# Contributing

Thanks for your interest in PloosAssist.

PloosAssist is in its initial design stage. Discuss architectural changes in an issue before implementing large features.

## Engineering expectations

- Prefer Go for shared server/client logic; use native APIs for platform-specific functionality.
- Keep cross-platform protocol semantics versioned and documented.
- Make user consent and safe termination explicit in every support flow.
- Add tests for behavior and security boundaries; don't assert feature support without testing it.
- Keep original contributions MIT-licensed. Review dependency and asset licenses before adding them.
- Never copy source code from incompatible-licensed remote desktop implementations.
- Do not commit credentials, private keys, or customer information.

A developer build guide will be added when executable components exist.
