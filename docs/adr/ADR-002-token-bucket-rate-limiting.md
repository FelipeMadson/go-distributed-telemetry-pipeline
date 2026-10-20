# ADR-002: Token Bucket Rate Limiting with Microsecond Resolution

## Status
Accepted

## Context
External ingest endpoints are vulnerable to traffic surges and denial-of-service spikes. Traditional fixed-window rate limiters create burst spikes at window boundaries ("boundary burst effect"), overwhelming downstream workers and exporters.

## Decision
We implemented a **Continuous Token Bucket Rate Limiter**:
1. **Time-Elapsed Token Accrual**: Rather than background tickers waking up every millisecond, tokens are dynamically calculated upon request arrival using `time.Since(lastUpdate).Seconds() * refillRate`.
2. **Atomic Synchronization**: Mutex contention is confined to token subtraction, ensuring thread safety across parallel worker pools.
3. **Smooth Traffic Shaping**: Excess requests are immediately rejected or shed with `ErrRateLimited`, preventing downstream buffer exhaustion.

## Consequences
- **Positive**: Perfectly smooth traffic egress without periodic boundary bursts; negligible CPU footprint when idle.
- **Trade-off**: Requires client coordination for backoff/retry handling (HTTP 429 Too Many Requests).
