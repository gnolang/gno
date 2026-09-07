# GRC20 EOA writes use private, explicitly delegated tellers

## Context

`UserTeller` supports direct EOA calls to a token realm. Registry user helpers
extended this to other realms through a per-token trusted-host list, adding
host management to the token library solely to support registry forwarding.

EOA classification alone cannot authorize a forwarding realm: when a user calls
an arbitrary realm, that realm has a valid `cur` and an EOA previous caller.
An unrestricted public user teller would let it transfer or approve that user's
tokens without a prior allowance.

## Decision

Keep `TrustHost`, `UntrustHost`, `UserTellerTrusted` and their stored host list
removed.
Use one `PrivateLedger.UserTeller(hostPath)` API for local entry points and
delegated registry entry points. Each teller stores an immutable, nonempty
host path, normalized by stripping any sub-realm suffix. Every write checks
the live frame, direct EOA caller, and exact host match (including that host's
sub-realms). CallerTeller uses the same host field fixed to the token origin;
readonly and fixed-account tellers do not require a host guard.

Provide registry EOA entry points only for tokens whose realm explicitly
passes `PrivateLedger.UserTeller(registryPath)` to `RegisterWithUserCalls`.
Registration validates the live frame, token origin, exact canonical teller
type, user-only mode, token pointer identity and registry host match before
storing token and teller in a private
realm-declared registration. Public lookups and every read-only tree traversal
expose only the token; Render unwraps the private registration too.

Leaked tellers cannot write from a different host, even for a direct EOA caller.
Registry tellers remain private. No host list, revocation, or teller getter is
introduced. Transfer and Approve act as the signing EOA; TransferFrom uses
the signing EOA's allowance as spender, never the registry's allowance.

## Alternatives considered

- Keep trusted hosts: supports registry forwarding but requires a separate
  authorization policy and opt-in for every token.
- Separate UserTeller and unrestricted UserRelayTeller constructors: requires
  learning two authority models and relies on secrecy for relay confinement.
  A single host-bound UserTeller makes the rule uniform and limits leaks.

## Consequences

EOAs use the token realm's own crossing entry points or, for opted-in tokens,
the registry's UserTransfer/UserApprove/UserTransferFrom. Realms retain the
registry's non-crossing wrappers backed by `RealmTeller`, acting as themselves
and spending user funds through explicit allowances. The removed APIs were
introduced on this branch; the existing caller/home and EOA regression tests
remain applicable. Ordinary Register does not opt in; foo20 demonstrates the
new registration path. grc20factory retains its existing registration policy.

This is pre-deployment: there is no persisted registry migration or retrofit
for previously registered tokens. Duplicate keys remain rejected. Package
tests cover canonical rejection, stale frames and lookup privacy; integration
tests cover persisted MsgCall writes, allowances and rejected relay callers.
All UserTeller callers now specify the intended host; UserRelayTeller and
IsCanonicalUserRelay are replaced by UserTeller and IsCanonicalUserTeller.

Registry source and registration changes affect the integration harness's
genesis storage cost. If absolute balances in the storage-price regression
test change, preserve its 500,000-byte allocation, storage/deposit assertions,
and refund expectation when recalibrating them.
