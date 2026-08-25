package review

import (
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/zhangkui/specimen-flow-audit/internal/alert"
)

var (
	ErrNotFound  = errors.New("review task not found")
	ErrCompleted = errors.New("review task completed")
)

type State string

const (
	Pending  State = "pending"
	Assigned State = "assigned"
	Closed   State = "closed"
)

type Task struct {
	ID         string      `json:"id"`
	JobID      string      `json:"job_id"`
	Station    string      `json:"station"`
	Alert      alert.Alert `json:"alert"`
	State      State       `json:"state"`
	CreatedAt  time.Time   `json:"created_at"`
	Assignee   string      `json:"assignee,omitempty"`
	Resolution string      `json:"resolution,omitempty"`
	ClosedAt   time.Time   `json:"closed_at,omitempty"`
}
type Queue struct {
	mu    sync.RWMutex
	tasks map[string]Task
}

func New() *Queue { return &Queue{tasks: make(map[string]Task)} }
func (q *Queue) Create(task Task) error {
	if task.ID == "" || task.JobID == "" || task.Station == "" || task.CreatedAt.IsZero() {
		return errors.New("invalid review task")
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	task.State = Pending
	q.tasks[task.ID] = task
	return nil
}
func (q *Queue) Assign(id, assignee string) (Task, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	task, ok := q.tasks[id]
	if !ok {
		return Task{}, ErrNotFound
	}
	if task.State == Closed {
		return Task{}, ErrCompleted
	}
	task.State = Assigned
	task.Assignee = assignee
	q.tasks[id] = task
	return task, nil
}
func (q *Queue) Close(id, resolution string, at time.Time) (Task, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	task, ok := q.tasks[id]
	if !ok {
		return Task{}, ErrNotFound
	}
	if task.State == Closed {
		return task, nil
	}
	task.State = Closed
	task.Resolution = resolution
	task.ClosedAt = at.UTC()
	q.tasks[id] = task
	return task, nil
}
func (q *Queue) Open(station string) []Task {
	q.mu.RLock()
	defer q.mu.RUnlock()
	items := []Task{}
	for _, task := range q.tasks {
		if task.Station == station && task.State != Closed {
			items = append(items, task)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.Before(items[j].CreatedAt) })
	return items
}
