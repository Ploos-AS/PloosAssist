# Ephemeral peer identity primitives

This change introduces **local cryptographic building blocks**, not an authenticated network service.

- Generate ephemeral Ed25519 key pairs.
- Derive a SHA-256 public-key fingerprint.
- Create random 256-bit, time-limited challenges.
- Sign and verify domain-separated challenge transcripts.

## What this does NOT guarantee

The verifier currently checks a signature and an expiry timestamp. It **does not consume challenges** or prevent reuse of the same valid challenge. No connection binding, trusted identity enrollment, host consent binding, WebSocket authentication, transport encryption, or user interface exists. A fingerprint is an identifier, not proof that the person behind the key is trustworthy.

## Requirements for the next implementation

- Maintain a concurrency-safe, server-side challenge registry with one-time consumption.
- Bind challenges to the exact connection, claimed peer key, and session context.
- Authenticate peers before processing signaling, and derive sender identity from the authenticated connection.
- Display technician identity to the support recipient before consent.
- Bind approved identities and permissions to the WebRTC peer fingerprint; verify revocation on every action.
- Add integration tests for challenge replay, identity substitution, stale connections, and cross-session authorization.

**Do not expose the development pairing API or use PloosAssist for customer support yet.**
