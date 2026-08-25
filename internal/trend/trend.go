package trend

import (
	"errors"
	"math"
	"sort"
	"time"

	"github.com/zhangkui/specimen-flow-audit/internal/report"
)

var ErrMixedStation = errors.New("quality history contains multiple stations")

type State string

const (
	Insufficient State = "insufficient"
	Stable       State = "stable"
	Improving    State = "improving"
	Degrading    State = "degrading"
)

type Config struct {
	MinPoints                    int     `json:"min_points"`
	NoiseRisePerHour             float64 `json:"noise_rise_per_hour"`
	CompletenessFallPerHour      float64 `json:"completeness_fall_per_hour"`
	SignificantNoiseDeviation    float64 `json:"significant_noise_deviation"`
	SignificantCompletenessDelta float64 `json:"significant_completeness_delta"`
}

type Snapshot struct {
	Station                  string    `json:"station"`
	FirstObservedAt          time.Time `json:"first_observed_at"`
	LastObservedAt           time.Time `json:"last_observed_at"`
	Points                   int       `json:"points"`
	MeanNoise                float64   `json:"mean_noise"`
	MeanCompleteness         float64   `json:"mean_completeness"`
	NoiseSlopePerHour        float64   `json:"noise_slope_per_hour"`
	CompletenessSlopePerHour float64   `json:"completeness_slope_per_hour"`
	LatestNoiseDeviation     float64   `json:"latest_noise_deviation"`
	LatestCompletenessDelta  float64   `json:"latest_completeness_delta"`
	State                    State     `json:"state"`
	Reasons                  []string  `json:"reasons"`
}

func DefaultConfig() Config {
	return Config{
		MinPoints:                    3,
		NoiseRisePerHour:             0.15,
		CompletenessFallPerHour:      0.01,
		SignificantNoiseDeviation:    2,
		SignificantCompletenessDelta: 0.08,
	}
}

func Analyze(station string, qualities []report.Quality, config Config) (Snapshot, error) {
	config = normalizeConfig(config)
	items := append([]report.Quality(nil), qualities...)
	if station == "" && len(items) > 0 {
		station = items[0].Station
	}
	for _, quality := range items {
		if quality.Station != station {
			return Snapshot{}, ErrMixedStation
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ObservedAt.Before(items[j].ObservedAt) })
	snapshot := Snapshot{Station: station, Points: len(items), State: Insufficient}
	if len(items) == 0 {
		snapshot.Reasons = []string{"没有可用的历史质量报告"}
		return snapshot, nil
	}
	snapshot.FirstObservedAt = items[0].ObservedAt
	snapshot.LastObservedAt = items[len(items)-1].ObservedAt
	noiseValues := make([]float64, len(items))
	completenessValues := make([]float64, len(items))
	hours := make([]float64, len(items))
	for index, quality := range items {
		noiseValues[index] = quality.NoiseRMS
		completenessValues[index] = quality.Completeness
		hours[index] = quality.ObservedAt.Sub(snapshot.FirstObservedAt).Hours()
		snapshot.MeanNoise += quality.NoiseRMS
		snapshot.MeanCompleteness += quality.Completeness
	}
	snapshot.MeanNoise /= float64(len(items))
	snapshot.MeanCompleteness /= float64(len(items))
	latest := items[len(items)-1]
	snapshot.LatestNoiseDeviation = standardized(latest.NoiseRMS, snapshot.MeanNoise, noiseValues)
	snapshot.LatestCompletenessDelta = latest.Completeness - snapshot.MeanCompleteness
	if len(items) < config.MinPoints {
		snapshot.Reasons = []string{"历史报告数量不足，无法判断趋势"}
		return snapshot, nil
	}
	noiseSlope, hasTimeRange := slope(hours, noiseValues)
	if !hasTimeRange {
		snapshot.Reasons = []string{"历史报告时间相同，无法判断趋势"}
		return snapshot, nil
	}
	completenessSlope, _ := slope(hours, completenessValues)
	snapshot.NoiseSlopePerHour = noiseSlope
	snapshot.CompletenessSlopePerHour = completenessSlope
	return classify(snapshot, config), nil
}

func normalizeConfig(config Config) Config {
	defaults := DefaultConfig()
	if config.MinPoints <= 0 {
		config.MinPoints = defaults.MinPoints
	}
	if config.NoiseRisePerHour <= 0 {
		config.NoiseRisePerHour = defaults.NoiseRisePerHour
	}
	if config.CompletenessFallPerHour <= 0 {
		config.CompletenessFallPerHour = defaults.CompletenessFallPerHour
	}
	if config.SignificantNoiseDeviation <= 0 {
		config.SignificantNoiseDeviation = defaults.SignificantNoiseDeviation
	}
	if config.SignificantCompletenessDelta <= 0 {
		config.SignificantCompletenessDelta = defaults.SignificantCompletenessDelta
	}
	return config
}

func slope(x, y []float64) (float64, bool) {
	meanX, meanY := 0.0, 0.0
	for index := range x {
		meanX += x[index]
		meanY += y[index]
	}
	meanX /= float64(len(x))
	meanY /= float64(len(y))
	numerator, denominator := 0.0, 0.0
	for index := range x {
		deltaX := x[index] - meanX
		numerator += deltaX * (y[index] - meanY)
		denominator += deltaX * deltaX
	}
	if denominator == 0 {
		return 0, false
	}
	return numerator / denominator, true
}

func standardized(value, mean float64, values []float64) float64 {
	variance := 0.0
	for _, item := range values {
		delta := item - mean
		variance += delta * delta
	}
	if variance == 0 {
		return 0
	}
	return (value - mean) / math.Sqrt(variance/float64(len(values)))
}

func classify(snapshot Snapshot, config Config) Snapshot {
	if snapshot.NoiseSlopePerHour >= config.NoiseRisePerHour {
		snapshot.State = Degrading
		snapshot.Reasons = append(snapshot.Reasons, "RMS 噪声持续上升")
	}
	if snapshot.CompletenessSlopePerHour <= -config.CompletenessFallPerHour {
		snapshot.State = Degrading
		snapshot.Reasons = append(snapshot.Reasons, "数据完整率持续下降")
	}
	if math.Abs(snapshot.LatestNoiseDeviation) >= config.SignificantNoiseDeviation && snapshot.LatestNoiseDeviation > 0 {
		snapshot.Reasons = append(snapshot.Reasons, "最新噪声显著高于历史均值")
	}
	if snapshot.LatestCompletenessDelta <= -config.SignificantCompletenessDelta {
		snapshot.Reasons = append(snapshot.Reasons, "最新完整率显著低于历史均值")
	}
	if snapshot.State == Degrading {
		return snapshot
	}
	if snapshot.NoiseSlopePerHour <= -config.NoiseRisePerHour || snapshot.CompletenessSlopePerHour >= config.CompletenessFallPerHour {
		snapshot.State = Improving
		return snapshot
	}
	snapshot.State = Stable
	return snapshot
}
