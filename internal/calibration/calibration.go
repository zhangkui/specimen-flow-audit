package calibration

import "errors"

var ErrInvalid = errors.New("invalid calibration")

type Curve struct {
	Station string
	Gain    float64
	Offset  float64
	Min     float64
	Max     float64
}

func (c Curve) Apply(samples []float64) ([]float64, error) {
	if c.Gain == 0 || c.Min >= c.Max {
		return nil, ErrInvalid
	}
	result := make([]float64, len(samples))
	for index, sample := range samples {
		result[index] = sample*c.Gain + c.Offset
	}
	return result, nil
}

func Clamp(samples []float64, min, max float64) []float64 {
	result := make([]float64, len(samples))
	for index, sample := range samples {
		if sample < min {
			sample = min
		}
		if sample > max {
			sample = max
		}
		result[index] = sample
	}
	return result
}
