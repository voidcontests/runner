package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"runner/internal/app/handler"
)

const PORT = 21003

func main() {
	err := os.MkdirAll("files", 0755)
	if err != nil {
		slog.Error("failed to create `./files/` directory", slog.Any("error", err))
		return
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthcheck", handler.Healthcheck)
	mux.HandleFunc("POST /test", handler.TestSolution)

	addr := fmt.Sprintf(":%d", PORT)
	log.Printf("listening on %s", addr)

	http.ListenAndServe(addr, mux)
}
