# GRC20 MsgCall entrypoints through private Teller delegation

## Context

PR #6123 restores structured MsgCall transfers and approvals, avoiding generated
MsgRun source. Its unrestricted public Token.UserTeller also let any directly
called realm spend or approve the user's tokens without an allowance. The review
requests explicit spending caps for ordinary apps.

External host policy was deliberately removed from the token package. Storing
one host string instead of an AVL allowlist would reintroduce that responsibility.
We need registry user entrypoints without putting external realm identities back
into Token or Teller.

## Decision

Move UserTeller from Token to PrivateLedger. The private constructor creates a
bearer capability: every write verifies IsCurrent, requires the immediate
Previous().IsUserCall(), and resolves the actor at call time. It stores no
external host path. Existing CallerTeller remains confined to the token origin;
RealmTeller and the allowance implementation retain their semantics.

A token opts into registry user calls by delegating a capability:

```go
grc20reg.RegisterWithUserCalls(cross(cur), Token, slug, ledger.UserTeller())
```

Ordinary Register keeps its signature and grants no user authority.
RegisterWithUserCalls accepts only the canonical non-nil user Teller for the
exact Token pointer, then applies the existing token-origin and duplicate-key
checks. Validation does not dispatch caller-supplied methods, so interface
embedding cannot bypass it.

The registry stores a private registration containing the public token and
private userTeller. Get and MustGet return only Token; GetRegistry uses the
existing rotree value transformation for all lookup and iterator operations.
No public accessor returns the capability. Crossing UserTransfer, UserApprove,
and UserTransferFrom use the private relay.

Users approve ordinary apps through the token's own MsgCall entrypoint or the
registry's UserApprove. Apps spend their allowance through RealmTeller or the
existing non-crossing grc20reg.TransferFrom wrapper. Registry user calls remain
crossing; realm-owned wrappers remain non-crossing to preserve the app as actor.

## Alternatives considered

- Token-level host allowlist or a single host string: rejected because external
  trust policy would return to the reusable token package.
- Host-bound Teller: keeps the list out of Token but still makes the token
  package responsible for external host identity.
- Token-only MsgCall entrypoints: avoids delegation but loses the registry's
  common user-write surface.
- Approving the registry as spender: supports capped transfers but cannot
  create a user's approval for a different app, so it does not replace UserApprove.
- Operation-specific callbacks or signed permits: introduce more implementation
  and validation machinery than reusing the existing Teller methods.

## Consequences

Public Token pointers no longer grant arbitrary user-relative authority. The
private ledger and delegated tellers must stay private. A leaked UserTeller IS
usable by another realm's direct EOA callers; frame checks do not bind it to its
intended holder. The filetest deliberately demonstrates this property and
per-call actor resolution. This design trades host confinement for capability
encapsulation, not for absence of trust.

The registry remains trusted for direct user calls and must not invoke untrusted
callbacks with a live user frame. No new host management or delegation-revocation
API is introduced. User spending allowances remain revocable with Approve(0),
which is distinct from revoking the registry's delegated capability. A malicious
app can spend its entire approved amount to any recipient; unused allowances
survive failed operations and retain the existing approval-ordering race.

Source compatibility changes are limited to the new PR surface: callers of
Token.UserTeller must use the private ledger or a delegated capability. Existing
Register users retain discovery and realm-owned wrappers; enabling User* requires
RegisterWithUserCalls. This is a pre-deployment change: no persisted registry
migration or genesis edit is included. MsgRun remains outside UserTeller's
policy and can still use RealmTeller under its ephemeral user address.

## Validation

Tests cover canonical capability validation (including nil, embedding, wrong
Token, and other Teller kinds), unauthorized and duplicate registration, public
views, direct MsgCall writes, intermediate callers, stale callbacks, and MsgRun
rejection. A compile-time regression verifies that public Token has neither
CallerTeller nor UserTeller constructors.

The registry integration test exercises approve 100, spend 30, leave 70, reject
71; spending the same allowance through the realm-owned registry wrapper;
revocation with zero; explicit renewal; and exhaustion. Ordinary realm approvals
and transfers cannot change the user's allowance or debit the user's balance.

Validation passed locally:

- `cd examples && go run ../gnovm/cmd/gno test ./...`
- `go test ./gno.land/pkg/integration -run TestTestdata -count=1`
- `git diff --check`

Only the storage-price regression's two absolute account balances changed,
each by 413,700 ugnot from HEAD for the changed genesis storage cost. Its
500,000-byte allocation, storage/deposit checks, and 50,031,600 ugnot refund
are unchanged.

The required `/simplify` command was attempted but the local Claude CLI was
not logged in. Manual simplification and authority-boundary review were
performed instead; no automated simplify result is claimed.
