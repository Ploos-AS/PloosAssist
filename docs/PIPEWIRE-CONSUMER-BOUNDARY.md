# Portal-authorized local PipeWire consumer boundary (PR #48)

This PR connects the existing Linux ScreenCast portal grant to a **local video-consumer interface**, but does **not** yet implement the native PipeWire video backend.

The desktop portal's `Start` response is now validated as a list of PipeWire node IDs and retained on the authorized session. `Streams()` returns an isolated copy only after successful Start. `LocalCapture.Run` selects the first approved node, opens the portal-scoped PipeWire FD, passes both to a `VideoConsumer`, and closes the FD and clears the bounded frame buffer on completion. `Stop` cancels the consumer and clears local frames.

**No video frames are captured by this PR.** PipeWire's Unix socket carries the PipeWire protocol, **not raw BGRA bytes**. A native backend must implement PipeWire protocol negotiation, node connection, SPA video format negotiation, memory mapping/buffer lifetime, frame conversion and callbacks before this can work. No such backend is supplied, and the host CLI does not invoke `LocalCapture.Run`. A nil backend fails closed with `ErrConsumerUnavailable`.

The next implementation should use a well-maintained native PipeWire binding or a separately reviewed Cgo adapter with explicit build tags and Linux CI packages. It must never attach to unapproved node IDs or expose local frames over the network without independent consent and transport security.
