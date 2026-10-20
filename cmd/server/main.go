package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/FelipeMadson/go-distributed-telemetry-pipeline/pkg/otlp"
	"github.com/FelipeMadson/go-distributed-telemetry-pipeline/pkg/ratelimit"
	"github.com/FelipeMadson/go-distributed-telemetry-pipeline/pkg/telemetry"
	"github.com/FelipeMadson/go-distributed-telemetry-pipeline/pkg/worker"
)

type Server struct {
	pool        *worker.WorkerPool
	ringBuffer  *telemetry.RingBuffer
	rateLimiter *ratelimit.TokenBucket
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ringBuffer := telemetry.NewRingBuffer(10000)
	pool := worker.NewWorkerPool(8, 5000, ringBuffer)
	rateLimiter := ratelimit.NewTokenBucket(1000, 500)

	pool.Start()

	srvState := &Server{
		pool:        pool,
		ringBuffer:  ringBuffer,
		rateLimiter: rateLimiter,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"service": "go-distributed-telemetry-pipeline",
			"status":  "healthy",
			"author":  "Felipe Madison (@FelipeMadson)",
			"version": "1.0.0",
		})
	})

	mux.HandleFunc("POST /v1/metrics", func(w http.ResponseWriter, r *http.Request) {
		if !srvState.rateLimiter.Allow() {
			http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
			return
		}

		var batch []telemetry.MetricRecord
		if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"payload JSON inválido: %s"}`, err.Error()), http.StatusBadRequest)
			return
		}

		accepted := 0
		dropped := 0
		for _, rec := range batch {
			if err := rec.Validate(); err != nil {
				continue
			}
			if srvState.pool.Submit(rec) {
				accepted++
			} else {
				dropped++
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "accepted",
			"accepted": accepted,
			"dropped":  dropped,
		})
	})

	mux.HandleFunc("GET /v1/stats", func(w http.ResponseWriter, r *http.Request) {
		samples := srvState.ringBuffer.Snapshot()
		latencyStats := telemetry.CalculatePercentiles(samples)
		pipelineStats := srvState.pool.Stats()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"pipeline": pipelineStats,
			"latency":  latencyStats,
			"tokens":   srvState.rateLimiter.AvailableTokens(),
		})
	})

	mux.HandleFunc("GET /v1/otlp/export", func(w http.ResponseWriter, r *http.Request) {
		samples := srvState.ringBuffer.Snapshot()
		records := make([]telemetry.MetricRecord, 0, len(samples))
		for _, v := range samples {
			records = append(records, telemetry.MetricRecord{
				Name:      "http.request.duration",
				Type:      telemetry.MetricTypeHistogram,
				Value:     v,
				Timestamp: time.Now().UTC(),
			})
		}

		payload, err := otlp.ConvertToOTLPPayload("telemetry-ingress", records)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(payload)
	})

	httpServer := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("[TelemetryPipeline] Servidor rodando na porta %s...", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erro no servidor HTTP: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[TelemetryPipeline] Sinal de encerramento recebido. Desligando graciosamente...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("Erro ao encerrar servidor HTTP: %v", err)
	}

	if err := pool.Stop(3 * time.Second); err != nil {
		log.Printf("Erro ao encerrar worker pool: %v", err)
	}

	log.Println("[TelemetryPipeline] Servidor encerrado.")
}
