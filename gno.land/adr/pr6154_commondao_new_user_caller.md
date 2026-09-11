# ADR: commondao `New` keys creation on a user caller

## Context

`New` authenticated with two different identities at once:

```go
orig := unsafe.OriginCaller()      // invite lookup, consumption, creators set
...
caller := cur.Previous().Address() // founding council seat
dao := createDAO(..., parseInitialMembers(caller, members)...)
```

When a realm relayed the call the two diverged: the invite was redeemed from
the transaction origin (a human who held one) while the council seat went to
the relaying realm. The user paid their one invite and got nothing; the realm
got a DAO. Both reads are individually defensible — the split is the defect.

Origin-keying was a deliberate choice, recorded in
`pr6012_commondao_ownership_rescope.md` and reaffirmed in
`pr6012_commondao_council.md`. Its stated reason was not the council seat but
spam: with the `creators` set keyed on the *caller*, a realm that created one
DAO became a permanent, un-invited DAO factory, since anyone could call it and
the realm would already be in `creators`. Origin-keying dodged that because
the origin is always an EOA.

So collapsing onto `cur.Previous()` alone fixes the identity split but
reopens the factory vector. Something has to keep the caller from being a
realm.

## Decision

Key everything on `cur.Previous().Address()`, and require that caller to be a
user:

```go
assertCurrent(cur)
assertCallerIsUser(cur)
caller := cur.Previous().Address()
```

`assertCallerIsUser` tests `IsUser()`, which is `IsUserCall() || IsUserRun()`:

| shape | `PkgPath()` | predicate | |
|---|---|---|---|
| `maketx call` from an EOA | `""` | `IsUserCall` | accept |
| `maketx run` from an EOA | `gno.land/e/<addr>/run` | `IsUserRun` | accept |
| a published realm relaying | `gno.land/r/...` | `IsCode` | reject |

Both accepted shapes report the signer's own address — `VMKeeper.Run` sets
`pkgAddr := caller`, and `IsUserRun` additionally requires the address
embedded in the path to equal the frame's own address, so a deployed package
can never present as one. The invite is therefore always redeemed by, and the
council seat always granted to, the account that signed the transaction. The
factory vector stays closed because a realm can never enter `creators`.

The realm's other entry points are unchanged: they read
`cur.Previous().Address()` and accept realm callers, which is what lets a
realm sit on a council and govern.

## Alternatives considered

- **`IsUserCall()` only** (reject MsgRun too): rejected. It reads as the
  stricter, safer option, but `IsUserCall`-over-`IsUser` exists to protect the
  `OriginSend` envelope (`effective-gno.md`, *Verifying inbound Coin
  payments*) — an ephemeral realm can consume the envelope before calling on.
  `New` accepts no payment, so that reasoning does not transfer, and the cost
  is real: `maketx run` is how a user batches "create the DAO, then configure
  it" into one atomic transaction, and the frame's address is already exactly
  the right account.
- **Keep origin-keying, use the origin for the council seat too.** Consistent,
  and it would keep realm-relayed creation working. Rejected: it makes the
  seat go to whoever signed rather than whoever called, so a realm that offers
  a "create a DAO" button silently seats its users instead of itself, and the
  realm cannot create a DAO it governs at all. Caller-keying expresses
  intent directly.
- **Drop the guard and let realms redeem invites.** Rejected — that is the
  factory vector `pr6012_commondao_ownership_rescope.md` set out to close.

## Consequences

- **A realm can no longer call `New`.** The supported way to give a realm a
  DAO is for a user to seat it via the `members` argument and resign:

  ```
  New(cur, "MemberDAO", purpose, "", <realmAddr>)   # council {realm, user}
  Resign(cur, id)                                   # council {realm}
  ```

  `commondao_dao_member.txtar` was rewritten to bootstrap its policy realm
  this way; its `mpolicy.CreateMemberDAO` entry point is gone.

- **The guard bounds creation, not membership.** `members` is not restricted
  to EOAs, and `CreateSubDAOProposal` sets a sub-DAO's council outright
  without appending its creator or consuming an invite. Both can seat a realm
  — deliberately: those are council decisions taken by an existing DAO's
  governance, whereas `New` is permissionless given an invite. `New`'s
  docstring says so, so the guard is not mistaken for an
  "only EOAs hold council seats" invariant. It is not one.

- **`Invite` still accepts any valid address**, and an invite issued to a
  realm address can never be redeemed. This predates the change (the old
  `unsafe.OriginCaller()` was always an EOA, so such invites were already
  dead) and is left alone here.

- **Tests.** `commondao_new_caller_shapes.txtar` pins all three caller shapes
  on a real node, so the predicate cannot silently narrow to `IsUserCall` or
  widen to `!IsCode`. `z_7_b_filetest.gno` restores caller≠origin coverage for
  the proposal entry points: its council is a realm reached the supported way
  and the origin driving it is a non-member, so swapping any
  `cur.Previous().Address()` back to `unsafe.OriginCaller()` fails the suite.
  `z_7_a` previously carried that coverage as a side effect of creating its
  DAO from a code realm, which `New` no longer permits.

## Supersedes

- `pr6012_commondao_ownership_rescope.md` — "Invite / creation gating
  re-homed to a `creators` set, keyed on the transaction origin", and the
  "Keying the creators set on the caller" alternative. The vector that
  rejection names is now closed by `assertCallerIsUser` instead.
- `pr6012_commondao_council.md` — "The invite check deliberately uses
  `unsafe.OriginCaller()` … confirmed intended".
