# Host-only consent request retrieval (PR #32)

After both sides complete Ed25519 proof and capability-backed `role_enroll`, the verified technician may send `{"type":"approval_request"}`. The server derives both live connection IDs from its role registry and creates a bounded, two-minute pending consent request. It responds with `approval_requested` but **never sends the host's one-use ticket to the technician**.

The verified host may send `{"type":"approval_check"}` to retrieve `approval_pending` containing the technician fingerprint, one-use ticket ID and expiration, or `approval_none`. This is **host-initiated polling, not push delivery**. Neither request nor retrieval grants consent; all signaling and remote input remain rejected. The UI and authenticated host decision message are not implemented.

**Important:** Development is loopback-only. Revocation and concurrent request lifecycle still require hardening. The server must not be exposed for remote support until active-socket revocation, verified on-screen approval, secure transport and end-to-end tests are completed.
