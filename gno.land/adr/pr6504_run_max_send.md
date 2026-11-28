# ADR: Cap what a run moves out of the signer's account

## Context

A `MsgRun` executes as `gno.land/e/<signer>/run`, and that realm's address is
the signer's own. The run's `Send` is therefore a transfer from the signer to
the signer: it moves nothing and bounds nothing. Any code that holds the run's
`cur`, the script or a realm the script handed it to, can build a
`BankerTypeRealmSend` banker over the signer's address and send the whole
balance. The signer had no field saying how much the run may cost.

`MsgCall` and `MsgAddPackage` do not have this gap: their `Send` lands in the
callee's address, and the callee spends from there.

## Decision

`MsgRun` gains `MaxSend`, set by `gnokey maketx run -max-send`. During the run,
the VM's banker sums every coin transfer whose sender is the signer and whose
recipient is another address, and panics on the transfer that takes the sum
past `MaxSend` with `run sends more coins than its max send allows`. The run
then fails and its effects revert.

An empty `MaxSend` refuses every such transfer, outside replayed history (see
Consequences). A run that sends nothing is unaffected.

The check sits in `RunSDKBanker.SendCoins`, which wraps `SDKBanker.SendCoins`,
the one path by which Gno code transfers coins between addresses. Only `Run`
builds a `RunSDKBanker`; every other message keeps the plain `SDKBanker`.

A realm can keep a banker built over a run realm and spend it in a later
transaction, under that transaction's banker. So `SDKBanker.SendCoins` refuses
a send out of an account that holds a public key, and `RunSDKBanker` skips that
check only for its own signer. No package account holds a public key, and a
run realm's address is its signer's, so only a run's own banker spends a user
account the ante handler has given a public key. Without this check the cap
bounds nothing.

A hardfork replay with `--skip-genesis-sig-verification` skips each replayed
signature before the line that stores its key, so every replayed account would
start the fork without one, and a banker kept on the source chain would spend
it there. The ante handler therefore stores the key a skipped replayed
signature carries, where the account holds none and the key hashes to the
signer's address.

`MaxSend` is the last field of `MsgRun`, proto field 5, and carries
`json:"max_send,omitempty"`. A run that does not set it encodes to the same
binary and signs the same bytes as before, so a client that predates the field
keeps working for every run that sends nothing.

## Alternatives considered

- **Reuse `Send` as the cap.** No new field, but `Send` on a call means coins
  that land in the callee, while on a run it would mean coins the script may
  move. One flag with two meanings was judged worse than a second flag.
- **An empty cap means unlimited.** Nothing breaks, and a signer who runs a
  script without reading it stays unprotected. Rejected: the cap exists for
  that signer.
- **Exempt every genesis run that sets no cap.** InitChain marks a new
  chain's own genesis transactions as a replay too, so its runs without a cap
  would spend without bound. Rejected: only replayed history, which carries
  its source block height, is exempt.
- **Leave stored bankers alone.** The cap would then stop only a realm that
  drains in the same transaction, which no realm keeping a banker needs to do.
- **Pin a run realm's banker in `chain/banker`**, as `BankerTypeOriginSend`
  is pinned, so the persistence walk refuses to store it. A chain upgraded in
  place keeps the stdlib source in its store and never reloads it, so the pin
  would reach only a fresh genesis, and it added gas to every `NewBanker` and
  every realm send. The Go check ships with the binary.
- **Check in the bank keeper, beside the session spend limit.** It would also
  catch transfers outside the VM, but the storage deposit and the fee already
  have their own caps and use the unrestricted send. Checking in the VM's
  banker counts exactly what Gno code moves.

## Consequences

- A run that moves coins out of the signer's account must pass `-max-send`.
  Scripts and wallets that build such runs need the field. This is a change to
  DeliverTx results, so it ships as a hardfork, and every node takes it at one
  height. A halt-height upgrade in place is enough: the change is Go code
  alone and touches no stdlib source the chain keeps in its store.
- The cap covers native coins only. A GRC20 balance is a number the token's
  realm keeps in its own state, not a coin in the bank, so the cap never sees
  it move.
- A user can no longer hand a realm a standing banker over their own account:
  a realm may keep one, but it spends only inside that user's own later run,
  within that run's `-max-send`.
- The storage deposit stays bounded by `MaxDeposit` and the fee by the gas
  fields, neither counted against `MaxSend`.
- `Send` on a run keeps its current behaviour.
- A replayed run from a past chain, delivered at its source block height,
  that sets no `MaxSend` runs uncapped, so history signed before the field
  replays as it ran. A fresh chain's own genesis runs at height 0 and is
  capped, and so is a replayed run that sets a cap. A replayed transaction
  that spent a user's account through a banker kept from that user's run
  fails, since the account check has no replay signal and the replay stores
  the key the user's replayed signature carries, whether or not it checks
  signatures. A source chain's own genesis transactions carry no metadata
  and replay at height 0, so one of them that sent coins from a run fails
  too.
- An account whose only signed transactions sit at height 0 of a genesis
  with `--skip-genesis-sig-verification`, a source chain's own genesis
  replayed or `gnodev` for instance, holds no public key until its owner
  signs again, since the ante handler skips those signatures before both the
  replay branch and the line that stores the key.
- A forked chain already running computes different account state at its
  genesis on this binary, since its replay now stores keys the old binary
  left out. A node rebuilding such a chain from genesis takes that step on
  the old binary, as for any hardfork. Upgraded in place, such a chain never
  replays again, so an account the old binary replayed holds no public key
  until its owner signs again.
- A run may burn or mint its own run-issued denom without `MaxSend`, since
  neither is a transfer. Neither touches `ugnot` or another realm's denom.
