# ScreenCast portal lifecycle (PR #42)

This PR defines the Go `Portal` / `PortalSession` abstraction and tests the consent-gated lifecycle: `CreateSession` → `SelectSources` → `Start` → `Close`. Errors and cancellation close the session; Stop irreversibly disables the controller. Mock portal tests verify ordering and cleanup. No extra Go dependency is introduced.

**This PR does not yet contain a D-Bus implementation.** `PortalCapture.Start` requires an injected portal adapter and returns `ErrNotImplemented` when none is provided. A fake portal used in tests is not evidence that a real desktop consent dialog was shown. The earlier `Controller.Start` remains a nonfunctional stub. No actual screen pixels, PipeWire streams, remote-control events or media transport exist.

## Next implementation step

Use the session D-Bus service `org.freedesktop.portal.Desktop`, interface `org.freedesktop.portal.ScreenCast` at `/org/freedesktop/portal/desktop`. Implement async Request.Response handling with request/session handle tokens and sender-specific paths, await desktop selection for `SelectSources` and `Start`, reject denial and malformed responses, close the portal session on cancellation/disconnect, and use the portal's `OpenPipeWireRemote` only after successful selection. Validate against actual GNOME and KDE Wayland desktops, never silently fall back to unrestricted capture.

A separate host-side action must explicitly authorize **screen sharing**; consent to technician identity alone must never authorize pixel capture.
