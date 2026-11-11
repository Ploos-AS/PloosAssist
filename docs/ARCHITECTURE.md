# Architecture (proposed)

PloosAssist is a **Go-first, MIT-licensed** remote assistance system, built independently rather than forking copyleft-licensed remote desktop implementations.

## Roles and data flow

```text
Linux technician console  <-- signaling -->  Self-hosted Go server  <-- signaling --> Support host
          |                           (no plaintext screen data)                 |
          +--------- WebRTC peer-to-peer, or TURN relay when necessary ----------+
```

The signaling service brokers session establishment and authorization. Media and input should travel over authenticated encrypted channels. TURN relays encrypted packets when direct connectivity fails.

## Component boundaries

1. `cmd/server`: signaling, short-lived pairing, session state, and future organization services.
2. `cmd/console`: Linux technician UI and display/control interaction.
3. `internal/protocol`: versioned message schemas with explicit compatibility policy.
4. `internal/session`: authorization, expiry, and revocation state machine.
5. `internal/transport`: WebRTC data/media transport abstraction; investigate Pion (MIT).
6. `platform/`: capture and input implementations specific to each support target.

These are **planned boundaries**, not implemented packages.

## Platform strategy

- Linux operator console only. A Linux support host is also required.
- Windows: screen capture and input using native APIs.
- macOS: native screen capture/accessibility authorization.
- ChromeOS: consent-based browser screen sharing first. Do not promise arbitrary OS-level input.
- Android: screen capture with user authorization; input only if permitted by device and platform rules.
- iOS/iPadOS: user-initiated screen broadcast/viewing, not general remote control.
- Avoid relying on Linux containers on ChromeOS as a route to control the entire ChromeOS desktop.

## Deliberate exclusions from the first prototype

- Persistent unattended access.
- Cloud-only services and mandatory accounts.
- Session recording, mass device management, or remote shell.
- Claiming support for platform capabilities we have not demonstrated.

## Open design questions

- Optimal screen media pipeline, encoder availability, and latency.
- Linux Wayland portal / PipeWire approach versus X11.
- Security identity/trust bootstrap for support hosts and technicians.
- TURN deployment defaults, authentication, bandwidth limits, and abuse controls.
- Packaging and release signing across operating systems.
