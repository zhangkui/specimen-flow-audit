package waveform

import "time"

type Segment struct {
	Trace       Trace
	Sequence    int
	ReceivedAt  time.Time
	Source      string
	QualityFlag string
}

func (s Segment) Window() (time.Time, time.Time) { return s.Trace.Start, s.Trace.End() }

func (s Segment) IsContiguous(next Segment) bool {
	_, end := s.Window()
	start, _ := next.Window()
	return start.Sub(end) == 0
}
