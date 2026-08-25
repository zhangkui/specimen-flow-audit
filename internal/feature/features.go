package feature

import (
	"github.com/zhangkui/specimen-flow-audit/internal/waveform"
	"math"
	"sort"
)

type Set struct {
	Mean              float64 `json:"mean"`
	Peak              float64 `json:"peak"`
	PeakIndex         int     `json:"peak_index"`
	StandardDeviation float64 `json:"standard_deviation"`
	P95               float64 `json:"p95"`
	ZeroCrossings     int     `json:"zero_crossings"`
}

func Extract(trace waveform.Trace) Set {
	if len(trace.Samples) == 0 {
		return Set{PeakIndex: -1}
	}
	set := Set{PeakIndex: 0}
	for _, sample := range trace.Samples {
		set.Mean += sample
	}
	set.Mean /= float64(len(trace.Samples))
	variance := 0.0
	sorted := append([]float64(nil), trace.Samples...)
	for index, sample := range trace.Samples {
		absolute := math.Abs(sample)
		if absolute > set.Peak {
			set.Peak = absolute
			set.PeakIndex = index
		}
		delta := sample - set.Mean
		variance += delta * delta
		if index > 0 && ((sample < 0) != (trace.Samples[index-1] < 0)) {
			set.ZeroCrossings++
		}
	}
	set.StandardDeviation = math.Sqrt(variance / float64(len(trace.Samples)))
	sort.Float64s(sorted)
	position := int(math.Ceil(float64(len(sorted))*0.95)) - 1
	if position < 0 {
		position = 0
	}
	set.P95 = sorted[position]
	return set
}
