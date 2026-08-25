package notify

import (
	"sort"
	"sync"
	"time"

	"github.com/zhangkui/specimen-flow-audit/internal/alert"
)

type Message struct {
	ID        string    `json:"id"`
	Station   string    `json:"station"`
	Recipient string    `json:"recipient"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	SentAt    time.Time `json:"sent_at,omitempty"`
	Attempts  int       `json:"attempts"`
}
type Outbox struct {
	mu       sync.RWMutex
	messages map[string]Message
}

func New() *Outbox { return &Outbox{messages: make(map[string]Message)} }
func (o *Outbox) Queue(id, recipient string, item alert.Alert, at time.Time) {
	o.mu.Lock()
	defer o.mu.Unlock()
	// Idempotent: a repeat delivery of the same job must not clobber a message
	// that may already be queued or sent.
	if _, ok := o.messages[id]; ok {
		return
	}
	o.messages[id] = Message{ID: id, Station: item.Station, Recipient: recipient, Subject: item.Code, Body: item.Detail, CreatedAt: at.UTC()}
}
func (o *Outbox) MarkSent(id string, at time.Time) bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	message, ok := o.messages[id]
	if !ok {
		return false
	}
	message.Attempts++
	message.SentAt = at.UTC()
	o.messages[id] = message
	return true
}
func (o *Outbox) Pending() []Message {
	o.mu.RLock()
	defer o.mu.RUnlock()
	items := []Message{}
	for _, message := range o.messages {
		if message.SentAt.IsZero() {
			items = append(items, message)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.Before(items[j].CreatedAt) })
	return items
}
