package detect

import (
	"github.com/zhangkui/specimen-flow-audit/internal/waveform"
	"math"
)

type Event struct {
	StartIndex int
	EndIndex   int
	Peak       float64
	Energy     float64
}

func Threshold(trace waveform.Trace, threshold float64) []Event {
	result := []Event{}
	var current *Event
	for index, sample := range trace.Samples {
		if math.Abs(sample) < threshold {
			if current != nil {
				current.EndIndex = index
				result = append(result, *current)
				current = nil
			}
			continue
		}
		if current == nil {
			current = &Event{StartIndex: index, Peak: math.Abs(sample)}
		}
		if math.Abs(sample) > current.Peak {
			current.Peak = math.Abs(sample)
		}
		current.Energy += sample * sample
	}
	if current != nil {
		current.EndIndex = len(trace.Samples)
		result = append(result, *current)
	}
	return result
}
