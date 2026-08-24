package assembly

import (
	"errors"
	"sort"
	"time"

	"github.com/zhangkui/specimen-flow-audit/internal/waveform"
)

var (
	ErrMixedStation = errors.New("segments have different stations")
	ErrMixedRate    = errors.New("segments have different sample rates")
)

type Result struct {
	Trace          waveform.Trace
	MissingSamples int
	OverlapSamples int
}

func Merge(segments []waveform.Segment, fill float64) (Result, error) {
	if len(segments) == 0 {
		return Result{}, nil
	}
	ordered := append([]waveform.Segment(nil), segments...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Trace.Start.Before(ordered[j].Trace.Start) })
	base := ordered[0].Trace
	merged := waveform.Trace{Station: base.Station, Start: base.Start, SampleRate: base.SampleRate}
	cursor := base.Start
	result := Result{Trace: merged}
	output := []float64{}
	for _, segment := range ordered {
		trace := segment.Trace
		if trace.Station != merged.Station {
			return Result{}, ErrMixedStation
		}
		if trace.SampleRate != merged.SampleRate {
			return Result{}, ErrMixedRate
		}
		if trace.Start.After(cursor) {
			missing := int(trace.Start.Sub(cursor).Seconds() * merged.SampleRate)
			for index := 0; index < missing; index++ {
				output = append(output, fill)
			}
			result.MissingSamples += missing
		}
		if trace.Start.Before(cursor) {
			overlap := int(cursor.Sub(trace.Start).Seconds() * merged.SampleRate)
			if overlap >= len(trace.Samples) {
				result.OverlapSamples += len(trace.Samples)
				continue
			}
			result.OverlapSamples += overlap
			trace.Samples = trace.Samples[overlap:]
		}
		output = append(output, trace.Samples...)
		cursor = trace.Start.Add(time.Duration(float64(len(trace.Samples)) / trace.SampleRate * float64(time.Second)))
	}
	result.Trace.Samples = output
	return result, nil
}
