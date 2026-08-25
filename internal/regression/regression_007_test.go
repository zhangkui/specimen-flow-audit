package regression

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/zhangkui/specimen-flow-audit/internal/importer"
)

type cancellingReader struct {
	ctx    context.Context
	cancel context.CancelFunc
	body   *strings.Reader
	fired  bool
}

func (r *cancellingReader) Read(p []byte) (int, error) {
	if !r.fired {
		r.fired = true
		r.cancel()
	}
	return r.body.Read(p)
}

func TestBug07_CancellationDuringDecodeStopsImport(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	reader := &cancellingReader{ctx: ctx, cancel: cancel, body: strings.NewReader(`{"station":"HN01","start":"2026-08-25T14:00:00Z","sample_rate":2,"samples":[1,2,3,4]}`)}
	result, err := importer.Read(ctx, importer.Request{ID: "job-7", Format: importer.JSON, Station: "HN01", Start: time.Date(2026, 8, 25, 14, 0, 0, 0, time.UTC), Body: reader, SubmittedAt: time.Now()})
	if err == nil || result.Trace.Station != "" {
		t.Fatalf("cancelled import completed: result=%+v err=%v", result, err)
	}
}
