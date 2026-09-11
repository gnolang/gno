# PR6154: Bounds on the structure of a multisig threshold key

## Context

`multisig.PubKeyMultisigThreshold` is a K-of-N key whose constituents are
themselves `crypto.PubKey` values, so a threshold key may hold other threshold
keys. Amino decodes one straight out of the signature field of an untrusted
transaction, populating `K` and `PubKeys` without going through
`NewPubKeyMultisigThreshold`, so nothing about the shape of such a key is
established before the verification path walks it.

Verification walks it twice, and both walks are driven by the key rather than by
anything the transaction pays for:

- `auth.DefaultSigVerificationGasConsumer` recurses over the constituent keys,
  decoding whatever signature bytes remain at each level, and charges gas only
  for the leaves it reaches.
- `PubKeyMultisigThreshold.VerifyBytes` recurses over the same structure, and
  `validate` scans each level's constituents pairwise to reject duplicates.

### Why TxSigLimit does not bound this

`auth.ValidateSigCount` limits `std.CountSubKeys` to `TxSigLimit` (default 7),
but `CountSubKeys` counts **leaves**, and neither axis of the structure is
visible to a leaf count:

- A chain of 1-of-1 threshold keys has exactly one leaf however deep it runs, so
  depth is unbounded.
- A constituent that holds no keys at all contributes no leaves, so a key over
  36,000 such constituents counts **zero** and is unbounded in width.

### The costs, as measured

Measured end to end through `NewAnteHandler` at production block params
(`MaxTxBytes` 1MB, `MaxGas` 3B), against the unfixed code:

| payload | tx bytes | ante handler | wire-decodable |
|---|---|---|---|
| 36,000 leafless constituents | 996,162 | **7.31 s** | yes |
| depth 63, signature padded | 902,712 | 33.9 ms | yes |
| depth 23,000 | 940,445 | 1.73 s | **no** |

Both quadratics are real, but they are not equally reachable, and this matters
for how the bounds are chosen. Amino's binary decoder enforces
`maxAnyDepth = 64` (`tm2/pkg/amino/binary_decode.go`), and each level of nesting
is one Any, so **the deepest key a transaction can carry is 63**. Depths beyond
that fail the decode a node performs before the ante handler ever runs, so the
reachable cost of the depth attack is ~34 ms, not the seconds a deeper in-process
key suggests. The width attack has no such encoding-layer ceiling: 36,000
constituents fit inside `MaxTxBytes` and are decoded without complaint.

In both cases the transaction is rejected, so `runTx` returns before
`msCache.MultiWrite()` and `CheckTx` charges nothing at all for the work.

Two further defects on the same path, found reviewing PR #22:

- The multisig case decoded the untrusted signature with `amino.MustUnmarshal`.
  Four bytes of `0xff` panicked past the ante handler's own `defer`, which
  recovers only `store.OutOfGasError` and repanics the rest, into `runTx`'s
  blanket recover as an `ErrInternal` carrying a stack trace. Not a consensus
  hazard — `ABCIResult` carries `Error`, not `Log`, and `std.InternalError` is a
  bare marker type, so the nondeterministic trace never reaches the results hash.
- `VerifyBytes` checked `len(Sigs)` against `K` and against the bit array size,
  but never against the number of **set bits**, while advancing one signature per
  set bit. Three keys, three set bits and two valid signatures panicked with
  `index out of range [2] with length 2`. The earlier signatures have to be valid
  for the loop to reach the overrun, which is why malformed-key tests did not
  cover it.

And two more, found reviewing this PR:

- The same count is unchecked in the other direction, and there it is not a panic
  but a shape nothing interprets: signatures past the last set bit are never
  indexed, so appending them to a valid multisignature leaves a transaction that
  authorizes exactly the same thing under different bytes. Measured through the
  full ante handler, a 404-byte transaction and a 421-byte one built from a single
  ed25519 signature both pass.

  An earlier draft of this ADR called that a transaction-hash malleability and
  justified rejecting it on those grounds. That justification does not hold, and
  the correction is recorded in **Non-goals** below: amino's decoder admits many
  byte encodings of one value, so the multisignature was never the only way to
  restate a transaction, and rejecting these does not make it canonical. The
  reason to reject them is narrower and stands on its own: a verifier should not
  accept fields it never reads.
- `execAddMultisig` pre-validated only the threshold, via
  `keys.ValidateMultisigThreshold`, and then called the panicking constructor. So
  every condition `validate` has grown since — duplicate constituents from PR #17,
  and now the structural bounds — reached that constructor unguarded and crashed
  `gnokey add multisig` with a stack trace on ordinary input.

And two more in the `CompactBitArray` those set-bit counts are derived from,
found reviewing this PR again once the counts were equal. The key half of the
signature was being validated thoroughly by then; the bit array half was still
taken entirely on trust, and it decides both how far verification walks and which
signatures it indexes:

- `Size()` is `(len(Elems)-1)*8 + ExtraBitsStored`, and `ExtraBitsStored` comes
  off the wire without ever being checked against `len(Elems)`. `GetIndex` bounds
  its index against `Size()`, not against `len(Elems)`, so nine bits claimed over
  one stored byte make `NumTrueBitsBefore` index `Elems[1]`. A **401-byte
  transaction with a 7-byte signature** panicked `index out of range [1] with
  length 1` out of the gas consumer — the same escape into `runTx`'s blanket
  recover that `MustUnmarshal` had, by a payload that decodes cleanly, so the
  `amino.Unmarshal` above does not catch it. `VerifyBytes` panics on the same
  bytes, and `gnogenesis verify` and `gnogenesis fork` call it on genesis
  documents.
- The bits of the final byte past `Size()` are never read, so setting them is the
  same unread-field problem trailing signatures are. Two 346-byte transactions
  differing only in those bits, with different sha256, both passed the full ante
  handler. `ExtraBitsStored` of 0 and 8 over the same `Elems` are a second such
  pair, for any key whose constituent count is a multiple of 8.

And one more in the key, found in a fifth review, which neither the structural
bounds nor the bit array checks cover:

- `validate` nil-checks its own constituents, but `Equals` recurses into the
  constituents of the keys it compares, so a nil nested one level below the top is
  reached by the duplicate scan rather than by the nil check above it, and
  dereferenced. A **277-byte transaction** panicked
  `invalid memory address or nil pointer dereference` at `Equals`, and the ante
  handler's own `defer` re-panicked it into `runTx`'s blanket recover as an
  `ErrInternal` — the third instance of that same escape, after `MustUnmarshal`
  and the malformed bit array, and the one both earlier fixes walked past.

  The shape survives an amino binary round-trip, so it arrives intact from the
  wire, and it clears every guard upstream: `std.CountSubKeys` counts it as 2
  leaves, `ValidateStructure` accepted it at 5 keys and depth 2, and with no bits
  set the whole gas consumer passed at zero gas. It reproduces through
  `amino.UnmarshalJSON` too, which is the genesis-document path `gnogenesis
  verify` and `gnogenesis fork` take with no recover of their own.

## Decision

Bound both axes of the structure, in one walk, as key invariants.

```go
const MaxNestingDepth = 6
const MaxTotalKeys    = 14
```

`ValidateStructure` walks the key once, spending a level budget and a key budget,
and rejecting a nil constituent at any node. All three are applied to a key
**before** its own constituents are looked at, so a key that exhausts either
budget is rejected without being walked to the bottom: the check never costs what
it is meant to prevent. A 100,000-level chain and a 100,000-wide key are both
rejected after examining a handful of nodes.

The nil rejection lives here rather than in `validate`'s own loop because this
walk is the only check that visits **every** node. A per-constituent nil check
covers one level, while `Equals` — which the duplicate scan calls — compares whole
subtrees, so a nil below the top level is reached by the comparison and not by the
check. `validate`'s own `pubkey == nil` clause is removed rather than left beside
it: a dead check in that position is what made the nil hazard look handled.

It is enforced in two places, deliberately sharing one implementation:

- `validate` calls it, so `NewPubKeyMultisigThreshold` and `VerifyBytes` both
  reject an out-of-bounds key. Placement inside `validate` matters: it runs
  **before** the pairwise duplicate scan, which is what makes that scan bounded
  and what keeps it from dereferencing a nil.
- `DefaultSigVerificationGasConsumer` calls `ValidateStructure` directly, because
  it recurses over the constituent keys itself and reaches them before
  `VerifyBytes` is ever called. It checks the key *before* decoding the
  signature: the key alone decides the shape, so there is no reason to decode
  attacker bytes for a key already rejected.

### Choice of values

`MaxNestingDepth = 6` is the deepest a key that branches at every level can nest
while spending no more than `TxSigLimit`'s 7 leaves.

An earlier draft set this to 3, on the reasoning that "a key that actually
branches at every level reaches at most depth 3". That is false, and it is worth
recording why, because it rejected keys nobody intended to reject. The reasoning
assumed a balanced tree, where depth d costs 2^d leaves. The deep shape that fits
in a leaf budget is the *unbalanced* one: a level holding one leaf and one subkey
branches — it has two constituents, so the degenerate-link argument does not apply
to it — and d such levels spend only d+1 leaves and 2d+1 keys. So 7 leaves reach
depth 6 at 13 keys, and a 2-of-2 where one party is itself a 2-of-2 where one
party is a 2-of-2 — an escalation hierarchy, and a perfectly ordinary key — was
being rejected at depth 4 with 5 leaves and 9 keys.

At 6, the two bounds agree: the deepest branching key inside `TxSigLimit` is also
the widest one `MaxTotalKeys` admits, so neither constant is the binding one for a
key that branches. Anything deeper needs a level holding a single subkey, which
delegates its whole threshold to that subkey and expresses nothing the subkey did
not already express. Those remain reachable within `MaxTotalKeys` alone — a 1-of-1
chain costs one key per level — which is why the depth bound stays a separate
check rather than a consequence of the key budget.

`MaxTotalKeys = 14` is twice the default `TxSigLimit`. A key that spends all 7 of
its permitted leaves and branches at every level needs at most 6 threshold keys
above them, so 14 admits every branching key `TxSigLimit` permits, with one node
to spare.

It does not admit every *degenerate* key `TxSigLimit` permits, and that is worth
stating plainly rather than glossing: 7 leaves each behind a 1-of-1 wrapper of
its own counts 7 against `TxSigLimit` but is 15 keys, and is rejected. The
"at most 6 threshold keys above them" bound holds only where every threshold key
has at least two constituents. Wrapping a lone subkey expresses nothing that
subkey did not already express — the same reasoning `MaxNestingDepth` rests on —
so no expressible key is lost, but the bound is not the strict superset of
`TxSigLimit` that a looser reading would suggest.

`MaxTotalKeys` is deliberately tighter than `TxSigLimit` strictly requires, chosen
to be conservative ahead of mainnet on the reasoning that a bound is easier to
relax later than to introduce later. Between the two new constants neither is
uniformly the tighter: they bite at the same shape for a key that branches, while
`MaxNestingDepth` is what stops a degenerate chain (one key per level, so 14 keys
would otherwise reach depth 13) and `MaxTotalKeys` is what stops a wide flat one.

`TxSigLimit`, though, is a governable consensus parameter — `WillSetParam` accepts
`p:tx_sig_limit` — while `MaxTotalKeys` is a compile-time constant that no
parameter change can raise. "A chain configuring `TxSigLimit` above 7 has to raise
`MaxTotalKeys` with it" therefore describes a binary release, not a config change,
and nothing previously stopped the two from being set inconsistently: a chain that
raised the parameter to 20 would accept a 15-of-20 key against its own limit and
then reject it with `ErrInvalidPubKey` in the signature check. So
`auth.Params.Validate` now refuses the desynchronised parameter instead:
`TxSigLimit` may not exceed `(MaxTotalKeys+1)/2`, the most leaves a branching key
can spend within the key budget. A chain wanting more raises both, together, in a
release — which was always the requirement, now enforced where it can be seen.

Alongside these:

- Require **exactly** one signature per set bit, in `VerifyBytes` and in the gas
  consumer alike, reusing the `NumTrueBitsBefore` result the `< K` check already
  computes.

  The lower half of that equality is what stops the index overrun, and it can
  only turn a panic into `false`, never a `true` into a `false`: reaching
  `return true` requires indexing all the set bits, which is the panic.

  The upper half *does* turn some `true` into `false`, and it is the breaking
  half. It rejects signatures the walk never indexes — see **Non-goals** for what
  that does and does not buy. Nothing a signer assembles is affected, because
  `AddSignature` maintains the equality; it did not before this change, which is
  the next item.

  The equality also subsumes `VerifyBytes`'s separate length check on `Sigs`,
  which is dropped: `nSigs` is at most `size` by construction and at least `K` by
  the first clause, so both of that check's clauses follow from it.
- Make `AddSignature` report an error rather than silently break that equality.
  It discarded `SetIndex`'s failure return, so for an index at or past the bit
  array's size it set no bit and appended a signature anyway:
  `NewMultisig(3).AddSignature(sig, 3)` produced 0 set bits and 1 signature, and
  `NewMultisig(0).AddSignature(sig, 0)` the same. `gnokey multisign` passes
  `len(PubKeys)` to both `NewMultisig` and the index lookup, so the equality held
  there by coincidence of one call site rather than by construction — and the
  claim being relied on above is about the function. Now that the equality is a
  consensus rule, the exported assembler has to refuse rather than hand back
  something the chain will reject with nothing having complained.
  `AddSignatureFromPubKey` propagates it, which is what its one non-test caller
  already checks.
- Add `bitarray.CompactBitArray.ValidateBasic` and call it before anything trusts
  `Size()`. It is the counterpart for this type of `bitarray.BitArray.ValidateBasic`
  in the sibling package, which already checks its own declared size against its
  backing slice.

  It establishes that `bA` holds exactly the `Size()` bits it claims, which is two
  conditions. `ExtraBitsStored < 8`, and non-zero only when `Elems` is non-empty,
  leaves `len(Elems) == (Size()+7)/8`, so every index `GetIndex` admits is backed
  by a byte and the walk cannot run off the end — this is the load-bearing half.
  The `8-ExtraBitsStored` bits of the final byte that `Size()` excludes must then
  be zero, which rejects a value setting bits the type cannot mean.
- Hold all three signature-shape checks in one place, `Multisignature.ValidateBasic(nKeys)`,
  called by `VerifyBytes` and by `consumeMultisignatureVerificationGas`. An earlier
  draft wrote them out separately at both sites, which contradicted the reasoning
  given for extracting `CompactBitArray.ValidateBasic` two items up: two callers
  walking one structure should not disagree about which shapes of it are walkable,
  and there was no test driving one input through both to catch it if they drifted.
  Writing the same predicate twice across a package boundary is how the
  `> len(Sigs)` / `!=` asymmetry this PR is fixing arose in the first place.

  A side effect worth naming: the bit-array-size mismatch now returns
  `ErrUnauthorized` from the gas consumer where it returned `ErrInvalidPubKey`, so
  every signature-shape rejection is one error class and every key-shape rejection
  the other. The split was previously incoherent — the same defect surfaced as
  either, depending on which clause caught it first.
- Decode the signature with `amino.Unmarshal` and return `ErrUnauthorized`. The
  amino error is deliberately not passed through: it renders the offending buffer
  as hex, and the buffer is attacker-sized.
- Add `NewPubKeyMultisigThresholdChecked`, the same construction reporting an
  error instead of panicking, and call it from `execAddMultisig`.
  `NewPubKeyMultisigThreshold` keeps its panic for the callers that want it.

## Non-goals

**This change does not make a transaction's encoding canonical, and nothing in it
should be read as claiming so.** An earlier draft did claim it, twice — of the
set-bit equality and of `CompactBitArray.ValidateBasic` — and justified the
breaking half of the equality on those grounds. Both claims were wrong at the byte
level, and the correction matters for judging what the break is worth.

`amino.DecodeUvarint` is Go's `binary.Uvarint` with no minimality check, and the
encoder omits zero-valued fields while the decoder accepts them written out. So a
multisignature still has many accepted encodings. Measured through the full ante
handler, after this change:

| malleation | canonical | malleated | both pass |
|---|---|---|---|
| `ExtraBitsStored` varint `08 01` written `08 81 00` | 288 B | 289 B | yes |
| `BitArray` field key `0a` written `8a 00` | 288 B | 289 B | yes |
| `ExtraBitsStored = 0` written out explicitly | 470 B | 472 B | yes |
| trailing signature past the last set bit | 404 B | 421 B | **no** — rejected |

Only the last is closed. And the same trick on the transaction *envelope* rather
than the signature gives two byte strings, 253 B and 254 B with different sha256,
that decode to a byte-identical `std.Tx` and both pass — for an ordinary
single-signer ed25519 transaction with no multisig anywhere in it. Amino's
tolerance is the malleability; the multisignature was one instance of it, not the
source. The earlier claim that "multisig is the only signature type this applies
to" was wrong for the same reason: `Tx.Hash` is `tmhash` over the raw wire bytes
while the signature covers a re-marshalled `SignDoc`, so any padded varint
anywhere in a transaction restates it under a new hash.

What the checks in this change actually buy, stated without the overreach:

- **The out-of-bounds index is closed.** This is the load-bearing part, and it is
  a panic on an untrusted path, not a cosmetic property.
- **The verifier no longer accepts fields it does not read.** A trailing signature
  and a bit set past `Size()` are values the type cannot mean, and the accepted set
  now equals the set a signer can produce. That is worth having on its own terms —
  a verifier that ignores part of its input is a verifier whose behaviour is harder
  to reason about — but it is a narrowing of the accepted set, not uniqueness of
  the encoding.

Making the transaction encoding genuinely canonical means rejecting non-minimal
varints in the amino decoder, which is a change to every message type on the chain
and belongs in its own PR with its own consensus analysis. It is filed separately.

## Alternatives considered

**Make `CountSubKeys` count nodes and reuse `TxSigLimit` as the budget.** One
constant would then bound both axes. Rejected: counting the threshold keys
themselves puts a flat 7-of-7 at 8 nodes against a limit of 7, so it would reject
keys the chain accepts today (verified: 6-of-6 passes at 7 nodes, 7-of-7 fails at
8). A node budget needs its own constant regardless, which is what `MaxTotalKeys`
is, and leaving `CountSubKeys` alone keeps `TxSigLimit`'s meaning unchanged.

**Replace the pairwise duplicate scan with an O(n) map keyed on the serialized
key.** This works and is equivalent for every registered key type. Rejected as
redundant: once the total is bounded at 14, the scan is at most ~98 comparisons.
The bound subsumes the need for it, and one mechanism is preferable to two.

**Rely on amino's `maxAnyDepth` for the depth bound.** Rejected: it is an
encoding-layer guard on Any nesting, not an invariant of the key type. It does
not constrain a key built in process or reached by any path that does not go
through a binary decode, and it says nothing about width. Depth 63 is also far
looser than anything a legitimate key needs.

**Charge gas per nesting level rather than bounding the structure.** Rejected: it
does not help where the cost is free. An aborted ante handler in `CheckTx`
commits nothing, so gas metering cannot price mempool work.

**Enforce the bounds only in the ante handler.** Rejected: `VerifyBytes` is also
called on genesis documents by `gnogenesis verify` and `gnogenesis fork`, which
crash or stall on a hand-crafted one. Putting the bound in `validate` covers
every caller.

**Reject non-minimal varints in the amino decoder, to make the encoding actually
canonical.** This is what the malleability argument in the earlier drafts of this
document would have required. Rejected *for this PR*, not on the merits: it changes
the accepted encoding of every message type on the chain, not just multisig
signatures, so it needs its own consensus analysis and its own compatibility
survey. Filed separately. Its absence is why the **Non-goals** section exists.

**Derive `MaxTotalKeys` from `params.TxSigLimit` inside `auth`** rather than
cross-validating the two. This is not the rejected alternative below — it keeps the
node budget as its own quantity but computes it per-chain, so raising the parameter
raises the budget. Rejected because the bound has to hold in `validate`, which is
in `crypto/multisig`, and that package cannot import `auth`: the dependency runs
`auth → std → multisig`. Threading params into every `VerifyBytes` caller —
including `gnogenesis`, which has no chain params — costs more than the cross-check
in `Params.Validate` buys back.

**Add `MaxTotalKeys` to `keys.ValidateMultisigThreshold`** rather than giving the
constructor a fallible form. That function is what `execAddMultisig` already
calls to pre-check the two conditions the constructor panicked on originally, so
mirroring a third there is the smaller diff. Rejected: mirroring is what produced
the bug. It already drifted once — PR #17's duplicate check has no mirror, so
`gnokey add multisig --multisig k1 --multisig k1` panics today — and it cannot
acquire one, because the function takes counts and never sees the keys. Every
condition added to `validate` would have to remember a call site in another
package. The checked constructor has one implementation and cannot drift.

## Consequences

**Breaking.** A multisig key nesting more than `MaxNestingDepth` levels deep,
holding more than `MaxTotalKeys` keys in total, or holding a nil constituent at
any depth, is now rejected with `ErrInvalidPubKey` where the first two were
previously verified and the third panicked. A malformed multisignature — signature
bytes that are not one, or a bit array claiming more bits than it stores — now
returns `ErrUnauthorized` where it previously panicked and surfaced as
`ErrInternal`; failed either way, but the result and the gas differ.

No key that anyone can sign for is affected. `gnokey multisign` resolves signers
against the flat top-level `PubKeys` list
(`tm2/pkg/crypto/keys/client/multisign.go`), so it cannot sign a nested key at
all, and the only non-test caller of the constructor is `gnokey add --multisig`
on locally held keys.

On the in-tree survey, stated exactly, because an earlier draft understated it:
the widest *flat* key is a 5-of-10 at 11 nodes, but the widest key of any shape is
`auth`'s `TestCountSubkeys` `multiLevelMultiKey` — a 2-of-3 over two 4-of-5 groups
plus a leaf — at **14 nodes, exactly `MaxTotalKeys`, with zero margin**. It is
still accepted, but widening either group by one constituent puts it at 15 and it
is rejected. Anyone raising `MaxTotalKeys` should know the tree already sits on the
bound rather than three nodes below it.

The one key of consequence that the in-tree survey cannot see is the GovDAO T1
multisig (`g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh`, referenced from
`misc/deployments/`), whose constituent pubkeys are held off-tree; per its holders
it is 7 constituents, each directly a user key. That is 8 nodes against
`MaxTotalKeys`'s 14 and depth 1 against `MaxNestingDepth`'s 6, so it is unaffected
with room on both axes, and the flat-7 shape is already pinned by
`TestTotalKeys`. Its only tight axis is `TxSigLimit`, at exactly the 7 leaves the
default permits — which this change does not touch.

Also not covered: an account whose stored pubkey is out of bounds cannot be
recovered by presenting a re-shaped key, because `processSig` replaces the
supplied pubkey with `sigAcc.GetPubKey()` for any account that already has one. On
a chain where such an account exists and holds a balance, that balance is
unspendable. No such key is known in-tree, and gno.land is pre-mainnet, which is
what makes taking this break now rather than later the cheaper option.

**Breaking, second.** A multisignature carrying signatures past its last set bit
is now rejected with `ErrUnauthorized` where it previously verified. No signer
produces one: `gnokey multisign` is the only thing that assembles a
multisignature, and `AddSignature` maintains the equality — which it now does by
construction rather than by coincidence, see the Decision above. So this only ever
rejects an encoding someone built by hand or restated out of a valid one. See
**Non-goals** for what it does not achieve. `gnokey`'s own signatures round-trip
unchanged, and the whole of `tm2/pkg/{crypto,sdk,std}` and
`gno.land/pkg/{gnoland,sdk}` is green with the equality in place.

**Breaking, third**, and for the same reason. A multisignature whose bit array does
not hold exactly the bits it claims is now rejected with `ErrUnauthorized`, where
one with bits set past the end previously verified and one claiming bits it does
not store previously panicked. No signer produces either: `NewCompactBitArray`
sizes `Elems` from the bit count and `SetIndex` refuses any index at or past
`Size()`, so both properties hold by construction for every bit array `gnokey
multisign` assembles. That direction is pinned by a test of its own: every size
from 1 to 64 bits, built and populated through those two functions, validates.

**Breaking, fourth**, and outside the signature path: `auth.Params.Validate` now
rejects a `TxSigLimit` above `(MaxTotalKeys+1)/2`, which is 7 — the current
default, so no chain running defaults is affected. A chain that had configured a
higher limit has an invalid parameter set until it is lowered or both constants are
raised in a release. This is deliberate: the alternative is a chain whose own limit
admits keys its verifier rejects.

`AddSignature` gains an error return, a source-compatible change for callers that
ignored it and a behavioural one for any that passed an out-of-range index — which
previously produced a multisignature that now cannot verify.

`NewPubKeyMultisigThreshold` gains a new panic condition, consistent with its
documented behaviour of panicking on invalid input. It is not reachable from
untrusted input: no node-side code calls it. The one command that does call it,
`gnokey add multisig`, no longer does — it uses the checked form, so the bounds,
the duplicate check, and anything added later surface there as an error rather
than a stack trace.

The `VerifyBytes` set-bit lower bound is consensus-neutral **as gno.land is
wired**, since #22's gas consumer rejects that shape before `VerifyBytes` is
reached, including for nested keys — it recurses over the same bytes. That is a
property of the wiring rather than of the code: `sigGasConsumer` is a parameter of
`NewAnteHandler` and an advertised extension point, and gno.land passes
`DefaultSigVerificationGasConsumer` (`gno.land/pkg/gnoland/app.go`). An app
supplying its own consumer that omits the check would see the result for those
bytes change from `ErrInternal` to `ErrUnauthorized`. The upper bound is not
neutral under any wiring: it is enforced in both places, and rejects transactions a
node accepts today.

Cost after the change, same payloads and same harness:

| payload | before | after |
|---|---|---|
| 36,000 leafless constituents, 996,162 B | 7.31 s | **7.93 ms** |
| depth 63, signature padded, 902,712 B | 33.9 ms | **241 µs** |
| 100,000-wide key, `VerifyBytes` | minutes | 0.01 s |

The residual is linear in the transaction: the amino decode of the key and the
one `Address()` marshal the ante handler performs before the signature check.

Raising `MaxNestingDepth` from 3 to 6 doubles the number of levels an *accepted*
key may have, and each level re-decodes whatever signature bytes remain below it,
so the accepted worst case rises with it. Measured through `NewAnteHandler` with
the innermost signature padded to 900 KB, three runs:

| accepted depth | ante handler |
|---|---|
| 3 (the old bound) | 1.8 – 2.2 ms |
| 6 (the new bound) | 2.0 – 4.2 ms |
| 7, rejected at the bound | 0.3 – 0.6 ms |

So the change costs roughly one extra millisecond on a near-`MaxTxBytes`
transaction, against the 7.31 s and 33.9 ms the bounds exist to prevent, and
`MaxTotalKeys` caps the walk at 14 nodes regardless of what the depth bound
permits. Rejection still happens at the bound rather than after descending, which
is why depth 7 and depth 63 cost less than depth 6 rather than more.

## Provenance

This change and the reviews that produced it were AI-assisted. Every cost figure
above comes from a test that was executed, not from reading, including the
`maxAnyDepth` ceiling that makes the deepest measurements unreachable in
production and the `7.31 s` width case that the depth bound alone did not
address.

The malleability, the `MaxTotalKeys` overclaim corrected above, and the
`gnokey add multisig` panic came out of a third review, of this PR as it then
stood. Each was reproduced before being fixed — two transactions differing only
in bytes both passing the ante handler, a 7-leaf key counting 7 against
`TxSigLimit` and 15 against `MaxTotalKeys`, and the panic itself out of a built
`gnokey` — and each new test was confirmed to fail against the unfixed code
rather than merely to pass against the fixed code.

A fifth review produced the nil-constituent panic, the `MaxNestingDepth = 3`
rationale error, the `AddSignature` equality claim, the duplicated shape checks,
the understated in-tree survey, the `TxSigLimit` parameter desynchronisation, and
the canonicity overclaim now recorded under **Non-goals**. The same rule was
applied: the 277-byte panic was reproduced end to end through `NewAnteHandler`
first, the depth-4 branching key and the `AddSignature` counter-example were
exhibited before either was changed, and the four malleation vectors in the
Non-goals table were each run against the fixed code to establish what the change
does *not* buy — the correction of a claim the earlier drafts of this document were
making. With each fix reverted and the tests kept, each fails in its own way:
`TestNilConstituentAtAnyDepth` and `TestAnteHandlerMultisigNilNestedConstituent`
panic with a nil pointer dereference, `TestBranchingKeyDepth` reports
`nested more than 3 levels deep`, `TestAddSignatureOutOfRange` gets no error where
one is expected, and the two new `TxSigLimit` parameter cases accept the value they
exist to reject.

Three findings from that review were judged not worth changing and are recorded
instead: `ValidateStructure` re-walks each subtree once per ancestor at every
enforcement site, which is O(n·depth) on a structure now bounded at 14 nodes and so
not worth the complexity of hoisting; `NewPubKeyMultisigThreshold` keeps the
shorter name despite having no non-test callers left, because renaming it churns
~20 test call sites for no behavioural gain; and the gas consumer still accepts
degenerate constituents (`K = 0`, no keys) that only `VerifyBytes` rejects, which
costs zero gas and fails the transaction either way.

The two `CompactBitArray` defects came out of a fourth review, of this PR once
the set-bit counts were equal, and the same rule was applied to them: the
401-byte panic and the two 346-byte transactions were reproduced end to end
through `NewAnteHandler` first. With the two call sites reverted and the tests
kept, each fails in its own way — `TestVerifyBytesMalformedBitArray` and
`TestAnteHandlerMultisigMalformedBitArray` panic `index out of range [1] with
length 1`, and `TestVerifyBytesBitArrayPadding` and the new gas-consumer case
accept the malleated encoding.
