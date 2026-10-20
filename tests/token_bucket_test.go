package tests

import (
	"sync"
	"testing"
	"time"

	"github.com/FelipeMadson/go-distributed-telemetry-pipeline/pkg/ratelimit"
)

func TestTokenBucket_BurstAndDepletion(t *testing.T) {
	// Capacidade 5, 2 tokens por segundo
	tb := ratelimit.NewTokenBucket(5, 2)

	// Deve permitir 5 requisições imediatas
	for i := 0; i < 5; i++ {
		if !tb.Allow() {
			t.Fatalf("requisição %d deveria ser permitida", i)
		}
	}

	// 6ª requisição deve ser bloqueada imediatamente
	if tb.Allow() {
		t.Fatalf("requisição 6 deveria ser bloqueada por falta de tokens")
	}

	// Aguarda meio segundo (deve regenerar ~1 token)
	time.Sleep(600 * time.Millisecond)

	if !tb.Allow() {
		t.Fatalf("requisição após recarga deveria ser permitida")
	}
}

func TestTokenBucket_ConcurrentAccess(t *testing.T) {
	tb := ratelimit.NewTokenBucket(100, 10)
	var wg sync.WaitGroup
	var allowedCount int
	var mu sync.Mutex

	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if tb.Allow() {
				mu.Lock()
				allowedCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if allowedCount > 105 {
		t.Fatalf("número de requisições permitidas (%d) excedeu a capacidade do bucket", allowedCount)
	}
}
