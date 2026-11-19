# ADR: Verifiable randomness from drand, checked on-chain

## Context

Realms that need randomness (games, raffles, lotteries, fair selection) have
no good source on gno.land. Execution is deterministic, so anything computed
inside a transaction is a function of data that already exists: block height,
time, a stored value. Anyone can dry-run a call and only send it when the
outcome suits them, and a proposer can reorder or drop transactions.

On Cosmos, Nois solved this with drand: a contract requests randomness pinned
to a drand round that is not published yet, relayers deliver the beacon, and
the chain verifies its BLS signature before using it. drand's main networks
sign on BLS12-381, which gno.land has no native support for. Verifying it in
interpreted Gno is far too expensive, and adding a BLS12-381 native means a
VM change and a chain upgrade.

drand also runs **evmnet** (chain hash `04f1e906...6ec8c3`, scheme
`bls-bn254-unchained-on-g1`, one beacon every 3 seconds), which signs on
BN254 for EVM precompile compatibility. gno.land already ships BN254 as
natives (`crypto/bn254`, EIP-196/197 layout), plus `crypto/keccak256`,
`crypto/modexp` and `crypto/sha256`.

## Decision

1. **Use drand evmnet and verify every beacon on-chain** with the existing
   natives. No VM change, no new native, no trusted relayer.
   - Message: `keccak256(uint64_be(round))`.
   - Hash to G1: RFC 9380 `expand_message_xmd` (keccak256) and the
     Shallue-van de Woestijne map, DST
     `BLS_SIG_BN254G1_XMD:KECCAK-256_SVDW_RO_NUL_`, written in Gno.
   - Check: `e(sig, G2) == e(H(m), pk)` as one `bn254.PairingCheck`.
   - Randomness: `sha256(signature)`, identical to what drand publishes.
2. **Push field arithmetic onto `modexp`.** A schoolbook 256x256 product in
   interpreted Gno cost about 640k gas per multiplication. Squaring,
   inversion and square roots go to `modexp` (x^2, x^(p-2), x^((p+1)/4)),
   multiplication uses quarter-squares, `ab = ((a+b)/2)^2 - ((a-b)/2)^2`, and
   limbs are converted with `encoding/binary`. The SVDW candidates x2 and x3
   are computed only when needed, and the G2 generator is negated once as a
   constant instead of negating H(m) per call. A full `Verify` went from
   about 24M to about 9.6M gas; an on-chain `Submit` (verify plus storage)
   measures 17.9M gas on gnodev.
3. **Split library and realm.**
   - `gno.land/p/drand/v0`: pure verifier and helpers (`Verify`,
     `HashToG1`, `Randomness`, `Derive`, `NewRand`, `RoundAt`, `TimeOf`).
   - `gno.land/r/drand/v0`: requests and beacons.
     `Request`/`RequestAfter` pin an id to the first round published at least
     `SafetyGap` (5s) after block time; `Submit` is permissionless and stores
     only requested rounds; `Get`, `MustGet` and `Rand` return
     `sha256(beacon || requester || ":" || id)`, so requests sharing a round
     get independent values.
   - `gno.land/r/drand/coinflip/v0`: a minimal consumer.
4. **Pull model, two transactions.** The consumer requests in one transaction
   and reads in a later one. Nothing is pushed into consumer realms.
5. **No admin.** The realm has no owner and no settable parameter. A
   randomness source nobody can retune is easier to trust; a change ships as
   a new version.
6. **Relayer in `contribs/gnodrand`.** Stateless: it polls `PendingRounds`,
   fetches published rounds from drand mirrors, verifies them locally with
   drand/kyber (so a broken mirror never costs gas) and batches `Submit`
   calls. The fee is set per beacon and scaled by the batch, because the
   whole fee is deducted whatever the transaction uses. Any number of
   relayers can run; a duplicate `Submit` is a no-op.
7. **Test vectors from the reference.** `vectors_test.gno` holds real evmnet
   beacons and hash-to-G1 outputs computed by drand/kyber.
   `contribs/gnodrand/internal/evmnet` keeps a `math/big` SVDW reference
   checked against kyber on 204 messages, which pins the constants. An
   opt-in e2e test (`-tags e2e`) runs gnodev against live drand from request
   to settlement.

## Alternatives considered

- **drand quicknet (BLS12-381).** The network Nois used, with a stronger
  curve, but it needs a new native and a chain upgrade.
- **Trusted relayer or oracle committee, no on-chain verification.** Cheap,
  but randomness would be only as honest as the relayer.
- **Block data or the latest stored beacon, used in the same transaction.**
  Known before the transaction commits, so it can be dry-run and gamed.
- **Store every round.** 28,800 beacons a day, about 1,500 GNOT a day at
  current prices, with no gain: a request must wait for a future round
  anyway.
- **Callbacks into consumer realms on `Submit`.** Saves the consumer a
  transaction, but invoking caller-supplied code under the realm's authority
  is a known attack surface. Deferred until a consumer needs it.
- **Hint-based verification** (the relayer supplies inverses and square
  roots, the realm only checks them). Not needed at the measured gas cost.

## Consequences

- Any realm gets unbiasable, verifiable randomness in about 9 to 15 seconds
  (5s gap, up to a 3s drand period, relayer poll and one block), with no
  chain upgrade.
- Security is BN254's (about 100 bits) rather than BLS12-381's (about 128),
  the trade-off Ethereum's precompiles make; fine for games and raffles.
  Trust reduces to drand's threshold assumption.
- Liveness needs at least one relayer. Correctness does not: an invalid
  beacon cannot pass `Verify`.
- A relayer pays about 0.018 GNOT of gas plus a 35,100 ugnot storage deposit
  per delivered round, shared by every request on that round.
- A `Request` costs about 6.5M gas plus a 2,415-byte storage deposit (about
  0.24 GNOT at 100 ugnot/byte), more than the beacon itself. Most of it is
  avl node and object overhead; a more compact request store is the obvious
  next optimization.
- `SafetyGap` covers normal drift between block time and wall clock. A
  proposer setting block time far behind real time could pin a request to an
  already published round.
- On a network with namespace enforcement, deploying under `drand` needs the
  namespace registered.
