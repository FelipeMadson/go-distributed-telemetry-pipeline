package ratelimit

import (
	"sync"
	"time"
)

// TokenBucket implementa um limitador de taxa de alto desempenho baseado em Token Bucket thread-safe.
type TokenBucket struct {
	mu         sync.Mutex
	capacity   float64
	tokens     float64
	refillRate float64 // tokens por segundo
	lastRefill time.Time
}

// NewTokenBucket cria uma instância de TokenBucket configurada.
func NewTokenBucket(capacity float64, refillRate float64) *TokenBucket {
	if capacity <= 0 {
		capacity = 100
	}
	if refillRate <= 0 {
		refillRate = 50
	}
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow verifica se 1 token está disponível para consumo.
func (tb *TokenBucket) Allow() bool {
	return tb.AllowN(1)
}

// AllowN tenta consumir n tokens de forma atômica.
func (tb *TokenBucket) AllowN(n float64) bool {
	if n <= 0 {
		return true
	}

	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.lastRefill = now

	// Adiciona tokens proporcionais ao tempo transcorrido
	tb.tokens += elapsed * tb.refillRate
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}

	if tb.tokens >= n {
		tb.tokens -= n
		return true
	}

	return false
}

// AvailableTokens retorna o saldo de tokens instantâneo (útil para telemetria).
func (tb *TokenBucket) AvailableTokens() float64 {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	return tb.tokens
}
