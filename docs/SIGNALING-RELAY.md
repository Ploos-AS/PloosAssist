# Authenticated signaling relay core

This PR implements a **transport-independent authorization policy**, not a network-facing WebSocket endpoint.

The relay registers a host and technician identity per session, starts in an unapproved state, permits only peer-to-peer signaling after approval, and denies all traffic following revocation. The authenticated sender is passed separately from the JSON envelope so the sender field alone cannot impersonate a peer.

## Critical trust boundary

`Register`, `Approve`, and `Authorize` are internal functions. The caller must independently verify the identities and the host's actual consent. In particular, **passing a client-supplied sender ID as the authenticated sender defeats the security model**. No network endpoint may invoke `Approve` solely because a client claims to be the host.

## Required before WSS deployment

- Establish TLS/WSS and a verified identity per connection using the single-use challenge store.
- Bind server-assigned connection IDs to cryptographic identity proofs.
- Tie the host's approval to the exact authenticated technician identity and active session.
- Revalidate session expiry and revocation for every message; current relay does not implement expiry.
- Add bounded queues, payload limits, per-peer rate limiting, heartbeat/timeouts and integration tests.
- Bind WebRTC DTLS fingerprints and media permissions to approved identities.

The existing HTTP pairing API remains loopback-only and development-only. This relay is not yet wired to it or to any network transport.
