# ADR: count a func literal as a call when it crossed a realm boundary

## Context

`runtime.AssertOriginCall()` is the VM's answer to "was this realm entered
straight from the message, with nobody in between". Realms that verify a
payment lean on it, because the send envelope they read is transaction-wide:
`unsafe.OriginSend()` returns the same coins to every realm in the call chain,
while the coins themselves were credited to exactly one address, the realm the
message entered. Without the origin check, reading the envelope says only what
the transaction carried, never who was paid.

The check is a frame count. `isOriginCall` (`gnovm/stdlibs/chain/runtime/native.go`)
requires the first frame to be a `MsgCall` root and then accepts at most two
call frames, which is entry realm plus `AssertOriginCall` itself.

`NumCallFrames` did the counting, and since #5407 it skipped two kinds of frame:

```go
if fr.Func != nil && !fr.Func.IsClosure {
    count++
}
```

Control-flow basic frames (`for`, `range`, `switch`, where `Func` is nil) are
skipped, which is right: they are not calls. Func literals were skipped for a
narrower reason, #3918: a realm that factors its body into a closure was
tripping the check over an implementation detail of its own, and that is not
what "who called me" is asking about.

### The gap

Skipping every func literal skips the ones that are entry points. Gno exports a
`var` of func type as a callable symbol, so

```go
var Relay = func(cur realm) {
    _ = unsafe.OriginSend()
    target.Deposit(cross(cur))
}
```

is reachable by `MsgCall`, and its frame is a closure frame. A message entering
`Relay` puts three frames on the stack:

| # | frame | closure | crossed | counted before |
|---|---|---|---|---|
| 0 | `Relay`, in the intermediary | yes | yes | no |
| 1 | `Deposit`, in the target | no | yes | yes |
| 2 | `AssertOriginCall` | no | no | yes |

Two counted frames, so the target's `AssertOriginCall` passed while an
intermediary realm sat between it and the message. An otherwise identical
*named* `func Relay` was rejected, which is what pins the frame counting as the
cause rather than anything about the call itself.

The consequence is not that the check is merely weaker. It inverts: the coins
of a message that entered `Relay` are credited to `Relay`, so a target that
accepts the origin check as proof of payment credits the sender for coins it
never received. `r/gnoland/wugnot` is the one shipped realm in this shape: its
`Deposit` pairs `AssertOriginCall` with `unsafe.OriginSend()` and mints one
wugnot per ugnot in the envelope, so an intermediary kept the coins and took
the wrapped supply.

`r/sys/namereg` and `r/demo/disperse` read the same envelope and check
`cur.Previous().IsUserCall()`, which is the pairing `chain/runtime/unsafe`
documents on `OriginSend` itself. That defeats the func-*literal* relay, which
has to put its own frame on the stack. ⚠️ **It does not defeat a func value**,
and this decision does not protect them either. See "What this does not reach".

## Decision

A func literal is transparent inside its own realm and opaque across one.

`Machine.NumCallBoundaryFrames` counts every named call, plus every func
literal that ran with **someone else's storage authority**. Walking outwards
from the innermost frame, it compares the realm the frame's body ran in against
the realm of whoever called it (`LastRealm`), with `ownsItsStorage` excluding
pure packages. A nil `LastRealm` means the call came from the message rather
than from Gno code, so that frame is the one the message entered and counts like
any other entry point.

`isOriginCall` pairs that with a second, different question: whether the message
named *this* code at all, `Frames[0].Func.PkgPath == OriginSendRecipientPath`.
Frame counting cannot answer that one, for the reason set out below.

Control-flow basic frames still never count, and a closure called within its
own realm still does not, so #3918 stays fixed and `std13`-`std16` are
unchanged.

### ⚠️ The crossing flags alone are not the test, and it is not close

An earlier revision counted any closure frame with either flag set. That is
wrong, and it breaks a realm calling **its own** crossing func literal, which
keeps `AssertOriginCall` on `master`:

| realm's own closure, invoked as | `WithCross` | `DidCrossing` | left the realm |
|---|---|---|---|
| `f(cur)`, cur forwarded | false | **true** | no |
| `f(cross(cur))` | **true** | true | no |
| `f()`, non-crossing (the #3918 shape) | false | false | no |

The first row is the trap: `installInheritedCur` gives a crossing function
entered *without* `cross()` the identity its caller was running as, and sets
`DidCrossing` doing it, even though nothing crossed. So neither flag on its own
and no combination of them separates "crossed a realm" from "is a crossing
function".

Gating on the flags the other way round is wrong too, and in the dangerous
direction: a plain non-crossing call into **another realm's** exported closure
sets neither flag, yet the closure runs with that realm's storage authority.
Skipping it lets an intermediary keep `AssertOriginCall` alive in a realm the
message never entered.

The declaring package is not the test either. A func literal in a `/p/` package
has no storage of its own and borrows its caller's, so it is the caller's code
however different the two paths look, and comparing `PkgPath` counts it as a
boundary for nothing.

What is left is the **storage realm on each side of the call**: the realm the
frame's body ran in, against the realm of whoever called it. That is what
`NumCallBoundaryFrames` compares, walking outwards from the innermost frame,
with `ownsItsStorage` excluding the `/p/` case because the VM names a pure
package as the current realm while its code runs even though its writes land in
the borrower.

Found in review, not by the tests, twice: every closure in `std13`-`std16` and
in cases 21 to 23 is non-crossing and same-realm, so none of them could reach
either half. Cases 24 to 28 cover it, and each fails against the rule it
replaced.

A wider sweep for the same class of false positive found none. A crossing func
literal cannot be declared in a `/p/` package at all, which the preprocessor
rejects outright (*"crossing function literal (realm first argument) declared in
non-realm package"*), and a package-level closure var, a non-crossing closure
returned by a `/p/` package, and a method calling its own crossing closure all
behave identically to `master`.

`isOriginCall` switches to the new counter, in the chain stdlib and in the
`gno test` one (`gnovm/tests/stdlibs/chain/runtime/testing_runtime.go`), which
have to agree or a realm's tests stop describing the chain.

### Why not change `NumCallFrames` in place

`NumCallFrames` has a second consumer: it is the `SizeNumCallFrames` input to
the native gas model, which bills `chain/runtime.getRealm` and
`chain/runtime/unsafe.getRealm` against stack depth. Redefining it would move
those gas costs in every call that has a crossing closure on the stack, which
is a consensus change in a place that has nothing to do with the question being
fixed. The two counts answer different questions, so they are two functions.

## ⚠️ What this does not reach

Both checks live inside `isOriginCall`, so **they only protect a realm that calls
`AssertOriginCall`**. A realm that authenticates with `cur.Previous().IsUserCall()`
alone never reaches them, and the alias shape defeats that guard on its own
terms rather than by evading it: when `var Register = namereg.Register` runs,
the aliasing realm genuinely has no frame on the stack, so `Previous()` *is* the
EOA and `IsUserCall()` is *correctly* true. Nothing is being spoofed. The realm
is simply asking a question whose honest answer does not mean what it needs.

Measured against `r/sys/namereg/v0` itself: an alias call registers the name and
returns `OK!`. Nothing is at risk there today because `registerPrice` is
`int64(0)`, but the comment on it says governance can raise it via
`ProposeNewRegisterPrice`, and the moment it does, a fee paid to the aliasing
realm buys a registration from namereg.

So the alias is a **message-dispatch** property, not an origin-call property:
the VM resolves the named symbol, credits the coins to the named path, and runs
whatever function the value holds. Closing it for every realm rather than for
opt-in callers means refusing that dispatch at the message boundary, which is a
larger change than this one and is filed separately. What is decided here is the
narrower thing: a realm that *does* call `AssertOriginCall` gets an answer that
is true of both the code and the coins.

## Consequences

A realm entered through an exported func value is now correctly seen as an
intermediary by anything it calls. This is a VM behaviour change: a call that
used to pass `AssertOriginCall` now panics with `invalid non-origin call`.
Nothing in `examples/` relied on it, and a caller that wants the old shape can
still reach the target directly.

`NumCallFrames` keeps its behaviour and its gas meaning, and its doc comment no
longer claims to be the origin-call count.

**No realm has to change**, and the claim that makes that true has to be stated
carefully, because an earlier revision of this ADR overstated it.

Frame counting alone proves only that one realm's code ran. It does **not**
prove the message named that realm: a message names a path, the VM credits its
coins to that path and then invokes whatever value the named symbol holds, so
an exported `var Dep = other.Deposit` runs *other's* own named function under
the aliasing realm's name. `Frames[0]` is then the target's own function, the
count is two, and `AssertOriginCall` returns while the coins sit in the alias.
No frame rule can see this, because only one realm's code is ever on the stack.

So `isOriginCall` checks both, and they are different questions:

| check | question it answers |
|---|---|
| `Frames[0].Func.PkgPath == OriginSendRecipientPath` | did the message name *this* code |
| `NumCallBoundaryFrames() <= 2` | did anything run between the message and here |

`OriginSendRecipientPath` is the path the message named, set for every
`MsgCall` whether or not it carries coins, and empty where there is no message
at all, which fails closed. Together they give the property
`r/gnoland/wugnot` was already relying on: `AssertOriginCall` returning means
the message entered *this* realm, so the send envelope it then reads was
credited *here*. Verified against the deployed `wugnot.gno` bytes, unmodified.

Coverage:

- `gno.land/pkg/integration/testdata/assertorigincall.txtar`, cases 21 to 28:
  an exported func value is still a legitimate entry point when nothing
  crosses (21), it is a call boundary when the message reaches another realm
  through it (22), a second same-realm closure between the entry and the
  crossing does not hide it (23), a realm calling its own crossing func literal
  keeps the origin call whether it forwards `cur` (24), crosses to it (25) or
  defers it (26, a separate frame-building path through
  `doOpReturnCallDefers`), a plain call into another realm's exported closure
  is a boundary even though it sets no crossing flag (27), and the same call
  into a `/p/` package's closure is not, because it borrows the caller's
  storage (28).
- `gno.land/pkg/integration/testdata/assertorigincall_alias.txtar`: the four
  ways to bind another package's function to an exported name (func value,
  func value with coins attached, interface method value, bound method), all
  refused, with the aliased functions still working when the message names the
  realm that declares them.
- `gno.land/pkg/integration/testdata/wugnot.txtar`: the end-to-end case, an
  intermediary with a func-value entry point refused against the shipped
  `r/gnoland/wugnot`, with supply unchanged after it.
