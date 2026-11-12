# Shared pairing and WebSocket state

The HTTP pairing API and authenticated WebSocket server are now assembled with the **same in-memory pairing store** by `handler()`. The server initializes a trusted `signaling.Enrollment` component against that store and its verified connection registry and role registry. An integration test checks that pairing capabilities created by the HTTP-side store can authorize role bindings for verified connections in the WebSocket-side registry.

**Not yet exposed:** the WebSocket read loop still rejects all application messages, including role enrollment. There is no new network endpoint, consent delivery, signaling or remote control. The test exercises in-process enrollment, not an actual role enrollment over a socket.

**Security work remaining:** authenticated socket-derived role enrollment, synchronous revocation propagation to role and consent state, transport shutdown on revocation, and host-visible approval decisions. The existing HTTP pairing approval endpoint must not be mistaken for the host's authenticated on-screen consent.
