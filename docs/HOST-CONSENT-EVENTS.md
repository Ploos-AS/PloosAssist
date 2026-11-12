# Host consent event binding (internal)

This PR introduces one-use, cryptographically random approval ticket IDs. A ticket is created only for an existing pending request, bound to its original authenticated host connection ID and exact technician fingerprint. A decision from a different connection is rejected; a successful decision consumes the ticket, preventing replay. Denial grants no access.

**No network integration yet.** The existing host UI remains a static prototype with disabled controls, and the WebSocket server still rejects all application messages. This module cannot establish that a person actually saw or clicked the host dialog. Tickets must not be treated as bearer authorization: the real host connection ID must be derived from the authenticated socket.

## Next integration

Implement a trusted host application surface and verified pairing/roles, deliver the ticket and technician fingerprint only to that host, then accept a user-initiated decision on the same socket. On disconnect, cancel pending tickets and approvals; on revoke, flush queued signaling and stop media/input. Add end-to-end tests before enabling network signaling.
