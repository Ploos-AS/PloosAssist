# Linux screen capture foundation (PR #41)

This PR introduces a **Linux-only, non-capturing** foundation. The `internal/screencapture` package checks whether the process appears to be running under Wayland and whether the `xdg-desktop-portal` and `pipewire` executables are on PATH. These are only diagnostic hints; presence of executables does **not** guarantee a working portal service or PipeWire stream.

The `Controller` denies capture without explicit local approval, refuses attempts after Stop, respects cancelled contexts and returns `ErrNotImplemented` even when prerequisites are present. It does **not** read pixels, open D-Bus, initiate a portal request, transmit video, or enable remote input. It is deliberately not connected to the existing approval UI: approval of a support request must **not automatically** authorize screen capture.

## Planned capture architecture

1. Establish a session-scoped, explicit **share my screen** action in the trusted Linux host UI, separate from approval of technician identity.
2. On Wayland, call the `org.freedesktop.portal.ScreenCast` D-Bus interface: `CreateSession`, `SelectSources`, `Start`; wait for the desktop's own interactive user confirmation. Bind resulting portal session to the support session.
3. Open the granted PipeWire remote, process only approved streams, and handle portal session closure, device changes, host Stop and disconnect immediately.
4. Add an authenticated, consent-bound, bandwidth-limited, encrypted media transport; no frame delivery before the consent and revocation model is integrated and tested.
5. Investigate X11 separately. Do not bypass Wayland portal selection or silently fall back to unrestricted capture.

The Go foundation has no new dependencies and intentionally cannot capture anything. It is a security boundary and API design step, **not functional screen sharing**.
