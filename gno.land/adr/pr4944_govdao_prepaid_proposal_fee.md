# ADR: non-members pay a prepaid fee to open a govDAO proposal

## Context

Issue #2520 asks that non-members of the govDAO be able to open proposals for a
fee, while members keep proposing for free.

A non-member can only build a `dao.ProposalRequest` in a `maketx run` script,
since a msg call cannot pass a struct argument. A run executes at
`gno.land/e/<signer>/run`, whose package address is the signer's own, so
`Run` in `gno.land/pkg/sdk/vm/keeper.go` sends `-send` from the signer to the
signer. `unsafe.OriginSend()` still reports the amount, so a check on it inside
`PreCreateProposal` passes with nothing paid, and passes again for every
proposal the same script creates.

## Decision

The fee is paid ahead of the proposal, in a msg call of its own:

- `impl.PayProposalFee` accepts exactly `ProposalFeeAmount` ugnot, only from a
  direct msg call (`cur.Previous().IsUserCall()`), so the coins land at the
  impl realm's address. It forwards them to `r/gov/dao/treasury/v0`, whose
  coins banker spends them on a treasury payment the members vote, and
  records one credit for the caller.
- `PreCreateProposal` lets members through, and spends one credit for a
  non-member. With no credit it returns an error naming the fee.
- `ProposalFeeAmount` defaults to 1 GNOT and changes through
  `NewProposalFeeAmountRequest`, which only members may open. Zero lets
  non-members propose for free.
- `NewAddMemberRequest` accepts a non-member proposer, who can nominate any
  address, themselves included, to T1 or T2. A member proposer still spends an
  invitation point. The tier filter still decides who votes.
- `PreCreateProposal` was the only member check for most requests, so a
  non-member who paid can now open any request whose constructor does not
  check membership itself. In `impl/v0` that is a treasury payment, a
  withdrawal, a promotion and a GRC20 token update; elsewhere it is every
  request the system realms build, `r/sys/params` (chain halt and parameter
  changes included), `r/sys/validators`, `r/sys/users`, `r/sys/namereg`,
  `r/sys/cla`, `r/sys/names`, `r/gnops/valopers` and `r/gnoland/blog`. Law
  changes, implementation upgrades, fee changes and tier changes keep their
  own member check. The vote, not the proposer, is the gate.

## Alternatives considered

- Checking `OriginSend()` in `PreCreateProposal`: free under `maketx run`, as
  described above.
- Requiring the proposal itself to arrive by msg call: impossible, since the
  request is a struct.
- A refundable deposit, returned when the proposal passes: needs payout logic
  that a fee does not, and the issue asks for a fee.

## Consequences

- A non-member signs two transactions per proposal: the payment, then the run.
- While `ugnot` is a restricted denom, the forward from `impl/v0` to the
  treasury is a restricted transfer and fails, so no non-member can pay. The
  mainnet genesis restricted it under Constitution §126.
- A member out of invitation points can still put a nomination to a vote by
  paying the fee from a second, non-member key.
- `PayProposalFee` does not check that `impl/v0` is still the active
  implementation, and an upgrade keeps it in `AllowedDAOs`. After an upgrade it
  still takes the fee into the treasury, for a credit the new implementation
  never reads.
- A credit outlives a fee change: it pays for one proposal at whatever the fee
  was when it was bought. A credit bought before the fee drops to zero stays
  unspent, since a free proposal does not consume it.
