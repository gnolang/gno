> **v0 — unaudited.** This realm has not been formally audited.
> Use it in production at your own risk.

# Vesting vault

`gno.land/r/nt/vesting/v0` is a factory. Anyone can lock a native coin
or a registered GRC20 on a cliff and a vesting curve. Each grant holds
its funds at `cur.Sub("g/<id>")`. The bank (or the GRC20 ledger) is the
balance sheet. The factory's own address is not a wallet.

The schedule math lives in [`gno.land/p/nt/vesting/v0`](../../../../p/nt/vesting/v0).
Why this is a grant contract, and not the genesis send-limit or the
quarantined timelock, is in
[`gno.land/adr/prxxxx_vesting_vault.md`](../../../../../../gno.land/adr/prxxxx_vesting_vault.md).

The beneficiary is `cur.Previous().Address()`. That is the signing user
on `maketx call` and on `maketx run`, and the calling realm on a
cross-call. Coin funding is stricter: `CreateCoin` requires
`cur.Previous().IsUserCall()`, because a `maketx run` script can spend
the `-send` envelope before this function sees it. A published realm
cannot relay that envelope either. The coins are credited to the realm
the message called, and an `OriginSend` banker will not spend them from
anywhere else.

GRC20 funding is an allowance pull. Approve **this factory's address**,
not the grant address. The token must already be registered in
`gno.land/r/nt/grc20reg/v0`.

A multisig safe can be the admin if its address is the admin and the
safe cross-calls `Revoke` or `TransferAdmin` while it is the current
realm. `r/nt/multisig/v0` does not invoke arbitrary realms on execute,
so that wiring is a later executor, not a dependency of this factory.

## Local node

From `examples/`:

```sh
gnodev
```

The dev node serves RPC at `127.0.0.1:26657` (chain id `dev`). Open
`http://127.0.0.1:8888/r/nt/vesting/v0` for the grant list. `devtest`
is pre-funded.

## Create a coin grant

`curve` is `1` for linear and `2` for step. `period` is `0` on a linear
curve. Times are unix seconds. This locks 1000 ugnot, waits 30 days,
then vests linearly over the following 335 days.

```sh
gnokey maketx call \
  -pkgpath "gno.land/r/nt/vesting/v0" \
  -func "CreateCoin" \
  -send "1000ugnot" \
  -args "g1...beneficiary" \
  -args "ugnot" \
  -args 1000 \
  -args 1767225600 \
  -args 2592000 \
  -args 31536000 \
  -args 0 \
  -args 1 \
  -args true \
  -args "contributor" \
  -gas-fee 1000000ugnot -gas-wanted 5000000 \
  -broadcast -chainid dev -remote 127.0.0.1:26657 \
  devtest
```

The `-send` amount must be exactly one coin, same denom and amount as
the grant. Anything else is rejected and no grant is stored.

## Create a GRC20 grant

Register the token, then approve the factory for `total` units. `tokenKey`
is the registry key (`realmPath.symbol`).

```sh
gnokey maketx call \
  -pkgpath "gno.land/r/nt/vesting/v0" \
  -func "CreateGRC20" \
  -args "g1...beneficiary" \
  -args "gno.land/r/demo/foo.TOK" \
  -args 1000 \
  -args 1767225600 \
  -args 0 \
  -args 31104000 \
  -args 2592000 \
  -args 2 \
  -args true \
  -args "monthly" \
  -gas-fee 1000000ugnot -gas-wanted 8000000 \
  -broadcast -chainid dev -remote 127.0.0.1:26657 \
  devtest
```

`31104000 / 2592000` is 12. A step curve requires the period to divide
the duration. A linear curve passes `0` for the period and `1` for the
curve.

## Claim, revoke, and pages

The beneficiary calls `Claim` with the grant id. The admin of a
revocable grant calls `Revoke`. The beneficiary calls `SetBeneficiary`.
The admin calls `TransferAdmin`.

```sh
gnokey maketx call \
  -pkgpath "gno.land/r/nt/vesting/v0" \
  -func "Claim" \
  -args 1 \
  -gas-fee 1000000ugnot -gas-wanted 3000000 \
  -broadcast -chainid dev -remote 127.0.0.1:26657 \
  beneficiary
```

`/r/nt/vesting/v0` lists the latest 20 grants. `/r/nt/vesting/v0:1`
is grant 1. A direct send to a grant address is not recorded in the
schedule and cannot be withdrawn by this contract.
