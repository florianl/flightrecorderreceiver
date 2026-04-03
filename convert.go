package flightrecorderreceiver

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/pprofile"
	"go.uber.org/zap"
	"golang.org/x/exp/trace"

	semconv "go.opentelemetry.io/otel/semconv/v1.39.0"
)

// extractMetricNameUnit extracts the name and unit from a flight recorder metric name.
// There is no 1:1 mapping between runtime/metrics and OTels semconv. Therefore,
// keep the naming of runtime/metrics.
func extractMetricNameUnit(traceName string) (name, unit string) {
	name, unit, _ = strings.Cut(traceName, ":")
	return name, unit
}

// convert converts a Flight Recorder trace from the provided reader into
// both OpenTelemetry Profiles and Metrics data structures.
func convert(ctx context.Context, logger *zap.Logger, f io.Reader) (pprofile.Profiles, pmetric.Metrics, error) {
	r, err := trace.NewReader(f)
	if err != nil {
		return pprofile.Profiles{}, pmetric.Metrics{}, err
	}
	lt := createLookupTable()

	profiles := pprofile.NewProfiles()
	rpSlice := profiles.ResourceProfiles()
	currentResourceProfile := rpSlice.AppendEmpty()
	currentResourceProfile.SetSchemaUrl(semconv.SchemaURL)
	spSlice := currentResourceProfile.ScopeProfiles()

	metrics := pmetric.NewMetrics()
	rmSlice := metrics.ResourceMetrics()
	currentResourceMetric := rmSlice.AppendEmpty()
	currentResourceMetric.SetSchemaUrl(semconv.SchemaURL)
	smSlice := currentResourceMetric.ScopeMetrics()
	currentScopeMetric := smSlice.AppendEmpty()
	currentScopeMetric.SetSchemaUrl(semconv.SchemaURL)
	metricsMap := make(map[string]pmetric.Metric)

	// most recent clock information from sync events
	var clockSnap *trace.ClockSnapshot

	// rangeState tracks an in-progress range for a single goroutine.
	type rangeState struct {
		startTS     time.Time
		lastEventTS time.Time
		profile     pprofile.Profile
		initialized bool // true once at least one sample has been added
	}
	// activeRanges maps goroutine ID to its current range state.
	activeRanges := make(map[trace.GoID]*rangeState)

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
			return pprofile.Profiles{}, pmetric.Metrics{}, err
		}
		switch ev.Kind() {
		case trace.EventSync:
			s := ev.Sync()
			if s.ClockSnapshot != nil {
				clockSnap = s.ClockSnapshot
			}
			continue eventLoop
		case trace.EventMetric:
			// Extract metrics from the event
			if clockSnap == nil {
				logger.Error("received EventMetric before clock synchronization")
				continue eventLoop
			}

			m := ev.Metric()
			metricName, metricUnit := extractMetricNameUnit(m.Name)

			// Get or create metric
			metric, exists := metricsMap[metricName]
			if !exists {
				metric = currentScopeMetric.Metrics().AppendEmpty()
				metric.SetName(metricName)
				metric.SetUnit(metricUnit)
				metric.SetEmptyGauge()
				metricsMap[metricName] = metric
			}

			// Add data point
			dp := metric.Gauge().DataPoints().AppendEmpty()
			dp.SetTimestamp(pcommon.NewTimestampFromTime(eventWallTime(ev.Time(), clockSnap)))
			// Flight recorder metrics are uint64 values
			dp.SetDoubleValue(float64(m.Value.Uint64()))

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
			// Finalize any previously initialized state for this goroutine before
			// replacing it; otherwise its profile would remain at time=0, duration=0.
			if existing, ok := activeRanges[ev.Goroutine()]; ok && existing.initialized {
				beginTS := eventWallTime(ev.Time(), clockSnap)
				existing.profile.SetTime(pcommon.NewTimestampFromTime(existing.startTS))
				existing.profile.SetDurationNano(uint64(beginTS.Sub(existing.startTS).Nanoseconds()) + 1)
			}
			activeRanges[ev.Goroutine()] = &rangeState{
				startTS: eventWallTime(ev.Time(), clockSnap),
			}
			// Fall through to add a sample at startTS.
		case trace.EventRangeEnd:
			goID := ev.Goroutine()
			state, ok := activeRanges[goID]
			if !ok {
				logger.Error("received EventRangeEnd without matching EventRangeBegin")
				continue eventLoop
			}
			delete(activeRanges, goID)
			if !state.initialized {
				// Do not generate data with empty profiles
				continue eventLoop
			}
			endTS := eventWallTime(ev.Time(), clockSnap)
			state.profile.SetTime(pcommon.NewTimestampFromTime(state.startTS))
			// Use +1 so the half-open range [startTS, endTS+1) includes any sample
			// that lands at exactly endTS (state transitions can share a timestamp
			// with the range-end event).
			state.profile.SetDurationNano(uint64(endTS.Sub(state.startTS).Nanoseconds()) + 1)
			continue eventLoop
		case trace.EventStateTransition:
			if clockSnap == nil {
				logger.Error("received EventStateTransition before clock synchonization")
				continue eventLoop
			}
			// Just unwind the stack — fall through to add a sample.
		default:
			logger.Debug(fmt.Sprintf("Skipping event kind %s", ev.Kind().String()))
			continue eventLoop
		}

		eventHasFrames := false
		for range ev.Stack().Frames() {
			eventHasFrames = true
			break
		}
		if !eventHasFrames {
			// Ignore events without a stack
			continue
		}

		// EventRangeBegin and EventStateTransition fall through to here.
		goID := ev.Goroutine()
		state, ok := activeRanges[goID]
		if !ok {
			state = &rangeState{
				startTS: eventWallTime(ev.Time(), clockSnap),
			}
			activeRanges[ev.Goroutine()] = state
			logger.Warn(fmt.Sprintf("Received event for GoID %v without prior EventRangeBegin", goID))
		}

		if !state.initialized {
			// Lazily create the ScopeProfile and Profile on the first sample.
			sp := spSlice.AppendEmpty()
			sp.SetSchemaUrl(semconv.SchemaURL)
			state.profile = sp.Profiles().AppendEmpty()
			initializeProfile(lt, state.profile)
			state.initialized = true
		}

		wallclockTS := eventWallTime(ev.Time(), clockSnap)
		state.lastEventTS = wallclockTS
		newSample := state.profile.Samples().AppendEmpty()

		// GoID is not part of OTel SemConv - so hardcode it here.
		goIDAttr := lt.AddKeyValueUnit("GoID", strconv.Itoa(int(goID)), "")
		newSample.AttributeIndices().Append(goIDAttr)

		if err := populateSample(lt, newSample, ev.Stack(), wallclockTS.UnixNano()); err != nil {
			return pprofile.Profiles{}, pmetric.Metrics{}, err
		}
	}

	if err := populateDictionary(lt, profiles.Dictionary()); err != nil {
		return pprofile.Profiles{}, pmetric.Metrics{}, err
	}

	// Dump remaining events that were missing proper EventRangeBegin and EventRangeEnd.
	for _, state := range activeRanges {
		if !state.initialized {
			continue
		}
		state.profile.SetTime(pcommon.NewTimestampFromTime(state.startTS))
		// Use +1 so the half-open range [startTS, lastEventTS+1) includes the
		// last sample, which is at lastEventTS.
		state.profile.SetDurationNano(uint64(state.lastEventTS.Sub(state.startTS).Nanoseconds()) + 1)
	}

	return profiles, metrics, nil
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
