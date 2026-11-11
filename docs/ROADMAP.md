# Roadmap

Milestones describe intended outcomes, not completed features or schedules.

## M0 — Foundation
- [x] Define product scope, license, role/platform matrix, and security principles.
- [ ] Choose concrete Go module/package layout and dependency policy.
- [ ] Add build, lint, unit test, and dependency/license CI.

## M1 — Safe connection prototype
- [ ] Minimal Go signaling server.
- [ ] Short-lived pairing with separate recipient approval.
- [ ] Linux-to-Linux screen viewing using encrypted transport.
- [ ] Explicit disconnect and consent expiry.

## M2 — Linux remote assistance
- [ ] Linux remote pointer/keyboard input, permission-gated.
- [ ] X11 and Wayland support paths evaluated separately.
- [ ] NAT traversal and TURN relay fallback.
- [ ] Usable Linux technician console.

## M3 — Desktop support hosts
- [ ] Windows screen viewing and remote input.
- [ ] macOS screen viewing and permitted remote input.
- [ ] Portable customer support experience where feasible.

## M4 — Other support hosts
- [ ] ChromeOS screen sharing, best-effort remote input only if supported.
- [ ] Android screen sharing and device-dependent controls.
- [ ] iOS/iPadOS screen broadcasting/view-only support.

## M5 — Production readiness
- [ ] File transfer and clipboard features with explicit permissions.
- [ ] Multi-monitor improvements and performance tuning.
- [ ] Signed installers, update integrity, security review, and documentation.
