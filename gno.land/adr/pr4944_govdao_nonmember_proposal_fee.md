# ADR: non-members propose a member for a fee

## Context

Issue #2520 asks that non-members of the govDAO be able to open proposals for a
fee, while members keep proposing for free.

A non-member can only build a `dao.ProposalRequest` in code, a `maketx run`
script or a realm, since a msg call cannot pass a struct argument. A run
executes at `gno.land/e/<signer>/run`, whose package address is the signer's
own, so `Run` in `gno.land/pkg/sdk/vm/keeper.go` sends `-send` from the signer
to the signer. `unsafe.OriginSend()` still reports the amount, so a fee check
on it inside `PreCreateProposal` passes with nothing paid, and passes again for
every proposal the same script creates.

## Decision

Non-members get one proposal, adding a member, through one entrypoint:

- `impl.ProposeAddMember(addr, tier, portfolio)` takes plain arguments, so it
  is a msg call, and refuses any caller but a direct one
  (`cur.Previous().IsUserCall()`). A non-member sends exactly
  `ProposalFeeAmount` ugnot with it. The coins land at the impl realm, which
  forwards them to `r/gov/dao/treasury/v0`, where a treasury vote can spend
  them. A member sends nothing.
- It builds the request and calls `dao.MustCreateProposal` itself.
  `PreCreateProposal` admits a non-member only when the proxy was entered from
  `gno.land/r/gov/dao/impl/v0`, and only `ProposeAddMember` creates a proposal
  from that realm. Every other request stays members-only, whoever
  built it.
- The request carries no unforgeable mark of its kind: its executor is
  unexported in the proxy, and a filter can be lifted from a genuine request
  and paired with another executor. Checking the route rather than the request
  is what keeps the other kinds closed.
- `NewAddMemberRequest` stays members-only, so a non-member's request never
  exists outside `ProposeAddMember`, where a member could submit it without
  spending an invitation point. A member proposer spends a point either way.
- `ProposalFeeAmount` defaults to 1 GNOT and changes through
  `NewProposalFeeAmountRequest`, which only members may open. Zero makes the
  proposal free.

## Alternatives considered

- Checking `OriginSend()` in `PreCreateProposal`: free under `maketx run`, as
  described above.
- A prepaid credit, bought by a msg call and spent by any later proposal: it
  opens every request kind to non-members, `r/sys/params` halts included, and
  takes two transactions per proposal.
- A refundable deposit, returned when the proposal passes: needs payout logic
  that a fee does not, and the issue asks for a fee.

## Consequences

- While `ugnot` is a restricted denom, no non-member can pay: one who is not
  whitelisted cannot send the fee at all, and for one who is, the forward from
  `impl/v0` to the treasury is a restricted transfer and fails. The mainnet
  genesis restricted it under Constitution §126.
- A member out of invitation points can still put a nomination to a vote by
  paying the fee from a second, non-member key.
- Opening another kind to non-members takes another entrypoint in `impl/v0`
  beside `ProposeAddMember`, taking the fee the same way.
