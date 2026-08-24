package spectrum_test

import (
	"testing"

	"github.com/zhangkui/specimen-flow-audit/internal/spectrum"
)

func TestEvaluateFindsDominantToneAndBandExcess(t *testing.T) {
	findings, err := spectrum.Evaluate(spectrum.Report{
		TotalPower:      10,
		DominantPower:   8,
		SpectralEntropy: 0.2,
		Bands:           []spectrum.BandPower{{Name: "high-frequency", Share: 0.7}},
	}, spectrum.Policy{MinEntropy: 0.3, MaxDominantShare: 0.7, MaxBandShare: map[string]float64{"high-frequency": 0.5}})
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 3 || findings[0].Code != "dominant-tone" || findings[0].Severity != spectrum.Failure {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestEvaluateHandlesSilentTraceAndInvalidPolicy(t *testing.T) {
	findings, err := spectrum.Evaluate(spectrum.Report{}, spectrum.Policy{})
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].Code != "silent-trace" {
		t.Fatalf("findings = %+v", findings)
	}
	_, err = spectrum.Evaluate(spectrum.Report{TotalPower: 1}, spectrum.Policy{MinEntropy: 1.1, MaxDominantShare: 0.5})
	if err != spectrum.ErrInvalidPolicy {
		t.Fatalf("error = %v", err)
	}
}
