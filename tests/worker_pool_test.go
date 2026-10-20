package tests

import (
	"testing"
	"time"

	"github.com/FelipeMadson/go-distributed-telemetry-pipeline/pkg/telemetry"
	"github.com/FelipeMadson/go-distributed-telemetry-pipeline/pkg/worker"
)

func TestWorkerPool_ThroughputAndGracefulShutdown(t *testing.T) {
	ringBuffer := telemetry.NewRingBuffer(100)
	pool := worker.NewWorkerPool(4, 50, ringBuffer)
	pool.Start()

	totalItems := 30
	for i := 0; i < totalItems; i++ {
		submitted := pool.Submit(telemetry.MetricRecord{
			Name:      "http.request.duration",
			Type:      telemetry.MetricTypeHistogram,
			Value:     float64(i * 10),
			Timestamp: time.Now().UTC(),
		})
		if !submitted {
			t.Fatalf("falha ao submeter registro %d", i)
		}
	}

	err := pool.Stop(2 * time.Second)
	if err != nil {
		t.Fatalf("erro ao encerrar pool: %v", err)
	}

	stats := pool.Stats()
	if stats.TotalProcessed != int64(totalItems) {
		t.Fatalf("esperado processamento de %d itens, obtido %d", totalItems, stats.TotalProcessed)
	}

	if ringBuffer.Size() != totalItems {
		t.Fatalf("esperado %d itens no ring buffer, obtido %d", totalItems, ringBuffer.Size())
	}
}
