package handler

import (
	"encoding/json"
	"net/http"

	"github.com/zhangkui/specimen-flow-audit/internal/service"
	"github.com/zhangkui/specimen-flow-audit/internal/waveform"
)

func New(app *service.Service) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /traces/analyze", func(writer http.ResponseWriter, request *http.Request) {
		defer request.Body.Close()
		trace, err := waveform.Decode(request.Body)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		quality, err := app.Analyze(trace)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(quality)
	})
	mux.HandleFunc("GET /traces/{station}", func(writer http.ResponseWriter, request *http.Request) {
		quality, err := app.Latest(request.PathValue("station"))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(writer).Encode(quality)
	})
	return mux
}
