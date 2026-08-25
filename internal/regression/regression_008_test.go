package regression

import (
	"testing"
	"time"

	"github.com/zhangkui/specimen-flow-audit/internal/ruleset"

	"github.com/zhangkui/specimen-flow-audit/internal/qc"
)

func TestBug08_DuplicateVersionIDCannotCreatePolicyFork(t *testing.T) {
	registry := ruleset.New()
	at := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)
	first := ruleset.Version{ID: "r-8", Station: "HN01", EffectiveAt: at, Author: "a", Policy: qc.Policy{MaxRMS: 1, MaxGapFraction: .2, MaxAbsolute: 2}}
	second := first
	second.Author = "b"
	second.Policy.MaxRMS = 9
	if err := registry.Publish(first); err != nil {
		t.Fatal(err)
	}
	if err := registry.Publish(second); err == nil {
		t.Fatal("duplicate version ID must not create two policy histories")
	}
	if len(registry.History("HN01")) != 1 {
		t.Fatalf("history contains a duplicate identity: %+v", registry.History("HN01"))
	}
}
