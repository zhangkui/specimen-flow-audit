package regression

import (
	"testing"
	"time"

	"github.com/zhangkui/specimen-flow-audit/internal/incident"

	"github.com/zhangkui/specimen-flow-audit/internal/alert"
)

func TestBug06_ResolveRequiresAcknowledgement(t *testing.T) {
	register := incident.New()
	opened := time.Date(2026, 8, 25, 11, 0, 0, 0, time.UTC)
	if err := register.Open(incident.Incident{ID: "incident-6", Station: "HN01", Severity: alert.Critical, OpenedAt: opened}); err != nil {
		t.Fatal(err)
	}
	if _, err := register.Resolve("incident-6", "closed", opened.Add(time.Minute)); err == nil {
		t.Fatal("an unacknowledged incident must not resolve")
	}
}
