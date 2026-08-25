package batch

import (
	"context"
	"github.com/zhangkui/specimen-flow-audit/internal/report"
	"github.com/zhangkui/specimen-flow-audit/internal/service"
	"github.com/zhangkui/specimen-flow-audit/internal/waveform"
	"sync"
)

type Result struct {
	Index  int
	Report report.Quality
	Err    error
}

func Analyze(ctx context.Context, app *service.Service, traces []waveform.Trace, workers int) []Result {
	if workers < 1 {
		workers = 1
	}
	jobs := make(chan int)
	results := make(chan Result, len(traces))
	var group sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case index, ok := <-jobs:
					if !ok {
						return
					}
					quality, err := app.Analyze(traces[index])
					results <- Result{Index: index, Report: quality, Err: err}
				}
			}
		}()
	}
	go func() {
		defer close(results)
		defer group.Wait()
		for index := range traces {
			select {
			case jobs <- index:
			case <-ctx.Done():
				close(jobs)
				return
			}
		}
		close(jobs)
	}()
	out := []Result{}
	for result := range results {
		out = append(out, result)
	}
	return out
}
