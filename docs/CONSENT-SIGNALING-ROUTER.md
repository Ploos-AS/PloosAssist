# Consent-gated signaling router (internal)

PR #18 adds a transport-independent signaling router, **not a public network endpoint**.

A verified connection ID maps to its authenticated session and Ed25519 fingerprint in the existing connection registry. `Attach` installs a bounded outbound queue for that identity. `Route` validates the signaling envelope, rejects spoofed sender/session fields, checks the currently active host consent for the exact host–technician pair **on every message**, and delivers to the recipient queue. Revocation blocks subsequent routing. A full queue returns an error rather than growing without limit.

**Critical trust boundary:** The caller is responsible for proving both identities, authenticating the actual live socket, and obtaining the host's explicit approval. The existing consent policy's `Approve` method must never be called from untrusted client-provided fingerprints. This PR does **not** connect the router to the WebSocket handler because no safe end-to-end host approval flow is available yet. Therefore, the WebSocket handler still rejects application messages; no offer, answer, or ICE candidate is forwarded over the network.

Next: bind the host's UI consent action to its authenticated socket, implement session membership/pairing, integrate this router with WebSocket connections, and test revocation/disconnect races before enabling real signaling. WSS and additional abuse controls are required before public deployment.
