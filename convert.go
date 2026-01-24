package flightrecorderreceiver

import (
	"context"
	"io"
	"strconv"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pprofile"
	"golang.org/x/exp/trace"

	semconv "go.opentelemetry.io/otel/semconv/v1.38.0"
)

// convert converts a Flight Recorder trace from the provided reader into
// an OpenTelemetry Profiles data structure.
//
// TODOs:
//   - deduplicate data
func convert(ctx context.Context, f io.Reader) (pprofile.Profiles, error) {
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
	currentScopeProfile := spSlice.AppendEmpty()
	currentScopeProfile.SetSchemaUrl(semconv.SchemaURL)
	currentProfile := currentScopeProfile.Profiles().AppendEmpty()

	// TODO: move LocationTable to lookupTable
	profiles.Dictionary().LocationTable().AppendEmpty()

	initializeProfile(lt, currentProfile)

	var baseTime time.Time
	var baseMono uint64

	startTS, endTS := time.Unix(0, 0), time.Unix(0, 0)
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
				baseTime = s.ClockSnapshot.Wall
				baseMono = s.ClockSnapshot.Mono
				startTS = baseTime
			}
			continue
		case trace.EventMetric:
			// Skip these events for the moment.
			// TODO: Figure out if and how these can be reported
			continue
		case trace.EventLabel, trace.EventExperimental:
			// Skip these events for the moment.
			// TODO: Figure out if and how these can be represented in OTel Profiles
			continue
		case trace.EventStateTransition:
		case trace.EventRangeBegin, trace.EventRangeEnd:
			if !addedFrames {
				// Do not generated data with empty profiles
				continue
			}
			addedFrames = false

			if startTS.IsZero() {
				// No stack was recorded so far.
				continue
			}
			currentProfile.SetTime(pcommon.NewTimestampFromTime(startTS))
			duration := endTS.Sub(startTS).Nanoseconds()
			currentProfile.SetDurationNano(uint64(duration))

			// Use a dedicated ScopeProfile per EventRange
			startTS = baseTime
			currentScopeProfile = spSlice.AppendEmpty()
			currentProfile = currentScopeProfile.Profiles().AppendEmpty()

			initializeProfile(lt, currentProfile)
		default:
			// TODO: Log skipped event kind
			continue
		}

		if ev.Kind() == trace.EventRangeBegin || ev.Kind() == trace.EventRangeEnd {
			// Skip stack unwinding for EventRagen[Begin|End]
			continue
		}

		endTS = convertTimestamp(baseTime, baseMono, ev.Time())

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

		if err := populateSample(lt, profiles.Dictionary(), newSample, ev.Stack(), endTS.UnixNano()); err != nil {
			return pprofile.Profiles{}, err
		}
		addedFrames = true
	}

	currentProfile.SetTime(pcommon.NewTimestampFromTime(startTS))
	duration := endTS.Sub(startTS).Nanoseconds()
	currentProfile.SetDurationNano(uint64(duration))

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
func populateSample(lt lookupTable, dic pprofile.ProfilesDictionary, s pprofile.Sample, stack trace.Stack, ts int64) error {
	s.TimestampsUnixNano().Append(uint64(ts))

	var li []int32
	for frame := range stack.Frames() {
		loc := dic.LocationTable().AppendEmpty()
		li = append(li, int32(dic.LocationTable().Len()-1))

		loc.SetAddress(frame.PC)
		ln := loc.Lines().AppendEmpty()

		ln.SetLine(int64(frame.Line))

		ln.SetFunctionIndex(lt.AddFunction(frame.Func, "", frame.File, int64(frame.Line)))
	}

	si := lt.AddStack(li)
	s.SetStackIndex(si)

	return nil
}

func convertTimestamp(base time.Time, baseMono uint64, ts trace.Time) time.Time {
	diff := uint64(ts) - baseMono
	return base.Add(time.Duration(diff))
}

func populateDictionary(lt lookupTable, dic pprofile.ProfilesDictionary) error {
	// mappings - currently not used
	// As MappingTable is not initizalied in createLookupTable
	// create the empty element here.
	dic.MappingTable().AppendEmpty()

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
