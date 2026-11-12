# Host consent UI integration boundary (PR #34)

The Norwegian consent view now contains a fail-closed prompt state machine: it displays the verified technician fingerprint, session, expiration, and explicit **Godkjenn / Avslå** controls only after a local adapter supplies a valid prompt. Expired or malformed prompts keep controls disabled. Decisions disable controls immediately and show success only after the adapter confirms the server response.

The new Go `internal/hostclient.ConsentClient` implements authenticated-host role enrollment, pending approval retrieval and one-use approval/denial over the existing WebSocket protocol. It must be used sequentially on a host-owned authenticated connection.

**Critical limitation:** the browser view and Go client are **not connected yet**. `window.PloosAssistHostConsent.installTrustedAdapter(...)` is a *JavaScript integration contract*, not an authentication mechanism: any script running in the page can call it. Never load untrusted scripts, expose this view on a public site or treat arbitrary JavaScript calls as proof of human consent. A future local Go host process must own the authenticated socket and expose a secured local UI bridge with origin/CSRF protection, explicit user interaction, and revocation handling. No screen sharing, remote input or signaling is enabled.
