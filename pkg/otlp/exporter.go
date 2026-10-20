package otlp

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/FelipeMadson/go-distributed-telemetry-pipeline/pkg/telemetry"
)

// OTLPResourceMetrics modela a estrutura de exportação compatível com OpenTelemetry v1.
type OTLPResourceMetrics struct {
	Resource      Resource       `json:"resource"`
	ScopeMetrics  []ScopeMetric  `json:"scopeMetrics"`
	SchemaURL     string         `json:"schemaUrl,omitempty"`
}

type Resource struct {
	Attributes []KeyValue `json:"attributes"`
}

type KeyValue struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type ScopeMetric struct {
	Scope   InstrumentationScope `json:"scope"`
	Metrics []Metric             `json:"metrics"`
}

type InstrumentationScope struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Metric struct {
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Unit        string     `json:"unit,omitempty"`
	Gauge       *GaugeData `json:"gauge,omitempty"`
	Sum         *SumData   `json:"sum,omitempty"`
}

type GaugeData struct {
	DataPoints []DataPoint `json:"dataPoints"`
}

type SumData struct {
	AggregationTemporality int         `json:"aggregationTemporality"`
	IsMonotonic            bool        `json:"isMonotonic"`
	DataPoints             []DataPoint `json:"dataPoints"`
}

type DataPoint struct {
	TimeUnixNano int64       `json:"timeUnixNano"`
	AsDouble     float64     `json:"asDouble"`
	Attributes   []KeyValue  `json:"attributes,omitempty"`
}

// ConvertToOTLPPayload transforma uma fatia de registros em payload OTLP JSON compatível.
func ConvertToOTLPPayload(serviceName string, records []telemetry.MetricRecord) ([]byte, error) {
	dataPoints := make([]DataPoint, 0, len(records))

	for _, rec := range records {
		attrs := make([]KeyValue, 0, len(rec.Tags))
		for k, v := range rec.Tags {
			attrs = append(attrs, KeyValue{Key: k, Value: map[string]string{"stringValue": v}})
		}

		dp := DataPoint{
			TimeUnixNano: rec.Timestamp.UnixNano(),
			AsDouble:     rec.Value,
			Attributes:   attrs,
		}
		dataPoints = append(dataPoints, dp)
	}

	payload := map[string]interface{}{
		"resourceMetrics": []OTLPResourceMetrics{
			{
				Resource: Resource{
					Attributes: []KeyValue{
						{Key: "service.name", Value: map[string]string{"stringValue": serviceName}},
						{Key: "telemetry.sdk.language", Value: map[string]string{"stringValue": "go"}},
						{Key: "telemetry.sdk.version", Value: map[string]string{"stringValue": "1.0.0"}},
					},
				},
				ScopeMetrics: []ScopeMetric{
					{
						Scope: InstrumentationScope{
							Name:    "go-distributed-telemetry-pipeline",
							Version: "1.0.0",
						},
						Metrics: []Metric{
							{
								Name: "pipeline.ingested.records",
								Unit: "1",
								Gauge: &GaugeData{
									DataPoints: dataPoints,
								},
							},
						},
					},
				},
			},
		},
	}

	bytes, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("falha ao serializar payload OTLP: %w", err)
	}
	return bytes, nil
}
