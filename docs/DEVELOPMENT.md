# Development

Requires Go 1.23 or newer.

```sh
go test ./...
go vet ./...
go run ./cmd/server -listen 127.0.0.1:8080
curl http://127.0.0.1:8080/healthz
```

The server currently exposes only a loopback health endpoint. It does **not** support pairing, signaling, screen sharing, or remote input. Do not expose it publicly or use it for customer support.

The `internal/protocol` package defines a small versioned JSON envelope and rejects unknown message kinds. Future authentication, rate limits, consent, and peer identity binding must be implemented before accepting session requests.
