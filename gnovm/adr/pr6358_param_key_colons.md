# Allow colons in realm param names

## Context

`chain/params.Set*` panicked on a key containing `:`. The IBC core realm
(gno-realms) stores packet commitments, receipts and acknowledgements as realm
params, under `clientID || type byte || uint64_be(sequence)`, because params
are provable from the store root by counterparty light clients. The
big-endian sequence contains 0x3a (`:`) for sequence 58, 314, 570, ..., 14906
and many others, so the 58th packet on any client could never be committed:
SendPacket, RecvPacket and acknowledgements on it panic, forever.

The ban did not protect namespacing: a realm's keys are always stored under
`vm:<realm>:`, so a colon in the name cannot reach another realm or module
(writing `bank:p:lockTransfer` lands at `vm:<realm>:bank:p:lockTransfer`). It
protected one reader, `realmFromKey` in the params deposit accounting, which
split on the last colon, while `ValidateRealmParamKey` split on the first.

## Decision

- `pkey` accepts colons in the name.
- `realmFromKey` takes the realm path to end at the first colon after `vm:`;
  package paths cannot contain one. Every reader now agrees.
- `ValidateRealmParamKey` (genesis) accepts the same names a realm can write.

## Alternatives considered

- **Escape the key in gno-realms.** Changes the stored key, hence the path
  counterparties prove, and breaks every deployed counterparty client.
- **Keep the ban and document it.** Leaves IBC unusable past 57 packets.

## Consequences

- `params.txtar` now asserts that a realm writing `bank:p:lockTransfer` lands
  in its own namespace and leaves the bank param unchanged.
- A consensus change: chains running the old rule reject transactions this
  one accepts.
