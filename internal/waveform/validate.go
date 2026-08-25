package waveform

import (
	"errors"
	"math"
)

var ErrInvalidTrace = errors.New("invalid trace")

func Validate(trace Trace) error {
	if trace.Station == "" || trace.Start.IsZero() || trace.SampleRate <= 0 || len(trace.Samples) == 0 {
		return ErrInvalidTrace
	}
	for _, sample := range trace.Samples {
		if math.IsInf(sample, 0) {
			return ErrInvalidTrace
		}
	}
	return nil
}
