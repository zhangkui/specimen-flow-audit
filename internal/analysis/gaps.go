package analysis

import "github.com/zhangkui/specimen-flow-audit/internal/waveform"

type Gap struct{ StartIndex, MissingSamples int }

func Gaps(trace waveform.Trace, missingValue float64) []Gap {
	gaps := []Gap{}
	for index := 0; index < len(trace.Samples); {
		if trace.Samples[index] != missingValue {
			index++
			continue
		}
		start := index
		for index < len(trace.Samples) && trace.Samples[index] == missingValue {
			index++
		}
		gaps = append(gaps, Gap{StartIndex: start, MissingSamples: index - start})
	}
	return gaps
}
