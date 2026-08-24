package health_test

import (
	"testing"

	"github.com/zhangkui/specimen-flow-audit/internal/availability"
	"github.com/zhangkui/specimen-flow-audit/internal/baseline"
	"github.com/zhangkui/specimen-flow-audit/internal/health"
	"github.com/zhangkui/specimen-flow-audit/internal/report"
)

func TestAssessUsesOperationalThresholds(t *testing.T) {
	healthy := health.Assess(health.Input{Station: "HN01", HasLatest: true, Latest: report.Quality{Completeness: 0.95}, Availability: availability.Summary{Availability: 0.95}, HasBaseline: true, Baseline: baseline.Model{Points: 2}, AnomalyScore: 2.9})
	if healthy.State != health.Healthy || len(healthy.Reasons) != 0 {
		t.Fatalf("healthy snapshot = %+v", healthy)
	}

	degraded := health.Assess(health.Input{Station: "HN01", HasLatest: true, Latest: report.Quality{Completeness: 0.94}, Availability: availability.Summary{Availability: 0.95}})
	if degraded.State != health.Degraded || len(degraded.Reasons) != 1 {
		t.Fatalf("degraded snapshot = %+v", degraded)
	}

	critical := health.Assess(health.Input{Station: "HN01", HasLatest: true, Latest: report.Quality{Completeness: 0.99}, Availability: availability.Summary{Availability: 0.81}, CriticalIncidents: 1})
	if critical.State != health.Critical || critical.Reasons[0] != "存在未关闭的严重事件" {
		t.Fatalf("critical snapshot = %+v", critical)
	}
}

func TestAssessMarksStationWithoutTraceUnknown(t *testing.T) {
	snapshot := health.Assess(health.Input{Station: "HN02"})
	if snapshot.State != health.Unknown || len(snapshot.Reasons) != 1 {
		t.Fatalf("snapshot = %+v", snapshot)
	}
}
