# Development-only pairing HTTP API

The Go server now exposes a **loopback-only** pairing API for local integration testing. It is not yet suitable for remote clients or customer use.

| Method | Route | Capability |
| --- | --- | --- |
| POST | /v1/pairings | Returns opaque session ID and host-only secret |
| POST | /v1/pairings/{id}/request | Returns technician secret, marks awaiting approval |
| POST | /v1/pairings/{id}/approve | Host Bearer secret |
| POST | /v1/pairings/{id}/deny | Host Bearer secret |
| POST | /v1/pairings/{id}/revoke | Host Bearer secret |
| GET | /v1/pairings/{id}/authorization | Technician Bearer secret; returns approved boolean |

The server defaults to 127.0.0.1:8080 and API routes reject non-loopback remote addresses. This restriction does **not** make reverse-proxy exposure safe: a reverse proxy can appear as a loopback peer. Do not publish these endpoints or tunnel them to untrusted users.

**Known limitations:** There is no user authentication, abuse/rate limiting, origin/CSRF policy, persistent storage, TLS termination, WebRTC signaling, client identity verification, or screen/input permission enforcement. Session identifiers are opaque 256-bit values, not user-friendly short codes. All state is lost on restart. The authorization check is not permission to view or control a screen.

## Next security milestone

Implement authenticated client identities and request binding, rate limits, trusted signaling channel, and explicit view/input permission negotiation **before** any internet-facing deployment.
