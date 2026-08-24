package alert

import (
	"github.com/zhangkui/specimen-flow-audit/internal/detect"
	"github.com/zhangkui/specimen-flow-audit/internal/qc"
	"github.com/zhangkui/specimen-flow-audit/internal/waveform"
	"sort"
	"time"
)

type Level string

const (
	Info     Level = "info"
	Warning  Level = "warning"
	Critical Level = "critical"
)

type Alert struct {
	Station string    `json:"station"`
	Level   Level     `json:"level"`
	Code    string    `json:"code"`
	At      time.Time `json:"at"`
	Detail  string    `json:"detail"`
}

func Build(trace waveform.Trace, findings []qc.Finding, events []detect.Event) []Alert {
	alerts := []Alert{}
	for _, finding := range findings {
		level := Warning
		if finding.Severity == qc.Failure {
			level = Critical
		}
		alerts = append(alerts, Alert{Station: trace.Station, Level: level, Code: finding.Rule, At: trace.Start, Detail: finding.Message})
	}
	for _, event := range events {
		if event.Peak > 0 {
			alerts = append(alerts, Alert{Station: trace.Station, Level: Info, Code: "wave-event", At: trace.Start.Add(time.Duration(float64(event.StartIndex) / trace.SampleRate * float64(time.Second))), Detail: "threshold event detected"})
		}
	}
	sort.Slice(alerts, func(i, j int) bool { return alerts[i].At.Before(alerts[j].At) })
	return alerts
}
