package main

import (
	"log"
	"net/http"
	"os"

	"github.com/zhangkui/specimen-flow-audit/internal/handler"
	"github.com/zhangkui/specimen-flow-audit/internal/service"
)

func main() {
	addr := os.Getenv("SPECIMEN_FLOW_AUDIT_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	log.Printf("seismic waveform quality service listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, handler.New(service.New())))
}
