# An unverifiable Merkle proof must error, not return an empty root

## Context

`computeHashFromAunts` used `nil` to mean "I cannot compute a root from this
proof". It has five such exits: a non-positive total, an out-of-range index, a
single-leaf tree carrying sibling hashes it should not have, a multi-leaf tree
carrying none, and two more propagating a recursive failure.

`SimpleProof.Verify` then compared that result against the caller's expected
root:

```go
computedHash := sp.ComputeRootHash()
if !bytes.Equal(computedHash, rootHash) { return errors.New("invalid root hash: ...") }
return nil // verified
```

`bytes.Equal(nil, []byte{})` is `true`. So whenever the computation gave up
*and* the expected root was empty or nil, the two "matched" and the proof was
declared valid — for an arbitrary leaf and an arbitrary proof. Reproduced across
all five exits:

| proof | expected root | verified (before) |
|---|---|---:|
| `total=0` | `[]byte{}` | true |
| `total=0` | `nil` | true |
| `index=5, total=3` | `[]byte{}` | true |
| `index=1, total=1` | `nil` | true |
| `total=4`, no sibling hashes | `[]byte{}` | true |
| `total=1`, one stray sibling hash | `[]byte{}` | true |
| `total=0` | a real 32-byte root | false |

The last row is the mitigation that keeps this latent rather than live: the
expected root has to be empty. tm2's own callers take the root from a signed
header — `bft/types.Part.Proof`, `TxProof`, `ABCIResults` — so it is always
32 bytes and they were never exposed.

What changes the exposure is `gnovm/stdlibs/crypto/merkle`. Its
`VerifySimpleProof(rootHash, leaf, index, total, aunts)` hands the root to a
realm as an ordinary argument, and `index`/`total` come from the proof under
test — that is, from whoever is trying to convince the realm. A realm that
reads the root from a message, from an uninitialised state slot, or simply
without checking it is 32 bytes long would accept any leaf. No realm uses it
yet, which is why this is hardening rather than an incident.

`Verify` also validated `Total < 0` and `Index < 0` but not `Total == 0` or
`Index >= Total`, so it forwarded proofs it should have refused outright.

## Decision

`computeHashFromAunts` and `SimpleProof.ComputeRootHash` return
`([]byte, error)`. Each of the five failure paths returns a distinct error.

The point is to make the failure *unrepresentable* rather than *guarded*. Adding
`Total == 0` and `Index >= Total` checks to `Verify` would have closed the cases
we happened to enumerate, at the one call site we happened to look at. It would
have missed the second one: `SimpleValueOp.Run` (`proof_simple_value.go`)
returned `[][]byte{op.Proof.ComputeRootHash()}` with no nil check, and that
value reaches `ProofOperators.Verify`, which does `bytes.Equal(root, args[0])`
— the identical collision, on the store/ICS23 proof path instead of the Merkle
one. A sentinel that cannot be constructed cannot be missed at a call site
nobody reviewed, including the next one somebody adds.

`Verify` keeps its `error` signature and gains the two missing range checks, so
a malformed proof is refused at the boundary with a clear message instead of
being handed to the hash walk. Those checks are now a diagnostic convenience
rather than the thing standing between a garbage proof and `true`.

## Alternatives considered

- **Guards in `Verify` only.** Cheapest diff, and it does close the reported
  case. Rejected: it leaves the sentinel in place, so `proof_simple_value.go`
  stays broken and every future caller inherits the same trap.
- **A `nil` check in each caller.** Same objection, spread over more sites.
- **Returning a sentinel error value instead of distinct ones.** The five exits
  have genuinely different causes, and the old code's single `nil` is exactly
  what made the failure hard to read. `Verify` previously reported
  `invalid root hash: wanted X got <empty>` for a malformed proof, sending the
  reader after a hash mismatch that never happened.
- **Fixing it in `gnovm/stdlibs/crypto/merkle` instead.** Wrong layer: both
  defects are in tm2, and the stdlib is only the newest consumer.

## Consequences

- **Breaking change to an exported tm2 API.** `SimpleProof.ComputeRootHash`
  changes signature. Two in-repo callers, both already returning `error` and
  both previously wrong: `Verify` and `SimpleValueOp.Run`.
  `rootmulti.MultiStoreProof.ComputeRootHash` is a different method and is
  untouched; `convert.go` only references the algorithm in a comment.
- **Behaviour-visible, but no state-encoding change.** Proofs that used to
  verify against an empty root now fail. `Verify`'s own signature is unchanged,
  so `X_verifySimpleProof` in the Gno stdlib needs no edit and no stdlib
  MemPackage source bytes move — the committed multistore root is unaffected.
- **Error messages stop misdirecting.** A malformed proof now names its actual
  defect rather than reporting a root mismatch.
- The unreachable `case 0: panic("Cannot call computeHashFromAunts() with 0
  total")` is gone; the `total <= 0` check above it always fired first.
- `ValidateBasic` still only checks `Total < 0` / `Index < 0`. Left alone: it is
  a wire-shape check with its own callers and its own error contract, and
  `Verify` no longer depends on it for safety. Tightening it is a separate
  question about what an acceptable serialized proof is.

## Verification

- `TestSimpleProofVerifyRejectsUncomputableProofs` covers all six rows of the
  table above; `TestComputeRootHashErrorsInsteadOfReturningNil` pins that a
  failure never comes back as a comparable value;
  `TestSimpleValueOpRunRejectsUncomputableProof` covers the second call site
  (its leaf hash is built the way `Run` rebuilds it, so the test reaches the
  root computation instead of stopping at the leaf-hash mismatch).
- Negative control: reinstating the old guards and swallowing the new error
  (`computedHash, _ := sp.ComputeRootHash()`) turns all six rows red.
- `./tm2/pkg/crypto/...`, `./tm2/pkg/store/...` and `./tm2/pkg/bft/types/...`
  pass, so the tightened range checks reject nothing legitimate — including the
  block-part, tx-proof and ABCI-result paths that build proofs for real trees.
- Through the Gno stdlib, `merkle.VerifySimpleProof` now returns `false` for all
  four reachable shapes; every one returned `true` before.
