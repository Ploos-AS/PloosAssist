# Real host/server handshake integration test

This test runs the real guarded Go HTTP handler on an ephemeral loopback port and connects the actual `internal/hostclient` implementation to it. It checks the Ed25519 challenge/proof round trip and the verified fingerprint. It then sends an unauthorized application message and confirms the server closes the connection with a WebSocket policy violation.

Run:

```sh
go test -race ./cmd/server ./internal/hostclient
```

**Scope:** This is an end-to-end authentication test, **not** a working approval UI or remote-support session. The server still rejects signaling and has no network-bound pairing/consent flow. The host diagnostic remains CLI-only. Do not expose the development server publicly.

## Next

Before enabling the prototype's approval controls, implement authenticated role enrollment, deliver a pending request over the exact verified host connection, and prove an explicit host-side decision travels back on that same socket. Then implement a trusted graphical Linux host surface and tests for disconnect, replay, denial and revocation.
