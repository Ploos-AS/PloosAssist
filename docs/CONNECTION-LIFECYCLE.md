# Authenticated connection lifecycle (internal)

This PR introduces an internal bounded connection registry with explicit registration, identity lookup, lease refresh, expiry, and disconnect cleanup.

**Trust boundary:** `Register` accepts already verified identities and MUST only be called by a transport adapter after authenticating a real socket. The registry does not authenticate connections itself, validate host consent, or authorize signaling.

The existing `/v1/ws` endpoint remains a **one-shot loopback-only experiment** and is not connected to this registry. This PR does not deliver a production WebSocket transport.

## Next steps

1. Adopt a maintained WebSocket library, including pinned module checksums and CI verification.
2. Bind the connection registry to real socket lifetimes, deadlines, and heartbeat handling.
3. Validate session membership and explicit host approval of the authenticated technician fingerprint before any relay.
4. Add bounded queues, backpressure, per-peer rate limits, and WSS before Internet deployment.

Do not expose the development server through reverse proxies or tunnels.
