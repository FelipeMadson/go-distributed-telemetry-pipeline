# Pipeline de Telemetria e Latência Distribuída em Go

[![CI](https://github.com/FelipeMadson/go-distributed-telemetry-pipeline/actions/workflows/ci.yml/badge.svg)](https://github.com/FelipeMadson/go-distributed-telemetry-pipeline/actions)
[![Licença: MIT](https://img.shields.io/badge/Licen%C3%A7a-MIT-yellow.svg)](LICENSE)
[![Versão do Go](https://img.shields.io/badge/Go-1.22+-00ADD8.svg)](https://golang.org/)
[![OpenTelemetry](https://img.shields.io/badge/OTLP-Compat%C3%ADvel-blue.svg)](#exporta%C3%A7%C3%A3o-opentelemetry)
[![Bilingue: Inglês & Português](https://img.shields.io/badge/Idioma-EN%20%7C%20PT--BR-blue.svg)](README.md)

Pipeline de ingestão de telemetria distribuída e cálculo concorrente de estatísticas de latência de altíssima vazão desenvolvida em **Go 1.22+**. Implementa um buffer circular com janela deslizante de alocação zero para cálculo contínuo de percentis P50/P90/P99, limitador de taxa Token Bucket thread-safe, pool de workers limitados com canais de contrapressão e exportação nativa compatível com **OpenTelemetry (OTLP)**.

De autoria de **Felipe Madison** ([@FelipeMadson](https://github.com/FelipeMadson)).

[English Version](README.md) · [Portfólio](https://felipemadson.github.io/portfolio/) · [Manual de Engenharia](https://github.com/FelipeMadson/software-engineering-handbook)

---

## Topologia Arquitetural

```
                         ┌──────────────────────────────────────────────┐
                         │             PRODUTORES DE INGRESSO           │
                         └──────────────────────┬───────────────────────┘
                                                │ HTTP POST /v1/metrics
                                                ▼
                         ┌──────────────────────────────────────────────┐
                         │       LIMITADOR TOKEN BUCKET CONCORRENTE     │
                         │    (Capacidade: 1000 reqs, Recarga: 500/s)   │
                         └──────────────────────┬───────────────────────┘
                                                │ (Aceito)
                                                ▼
                         ┌──────────────────────────────────────────────┐
                         │      POOL DE WORKERS LIMITADO & DESPACHO     │
                         │     Canal Bufferizado (5000) + Contrapressão │
                         └──────────────┬───────────────────────────────┘
                                        │
                         ┌──────────────┴──────────────┐
                         ▼                             ▼
             ┌──────────────────────┐      ┌──────────────────────┐
             │ Goroutines de Worker │      │   BUFFER CIRCULAR    │
             │(Normalização Paralela│      │(Janela 10.000 amostra│
             └──────────┬───────────┘      └──────────┬───────────┘
                        │                             │
                        ▼                             ▼
             ┌──────────────────────┐      ┌──────────────────────┐
             │ Exportador OTLP Json │      │ Cálculo P50/P90/P99  │
             │   Padrão OpenTelemetry      │  Percentis Contínuos │
             └──────────────────────┘      └──────────────────────┘
```

---

## Principais Recursos

- **Buffer Circular de Janela Deslizante Concorrente**: Estrutura de dados thread-safe com custo de realocação zero em regime permanente.
- **Motor de Percentis Preciso**: Algoritmo de interpolação linear para cálculo determinístico de latências P50, P90 e P99 sobre amostras móveis.
- **Controle de Vazão Token Bucket**: Limitador de picos de alto desempenho com reabastecimento contínuo protegendo serviços a montante.
- **Pool de Workers com Contrapressão**: Goroutines dedicadas processando canais com buffer estrito e contabilização de descarte sob sobrecarga severa.
- **Exportador OpenTelemetry (OTLP)**: Geração de payloads compatíveis com OTLP JSON para integração direta com Grafana Tempo, Jaeger e Datadog.

---

## Execução Rápida

### Pré-requisitos
- [Go 1.22+](https://golang.org/dl/)

### Executar Testes
```bash
go test -v -race -cover ./...
```

### Iniciar Servidor
```bash
go run cmd/server/main.go
```

O servidor escuta na porta `:8080` por padrão.

---

## Licença

Este projeto é licenciado sob a [Licença MIT](LICENSE).
De autoria de **Felipe Madison** ([@FelipeMadson](https://github.com/FelipeMadson)).
