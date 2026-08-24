package remediation_test

import (
	"testing"

	"github.com/zhangkui/specimen-flow-audit/internal/availability"
	"github.com/zhangkui/specimen-flow-audit/internal/health"
	"github.com/zhangkui/specimen-flow-audit/internal/remediation"
	"github.com/zhangkui/specimen-flow-audit/internal/report"
	"github.com/zhangkui/specimen-flow-audit/internal/spectrum"
	"github.com/zhangkui/specimen-flow-audit/internal/trend"
)

func TestBuildOrdersCriticalOperationalActions(t *testing.T) {
	plan := remediation.Build(remediation.Input{
		Health:      health.Snapshot{Station: "HN01", State: health.Critical, HasLatest: true, Latest: report.Quality{Completeness: 0.70}, Availability: availability.Summary{Availability: 0.60}},
		Trend:       trend.Snapshot{State: trend.Degrading},
		Spectrum:    spectrum.Report{TotalPower: 10, SpectralEntropy: 0.2},
		HasSpectrum: true,
	})
	if len(plan.Actions) < 5 || plan.Actions[0].Code != "dispatch-on-call" || plan.Actions[0].Priority != remediation.Urgent {
		t.Fatalf("plan = %+v", plan)
	}
	if !hasAction(plan, "inspect-uplink") || !hasAction(plan, "backfill-waveform") || !hasAction(plan, "schedule-calibration") || !hasAction(plan, "inspect-narrowband-noise") {
		t.Fatalf("plan actions = %+v", plan.Actions)
	}
}

func TestBuildProvidesObservationForHealthyStation(t *testing.T) {
	plan := remediation.Build(remediation.Input{Health: health.Snapshot{Station: "HN02", State: health.Healthy, HasLatest: true, Latest: report.Quality{Completeness: 0.99}, Availability: availability.Summary{Availability: 0.99}}, Trend: trend.Snapshot{State: trend.Stable}})
	if len(plan.Actions) != 1 || plan.Actions[0].Code != "continue-observation" {
		t.Fatalf("plan = %+v", plan)
	}
}

func TestBuildRequestsTelemetryWhenNoTraceExists(t *testing.T) {
	plan := remediation.Build(remediation.Input{Health: health.Snapshot{Station: "HN03", State: health.Unknown}})
	if len(plan.Actions) != 1 || plan.Actions[0].Code != "verify-telemetry" || plan.Actions[0].Priority != remediation.Urgent {
		t.Fatalf("plan = %+v", plan)
	}
}

func hasAction(plan remediation.Plan, code string) bool {
	for _, action := range plan.Actions {
		if action.Code == code {
			return true
		}
	}
	return false
}
