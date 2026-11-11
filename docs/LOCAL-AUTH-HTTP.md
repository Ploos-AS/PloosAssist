# Local HTTP proof-of-possession prototype

This change exposes a **loopback-only** HTTP handshake, not a WebSocket transport.

- `POST /v1/auth/start`: JSON `session_id` and base64url `public_key` (32-byte Ed25519 key). Returns server-generated `connection_id`, `challenge_id`, `nonce`, and `expires_at`.
- `POST /v1/auth/verify`: JSON `connection_id`, `challenge_id`, and base64url Ed25519 `signature`. Signature covers the domain-separated challenge transcript defined in `internal/identity`. Returns the verified public-key fingerprint.

The API is guarded by the existing loopback-only, no-browser-Origin and process-local rate-limit middleware. The challenge can be used only once. JSON requests are size-limited.

**This is not authenticated network signaling.** The HTTP requests are not bound to a persistent socket. Any caller can claim any session ID and register its own key; the returned fingerprint does not establish trusted identity or authorization. Connections are in memory, without lifecycle expiry or a network disconnect callback, so they can accumulate. Do not expose via a reverse proxy, tunnel or public interface, and do not use for customer support.

Next: implement a persistent TLS/WSS connection, bind each proof to its actual connection, enforce limits and cleanup, and connect authenticated identity to consent-gated relay authorization.
