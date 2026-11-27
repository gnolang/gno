> **v0 — unaudited.** This package has not been formally audited.
> Use it in production at your own risk.

# `p/nt/multisig/v0`

State machine for an M-of-N treasury: signers, proposals, approvals,
a timelock, and expiry. It does not move coins. The factory realm
[`gno.land/r/nt/multisig/v0`](../../../../r/nt/multisig/v0) owns the
safes, authenticates callers with a live `cur`, and sends.

See that realm's README for `gnodev` / `gnokey` usage, and
[`gno.land/adr/prxxxx_multisig_treasury.md`](../../../../../../gno.land/adr/prxxxx_multisig_treasury.md)
for why this is a realm rather than a tm2 multisig account.

A safe's funds live at `chain.DerivePkgSubAddr(host, "s/<id>")`, the
address of `cur.Sub("s/<id>")` on the factory. Native coins and GRC20
are sent to that address. Spending them is a proposal.

Call proposals do not invoke the target. After execution the target
realm calls `Consume` and receives the payload once. Dynamic
`std.Call` is not available.
