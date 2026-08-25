package daily

import (
	"github.com/zhangkui/specimen-flow-audit/internal/report"
	"sort"
	"time"
)

type Summary struct {
	Station             string  `json:"station"`
	Day                 string  `json:"day"`
	TraceCount          int     `json:"trace_count"`
	AverageCompleteness float64 `json:"average_completeness"`
	MaximumNoise        float64 `json:"maximum_noise"`
	TotalGaps           int     `json:"total_gaps"`
}

func Summarize(reports []report.Quality, location *time.Location) []Summary {
	if location == nil {
		location = time.UTC
	}
	buckets := map[string]Summary{}
	for _, item := range reports {
		day := item.ObservedAt.In(location).Format("2006-01-02")
		key := item.Station + "|" + day
		summary := buckets[key]
		summary.Station = item.Station
		summary.Day = day
		summary.TraceCount++
		summary.AverageCompleteness += item.Completeness
		summary.TotalGaps += item.GapCount
		if item.NoiseRMS > summary.MaximumNoise {
			summary.MaximumNoise = item.NoiseRMS
		}
		buckets[key] = summary
	}
	result := make([]Summary, 0, len(buckets))
	for _, summary := range buckets {
		summary.AverageCompleteness /= float64(summary.TraceCount)
		result = append(result, summary)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Station < result[j].Station })
	return result
}
