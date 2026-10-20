package tests

import (
	"testing"

	"github.com/FelipeMadson/go-distributed-telemetry-pipeline/pkg/telemetry"
)

func TestRingBuffer_PushAndSnapshot(t *testing.T) {
	rb := telemetry.NewRingBuffer(5)

	// Inserção inferior à capacidade
	rb.Push(10.0)
	rb.Push(20.0)
	rb.Push(30.0)

	if rb.Size() != 3 {
		t.Fatalf("esperado tamanho 3, obtido %d", rb.Size())
	}

	snap := rb.Snapshot()
	if len(snap) != 3 {
		t.Fatalf("esperado slice com 3 elementos, obtido %d", len(snap))
	}
	if snap[0] != 10.0 || snap[1] != 20.0 || snap[2] != 30.0 {
		t.Errorf("conteúdo inesperado do snapshot: %v", snap)
	}

	// Rollover da capacidade (5 elementos)
	rb.Push(40.0)
	rb.Push(50.0)
	rb.Push(60.0) // Sobrescreve 10.0

	if rb.Size() != 5 {
		t.Fatalf("esperado tamanho 5 após overflow, obtido %d", rb.Size())
	}

	snap2 := rb.Snapshot()
	expected := []float64{20.0, 30.0, 40.0, 50.0, 60.0}
	for i, v := range expected {
		if snap2[i] != v {
			t.Errorf("índice %d: esperado %.1f, obtido %.1f", i, v, snap2[i])
		}
	}
}

func TestCalculatePercentiles(t *testing.T) {
	samples := []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}

	stats := telemetry.CalculatePercentiles(samples)

	if stats.Count != 10 {
		t.Errorf("esperado count 10, obtido %f", stats.Count)
	}
	if stats.Min != 10 {
		t.Errorf("esperado min 10, obtido %f", stats.Min)
	}
	if stats.Max != 100 {
		t.Errorf("esperado max 100, obtido %f", stats.Max)
	}
	if stats.Mean != 55 {
		t.Errorf("esperado mean 55, obtido %f", stats.Mean)
	}
	if stats.P50 != 55 {
		t.Errorf("esperado P50 55, obtido %f", stats.P50)
	}
	if stats.P90 != 91 {
		t.Logf("P90 calculado: %f", stats.P90)
	}
}
