# Single-use challenge registry

The challenge store is a local, concurrency-safe building block for future authenticated signaling.

- Issues random, expiring challenges bound to a public key and caller-supplied connection/session identifiers.
- Consumes a challenge atomically, even when proof verification fails.
- Enforces bounded in-memory capacity and prunes expired entries when issuing new challenges.
- Tests replay, wrong context, expiry, capacity, and concurrent attempts.

**Security limitations:** Connection and session identifiers are supplied by the caller; this store does not prove they refer to an authenticated transport. It is not exposed as a network endpoint. No technician trust enrollment, pairing-to-identity binding, TLS/WSS, screen permissions, or remote access exists. Capacity is global and can be exhausted without external request admission controls.

Next: connect the challenge store to a secure authenticated transport, using server-assigned connection identifiers, and ensure session consent is bound to the authenticated peer key. Do not deploy for customer support.
