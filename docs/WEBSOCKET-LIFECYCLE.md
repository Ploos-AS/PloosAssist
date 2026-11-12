# WebSocket disconnect lifecycle (PR #29)

The loopback-only WebSocket server now owns shared in-memory role, pending approval and signaling router state for its authenticated connections. Each socket exit invokes cleanup using its **server-generated connection ID**, before the connection registry entry is removed. The router now tracks exact connection ownership so a stale or expired disconnect cannot remove a newer peer's queue or strand a stale destination.

The server **still rejects all application messages**. There is no role-enrollment message, host consent delivery, remote screen sharing or remote input. Cleanup currently only affects internal state; the server does not yet use these registries to run a support session.

Next: integrate the pairing store with the same server instance, design an authenticated and capability-backed role enrollment message, and add transport tests. Host revocation must terminate active sockets and check consent immediately before delivering data before any support traffic is enabled.
