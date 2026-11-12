# Linux local browser consent dialog (PR #37)

Start the loopback development server and create a pairing with the local HTTP API. Run the host with `go run ./cmd/host -session <pairing-id> -host-capability <host-capability>`. Open the printed **127.0.0.1** local URL in the browser on the **same Linux host**.

The Go host process serves a bundled Norwegian consent page, stylesheet and JavaScript. A random per-process token is inserted in the local HTML document (never URL or logs); same-origin JavaScript polls the token-protected consent API every two seconds and presents technician fingerprint and session. **Godkjenn** and **Avslå** send explicit decisions to the local bridge, which serializes decisions through the authenticated WebSocket owner. Failed requests disable the controls.

The page is served only on loopback, checks the Host header to resist DNS rebinding, and applies no-store, no-referrer, frame denial and restrictive CSP. This is still a development prototype: any local process or browser extension with access to the host user's browser may access the local UI; it does not provide OS-level user isolation or prove that the human user is physically present. Active-session revocation, comprehensive lifecycle testing and real media/input remain outstanding. Never expose this HTTP endpoint outside loopback.
