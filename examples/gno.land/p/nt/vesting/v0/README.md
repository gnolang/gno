> **v0 — unaudited.** This package has not been formally audited.
> Use it in production at your own risk.

# `p/nt/vesting/v0`

State machine for one cliff-and-vest grant. It does not move coins.
The factory realm [`gno.land/r/nt/vesting/v0`](../../../../r/nt/vesting/v0)
owns the grants, authenticates callers with a live `cur`, and sends.

Design notes, the survey of existing vesting and timelock code, and
the threat model are in
[`gno.land/adr/prxxxx_vesting_vault.md`](../../../../../../gno.land/adr/prxxxx_vesting_vault.md).

A grant is a schedule plus books:

- `start`, `cliff`, `duration` are unix seconds. `cliff` is an offset
  from `start` and must sit inside the duration. `cliff == 0` vests
  from the start. `cliff == duration` unlocks the whole amount at the
  end.
- `curve` is linear (`1`) or step (`2`). A step curve also has a
  `period` that divides the duration. Nothing vests until a whole
  period has elapsed.
- `total` is a positive int64. `VestedAmount` is `floor(total * elapsed / duration)`
  (or periods, for a step curve). The remainder stays in the vault
  until the end, when the function returns `total` exactly.
- Before `start + cliff`, vested is zero. At the cliff, the amount
  accrued since `start` unlocks in one jump.
- `Claim` moves vested-but-unclaimed into `claimed`. `Revoke` on a
  revocable grant freezes vested, returns `total - vested` to the
  admin, and leaves the vested part claimable. A second revoke fails.
  An irrevocable grant cannot be revoked. There is no pause: an admin
  must not be able to freeze tokens that have already vested.
- The beneficiary may name a new beneficiary. The admin may name a new
  admin. Neither may set the address to the factory or the grant.

The package rejects a negative clock, a total that does not fit in
int64 arithmetic, and a label that contains a control character.
Rounding favors the vault. `Grant.Invariant` checks, at a unix second:

- `claimed + returned <= total`, and both are non-negative
- the stored address is still `DerivePkgSubAddr(host, "g/<id>")`
- while the grant is open, nothing has been returned
- after revoke, `returned == total - frozen` and `frozen >= claimed`
- `claimed` never exceeds the curve, and `releasable == curve - claimed`

After a revoke, `claimed + releasable + returned == total`.
