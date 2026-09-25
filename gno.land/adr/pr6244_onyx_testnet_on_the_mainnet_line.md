# ADR: Onyx, a testnet that runs mainnet's binaries and has no branch of its own

## Context

Pearl (test16) runs its own code line, cut on 2026-08-27 from a commit that
predates mainnet's launch code, with three gno-core validators and no
upgrade since. Mainnet (`gnoland-1`) has been upgraded three times in its
first two weeks, and #6228 wrote down the process for the next ones: every
release is a `vX.Y.Z` tag, cut first as a release candidate on the same
commit and rehearsed on the testnet through the real operator procedure. A
testnet on a different code line cannot rehearse anything.

The mainnet genesis builder (`misc/deployments/mainnet.gno.land/`) is the
current shape of a gno.land genesis: the `/v0` package set, a solo GovDAO T1
seed, namespace preregistration and enforcement, valoper seeding, the inert
code-submission policy with the gpao oracle, a two-pass fee-payer
measurement and a checksum-locked build. Most of its complexity is
mainnet-only: the independence-day allocation, the §126 transfer lock and
its exemption list, the §132 vesting guards, and the funding guards that
exist because mainnet has no faucet.

## Decision

1. **Onyx runs mainnet's exact binaries.** It launches on `v1.5.0`, the
   version mainnet runs on launch day, and is upgraded whenever mainnet is:
   from the next release on, onyx runs the candidate (`v1.6.0-rc.N`) and
   mainnet the final tag on the same commit. Its ledger names what it ran.
2. **No `chain/onyx` branch.** A branch that never differs in code is a trap
   for whoever checks it out. The deployment folder lands on `master` and is
   cherry-picked to `chain/mainnet`; the genesis is built from that tree
   (the package set is read from `examples/`, and `master`'s already
   differs from what the network runs); the `chain/onyx` launch tag marks
   that commit; its release carries `genesis.json` and its checksums, no
   binaries.
3. **The genesis is mainnet's shape with a testnet's money.** Derived from
   the mainnet builder: same package set (the package list's checksum is
   byte-identical), same bootstrap MsgRun (aeddi as sole T1 member with
   nine invitation points, `AllowedDAOs` locked), same seven namespaces,
   same names-admin gate, same inert policy with the gpao oracle as sole
   approver and `run_submitters` restricted to the seeded member. Removed:
   the allocation download and reconciliation, the transfer lock and
   exemption list, the vesting machinery, the founding-validator funding
   guard. Added: a typed `FUNDED_ACCOUNTS` sheet — the two faucet accounts
   and aeddi at `1e18` ugnot (pearl's ceiling), the gpao oracle at
   1,000,000 GNOT — with its own format, duplicate and int64-headroom
   guards, and a reconciliation of the shipped supply against the sheet plus
   the measured burn.
4. **gpao runs on onyx exactly as on mainnet.** The inert policy is on from
   block 1, the oracle is listed and funded in the genesis; an approver added
   post-genesis would have left every deployment parked in the meantime. The
   corollary holds too: `maketx run` is restricted to the seeded T1 member,
   because an open `run_submitters` would let `MsgRun` bypass package
   approval. A developer who needs it goes into `RUN_SUBMITTERS_EXTRA` by
   name.
5. **Sums in 64-bit integers.** Mainnet's reconciliation summed in awk
   doubles, exact under 2^53 (9.007e15) and fine for a 1.333e15 supply; a
   single faucet here holds 1e18. `sheet_total` sums in bash with an
   overflow check per addition.

## Alternatives considered

- **Derive from pearl's builder.** Rejected: pearl predates the `/v0`
  layout, the solo-T1 seed, namespace preregistration and the inert
  policy; onyx would have rehearsed a genesis mainnet does not have.
- **A `chain/onyx` branch, like every earlier testnet.** Rejected, see
  decision 2.
- **Launch on a release candidate of the next version.** Rejected: there
  is no new node code on `chain/mainnet`; a candidate would rehearse
  nothing.
- **Open `maketx run` on the testnet.** Rejected: it defeats the policy the
  testnet exists to rehearse. Named additions are the escape hatch.
- **Fund the gpao oracle from the faucet after launch.** Rejected: an
  unfunded approver approves nothing while every process reports healthy;
  the balance is asserted in the build against the same sheet that funds it.

## Consequences

- Onyx's `upgrades.json` starts with one genesis entry (`v1.5.0`) and gains
  a row per mainnet release, one candidate ahead of mainnet's.
- The build reproduces only from the `chain/mainnet` tree at the launch
  commit; a run from `master` fails on the package list by design, and the
  script's manifest comment says so.
- The validator's signing address holds nothing at genesis; the operator
  (aeddi) is funded and the faucet covers the rest.
