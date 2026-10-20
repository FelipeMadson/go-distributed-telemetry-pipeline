package worker

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/FelipeMadson/go-distributed-telemetry-pipeline/pkg/telemetry"
)

// PipelineStats mantém métricas atômicas agregadas sobre a execução da pipeline.
type PipelineStats struct {
	TotalIngested  int64 `json:"total_ingested"`
	TotalProcessed int64 `json:"total_processed"`
	TotalDropped   int64 `json:"total_dropped"`
}

// WorkerPool gerencia um pool de goroutines concorrentes com canais limitados e contrapressão.
type WorkerPool struct {
	workerCount int
	bufferSize  int
	queue       chan telemetry.MetricRecord
	ringBuffer  *telemetry.RingBuffer
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc

	ingestedCount  atomic.Int64
	processedCount atomic.Int64
	droppedCount   atomic.Int64
}

// NewWorkerPool inicializa um WorkerPool configurado.
func NewWorkerPool(workerCount, bufferSize int, ringBuffer *telemetry.RingBuffer) *WorkerPool {
	if workerCount <= 0 {
		workerCount = 4
	}
	if bufferSize <= 0 {
		bufferSize = 1000
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &WorkerPool{
		workerCount: workerCount,
		bufferSize:  bufferSize,
		queue:       make(chan telemetry.MetricRecord, bufferSize),
		ringBuffer:  ringBuffer,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start dispara as goroutines trabalhadoras.
func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.runWorker(i)
	}
}

func (wp *WorkerPool) runWorker(id int) {
	defer wp.wg.Done()

	for {
		select {
		case <-wp.ctx.Done():
			// Drena as mensagens residuais da fila antes de encerrar
			for rec := range wp.queue {
				wp.processRecord(rec)
			}
			return

		case rec, ok := <-wp.queue:
			if !ok {
				return
			}
			wp.processRecord(rec)
		}
	}
}

func (wp *WorkerPool) processRecord(rec telemetry.MetricRecord) {
	start := time.Now()

	// Simula processamento / normalização
	if rec.Type == telemetry.MetricTypeHistogram || rec.Name == "http.request.duration" {
		wp.ringBuffer.Push(rec.Value)
	}

	wp.processedCount.Add(1)

	// Mede a latência interna do worker
	_ = time.Since(start)
}

// Submit submete uma métrica à fila de processamento com contrapressão não bloqueante.
func (wp *WorkerPool) Submit(rec telemetry.MetricRecord) bool {
	wp.ingestedCount.Add(1)

	select {
	case wp.queue <- rec:
		return true
	default:
		// Fila cheia: contrapressão descarta para proteger integridade da memória
		wp.droppedCount.Add(1)
		return false
	}
}

// Stats retorna as métricas atômicas do pool.
func (wp *WorkerPool) Stats() PipelineStats {
	return PipelineStats{
		TotalIngested:  wp.ingestedCount.Load(),
		TotalProcessed: wp.processedCount.Load(),
		TotalDropped:   wp.droppedCount.Load(),
	}
}

// Stop encerra o pool graciosamente, esvaziando as mensagens pendentes.
func (wp *WorkerPool) Stop(timeout time.Duration) error {
	wp.cancel()
	close(wp.queue)

	done := make(chan struct{})
	go func() {
		wp.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("timeout expirado ao encerrar worker pool")
	}
}
