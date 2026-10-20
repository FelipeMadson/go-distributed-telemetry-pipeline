package telemetry

import (
	"fmt"
	"time"
)

// MetricType define os tipos fundamentais de métricas suportadas pela pipeline.
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
)

// MetricRecord representa uma amostra de métrica individual imutável.
type MetricRecord struct {
	Name      string            `json:"name"`
	Type      MetricType        `json:"type"`
	Value     float64           `json:"value"`
	Timestamp time.Time         `json:"timestamp"`
	Tags      map[string]string `json:"tags,omitempty"`
}

// Validate assegura as invariantes de integridade do registro.
func (m *MetricRecord) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("nome da métrica não pode ser vazio")
	}
	if m.Type != MetricTypeCounter && m.Type != MetricTypeGauge && m.Type != MetricTypeHistogram {
		return fmt.Errorf("tipo de métrica inválido: %s", m.Type)
	}
	if m.Timestamp.IsZero() {
		m.Timestamp = time.Now().UTC()
	}
	return nil
}
