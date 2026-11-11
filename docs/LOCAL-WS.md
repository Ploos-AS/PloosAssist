# Experimental local WebSocket proof handshake

`GET /v1/ws?session_id=...&public_key=...` is a **loopback-only, one-shot RFC 6455 development experiment**. The client sends an Ed25519 public key (base64url), receives a nonce, and responds with a masked text frame containing JSON `{"type":"proof","signature":"<base64url>"}`. The signature covers the identity package's domain-separated challenge transcript. On success the server responds with `proof_verified` and immediately disconnects.

**No signaling, relay, screen viewing, remote control, or trusted peer membership is provided.**

The tiny built-in frame parser deliberately supports only small, masked, unfragmented text frames. It does not support ping/pong, fragmentation, close frames, or production WebSocket lifecycle semantics. The handshake is limited to 15 seconds and is not suitable for a persistent connection. It uses plaintext WebSocket over loopback, not WSS. Never expose it through a reverse proxy or tunnel.

The client can claim any session ID and supply its own key. Key possession alone does not authorize a support session. A future production transport should use a maintained WebSocket implementation with TLS, bounded queues, ping/pong, backpressure, cancellation, authenticated session membership, and host consent bound to the verified technician key.
