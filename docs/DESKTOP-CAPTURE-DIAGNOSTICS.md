# Linux desktop smoke-test preparation (PR #53)

This PR adds a **read-only** host diagnostic mode. It does not run an actual desktop smoke test, request permission, capture frames, or contact the signaling server.

## Run on your Linux desktop

From a checkout of this branch:

```sh
go run ./cmd/host -capture-diagnostics
```

For the opt-in native backend (requires `libpipewire-0.3-dev`, `pkg-config`, a C compiler and a real Wayland desktop):

```sh
CGO_ENABLED=1 go run -tags pipewire ./cmd/host -capture-diagnostics
```

The command prints the session type, desktop name, whether the session bus and Wayland display are set (without exposing their values), executable availability, and whether the native backend is compiled. It **does not** prove that the portal works or that PipeWire delivers video.

After the diagnostic reports prerequisites, follow [LINUX-CAPTURE-ACCEPTANCE.md](LINUX-CAPTURE-ACCEPTANCE.md) for a real manual test. Record the desktop version, PipeWire version, portal backend, and the result of approving and stopping local screen capture. No technician media transmission is implemented.

CI covers only the diagnostic command's safety properties and build compatibility. Real GNOME/KDE Wayland capture must be tested on a desktop.
