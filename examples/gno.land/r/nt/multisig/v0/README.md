> **v0 — unaudited.** This realm has not been formally audited.
> Use it in production at your own risk.

# Multisig treasury

`gno.land/r/nt/multisig/v0` is a factory. Anyone can create an M-of-N
safe. Each safe has its own address, derived with `cur.Sub`, and only
its signers can move what that address holds.

The rules live in [`gno.land/p/nt/multisig/v0`](../../../../p/nt/multisig/v0).
The design notes are in
[`gno.land/adr/prxxxx_multisig_treasury.md`](../../../../../../gno.land/adr/prxxxx_multisig_treasury.md).

This is not the tm2 multisig that holds the genesis treasuries. Those
are accounts. This is a realm: it can queue a payment, hold GRC20, and
show its state. It does not move the GovDAO or NT balances by being
deployed.

## What a proposal can do

- Send native coins (`ProposeCoin`).
- Send a GRC20 registered in `gno.land/r/nt/grc20reg/v0` (`ProposeGRC20`).
- Authorize another realm to pull a payload once (`ProposeCall`, then
  that realm's `Consume`). The safe does not call the target. There is
  no dynamic cross-realm call in the VM today.
- Add or remove a signer, change the threshold, change the timelock,
  pause, or unpause.

Pause blocks coins, GRC20, and call authorizations. Signer changes and
unpause still run. Removing a signer drops that signer's outstanding
approvals. A proposal that has reached the threshold cannot be
cancelled until someone revokes a vote. Revoking under the threshold
restarts the timelock.

Signers are whatever `cur.Previous().Address()` is. A user must call
with `maketx call`. `maketx run` is a different address. A realm can be
a signer if its own address is on the list.

Limits: 32 signers, 64 pending proposals, 128 retained proposals
(older finished ones are dropped; an unconsumed call authorization
is kept), 1024 safes on this factory, delay up to 30 days, expiry
at most 366 days out and always later than `now + delay`. A native
denom has to match the chain's denom grammar. One signer can fill
the pending list; any other signer can cancel those proposals while
they are below the threshold.

## Local node

From `examples/`:

```sh
gnodev
```

The dev node serves RPC at `127.0.0.1:26657` (chain id `dev`) and the
web UI at `http://127.0.0.1:8888/r/nt/multisig/v0`. `devtest` is
pre-funded. `gnokey list` prints its address. Add more keys for a real
threshold:

```sh
gnokey add alice
gnokey add bob
```

Fund them from `devtest` with `gnokey maketx send` if you want more
than one signer. A 1-of-1 safe owned by `devtest` is enough to see the
flow.

## Create a safe

Signers are a comma-separated list. `threshold` is how many must
approve. `delay` is the timelock in seconds; `0` executes as soon as
the threshold is met.

```sh
gnokey maketx call \
  -pkgpath "gno.land/r/nt/multisig/v0" \
  -func "Create" \
  -args "ops" \
  -args "g1...alice,g1...bob,g1...carol" \
  -args 2 \
  -args 3600 \
  -gas-fee 1000000ugnot -gas-wanted 5000000 \
  -broadcast -chainid dev -remote 127.0.0.1:26657 \
  devtest
```

The call returns the safe id. Its address:

```sh
gnokey query vm/qeval \
  -data 'gno.land/r/nt/multisig/v0.SafeAddress(1)' \
  -remote 127.0.0.1:26657
```

## Put coins in

Send to the safe address, not to the factory. A `-send` on `Create`
lands on the factory and stays there.

```sh
gnokey maketx send \
  -to "<safe address>" \
  -send "1000000ugnot" \
  -gas-fee 2000ugnot -gas-wanted 200000 \
  -broadcast -chainid dev -remote 127.0.0.1:26657 \
  devtest
```

`Deposit` is the alternative when the coins are attached to a call.
It only works as a direct user call, and the amount has to fit in
that message's `-send`:

```sh
gnokey maketx call \
  -pkgpath "gno.land/r/nt/multisig/v0" \
  -func "Deposit" \
  -args 1 -args "ugnot" -args 1000000 \
  -send "1000000ugnot" \
  -gas-fee 1000000ugnot -gas-wanted 3000000 \
  -broadcast -chainid dev -remote 127.0.0.1:26657 \
  devtest
```

GRC20 is an ordinary transfer to the same address, then `WatchToken`
with the registry key (`realmPath.SYMBOL`) so the page shows the
balance.

## Propose, approve, execute

`expiresAt` is a unix timestamp, later than now plus the timelock.

```sh
gnokey maketx call \
  -pkgpath "gno.land/r/nt/multisig/v0" \
  -func "ProposeCoin" \
  -args 1 \
  -args "g1...dest" \
  -args "ugnot" \
  -args 500000 \
  -args 1800000000 \
  -args "payroll" \
  -gas-fee 1000000ugnot -gas-wanted 5000000 \
  -broadcast -chainid dev -remote 127.0.0.1:26657 \
  alice
```

The call returns the proposal id. Each signer:

```sh
gnokey maketx call \
  -pkgpath "gno.land/r/nt/multisig/v0" \
  -func "Approve" \
  -args 1 -args 1 \
  -gas-fee 1000000ugnot -gas-wanted 3000000 \
  -broadcast -chainid dev -remote 127.0.0.1:26657 \
  alice
```

After the threshold and the timelock:

```sh
gnokey maketx call \
  -pkgpath "gno.land/r/nt/multisig/v0" \
  -func "Execute" \
  -args 1 -args 1 \
  -gas-fee 1000000ugnot -gas-wanted 5000000 \
  -broadcast -chainid dev -remote 127.0.0.1:26657 \
  bob
```

`Revoke` and `Cancel` take the same two ids. A second `Execute` fails.
`ProposalStatus` and `Balance` are reads:

```sh
gnokey query vm/qeval \
  -data 'gno.land/r/nt/multisig/v0.ProposalStatus(1, 1)' \
  -remote 127.0.0.1:26657
```

## Call authorization

`ProposeCall` records `targetPkg` and a payload. `Execute` does not
enter that realm. The target, in its own transaction, calls:

```sh
gnokey maketx call \
  -pkgpath "gno.land/r/nt/multisig/v0" \
  -func "Consume" \
  -args 1 -args 1 \
  ...
```

`Consume` checks that the caller realm's path is `targetPkg`, returns
the payload, and refuses a second call. The target then does its own
work. That is the whole callback: the safe never runs the target's code.

## Pages

- `/r/nt/multisig/v0` lists safes.
- `/r/nt/multisig/v0:<id>` shows signers, balances, and proposals.
- `/r/nt/multisig/v0:<id>/<proposal>` shows one proposal and help links.

Only `ugnot`, denoms added with `WatchDenom`, and tokens added with
`WatchToken` are shown. Listing every denom would let a stranger set
the gas cost by sending a new one.
