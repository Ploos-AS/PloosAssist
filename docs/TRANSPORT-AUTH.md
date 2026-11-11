# Connection authentication state machine

This PR introduces a **transport-independent authentication state machine**, not a network listener.

A future network adapter can call:
1. `Open(sessionID, publicKey)` to allocate an unpredictable server-side connection identifier.
2. `Challenge(connectionID)` to issue a short-lived, one-time challenge bound to that connection and session.
3. `Authenticate(connectionID, challengeID, signature)` to verify Ed25519 key possession.
4. `Identity(connectionID)` to retrieve the verified public-key fingerprint and session ID.
5. `Close(connectionID)` on disconnect.

The registry prevents unauthenticated identity reads, replayed challenges, cross-connection proof reuse, and successful reauthentication of the same connection.

## Security boundaries

**This is not yet a secure remote support connection.** A caller can currently supply any session ID and public key when opening a connection. The fingerprint proves possession of a key, **not** that the key belongs to a trusted technician or an approved customer. The implementation has no WebSocket/TLS server, transport binding, consent-to-identity binding, session expiry, connection count limits, automatic cleanup of abandoned connections, or WebRTC media permissions. The in-memory connection registry is intended for a controlled prototype only.

A production network adapter must bind connection IDs to its actual accepted sockets, enforce limits/timeouts, verify session membership, and connect identity to host consent and the signaling relay before forwarding any messages.
