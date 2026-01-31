package flightrecorderreceiver

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pprofile"
	"go.uber.org/zap"
	"golang.org/x/exp/trace"

	semconv "go.opentelemetry.io/otel/semconv/v1.38.0"
)

// convert converts a Flight Recorder trace from the provided reader into
// an OpenTelemetry Profiles data structure.
//
// TODOs:
//   - deduplicate data
func convert(ctx context.Context, logger *zap.Logger, f io.Reader) (pprofile.Profiles, error) {
	r, err := trace.NewReader(f)
	if err != nil {
		return pprofile.Profiles{}, err
	}
	lt := createLookupTable()

	profiles := pprofile.NewProfiles()
	rpSlice := profiles.ResourceProfiles()
	currentResourceProfile := rpSlice.AppendEmpty()
	currentResourceProfile.SetSchemaUrl(semconv.SchemaURL)
	spSlice := currentResourceProfile.ScopeProfiles()

	// most recent clock information from sync events
	var clockSnap *trace.ClockSnapshot
	// startTS keeps track of the start time of a range
	var startTS time.Time

	// currentXYZ keeps track of the current scope
	var currentScopeProfile pprofile.ScopeProfiles
	var currentProfile pprofile.Profile
	addedFrames := false

eventLoop:
	for {
		select {
		case <-ctx.Done():
			break eventLoop
		default:
		}

		ev, err := r.ReadEvent()
		if err != nil {
			if err == io.EOF {
				break
			}
			return pprofile.Profiles{}, err
		}
		switch ev.Kind() {
		case trace.EventSync:
			s := ev.Sync()
			if s.ClockSnapshot != nil {
				clockSnap = s.ClockSnapshot
			}
			continue eventLoop
		case trace.EventMetric:
			// Skip these events for the moment.
			// TODO: Figure out if and how these can be reported
			continue eventLoop
		case trace.EventLabel, trace.EventExperimental:
			// Skip these events for the moment.
			// TODO: Figure out if and how these can be represented in OTel Profiles
			continue eventLoop
		case trace.EventRangeBegin:
			if clockSnap == nil {
				logger.Error("received EventRangeBegin before clock synchonization")
				continue eventLoop
			}
			startTS = eventWallTime(ev.Time(), clockSnap)

			// Use a dedicated ScopeProfile per EventRange
			currentScopeProfile = spSlice.AppendEmpty()
			currentProfile = currentScopeProfile.Profiles().AppendEmpty()
			currentScopeProfile.SetSchemaUrl(semconv.SchemaURL)

			initializeProfile(lt, currentProfile)
		case trace.EventRangeEnd:
			if !addedFrames {
				// Do not generated data with empty profiles
				continue eventLoop
			}
			addedFrames = false

			if startTS.IsZero() {
				logger.Error("received EventRangeEnd before EventRangeBegin")
				continue eventLoop
			}
			currentProfile.SetTime(pcommon.NewTimestampFromTime(startTS))
			endTS := eventWallTime(ev.Time(), clockSnap)
			duration := endTS.Sub(startTS).Nanoseconds()
			currentProfile.SetDurationNano(uint64(duration))
		case trace.EventStateTransition:
			// Just unwind the stack
		default:
			logger.Debug(fmt.Sprintf("Skipping event kind %s", ev.Kind().String()))
			continue eventLoop
		}

		wallclockTS := eventWallTime(ev.Time(), clockSnap)

		eventHasFrames := false
		for range ev.Stack().Frames() {
			eventHasFrames = true
			break
		}
		if !eventHasFrames {
			// Ignore events without a stack
			continue
		}

		newSample := currentProfile.Samples().AppendEmpty()

		// GoID is not part of OTel  SemConv - so hardcode it here.
		goIDAttr := lt.AddKeyValueUnit("GoID", strconv.Itoa(int(ev.Goroutine())), "")
		newSample.AttributeIndices().Append(goIDAttr)

		if err := populateSample(lt, newSample, ev.Stack(), wallclockTS.UnixNano()); err != nil {
			return pprofile.Profiles{}, err
		}
		addedFrames = true
	}

	if err := populateDictionary(lt, profiles.Dictionary()); err != nil {
		return pprofile.Profiles{}, err
	}

	return profiles, nil
}

// initializeProfile sets default values for Profile.
func initializeProfile(lt lookupTable, p pprofile.Profile) {
	// Report the flightrecord as wallclock profile.
	p.SampleType().SetTypeStrindex(lt.AddString("wall"))
	p.SampleType().SetUnitStrindex(lt.AddString("nanoseconds"))
}

// populateSample converts a single trace.Stack into a pprofile.Sample.
func populateSample(lt lookupTable, s pprofile.Sample, stack trace.Stack, ts int64) error {
	s.TimestampsUnixNano().Append(uint64(ts))

	var li []int32
	for frame := range stack.Frames() {
		loc := lt.AddLocation(frame)
		li = append(li, loc)
	}

	si := lt.AddStack(li)
	s.SetStackIndex(si)

	return nil
}

// eventWallTime calculates the absolute time.Time for a given event based
// on a reference trace.ClockSnapshot.
func eventWallTime(evTime trace.Time, snap *trace.ClockSnapshot) time.Time {
	diff := time.Duration(evTime - snap.Trace)
	return snap.Wall.Add(diff)
}

func populateDictionary(lt lookupTable, dic pprofile.ProfilesDictionary) error {
	// mappings - currently not used
	// As MappingTable is not initizalied in createLookupTable
	// create the empty element here.
	dic.MappingTable().AppendEmpty()

	// locations
	for range lt.locations {
		dic.LocationTable().AppendEmpty()
	}
	for _, l := range lt.locations {
		dic.LocationTable().At(int(l.locationTableIdx)).SetAddress(l.address)
		for _, line := range l.lines {
			ln := dic.LocationTable().At(int(l.locationTableIdx)).Lines().AppendEmpty()
			ln.SetFunctionIndex(line.functionIdx)
			ln.SetLine(line.line)
			ln.SetColumn(line.column)
		}
	}

	// functions
	for range lt.functions {
		dic.FunctionTable().AppendEmpty()
	}
	for a, idx := range lt.functions {
		dic.FunctionTable().At(int(idx)).SetNameStrindex(a.nameIdx)
		dic.FunctionTable().At(int(idx)).SetSystemNameStrindex(a.systemNameIdx)
		dic.FunctionTable().At(int(idx)).SetFilenameStrindex(a.fileNameIdx)
		dic.FunctionTable().At(int(idx)).SetStartLine(a.startLine)
	}

	// links - currently not used
	// As LinkTable is not initizalied in createLookupTable
	// create the empty element here.
	dic.LinkTable().AppendEmpty()

	// strings
	for range lt.strings {
		dic.StringTable().Append("")
	}
	for s, idx := range lt.strings {
		dic.StringTable().SetAt(int(idx), s)
	}

	// attributes
	for range lt.attributes {
		dic.AttributeTable().AppendEmpty()
	}
	for a, idx := range lt.attributes {
		dic.AttributeTable().At(int(idx)).SetKeyStrindex(a.keyIdx)
		dic.AttributeTable().At(int(idx)).SetUnitStrindex(a.unitIdx)
		dic.AttributeTable().At(int(idx)).Value().SetStr(a.value)
	}

	// stacks
	for range lt.stacks {
		dic.StackTable().AppendEmpty()
	}
	for _, s := range lt.stacks {
		dic.StackTable().At(int(s.stackTableIdx)).LocationIndices().Append(s.locationIndices...)
	}

	return nil
}
