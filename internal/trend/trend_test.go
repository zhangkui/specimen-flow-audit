package trend_test

import (
	"testing"
	"time"

	"github.com/zhangkui/specimen-flow-audit/internal/report"
	"github.com/zhangkui/specimen-flow-audit/internal/trend"
)

func TestAnalyzeMarksRisingNoiseAsDegrading(t *testing.T) {
	start := time.Date(2026, 8, 24, 0, 0, 0, 0, time.UTC)
	snapshot, err := trend.Analyze("HN01", []report.Quality{
		{Station: "HN01", ObservedAt: start, NoiseRMS: 1, Completeness: 0.99},
		{Station: "HN01", ObservedAt: start.Add(time.Hour), NoiseRMS: 1.3, Completeness: 0.98},
		{Station: "HN01", ObservedAt: start.Add(2 * time.Hour), NoiseRMS: 1.6, Completeness: 0.97},
	}, trend.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.State != trend.Degrading || snapshot.NoiseSlopePerHour < 0.29 || len(snapshot.Reasons) < 1 {
		t.Fatalf("snapshot = %+v", snapshot)
	}
}

func TestAnalyzeDetectsStableBoundaryAndInsufficientHistory(t *testing.T) {
	start := time.Date(2026, 8, 24, 0, 0, 0, 0, time.UTC)
	stable, err := trend.Analyze("HN01", []report.Quality{
		{Station: "HN01", ObservedAt: start, NoiseRMS: 2, Completeness: 0.95},
		{Station: "HN01", ObservedAt: start.Add(time.Hour), NoiseRMS: 2.1, Completeness: 0.95},
		{Station: "HN01", ObservedAt: start.Add(2 * time.Hour), NoiseRMS: 2, Completeness: 0.95},
	}, trend.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if stable.State != trend.Stable {
		t.Fatalf("stable snapshot = %+v", stable)
	}
	insufficient, err := trend.Analyze("HN01", []report.Quality{{Station: "HN01", ObservedAt: start, NoiseRMS: 2, Completeness: 0.95}}, trend.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if insufficient.State != trend.Insufficient || len(insufficient.Reasons) != 1 {
		t.Fatalf("insufficient snapshot = %+v", insufficient)
	}
}

func TestAnalyzeRejectsMixedStationHistory(t *testing.T) {
	_, err := trend.Analyze("HN01", []report.Quality{{Station: "HN01"}, {Station: "HN02"}}, trend.Config{})
	if err != trend.ErrMixedStation {
		t.Fatalf("error = %v", err)
	}
}
