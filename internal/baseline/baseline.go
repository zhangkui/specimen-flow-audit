package baseline

import (
	"errors"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/zhangkui/specimen-flow-audit/internal/report"
)

var ErrNoHistory = errors.New("no baseline history")

type Point struct {
	At           time.Time `json:"at"`
	NoiseRMS     float64   `json:"noise_rms"`
	Completeness float64   `json:"completeness"`
}
type Model struct {
	Station          string  `json:"station"`
	Window           int     `json:"window"`
	MeanNoise        float64 `json:"mean_noise"`
	StdNoise         float64 `json:"std_noise"`
	MeanCompleteness float64 `json:"mean_completeness"`
	Points           int     `json:"points"`
}
type Registry struct {
	mu      sync.RWMutex
	history map[string][]Point
}

func New() *Registry { return &Registry{history: make(map[string][]Point)} }
func (r *Registry) Add(quality report.Quality) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.history[quality.Station] = append(r.history[quality.Station], Point{At: quality.ObservedAt, NoiseRMS: quality.NoiseRMS, Completeness: quality.Completeness})
}
func (r *Registry) Build(station string, window int) (Model, error) {
	r.mu.RLock()
	points := append([]Point(nil), r.history[station]...)
	r.mu.RUnlock()
	if len(points) == 0 {
		return Model{}, ErrNoHistory
	}
	sort.Slice(points, func(i, j int) bool { return points[i].At.Before(points[j].At) })
	if window > 0 && len(points) > window {
		points = points[len(points)-window:]
	}
	model := Model{Station: station, Window: window, Points: len(points)}
	for _, point := range points {
		model.MeanNoise += point.NoiseRMS
		model.MeanCompleteness += point.Completeness
	}
	model.MeanNoise /= float64(len(points))
	model.MeanCompleteness /= float64(len(points))
	for _, point := range points {
		delta := point.NoiseRMS - model.MeanNoise
		model.StdNoise += delta * delta
	}
	model.StdNoise = math.Sqrt(model.StdNoise / float64(len(points)))
	return model, nil
}
func Score(model Model, quality report.Quality) float64 {
	if model.StdNoise == 0 {
		return 0
	}
	return math.Abs(quality.NoiseRMS-model.MeanNoise) / model.StdNoise
}
