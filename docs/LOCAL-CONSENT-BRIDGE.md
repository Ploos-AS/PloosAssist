# Local consent bridge — groundwork (PR #35)

A Go-only loopback HTTP bridge now exposes `GET /api/consent` and `POST /api/consent/decision` for a future **same-origin** local consent UI. The trusted host WebSocket owner supplies pending requests and a decision callback. The browser may only decide on the exact pending ticket, with an explicit boolean; the server serializes decisions and consumes the ticket before invoking the callback. Failed callbacks fail closed.

Requests require exact loopback peer, exact host, exact Origin and a random per-process 256-bit token in `X-PloosAssist-Token`. There is no CORS opt-in. The token must be bootstrapped only by the trusted local host UI renderer, never embedded in URLs, logged, or sent to remote websites.

**Not connected:** the host CLI does not start this HTTP bridge, no HTML is served and the browser adapter is not installed automatically. This PR establishes a restricted local boundary only. A future PR must implement secure same-origin asset serving, token bootstrap, polling, WebSocket ownership and lifecycle tests. Local malware or other processes running as the same OS user remain outside this prototype's security guarantees. Screen sharing and remote input remain disabled.
