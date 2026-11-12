# Linux portal status reporting (PR #45)

The Linux host's portal action now tracks one of `idle`, `waiting`, `approved`, `denied`, `error`, `stopped`. The local token-guarded endpoint `GET /api/screen-share/status` exposes only the current status, and the Norwegian browser UI polls it every two seconds.

The status `approved` means the desktop portal returned a successful ScreenCast Start response, **not** that any pixels are captured, encoded or sent. `denied` covers rejected/malformed portal responses, while unavailable desktop services or runtime errors map to `error`. The request is one-shot: retry requires a fresh host process. Stopping a support session permanently transitions to `stopped`, even if an outstanding portal callback returns later.

This is a diagnostic prototype. Desktop-specific GNOME/KDE Wayland testing, more precise portal error classification and PipeWire capture remain future work.
