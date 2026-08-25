package incident

import (
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/zhangkui/specimen-flow-audit/internal/alert"
)

var (
	ErrNotFound       = errors.New("incident not found")
	ErrClosed         = errors.New("incident is closed")
	ErrNotAcknowledged = errors.New("incident must be acknowledged before it can be resolved")
)

type State string

const (
	Open         State = "open"
	Acknowledged State = "acknowledged"
	Resolved     State = "resolved"
)

type Incident struct {
	ID             string      `json:"id"`
	Station        string      `json:"station"`
	State          State       `json:"state"`
	Severity       alert.Level `json:"severity"`
	OpenedAt       time.Time   `json:"opened_at"`
	AcknowledgedAt time.Time   `json:"acknowledged_at,omitempty"`
	ResolvedAt     time.Time   `json:"resolved_at,omitempty"`
	Assignee       string      `json:"assignee,omitempty"`
	Notes          []string    `json:"notes"`
}
type Register struct {
	mu      sync.RWMutex
	entries map[string]Incident
}

func New() *Register { return &Register{entries: make(map[string]Incident)} }
func (r *Register) Open(item Incident) error {
	if item.ID == "" || item.Station == "" || item.OpenedAt.IsZero() {
		return errors.New("invalid incident")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	item.State = Open
	item.Notes = append([]string(nil), item.Notes...)
	r.entries[item.ID] = item
	return nil
}
func (r *Register) Acknowledge(id, assignee, note string, at time.Time) (Incident, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	item, ok := r.entries[id]
	if !ok {
		return Incident{}, ErrNotFound
	}
	if item.State == Resolved {
		return Incident{}, ErrClosed
	}
	item.State = Acknowledged
	item.Assignee = assignee
	item.AcknowledgedAt = at.UTC()
	if note != "" {
		item.Notes = append(item.Notes, note)
	}
	r.entries[id] = item
	return item, nil
}
func (r *Register) Resolve(id, note string, at time.Time) (Incident, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	item, ok := r.entries[id]
	if !ok {
		return Incident{}, ErrNotFound
	}
	if item.State == Resolved {
		return item, nil
	}
	if item.State != Acknowledged {
		return Incident{}, ErrNotAcknowledged
	}
	item.State = Resolved
	item.ResolvedAt = at.UTC()
	if note != "" {
		item.Notes = append(item.Notes, note)
	}
	r.entries[id] = item
	return item, nil
}
func (r *Register) Get(id string) (Incident, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, ok := r.entries[id]
	if !ok {
		return Incident{}, ErrNotFound
	}
	item.Notes = append([]string(nil), item.Notes...)
	return item, nil
}
func (r *Register) OpenForStation(station string) []Incident {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := []Incident{}
	for _, item := range r.entries {
		if item.Station == station && item.State != Resolved {
			item.Notes = append([]string(nil), item.Notes...)
			items = append(items, item)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].OpenedAt.Before(items[j].OpenedAt) })
	return items
}
