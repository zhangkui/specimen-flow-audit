package report

import (
	"time"

	"github.com/zhangkui/specimen-flow-audit/internal/analysis"
	"github.com/zhangkui/specimen-flow-audit/internal/waveform"
)

type Quality struct {
	Station         string    `json:"station"`
	ObservedAt      time.Time `json:"observed_at"`
	DurationSeconds float64   `json:"duration_seconds"`
	GapCount        int       `json:"gap_count"`
	NoiseRMS        float64   `json:"noise_rms"`
	Completeness    float64   `json:"completeness"`
}

func Build(trace waveform.Trace) Quality {
	gaps := analysis.Gaps(trace, 0)
	missing := 0
	for _, gap := range gaps {
		missing += gap.MissingSamples
	}
	return Quality{Station: trace.Station, ObservedAt: trace.Start, DurationSeconds: trace.End().Sub(trace.Start).Seconds(), GapCount: len(gaps), NoiseRMS: analysis.RMS(trace), Completeness: 1 - float64(missing)/float64(len(trace.Samples))}
}
