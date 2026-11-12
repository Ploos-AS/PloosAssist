# Local-only screen preview (PR #51)

The Linux host can now inspect its own captured screen through a **loopback-only, token-authenticated PNG endpoint**. The preview appears in the local consent UI when native PipeWire capture is running.

- `GET /api/screen-share/preview.png` uses the existing loopback, Host and token guard and sets `Cache-Control: no-store`. It rejects requests after Stop.
- The host reads an isolated BGRA snapshot, validates its layout, converts to PNG in memory and returns the PNG only to the local browser.
- The browser fetches images with the secret token in a header (never a URL), displays them as temporary object URLs and revokes them on replacement, Stop and tab hiding.
- No remote media protocol, file recording, or technician screen view is introduced.

Build the host with `CGO_ENABLED=1 go build -tags pipewire ./cmd/host` after installing `libpipewire-0.3-dev`, then follow the existing loopback host workflow. Click **Del skjermen min**, approve the Linux desktop prompt, and verify the local preview. Click **Stopp supportøkten** and confirm the preview disappears.

**Experimental:** Manual GNOME/KDE Wayland verification is required. The preview is a series of PNG snapshots, not an optimized video renderer. Sensitive screen content is processed in the host process and browser memory; avoid sharing the loopback browser session. No pixels are sent to a technician.
