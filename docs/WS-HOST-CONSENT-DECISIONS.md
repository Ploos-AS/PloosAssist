# Authenticated host approval and denial (PR #33)

Once the technician requests support and the host retrieves an `approval_pending` ticket, the **host's verified, capability-enrolled WebSocket connection** may submit:

`{"type":"approval_decide","ticket_id":"<one-use ticket>","approve":true}`

For denial, use `"approve":false`. The server derives the host connection ID from the authenticated socket, checks the session's role pair and consumes the exact one-use ticket. A successful decision returns `{"type":"approval_decided","state":"approved"}` or `"denied"`. Replay or technician-originated decisions are rejected.

**Important:** This is protocol support for a host decision, **not yet a human-visible consent UI**. It does not activate signaling, screen sharing, remote input or any technician control. The approval state exists only in memory. Concurrent revocation, active socket shutdown and a clear graphical consent prompt must be completed and tested before exposing remote support.
