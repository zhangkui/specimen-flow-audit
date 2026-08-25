package qc

import (
	"github.com/zhangkui/specimen-flow-audit/internal/analysis"
	"github.com/zhangkui/specimen-flow-audit/internal/waveform"
	"math"
)

type Severity string

const (
	Warning Severity = "warning"
	Failure Severity = "failure"
)

type Finding struct {
	Rule     string   `json:"rule"`
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
}
type Policy struct {
	MaxRMS         float64
	MaxGapFraction float64
	MaxAbsolute    float64
}

func Evaluate(trace waveform.Trace, policy Policy) []Finding {
	findings := []Finding{}
	if analysis.RMS(trace) > policy.MaxRMS {
		findings = append(findings, Finding{"noise-rms", Failure, "RMS exceeds station limit"})
	}
	missing := 0
	for _, gap := range analysis.Gaps(trace, 0) {
		missing += gap.MissingSamples
	}
	if float64(missing)/float64(len(trace.Samples)) > policy.MaxGapFraction {
		findings = append(findings, Finding{"missing-data", Failure, "missing sample fraction exceeds limit"})
	}
	for _, sample := range trace.Samples {
		if math.Abs(sample) > policy.MaxAbsolute {
			findings = append(findings, Finding{"amplitude", Warning, "sample exceeds expected range"})
			break
		}
	}
	return findings
}
