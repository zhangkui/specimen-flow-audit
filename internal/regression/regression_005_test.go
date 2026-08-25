package regression

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/zhangkui/specimen-flow-audit/internal/calibration"
	"github.com/zhangkui/specimen-flow-audit/internal/ledger"
	"github.com/zhangkui/specimen-flow-audit/internal/qc"
	"github.com/zhangkui/specimen-flow-audit/internal/ruleset"
	"github.com/zhangkui/specimen-flow-audit/internal/station"
	"github.com/zhangkui/specimen-flow-audit/internal/workflow"
)

func TestBug05_ConcurrentSameJobKeepsOneConsistentLifecycle(t *testing.T) {
	for attempt := 0; attempt < 100; attempt++ {
		now := time.Date(2026, 8, 25, 15, 0, 0, 0, time.UTC)
		stations := station.New()
		policy := qc.Policy{MaxRMS: 1, MaxGapFraction: .9, MaxAbsolute: 2}
		for _, code := range []string{"HN01", "HN02"} {
			if err := stations.Upsert(station.Profile{Code: code, Network: "HN", Channels: []string{"HHZ"}, Calibration: calibration.Curve{Gain: 1, Min: -10, Max: 10}, Policy: policy, Trigger: 3}); err != nil {
				t.Fatal(err)
			}
		}
		rules := ruleset.New()
		for _, code := range []string{"HN01", "HN02"} {
			if err := rules.Publish(ruleset.Version{ID: "r1", Station: code, EffectiveAt: now.Add(-time.Hour), Policy: policy, Author: "operator"}); err != nil {
				t.Fatal(err)
			}
		}
		app := workflow.New(stations, rules, ledger.New())
		start := make(chan struct{})
		type result struct {
			completed workflow.Completed
			err       error
		}
		results := make(chan result, 2)
		var group sync.WaitGroup
		for _, code := range []string{"HN01", "HN02"} {
			group.Add(1)
			go func(stationCode string) {
				defer group.Done()
				<-start
				body := strings.NewReader(fmt.Sprintf(`{"station":"%s","start":"%s","sample_rate":2,"samples":[9,9,9,9]}`, stationCode, now.Format(time.RFC3339)))
				completed, err := app.Submit(context.Background(), workflow.SubmitRequest{JobID: "same-job", Format: "json", Station: stationCode, Start: now, SubmittedAt: now, Body: body})
				results <- result{completed: completed, err: err}
			}(code)
		}
		close(start)
		group.Wait()
		close(results)
		successes := []workflow.Completed{}
		for item := range results {
			if item.err == nil {
				successes = append(successes, item.completed)
			}
		}
		if len(successes) != 1 {
			t.Fatalf("attempt %d successes=%d", attempt, len(successes))
		}
		cards := app.Dashboard()
		if len(cards) != 1 || cards[0].Station != successes[0].Job.Station {
			t.Fatalf("attempt %d inconsistent lifecycle: success=%+v dashboard=%+v", attempt, successes[0].Job, cards)
		}
	}
}
