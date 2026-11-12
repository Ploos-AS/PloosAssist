# Portal-scoped PipeWire remote (PR #46)

The Linux D-Bus ScreenCast adapter now implements `OpenPipeWireRemote(ctx)` on the concrete `dbusPortalSession`. This calls `org.freedesktop.portal.ScreenCast.OpenPipeWireRemote` with the previously granted portal session handle and returns a local `*os.File` wrapping the received D-Bus Unix file descriptor.

## Security boundaries

- The remote can only be requested **after** `Start` has succeeded and the desktop portal has granted screen selection.
- Closed sessions and cancelled contexts are rejected. A descriptor received after session closure is closed immediately.
- The caller owns the returned file and **must close it**. The method does not select arbitrary PipeWire nodes, read frames, encode video, transmit media, or grant remote input.
- The method is deliberately **not wired to the host CLI** yet; no descriptor is opened during ordinary use. This allows the API to be reviewed and tested before adding a PipeWire consumer.

## Follow-up

Implement a Linux PipeWire client that consumes only portal-authorized stream node IDs and the returned remote FD, closes everything on Stop/disconnect, and never transmits media without separate authenticated session consent. Real GNOME/KDE Wayland integration testing is required because CI cannot provide a graphical portal service.
