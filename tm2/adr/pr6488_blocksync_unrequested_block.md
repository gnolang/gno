# ADR-pr6488: Decode a fast-sync block only from a peer the pool waits on, and hash its body only once signed

## Context

`BlockchainReactor.Receive` decoded every message and ran `ValidateBasic` on it before the pool
checked whether it had asked the sender for that block. `Block.ValidateBasic` builds a Merkle tree
over `Data.Txs`, and an empty tx costs 2 bytes on the wire, so `maxMsgSize` admits about four
million of them in one block response. Measured with such an 8000325-byte message, decoding
allocated 674 MiB and `ValidateBasic` 1190 MiB more, after which `AddBlock` found no requester and,
within 100 heights of the pool's, dropped the block without reporting the peer. The reactor is
registered on every node, so any peer could repeat this against any node.

## Decision

`Receive` drops a message larger than `maxSmallMsgSize`, 64 bytes, undecoded, unless the pool is
running and waits on at least one block from its sender. Every message of this reactor but a block
response carries one height and is at most 34 bytes at `math.MaxInt64`.

`Receive` checks a block response against its header alone, through `Block.ValidateHeaderBasic`:
every check of `ValidateBasic` but the two that hash the body against `LastCommitHash` and
`DataHash`. It also refuses an empty hash over a non-empty body, since `Block.Hash` fills an empty
one from the body. The pool routine runs the full `ValidateBasic` once `VerifyCommit` has accepted
the block, before saving it, so a block the validators signed over a body its header does not
describe is refused there, where `ApplyBlock` would panic on it. A body altered under a signed
header already fails `VerifyCommit`, whose `PartsHeader` covers the serialized block.

A decoded block response whose header checks out goes to `AddBlock`, which stores it only if a
requester for its height is assigned to the sender and still empty. Otherwise the peer is reported,
whatever the height, but for one late answer. A height from the pool's up gets its requester before
any peer is asked for it and keeps it until popped, so nobody was asked for such a block with no
requester. Below the pool's height an answer can be honest and late: the request redone to that
peer after its first deliverer dropped while the pool verified the block, then popped before the
answer arrived. The pop marks that height on the peer the request was redone to, and the first
block for it that peer sends clears the mark and is dropped unreported. A requester popped before
its block arrived, its request redone after the block it held was verified, gives back the pending
count it held, so the peer it re-asked is not later dropped as timed out. That give-back and a
delivery's both reach the pool, and the peer only when the pool's entry for that ID is the one that
counted the request, since one removed and re-added under the same ID never did. A requester is
assigned its peer in the lock section that counts the request against it, so no removal lands
between counting a request and assigning it, and a requester popped while its pick waited for that
lock is counted against nobody. A report is sent while the pool's lock is held, and the consumer of
the reports takes that lock to remove a peer, so a report finding the channel full is dropped and
logged, where the base blocked there for good.

Measured on the same 8000325-byte message: nothing is decoded from a peer the pool is not waiting on.

## Alternatives considered

Only checking the requester after the decode, without the size gate. Rejected because the decode
alone allocates 674 MiB, so an unsolicited message would still cost that much.

Disconnecting the sender of a large message the pool is not waiting on. Rejected because an honest
peer answering a request after the pool stopped, when the node switched to consensus, sends exactly
that.

A cap on the number of txs in `Block.ValidateBasic`. Rejected for this change: no stateless bound
separates an honest block from a stuffed one, since `MaxBlockDataBytesLimit` admits four million
empty txs, and the cap would still run after the decode.

Checking the requester before the header. Dropped: with the body hashed only after the commit, the
header check costs 3.1 µs on the 8000325-byte message, against 906 ms for its decode, and
`AddBlock` checks the requester anyway. The header check also stops the sender of a malformed block
the pool did not ask for, as on the base.

Making the hash cheaper, as `pr6488_merkle_root_no_alloc.md` does, without moving it. Not enough alone: it takes
`Txs.Hash` from 1190 MiB to 214 MiB on four million txs, where checking the header first leaves any
block no commit signed unhashed.

## Consequences

A block from a peer the pool asked is decoded but not hashed until its commit verifies; one that
fails `VerifyCommit` drops the peer. While the pool waits on a peer, a stuffed block for a height it was not asked for, below the pool's or not, is still
decoded, and gets it stopped, which removes it from the pool, so the size gate decodes no large
message from it after that. A redo can still be lost when the requester's one-slot redo channel is
full, as on the base, and the requester then waits on a peer the pool no longer does. A peer
dropped in the last 20µs or so before a pop can still leave the pool's pending count one too high,
where on the base that window reaches past 2ms.

A large message from a peer the pool is not waiting on is dropped and the peer kept, where the base
disconnected it when the message failed to decode or to validate, or carried a block more than 100
heights from the pool's or for a height assigned to another peer. With the pool stopped, on a node
not fast-syncing, that covers every large message on the channel. Such a message costs the node its
bytes and nothing more.

A block the validators signed over an invalid body now disconnects the senders of both blocks the
pool routine checks, as a failed `VerifyCommit` does, where the base disconnected its sender in
`Receive`. `ValidateBasic` runs its header checks before its hash comparisons, so a block failing
both kinds reports the header error first.
