# Local authenticated WebSocket transport (development only)

`GET /v1/ws?session_id=...&public_key=...` uses `github.com/gorilla/websocket` v1.5.3 instead of the experimental hand-written frame parser. The old `internal/wsproto` package is retained temporarily for historical tests but is **not used by this endpoint**.

The client receives a challenge and signs it using its Ed25519 private key. On successful verification, the server responds with `proof_verified`, registers the authenticated identity, and keeps the socket open.

- Maximum 128 concurrent sockets per server process, including unauthenticated handshakes.
- 4 KiB message limit, 15-second authentication timeout.
- Ping every 20 seconds; 60-second idle timeout refreshed on pong.
- Connection registry lease refreshed on pong, cleaned up on disconnect.
- Application text/binary messages are **rejected**. No signaling or media is permitted.
- The HTTP guard requires numeric loopback clients, rejects browser Origin headers, and limits requests.
- This remains plaintext ws:// on loopback. Do **not** expose it through a reverse proxy or tunnel.

**Security:** Proof of key possession is not proof of session membership or host consent. Future signaling must bind authenticated transport identity to approved host/technician session and check revocation on each message.

## Dependency integrity

The pinned `gorilla/websocket` v1.5.3 module is resolved and verified in CI via `go mod download` and `go mod verify`. This branch does not include a checked-in `go.sum` because the authoring environment cannot access the public Go module proxy. Before merging, preferably generate and commit `go.sum` from a networked Go environment using `go mod tidy`; do not fabricate checksums. CI can confirm dependency retrieval but does not replace a checked-in checksum file.
