# ADR: Reject reassigning a realm value (`cur`) to close the identity-laundering primitive

## Status

Proposed

## Context

A crossing function receives its realm identity in the first parameter, `cur realm`.
The runtime treats that value as an authority capability: `cur.IsCurrent()`, `cur.Address()`,
`cur.PkgPath()`, `cur.Previous()` and `cur.IsUserCall()` are the predicates realms and
stdlibs authenticate callers with, and `cross(rlm)` accepts a realm value only when it
`IsCurrent()`.

`IsCurrent()` is decided by HIV pointer identity against the topmost crossing frame's `Cur`
(`realmIsCurrentOnMachine`, `uverse.go`). A crossing function entered **without** `cross()`
inherits its frame's `Cur` verbatim from the caller-supplied first argument
(`doOpCall`, `op_call.go`) — the transparent tail-call semantics that let a realm split its
logic across local crossing helpers. Preprocess only allows such a no-cross call when the
argument is the bare name `cur` resolving to a containing crossing function's parameter
(`preprocess.go`); it never checked that the variable had not been reassigned.

That gap is a general **caller-identity laundering primitive**. A hostile realm writes:

```go
func Lure(cur realm, ...) {
	cur = cur.Previous()   // rebind cur to the tx origin, or to a realm that crossed in
	steal(cur, ...)        // no-cross entry: steal's frame Cur BECOMES the forged value
}
func steal(cur realm, ...) { /* cur.IsCurrent() is now true for the forged identity */ }
```

Once the forged value is a frame's `Cur`, it satisfies `IsCurrent()` and is accepted by
`cross(rlm)`, so the attacker can present the identity of whoever called it (the transaction
origin — any user who signs one call to the attacker realm — or any realm induced to cross
in) to any downstream realm or stdlib. Verified reach (each reproduced end to end against a
live/txtar gno.land node; see the regression tests):

- `r/nt/grc20reg/v0` `Transfer`/`Approve`/`TransferFrom` — drains a lured holder's registered
  GRC20 tokens (dora `7cd298d9`), and the same drain is reachable directly through the public
  `grc20.Token.RealmTeller` returned by `MustGet`, so it is not a grc20reg-only bug.
- `chain/banker` `NewBanker(BankerTypeRealmSend, cur)` — drains **native ugnot** from any
  lured account; every account holds the gas token, so the blast radius is the whole chain.
  `NewBanker`'s guard comment explicitly assumed `cur.Previous()` fails `IsCurrent()`, which
  is exactly what the laundering defeats.
- Any realm authorizing callers by the ordinary `cur.Previous().Address()` / `.PkgPath()` /
  `.IsUserCall()` route — the attacker re-crosses into it with the forged value and it reads
  the laundered identity as the immediate caller.

The primitive forges *caller identity only*: writes still execute in the victim realm's own
code, so it does not by itself bypass the storage-ownership/borrow model. But since nearly all
realm authorization rests on caller identity, impersonation is equivalent to acting with the
lured party's authority.

## Decision

Treat the crossing `cur` parameter as a non-reassignable capability, and enforce the frame
invariant that a no-cross crossing entry inherits only its caller's own cur.

1. **Preprocess, assignment (primary):** reject reassigning (`=` or any op-assign) a
   realm-typed `cur`. The parameter is the one realm value a no-cross crossing call can forward
   (that call requires the bare name `cur`), so rebinding it to `cur.Previous()` smuggles a
   non-caller realm value into the slot that call accepts; making it a compile-time error means
   the attacker realms fail to deploy. `cur` is reserved as a *parameter* name — the preprocessor
   lets only a crossing function's first realm parameter use it — but a local may still shadow it
   in an inner block (`cur := cur.Previous()`), so a realm-typed `cur` on the left of `=` is
   usually, not always, the parameter. Both are rejected: a shadowing local cannot be cur-called
   anyway (its defining block is not a crossing `FuncDecl`/`FuncLitExpr`, which the pre-existing
   provenance check rejects at the call), so refusing it costs nothing and spares the rule from
   having to work out which `cur` it is looking at. Storing a realm value into some *other*
   variable is untouched — not a vector, and it keeps its own persist-time guard. Implemented in
   the `AssignStmt` handler in `preprocess.go`.

2. **Preprocess, address-of and range:** reject taking the address of a realm-typed `cur`, and
   reject assigning one in a `range` clause. (1) matches only
   a `*NameExpr` on the left of an assignment, so it does not see `p := &cur; *p = cur.Previous()`
   — a `StarExpr` write to the same parameter slot, which preprocesses, **deploys**, and reaches
   the callee's frame `Cur`, stopped only by (3) at call time rather than at deploy time. Chasing
   assignment shapes is the losing end of this problem: a slot whose address cannot be taken
   cannot be written through any of them, and refusing the address also covers `&cur` escaping
   into a helper that assigns through it, which no LHS rule would see. Implemented in the
   `RefExpr` handler in `preprocess.go`.

   `for _, cur = range rs` is a **third** write shape into the same slot, found by review *after*
   the address-of rejection was written: a `RangeStmt` is not an `AssignStmt`, so (1) never sees
   it, and it takes no address, so the `RefExpr` rule does not apply either. It likewise
   preprocessed, deployed and was caught only by (3) at call time. Rejected in the `RangeStmt`
   handler. The `DEFINE` form (`for _, cur := range`) binds a new name and stays legal, because a
   name whose declaring block is not a crossing `FuncDecl`/`FuncLitExpr` cannot be cur-called.

   That two separate write shapes were missed by an enumeration written specifically to catch
   them is the standing argument for (3). A binding-based rule — mark the crossing parameter's
   static block slot at `FuncTypeExpr` time and refuse any write to a marked slot, whatever the
   statement shape — would subsume all three and is the right long-term shape; it is not taken
   here because it is a larger change than this fix, and (3) covers the residue meanwhile.

3. **The `cur` naming rules decide on the resolved type.** They used to run at `TRANS_ENTER` of
   `FuncTypeExpr` against the *spelling* of the type expression (`p.Type.(*NameExpr).Name ==
   "realm"`), while `IsCrossing()` compares the resolved type to `gRealmType`. A type alias fits
   exactly in that gap: `type R = realm` leaves a function crossing while neither naming rule
   fires, so it could take a **second** realm parameter — arbitrary, caller-supplied, and literally
   named `cur`. That parameter passes the cur-call provenance check, because it really is declared
   by a crossing function, and can be forwarded by bare name into a callee's identity. No
   assignment and no address taken anywhere, so none of the write-shape rules above could see it.
   The rules now run at `TRANS_LEAVE`, where the type is resolved, and ask the same question of the
   same value that `IsCrossing()` asks — which is what stops the two answers drifting apart again.
   Aliases stay legal; they are simply held to the same rules as the spelling.

4. **The rebuild recognises its placeholder by identity, not by description.**
   A crossing entry whose inherited `cur` is the preprocessor's placeholder gets a real
   origin-rooted one built in its place. That test used to be "prev is origin-shaped **and** its
   address is empty" — a description of the placeholder rather than the placeholder itself.
   `buildOriginRealm` mints a **fresh** realm whose address is whatever `OriginCaller` holds, and
   `OriginCaller` is deliberately left empty wherever there is no signer: every query
   (`withQueryEvalMachine`) and the test harness. The materialized value therefore answered the
   same description, matched again at the next hop, and the rebuild cascaded down the whole chain
   — every callee receiving a fresh identity of its own instead of its caller's, with the check in
   (5) skipped at each level, because a rebuilt cur is exempt from it. Fail-open: the minted
   identity's `Previous()` reads as a direct user call, which is *more* authority than the caller
   had, not less.

   The placeholder is a single package-level value created once (`gOriginRealmTV`), so pointer
   identity asks exactly the question that was meant — "is this THE placeholder?" — and a freshly
   allocated realm can never be mistaken for it, however empty its address. Gating the rebuild on
   "no crossing caller" was tried first and is wrong: a crossing function literal in a test
   legitimately carries a placeholder while it *does* have a live crossing caller, and that
   guard broke `r/sys/users`, `r/gnops/valopers` and `r/morgan/chess`.

5. **Runtime backstop:** when a crossing function is entered without `cross()` and there is a
   live crossing caller, assert that the inherited `Cur` is that caller's **own** current cur, by
   HIV identity (the same comparison `IsCurrent()` makes). The only realm value a crossing
   function can forward by the name `cur` is its own cur (preprocess rejects any other first arg
   to a no-cross crossing call), so threading `cur` unchanged — and delegating it to another
   realm's executor/callback — both preserve HIV identity, while `cur = cur.Previous()` does not.
   A mismatch means a laundered identity. The check is skipped when there is no crossing caller
   (a bootstrap entry: `main`/`init`/origin, a top-level MsgRun/query, or the test harness seeding
   `cur` through non-crossing frames) and when `cur` was rebuilt from the preprocess-origin
   placeholder. Implemented in `doOpCall` in `op_call.go`.

   **This must not be deleted as dead code.** An earlier revision of this ADR called it
   "unreachable from source once (1) is in place", and that was false in the same commit: the
   `&cur` route above matched no assignment rule, deployed cleanly, and was caught here and
   nowhere else.

   **And it has to actually run, which for methods it did not.** `doOpCall` asked
   `ft.IsCrossing()`, where `ft` is the *unbound* type: for `func (s *S) M(cur realm)` its
   `Params[0]` is the receiver, not `cur`, and `IsCrossing()` only ever looks at `Params[0]`. So
   the answer was false for every crossing method and the whole block was skipped — no frame
   identity installed, no check run, and the `curIdx = 1` receiver adjustment directly above it
   unreachable. It now asks `fv.IsCrossing()`, which is set at preprocess from the **declared**
   type (receiver excluded) and is the same predicate `doOpPrecall` already uses a few lines up.
   A one-word change rather than a second way of asking the same question.

   **And it has to look in the right slot, which for one spelling it did not.** A method's
   receiver occupies block slot 0 however the call is written, so `cur` sits at slot 1 for every
   method — but the index was decided from `fr.Receiver`, which is set only for a *bound* call. A
   method **expression** (`f := (*S).m; f(s, cur)`) passes the receiver as an ordinary argument, so
   `fr.Receiver` is undefined while the block still starts with the receiver: the index stayed 0
   and the check read the receiver as if it were the cur. A `*S` matches no caller identity, so a
   legitimate call passing its genuine `cur` was rejected. This was latent and harmless while
   methods were skipped entirely; making them enter the check turned it into a hard panic on valid
   source, which is worse than the hole being closed. It keys off `fv.IsMethod` now — a property of
   the declaration, true for both spellings.

   Worth stating as a general point, because it is the one that cost the most here: **enabling a
   dormant code path can expose assumptions that were never exercised while it slept.** A fix that
   is correct in isolation is not finished until the paths it makes reachable have been probed, and
   the tests written alongside it cannot do that — they share its assumptions.

   **And it has to run on every dispatch path, which meant not writing it twice.**
   `doOpReturnCallDefers` duplicates `doOpCall`'s block setup — its own comment says so — and the
   copy omitted the identity install and the check entirely, so `defer target(cur)` ran unchecked
   while the identical undeferred call was refused. The fix is not a second copy of the check but
   one `installInheritedCur` called from both, because the gap existed *because* the logic was
   duplicated; a second copy would have been the same bug waiting to recur. A deferred crossing
   call takes the identity of the frame that deferred it — the same frame its arguments were
   evaluated in, since `defer` evaluates them at the statement. (2) closes that particular route, but "unreachable from source" is a claim about
   having enumerated every syntactic route to writing the slot, and that claim has already been
   wrong once. This check does not depend on the enumeration being complete, which is the only
   reason it is worth keeping.

   A pkgpath-equality invariant (inherited cur's pkgpath must equal the callee's) was tried first
   and rejected: legitimate crossing code runs with an inherited cur belonging to a *different*
   realm — DAO/gov executors and cross-realm callbacks are invoked with the caller's cur, and
   `p/` test entry points are seeded with an origin cur (pkgpath `""`). HIV identity is the
   invariant that actually separates delegation (allowed) from forgery (rejected).

## Alternatives considered

- **Delete `grc20reg`'s write wrappers (dora's proposed fix).** Necessary-insufficient: the
  same drain is reachable through the public `grc20.Token.RealmTeller`, and the native-banker
  and cross-realm-relay variants are untouched. The flaw is at the VM identity layer, not in any
  one realm, so it must be fixed there.

- **Harden each call site (make wrappers crossing, re-check identity).** `cross(rlm)` validates
  with the same launderable `IsCurrent()` predicate, so hub-side "act as the realm that called
  me" cannot be made safe. Fixing the primitive removes the need.

- **Runtime backstop only (no preprocess rejection).** Would leave `cur = cur.Previous()`
  compiling and only fail at execution, a worse developer experience and a larger trusted
  surface. Rejecting at preprocess is the tighter, earlier gate; the runtime check is kept as
  defense in depth.

- **Reject reassigning *any* realm-typed variable, not just `cur`.** Tried first and rejected:
  it breaks legitimate code that stores a realm value into an ordinary variable to exercise the
  persist-time guard (`stashed = cur`, `stash = cur.Sub("vault")` in the `zrealm_*_persist`
  filetests). Only the `cur` parameter is a laundering vector — it is the sole name a no-cross
  crossing call can forward — so the rule is scoped to it; other realm-typed assignments keep
  their existing persist-time refusal.

## Known gaps, not closed here

None outstanding. The four found by review after the first revision — the type
alias defeating the `cur` naming rules, crossing methods never entering the
backstop, deferred cur-calls bypassing it, and the rebuild cascading through
QueryEval — are each closed above, with a test that fails without the fix.

What is worth carrying forward is not the list but its shape: **every one of
them was a path that skipped a check, not a check that was wrong.** Three were
dispatch paths (methods, defers, queries) and one was a declaration path
(aliases). The rules themselves held up each time. When extending this area,
the question to ask is "which paths reach this, and do they all go through the
same code?" — the defer gap existed purely because `doOpReturnCallDefers`
duplicated `doOpCall`'s setup, and it is now one shared helper for that reason.

## Every dispatch path is covered, by enumeration

The gaps in this area were paths that skipped a check, so the useful question is
not "is the check right?" but "does every path reach it?". That is answerable
rather than a matter of judgement: a crossing function can only be entered
through a pushed call frame, and `PushFrameCall` has exactly three callers —

| site | path | reaches the check |
|---|---|---|
| `op_call.go:19` | `doOpPrecall`, plain func | yes, via `doOpCall` |
| `op_call.go:38` | `doOpPrecall`, bound method | yes, via `doOpCall` |
| `op_call.go:741` | `doOpReturnCallDefers` | yes, directly |

Both terminal paths call `installInheritedCur`, which is why it is one shared
function rather than two copies. **If a fourth `PushFrameCall` caller is ever
added, it must call it too** — that is the whole maintenance burden of this
design, and grepping the constructor is how to check it.

## What it costs

Per crossing entry, roughly: one `topCrossingCur` walk (three branches per
frame, stopping at the nearest crossing frame — typically one or two), one
`realmMatchesCurHIV` (a handful of pointer derefs), and one
`curUsesPreprocessOrigin` (now a single pointer comparison). On the order of
60–90ns against a crossing call that costs 1–3µs in a tree-walking interpreter,
so a few percent of crossing calls and nothing at all for non-crossing ones.
These are reasoned from operation counts, not benchmarked.

Two notes for anyone tempted to optimise it:

- **`topCrossingCur` is O(distance to the nearest crossing frame), not O(stack
  depth).** It only degrades with many *non-crossing* frames between a call and
  its nearest crossing ancestor. Giving each frame a pointer to that ancestor,
  set at push, would make it O(1) — but that adds a field to `Frame` and upkeep
  on every call to remove a cost nothing has been shown to notice. Measure first.
- **None of this work is gas-metered.** Real CPU per crossing call varies with
  frame distance while the gas charged is constant. That predates this change;
  it is noted because the change widened where it applies.

## Consequences

- `cur = cur.Previous()` and any realm-value reassignment are now compile errors. This is a
  source-compatibility change, but no in-tree realm relied on it; realm reassignment only ever
  served the laundering shape.
- Legitimate crossing code is unaffected: threading `cur` unchanged into local no-cross helpers,
  `cross(cur)` calls, and the `(_ int, rlm realm)` non-crossing convention all continue to work
  (the non-crossing convention does not reassign `cur` and does not hit the inheritance path).
- The `grc20reg` write wrappers keep working for their intended use (a realm moving its **own**
  registered tokens with its genuine `cur`), so the fix closes the hole without removing
  legitimate functionality.

## Tests

- `gnovm/tests/files/zrealm_cur_reassign.gno` — preprocess rejects the reassignment.
- `gnovm/tests/files/zrealm_cur_addr.gno` — preprocess rejects `&cur`, the `StarExpr` route that
  the assignment rule alone let through to deploy.
- `gnovm/tests/files/zrealm_cur_range.gno` — preprocess rejects `for _, cur = range rs`, the third
  write shape, which neither of the other two rules matched.
- `gnovm/tests/files/zrealm_cur_shadow.gno` — pins the deliberate over-rejection of a shadowing
  local named `cur`, so the next reader does not "fix" it back into a gap.
- `gnovm/tests/files/zrealm_cur_shadow_{call,if,range}.gno` — a shadowing `cur` (block, if-init,
  range-define) cannot be passed to a crossing helper. Pins the provenance check the shadow test
  relies on; all three fail if its `default:` branch stops panicking.
- `gnovm/tests/files/zrealm_cur_backstop.gno` — the runtime backstop (3) actually fires. This is
  the test that makes it deletable-**detectable**: without it the whole suite passes with the
  check removed, so a dead-code pass takes it out and CI says nothing. It reaches the backstop
  through a closure that captures `cur` in one crossing frame and is called from another — a
  stale capture, which is an ordinary mistake rather than an attack, and is the shape a developer
  will actually hit. An earlier version used the type-alias gap; that gap is closed now.
- `gnovm/tests/files/zrealm_cur_addr_arg.gno` — the helper form, `set(&cur, cur.Previous())`,
  where `&cur` never appears on the left of anything. Kept separate so that narrowing the
  `RefExpr` rule to assignment context fails a test rather than silently reopening the route.
- `gno.land/pkg/sdk/vm/keeper_cur_query_test.go` — a query applies the same check as a
  transaction. Fails before the identity fix ("An error is expected but got nil"), because the
  cascading rebuild exempted every hop.
- `gnovm/tests/files/zrealm_cur_alias.gno` — the naming rules see through a type alias.
- `gnovm/tests/files/zrealm_cur_alias_second.gno` — and at a position other than the first.
- `gnovm/tests/files/zrealm_cur_alias_legal.gno` — positive control: an alias for `realm` is not
  banned, only held to the same rules.
- `gnovm/tests/files/zrealm_cur_defer.gno` — a deferred crossing call is checked. Fails with the
  one line removed from the defer path.
- `gnovm/tests/files/zrealm_cur_defer_legal.gno` — positive control: deferred crossing calls in
  all four spellings (plain, bound method, method value, method expression) from the frame whose
  identity they carry. It passes with and without the fix by design; what it catches is a fix that
  refused deferred crossing calls outright instead of inheriting for them.
- `gnovm/tests/files/zrealm_cur_method_expr.gno` — a bound call, a method expression and a method
  value, all passing their genuine `cur`. Fails if the slot index is keyed off `fr.Receiver`.
- `gnovm/tests/files/zrealm_cur_method_backstop.gno` — a crossing **method** is checked like a
  crossing function. Fails before the `fv.IsCrossing()` change (the forged value sails through
  and the method prints an empty pkgpath), passes after.
- `gnovm/tests/files/zrealm_cur_method_legal.gno` — positive control for the same: an honest
  crossing method still sees itself as the current realm, cur-called and through `cross()`.
- `gnovm/tests/files/zrealm_cur_legal.gno` — positive control: threading `cur`, `cross(cur)`,
  the `DEFINE` form `for _, cur := range`, and a non-realm variable named `cur` including `&cur`.
- `gno.land/pkg/integration/testdata/cur_laundering_rejected.txtar` — the grc20reg-wrapper,
  native-banker, cross-realm-relay and `cur = forged` attack shapes all fail to deploy, and a
  positive-control realm that threads `cur` correctly still works.
- The runtime backstop was verified by temporarily disabling the preprocess rejection and
  confirming `doOpCall` rejects the laundered native-banker drain at execution
  (`... was entered without cross(), so it takes its caller's identity — but the value passed is not the caller's own cur ...`), while
  the full `examples/` gno-test suite (DAO executors, cross-realm callbacks, `p/` test entry
  points) and the gnovm filetest suite continue to pass.
