# Go Distributed Telemetry & Latency Pipeline

[![CI](https://github.com/FelipeMadson/go-distributed-telemetry-pipeline/actions/workflows/ci.yml/badge.svg)](https://github.com/FelipeMadson/go-distributed-telemetry-pipeline/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8.svg)](https://golang.org/)
[![OpenTelemetry](https://img.shields.io/badge/OTLP-Compatible-blue.svg)](#opentelemetry-export)
[![Bilingual: English & Portuguese](https://img.shields.io/badge/Language-EN%20%7C%20PT--BR-blue.svg)](README.pt-BR.md)

High-throughput, concurrent distributed telemetry ingestion and latency statistics pipeline engineered in **Go 1.22+**. Features a zero-allocation sliding-window ring buffer for real-time P50/P90/P99 latency calculations, a concurrent token-bucket rate limiter, bounded worker pools with backpressure channels, and native **OpenTelemetry (OTLP)** export contracts.

Authored by **Felipe Madison** ([@FelipeMadson](https://github.com/FelipeMadson)).

[Versão em Português](README.pt-BR.md) · [Portfolio](https://felipemadson.github.io/portfolio/) · [Architecture Handbook](https://github.com/FelipeMadson/software-engineering-handbook)

---

## Architectural Topology

```
                         ┌──────────────────────────────────────────────┐
                         │              INGRESS PRODUCERS               │
                         └──────────────────────┬───────────────────────┘
                                                │ HTTP POST /v1/metrics
                                                ▼
                         ┌──────────────────────────────────────────────┐
                         │       CONCURRENT TOKEN BUCKET LIMITER        │
                         │    (Capacity: 1000 reqs, Refill: 500/sec)    │
                         └──────────────────────┬───────────────────────┘
                                                │ (Accepted)
                                                ▼
                         ┌──────────────────────────────────────────────┐
                         │       BOUNDED WORKER POOL & DISPATCHER       │
                         │   Buffered Channel (5000) + Backpressure     │
                         └──────────────┬───────────────────────────────┘
                                        │
                         ┌──────────────┴──────────────┐
                         ▼                             ▼
             ┌──────────────────────┐      ┌──────────────────────┐
             │   Worker Goroutines  │      │  SLIDING RING BUFFER │
             │  (Concurrent Normal) │      │ (10,000 Sample Window│
             └──────────┬───────────┘      └──────────┬───────────┘
                        │                             │
                        ▼                             ▼
             ┌──────────────────────┐      ┌──────────────────────┐
             │  OTLP Exporter Json  │      │ P50 / P90 / P99 Calc │
             │  OpenTelemetry Spec  │      │ Lock-Free Percentiles│
             └──────────────────────┘      └──────────────────────┘
```

---

## Key Capabilities

- **Concurrent Sliding-Window Ring Buffer**: Thread-safe circular buffer with zero reallocation overhead during steady-state ingestion.
- **Accurate Percentile Engine**: Linear interpolation algorithm calculating continuous P50, P90, and P99 latency percentiles over arbitrary rolling windows.
- **Token Bucket Rate Limiting**: Zero-lock fast-path burst limiter with smooth replenishment protecting upstream systems from overload.
- **Bounded Worker Pool with Contrapressure**: Fixed goroutine pool draining bounded buffered channels with deterministic drop metrics under extreme load.
- **OpenTelemetry (OTLP) Exporter**: Generates compliant OTLP JSON telemetry payloads directly consumable by Grafana Tempo, Jaeger, and Datadog.

---

## Quickstart

### Prerequisites
- [Go 1.22+](https://golang.org/dl/)

### Run Tests
```bash
go test -v -race -cover ./...
```

### Start Server
```bash
go run cmd/server/main.go
```

The server listens on `:8080` by default.

---

## API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/health` | Ingress health probe and service metadata |
| `POST` | `/v1/metrics` | Ingest batch of telemetry metrics with rate limit guard |
| `GET` | `/v1/stats` | Retrieve real-time P50, P90, P99 latency and pool statistics |
| `GET` | `/v1/otlp/export` | Export sliding-window metrics in OpenTelemetry OTLP JSON format |

---

## License

This project is licensed under the [MIT License](LICENSE).
Authored by **Felipe Madison** ([@FelipeMadson](https://github.com/FelipeMadson)).
