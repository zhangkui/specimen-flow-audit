package sla

import (
	"github.com/zhangkui/specimen-flow-audit/internal/incident"
	"time"
)

type Target struct {
	AcknowledgeWithin time.Duration
	ResolveWithin     time.Duration
}
type Result struct {
	IncidentID         string        `json:"incident_id"`
	AcknowledgedIn     time.Duration `json:"acknowledged_in"`
	ResolvedIn         time.Duration `json:"resolved_in"`
	AcknowledgementMet bool          `json:"acknowledgement_met"`
	ResolutionMet      bool          `json:"resolution_met"`
}

func Evaluate(item incident.Incident, target Target, now time.Time) Result {
	result := Result{IncidentID: item.ID}
	ackEnd := item.AcknowledgedAt
	if ackEnd.IsZero() {
		ackEnd = now
	}
	resolveEnd := item.ResolvedAt
	if resolveEnd.IsZero() {
		resolveEnd = now
	}
	result.AcknowledgedIn = ackEnd.Sub(item.OpenedAt)
	result.ResolvedIn = resolveEnd.Sub(item.OpenedAt)
	result.AcknowledgementMet = result.AcknowledgedIn <= target.AcknowledgeWithin
	result.ResolutionMet = result.ResolvedIn <= target.ResolveWithin
	return result
}
