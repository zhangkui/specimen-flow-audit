package analysis

import (
	"github.com/zhangkui/specimen-flow-audit/internal/waveform"
	"math"
)

func RMS(trace waveform.Trace) float64 {
	var sum float64
	for _, sample := range trace.Samples {
		sum += sample * sample
	}
	return math.Sqrt(sum / float64(len(trace.Samples)))
}
