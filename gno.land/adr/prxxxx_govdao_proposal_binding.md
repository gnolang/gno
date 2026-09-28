# ADR-prxxxx: Bind a govDAO proposal to the data it checked

## Context

A govDAO proposal can execute something other than what its creation-time
checks validated and its voters were shown. Four sites in the genesis realms:

- `NewValidatorProposalRequest` in `r/sys/validators/v0/proposal.gno` validates
  and renders the caller's `changes` slice, and its executor closure captures
  the same slice.
- `NewPropRequest` in `r/sys/validators/v2/poc.gno` captures the slice returned
  by the caller's `changesFn`.
- `ProposeSetRunSubmitters` in `r/sys/params/run_submitters.gno` captures
  `addrs`.
- The five generic factories in `r/sys/params/params.gno` capture a
  reference-typed `value`. Their description shows the key, not the value, so
  this one is latent.

Cross-realm calls pass a slice's header, not its array, so the caller keeps a
write handle after the proposal is stored. A rewrite during the vote changes
what executes.

`ProposeSetRunSubmitters` also has a weak floor. It takes the proposer from
`cur.Previous().Address()`, which is a keyless realm address when a member
relays the call through a deployed realm. It also assumed the builder is a
member, but GovDAO checks the account that creates the proposal, and a
`ProposalRequest` built by a non-member can be created later by a member.
Either way the list can name only a key that cannot sign or propose, and since
proposing needs MsgRun, the list can never be changed back.

## Decision

1. **Snapshot at construction.** Each site copies the caller's slice into an
   array its own realm allocates before rendering or capturing it: a `make` and
   `copy` in v0 and v2, `copyStrings` and `copyBytes` in `params.gno`. `copy`
   duplicates each element's fields but would still share a slice, map or
   pointer inside one. The copy is complete here only because `ValoperChange`
   and `validators.Validator` hold nothing but addresses, strings and integers.
2. **The floor reads the signer.** `ProposeSetRunSubmitters` requires
   `cur.Previous().IsUser()`, which accepts a user and the signer's own
   `/e/<addr>/run` realm and refuses a relaying realm. It then requires
   membership of the current implementation, through the new `dao.IsMember`:
   the proxy forwards it to a new `IsMember` method on the `DAO` interface,
   which `impl/v0` answers from `memberstore/v0`.

## Alternatives considered

- *Re-validate at execution.* The closure cannot recover what voters saw.
- *Copy inside `dao.NewSimpleExecutor`.* It does not know what a callback
  captures.
- *`IsUserCall` for the floor.* It refuses the MsgRun path.
- *Read `memberstore/v0` directly.* After an upgrade to an implementation
  that keeps its members elsewhere, a member of the new DAO fails that floor
  and a `memberstore/v0` member fails the new DAO's creation check, so nobody
  can replace the list, and r/sys/params cannot be redeployed to fix it.

## Consequences

The executed change is the one validated and shown. A realm can no longer build
a run_submitters proposal for a member, and a non-member can no longer build
one at all; the in-tree callers use MsgRun directly. The cost is one copy of at
most 40 entries or one list per proposal. The `DAO` interface gains
`IsMember`, so every implementation, including one GovDAO upgrades to, must say
who its members are.

Regression tests, each failing on the unfixed code:
`r/sys/validators/v0/proposal_capture_test.gno`,
`r/sys/validators/v2/poc_test.gno`,
`r/sys/validators/v2/filetests/crossrealm_rewrite_filetest.gno`,
`r/sys/params/run_submitters_alias_test.gno`,
`r/sys/params/params_factories_alias_test.gno`,
`r/sys/params/run_submitters_test.gno`,
`gno.land/pkg/integration/testdata/run_submitters_nonmember_builder.txtar` and
`gno.land/pkg/integration/testdata/run_submitters_dao_upgrade.txtar`, the last
failing on a floor that reads `memberstore/v0`.
