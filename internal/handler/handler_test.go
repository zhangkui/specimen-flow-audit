package handler_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/zhangkui/specimen-flow-audit/internal/handler"
	"github.com/zhangkui/specimen-flow-audit/internal/service"
)

func TestAnalyzeTraceReturnsQualityReport(t *testing.T) {
	server := handler.New(service.New())
	body := `{"station":"HN01","start":"2026-08-24T00:00:00Z","sample_rate":4,"samples":[1,2,0,0,3,4]}`
	request := httptest.NewRequest(http.MethodPost, "/traces/analyze", bytes.NewBufferString(body))
	recorder := httptest.NewRecorder()

	server.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"gap_count":1`) {
		t.Fatalf("expected one missing-data gap, body = %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"completeness":0.666`) {
		t.Fatalf("expected completeness derived from missing samples, body = %s", recorder.Body.String())
	}
}
