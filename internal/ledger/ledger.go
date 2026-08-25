package ledger

import (
	"sort"
	"sync"
	"time"
)

type State string

const (
	Received   State = "received"
	Processing State = "processing"
	Completed  State = "completed"
	Failed     State = "failed"
)

type Job struct {
	ID          string    `json:"id"`
	Station     string    `json:"station"`
	State       State     `json:"state"`
	SubmittedAt time.Time `json:"submitted_at"`
	StartedAt   time.Time `json:"started_at,omitempty"`
	FinishedAt  time.Time `json:"finished_at,omitempty"`
	RuleVersion string    `json:"rule_version,omitempty"`
	Error       string    `json:"error,omitempty"`
}
type Ledger struct {
	mu   sync.RWMutex
	jobs map[string]Job
}

func New() *Ledger { return &Ledger{jobs: make(map[string]Job)} }
func (l *Ledger) Submit(job Job) {
	l.mu.Lock()
	defer l.mu.Unlock()
	job.State = Received
	l.jobs[job.ID] = job
}
func (l *Ledger) Start(id, version string, at time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	job, ok := l.jobs[id]
	if !ok || job.State != Received {
		return false
	}
	job.State = Processing
	job.StartedAt = at.UTC()
	job.RuleVersion = version
	l.jobs[id] = job
	return true
}
func (l *Ledger) Finish(id string, at time.Time, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	job, ok := l.jobs[id]
	if !ok {
		return
	}
	job.FinishedAt = at.UTC()
	if err == nil {
		job.State = Completed
	} else {
		job.State = Failed
		job.Error = err.Error()
	}
	l.jobs[id] = job
}
func (l *Ledger) Get(id string) (Job, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	job, ok := l.jobs[id]
	return job, ok
}
func (l *Ledger) ByStation(station string) []Job {
	l.mu.RLock()
	defer l.mu.RUnlock()
	items := []Job{}
	for _, job := range l.jobs {
		if job.Station == station {
			items = append(items, job)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].SubmittedAt.Before(items[j].SubmittedAt) })
	return items
}
