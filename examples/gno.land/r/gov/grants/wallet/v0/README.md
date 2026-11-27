> **v0 - Unaudited**
> This is an initial version of this package that has not yet been formally audited.
> A fully audited version will be published as a subsequent release.
> Use in production at your own risk.

# `r/gov/grants/wallet/v0` — grant accounts for teams and projects

The way to take GovDAO grants from [`r/gov/grants/v0`](../../v0) without deploying code.

1. **Open** an account named after a namespace you may deploy under: your registered name or your
   g1 address, and `<namespace>/<project>` for a second project. You become the owner.
2. **Apply** from the account's page: a title, a pitch, and tranches such as
   `v3 design published: 1000; v3 live with 100 players: 4000` in whole GNOT.
3. Once GovDAO approves, **report** on each payment (with a link to what shows it). A payment falls
   due after its report and GovDAO's review window; anyone then releases it with Claim on the grant's
   page at the desk, and it arrives at the account's own address,
   `chain.PackageAddress("gno.land/r/gov/grants/wallet/v0#<name>")`. A first payment of at most the
   desk's advance share (20%) needs no report: claim it, then report on payment 2.
4. **Spend** with a memo; every spend and every change of owner is in the account's public journal.
   Hand the account to a multisig account or a DAO realm with `TransferOwnership`; it changes hands
   when the new owner calls `AcceptOwnership`, so an address that cannot act never ends up owning it.
   Offering the account to yourself withdraws an offer, even while GovDAO has the account frozen.
   An account page warns when its owner can no longer deploy under the account's namespace.
   A realm owner must call this wallet from its own code (`AcceptOwnership`, then `Spend`): a realm
   that never calls out, such as a safe that only lets targets pull payloads, cannot own an account.

There is no other way for coins to leave an account, no admin and no upgrade path. GovDAO can
freeze an account through the desk; a frozen account cannot apply, spend or change hands, and
receives no payment. If a realm owns the account, that realm's code decides who spends: voters
read it.
