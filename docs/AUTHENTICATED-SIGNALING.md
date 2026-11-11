# Authenticated signaling: design and security gates

**Status: design only.** This change adds a validated, transport-independent envelope for future WebRTC offer/answer/ICE messages. No network signaling endpoints are added. The existing development pairing API is still loopback-only and is not suitable for real customer sessions.

## Proposed security design

1. **Client identity:** each support host and technician establishes an authenticated connection with an ephemeral cryptographic identity; proof of key possession is required. Do not trust a sender ID supplied in a JSON message.
2. **Transport:** HTTPS/WSS with TLS and strict origin/host handling. WebSocket upgrade must not bypass authentication or per-session authorization.
3. **Pairing:** host consent is bound to the authenticated technician identity, session ID, and expiration; host UI displays technician identity before approval. A session code alone never grants access.
4. **Message authorization:** only the two approved peers can exchange SDP and ICE candidates. Derive sender identity from the authenticated connection, not the envelope; reject unauthorized recipient IDs and stale/replayed messages.
5. **Media security:** use authenticated WebRTC DTLS/SRTP and bind peer fingerprints to the approved signaling session. TURN only relays encrypted media; a signaling server must not be able to substitute peers unnoticed.
6. **Permissions:** screen viewing and remote input are separately granted and revocable. Signaling authorization is not remote-control authorization.
7. **Operational protections:** per-identity/IP limits, bounded message sizes, timeouts, session cleanup, audit events without secrets, and explicit reverse-proxy configuration.

## Planned implementation sequence

- Add ephemeral identity registration and challenge-response tests.
- Bind consent to verified host and technician identities.
- Add authenticated WSS message relay with session-scoped authorization.
- Add integration tests for impersonation, replay, revocation, cross-session messages and concurrent sessions.
- Only then connect WebRTC peer negotiation and screen sharing.

## Non-goals in this PR

No WSS server, authentication implementation, NAT traversal, TURN, screen capture, or remote input. Do not expose the development server on the public internet.
