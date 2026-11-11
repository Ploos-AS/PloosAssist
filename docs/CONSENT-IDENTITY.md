# Host consent bound to a specific technician identity

The new `ConsentPolicy` is an **internal policy component**, not a network API.

A session records two distinct verified public-key fingerprints and an expiry. The host can approve **only the exact technician fingerprint** registered for that session. Authorization requires an approved, unexpired session and the exact host–technician pair in either direction. Revocation immediately blocks further policy checks.

## Trust boundaries

- `Create` requires independently authenticated peers and a trusted session orchestrator.
- `Approve` and `Revoke` accept an *already authenticated* host fingerprint; simply forwarding a client-supplied string is unsafe.
- The existing pairing API and WebSocket experiment are **not wired** to this policy. No signaling is enabled by this PR.
- This policy does not persist approvals across restarts, verify user interaction, or enforce network-level authorization.
- Production integration must validate the real transport connection identity and active pairing session on **every message**, and bind approved WebRTC fingerprints to the authenticated peers.

The server remains loopback-only and development-only. Do not expose it through a reverse proxy.
