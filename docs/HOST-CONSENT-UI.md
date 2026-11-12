# Host consent UI prototype

This static, Norwegian-language HTML/CSS/JS prototype is located at `web/host-consent/`. It shows the intended host-facing approval experience with a clear warning, verified technician fingerprint placeholder, session placeholder, and approve/deny buttons.

**Security:** Both buttons are intentionally **disabled**. There is no connection to WebSocket, the approval state machine, or any remote-control capability. Do not interpret this UI as a functioning consent mechanism. The prototype does not read untrusted URL parameters or pretend to authenticate a technician.

## Integration acceptance criteria

1. Provide a trusted, installed host application surface; a browser preview alone cannot prove control of the host device.
2. Establish host and technician roles through verified pairing, not client-supplied labels.
3. Deliver approval requests only to the host's authenticated, live transport; show the verified technician fingerprint and session.
4. Require an explicit click; bind the decision to the same authenticated connection and the exact request, prevent replay, and fail closed on disconnect/expiry.
5. On denial or revocation, immediately stop signaling and clear queued messages, then terminate media/input channels.
6. Add accessibility tests, keyboard navigation, screen reader announcements, and end-to-end integration tests.

The existing server remains loopback-only and development-only. This page is not served by the Go server.
