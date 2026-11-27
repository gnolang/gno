# ADR: GovDAO grants for realms (`r/gov/grants/v0`, `r/gov/grants/wallet/v0`, `p/nt/grants/v0`)

## Context

Projects and teams (a game shipping versions, a marketing team with a budget) want GovDAO funding
that is automated, transparent and accountable. The ask: the grantee is **code**, a realm, not a
person, and it pays for itself with progress reports. Sub-DAOs should be able to build on the same
package. Two concerns were raised: when the grantee's code is the grantee's own, it could divert
funds; and GovDAO must be able to review that code quickly, with automated help where possible.

Constraints found in the code (master and mainnet are byte-identical here):

- Only `gno.land/r/gov/dao/impl/v0` may spend the treasury (`AllowedDAOs`). Allowlisting any other
  realm would also let it replace the implementation.
- A GovDAO proposal carries a func value, so it can only be created from a member's MsgRun; on
  mainnet MsgRun is allowlisted.
- A third-party realm cannot read GovDAO membership or proposal status, and `Executor` is an open
  interface: a proposal's title, description and executor strings prove nothing about its code.
- `SimpleExecutor.Execute` only runs when invoked from `r/gov/dao`. A panic in its callback aborts
  execution and cannot be denied; a returned error can (`ExecuteOrRejectProposal`).
- Code at a public path cannot change. A **private** realm can be redeployed by its creator at
  the same path and address (#6164, #6171); a realm cannot detect that about another one, and the
  chain exposes no package source to compare.
- `cur.Sub(name)` mints a sub-realm identity `host#name` with an address of its own, which a
  banker can send from (as `r/nt/commondao` does).

## Decision

A pure lifecycle package, a GovDAO desk on it, and a wallet that hosts grantee accounts, with no
change to GovDAO.

- **Pool.** The desk's address holds the funds. Whoever holds the money GovDAO assigns to grants
  sends it there: a treasury payment proposal when the GovDAO treasury holds funds, or the holder
  of the Ecosystem Treasury account. Coins leave only to a grantee's address (derived from the
  caller) or, by proposal (`NewReturnRequest`, `NewCloseRequest`), to the address the vote names,
  normally where they came from. Approval requires full coverage on
  top of what is committed, so every approved grant is payable.
- **Grantees are code, hosted by default.** `r/gov/grants/wallet/v0` holds one account per
  namespace (`gnogolf`, or `gnogolf/cup`), opened by whoever may deploy under that namespace
  (`r/sys/names`), owned by one address: a user, a multisig account, or a realm whose code calls
  the wallet (a handover takes the new owner's `AcceptOwnership`, so no account can be handed to an
  address that never acts, such as another account's), and identified as
  the sub-realm `wallet/v0#<name>` with its own address. Only the owner spends, each spend with a
  memo in the account's public journal. GovDAO reviews this code once, so a grant to an account
  needs no code review of its own: Thomas's diversion concern is bounded by code every account
  shares, and Manfred's review becomes "who owns it" (and, if the owner is a realm, that realm's
  code). The pages do not claim GovDAO reviewed the wallet; they link its source. A team may still apply from a realm of its own
  (`r/gnogolf/grant`); that realm's code must then be read, and every grant page says which case
  it is and prints the matching checklist.
- **One vote per grant, a report before every payment.** Payment *i* is due after a report on it,
  a 14-day window, and at least the grant's period (≥ 30 days) after payment *i−1*. The one
  exception is an advance: a first payment of at most 20% of the grant, due 14 days after
  approval. A single-payment grant is therefore paid after its work is reported. A report on an
  advance not yet claimed is refused, since claiming it would discard the report; so is a report
  past its deadline, so a grantee cannot take back an expiry anyone may already call. Reports carry an
  optional link (a gno.land path or an https URL, restricted to characters that cannot end a
  markdown link, and without `$`, which on gnoweb opens a page that can prefill a transaction). The periods add up to at most a year; a slower grantee takes longer.
- **A named sponsor.** `NewApproveRequest` takes a review statement (who controls the grantee
  and how they are known, what was checked, any conflict of interest), checked as the approval
  will check it, and must be built in the sponsor's own MsgRun; the proposal shows the review in
  full and the grant keeps it with the address that built the request. GovDAO records the member
  who filed it as the proposal's author; a request built by one address and filed by another
  shows both.
- **Brakes.** Cancel by proposal at any time; a cancellation fails if the grant's status changed
  since it was proposed (a request approved meanwhile is a funded grant voters did not weigh).
  The author of the newest proposal, if it is the cancellation of that grant (exact stored title
  and an executor created by the desk, never called) and the desk built it for that author in the
  same block, can push the next payment back 14 days, once per payment. The same-block rule stands
  in for a status check GovDAO does not expose: a hold cannot follow a vote already lost. GovDAO can freeze a grantee (`NewFreezeRequest`): the
  desk pays it nothing, it cannot apply or be approved, and a wallet account cannot spend or change
  hands, until it is unfrozen. Freezing a realm also stops identities it hosts from applying or
  being approved, even when it never applied under its own path; those already approved stay
  payable until each is frozen, and their pages do not show them as frozen. Other wallet accounts
  of the same namespace are grantees of their own, and the freeze proposal says so. The grantee can
  withdraw. Anyone can expire a request pending 60 days, or a grant whose next payment is still
  unreported 60 days after its period. `NewCloseRequest` stops new requests and approvals and
  returns the uncommitted pool; running grants go on unless the vote also cancels them, so a move
  to a v1 desk does not re-vote every running grant; a later close can still cancel them all. Close walks only active grants, which only
  GovDAO creates.
- **Bounded work.** Active grants and pending requests are listed separately, pending ones oldest
  first, so a flood of requests sinks below the approved grants. Pitches, reports and reviews are
  capped in bytes and lines, a page shows the last three reports, namespace counts stop at 1,000,
  and journals and account lists are paged.
- **Identity.** `GranteeOf`: the calling realm, never a user or a MsgRun. Sub-realm identities
  (`host#sub`) are accepted, which is how wallet accounts and DAOs hosted per the Common DAO Spec
  apply; each identity has one pending request at a time.
- **Execution.** Callbacks check, then write. An underfunded approval panics so it stays
  executable (tested on a node); a stale one returns an error so it can be rejected. Every change
  of a grant's status emits an event naming it, including a superseded request and each grant a
  closing vote cancels.
- **Voters and observers.** Proposal descriptions lead with facts the desk computes; grantee and
  sponsor words are escaped, clamped and labelled, and quoted line by line on the pages, so they
  cannot open a heading, link or code fence.
- **Pages.** gnoweb extensions only: cards, alerts, collapsible sections, icons, and one-click
  call forms for the calls anyone may make (Claim, Expire) and for an account's owner. "Needs
  attention" lists payments to release, grants to expire, and reports GovDAO can still object to.
  A pending grant's page names the title its vote carries in GovDAO's proposals rather than
  linking a proposal id, which a member could point at a look-alike.

## Alternatives considered

- **A template each team copies and GovDAO diffs.** The first version. Nothing on chain can check
  the copy, a private copy can be redeployed after the vote, and copying is the habit Gno's own
  guide argues against. Replaced by the hosted wallet; own realms remain possible with a review.
- **One treasury payment per payment.** No custody, but a supermajority vote per payment.
- **Allowlist the desk in `AllowedDAOs`.** Grants it `UpdateImpl`. Rejected.
- **Change GovDAO core** (a spender allowlist, a batch executor, a status getter). Cleaner, but a
  new impl and an upgrade vote; listed as follow-ups.
- **A desk any DAO can run, now.** Grantees can already be any DAO (a realm, a hosted `host#dao`
  identity, or the owner of a wallet account), and the lifecycle in `p/nt/grants/v0` knows nothing
  of GovDAO. The desk does: its proposals, its hold and its pages are GovDAO's. Making it generic
  before a second grantor exists would mean guessing what "approve", "hold" and "freeze" mean for
  DAOs that decide differently. Deferred; see "Next: a desk for any DAO".
- **Time-based vesting** (#6542 proposes `r/nt/vesting/v0`). A vault releases coins on a schedule
  whatever is delivered, and only a user's direct call can fund one, so GovDAO could not. Grants
  here are paid against reports instead. The two fit side by side: a funded team can vest, from its
  owner's address, what it pays its own contributors.
- **Reuse `p/moul/grants/v0`.** Its grantees are addresses and its members vote on each milestone; here identity is the caller's realm, the whole grant is reserved, and payments follow
  reports without a vote unless GovDAO objects.
- **Link the open vote from a member-supplied proposal id.** A member could point it at a
  look-alike proposal; the page names the title instead.
- **Refuse sub-realm identities.** Shuts out the wallet and every DAO hosted per the Common DAO
  Spec.

## Consequences

- Deployable with the current GovDAO, by the `gov` namespace owner and, on mainnet, an approver.
- Worst case after approval: what was paid before GovDAO cancels or freezes, bounded by the
  advance, the payment sizes, the spacing and the window. Coins already in a wallet account can be
  frozen, not taken back; coins already sent to a realm of its own cannot be frozen at all.
- A wallet account's owner can hand it over after the vote (journaled, refused while frozen);
  voters are told so. The wallet calls desk v0: a v1 desk needs a v1 wallet, and owners move their
  balance by spending it to the new account.
- One member can delay one payment 14 days, no more, and only while filing the cancellation; the
  hold lapses after that even if the vote is still open, until GovDAO exposes proposal status.
- **Transfer lock.** Coins leave the pool and wallet accounts through `bank.SendCoins`. If GovDAO
  ever restores `bank:p:restricted_denoms=["ugnot"]` (empty on mainnet, read 2026-11-27), every
  payment, return and spend fails until the pool and each account's address are in
  `auth:p:unrestricted_addrs`: whitelist them in the same params vote, or before funding.
- **Mainnet today (read 2026-11-26):** the GovDAO treasury realm holds no GNOT, so the pool would
  be funded from the Ecosystem Treasury account; GovDAO has 6 T1 members and no T2/T3, so every
  vote needs 4 of 6; only the 3 MsgRun-allowlisted addresses can sponsor, cancel, hold or freeze;
  a team can self-register only a `nym-…123` name or use its g1 address as namespace, so a plain
  account name such as `gnogolf` needs a GovDAO name decision first.
- **Open for GovDAO, not solved by code:** the draft Constitution asks for an accountability
  framework, a known human identity for Ecosystem Treasury recipients (here, the sponsor's review
  names the controller), purpose-bound treasuries (one pool serves every purpose here), and a
  Simple Majority freeze (GovDAO votes by supermajority, so `NewFreezeRequest` does too).
- Non-ugnot coins sent to the pool are stuck. Closed grants are never pruned.
- Follow-ups: a proposal-status getter in `impl` (a hold that lasts until its vote ends), a
  private-package getter, a batch executor (fund and approve in one vote).

## Next: a desk for any DAO

If another DAO, a sub-DAO for instance, wants to fund grants, the desk becomes a hosted realm, as
the wallet is for grantees, so that no grantor has to write and audit its own:

- One realm holds a desk per grantor, each a sub-identity (`grants#<name>`) with its own pool
  address, its own `Rules` and its own pages.
- Each desk names an **authority**: the identity whose crossing calls may approve, cancel, freeze,
  return funds and close. Any DAO whose executed proposals can make a crossing call qualifies;
  `p/nt/commondao` hands its executors a DAO-scoped sub-identity that can.
- GovDAO keeps a thin adapter with what is specific to it: the proposal builders, the sponsor's
  review, and the hold, which reads GovDAO's proposals. A generic desk would offer a pause by its
  authority, or by guardians the authority names, instead.
- The wallet names the desk it applies to. `wallet/v0` is bound to this desk, so its accounts move
  to a `wallet/v1` account by spending their balance there.
