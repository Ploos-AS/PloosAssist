# Consent-first pairing design

This PR implements **only an in-memory state machine**. No HTTP pairing endpoints are exposed, no clients connect, and no screen data or remote input is transmitted.

## Intended flow

1. The support host creates a session. It receives an opaque random session identifier and a separate host-only capability.
2. The technician requests a session using the identifier. This creates a separate technician capability but **does not authorize** access.
3. The host explicitly approves or denies the request, using its host-only capability.
4. Approval permits only the next authenticated connection negotiation stage, **not** screen viewing or remote input on its own.
5. The host can revoke the session at any time; sessions expire automatically.

## Security constraints before network exposure

- Bind identities and permissions to authenticated, encrypted signaling and WebRTC transport.
- A support code visible to a customer must not be equivalent to a bearer authentication credential.
- Rate-limit all attempts and provide abuse protection.
- Show the technician identity to the host before approval; avoid blind approval.
- Require separate permissions for screen viewing and remote input.
- Do not expose capability tokens in URLs, logs, or analytics.
- Persist session state only if a later design requires it, with careful token handling.
- Add concurrency, clock-boundary, and fuzz tests before production use.

The current random session ID is intentionally **not** a short human-friendly support code. A human-friendly code must be added with guessing protection and a separate authenticated exchange.
