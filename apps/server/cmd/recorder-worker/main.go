package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"robot-center/apps/server/internal/config"
	"robot-center/apps/server/internal/recording"
)

func main() {
	cfg := config.LoadRecorderWorkerConfig()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	worker := recording.NewWorker(cfg)
	go worker.Run(ctx)

	mux := http.NewServeMux()
	started := time.Now().UTC()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":     "ok",
			"service":    "recorder-worker",
			"startedAt":  started.Format(time.RFC3339),
			"subscriber": worker.SubscriberStatus(),
		})
	})

	server := &http.Server{
		Addr:              cfg.HTTPAddress,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("recorder-worker health endpoint listening on %s", cfg.HTTPAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("recorder-worker health endpoint failed: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("recorder-worker shutdown failed: %v", err)
	}
}
