# Security model (initial design)

**This project does not yet provide secure remote support software.** The following are non-negotiable engineering requirements before any real-world support use.

## Session authorization

1. The support recipient deliberately starts a support session and sees a short-lived identifier.
2. The recipient explicitly approves the intended technician and viewing scope.
3. Remote input requires separate, visible consent where applicable.
4. Either party can immediately revoke access and terminate the session.
5. Codes expire, are rate-limited, and never function as sufficient authorization by themselves.

## Transport and identity

- Use established authenticated encryption from audited protocol implementations (for example, WebRTC DTLS/SRTP). Do not design custom cryptography.
- Authenticate signaling and bind session permissions to the media channel and peer identity; prevent code guessing, replay, and session hijack.
- Protect credentials in transit and at rest; never log secrets or access codes.
- Relay service must only relay encrypted traffic. Verify actual end-to-end properties before claiming end-to-end encryption.
- Require TLS for public server endpoints; handle ICE/TURN credentials with expiration and rotation.

## Client safeguards

- Show a persistent support-active indicator on the support target.
- Deny unattended access by default.
- Respect operating-system permissions, sandbox boundaries, and consent dialogs.
- Treat incoming input, filenames, clipboard, and session messages as untrusted.
- No hidden persistence, stealth access, security bypass, or privileged escalation.

## Release gates

Before external testing: threat model, dependency/license audit, reproducible security tests, authentication/authorization tests, and clear incident disclosure process.

Report vulnerabilities privately to the maintainers rather than opening a public issue containing exploitable details. A formal security reporting channel will be published before beta.
