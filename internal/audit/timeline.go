package audit

import (
	"sort"
	"sync"
	"time"
)

type Entry struct {
	ID      string    `json:"id"`
	Subject string    `json:"subject"`
	Action  string    `json:"action"`
	Actor   string    `json:"actor"`
	At      time.Time `json:"at"`
	Detail  string    `json:"detail"`
}
type Timeline struct {
	mu      sync.RWMutex
	entries []Entry
}

func New() *Timeline { return &Timeline{} }
func (t *Timeline) Add(entry Entry) {
	t.mu.Lock()
	defer t.mu.Unlock()
	entry.At = entry.At.UTC()
	t.entries = append(t.entries, entry)
}
func (t *Timeline) ForSubject(subject string) []Entry {
	t.mu.RLock()
	defer t.mu.RUnlock()
	items := []Entry{}
	for _, entry := range t.entries {
		if entry.Subject == subject {
			items = append(items, entry)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].At.Before(items[j].At) })
	return items
}
