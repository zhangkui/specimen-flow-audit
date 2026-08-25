package runlog

import (
	"sort"
	"sync"
	"time"
)

type Entry struct {
	ID         string    `json:"id"`
	Station    string    `json:"station"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
	Status     string    `json:"status"`
	Error      string    `json:"error,omitempty"`
}
type Log struct {
	mu      sync.RWMutex
	entries map[string]Entry
}

func New() *Log { return &Log{entries: make(map[string]Entry)} }
func (l *Log) Start(entry Entry) {
	l.mu.Lock()
	defer l.mu.Unlock()
	entry.Status = "running"
	l.entries[entry.ID] = entry
}
func (l *Log) Finish(id string, finished time.Time, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	entry, ok := l.entries[id]
	if !ok {
		return
	}
	entry.FinishedAt = finished
	if err != nil {
		entry.Status = "failed"
		entry.Error = err.Error()
	} else {
		entry.Status = "completed"
	}
	l.entries[id] = entry
}
func (l *Log) List() []Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	items := make([]Entry, 0, len(l.entries))
	for _, entry := range l.entries {
		items = append(items, entry)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].StartedAt.Before(items[j].StartedAt) })
	return items
}
