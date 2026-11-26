# PR #6504: Reject threshold multisig keys that reuse one signer across positions

## Context

A `PubKeyMultisigThreshold` is a K-of-N threshold key: it should require K
distinct signers. `validate` (in `tm2/pkg/crypto/multisig/threshold_pubkey.go`)
enforced distinctness by comparing the constituents of a single level with
`pubkey.Equals`. `Equals` is type-specific: it first asserts that the other key
has the same concrete type, then compares the whole subtree.

That leaves a gap. A bare leaf key (e.g. `PubKeyEd25519`) and a threshold key
wrapping that same leaf (e.g. `{K:1, [leaf]}`) are different concrete types, so
`leaf.Equals({K:1,[leaf]})` is always false and the per-level scan never fires.
`VerifyBytes` then satisfies each threshold position independently: the wrapper
position defers verification to the same leaf with the same signature, re-wrapped
in an inner `Multisignature`. The result is that `[leaf, {K:1,[leaf]}]` with K=2
is satisfied by the single private key behind `leaf`. Deeper nesting reduces K=4
to one or two signers while staying within `MaxNestingDepth` (6) and
`MaxTotalKeys` (14). `ValidateStructure` bounds the tree's size and depth but not
which private keys occupy it.

`validate` runs on untrusted input: `VerifyBytes` calls it on a
`PubKeyMultisigThreshold` decoded straight from a transaction signature, so this
is reachable from the auth ante handler, which accepts a caller-supplied pubkey
on an account's first transaction (the account address is the hash of that
pubkey, so it matches by construction).

This is distinct from the already-handled same-level, same-type exact duplicate
(`[A, A, B]`), which `Equals` still catches.

## Decision

`validate` collects every terminal leaf across the whole key tree, in the same
walk that bounds the tree's size and depth and rejects nils (the walk behind
`ValidateStructure`), and rejects the key if any two leaves are `Equals`. A
terminal leaf is any node that is not a `PubKeyMultisigThreshold`, or a threshold
node holding no constituents; a non-empty threshold node contributes its
constituents recursively rather than itself. Because
verification of any position that reduces to a leaf reduces to that leaf's key,
two positions reaching the same leaf need one signer, not two. The error is the
existing `"threshold k of n multisignature: duplicate pubkey"`, so callers and
tests that match on the message are unaffected.

This replaces the previous per-level `Equals` scan rather than adding to it. The
leaf scan subsumes it: two `Equals` siblings are either both leaves (compared
directly here) or both identical threshold subtrees (whose identical leaf sets
collide here, including two identical empty subtrees, each of which is a leaf);
different-type siblings were caught by neither. So the
same-level duplicate `[A, A, B]` is still rejected.

The scan runs only after that walk has bounded the node count and rejected
nils, so it is finite, dereferences nothing, and its pairwise cost is
charged against the same transaction-linear budget the rest of validation
enforces.

`Equals` is used for the leaf comparison rather than `Address()`. For the account
key types auth admits (ed25519, secp256k1) an address is the hash of the raw
pubkey, so either comparison identifies the same key; `Equals` avoids calling
`Address()` on caller-chosen keys, which does not widen the panic surface, and
avoids treating a 160-bit address collision as a shared key.

Distinct leaves are not yet distinct signers when a leaf is an ed25519 key of
small order. `crypto/ed25519.Verify` does not reject such a key, and a signature
with `S = 0` and a small-order `R` verifies under it: for the identity, for
every message. The identity also has a second encoding, with the x sign bit set,
that `Equals` treats as a different key, so `{K:3, [A, id, id']}` passed the
duplicate scan and was satisfied by A alone. `validate` therefore also rejects
any ed25519 leaf for which `PubKeyEd25519.IsSmallOrder` holds, with
`"threshold k of n multisignature: small-order ed25519 pubkey"`.
`IsSmallOrder` compares the key, with its top bit cleared, against libsodium's
`ed25519_is_small_order` list of seven encodings, which covers every point of
order 1, 2, 4 and 8, including the non-canonical encodings of 0 and 1. A key
derived from a private key has prime order, so no honest key is affected.
secp256k1 has cofactor 1, so it has no counterpart. The check is scoped to
multisig leaves. Rejecting small-order keys in `PubKeyEd25519.VerifyBytes`
itself would also cover single-key accounts, validator votes and p2p, and is
left to its own change.

## Alternatives considered

- **Walk every node and compare by `Address()`.** Catches the same shapes, but
  a multisig's address is the hash of its amino form and a custom key's
  `Address()` can panic or collide; comparing terminal leaves by `Equals`
  identifies the same private key without hashing inner multisigs or invoking
  `Address()` on untrusted input.
- **Reject any 1-of-1 wrapper as structurally vacuous.** Smaller rule, but it
  does not stop `[leaf, {K:1,[leaf, other]}]` (a 1-of-N wrapper), and it would
  break legitimate single-key wrappers.
- **Flatten the policy before comparing.** Changes the meaning of nested
  threshold composition; out of scope for a distinctness fix.
- **Compute the minimal satisfying signer set.** Precise but combinatorial and
  unnecessary for this class.

## Consequences

- A transaction presenting a multisig key whose positions share a leaf is now
  rejected by the ante handler with `Unauthorized` instead of accepted. This
  changes tx admission and therefore the resulting state, so it is a
  consensus-affecting validation change: mixed-version validator sets can
  diverge on `AppHash`, and it needs a coordinated binary upgrade (or an
  activation height). CheckTx results and consumed gas also change.
- Addresses and wire encodings of legitimate multisig keys are unchanged.
- Leaf distinctness is *sufficient* for the property we want — a key whose
  leaves are all distinct cannot under-count its signers — but it is not
  *necessary*. A key whose positions share a leaf yet still require K distinct
  signers is rejected anyway: `{K:2, [A, {K:2,[A,B]}]}` needs A and B, and
  `{K:2, [{K:2,[A,B]}, {K:2,[A,C]}]}` needs A, B and C, but both repeat a leaf
  and are now refused. This is a deliberately conservative bound: the precise
  test is the minimal-satisfying-signer-set analysis rejected under
  *Alternatives considered* as combinatorial and out of scope. Such
  AND-of-AND policy keys were valid before and are not after.
- Because of the above, it is **not** true that every previously valid multisig
  is unaffected. If an existing chain holds an account whose stored pubkey
  repeats any leaf — whether it genuinely under-counts signers or is an honest
  AND-of-AND composition — that account can no longer sign after the upgrade.
  Genesis replay that verifies signatures may fail on such historical
  transactions; replay that skips verification is unaffected. The
  pre-deployment survey must flag *any* stored multisig whose positions share a
  leaf, not only the vacuous-wrapper shapes, and replay targets should be
  surveyed the same way.
- A multisig with a small-order ed25519 leaf is rejected. Such a leaf never
  stood for a private key, so no key a signer could hold is refused.
- Custom (app-registered) key types that verify by delegation without being a
  `PubKeyMultisigThreshold` are treated as opaque leaves; distinctness among
  them still rests on their own `Equals`.

## Testing

- `TestVerifyBytesRejectsRepeatedLeafAcrossNestedPositions`
  (`tm2/pkg/crypto/multisig`): the `VerifyBytes` seam over the shapes above.
- `TestValidateRejectsDuplicateEmptySubtrees` (`tm2/pkg/crypto/multisig`): two
  identical empty threshold nodes collide as leaves.
- `TestValidateOverRejectsGenuineKofN` (`tm2/pkg/crypto/multisig`): pins the
  conservative rejection of AND-of-AND keys described under *Consequences*.
- `TestValidateRejectsSmallOrderEd25519Leaf` (`tm2/pkg/crypto/multisig`):
  `{K:3, [A, id, id']}` is rejected by `validate`, and `VerifyBytes` refuses it
  with one real signature and two forged ones; a small-order leaf nested below
  the top level is rejected too.
- `TestPubKeyIsSmallOrder` (`tm2/pkg/crypto/ed25519`): every listed encoding, with
  either sign bit, is small-order; freshly generated keys are not.
- `TestAnteHandlerMultisigSameSignerAcrossTypes` (`tm2/pkg/sdk/auth`): the real
  entry point — a funded account at the multisig's address, one signer, amino
  round-trip, ante rejection with `Unauthorized`; plus a baseline K=2 over two
  distinct leaves that still passes.
