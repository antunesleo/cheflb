# cheflb

A learning-project load balancer written from scratch in Go.

## What it does

Forwards incoming traffic to a pool of backend servers using one of several balancing strategies. Supports both layer 4 (raw TCP proxying) and layer 7 (HTTP, via reverse proxy or 307 redirect).

## Layout

- `cmd/server/main.go` — entry point. Reads `CHEFLB_LAYER` and `CHEFLB_ALGORITHM`, builds the LB, and starts the chosen server.
- `internal/server/layer4.go` — TCP listener that bidirectionally pipes bytes between client and chosen backend.
- `internal/server/layer7.go` — HTTP handler that either reverse-proxies (`forwardMode = "request"`) or redirects (`"redirect"`).
- `internal/lbs/lbs.go` — the `LoadBalancer` interface and its implementations.
- `targetserver/` — two Python HTTP servers (`server1.py` on `:7171`, `server2.py` on `:8181`) used as backends for local testing. They sleep a random 1–3s to simulate variable latency.

## Balancing strategies

All implement `Balance(ipAddress string) *Server`:

- **Round robin** (`round_robin`, `RoundRobinLb`) — cycles through the pool in order via an atomic counter.
- **Hash** (`hash`, `HashLb`) — MurmurHash3 of the client IP modulo pool size, giving stable client→backend affinity.
- **Least response time** (`least_response_time`, `LeastRespTimeLb`) — picks the server with the lowest EWMA of response time.

## Run it

Start the backends in two terminals:

```sh
python3 targetserver/server1.py
python3 targetserver/server2.py
```

Then build and run the LB on `:8080`:

```sh
make run
```

## Configuration

Configured via environment variables:

- `CHEFLB_LAYER` — `4` or `7`. Default `7`.
- `CHEFLB_ALGORITHM` — `round_robin`, `hash`, or `least_response_time`. Default `round_robin`.

Example:

```sh
CHEFLB_LAYER=4 CHEFLB_ALGORITHM=hash make run
```

Invalid values cause the binary to exit at startup with a clear error.
