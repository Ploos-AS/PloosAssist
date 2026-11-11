# PloosAssist

**Open-source remote assistance. Your infrastructure. Your control.**

PloosAssist is an early-stage, self-hostable remote support project by **Ploos AS**, intended as an open alternative to TeamViewer.

> **Status:** Architecture and project foundation only. No working remote-control functionality yet. Do not deploy for customer support.

## Principles

- **MIT-licensed original code**; third-party dependencies retain their own licenses.
- **Go-first** server, protocol, and Linux operator implementation.
- **Linux is the operator platform**. Other platforms only need to be support targets, not controllers.
- **Support targets:** Linux, Windows, macOS, ChromeOS, Android, iOS/iPadOS.
- **Remote control when permitted; otherwise consent-based screen viewing.**
- **Self-hostable by design**, without requiring a Ploos account or Ploos Cloud.
- **Explicit user authorization** and secure session lifecycle; no hidden access or unauthorized control.

## Planned platform capabilities

| Support target | Screen sharing | Remote input | Notes |
| --- | --- | --- | --- |
| Linux | Planned | Planned | X11 and Wayland need separate integration paths. |
| Windows | Planned | Planned | Privileged desktops/UAC require special handling. |
| macOS | Planned | Planned | Screen capture and accessibility permissions required. |
| ChromeOS | Planned | Best effort / restricted | Browser sharing first; system-wide input may be unavailable. |
| Android | Planned | Device-dependent | MediaProjection; remote input may need accessibility permissions/policy review. |
| iOS/iPadOS | Planned | Not planned | Broadcast-based screen sharing; no general third-party remote input. |

All rows describe **future targets**, not existing functionality.

## Proposed architecture

- Go-based signaling and session orchestration.
- WebRTC transport using Pion where appropriate; ICE/STUN/TURN for connectivity.
- A Linux-only technician console.
- Native support-host components using the minimum platform-specific code required.
- A browser-based sharing option where supported by platform APIs.
- Operator-to-host consent and identity verification, with short-lived session credentials.

See [Architecture](docs/ARCHITECTURE.md), [Security](docs/SECURITY.md), [Roadmap](docs/ROADMAP.md), and [Contributing](CONTRIBUTING.md).

## License

The project's own source code is licensed under the [MIT License](LICENSE). Dependencies must be reviewed for compatible redistribution terms.
