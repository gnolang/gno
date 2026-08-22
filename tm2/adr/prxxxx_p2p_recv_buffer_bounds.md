# ADR: Bounding per-connection recv-buffer growth in the p2p layer

Filed as `prxxxx_` because the eventual gnolang/gno PR number is not yet known;
the change lands via gnolang/gno-fixes#26.

## Context

`MConnection.recvPacketMsg` in `tm2/pkg/p2p/conn/connection.go` assembles
partial `PacketMsg`s (`EOF=0`) into each channel's `recving` buffer, bounded
only by that channel's `RecvMessageCapacity`. Nothing bounded how long an
incomplete message could sit there. A peer could start a message on every
channel, never send `EOF`, and stay alive by answering pings, pinning the sum
of every channel's `RecvMessageCapacity` for the life of the connection:

| channel | `RecvMessageCapacity` |
|---|---|
| blockchain | `MaxBlockSizeBytes + 5` ≈ 100MB |
| consensus (4 channels) | 1MB each |
| mempool | 21MB (`defaultRecvMessageCapacity`) |
| discovery | 5MB |

That is ~130MB per connection, and `MaxNumInboundPeers` defaults to 40, so
~5.2GB from unauthenticated inbound peers — enough to OOM a node. The recv
rate limiter only throttles throughput (its return value is discarded); it
never caps accumulated bytes. Separately, `recving[:0]` on completion retained
the grown backing array, so one large message pinned its buffer for the life of
the connection.

Reproduced at the unit level (partial floods pin the buffer, the connection
stays alive, `onReceive` never fires) and, per the originating finding, against
a live single-validator node (~600MB across 2 connections).

Investigating the exposure surfaced four adjacent defects, all fixed here
because the DoS analysis depends on them: the envelope that made the blockchain
channel 100MB, the `MaxDataBytes` bounds that envelope now relies on, a
proposal-block sizing bug in the same constant, and the removal of the guard
that let one host hold every inbound slot.

## Decision

### 1. Bound the recving buffers (`p2p/conn/connection.go`)

Three defenses in `recvPacketMsg`:

- **`RecvAssemblyTimeout`** (default 30s): a per-channel deadline anchored to
  the *first* partial packet of a message and never reset by later packets, so
  an incomplete message cannot be held past the deadline by dribbling. On
  expiry the connection is torn down via `stopForError`, which closes the
  connection and unblocks the read — no dependency on `sendRoutine`.

  The deadline measures time the *peer* had to make progress, not wall clock.
  `recvRoutine` calls `onReceive` inline — the code says so — and real reactors
  block there for unbounded time: `mempool.Reactor.Receive` waits on the mutex
  `BlockExecutor.Commit` holds across `CommitSync` and a recheck of the whole
  mempool, and the consensus reactor does a blocking send into a `peerMsgQueue`
  that `receiveRoutine` may not be draining. Nothing on the connection is read
  meanwhile, so charging that time to a peer half way through a message on
  another channel drops an innocent peer for our own stall — and it drops it
  exactly when the node is already behind and needs its peers. Reproduced: with
  a 300ms deadline and a reactor that blocks for 1.2s, an honest peer's
  multi-packet message was torn down with `recv assembly timeout: channel 2 did
  not complete message within 300ms`.

  So `recvRoutine` publishes when it enters a callback (`recvStallSince`) and
  credits the elapsed time to every armed deadline when it leaves. A timer that
  fires mid-stall sees the stall in progress and re-arms; a timer that fires
  after one sees the moved deadline and re-arms. Credit only ever flows from
  *our* stalls, so the anti-dribble property is untouched —
  `TestRecvAssemblyTimeoutNotResetByDribble` still passes, and
  `TestRecvAssemblyDeadlineNotChargedForReactorStall` additionally asserts the
  deadline still fires for the peer's own next unfinished message.

  This is the one piece of recv state that is not `recvRoutine`-only, since the
  timer goroutine now re-arms rather than only tearing down, so the deadline sits
  behind a per-channel mutex. Re-arming on the timer's own schedule (rather than
  `Reset`ting from `recvRoutine`) costs at most one spurious wakeup per credited
  stall and avoids racing `Reset` against the firing callback; `tm2/pkg/p2p/...`
  is clean under `-race -count=3`.
- **`MaxRecvBufferBytes`** (default 20MB): a total per-connection budget across
  all channels, on top of the per-channel cap. Tracked in `recvRoutine` only,
  so it needs no locking.
- **Free the backing array** once it has grown past `RecvBufferCapacity` *and*
  the message that just completed no longer needed it. A channel steadily
  carrying large messages keeps its array and stays allocation-free; one that
  saw a single outsized message hands the memory back on its next ordinary
  message. See alternative E for why neither an unconditional free nor a free
  keyed only on the grown capacity works.

Both limits have live defaults in `DefaultMConnConfig`, which `MConfigFromP2P`
starts from and only partially overrides, so they are active in production.

The budget is sized against real traffic rather than against the channel-cap
sum, because it is the only bound on this memory and therefore sets the node's
aggregate worst case. A connection assembles at most one incomplete message per
channel, and the largest legitimate ones total ~13MB with `MaxDataBytes` at its
ceiling (8MB blockchain + 4×1MB consensus + 1MB mempool + KBs of discovery), or
~7MB at the 2MB default.

### 2. Right-size the fast-sync envelope (`bft/blockchain/reactor.go`)

`maxMsgSize` — both the decode limit and the channel's `RecvMessageCapacity` —
was `MaxBlockSizeBytes` (100MB), the single largest contributor to the
exposure. It is now `MaxBlockDataBytesLimit + 64KB`.

The 64KB covers only the `bcBlockResponseMessage` envelope wrapped around an
already length-prefixed block (measured: 25 bytes). It deliberately does *not*
budget for the header or `LastCommit`, because `MaxDataBytes` is also the max
size `ConsensusState.addProposalBlockPart` passes to amino, so a block is only
committable if its whole serialized form already fits in `MaxDataBytes`.
Deriving from the ceiling rather than the live value covers blocks committed
before `MaxDataBytes` was lowered.

### 3. Bound `MaxDataBytes` at both ends (`bft/types/params.go`)

`ValidateConsensusParams` now rejects `MaxDataBytes > MaxBlockDataBytesLimit`
(8MB), without which governance could raise the block size above the envelope
and produce blocks that fast-syncing peers reject. 8MB is 4x the current
default and every `MaxDataBytes` in the tree is 2MB, so no deployment is
affected. `MaxBlockSizeBytes` is left alone; it also drives the consensus
block-part count.

It also rejects non-positive values, which were reachable and fatal: `0` panics
the proposer in `ReapMaxBytesMaxGas`, and a negative value disables the reaping
limit entirely (bypassing the ceiling above) while panicking amino, which
rejects a negative `maxSize`. `ConsensusParams.Update` copies the whole `Block`
subparam, so a genesis that sets the other block fields but omits
`max_data_bytes` decodes to 0 — validated clean, then halts at the first
proposal with `CONSENSUS FAILURE`.

Finally, it requires `MaxTxBytes` to leave `MaxBlockOverheadBytes` (128KB) free
inside `MaxDataBytes`. `MaxTxBytes` was bounded only by `MaxBlockSizeBytes`
(100MB), so `MaxTxBytes >= MaxDataBytes` was a legal configuration — and a fatal
one, because `MaxDataBytes` bounds the *whole serialized block*. A tx whose raw
size fits `MaxTxBytes` but whose framed size does not fit `MaxDataBytes` is
admitted by `CheckTx`, reaped on its own (`ReapMaxBytesMaxGas` stops at the first
tx that does not fit rather than skipping it), and then trimmed straight back out
by the loop in decision 4. Measured with `MaxTxBytes = MaxDataBytes = 64KB` and
one 64KB tx followed by a 512-byte one: before this bound, every proposal came
out empty (193 bytes) with both txs still in the mempool, at every height,
forever, and nothing logged it — the chain accepts no txs at all and looks
healthy. Prior to decision 4 the same configuration produced a 65776-byte block
that every peer rejected, i.e. a loud stall rather than silent starvation; the
trim loop would otherwise have converted one into the other.

128KB is sized from measurement: a serialized block costs 428 bytes empty plus
~167 bytes per validator in its `LastCommit`, plus 44 bytes of framing per tx, so
it covers a commit for roughly 780 validators. It also happens to be what makes
the 20MB recv budget sound: the worst *legal* concurrent assembly is
`8MB (blockchain) + 4MB (consensus) + (8MB - 128KB) (mempool) + discovery`, which
`TestDefaultBudgetCoversWorstLegalConfig` measures at 20,908,032 bytes against
the 20,971,520 budget. That is only 63KB of headroom, and it exists only because
`MaxTxBytes` cannot reach `MaxDataBytes`; the test fails if either side drifts.

One consequence to note: with `MaxTxBytes + 128KB <= MaxDataBytes <= 8MB`, the
effective ceiling on `MaxTxBytes` is 8MB − 128KB, so the pre-existing
`MaxTxBytes > MaxBlockSizeBytes` (100MB) check is now unreachable. It is left in
place rather than removed — it is cheap, and removing a consensus-validation
rule is a larger change than this ADR wants to make.

### 4. Keep proposal blocks within the decode limit (`bft/state/execution.go`)

`CreateProposalBlock` reaped up to `MaxDataBytes` of *raw tx bytes* while
`addProposalBlockPart` applies the same value to the *whole serialized block*,
so a block filling the tx budget was undecodable by every peer:

```
read overflow, maxSize is 2000000 but this amino binary object is 2000184 bytes
```

Peers prevote nil, the round fails, and the next proposer reaps the same
mempool — so the chain stalls while the mempool holds that much data, which
anyone can arrange (default mempool 5000 txs / 1GB, and the byte budget binds
before gas). Fixed by trimming txs from the tail until the serialized block
fits. Removing n bytes of tx data removes at least n bytes from the block, so
trimming by the excess converges. `PartSet.ByteSize` reports the size without
an extra marshal, since `MakePartSet` already built the set from that encoding.

### 5. Restore the same-IP guard (`p2p/config`, `p2p/switch.go`)

The only dedup on an inbound connection was by peer ID, which is a
self-generated node key, so one host could mint identities and occupy all 40
inbound slots — verified with a real switch and real dials. That is what makes
the per-connection budget the wrong unit for the aggregate.

`AllowDuplicateIP` returns, defaulting to `false`. tm2 had this before #2852,
also on by default; that PR removed it as a "useless flag" in one line of the
description, with no discussion, in a change reviewers said was too large to
review thoroughly — and the test2–test5 deployment configs still carry a
now-silently-ignored `allow_duplicate_ip = false`.

Rather than resurrect that PR's `ConnSet`/`resolveIPs`/goroutine-per-filter
machinery, the check sits beside the existing duplicate-ID check in the accept
loop and reads `PeerConn.RemoteIP()`. Inbound only: no DNS resolution is needed
for a socket peer, and our own dials are explicitly configured. Local clusters
share the loopback address, so it is lifted in the bft test config and the
internal p2p test-cluster helper.

A rejected connection is closed explicitly (`p.CloseConn()`) rather than only
dropped from the transport. `transport.Remove` deletes an `activeConns` entry and
nothing more, and the peer was never started, so no `Stop()` path runs either —
the socket the STS handshake just established would be closed only when the
`netFD` finalizer ran. Measured with the GC disabled, 20 connections from one
host left 20 sockets `ESTABLISHED` on the victim instead of 1. That matters more
here than for the max-inbound and duplicate-ID branches that share the shape,
because this branch is reachable from a peer's *second* connection, so an
attacker can open sockets faster than the GC reclaims them.
`TestSwitchClosesRejectedDuplicateIPConn` observes the close from the dialer's
side and fails (by timeout) without it.

`recv_assembly_timeout` and `max_recv_buffer_bytes` are exposed on `P2PConfig`
and copied through `MConfigFromP2P`. They carried toml tags from the start but
had no `P2PConfig` counterpart, so they were not actually reachable from
`config.toml`; an operator whose peers are being dropped mid-transfer needs to be
able to raise the deadline without a recompile. Both accept 0 to disable, which
is what the `MConnConfig` fields already meant.

## Alternatives considered

**A. Reset the assembly deadline on each partial packet** — rejected. An
attacker dribbles packets to keep resetting it and the timer never fires. This
was the flaw in the originally proposed patch for this finding, along with a
`MaxTotalRecvBuffer` that defaulted to 0 and was not mapped in
`MConfigFromP2P`, making it dead code in production.

**B. Raise the proposal decode limit instead of trimming** — rejected. It would
make upgraded nodes accept blocks un-upgraded nodes reject, needing a
coordinated upgrade to avoid a split. Trimming needs none, and old peers accept
the smaller blocks unchanged.

**C. Use `BlockParams.MaxBlockBytes` as the whole-block limit** — it looks like
the field the decode should have used, but it is vestigial: never defaulted,
never validated, never read. Wiring it up would change
`ConsensusParams.Hash()` and so every block's `ConsensusHash`, breaking
existing chains.

**D. Re-reap with a reduced budget rather than trimming** — rejected. It returns
the same tail txs (`ReapMaxBytesMaxGas` walks in order and stops) but adds a
second mempool read that a concurrent `CheckTx` could change under the
measurement.

**E. Free the recving array on every completed message** — rejected on cost, in
two rounds.

Freeing it *unconditionally* allocates `RecvBufferCapacity` per message, which is
200KB on the consensus data and blockchain channels: 20931 ns/op and 204936 B/op
against 1070 ns/op and 136 B/op when the free is conditional.

Making it conditional on `cap(recving) > RecvBufferCapacity` alone is not enough,
which the first version of this change got wrong. That condition holds for
*every* message larger than `RecvBufferCapacity`, so it puts a
realloc-and-regrow on the path of exactly the messages that matter most:
blockchain and consensus data configure 200KB against multi-MB blocks, and the
mempool channel leaves it at the 4KB default against `MaxTxBytes`-sized txs.
Measured on a 2MB message with a 200KB capacity, and on a 100KB tx with the 4KB
default:

| | ns/op | B/op | allocs/op |
|---|---|---|---|
| 2MB message, free when grown | 2500106 | 9068605 | 10 |
| 2MB message, array reused | 47649 | 0 | 0 |
| 100KB tx, free when grown | 17692 | 176129 | 3 |
| 100KB tx, array reused | 1792 | 0 | 0 |

The benchmark that originally justified the conditional free used a 64KB message
on a 200KB channel — under the capacity, so it took the reuse path either way and
could not see the regression (1135 ns/op and no allocations in both).

Hence the extra `len(msgBytes)*2 < cap(ch.recving)` term: release the array only
when the completed message was not really using it. The residual is that a peer
which sends one max-size message per channel and then goes quiet keeps those
arrays until its next message, so retained-but-empty capacity is bounded by the
channel-cap sum (~38MB) rather than by the 20MB budget — but getting there costs
it ~38MB of *complete, reactor-accepted* messages, where the attack this ADR is
about needs no valid message at all. The 20MB budget still bounds everything
in flight.

Reuse is safe because amino's `DecodeByteSlice` copies byte slices while
decoding, so no decoded message aliases `recving`.

## Consequences

- Per-connection exposure for in-flight assembly drops from ~130MB to 20MB, and
  the slots one host can hold from 40 to 1, so a single-host worst case goes
  from ~5.2GB to 20MB. A distributed attacker still reaches 40 × 20MB = 800MB;
  the same-IP guard is single-host protection, not DDoS protection. Empty
  backing arrays retained across messages are bounded separately, by the
  channel-cap sum (~38MB) — see alternative E.
- An incomplete message can no longer be held indefinitely; the deadline is a
  throughput floor of roughly `messageSize/30s` (~68KB/s for a 2MB block, and
  ~273KB/s at the 8MB `MaxDataBytes` ceiling), so very slow peers will be
  dropped mid-transfer during fast sync. Measured against the 5MB/s default
  `recv_rate`, a 2MB message assembles in 0.4s, so the floor only binds on peers
  two orders of magnitude slower than the rate limiter allows.
- A reactor that blocks `recvRoutine` no longer costs the peer its deadline, but
  it still stops the connection being read, so the pong timeout can still fire
  during a long stall. That path is pre-existing and out of scope here.
- `recv_assembly_timeout` and `max_recv_buffer_bytes` are reachable from
  `config.toml`. Configs written before they existed keep the defaults, since
  `LoadConfigFile` decodes onto `DefaultConfig()` and absent keys are left alone.
- `Block.MaxDataBytes` is now bounded at both ends; genesis files with a partial
  `block` params object are rejected instead of halting the node later.
- Full blocks are decodable by peers; a block filling the tx budget loses the
  tail txs that the overhead displaces.
- `allow_duplicate_ip` must be set to `true` for multi-node single-host
  clusters. The stale key in the deployment configs becomes meaningful again.
- Two pre-existing p2p defects fixed in passing: `errors.As` on the
  transport-closed sentinel wrote to a package-level variable from every accept
  loop (a data race, and an over-broad match that would have turned any tm2
  error from `Accept` into a permanent exit from the loop), and two switch tests
  raced on state the accept and redial loops read. Over 30 `-race` runs of
  `tm2/pkg/p2p` the race count goes from 6 to 0. The accept-loop tests also had
  to learn to stop: their `Accept` mock ignored the context and `runAcceptLoop`
  only leaves the loop on an error, so cancelling did nothing and the loop spun
  at ~555k iterations/second for the rest of the test binary — burning two cores
  through a package full of timing-sensitive tests. The mock now returns
  `ctx.Err()`.
- Known unrelated flakiness in `tm2/pkg/bft/consensus` was hit while verifying
  and is *not* addressed here: a `FireEvent`-under-`cs.mtx` deadlock in
  `TestStateFullRound1`, a `TestHandshakeReplayNone` WAL replay failure, and a
  package-level `config` variable written by parallel tests in both
  `consensus` and `blockchain` (9 races per 5 `-race` runs on the base). These
  want their own investigation.
