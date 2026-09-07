# GRC20 EOA writes stay in the token realm

## Context

`UserTeller` supports direct EOA calls to a token realm. Registry user helpers
extended this to other realms through a per-token trusted-host list, adding
host management to the token library solely to support registry forwarding.

EOA classification alone cannot authorize a forwarding realm: when a user calls
an arbitrary realm, that realm has a valid `cur` and an EOA previous caller.
An unrestricted public user teller would let it transfer or approve that user's
tokens without a prior allowance.

## Decision

Remove `TrustHost`, `UntrustHost`, `UserTellerTrusted`, their stored host list,
and the registry's `UserTransfer`, `UserApprove`, and `UserTransferFrom` helpers.
Keep `PrivateLedger.UserTeller` confined to the token's originating realm, with
both current-frame validation and the direct-EOA check.

## Alternatives considered

- Keep trusted hosts: supports registry forwarding but requires a separate
  authorization policy and opt-in for every token.
- Remove only the host guard: admits arbitrary realms acting for their EOA
  callers, so it is not an equivalent security boundary.

## Consequences

EOAs use the token realm's own crossing entry points. Realms retain the
registry's non-crossing wrappers backed by `RealmTeller`, acting as themselves
and spending user funds through explicit allowances. The removed APIs were
introduced on this branch; the existing caller/home and EOA regression tests
remain applicable.

Removing registry source also lowers the integration harness's genesis storage
cost. Update the two absolute account balances in the storage-price regression
test, preserving its 500,000-byte allocation, storage/deposit assertions, and
refund expectation.
