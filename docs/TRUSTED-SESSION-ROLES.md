# Verified session role binding

This internal component binds distinct, verified host and technician connection IDs to one session, rejects duplicate roles and self-pairing, and fails closed if either connection disappears or the binding expires. Disconnect removes the entire binding.

**Critical trust boundary:** `Roles.Bind` does **not** validate host or technician capabilities. It must only be called by trusted server orchestration after verifying independent host/technician role credentials, for example credentials derived from `internal/session.Store` and bound to the authenticated connection. Do not expose `Bind` as an endpoint that accepts a caller-supplied role. The current pairing store does not yet provide a safe end-to-end role-enrollment API.

No WebSocket application messages are enabled, and the UI remains inactive. Before live consent delivery, implement capability-backed enrollment on authenticated sockets, then connect the verified role pair to HostDelivery and ApprovalEvents. On disconnect, clear pending consent and any active traffic.
