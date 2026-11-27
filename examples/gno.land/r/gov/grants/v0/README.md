> **v0 - Unaudited**
> This is an initial version of this package that has not yet been formally audited.
> A fully audited version will be published as a subsequent release.
> Use in production at your own risk.

# `r/gov/grants/v0` — GovDAO grants for realms

Projects, teams and DAOs ask GovDAO for ugnot **as code, not as people**. Identity is a realm's
package path, and money only ever goes to that grantee's own address. One GovDAO vote approves a
grant; each payment then waits for a public report on what it pays for and a window in which
GovDAO can stop it. A small first payment, an advance, is the only one paid without a report.

- **Teams and projects:** open an account in [`grants/wallet/v0`](../wallet/v0) named after a
  namespace you can deploy under (your registered name or your g1 address; `gnogolf/cup` for a
  second project) and apply, report, withdraw and spend from its page. Every account runs the same
  code, so voters only weigh what you ask for and who owns the account. Or call
  `grants.Apply` and `grants.Report` from a realm of your own; GovDAO members then read its code.
  The package doc of `grants.gno` lists every path coins can take.
- **GovDAO members:** each grant page prints the MsgRun that sponsors it (with your review, kept on
  the grant), cancels it, holds its next payment or freezes its grantee, and the checklist to go
  through first.
- **Everyone:** the desk's page shows the pool, what needs attention (payments to release, grants
  to expire, reports to read), and the open grants with their next step. Each grant has its own
  page: status, a one-click Claim or Expire form when one is due, payments, timeline, pitch,
  sponsor's review and reports. `ns/<name>` sums up a namespace, wallet accounts included.

The rules and their durations are on the realm's page; the lifecycle is the pure package
[`p/nt/grants/v0`](../../../../p/nt/grants/v0), and the reasoning is in
`gno.land/adr/prxxxx_govdao_grants.md`.
