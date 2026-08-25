package pipeline

import (
	"github.com/zhangkui/specimen-flow-audit/internal/calibration"
	"github.com/zhangkui/specimen-flow-audit/internal/detect"
	"github.com/zhangkui/specimen-flow-audit/internal/qc"
	"github.com/zhangkui/specimen-flow-audit/internal/report"
	"github.com/zhangkui/specimen-flow-audit/internal/waveform"
)

type Request struct {
	Trace          waveform.Trace
	Calibration    calibration.Curve
	Policy         qc.Policy
	EventThreshold float64
}
type Result struct {
	Trace    waveform.Trace `json:"-"`
	Quality  report.Quality `json:"quality"`
	Findings []qc.Finding   `json:"findings"`
	Events   []detect.Event `json:"events"`
}

func Run(request Request) (Result, error) {
	if err := waveform.Validate(request.Trace); err != nil {
		return Result{}, err
	}
	samples, err := request.Calibration.Apply(request.Trace.Samples)
	if err != nil {
		return Result{}, err
	}
	trace := request.Trace
	trace.Samples = calibration.Clamp(samples, request.Calibration.Min, request.Calibration.Max)
	return Result{Trace: trace, Quality: report.Build(trace), Findings: qc.Evaluate(trace, request.Policy), Events: detect.Threshold(trace, request.EventThreshold)}, nil
}
