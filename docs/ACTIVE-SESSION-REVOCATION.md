# Active session revocation (PR #38)

Both the existing host-capability-protected HTTP `POST /v1/pairings/{id}/revoke` endpoint and the new authenticated host WebSocket `{"type":"session_revoke","capability":"<host capability>"}` action now invoke the shared in-memory `signaling.Revocation` coordinator.

The coordinator first validates the independent host capability and revokes the pairing, then clears consent, role bindings, pending approvals, tickets and routing queues. The server subsequently closes **all tracked WebSocket connections** for the session (host and technician), not merely the requesting connection. Unknown credentials must not trigger socket closure. WebSocket revocation intentionally closes the socket without sending an acknowledgment.

This is a server-side foundation. The Linux browser UI still lacks an explicit Stop button and the host client does not yet invoke the new WebSocket action. The next change must implement an always-visible customer Stop action, tests for both transports and lifecycle/race handling. There are no media streams or remote input channels to stop yet. Loopback-only development; do not treat this as production-ready.
