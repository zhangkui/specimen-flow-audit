package compare

import (
	"errors"
	"math"

	"github.com/zhangkui/specimen-flow-audit/internal/waveform"
)

var ErrIncompatible = errors.New("traces are incompatible")

type Result struct {
	Left            string  `json:"left"`
	Right           string  `json:"right"`
	Correlation     float64 `json:"correlation"`
	LagSamples      int     `json:"lag_samples"`
	ComparedSamples int     `json:"compared_samples"`
}

func Correlate(left, right waveform.Trace, maxLag int) (Result, error) {
	if left.SampleRate != right.SampleRate || len(left.Samples) == 0 || len(right.Samples) == 0 {
		return Result{}, ErrIncompatible
	}
	best := Result{Left: left.Station, Right: right.Station, Correlation: -2}
	for lag := -maxLag; lag <= maxLag; lag++ {
		score, count := correlationAt(left.Samples, right.Samples, lag)
		if count > 0 && score > best.Correlation {
			best.Correlation = score
			best.LagSamples = lag
			best.ComparedSamples = count
		}
	}
	if best.Correlation == -2 {
		return Result{}, ErrIncompatible
	}
	return best, nil
}
func correlationAt(left, right []float64, lag int) (float64, int) {
	sumX, sumY, sumXX, sumYY, sumXY := 0.0, 0.0, 0.0, 0.0, 0.0
	count := 0
	for index, value := range left {
		other := index + lag
		if other < 0 || other >= len(right) {
			continue
		}
		second := right[other]
		sumX += value
		sumY += second
		sumXX += value * value
		sumYY += second * second
		sumXY += value * second
		count++
	}
	if count < 2 {
		return 0, count
	}
	numerator := float64(count)*sumXY - sumX*sumY
	denominator := math.Sqrt((float64(count)*sumXX - sumX*sumX) * (float64(count)*sumYY - sumY*sumY))
	if denominator == 0 {
		return 0, count
	}
	return numerator / denominator, count
}
