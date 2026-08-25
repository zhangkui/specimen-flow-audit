package waveform

import "time"

type Trace struct {
	Station    string    `json:"station"`
	Start      time.Time `json:"start"`
	SampleRate float64   `json:"sample_rate"`
	Samples    []float64 `json:"samples"`
}

func (t Trace) End() time.Time {
	if t.SampleRate <= 0 || len(t.Samples) == 0 {
		return t.Start
	}
	return t.Start.Add(time.Duration(float64(len(t.Samples)) / t.SampleRate * float64(time.Second)))
}
