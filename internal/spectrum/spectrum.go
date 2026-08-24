package spectrum

import (
	"errors"
	"math"
	"sort"

	"github.com/zhangkui/specimen-flow-audit/internal/waveform"
)

var (
	ErrInvalidBand   = errors.New("invalid spectral band")
	ErrTooFewSamples = errors.New("at least four samples are required for spectral analysis")
)

type Band struct {
	Name   string  `json:"name"`
	LowHz  float64 `json:"low_hz"`
	HighHz float64 `json:"high_hz"`
}

type Config struct {
	Bands        []Band  `json:"bands"`
	MaxSamples   int     `json:"max_samples"`
	MinFrequency float64 `json:"min_frequency"`
}

type BandPower struct {
	Name   string  `json:"name"`
	LowHz  float64 `json:"low_hz"`
	HighHz float64 `json:"high_hz"`
	Power  float64 `json:"power"`
	Share  float64 `json:"share"`
}

type Report struct {
	Station             string      `json:"station"`
	SampleRate          float64     `json:"sample_rate"`
	OriginalSamples     int         `json:"original_samples"`
	AnalyzedSamples     int         `json:"analyzed_samples"`
	FrequencyResolution float64     `json:"frequency_resolution"`
	TotalPower          float64     `json:"total_power"`
	DominantFrequency   float64     `json:"dominant_frequency"`
	DominantPower       float64     `json:"dominant_power"`
	SpectralEntropy     float64     `json:"spectral_entropy"`
	Bands               []BandPower `json:"bands"`
}

func DefaultConfig() Config {
	return Config{
		Bands: []Band{
			{Name: "microseism", LowHz: 0.05, HighHz: 0.30},
			{Name: "local-motion", LowHz: 0.30, HighHz: 0.80},
			{Name: "high-frequency", LowHz: 0.80, HighHz: 0.95},
		},
		MaxSamples:   2048,
		MinFrequency: 0.01,
	}
}

func Analyze(trace waveform.Trace, config Config) (Report, error) {
	if err := waveform.Validate(trace); err != nil {
		return Report{}, err
	}
	if len(trace.Samples) < 4 {
		return Report{}, ErrTooFewSamples
	}
	config = normalizeConfig(config)
	samples, sampleRate := decimate(trace.Samples, trace.SampleRate, config.MaxSamples)
	if err := validateBands(config.Bands, sampleRate/2); err != nil {
		return Report{}, err
	}
	report := Report{
		Station:             trace.Station,
		SampleRate:          sampleRate,
		OriginalSamples:     len(trace.Samples),
		AnalyzedSamples:     len(samples),
		FrequencyResolution: sampleRate / float64(len(samples)),
		Bands:               makeBandPowers(config.Bands),
	}
	powers := periodogram(samples)
	for bin, power := range powers {
		frequency := float64(bin) * report.FrequencyResolution
		if bin == 0 {
			continue
		}
		report.TotalPower += power
		if frequency >= config.MinFrequency && power > report.DominantPower {
			report.DominantFrequency = frequency
			report.DominantPower = power
		}
		for index := range report.Bands {
			band := &report.Bands[index]
			if frequency >= band.LowHz && frequency < band.HighHz {
				band.Power += power
			}
		}
	}
	for index := range report.Bands {
		if report.TotalPower > 0 {
			report.Bands[index].Share = report.Bands[index].Power / report.TotalPower
		}
	}
	report.SpectralEntropy = entropy(powers[1:])
	return report, nil
}

func normalizeConfig(config Config) Config {
	defaults := DefaultConfig()
	if len(config.Bands) == 0 {
		config.Bands = defaults.Bands
	}
	if config.MaxSamples <= 0 {
		config.MaxSamples = defaults.MaxSamples
	}
	if config.MinFrequency < 0 {
		config.MinFrequency = 0
	}
	return config
}

func validateBands(bands []Band, nyquist float64) error {
	seen := make(map[string]struct{}, len(bands))
	for _, band := range bands {
		if band.Name == "" || band.LowHz < 0 || band.HighHz <= band.LowHz || band.HighHz > nyquist {
			return ErrInvalidBand
		}
		if _, exists := seen[band.Name]; exists {
			return ErrInvalidBand
		}
		seen[band.Name] = struct{}{}
	}
	return nil
}

func decimate(samples []float64, sampleRate float64, maxSamples int) ([]float64, float64) {
	if len(samples) <= maxSamples {
		return append([]float64(nil), samples...), sampleRate
	}
	step := int(math.Ceil(float64(len(samples)) / float64(maxSamples)))
	result := make([]float64, 0, (len(samples)+step-1)/step)
	for index := 0; index < len(samples); index += step {
		result = append(result, samples[index])
	}
	return result, sampleRate / float64(step)
}

func periodogram(samples []float64) []float64 {
	count := len(samples)
	bins := count/2 + 1
	powers := make([]float64, bins)
	for bin := 0; bin < bins; bin++ {
		realPart, imaginaryPart := 0.0, 0.0
		for index, sample := range samples {
			angle := 2 * math.Pi * float64(bin*index) / float64(count)
			realPart += sample * math.Cos(angle)
			imaginaryPart -= sample * math.Sin(angle)
		}
		powers[bin] = (realPart*realPart + imaginaryPart*imaginaryPart) / float64(count*count)
	}
	return powers
}

func entropy(powers []float64) float64 {
	if len(powers) < 2 {
		return 0
	}
	total := 0.0
	for _, power := range powers {
		total += power
	}
	if total == 0 {
		return 0
	}
	value := 0.0
	for _, power := range powers {
		if power == 0 {
			continue
		}
		probability := power / total
		value -= probability * math.Log(probability)
	}
	return value / math.Log(float64(len(powers)))
}

func makeBandPowers(bands []Band) []BandPower {
	result := make([]BandPower, len(bands))
	for index, band := range bands {
		result[index] = BandPower{Name: band.Name, LowHz: band.LowHz, HighHz: band.HighHz}
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].LowHz < result[j].LowHz })
	return result
}
