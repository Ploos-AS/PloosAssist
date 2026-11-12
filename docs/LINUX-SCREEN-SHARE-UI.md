# Linux local screen-sharing action (PR #44)

The bundled local host UI adds a distinct **Del skjermen min** button. Clicking it makes a token-protected, same-origin POST to `/api/screen-share/start`. The host process owns a one-shot `PortalCapture`, starts a desktop portal request asynchronously and cancels it when the support session is stopped or the host process exits.

This button is **separate from approving a technician**. The browser cannot supply a portal session handle or arbitrary screen source. The portal itself must ask the user to select a monitor; no unattended or silent capture is permitted.

**Limitations:** A successful HTTP response means the local process accepted the request, **not** that the desktop portal prompt appeared or that the user approved it. Desktop portal failures are not yet reflected back in the browser. The Linux adapter is not wired to PipeWire frames; **no pixels are captured or transmitted**. The host CLI remains a loopback-only development tool, not a production remote support client.

## Manual desktop smoke test

On a graphical Linux Wayland session with a working XDG Desktop Portal and PipeWire service:

1. Start the loopback PloosAssist development server.
2. Create a pairing with the existing local pairing API, then launch the host CLI with its session ID and host capability.
3. Open the printed local consent dialog and click **Del skjermen min**.
4. Verify that the desktop presents its own monitor-selection permission dialog. Accept or deny it; no video should appear in either case.
5. Click **Stopp supportøkten** while a portal dialog is pending; the request must be cancelled and never resume.
6. Repeat on GNOME and KDE. Record the portal backend/version and any D-Bus errors.

Do not expose the local bridge or development WebSocket outside loopback. Before implementing media, add explicit status reporting for portal success/denial, test session cancellation on real desktops, and bind media delivery to server-side session consent and revocation.
