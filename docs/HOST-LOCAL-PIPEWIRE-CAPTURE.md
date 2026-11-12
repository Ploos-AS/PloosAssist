# Linux host local PipeWire capture (PR #50)

This PR connects the existing opt-in native PipeWire backend to the **Linux host CLI**. The customer still initiates capture with the local **Del skjermen min** button and must approve the Linux desktop portal's monitor-selection dialog.

In native builds, the host uses the same portal session to retrieve its approved stream node ID and PipeWire remote FD, then runs the native video consumer. A local in-memory buffer receives frames. The browser reports **capturing** only after a frame has appeared in the buffer. No frames are sent to the technician or written to disk.

## Build and test

```sh
sudo apt-get install libpipewire-0.3-dev pkg-config gcc
CGO_ENABLED=1 go test -tags pipewire ./...
CGO_ENABLED=1 go build -tags pipewire ./cmd/host
```

Launch the development server and host with the existing loopback session/capability workflow. Open the printed local consent UI, click **Del skjermen min**, and approve the desktop monitor-selection dialog. The UI should transition from waiting to capturing when frames arrive. Click **Stopp supportøkten** to cancel the native consumer and close the portal session. Test denial and disconnect as well.

**Default builds** without `-tags pipewire` retain permission-only behavior and do not capture pixels.

## Security and remaining limitations

- No media transport, recording, remote input, unattended access or background capture.
- Native Cgo video capture requires real GNOME/KDE Wayland tests; passing CI only proves compilation and unit tests.
- The current `LocalCapture.Run` clears the frame buffer when the consumer exits, so the UI reports activity but does not expose a preview.
- Native stream errors currently map to a generic error state. Add a richer state/error channel and live preview before treating the feature as production-ready.
- Native video callback safety, supported buffer types and portal cancellation should be audited on real desktop hardware.
