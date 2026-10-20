package telemetry

import (
	"math"
	"sort"
)

// LatencyStats agrupa as estatísticas calculadas a partir de uma janela de latência.
type LatencyStats struct {
	Count float64 `json:"count"`
	Min   float64 `json:"min_ms"`
	Max   float64 `json:"max_ms"`
	Mean  float64 `json:"mean_ms"`
	P50   float64 `json:"p50_ms"`
	P90   float64 `json:"p90_ms"`
	P99   float64 `json:"p99_ms"`
}

// CalculatePercentiles calcula P50, P90 e P99 com interpolação precisa sobre uma amostra.
func CalculatePercentiles(samples []float64) LatencyStats {
	if len(samples) == 0 {
		return LatencyStats{}
	}

	sorted := make([]float64, len(samples))
	copy(sorted, samples)
	sort.Float64s(sorted)

	total := 0.0
	minVal := sorted[0]
	maxVal := sorted[len(sorted)-1]

	for _, v := range sorted {
		total += v
	}

	mean := total / float64(len(sorted))

	return LatencyStats{
		Count: float64(len(sorted)),
		Min:   round2(minVal),
		Max:   round2(maxVal),
		Mean:  round2(mean),
		P50:   round2(percentile(sorted, 50)),
		P90:   round2(percentile(sorted, 90)),
		P99:   round2(percentile(sorted, 99)),
	}
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if len(sorted) == 1 {
		return sorted[0]
	}

	rank := (p / 100.0) * float64(len(sorted)-1)
	lower := int(math.Floor(rank))
	upper := int(math.Ceil(rank))
	weight := rank - float64(lower)

	if lower == upper {
		return sorted[lower]
	}

	return sorted[lower]*(1.0-weight) + sorted[upper]*weight
}

func round2(val float64) float64 {
	return math.Round(val*100) / 100
}
