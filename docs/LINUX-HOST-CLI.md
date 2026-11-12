# Linux host diagnostic (PR #23)

Run the development server locally, then run:

```sh
go run ./cmd/server
go run ./cmd/host -session demo
```

The host CLI creates an ephemeral Ed25519 identity, connects to `ws://127.0.0.1:8080/v1/ws`, signs the server's challenge and checks that the server's verified fingerprint matches the local key. It then waits for disconnection or Ctrl+C. The endpoint is restricted to numeric loopback IPs to prevent sending development credentials to remote services.

**This is not yet the promised graphical Linux client.** The CLI cannot receive approval requests, show the approval UI, stream the screen, or allow remote input. The server currently rejects application messages. This PR establishes a working local client/server authentication smoke-test before integrating the UI and consent state machine.

Security limitations: the server itself is not authenticated to the client and ws:// is unencrypted. Only run on a trusted local machine; do not expose via a proxy or tunnel. A production implementation needs TLS/WSS, server identity validation, trusted role enrollment and a genuine host-side consent surface.
