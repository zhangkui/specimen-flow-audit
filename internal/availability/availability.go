package availability

import (
	"github.com/zhangkui/specimen-flow-audit/internal/maintenance"
	"sort"
	"time"
)

type Interval struct {
	Station         string    `json:"station"`
	Start           time.Time `json:"start"`
	End             time.Time `json:"end"`
	ReceivedSamples int       `json:"received_samples"`
	ExpectedSamples int       `json:"expected_samples"`
	Maintenance     bool      `json:"maintenance"`
}
type Summary struct {
	Station              string    `json:"station"`
	Start                time.Time `json:"start"`
	End                  time.Time `json:"end"`
	ExpectedSamples      int       `json:"expected_samples"`
	ReceivedSamples      int       `json:"received_samples"`
	CoveredByMaintenance int       `json:"covered_by_maintenance"`
	Availability         float64   `json:"availability"`
}

func Measure(station string, start, end time.Time, sampleRate float64, received int, calendar *maintenance.Calendar) Summary {
	expected := int(end.Sub(start).Seconds() * sampleRate)
	summary := Summary{Station: station, Start: start, End: end, ExpectedSamples: expected, ReceivedSamples: received}
	if expected <= 0 {
		return summary
	}
	for cursor := start; cursor.Before(end); cursor = cursor.Add(time.Second) {
		if _, ok := calendar.At(station, cursor); ok {
			summary.CoveredByMaintenance += int(sampleRate)
		}
	}
	effective := expected - summary.CoveredByMaintenance
	if effective < 1 {
		effective = 1
	}
	summary.Availability = float64(received) / float64(effective)
	if summary.Availability > 1 {
		summary.Availability = 1
	}
	return summary
}
func Merge(intervals []Interval) []Interval {
	items := append([]Interval(nil), intervals...)
	sort.Slice(items, func(i, j int) bool { return items[i].Start.Before(items[j].Start) })
	return items
}
