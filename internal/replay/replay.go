package replay

import (
	"context"
	"sort"
	"time"

	"github.com/zhangkui/specimen-flow-audit/internal/workflow"
)

type Item struct {
	Request    workflow.SubmitRequest
	OriginalAt time.Time
}
type Result struct {
	JobID      string
	StartedAt  time.Time
	FinishedAt time.Time
	Err        error
}

func Run(ctx context.Context, app *workflow.Coordinator, items []Item, scale float64) []Result {
	ordered := append([]Item(nil), items...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].OriginalAt.Before(ordered[j].OriginalAt) })
	if scale <= 0 {
		scale = 1
	}
	results := make([]Result, 0, len(ordered))
	var previous time.Time
	for _, item := range ordered {
		if !previous.IsZero() {
			delay := item.OriginalAt.Sub(previous)
			timer := time.NewTimer(time.Duration(float64(delay) / scale))
			select {
			case <-ctx.Done():
				timer.Stop()
				return results
			case <-timer.C:
			}
		}
		started := time.Now().UTC()
		_, err := app.Submit(ctx, item.Request)
		results = append(results, Result{JobID: item.Request.JobID, StartedAt: started, FinishedAt: time.Now().UTC(), Err: err})
		previous = item.OriginalAt
	}
	return results
}
