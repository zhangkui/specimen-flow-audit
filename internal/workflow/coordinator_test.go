package workflow_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/zhangkui/specimen-flow-audit/internal/calibration"
	"github.com/zhangkui/specimen-flow-audit/internal/ledger"
	"github.com/zhangkui/specimen-flow-audit/internal/maintenance"
	"github.com/zhangkui/specimen-flow-audit/internal/qc"
	"github.com/zhangkui/specimen-flow-audit/internal/ruleset"
	"github.com/zhangkui/specimen-flow-audit/internal/spectrum"
	"github.com/zhangkui/specimen-flow-audit/internal/station"
	"github.com/zhangkui/specimen-flow-audit/internal/trend"
	"github.com/zhangkui/specimen-flow-audit/internal/workflow"
)

func TestSubmitRunsFullQualityWorkflow(t *testing.T) {
	stations := station.New()
	if err := stations.Upsert(station.Profile{Code: "HN01", Network: "HN", Channels: []string{"HHZ"}, Calibration: calibration.Curve{Gain: 2, Min: -10, Max: 10}, Policy: qc.Policy{MaxRMS: 5, MaxGapFraction: 0.4, MaxAbsolute: 8}, Trigger: 3}); err != nil {
		t.Fatal(err)
	}
	rules := ruleset.New()
	now := time.Date(2026, 8, 24, 8, 0, 0, 0, time.UTC)
	if err := rules.Publish(ruleset.Version{ID: "r1", Station: "HN01", EffectiveAt: now.Add(-time.Hour), Policy: qc.Policy{MaxRMS: 5, MaxGapFraction: 0.4, MaxAbsolute: 8}, Author: "operator"}); err != nil {
		t.Fatal(err)
	}
	app := workflow.New(stations, rules, ledger.New())
	completed, err := app.Submit(context.Background(), workflow.SubmitRequest{JobID: "job-1", Format: "json", Station: "HN01", Start: now, SubmittedAt: now, Body: strings.NewReader(`{"station":"HN01","start":"2026-08-24T08:00:00Z","sample_rate":2,"samples":[9,0,0,0,9]}`)})
	if err != nil {
		t.Fatal(err)
	}
	if completed.Job.State != ledger.Completed {
		t.Fatalf("state = %s", completed.Job.State)
	}
	if completed.Result.Quality.Station != "HN01" {
		t.Fatalf("station = %s", completed.Result.Quality.Station)
	}
	if len(completed.Alerts) == 0 {
		t.Fatal("expected quality alert from calibrated high amplitude")
	}
	if len(app.OpenReviews("HN01")) == 0 {
		t.Fatal("expected review task for critical alert")
	}
	if len(app.PendingNotifications()) == 0 {
		t.Fatal("expected notification for critical alert")
	}
	if len(app.AuditTrail("job-1")) < 2 {
		t.Fatal("expected submitted and completed audit entries")
	}
	if len(app.Dashboard()) != 1 {
		t.Fatal("expected one dashboard card")
	}
	model, err := app.Baseline("HN01", 10)
	if err != nil {
		t.Fatal(err)
	}
	if model.Points != 1 || model.Station != "HN01" {
		t.Fatalf("baseline = %+v", model)
	}
	if _, err := app.AnomalyScore("job-1", 10); err != nil {
		t.Fatal(err)
	}
	content, manifest, err := app.Package("HN01")
	if err != nil {
		t.Fatal(err)
	}
	if len(content) == 0 || manifest.Reports != 1 || manifest.Alerts != len(completed.Alerts) {
		t.Fatalf("manifest = %+v, bytes = %d", manifest, len(content))
	}
	snapshot, err := app.Health("HN01", now, now.Add(2500*time.Millisecond), 10)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.State != "critical" || snapshot.CriticalIncidents == 0 {
		t.Fatalf("health snapshot = %+v", snapshot)
	}
	spectral, err := app.Spectrum("job-1", spectrum.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if spectral.AnalyzedSamples != 5 || spectral.TotalPower == 0 {
		t.Fatalf("spectral report = %+v", spectral)
	}
	findings, err := app.SpectralFindings("job-1", spectrum.Policy{})
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) == 0 {
		t.Fatal("expected spectral policy finding for the short pulse waveform")
	}
	comparison, err := app.CompareSpectrum("job-1", "job-1", spectrum.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if comparison.Similarity != 1 {
		t.Fatalf("spectral comparison = %+v", comparison)
	}
	trendSnapshot, err := app.Trend("HN01", time.Time{}, trend.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if trendSnapshot.State != trend.Insufficient {
		t.Fatalf("trend snapshot = %+v", trendSnapshot)
	}
	plan, err := app.Remediation("job-1", now, now.Add(2500*time.Millisecond), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Actions) == 0 || plan.Actions[0].Code != "dispatch-on-call" {
		t.Fatalf("remediation plan = %+v", plan)
	}
}

func TestMaintenanceSuppressesQualityIncidents(t *testing.T) {
	stations := station.New()
	now := time.Date(2026, 8, 24, 8, 0, 0, 0, time.UTC)
	if err := stations.Upsert(station.Profile{Code: "HN02", Network: "HN", Channels: []string{"HHZ"}, Calibration: calibration.Curve{Gain: 1, Min: -10, Max: 10}, Policy: qc.Policy{MaxRMS: 1, MaxGapFraction: 0.1, MaxAbsolute: 2}, Trigger: 99}); err != nil {
		t.Fatal(err)
	}
	rules := ruleset.New()
	if err := rules.Publish(ruleset.Version{ID: "r2", Station: "HN02", EffectiveAt: now.Add(-time.Hour), Policy: qc.Policy{MaxRMS: 1, MaxGapFraction: 0.1, MaxAbsolute: 2}, Author: "operator"}); err != nil {
		t.Fatal(err)
	}
	app := workflow.New(stations, rules, ledger.New())
	if err := app.ScheduleMaintenance(maintenance.Window{ID: "mw-1", Station: "HN02", Start: now.Add(-time.Minute), End: now.Add(time.Minute), Reason: "sensor swap", Planned: true}); err != nil {
		t.Fatal(err)
	}
	_, err := app.Submit(context.Background(), workflow.SubmitRequest{JobID: "job-2", Format: "json", Station: "HN02", Start: now, SubmittedAt: now, Body: strings.NewReader(`{"station":"HN02","start":"2026-08-24T08:00:00Z","sample_rate":2,"samples":[0,0,9,9]}`)})
	if err != nil {
		t.Fatal(err)
	}
	if len(app.OpenIncidents("HN02")) != 0 {
		t.Fatal("maintenance window should suppress quality incidents")
	}
}
