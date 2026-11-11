# Development API guardrails

The pairing HTTP API remains **loopback-only and development-only**.

This change adds:
- An explicit check preventing the server from binding to non-loopback addresses.
- A global, in-memory request budget (120 requests/minute) for the entire HTTP handler.
- Rejection of requests with an `Origin` header (no browser client support yet).
- Cache, content-type-sniffing and referrer restrictions on responses.
- Tests for rate limiting, origin rejection and bind-address validation.

These are defense-in-depth measures, **not** production security. A reverse proxy, SSH tunnel or local untrusted process can still reach the API. The global rate limit can also cause denial of service and resets on restart. No verified technician identity, authenticated public signaling channel, end-to-end peer binding or separate screen/input permissions exist yet.

Do not deploy this API on a public host or use it to assist customers. A future internet-facing signaling API must be designed around authenticated client identity, anti-enumeration/rate limiting, TLS, replay protection, and authorization bound to a secure media channel.
