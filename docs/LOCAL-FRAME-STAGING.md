# Local PipeWire frame staging foundation (PR #47)

This change adds a **bounded local frame buffer** for a future PipeWire capture adapter. It is not yet a working PipeWire frame consumer.

- `internal/framebuffer` stores at most one uncompressed BGRA frame in memory, capped at 16 MiB with dimension and stride validation.
- Each frame is copied on ingestion and on retrieval; caller memory cannot mutate the stored frame.
- Replaced frames are zeroed and `Close` zeroes and discards the last frame. After Close, no more frames are accepted.
- Race-oriented tests cover concurrent writes, snapshots and Stop.
- `internal/screencapture/streams_linux.go` introduces a strict parser for portal stream node metadata, to be used by the real PipeWire adapter in a later PR.

**No real frames are acquired.** This PR does not call PipeWire APIs, consume a PipeWire stream, capture screen pixels, write files or transmit video. The host CLI does not use the buffer. The parser is not yet wired to the ScreenCast Start response. Frame staging alone is not a capture implementation.

## Next steps

Implement a real PipeWire consumer using the portal's Unix FD and **only** the node IDs returned by the approved ScreenCast Start response. Validate negotiated video format, dimensions, stride, memory layout and buffer lifetime before copying frames. Keep all processing local initially, then bind any future media transport to authenticated session consent and revocation. Test on GNOME and KDE Wayland desktops.
