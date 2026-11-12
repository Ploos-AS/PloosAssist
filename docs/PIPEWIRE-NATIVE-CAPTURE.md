# Experimental native PipeWire frame consumer (PR #48)

An **opt-in Linux/cgo** implementation now uses libpipewire to connect to a **portal-scoped Unix FD** and a **portal-approved PipeWire node ID** and attempts to copy **one BGRA frame** into the bounded local `framebuffer.Latest` store.

## Build prerequisites

This implementation is gated behind the `pipewire` Go build tag and requires Linux, cgo, a C compiler, pkg-config and `libpipewire-0.3` development headers:

```sh
sudo apt install build-essential pkg-config libpipewire-0.3-dev
go test -tags pipewire ./internal/pipewirecapture ./internal/framebuffer
```

Default `go test ./...` does **not** build the optional native code, and the standard CI does not validate its compilation or runtime behavior yet. It must be compiled and smoke-tested on a Linux machine before it can be relied upon.

## Scope and security

- `CaptureOne(ctx, remote, nodeID, latest)` is an internal API, **not wired into the host CLI**. It requires a portal-provided file descriptor and a node ID from a successful ScreenCast Start response.
- The native stream negotiates BGRA only; other formats and DMA-BUF-only buffers may not yield a frame. The staging buffer rejects oversized frames.
- The implementation requests one local frame only. It does not persist frames, encode video, transmit over a network or enable remote input.
- The caller must close its portal FD, close the frame buffer on Stop, and close the portal session.
- This is a first native implementation and has not yet been verified against actual GNOME/KDE portal backends. The portal Start stream metadata parser is not yet connected to the host CLI. Desktop integration and fd lifecycle tests are required before production.

## Next

Wire validated stream IDs and `OpenPipeWireRemote` into a trusted host-only capture workflow; test the native adapter with an actual portal-granted session, cancellation, and GNOME/KDE. Add a separate CI job with libpipewire development headers before any default enablement.
