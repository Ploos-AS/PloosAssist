# Linux screen capture stabilization and manual acceptance (PR #52)

This PR tightens the existing local-only preview behavior and adds deterministic tests for Stop/preview races. It does **not** claim that a graphical desktop test was performed.

## Automated verification

The CI jobs must both pass:

- Default Go vet, race tests and build.
- Opt-in native PipeWire Cgo build/test with `libpipewire-0.3-dev`.
- The new preview tests verify that a snapshot is unavailable before capture, remains unavailable after Stop, and cannot escape if Stop occurs while the bridge callback is running.

## Manual Wayland acceptance checklist

Use a disposable Linux test account, not a production customer machine.

1. Install PipeWire, the desktop portal and the GNOME or KDE portal backend. Verify `XDG_SESSION_TYPE=wayland`.
2. Build with `CGO_ENABLED=1 go build -tags pipewire -o ploosassist-host ./cmd/host`. Start the existing loopback development server and host with an authorized test session.
3. Open the host's local consent URL. Confirm that no preview appears before explicit approval.
4. Click **Del skjermen min**, select one monitor in the desktop portal, and approve. Confirm a real local preview appears and its colors/orientation match the monitor.
5. Check multiple resolutions, fractional scaling, and a window with moving content. Observe CPU/memory growth for 10 minutes.
6. Click **Stopp supportøkten** while the preview is updating. Confirm the preview disappears, subsequent PNG requests fail, the portal indicator closes, and no new frames are captured.
7. Repeat with portal denial, cancellation, closing the browser, and disconnecting the server. Confirm no unintended capture or remote transmission.
8. Repeat on GNOME and KDE. Record distribution, desktop version, PipeWire version, and observed failures.

**Known limits:** No actual GUI acceptance has been run as part of this PR. No technician video view, recording, unattended access, or remote control is implemented. The native consumer remains experimental and requires platform testing.
