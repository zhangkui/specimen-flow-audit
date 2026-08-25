package regression

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/zhangkui/specimen-flow-audit/internal/calibration"
	"github.com/zhangkui/specimen-flow-audit/internal/ledger"
	"github.com/zhangkui/specimen-flow-audit/internal/qc"
	"github.com/zhangkui/specimen-flow-audit/internal/ruleset"
	"github.com/zhangkui/specimen-flow-audit/internal/station"
	"github.com/zhangkui/specimen-flow-audit/internal/workflow"
)

func bug10Workflow(t *testing.T) (*workflow.Coordinator, time.Time) {
	t.Helper()
	now := time.Date(2026, 8, 25, 13, 0, 0, 0, time.UTC)
	stations := station.New()
	if err := stations.Upsert(station.Profile{Code: "HN01", Network: "HN", Channels: []string{"HHZ"}, Calibration: calibration.Curve{Gain: 1, Min: -10, Max: 10}, Policy: qc.Policy{MaxRMS: 1, MaxGapFraction: .9, MaxAbsolute: 2}, Trigger: 3}); err != nil {
		t.Fatal(err)
	}
	rules := ruleset.New()
	if err := rules.Publish(ruleset.Version{ID: "r1", Station: "HN01", EffectiveAt: now.Add(-time.Hour), Policy: qc.Policy{MaxRMS: 1, MaxGapFraction: .9, MaxAbsolute: 2}, Author: "operator"}); err != nil {
		t.Fatal(err)
	}
	return workflow.New(stations, rules, ledger.New()), now
}

func bug10Body(start time.Time) *strings.Reader {
	return strings.NewReader(fmt.Sprintf(`{"station":"HN01","start":"%s","sample_rate":2,"samples":[9,9,9,9]}`, start.UTC().Format(time.RFC3339)))
}

func TestBug10_HealthDeduplicatesOverlappingResults(t *testing.T) {
	app, now := bug10Workflow(t)
	for _, job := range []string{"job-10-a", "job-10-b"} {
		if _, err := app.Submit(context.Background(), workflow.SubmitRequest{JobID: job, Format: "json", Station: "HN01", Start: now, SubmittedAt: now, Body: bug10Body(now)}); err != nil {
			t.Fatal(err)
		}
	}
	snapshot, err := app.Health("HN01", now, now.Add(2*time.Second), 10)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Availability.ReceivedSamples != 4 {
		t.Fatalf("health=%+v", snapshot.Availability)
	}
}
