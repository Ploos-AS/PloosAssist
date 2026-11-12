# Linux host: local consent API diagnostic (PR #36)

Run the development server and create a pairing through the loopback-only HTTP pairing API. Start the host with the resulting **host** capability:

`go run ./cmd/host -session <pairing-id> -host-capability <host-capability>`

The host authenticates its ephemeral Ed25519 key, enrolls the host role, binds a randomly selected loopback HTTP port and polls for consent requests every two seconds. The bridge retains the pending request and routes an explicit decision through a single WebSocket-owner event loop, avoiding concurrent socket readers.

**Not a browser UI yet.** The API requires a secret token which is deliberately never printed, logged or placed in a URL. No trusted HTML token-bootstrap mechanism is implemented, so this is a diagnostic integration foundation, not an end-user usable consent flow. Do not weaken the token requirement to make the page work. The existing static HTML is still disconnected. Future work: authenticated same-origin local asset serving and token provisioning, browser integration, shutdown/revocation tests and secure lifecycle handling.

Only loopback development is supported. No screen sharing, remote input or signaling.
