package telemetry

import (
	"sync"
)

// RingBuffer implementa um buffer circular com janela deslizante de tamanho fixo
// thread-safe para cálculo contínuo de latência com contenção minimizada.
type RingBuffer struct {
	mu       sync.RWMutex
	capacity int
	data     []float64
	cursor   int
	isFull   bool
}

// NewRingBuffer inicializa um buffer circular com capacidade pré-alocada.
func NewRingBuffer(capacity int) *RingBuffer {
	if capacity <= 0 {
		capacity = 1000
	}
	return &RingBuffer{
		capacity: capacity,
		data:     make([]float64, capacity),
		cursor:   0,
		isFull:   false,
	}
}

// Push insere uma nova amostra no buffer circular.
func (r *RingBuffer) Push(val float64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.data[r.cursor] = val
	r.cursor++
	if r.cursor >= r.capacity {
		r.cursor = 0
		r.isFull = true
	}
}

// Snapshot retorna uma cópia das amostras atualmente presentes na janela deslizante.
func (r *RingBuffer) Snapshot() []float64 {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []float64
	if !r.isFull {
		result = make([]float64, r.cursor)
		copy(result, r.data[:r.cursor])
	} else {
		result = make([]float64, r.capacity)
		// Ordem cronológica da janela deslizante
		n := copy(result, r.data[r.cursor:])
		copy(result[n:], r.data[:r.cursor])
	}
	return result
}

// Size retorna o número atual de amostras armazenadas.
func (r *RingBuffer) Size() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.isFull {
		return r.capacity
	}
	return r.cursor
}

// Reset esvazia o buffer circular.
func (r *RingBuffer) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.cursor = 0
	r.isFull = false
}
