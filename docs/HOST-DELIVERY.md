# Authenticated host approval delivery boundary

This PR adds an internal `HostDelivery` component. A trusted pairing orchestrator can queue a pending approval request for one verified host connection, retrieve the ticket for that exact connection, and accept a one-use approve/deny decision only from the same authenticated connection. A disconnect cancels pending requests and tickets. The delivery boundary never trusts a host connection ID supplied in a client message.

**Not yet wired to WebSocket.** The server still rejects application messages, and the host UI is disabled. There is no real host-facing notification or UI click handling. The orchestrator must independently establish which verified connection is the host and which is the technician; this component does not provide role enrollment.

Future work: trusted role enrollment, authenticated WebSocket delivery and decisions, a genuine Linux host UI, and immediate revocation of active signaling and media. The current implementation does not terminate active sessions on disconnect; that must be enforced before allowing real support traffic.
