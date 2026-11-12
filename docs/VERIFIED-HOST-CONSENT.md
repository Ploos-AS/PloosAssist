# Verified host consent coordinator

This internal coordinator translates **authenticated connection IDs** into verified public-key fingerprints before invoking the consent policy. Enrollment requires a live host and technician connection belonging to the same session. Host approval and revocation derive the host identity from the connection registry; a technician cannot approve on behalf of the host.

## Critical boundary

The coordinator **does not establish human consent** or validate that the person operating the host actually clicked an approval button. A caller must invoke `ApproveHost` only after an explicit trusted host-side UI action on the same authenticated socket. The caller must also establish the host/technician roles through trusted session orchestration; roles cannot be assigned from arbitrary client parameters.

No public endpoints are introduced. The current WebSocket handler still rejects application messages. The coordinator and router are not yet wired to it. The next integration must enforce host approval and active session membership on the real socket and must recheck consent at message delivery to prevent queued messages leaking after revocation.
