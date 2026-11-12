# Host approval requests (internal workflow)

PR #20 adds a bounded, expiring host approval request state machine. A trusted orchestration layer identifies the host and technician by their **live, authenticated connection IDs**, and the resulting request contains the exact technician fingerprint for display to the host. The host can approve or deny; only an approval originating from the original authenticated host connection enrolls and approves the pair. A disconnected technician cannot be approved. Denial never grants signaling permission.

**Not yet an actual user interface or network API.** No host UI exists yet and the WebSocket handler continues to reject application messages. The future integration MUST authenticate the host UI action on the original socket, establish roles through trusted pairing, display the fingerprint clearly, require deliberate confirmation, and implement immediate revocation plus queue flushing before forwarding signaling.

This state machine is process-local and does not survive restart. Requests expire after at most 15 minutes. Do not expose the development server through a reverse proxy.
