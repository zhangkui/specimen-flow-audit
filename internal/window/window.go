package window

import (
	"errors"
	"github.com/zhangkui/specimen-flow-audit/internal/waveform"
	"time"
)

var ErrOutside = errors.New("trace is outside requested window")

type Range struct {
	Start time.Time
	End   time.Time
}

func (r Range) Valid() bool { return !r.Start.IsZero() && r.End.After(r.Start) }

func Crop(trace waveform.Trace, requested Range) (waveform.Trace, error) {
	if !requested.Valid() {
		return waveform.Trace{}, ErrOutside
	}
	traceEnd := trace.End()
	if trace.Start.Before(requested.Start) || traceEnd.After(requested.End) {
		return waveform.Trace{}, ErrOutside
	}
	return trace, nil
}

func Split(trace waveform.Trace, size int) []waveform.Segment {
	if size <= 0 {
		return nil
	}
	segments := []waveform.Segment{}
	for offset, sequence := 0, 0; offset < len(trace.Samples); offset, sequence = offset+size, sequence+1 {
		end := offset + size
		if end > len(trace.Samples) {
			end = len(trace.Samples)
		}
		part := trace
		part.Samples = append([]float64(nil), trace.Samples[offset:end]...)
		part.Start = trace.Start.Add(time.Duration(float64(offset) / trace.SampleRate * float64(time.Second)))
		segments = append(segments, waveform.Segment{Trace: part, Sequence: sequence})
	}
	return segments
}
