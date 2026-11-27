# PR #6504: Reject small-order ed25519 keys in signature verification

## Context

`PubKeyEd25519.VerifyBytes` calls `crypto/ed25519.Verify`. That function
implements RFC 8032's cofactorless equation and does not reject public keys of
small order. Under such a key a signature with `S = 0` and a small-order `R`
verifies: with the identity (`01 00…00`) as both key and `R`, for every
message; for the other small-order points, for a message the attacker chooses
(vary a memo; expect roughly one working `R` among the eight small-order points
per attempt). Go also accepts the non-canonical encodings of 0 and 1 (`y = p`,
`y = p+1`) and the x sign bit, so 14 byte strings behave this way.

A key derived from a private key has prime order and is never small order. So
for these keys `VerifyBytes` returning true stops meaning "somebody holds the
private key". The key identifies no signer.

The multisig unique-leaf check (`pr6504_multisig_unique_leaf_keys.md`) added `PubKeyEd25519.IsSmallOrder` and rejected small-order leaves in
`PubKeyMultisigThreshold.validate`. Its ADR left the rest of `VerifyBytes`'s
callers to this change; the gno stdlib, which it did not consider, is added
here:

- the auth ante handler, for single-key accounts. An address is the hash of its
  pubkey, so an account created with a small-order key is spendable by anyone.
- consensus: validator votes, commits, proposals and duplicate-vote evidence
  are verified against the validator's key, and a validator admitted with a
  small-order key (a governance proposal, or an operator bug) can be signed for
  by anyone.
- delegated sessions. `MsgCreateSession` took any `SessionKey`
  (`ValidateBasic` only checked it was non-nil), and in ante the session
  account then signs for its master's messages. A small-order session key is
  bound to an honest master's authority, not to a fresh address: a session
  created with one (say a dApp or wallet that sends an uninitialized all-zero
  key) let anyone act as the master within the session's allow-paths and
  spend limit.
- p2p `SecretConnection` and the remote-signer handshake authenticate the
  peer's ed25519 key. The remote-signer allowlists compare raw pubkeys, so a
  small-order entry authenticates nobody.
- the gno standard library `crypto/ed25519.Verify` (`X_verify`), which a realm
  calls with a caller-supplied pubkey. A realm that treats a passing signature
  as proof that the caller holds that key is forgeable. Unlike the chain paths,
  a realm has no address to bind the key to: it may use the pubkey itself as
  the identity.

Every chain-side consumer checks the key's derived address against the identity
it is authenticating (`ante.go`, `Vote.Verify`, `VerifyCommit`, evidence, the
p2p peer ID). That is why this is not an impersonation primitive against an
honest key: a small-order key cannot be substituted for an existing account,
validator or node ID. What it creates is an identity nobody controls. Sessions
are the chain-side exception: the session key's address identifies the
session, but the authority it carries is the master's. The stdlib path is the
other: there is no derived address unless the realm computes one itself, so a
realm that authenticates by pubkey can be forged today.

## Decision

Reject small-order public keys in `PubKeyEd25519.VerifyBytes`, and route the
gno stdlib `crypto/ed25519.Verify` native through it. Also refuse them where a
validator or a session key is admitted (below).
`IsSmallOrder` (already present from the multisig unique-leaf check) normalizes by clearing the top bit
and compares against libsodium's seven `ed25519_is_small_order` entries, which
covers all eight points of order 1, 2, 4 and 8 plus the non-canonical encodings
of 0 and 1; with both sign bits that is 14 byte strings. `VerifyBytes` calls it
before `crypto/ed25519.Verify`. The native takes `[]byte`: a 32-byte key
becomes a `PubKeyEd25519` and goes through `VerifyBytes`, which otherwise
matches `crypto/ed25519.Verify` (both return false for a signature that is not
64 bytes). The native returns false for a key that is not 32 bytes or a
signature that is not 64 bytes before either is copied.

One gate. `VerifyBytes` is the single check for tx signatures, consensus,
evidence, p2p, the remote signer, sessions, multisig leaves and realms, so a
later tightening of it reaches every caller.

Only the public key is checked, not `R`. With a prime-order `A`,
`R = [S]B - [k]A` lies in the prime-order subgroup, so `R` can be small order
only if it is the identity, and an identity `R` still requires the private
scalar `a` (`S = k·a mod L`; Go accepts such a holder-made signature). The
forgery exists because `A` is small order, so rejecting `A` is sufficient and
leaves every honest signature, whatever `R` it carries, untouched.

Three admission points also refuse a small-order key, so that it is turned
away where it is introduced rather than accepted and left unable to sign:

- `valset:proposed` (`nodeParamsKeeper.WillSetParam`) panics on a small-order
  ed25519 entry with power > 0. Such a validator can never sign, so above 1/3
  of the power it halts the chain; before this change it was a safety problem
  (anyone could sign for it), now it is a liveness one. Power 0 stays allowed,
  so governance can still remove one already in the set. The check is not in
  `abci.ParseValidatorUpdate`, which also reads `valset:current` in
  `WillSetParam` and the EndBlocker: a set that already held such a key would
  become unreadable (a dropped proposal or a panic at genesis).
- tm2's `validateValidatorUpdates` returns an error for a small-order ed25519
  update with power > 0, beside its key-type check. It runs when a block is
  applied, so for any ABCI app it is a backstop that stops the chain on such
  an update instead of admitting a validator that can never sign. gno.land
  does not reach it: `WillSetParam` refuses the entry first.
- `MsgCreateSession.ValidateBasic` returns `ErrInvalidPubKey` for a
  small-order ed25519 `SessionKey`, for the same reason the multisig unique-leaf check kept the multisig
  check at admission. `MsgRevokeSession` keeps accepting one, so a session
  created before the upgrade can still be removed.

The existing multisig leaf check stays, and it is not redundant with the
gate. `PubKeyMultisigThreshold.VerifyBytes` calls `validate()` first, so it is
`validate`'s check that rejects a key with a small-order leaf. Leaves are only
verified for set bits: under a 1-of-2 key {honest, small-order}, a signature
by the honest leaf alone never reaches the small-order leaf's `VerifyBytes`,
and without `validate`'s check that key would verify. The gate guarantees
that a small-order leaf cannot contribute a signature; `validate` also refuses
keys that merely contain one, and gives a precise `ErrInvalidPubKey` at
admission.

The check is a fixed seven-entry array comparison: constant work, a few
nanoseconds, inside the fixed base the gas meter already charges for an
ed25519 verification (`native_gas.go`, `crypto/ed25519.verify`). It adds no
size-dependent work, so no gas recalibration.

## Alternatives considered

- **Reject only at each admission point** (account pubkey, validator-set
  update, session creation, p2p key load, remote-signer allowlist, multisig
  leaves), with no gate in `VerifyBytes`. Keeps `VerifyBytes` a pure
  primitive, but validators enter the set through the app rather than one
  admission function, and the stdlib path is not covered at all. A future
  caller would reopen the hole. The admission checks above complement the
  gate; they do not replace it.
- **Reject only in the ante handler.** Fixes accounts and leaves consensus,
  p2p and realms open. The issue's first bullet is the least damaging of the
  four.
- **Reject small-order `R` as well.** libsodium does, but it buys nothing for
  this forgery (see Decision), and it is a consensus change of its own, not a
  no-op: it refuses holder-made `R = identity` signatures that verify today.
- **Switch to a stricter ed25519 implementation** (libsodium semantics,
  `ed25519-dalek`'s `verify_strict`). Same result, but a new dependency and a
  much larger consensus-surface change. ZIP-215, which CometBFT and
  `ed25519-consensus` use, is not an option: it explicitly allows small-order
  `A` and, because the equation is cofactored, accepts the identity forgery for
  every message.
- **Leave it to realm authors.** The native is a general primitive, but it is
  the only signature check a realm has, and the natural reading of
  `ed25519.Verify(pub, msg, sig)` as "the holder of `pub` signed `msg`" is the
  vulnerable one.

## Consequences

- Consensus-affecting validation change. `VerifyBytes` returns false for 14
  encodings it used to accept, so tx admission, CheckTx results, consumed gas,
  block validity, p2p handshakes and realm results can all change, and
  `MsgCreateSession` and `valset:proposed` writes that used to succeed now
  fail. Deploy as a coordinated binary upgrade (or gate the admission paths at
  an activation height; the check inside the pure function has no chain
  context). Mixed versions must not coexist in one validator set.
- No honest key is affected. A key derived from a private key has prime order.
- Existing identities must be surveyed before activation. On gno.land mainnet
  at height ~950k, and again at 987,099, there are six validators, all with
  ordinary keys, and the 14 possible small-order addresses are exhaustive
  (`Address` is the truncated hash of the key bytes); none has an account. An
  account is created on first receipt and never deleted (`RemoveAccount` has
  no caller outside tests), and every signer needs one (`GetSignerAcc`), so no
  account also means no transaction was ever signed under one of these keys:
  that settles the account path's history without a replay. The account query
  cannot see sessions, which are stored per master
  (`/a/<master>/s/<session>`), not under `auth/accounts/<addr>`; through
  indexer.gno.land, gnoland-1 has 22 `create_session` transactions (heights
  113002 to 838686), all with 33-byte secp256k1 keys, so no live session is
  affected. Testnets and other networks need the same three checks, and
  genesis/fork replay that verifies signatures will reject historical
  transactions signed under such a key. A validator that is already in
  `valset:current` with a small-order key stops signing, which is a halt if it
  holds power, so include it in the survey.
- On the realm side, 4 of the 682 non-stdlib packages on mainnet at ~951k
  import `crypto/ed25519` (listed with `vm/qpaths`, read with `vm/qfile`), but
  that count misses transitive use: a realm can reach `Verify` through a `p/`
  package. At ~968k (683 packages) the users are:
  - `p/g1zdaz6sypk5tdsjc2fzsqrrn4ttn8qzmnl4waqp/hyperlane/v0` takes the keys
    as an argument; it has no caller on mainnet.
  - `r/g1zdaz6sypk5tdsjc2fzsqrrn4ttn8qzmnl4waqp/hlverifier` takes its keys
    from `hlmailbox`, hard-coded in its `init`.
  - `r/samcrew/memba_quest_attestation_v1` verifies against an owner-set key.
  - `p/moul/x/wesh/v0` and `r/moul/x/wesh/v0` take a caller-supplied key:
    `Register` and `RegisterCommitted` check `accountPKHex` with
    `p/moul/x/wesh/v0.VerifyBind` and then store it. That is the forgeable
    pattern from the Context section, live before this change: anyone could
    register a Wesh identity under a key nobody holds. The realm imports
    `crypto/ed25519` only in `wesh_test.gno`, which is why the count missed
    it. The directory is empty at ~968k, so nothing stored is affected.

  The native has no activation height either, so historical transactions are
  settled by a mainnet replay of this branch.
- Multisig keys with a small-order leaf stay refused by the multisig unique-leaf check at
  admission, which the gate does not replace (see Decision).
- Realm behavior changes: `crypto/ed25519.Verify` returns false for small-order
  keys. A realm that authenticated a caller by a caller-supplied pubkey was
  forgeable before and is not after. No legitimate use requires a key that no
  private key can produce.
- Not addressed here: ed25519 malleability classes unrelated to small-order
  keys. Go already rejects `S >= L`, and `R` is not checked, which this change
  does not alter.
- The gno.land app-hash pin (`TestAppHashCrossrealm38`) moves, because stdlib
  `_test.gno` source bytes are committed into the genesis MemPackage. This is
  attributed by bisection to `crypto/ed25519/ed25519_test.gno` alone; the
  native change is not in the Merkle root and crossrealm38 does not use
  ed25519. Same class as the earlier `math/rand/example_test.gno` and
  `modexp_test.gno` bumps recorded above that constant.

## Testing

- `TestPubKeyIsSmallOrder` (`tm2/pkg/crypto/ed25519`) derives the
  small-order encodings from the curve instead of copying the table under
  test, with `filippo.io/edwards25519` (the upstream of Go's decoder, a test
  dependency). The small-order points are the multiples of a point of order 8,
  found as `[L]P`, and an unreduced `y + p` fits in 255 bits only for
  `y <= 18`, so every small-order encoding lies in a closed set: those
  multiples' encodings and `y + p` for `y <= 18`, with either sign bit (plus
  `y <= 18` itself as negatives). `IsSmallOrder` must match exactly the
  members that decode to a point with `[8]P` the identity, and there must be
  14. A row dropped from the table fails it, which a hand-copied list would
  not catch.
- `TestVerifyBytesRejectsSmallOrderKeys` (`tm2/pkg/crypto/ed25519`): each of
  those 14 encodings, with an `S = 0` forgery that `crypto/ed25519` accepts (a
  small-order `R`, the memo varied until one verifies); and an honest signature
  still verifies.
- `TestAnteHandlerRejectsSmallOrderPubKey` (`tm2/pkg/sdk/auth`): a funded
  address, the forged tx, ante rejection with `UnauthorizedError`, and the
  account's pubkey and sequence unchanged.
- `TestSmallOrderSessionKey` (`tm2/pkg/sdk/auth`): `MsgCreateSession` with the
  all-zero key fails `ValidateBasic` with `ErrInvalidPubKey`; a session
  installed with it directly through the keeper, as one created before this
  change would be, cannot be signed for: the `S = 0` forgery (memo varied
  until a small-order `R` verifies under `crypto/ed25519`) is rejected in ante
  with `UnauthorizedError`. With the gate removed, that forgery passes ante.
- `TestNodeParamsKeeperWillSetParam` (`gno.land/pkg/gnoland`): a
  `valset:proposed` entry with the identity key and power 10 panics; the same
  key with power 0 (a removal) passes.
- `TestValidateValidatorUpdates` (`tm2/pkg/bft/state`): a small-order ed25519
  update with power 20 is an error; with power 0 it is not.
- `TestSmallOrderValidatorCannotSign` (`tm2/pkg/bft/types`): the same forgery
  at `Vote.Verify` (`ErrVoteInvalidSignature`) and at `VerifyCommit`.
- `TestValidateRejectsSmallOrderEd25519Leaf` (`tm2/pkg/crypto/multisig`):
  behavior unchanged; the premise now calls `crypto/ed25519` directly for the
  primitive and `VerifyBytes` for the gate.
- `TestVerifyRejectsSmallOrderKey`
  (`gnovm/stdlibs/crypto/ed25519/ed25519_test.gno`): the realm-facing native
  rejects `R = identity, S = 0` under the 14 encodings. That signature verifies
  under 8 of them without the gate; the native goes through `VerifyBytes`, so
  the tm2 tests above cover the rest. Its encodings stay a hand-written list:
  Gno has no curve arithmetic to derive them, and completeness is asserted on
  the Go side.
- `TestAppHashCrossrealm38` (`gno.land/pkg/sdk/vm`): the pinned multistore root
  is re-derived for the stdlib test-source change, after bisecting that file as
  the only cause.
- `cd examples && go run ../gnovm/cmd/gno test ./...` and the `gnovm` stdlib
  suites, per the stdlib change.
