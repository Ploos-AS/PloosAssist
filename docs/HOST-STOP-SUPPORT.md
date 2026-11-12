# Customer Stop button (PR #39)

The local Norwegian consent page now displays **Stopp supportøkten** independently of pending approval requests. Pressing it immediately disables all browser consent actions and polling, then POSTs to the token- and Origin-protected loopback endpoint `/api/consent/stop`.

The Linux host CLI routes the stop callback through the sole WebSocket-owning event loop, sends `session_revoke` with the host's independent capability, and disconnects its local WebSocket regardless of whether the remote server can be reached. The server-side PR #38 coordinator revokes pairing/consent and closes both session WebSockets when the command arrives.

**Limitations:** The server intentionally closes the socket without an acknowledgment. A successful local write therefore **does not prove that server-side revocation completed**. A disconnected host cannot guarantee the remote server received the command. The UI reports local shutdown only, not verified remote revocation. Full server-side integration and race tests, revocation confirmation, and actual screen/media/input shutdown must precede production use. No screen sharing or remote input is enabled.

## Revocation acknowledgment (PR #40)

The server now sends `session_revoked` **after committing** in-memory revocation and immediately before closing all session sockets. The host client waits up to five seconds for this acknowledgment and reports success only when it is received. If the network fails before acknowledgment, the host still disconnects locally but reports remote status as unconfirmed. An ACK confirms server-side revocation state was committed, not delivery of a separate media shutdown (media is not implemented).

Integration tests cover WebSocket host revocation, host and technician socket closure, rejected reenrollment, and HTTP revocation closing an authenticated socket. Run CI race tests before merging.
