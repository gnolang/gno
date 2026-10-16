# secp256k1 validator consensus keys

## Context

Validators signed consensus votes with ed25519 only: `DefaultValidatorParams`
allowed nothing else and new validator keys were always ed25519.

An IBC light client of a gno.land chain on Ethereum has to verify commits
on-chain. The EVM has no ed25519 precompile, so verifying one ed25519
signature in Solidity costs hundreds of thousands of gas, while secp256k1 is
`sha256` plus the `ecrecover` precompile (about 3k gas). With secp256k1
validators a 3-validator commit verifies for roughly 170k gas in total
(`GnoLightClient` in gno-eth-ibc).

tm2 already had a secp256k1 `PubKey`/`PrivKey` (used for accounts) whose
signatures are 64-byte r||s over sha256 of the message, low-S enforced, and the
consensus code is key-type agnostic: votes and commits go through
`PubKey.VerifyBytes` and the validator set hash through the amino encoding of
the key.

## Decision

- `DefaultValidatorParams` allows both `/tm.PubKeySecp256k1` and
  `/tm.PubKeyEd25519`. Chains restrict the list in their genesis consensus
  params; a chain whose commits are verified on Ethereum should allow
  secp256k1 only.
- New validator keys default to secp256k1: `privval/signer/local` takes a
  `KeyType` (`GenerateFileKeyOfType`, `GeneratePersistedFileKeyOfType`), and
  `gnoland secrets init` gains `-validator-key-type`.
- P2P node keys and remote-signer transport keys stay ed25519: they never
  reach a light client.

## Alternatives considered

- **secp256k1 only.** Simpler invariant, but breaks every existing chain and
  test whose genesis carries ed25519 validators.
- **Keep ed25519 and verify on Ethereum with a ZK proof** (as the SP1
  Tendermint client of solidity-ibc-eureka does). Works for existing chains,
  but needs a prover service and much more machinery.

## Consequences

- Existing ed25519 chains are unaffected: their genesis lists ed25519 and their
  keys load as before.
- `gnoland secrets init` creates secp256k1 validator keys unless told
  otherwise; the default allow-list accepts them.
- The `valopers` realm accepts secp256k1 signing keys on chains using the
  default list; `valopers.txtar` no longer asserts the opposite.
- AtomOne's `10-gno` light client currently converts only ed25519 validator
  keys, so it cannot follow a secp256k1 gno chain until it learns the type.
- `TestReactorBasicSecp256k1` runs a 4-validator secp256k1 network.
