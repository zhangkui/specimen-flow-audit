package store

import (
	"errors"
	"sync"

	"github.com/zhangkui/specimen-flow-audit/internal/waveform"
)

var ErrNotFound = errors.New("trace not found")

type Store struct {
	mu     sync.RWMutex
	traces map[string]waveform.Trace
}

func New() *Store { return &Store{traces: make(map[string]waveform.Trace)} }
func (s *Store) Save(trace waveform.Trace) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.traces[trace.Station] = trace
}
func (s *Store) Load(station string) (waveform.Trace, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	trace, ok := s.traces[station]
	if !ok {
		return waveform.Trace{}, ErrNotFound
	}
	trace.Samples = append([]float64(nil), trace.Samples...)
	return trace, nil
}
