# Authenticated WebSocket lifecycle integration

The existing **one-shot, loopback-only** WebSocket proof-of-possession endpoint now registers its cryptographically verified connection identity with the bounded, expiring signaling connection registry. Both authentication state and the registry entry are removed when the socket handler exits. The endpoint still returns a proof confirmation and disconnects immediately; it does not relay messages.

This PR does **not** introduce a maintained WebSocket library, persistent connections, TLS/WSS, or production-ready support. The local experimental frame parser remains in use. An attempt to retrieve and verify an external dependency was blocked by unavailable network access, so no unverified module version or fabricated go.sum was committed.

Before enabling persistent signaling, replace the experimental framing with a maintained, audited dependency, pin its version and checksums, add heartbeat and backpressure, and verify the actual connected peer is an approved member of the session. Proof of key possession alone is not authorization.
