# ADR: Bound per-connection P2P send buffers

## Context

P2P channels limit their send queues by message count, but connections have no limit on total buffered bytes. A slow or unresponsive peer can therefore cause unbounded memory usage.

A byte limit introduces another problem: block synchronization allows up to 20 outstanding requests per peer, which can exceed a reasonable send buffer budget. Dropping responses causes the requester to time out and disconnect because it does not retry individual requests.

Block responses must therefore wait for available capacity without blocking the connection's receive routine, which must continue processing control messages such as pings.

## Decision

Introduce `MaxSendBufferBytes` as a per-connection limit shared across all P2P channels.

- The default is 20 MiB, enough for two maximum-size block responses.
- Zero disables the limit. Negative values and positive values below `MinSendBufferBytes` (8 MiB + 64 KiB) are rejected.
- `TrySend` fails immediately when either the byte budget or channel queue is full.
- `Send` waits for both, subject to `defaultSendTimeout`. Messages larger than the total budget fail immediately.
- Buffer capacity is reserved atomically and released once the message's last packet is handed to the connection writer. A write failing mid-message releases nothing; the connection is stopping by then.

Apply the same minimum-size requirement to `MaxRecvBufferBytes` so a connection can receive at least one maximum-size block response.

### Block response handling

Process block requests asynchronously using a dedicated response routine per peer.

The receive routine queues requested block heights instead of sending responses directly. The response routine processes requests in order and retries timed-out sends until they succeed or the peer stops. This prevents buffer pressure from dropping responses or blocking incoming control messages.

The request queue is bounded by `maxPendingRequestsPerPeer`, matching the maximum number of outstanding requests from an honest peer. Excess requests are dropped.

The response routine recovers a panic from loading a block, such as on an unreadable block store record, and stops the requesting peer. That is what the receive routine's recover did when it answered requests itself; without it, any peer asking for that height would crash the node.

Block requests that `TrySend` cannot queue are retried every 100ms until they are queued or the peer is gone. `poolRoutine` serves every peer, so it does not block on one.

## Alternatives considered

- **Larger send buffers:** Sizing the buffer for all outstanding block responses would significantly increase per-peer memory usage. Sizing it against the current `MaxDataBytes` would also become insufficient if the block size increases.
- **Blocking sends in the receive routine:** Prevents timely processing of control messages and can still lose responses when sends time out.
- **Retrying requests on the receiving side:** Does not address receive routine blocking and would require changes to existing peers.
- **Per-channel quotas:** Adds complexity. A shared budget preserves the existing channel scheduler.

## Consequences

Outgoing payload buffering is bounded per connection, regardless of peer behavior. Framing overhead, writer buffers, and kernel socket buffers remain outside this limit.

Block responses may be delayed under backpressure, but the receive routine remains responsive. Each peer has one additional response routine that exits when the peer or reactor stops.

A shared budget also means bulk traffic can temporarily consume capacity needed by higher-priority channels.