package spectrum_test

import (
	"math"
	"testing"

	"github.com/zhangkui/specimen-flow-audit/internal/spectrum"
)

func TestCompareReturnsOneForMatchingFingerprint(t *testing.T) {
	report := spectrum.Report{Station: "HN01", FrequencyResolution: 0.1, DominantFrequency: 0.5, SpectralEntropy: 0.4, Bands: []spectrum.BandPower{{Name: "microseism", Share: 0.6}, {Name: "high-frequency", Share: 0.2}}}
	comparison := spectrum.Compare(report, report)
	if comparison.Similarity != 1 || comparison.BandShareDistance != 0 {
		t.Fatalf("comparison = %+v", comparison)
	}
}

func TestCompareIncludesMissingBandAndFrequencyDifference(t *testing.T) {
	left := spectrum.Report{Station: "HN01", FrequencyResolution: 0.1, DominantFrequency: 0.2, SpectralEntropy: 0.2, Bands: []spectrum.BandPower{{Name: "microseism", Share: 0.8}}}
	right := spectrum.Report{Station: "HN02", FrequencyResolution: 0.1, DominantFrequency: 0.5, SpectralEntropy: 0.5, Bands: []spectrum.BandPower{{Name: "high-frequency", Share: 0.8}}}
	comparison := spectrum.Compare(left, right)
	if math.Abs(comparison.BandShareDistance-0.8) > 0.000001 || comparison.Similarity >= 0.25 {
		t.Fatalf("comparison = %+v", comparison)
	}
}
