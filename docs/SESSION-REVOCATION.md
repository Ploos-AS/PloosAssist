# Coordinated in-process session revocation

PR #28 adds an internal `Revocation` coordinator. After the independent host pairing capability is verified by the pairing store, it marks the pairing revoked, closes consent, clears role bindings, cancels pending approvals and one-use tickets, and removes buffered signaling queues.

An invalid host capability leaves the session unchanged. Newly routed messages fail the consent check after revocation.

**Limitations:** This is not yet connected to the loopback HTTP API or authenticated WebSocket connections. It does not terminate transport sockets, screen streams or input channels, and cannot retract a message already dequeued by a writer. Future transport writers MUST recheck current consent before sending and stop immediately on revocation. Full shutdown also needs to coordinate concurrent enrollment/approval against revocation to prevent races; do not expose signaling or remote input before that work is complete.
