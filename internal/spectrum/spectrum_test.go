package spectrum_test

import (
	"math"
	"testing"
	"time"

	"github.com/zhangkui/specimen-flow-audit/internal/spectrum"
	"github.com/zhangkui/specimen-flow-audit/internal/waveform"
)

func TestAnalyzeFindsDominantFrequencyAndBandShare(t *testing.T) {
	trace := sineTrace("HN01", 20, 1.0, 200)
	report, err := spectrum.Analyze(trace, spectrum.Config{Bands: []spectrum.Band{{Name: "target", LowHz: 0.9, HighHz: 1.1}}, MaxSamples: 512})
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(report.DominantFrequency-1) > report.FrequencyResolution {
		t.Fatalf("dominant frequency = %.3f, resolution = %.3f", report.DominantFrequency, report.FrequencyResolution)
	}
	if len(report.Bands) != 1 || report.Bands[0].Share < 0.95 {
		t.Fatalf("band power = %+v", report.Bands)
	}
	if report.SpectralEntropy <= 0 || report.SpectralEntropy >= 0.5 {
		t.Fatalf("spectral entropy = %f", report.SpectralEntropy)
	}
}

func TestAnalyzeDecimatesLargeTraceAndRejectsInvalidBand(t *testing.T) {
	trace := sineTrace("HN01", 100, 5.0, 500)
	report, err := spectrum.Analyze(trace, spectrum.Config{Bands: []spectrum.Band{{Name: "signal", LowHz: 4, HighHz: 6}}, MaxSamples: 100})
	if err != nil {
		t.Fatal(err)
	}
	if report.AnalyzedSamples != 100 || report.SampleRate != 20 {
		t.Fatalf("decimated report = %+v", report)
	}
	_, err = spectrum.Analyze(sineTrace("HN01", 10, 1, 20), spectrum.Config{Bands: []spectrum.Band{{Name: "beyond-nyquist", LowHz: 1, HighHz: 6}}})
	if err != spectrum.ErrInvalidBand {
		t.Fatalf("error = %v", err)
	}
}

func TestAnalyzeRequiresEnoughSamples(t *testing.T) {
	_, err := spectrum.Analyze(waveform.Trace{Station: "HN01", Start: time.Date(2026, 8, 24, 0, 0, 0, 0, time.UTC), SampleRate: 10, Samples: []float64{1, 2, 3}}, spectrum.Config{})
	if err != spectrum.ErrTooFewSamples {
		t.Fatalf("error = %v", err)
	}
}

func sineTrace(station string, sampleRate, frequency float64, count int) waveform.Trace {
	samples := make([]float64, count)
	for index := range samples {
		samples[index] = math.Sin(2 * math.Pi * frequency * float64(index) / sampleRate)
	}
	return waveform.Trace{Station: station, Start: time.Date(2026, 8, 24, 0, 0, 0, 0, time.UTC), SampleRate: sampleRate, Samples: samples}
}
