package maintenance

import (
	"errors"
	"sort"
	"sync"
	"time"
)

var ErrInvalidWindow = errors.New("invalid maintenance window")

type Window struct {
	ID      string    `json:"id"`
	Station string    `json:"station"`
	Start   time.Time `json:"start"`
	End     time.Time `json:"end"`
	Reason  string    `json:"reason"`
	Planned bool      `json:"planned"`
}

func (w Window) Validate() error {
	if w.ID == "" || w.Station == "" || w.Start.IsZero() || !w.End.After(w.Start) || w.Reason == "" {
		return ErrInvalidWindow
	}
	return nil
}
func (w Window) Covers(at time.Time) bool { return !at.Before(w.Start) && at.Before(w.End) }

type Calendar struct {
	mu      sync.RWMutex
	windows map[string][]Window
}

func New() *Calendar { return &Calendar{windows: make(map[string][]Window)} }
func (c *Calendar) Schedule(window Window) error {
	if err := window.Validate(); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	items := append(c.windows[window.Station], window)
	sort.Slice(items, func(i, j int) bool { return items[i].Start.Before(items[j].Start) })
	c.windows[window.Station] = items
	return nil
}
func (c *Calendar) At(station string, at time.Time) (Window, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, window := range c.windows[station] {
		if window.Covers(at) {
			return window, true
		}
	}
	return Window{}, false
}
func (c *Calendar) Between(station string, start, end time.Time) []Window {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := []Window{}
	for _, window := range c.windows[station] {
		if window.Start.Before(end) && window.End.After(start) {
			result = append(result, window)
		}
	}
	return result
}
