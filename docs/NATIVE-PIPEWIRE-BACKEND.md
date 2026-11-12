# Experimental native PipeWire video consumer (PR #49)

This PR introduces a **real native PipeWire stream consumer** behind the explicit Linux build tags `cgo,pipewire`. It is not enabled in normal builds.

## Build

Install the Linux development package `libpipewire-0.3-dev` and a C compiler. Then build or test with:

```sh
CGO_ENABLED=1 go test -tags pipewire ./internal/screencapture/...
CGO_ENABLED=1 go build -tags pipewire ./...
```

The native consumer uses `pw_context_connect_fd` with a duplicate of the portal-provided PipeWire FD, connects to the **portal-approved node ID**, negotiates raw BGRA video, and copies validated frames into the bounded local `framebuffer.Latest` store. A small native staging buffer is polled from Go and copied again into Go-owned memory. The stream is closed and memory cleared on cancellation.

## Current limitations and review requirements

- This is an **experimental adapter**. The normal host CLI does **not** instantiate it; it must be integrated with `LocalCapture.Run` in a separate PR after native CI and manual Wayland tests.
- Only raw BGRA frames backed by mapped CPU memory are supported. DMA-BUF, alternative pixel formats, metadata transforms and multi-monitor capture are not implemented.
- Native PipeWire builds require Linux development headers; ordinary CI does not exercise the opt-in Cgo code. **Do not treat this PR as tested on a real desktop until verified.**
- No screen images are written to disk or sent over the network. The adapter has no remote input capability.
- A production implementation must test stop/revocation while native callbacks are active, PipeWire stream state/error callbacks, and GNOME/KDE compatibility.

This is the first implementation of a frame-producing native backend, not a claim of a completed remote desktop feature.
