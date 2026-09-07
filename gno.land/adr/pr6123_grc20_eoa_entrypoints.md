# Public GRC20 user tellers for direct EOA calls

## Context

The September 7 Gno Core Weekly decision for PR #6123 removes trusted-host
policy, UserTeller host arguments, and stored host paths. The specified behavior
allows a direct EOA to use a public UserTeller through any foreign realm, while
rejecting intermediate cross-calls and MsgRun.

## Decision

Expose `Token.UserTeller()` with no arguments. Each write checks `IsCurrent()`
before reading the previous frame, requires `Previous().IsUserCall()`, and
resolves its actor from `Previous().Address()` at the time of the write.
Transfer and Approve use that actor as owner; TransferFrom uses it as spender
and requires an allowance from the supplied owner.

Remove private relay registration, stored relay capabilities, and their
canonical-token/host validator. Ordinary `Register` makes the token available
to all three registry user helpers, which construct UserTeller directly.
Registry storage and read-only views contain token pointers again.

CallerTeller retains its separate token-origin restriction using a boolean
and the existing `Token.origRealm`; no teller stores a host path. Readonly,
RealmTeller, RealmSubTeller, and ImpersonateTeller keep their existing semantics.

## Alternatives considered

- Trusted-host lists or host-bound private tellers restrict which realms may
  act for an EOA, but require token-specific delegation and host policy.
- Private UserTeller with no host parameter still requires explicit delegation
  and does not support the requested public foreign-realm constructor.

## Consequences

A user who directly calls a foreign realm allows it to transfer or approve
that user's tokens through any public Token pointer, without prior allowance.
This is the accepted authority model, not protection against a malicious realm
that the user directly invokes. The IsUserCall check authenticates the actor;
it does not establish the user's intent for each token operation.

Bob calling the same realm cannot become Alice: actor resolution is per write,
including when a teller is retained across calls. An intermediate explicit
cross-call changes the previous frame to a realm and is rejected. MsgRun and
stale frames are rejected as well. Non-crossing code given a live `cur` shares
that frame's authority; it is not isolated by the direct-EOA check.

Package/filetests cover caller classification, nil tokens, stale frames, stored
teller reuse, and unchanged CallerTeller confinement. The registry txtar tests
exercise persisted EOA transfers and allowances, foreign-realm access, actor
separation, and rejected intermediate/MsgRun writes with unchanged state.

This is a pre-deployment change; no persisted registry migration is included.
Removing source and relay state changes the integration harness's genesis
storage cost. Storage-price test balance expectations are recalibrated without
changing its 500,000-byte allocation, storage/deposit checks, or refund amount.
