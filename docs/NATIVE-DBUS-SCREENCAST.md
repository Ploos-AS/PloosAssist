# Native Linux ScreenCast D-Bus adapter (PR #43)

The Linux-only `DBusPortal` uses `github.com/godbus/dbus/v5` to connect to the user's **session D-Bus** and invoke `org.freedesktop.portal.ScreenCast`. It implements the full portal permission handshake: `CreateSession`, `SelectSources` (single monitor only), `Start`, and `Session.Close`.

Each request subscribes to asynchronous `org.freedesktop.portal.Request.Response` signals **before** invoking the portal method and checks the returned request object path, response status and result type. A cancelled request attempts to close the outstanding portal Request. Stop cancels an in-progress desktop prompt without waiting for the portal callback, and closes any established portal session. The controller is one-shot to avoid concurrent capture attempts.

**This is an opt-in adapter, not yet wired into the Linux host UI or CLI.** The existing `Controller.Start` remains nonfunctional. A caller must construct `NewDBusPortal()` and pass it into `PortalCapture.Start` after a distinct explicit screen-sharing consent action. Portal approval may create a desktop capture grant but **no frames are read or transmitted**. The code does not call `OpenPipeWireRemote`, consume PipeWire streams, or expose remote input.

### Before production

Test actual GNOME/KDE Wayland desktops with a running session bus, verify desktop portal versions and stream response types, test cancellation/denial/desktop logout and session revocation. The presence of binaries on PATH is not proof of a working portal. Ensure the app cannot silently retry or fall back to unrestricted X11 capture. Require TLS and authenticated, consent-bound media transport before transmitting any pixels.

CI runs Go tests with race detection; integration tests that show real desktop prompts still require a graphical Linux test machine.
