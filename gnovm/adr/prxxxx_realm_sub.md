# `cur.Sub(subpath)` — Sub-Realm Identity Extension

## Status

Consolidated proposal, revised after a three-lens review (security,
implementation feasibility, design coherence) verified against master.
Supersedes the initial SUB.md draft and the CROSSUB.md draft; the
decisions inherited from those drafts are restated self-contained in
"Alternatives considered" so this document stands alone.

Decisions made in this revision (rationale in the named section,
table row, or sequencing step):

- `unsafe.PreviousRealm()` parity is **upheld** by extending the frame
  walker to read the presented-identity chain (§ unsafe interaction).
- `BankerTypeOriginSend` and `BankerTypeRealmIssue` are **forbidden**
  for sub-tokens (defensive checks at `NewBanker`).
- `Sub()` on ephemeral `/e/` run realms is **forbidden**.
- Subpath length cap is **256 bytes** (grounded in the existing pkgpath
  limit).
- gnobuiltins: `Sub` is added to the existing `gno0p9` realm interface
  **in place** — 0.9 is the latest language version, and the change is
  purely additive on a sealed interface, so no existing package
  becomes invalid. Activation rides a coordinated node upgrade
  (type-checking is consensus).
- The legacy `cross1` migration sentinel is **removed** as a precursor
  step (Sequencing step 0); its gno0p9-shim removal batches into the
  same upgrade as the `Sub` addition.

## Problem

In Gno v2, every actor that calls into a realm surfaces to the callee as
an address via `cur.Previous().Address()`. The address is either:

- an EOA (when the previous-realm is the chain root, pkgpath = ""), or
- a realm's package address (when the previous-realm is `/r/foo`).

Objects *inside* a realm — DAOs in `/r/nt/commondao/v0`, accounts in a
registry realm, etc. — have no native identity surface. When such an
object "acts" by having its host realm call downstream, the downstream
only sees the host realm's pkg address. Two DAOs in the same registry
are indistinguishable from the callee's perspective without additional
out-of-band parameters or ad-hoc vouching APIs.

This forces every consumer of host realms (boards, banker, etc.) to
either:

1. Accept extra `principal address` parameters and verify them via a
   per-host vouching list, or
2. Tie authority to the host realm's pkg address (so all sub-actors
   collapse into a single identity from any downstream's view).

Both are unsatisfying. The first scatters trust lists across every
downstream realm; the second loses identity entirely.

## Goal

Extend the `cur realm` capability so that a host realm can mint a
sub-realm representing one of its internal actors. Downstream callees
see the sub-actor as a normal `address` via the existing
`cur.Previous().Address()` idiom AND can also distinguish it from the
host's primary identity via a distinct `PkgPath()`. No new threading
parameter, no per-host vouching list, no separate principal type.

Constitutional fit (`docs/CONSTITUTION.md:1534`, "Each DAO may have an
associated crypto address"): sub-realms give each DAO a chain-derived
crypto address controlled by its host realm.

## Proposal

Add one method to the `realm` interface:

```go
// Sub returns a derived sub-realm token representing a sub-actor within
// the calling realm's namespace.
//
// The returned token has:
//   - Address() = chain.PackageAddress(cur.PkgPath() + ":" + subpath)
//   - PkgPath() = cur.PkgPath() + ":" + subpath   (distinct from any
//     primary pkgpath; the prefix before ":" encodes the host realm)
//   - Previous() = cur.Previous()  (the sub-token represents the host
//     realm acting under a sub-identity, not a new delegation step;
//     chain depth is unchanged from a regular cross)
//   - IsCurrent() = true iff the host realm is the topmost live
//     crossing frame's cur (compared by pointer identity against an
//     internal parent-cur reference stored on the sub-token at mint
//     time; not visible via Previous)
//
// Sub() panics unless ALL of the following hold:
//   1. The receiver's own HIV is pointer-identical to the topmost live
//      crossing frame's Cur. This is strictly stronger than
//      IsCurrent() — it does NOT go through the sub-token relaxation
//      (see "Sub-tokens cannot be Sub'd"). Consequence: sub-tokens can
//      never be Sub'd, because a sub-token's fresh HIV is never a
//      frame's Cur.
//   2. m.Realm is non-nil and m.Realm.Path == cur.PkgPath() (the
//      caller-pkgpath check; see that section). Sub() called on a
//      passed-around cur from a foreign realm panics. Sub() called at
//      the chain root or from stdlib context (m.Realm == nil) panics.
//   3. The receiver is not ephemeral. Sub() on an /e/ run-realm cur
//      panics (see "Ephemeral (/e/) hosts").
//
// subpath constraints:
//   - non-empty
//   - does not contain ":"
//   - does not contain NUL bytes
//   - length ≤ 256 bytes
//
// Sub-tokens cannot be persisted (refusePersistRealmHIV applies; see
// "Cross-time validation" for the isOriginRealmHIV carve-out that must
// exclude sub-tokens).
//
// Sub-tokens can be passed to cross(...) — as a bare identifier:
// cross's argument must be a name, not an expression, so mint to a
// local first (`sub := cur.Sub(x); f(cross(sub), ...)`). The extended
// realmIsCurrentOnMachine accepts them iff their parent's cur is the
// topmost live crossing frame's cur.
Sub(subpath string) realm
```

## Synthesized pkgpath

The sub-token's `PkgPath()` is the synthesized form:

```
sub_pkgpath = parent_pkgpath + ":" + subpath
```

`:` is reserved as a separator.

### Where the reservation is enforced

`Re_gnoUserPkgPath` (`gnovm/pkg/gnolang/mempackage.go:48-53`) and the
tm2-level `rePkgPathURL` (`tm2/pkg/std/memfile.go`) are both anchored
and colon-free, so `:` is already rejected in every package-ingestion
path today (keeper AddPackage, genesis, gnodev, and
`defaultStore.AddMemPackage` itself all funnel through
`ValidateMemPackageAny`). To make the reservation an explicit contract
rather than a regex side-effect, add the check where path validation
actually lives:

```go
// In gnovm/pkg/gnolang/mempackage.go, ValidateMemPackageAny (or the
// shared path-validation helper it uses):
if strings.Contains(mpkg.Path, ":") {
    return fmt.Errorf("invalid package path %q: ':' is reserved for sub-realm derivation", mpkg.Path)
}
```

(`MsgAddPackage.ValidateBasic` itself performs no path-format
validation; the mempackage layer is where every ingestion path
converges.)

### Derivation safety coupling (load-bearing)

The safety of sub-address derivation itself rests on real pkgpaths
being colon-free. `DerivePkgBech32Addr`
(`gnovm/pkg/gnolang/misc.go:201-211`) special-cases run paths: if the
input matches `IsGnoRunPath` (`<domain>/e/<addr>/run`), it returns the
embedded bech32 address **verbatim, skipping the hash**. A synthesized
`host:subpath` contains `:` and fails that anchored regex today, so
derivation falls through to the hash — but subpaths permit `/`, so if
the pkgpath rules were ever relaxed to admit `:`, a subpath like
`"/e/g1victim/run"` could alias run-path extraction and yield an
arbitrary account's address. Two defensive requirements:

1. `Sub()` and `DerivePkgSubAddr` assert that the *host* pkgpath
   contains no `:` (also forecloses nested synthesis through any
   unforeseen path), and
2. assert that the synthesized `host + ":" + subpath` is not a run
   path. In the Sub native this is a literal `!IsGnoRunPath` call; on
   the `.gno` side (`DerivePkgSubAddr`) no run-path predicate exists
   and none is needed — the host-colon-free assert subsumes it, since
   the run-path regex is anchored and colon-free.

These are one-line panics; they turn a silent regex coupling into an
explicit invariant.

### Stdlib helpers

```go
// In gnovm/stdlibs/chain/address.gno (alongside PackageAddress).
// Pure Gno; reuses the existing packageAddress native — no new native
// binding. Validates the same subpath rules as Sub() plus the
// host-colon-free assert; the run-path assert is native-only
// (subsumed here — see "Derivation safety coupling").
func DerivePkgSubAddr(pkgpath, subpath string) address {
    assertValidSubpath(pkgpath, subpath)
    return PackageAddress(pkgpath + ":" + subpath)
}

// SplitPkgSubPath splits a possibly-synthesized pkgpath into host and
// subpath. ok reports whether p is a synthesized sub-realm pkgpath.
func SplitPkgSubPath(p string) (host, subpath string, ok bool) {
    return strings.Cut(p, ":")
}
```

These let off-chain tooling and other realms compute and parse
sub-identities without consulting the host realm.

Examples:

- `cur.PkgPath() == "gno.land/r/nt/commondao/v0"`
- `cur.Sub("dao/42").PkgPath() == "gno.land/r/nt/commondao/v0:dao/42"`
- `cur.Sub("dao/42").Address() == chain.PackageAddress("gno.land/r/nt/commondao/v0:dao/42")`

### Matching sub-identities (the safe idiom)

A callee that does `cur.Previous().PkgPath() == "gno.land/r/nt/commondao/v0"`
matches the primary only — sub-tokens have a different pkgpath, so the
exact-match check is not silently broadened.

A callee that wants to accept the host **or** any of its subs must
anchor the separator:

```go
p := cur.Previous().PkgPath()
ok := p == host || strings.HasPrefix(p, host+":")
```

A bare `strings.HasPrefix(p, host)` is **wrong**: it also matches
sibling packages (`.../v0extra`), subdirectory packages (`.../v0/sub`),
and their subs — each a distinct authority. Every example in this
document uses the anchored form; docs and linters should too. (See
Migration for deployed code that already uses bare prefix checks.)

## IsCurrent semantics

`cur.Sub(subpath).IsCurrent()` returns **true** iff the host realm (the
parent cur at the moment of Sub()) is the topmost live crossing frame's
cur. The sub-token stores the parent's HIV in an internal field at
construction; `IsCurrent()` compares this stored reference against the
topmost crossing frame's Cur by pointer identity.

Note that the sub-token's `Previous()` does NOT expose this internal
parent reference — `Previous()` skips the host realm and returns
whatever was above the host in the chain (`cur.Previous()` at Sub()
time). The internal parent reference exists only to support IsCurrent
and the cross-time staleness check; it is not surfaced at the
language level.

This relaxation is **safe** because the synthesized pkgpath prevents
silent broadening of pkgpath-keyed authentication checks. Specifically:

- Pattern `if !rlm.IsCurrent() { panic }` (anti-spoof) — sub-tokens are
  not stolen or stale; they're held by code legitimately operating in
  the live parent's frame. Accepting them is correct.
- Pattern `if !rlm.IsCurrent() || rlm.Address() != owner { panic }`
  (owner check) — sub-tokens with the right address match (owner was
  intentionally set to the sub-address at banker construction); others
  fail the address check. No regression.
- Pattern `if !rlm.IsCurrent() || rlm.PkgPath() != expectedPkg { panic }`
  (pkgpath check) — sub-tokens have a different (synthesized) pkgpath
  from the primary; they fail this check. No silent broadening.

Distinct pkgpaths are the key insight that makes the relaxation safe:
pkgpath-keyed checks cannot silently accept sub-identities.

The existing invariant "`rlm.IsCurrent()` ⇔ `cross(rlm)` accepts"
(documented at `uverse.go:404-407`) is preserved: both route through
the same extended `realmIsCurrentOnMachine`. The proposal deliberately
does **not** extend that symmetry to `Sub()`, whose entry guard is
strictly stronger (next sections).

## Sub-token classification methods

Sub-tokens implement the full `realm` interface. The classification
methods are specified for v1 as follows:

| Method | Sub-token result | Rationale |
|---|---|---|
| `IsCode()` | true | A sub-identity is code-controlled by its host. |
| `IsUser()` | false | Never a user; keeps the `IsUser` payment-guard bug class from extending to sub-identities. |
| `IsUserCall()` | false | A sub-token cross is realm-mediated, never a direct user call. |
| `IsUserRun()` | false | Same reasoning. |
| `IsEphemeral()` | false | Ephemeral hosts cannot mint subs (Sub() panics on `/e/` curs), so every sub-token's host is a persistent realm. |
| `Subpath()` | the subpath (`""` for primaries) | The canonical "am I a sub, of what subpath" accessor. Derived from `PkgPath()` (substring after the first `:`), so it agrees with `chain.SplitPkgSubPath` for any realm value. Consumers should prefer it over string-parsing `PkgPath()`; the banker `RealmIssue`/`OriginSend` ban is the first consumer. |
| `String()` | `realm{<synthesized>:<addr>}` | The existing format already uses `:` internally; sub-tokens render with an extra colon. Cosmetic; tooling parses via `SplitPkgSubPath`/`Subpath()`, not `String()`. |

All user-ness predicates returning false is load-bearing: downstream
guards like `Previous().IsUserCall()` (payment envelopes) must not
treat a programmatic sub-identity as a user.

`Subpath()` (not a `Host() realm` accessor) is deliberately the only
new introspection: returning a host *realm value* would let any callee
holding `cur.Previous()` reconstruct a host-shaped realm and construct
a `RealmSend` banker over the host's primary treasury. The host
pkgpath as a *string* is available via `SplitPkgSubPath`.

## Chain visibility and forensics

Because `cur.Sub(x).Previous() = cur.Previous()` (the host realm is NOT
inserted as a separate step in the Previous chain), downstream callees
walking the chain see exactly the same depth they'd see for a non-sub
crossing. A callee doing `cur.Previous().Previous()` to reach the
chain root sees the chain root regardless of whether a sub-token was
used at the immediate previous level.

The host realm's identity is still recoverable forensically: the
synthesized pkgpath encodes it as the prefix before the `:` separator.

```go
host, subpath, ok := chain.SplitPkgSubPath(cur.Previous().PkgPath())
// host == "gno.land/r/nt/commondao/v0"
// subpath == "dao/42"
// ok == true (this is a sub-token, not a primary)
```

Tooling that wants to identify which host minted a sub-identity does
this parse rather than a chain walk. The information is present; it's
just at the pkgpath layer rather than as a separate Previous step.

## `unsafe.CurrentRealm()` / `unsafe.PreviousRealm()` interaction

The VM tracks "which realm" through two channels:

1. **The presented-identity chain** — realm *values*. At each crossing,
   `installCrossingCur` (`op_call.go:85-102`) stores the realm value
   passed to `cross(...)` as the new cur's `prev` field. This is what
   `cur.Previous()` reads, and it is where a sub-token lives.
2. **The crossing-context chain** — `*Realm` objects. `Frame.LastRealm`
   snapshots `m.Realm` at frame push (`machine.go:2252`);
   `execctx.GetRealm` (`gnovm/stdlibs/internal/execctx/realm.go`)
   counts `WithCross` frames and samples `LastRealm` at frames adjacent
   to crossing frames, reconstructing which realm *contexts* were
   explicitly crossed into. Borrowed (storage-switched) realms never
   surface in it. This is what `unsafe.*` reads today.

For every ordinary cross the two chains coincide: `cross(cur)` presents
an identity equal to the crossed-from context. A sub-token cross is the
first construct that splits them — the context crossed from is the
host; the identity presented is the sub. Without changes,
`unsafe.PreviousRealm()` inside the callee would return the host's
primary path and address (re-derived from `Frame.LastRealm.Path`)
while `cur.Previous()` returns the sub-identity, violating the v2 §5.4
agreement contract (§5.4 is titled for the pre-split `runtime.*`
names; `unsafe.*` is the current surface).

**Decision: uphold parity.** Extend `execctx.GetRealm` to read the
presented-identity chain: locate the innermost crossing frame, take its
`Cur` (a `.grealm` value already on the frame — crossing functions
entered *without* cross inherit the same Cur, `op_call.go:324-371`, so
the walk sees them identically), and walk `.prev` while
`height < crosses`, returning that value's `Address()`/`PkgPath()`.
The fallback boundary stays at `height == crosses` exactly as today
(EOA/origin fallbacks unchanged), and stacks with no Cur anywhere
(plain `init()`, StageAdd non-crossing, pure `/p/` entries) keep the
full legacy fallback. Only heights 0 and 1 are exposed
(`unsafe.CurrentRealm` → 0, `unsafe.PreviousRealm` → 1;
`chain/runtime` exposes no stack walker). The identity chain is already materialized on the
frames — no new threading is required. The test-stdlib mirror
(`gnovm/tests/stdlibs/chain/runtime/testing_runtime.go`, marked "keep
in sync") gets the same change and must reconcile the prev-walk with
its frame-index-keyed `RealmOverride` handling.

Inside a callee invoked via `sub := cur.Sub("dao/42");
target.Foo(cross(sub), ...)` (two-step: cross takes a bare
identifier):

```go
cur.Previous().Address()         // = sub-address (DAO)
cur.Previous().PkgPath()         // = "<host>:dao/42" (synthesized)
unsafe.PreviousRealm().Address() // = same sub-address
unsafe.PreviousRealm().PkgPath() // = same synthesized form
```

The rejected alternative — leave the walker reading the context chain
and document the divergence — is recorded in "Alternatives considered".

### Asymmetry note: `unsafe.CurrentRealm()` vs sub-token holder's view

Inside a frame whose code is mid-execution and holds a sub-token (not
yet crossed with it):

- `unsafe.CurrentRealm().PkgPath()` returns the host realm's pkgpath
  (the topmost crossing frame's Cur is still the host's primary cur —
  minting a sub-token shifts nothing until it is used).
- `subToken.PkgPath()` returns the synthesized sub-pkgpath.

This is intentional. The sub-token is a capability handle the host
holds; until it's used via `cross(subToken)` (or handed to a
token-style API), downstream observers see the host's primary identity.
Per v2 guidance, code uses the `cur` API for agency decisions;
`unsafe.*` remains debug-only.

## Caller-pkgpath check

`Sub()` requires `m.Realm != nil && m.Realm.Path == cur.PkgPath()`.
This forbids:

```go
// In /r/attacker's frame, with rlm passed from /r/victim:
rlm.Sub("anything")   // PANIC: m.Realm=/r/attacker, rlm.PkgPath=/r/victim
```

Without this check, a foreign realm receiving cur could mint sub-tokens
in the caller's namespace and use them to drain treasuries or
impersonate the caller in downstream realms. The check naturally permits
the legitimate case: `/p/`-methods on victim-stamped receivers, called
by victim with victim's own cur, where m.Realm is borrowed to /r/victim
via borrow rule #2 and cur.PkgPath() is also /r/victim.

Trace through scenarios:

| Caller | Function | m.Realm | rlm.PkgPath | Sub() allowed |
|---|---|---|---|---|
| /r/X | own /r/X code | /r/X | /r/X | ✓ |
| /r/X | /p/ method on /r/X-stamped receiver | /r/X (borrow #2) | /r/X | ✓ |
| /r/X | /p/ method on /r/Y-stamped receiver | /r/Y (borrow #2) | /r/X | ✗ |
| /r/Y | /r/Y code with passed /r/X cur | /r/Y | /r/X | ✗ |
| chain root | /p/ top-level func | nil | anything | ✗ (nil guard) |

Because `m.Realm.Path` is always a real, colon-free package path, the
check also structurally rejects any receiver with a synthesized
pkgpath — a second line of defense against nested synthesis, on top of
the strict entry guard below.

## Sub-tokens cannot be Sub'd

**Important distinction**: the entry guard inside `Sub()`'s native is
**not** a call to the receiver's `IsCurrent()` method. It is a *strict
pointer-identity check* directly against the topmost crossing frame's
Cur HIV. Pseudocode (all guards shown):

```go
func nativeSub(m *Machine, recv TypedValue, subpath string) realm {
    host := realmPkgPath(recv)
    assertValidSubpath(host, subpath)              // subpath rules + derivation asserts

    topCurHIV := topmostCrossingFrameCurHIV(m)     // nil if no crossing frame
    if topCurHIV == nil {
        panic("Sub: no live crossing frame")
    }
    recvHIV := realmHIV(recv)                      // nil for degenerate values
    if recvHIV == nil || recvHIV != topCurHIV {
        panic("Sub: receiver is not the live cur (stale, sibling, or a sub-token)")
    }

    if realmIsEphemeral(host) {
        panic("Sub: ephemeral realms cannot mint sub-identities")
    }
    if m.Realm == nil || m.Realm.Path != host {
        panic("Sub: caller is not operating in the receiver's namespace")
    }

    // Derive sub-token: synthesized pkgpath, derived address,
    // parent = the topmost Cur (stored like prev: a *.grealm-typed
    // value whose base is the parent HIV), prev = receiver's prev.
    ...
}
```

Helper definitions: `assertValidSubpath(host, subpath)` enforces the
subpath rules (see "Subpath validation") plus the derivation asserts —
host contains no `:`, and the synthesized `host+":"+subpath` is not a
run path. `topmostCrossingFrameCurHIV` walks frames topmost-down and
returns the Cur HIV of the first frame with `WithCross ||
DidCrossing` (mirroring today's `realmIsCurrentOnMachine` loop at
`uverse.go:414-433`); non-crossing frames — including the native
call's own — are skipped. The same helper backs the extended
`realmIsCurrentOnMachine`. `realmIsEphemeral` is the existing broad
`/e/` predicate backing `IsEphemeral()` (`uverse.go:437-447`).

This is intentionally stricter than the `IsCurrent()` method's
relaxation. The relaxation in `IsCurrent()` is *observational* — it
lets a sub-token report "I am live" when its parent is live, so that
non-crossing wrappers like the treasury banker accept sub-tokens
gracefully. But Sub()'s native bypasses the observational layer
entirely and demands the receiver's own HIV match the topmost Cur.

Consequence:

- Primary cur token (held by its frame as the topmost Cur): receiver
  HIV matches topmost Cur HIV. Sub() succeeds.
- Sub-token (fresh HIV allocated when Sub() was originally called):
  receiver HIV is *not* the topmost frame's Cur HIV — only the
  sub-token's internal `parent` field matches. The check fails.
  Sub-of-sub is structurally blocked.

This is load-bearing for defense in depth: if Sub() routed through
`IsCurrent()` (and thus through the parent-reference relaxation),
nested sub-minting would be stopped only by the caller-pkgpath check —
a single line of defense resting on the invariant that real pkgpaths
are colon-free. The strict pointer-identity check keeps sub-of-sub
structurally foreclosed, independent of path-shape invariants. A
regression test must assert that Sub() and IsCurrent() give different
answers for a live-parent sub-token (IsCurrent true; Sub panics), so a
future refactor cannot silently merge the two paths.

## Cross-time validation

`cross(rlm)` validates `rlm` via `realmIsCurrentOnMachine`
(`gnovm/pkg/gnolang/uverse.go:414-433`). Extension needed: when `rlm`
is a sub-token (non-empty subpath field), check the sub-token's
internal parent reference (not its own HIV) against the topmost
crossing frame's Cur.

Pseudocode for the extended check:

```go
func realmIsCurrentOnMachine(m *Machine, tv *TypedValue) bool {
    recvHIV := realmHIV(tv)
    if recvHIV == nil { return false }
    topCurHIV := topmostCrossingFrameCurHIV(m)  // skipping cross's own (non-crossing) frame
    if topCurHIV == nil { return false }

    // For sub-tokens, compare via the internal parent reference;
    // for primary cur tokens, compare directly.
    checkHIV := recvHIV
    sv := derefRealmStruct(tv)
    if len(sv.Fields) > fieldIdxSubpath && realmSubpathOf(sv) != "" {
        checkHIV = realmParentOf(sv)  // internal field, not exposed via Previous()
    }
    return checkHIV == topCurHIV
}
```

(The `len(sv.Fields)` guard is mandatory: AST-persisted origin
placeholders carry the legacy 3-field shape and must not
index-out-of-range. The same hazard exists at every new-field read
site — `isOriginRealmHIV`, `curUsesPreprocessOrigin`, this function,
and the Sub native. Centralize access behind
`realmSubpathOf`/`realmParentOf` accessors that treat missing fields
as zero.)

The same routine backs:
- `cross(rlm)`'s acceptance check.
- `rlm.IsCurrent()`'s relaxed evaluation for sub-tokens.

It is **NOT** used by Sub()'s native entry guard — that uses the
strict `recvHIV == topCurHIV` check directly (see previous section).
The distinction is load-bearing.

A stale sub-token — held in a local after the parent's crossing frame
has exited, or smuggled to a sibling frame within the same
transaction — fails the check because its parent's HIV is no longer
the topmost Cur. Capture *across* transactions is impossible outright:
sub-tokens cannot be persisted.

On that persistence guarantee: `refusePersistRealmHIV`
(`uverse.go:270-277`, fired in the save walk at `realm.go:985-991`) is
type-keyed on the `.grealm` type and therefore covers sub-tokens
automatically — **except** the `isOriginRealmHIV` carve-out
(`uverse.go:251-260`), which exempts realms whose `prev` is truly nil.
A sub-token minted from a truly-nil-prev cur (test frames) would be
wrongly exempt. The carve-out must gain a `subpath == ""` condition.
One line, but load-bearing.

## Banker compatibility (one `.gno`-layer guard, no native change)

The stdlib banker (`gnovm/stdlibs/chain/banker/banker.gno`) accepts any
`realm` value at `NewBanker(bt BankerType, rlm realm)`, stores
`rlm.Address()`/`rlm.PkgPath()`, and authorizes `SendCoins(from, to,
amt)` via `b.pkgAddr == from` (`banker.gno`). The native
`X_bankerSendCoins` (`banker.go:41-67`) delegates to the chain bank
keeper with no auth of its own — authorization lives entirely at the
`.gno` layer. Sub-tokens implement `realm` and report the correct
sub-address / synthesized pkgpath, so they flow through unchanged:

```go
sub := cur.Sub("dao/42")
b := banker.NewBanker(banker.BankerTypeRealmSend, sub)
// b.pkgAddr = chain.PackageAddress(synthesized pkgpath) = DAO sub-address
b.SendCoins(sub.Address(), dest, coins)   // .gno check: b.pkgAddr == from ✓
```

### Prerequisite fix: NewBanker must require `IsCurrent()`

The rlm-as-capability model — "you can only hold an rlm whose Address
represents authority you legitimately have" — is only sound if
`NewBanker` refuses realm values that are *not the live cur*. It did
not: `cur.Previous()` hands every callee a realm value bearing its
**caller's** address (`IsCurrent()==false`), and the old `NewBanker`
had no `IsCurrent()` gate for `RealmSend`/`RealmIssue`. A callee could
therefore build `NewBanker(RealmSend, cur.Previous())` (pkgAddr =
caller) and `SendCoins(caller, attacker, …)` — the `pkgAddr == from`
gate passes and the bank keeper moves the caller's coins. For a
MsgCall, that means any realm could drain the user who called it. This
is a **pre-existing base-v2 hole**, independent of sub-realms, but it
is exactly the path this section relies on, so the fix lands here:

```go
// banker.gno NewBanker, before the type-specific checks:
if !rlm.IsCurrent() {
    panic("banker can only be instantiated for the current realm")
}
```

Construction-time is the correct check point: the `banker` value holds
no realm handle, so use-time re-checking is impossible, and the
legitimate pattern is construct-from-live-cur-then-store/reuse (the
treasury `/p/` wrappers, `banker_persistence.txtar`). Live sub-tokens
pass (`IsCurrent` is parent-anchored for them), so the sub-token flow
above is preserved. `docs/resources/gno-security.md` Class 2 ("never
trust `rlm.Address()` without `IsCurrent()`") already prescribes this;
`NewBanker` was an unfixed instance. Regression tests:
`banker_security.txtar` TEST 6 (real-keeper drain blocked) and
`zrealm_banker_previous.gno` (VM-level).

With the gate, the rlm-as-capability invariant holds: `rlm` values are
produced only by frame mechanics, `Sub()` (gated), and `Previous()`
walks — and only the *live* one can mint a banker.

### Forbidden banker types for sub-tokens

Two defensive checks at `NewBanker`:

```go
// banker.gno NewBanker additions
if _, _, isSub := chain.SplitPkgSubPath(rlm.PkgPath()); isSub {
    switch bt {
    case BankerTypeRealmIssue:
        panic("BankerTypeRealmIssue not supported for sub-tokens")
    case BankerTypeOriginSend:
        panic("BankerTypeOriginSend not supported for sub-tokens")
    }
}
```

- **RealmIssue**: `assertCoinDenom` (`banker.gno:180-190`) prefixes
  denoms with `"/" + pkgPath + ":"`; a sub-token-issued denom would
  have a double-`:` shape (`"/host:dao/42:base"`). No forgery is
  possible either way (`isValidBaseDenom` rejects `:`, so a primary
  cannot fabricate a sub's denom), but coin issuance under a
  sub-namespace is an exotic capability; keeping the denom space clean
  for v1 costs nothing.
- **OriginSend**: the existing guard is `rlm.Previous().IsUserCall()`
  (`banker.gno:96`). Because `Previous()` skips the host, a sub-token
  would pass that guard whenever an EOA called the host — yielding an
  origin-send banker whose `pkgAddr` is a sub-address holding none of
  the origin-send coins. Nonsensical; forbid.

`BankerTypeRealmSend` (the treasury case) and `BankerTypeReadonly` are
unaffected.

### GRC20 / boards / treasury consumers

Standard crossing-function APIs read source from
`cur.Previous().Address()`, which is the sub-address after crossing
with a sub-token. Non-crossing token-style (`_ int, rlm
realm`) wrappers that gate on `rlm.IsCurrent() && rlm.Address() ==
owner` — e.g. the treasury banker `Send(_ int, rlm realm, p Payment)`
(`examples/gno.land/p/nt/treasury/v0/banker_coins.gno`) — work
because:

- `rlm.IsCurrent()` is true for sub-tokens of a live parent (relaxed
  semantics, safe due to synthesized pkgpath).
- `rlm.Address()` matches if the wrapper was constructed with the
  sub-token's address as owner.

No GRC20 standard change. No boards extension. No new banker entry.

On the `_ int` discriminator, since this document leans on it: a
`realm`-typed first parameter makes a function *crossing*, which `/p/`
code cannot declare. The `_ int, rlm realm` shape keeps a function
non-crossing while still accepting a realm token; callers pass `0` for
the discriminator (see `p/demo/tokens/grc20/tellers.gno`,
`p/nt/treasury/v0`). Rule of thumb: use the crossing form when the
callee authenticates via `cur.Previous()`; use the token form when the
callee takes `rlm` and checks `IsCurrent()` + `Address()`.

### Prior art: grc20 `RealmSubTeller`

`p/demo/tokens/grc20/tellers.gno` already ships `RealmSubTeller(_ int,
rlm realm, slug string)` deriving per-slug accounts as
`chain.PackageAddress(addr.String() + "/" + slug)` — an address-based,
`/`-separated scheme, incompatible with this proposal's pkgpath-based,
`:`-separated derivation (the code carries an "XXX: use a new std.XXX
call for this" note anticipating exactly this feature). Balances held
at RealmSubTeller-derived addresses will NOT match
`cur.Sub(slug).Address()`. Migration of RealmSubTeller onto `Sub()` is
out of scope but should be scheduled; until then the two schemes
coexist and must not be conflated in docs.

## Ephemeral (`/e/`) hosts

`Sub()` panics when the receiver is ephemeral (an `/e/<addr>/run`
cur). Without an explicit guard the other checks would naturally
permit it (`m.Realm` is the user's run realm and `cur.PkgPath()`
matches), but sub-identities of an ephemeral namespace make no sense:
run scripts are transient, arbitrary code acting for a user who
already has a first-class identity (their EOA), and any future ad-hoc
script by the same user would share the path and could re-mint the
same subs. Identity minting is reserved for persistent realms whose
code is on-chain and auditable. The guard is one predicate in the
Sub() native: the existing `realmIsEphemeral` (`uverse.go:437-447`),
the broad `/e/` check backing `IsEphemeral()` — deliberately not the
narrower `IsGnoRunPath`.

## Event attribution

`chain.Emit` attributes events by the emitting code's declaring
package. A host acting under a sub-identity therefore emits as the
host; no event surface carries the synthesized pkgpath. This is
intended for v1 — sub-realms are identity-only, and storage/event
attribution stays with the host. Indexers that need sub-level
attribution correlate via bank transfer events (sub-addresses are
ordinary addresses) or via callee-side emits that include
`cur.Previous().PkgPath()`. A dedicated attribution mechanism can
layer on later without changing this proposal.

## What this gives commondao

The `CommonDAO` type lives in `gno.land/p/nt/commondao/v0` (unexported
`id`, exposed via `ID()`); the realm cannot declare methods on it. The
sub-identity helpers are realm-level:

```go
// /r/nt/commondao/v0 (currently quarantined; see Verification)
package commondao

const pkgPath = "gno.land/r/nt/commondao/v0"

func subpathOf(daoID uint64) string {
    return "dao/" + strconv.FormatUint(daoID, 10)
}

// DAOAddress returns the DAO's chain-derived address. Pure; callable
// by anyone, matching chain.DerivePkgSubAddr(pkgPath, subpathOf(id)).
func DAOAddress(daoID uint64) address {
    return chain.DerivePkgSubAddr(pkgPath, subpathOf(daoID))
}

// Treasury spend, executed by the host realm's proposal machinery.
func executeTreasurySpend(cur realm, daoID uint64, dest address, coins chain.Coins) {
    sub := cur.Sub(subpathOf(daoID))
    b := banker.NewBanker(banker.BankerTypeRealmSend, sub)
    b.SendCoins(sub.Address(), dest, coins)
}

// GRC20 spend. Token-style /p/ tellers take the sub-token directly:
func executeTokenSpend(cur realm, daoID uint64, teller grc20.Teller, dest address, amount int64) {
    sub := cur.Sub(subpathOf(daoID))
    teller.Transfer(0, sub, dest, amount)
}
// Crossing /r/ token wrappers (e.g. wugnot) take a cross instead:
//   wugnot.Transfer(cross(sub), dest, amount)
```

Note `executeTokenSpend` crosses into (or hands a live token to) a
caller-supplied interface value; per the interrealm spec's
arbitrary-interface caveat, the token/teller must be whitelisted by
the DAO's proposal machinery, not accepted from arbitrary proposers.

For a DAO acting as a board admin (boards2 permissions key on
`cur.Previous().Address()` — no boards2 changes needed):

```go
sub := cur.Sub(subpathOf(daoID))
boards2.ChangeMemberRole(cross(sub), boardID, member, role)
// Inside boards2: caller = cur.Previous().Address() = DAO sub-address
```

Granting the role in the first place is ordinary address-based
administration — some board-admin code runs
`perms.SetUserRoles(commondao.DAOAddress(daoID), boards2.RoleAdmin)`.

## VM changes summary

| Change | Where | Size |
|---|---|---|
| Add unexported `subpath` (string) and `parent` (`*.grealm`-typed, HIV-compared) fields to `.grealm` — transient, never persisted, inaccessible from user code | `uverse.go` (`gConcreteRealmType`) | 2 fields |
| Update `newRealmHIVPointer` + its construction sites; add `subpath == ""` condition to `isOriginRealmHIV`; `len(Fields)` guards wherever new fields are read (legacy 3-field AST-persisted origins exist) | `uverse.go` | ~20 lines |
| Add `Sub()` to the `realm` interface (`gRealmType`) and as a native pointer method with the strict entry guard + caller-pkgpath check | `uverse.go` | ~50 lines |
| Add `Sub` to the gno0p9 go/types realm shim **in place** (0.9 is the latest language version; additive on a sealed interface — strictly more programs type-check, no existing package breaks). Type-checking is consensus: activation must ride a coordinated node upgrade so all validators accept `Sub`-using packages from the same height | `gotypecheck.go` | ~2 lines |
| Extend `realmIsCurrentOnMachine` for sub-tokens; extract `topmostCrossingFrameCurHIV` helper shared with Sub's strict guard | `uverse.go` | ~15 lines |
| Extend `execctx.GetRealm` to walk the presented-identity chain (crossing frame's `Cur` → `.prev`^height) with existing origin fallbacks; same change in the test-stdlib mirror (incl. `RealmOverride` reconciliation) | `gnovm/stdlibs/internal/execctx/realm.go`, `gnovm/tests/stdlibs/chain/runtime/testing_runtime.go` | ~60–100 lines across 2 files |
| Add `chain.DerivePkgSubAddr` + `chain.SplitPkgSubPath` (pure Gno, reuse existing native) | `gnovm/stdlibs/chain/address.gno` | ~15 lines |
| Explicit `:` rejection in package path validation | `gnovm/pkg/gnolang/mempackage.go` (`ValidateMemPackageAny`) | 1 conditional |
| Defensive derivation asserts (host colon-free in both sites; `!IsGnoRunPath(synthesized)` native-only — the `.gno` helper's host-colon-free assert subsumes it, see "Derivation safety coupling") | Sub native + `DerivePkgSubAddr` | ~4 lines |
| Forbid `RealmIssue`/`OriginSend` for sub-token rlm | `gnovm/stdlibs/chain/banker/banker.gno` | ~8 lines |
| Meter `Sub` inside the native body (`m.incrCPU` as the body's first statement, so failed calls pay too; base + per-byte slope over the synthesized path, mirroring `packageAddress`'s calibrated entry). A `native_gas.go` table entry would be dead code: uverse natives have `NativePkg == ""`, and `chargeNativeGas` short-circuits to flat `OpCPUCallNativeBody` without consulting the table | Sub native in `uverse.go` | ~5 lines |

Total: roughly 150–300 lines of runtime/stdlib changes across ~7
files, excluding tests, docs, and the step-0 `cross1` removal
(first commit of the same PR).

**Notable absence**: no *native* banker changes are required. The
stdlib banker accepts sub-tokens through the standard
rlm-as-capability path; authorization remains the `.gno`-layer address
equality check. There is no native banker authorization mechanism to
extend, and adding pkgpath-keyed checks to trusted native code would
import the prefix-matching footgun — deliberately avoided.

## Subpath validation

Subpaths follow a deliberately strict, **frozen-at-introduction**
grammar (via `subRealmPathError`):

```
subpath := segment ("/" segment)*
segment := [a-z0-9] ( [a-z0-9_.-]* [a-z0-9] )?
constraint: len(host) + 1 + len(subpath) <= 256
```

Lowercase-alphanumeric segments, `/`-separated, with `_`, `.`, `-`
allowed only *inside* a segment. This excludes: uppercase, whitespace,
control bytes, non-ASCII (so no NFC/NFD or RTL-override address
ambiguity — two visually identical subpaths can't derive two
addresses), `:` and NUL, empty segments (leading/trailing/double `/`),
edge punctuation, and `..`/`.` traversal segments. Examples that pass:
`dao/42`, `treasury`, `role_admin`, `v1.2`, `g1…` addresses. Examples
that fail: `Dao`, `a b`, `../x`, `dao/`, `a//b`, `_x`.

Why frozen: derived sub-addresses are permanent. Loosening the grammar
later is additive (safe); **tightening it later would strand funds**
already sent to addresses derived from a now-illegal subpath. So the
narrow set is chosen up front.

The **total** cap is on the synthesized `host:subpath` (≤ 256, the
existing `pkgPathLimit`), not the subpath alone — this keeps every
downstream pkgpath-sized buffer/field valid, rather than letting
`PkgPath()` reach ~513 bytes.

`chain.DerivePkgSubAddr` enforces the identical grammar (its
`assertValidSubpath` mirrors `subRealmPathError` — a byte-for-byte
transliteration, kept in sync by comment and by
`TestIsValidSubpath`), so an address can never be derived off-chain
that `cur.Sub` would refuse to mint.

## What this does NOT do

- **Per-sub-realm storage isolation.** The host realm's state stays in
  the host realm's storage. Sub-realms are identity-only. If you want
  per-DAO storage isolation, a separate proposal (virtual realm
  instances) can layer on top.
- **Off-chain custody.** Sub-addresses can receive coins, and the host
  realm can spend from them via banker. No off-chain key material is
  involved. For multisig-style custody, use the constitution's
  alternative model (DAO as m-of-n multisig account).
- **Cross-host identity.** If a host realm at `/r/foo/v0` migrates to
  `/r/foo/v1`, sub-tokens minted under v0 have addresses different
  from sub-tokens of v1 (different parent pkgpath). For treasuries
  this is a **fund-stranding risk**: only the v0 host can mint v0
  sub-tokens, so retiring v0 without first executing a sub-treasury
  migration permanently locks coins held at v0 sub-addresses. Hosts
  holding meaningful sub-treasuries must ship a migration executor
  (drain each sub-address to its v1 counterpart) before upgrade.
  Address continuity across host upgrades requires application-level
  mapping.

## Footguns

**1. Passed-cur in `/p/` code.** If `/r/X` invokes `/p/` code with its
own cur — a top-level `/p/` function (m.Realm inherited = /r/X) or a
`/p/`-method on an /r/X-stamped receiver (borrow rule #2) — that code
can call `cur.Sub("anything")` and mint subs in /r/X's namespace. This
is a trust extension /r/X makes to its `/p/` imports, same in kind as
today (those packages can already `cross(cur)` to act as /r/X's
primary identity), but larger in blast radius:

**Banker-drain corollary**: because `DerivePkgSubAddr` is public and
pure, every address in /r/X's sub-namespace is enumerable off-chain. A
malicious `/p/` helper that receives /r/X's live cur — a top-level
`/p/` function or a method on an /r/X-stamped receiver, the two shapes
that still execute with `m.Realm = /r/X`; a method on a
foreign-stamped receiver borrows `m.Realm` away, so its Sub() panics —
can construct
`banker.NewBanker(banker.BankerTypeRealmSend, rlm.Sub(target))` for
the specific sub-address holding the largest balance, and the native
send path performs no auth. The reachable set is not "subs /r/X has
minted" but **every sub /r/X could ever mint**. Treat passing `cur` to
a `/p/` helper as granting signing authority over the entire
sub-address namespace. Mitigations: import only trusted `/p/`
packages; clear agency (`safely()`-style, per gno-interrealm) before
invoking untrusted helpers with a live cur.

**2. Subpath collisions across logical types.** A single realm using
the same subpath for two unrelated purposes (e.g., `"42"` for a DAO ID
AND a committee ID) creates a sub-address collision. Mitigation:
caller convention — `"dao/42"`, `"committee/42"` (`:` is reserved;
use `/` or `.` for structure).

**3. Tooling parsing of synthesized pkgpath.** Off-chain tools may not
recognize `gno.land/r/nt/commondao/v0:dao/42` as a non-package. The
`:` separator is the marker; `chain.SplitPkgSubPath` is the canonical
parse. Standardize in tooling docs.

**4. Salt grinding vs known addresses.** Address derivation is a
preimage-resistant hash (SHA-256-truncated over the synthesized path);
finding a subpath that collides with a target EOA address requires
breaking the hash. Not feasible. (The run-path extraction shortcut,
the one non-hash derivation, is foreclosed by the defensive asserts —
see "Derivation safety coupling".)

## Migration

Existing code is unchanged by default. No callee needs to update
unless it wants to support sub-tokens.

When a callee opts in:

- Address-keyed auth: works automatically (sub-addresses are just
  addresses). Note that granting a role to a sub-address is trusting
  the host realm's current *and future* code, exactly as granting to
  the host's primary address is.
- PkgPath-keyed auth: opt in via the anchored idiom
  (`p == host || strings.HasPrefix(p, host+":")`) or explicit
  synthesized-form match.
- IsCurrent+Address-keyed wrappers (treasury banker, token tellers):
  work automatically if constructed with the sub-token.

**Audit note — deployed bare-prefix checks**: any *existing*
`strings.HasPrefix(rlm.PkgPath(), ...)` auth check will start
accepting sub-identities of matching hosts without its author opting
in. The shift is bounded — only hosts already inside the matched
namespace can mint such subs — but it is a semantic change. Known
instance: boards2's `assertCallerHasBoardsNS`
(`r/gnoland/boards2/v1/protected.gno`) prefix-matches
`"gno.land/r/gnoland/boards2/"`. Before landing, sweep `examples/` and
deployed realms for `HasPrefix` over pkgpaths and confirm each is
acceptable or convert to the anchored idiom.

commondao opts in by minting `cur.Sub(subpathOf(daoID))` in its
proposal executors and exposing `DAOAddress()`.

## Sequencing

0. **Precursor (first commit of this PR) — remove `cross1`**: the legacy v1→v2
   migration sentinel that lowers to the `.origin`-shaped AST. Touch
   points: uverse def (`uverse.go:1477`), gno0p9 shim line
   (`gotypecheck.go:82`), first-arg recognizer (`nodes.go:439`),
   preprocess cases (`preprocess.go:1327-1334`, `:2030,2042`),
   op_call comment references (`op_call.go:74,92` — the `.origin`
   branch itself stays for compiler-synthesized origin), and the
   transpiler's ident rewrite (`transpiler.go:251-262`). The one
   legacy filetest (`gnovm/tests/files/zrealm_cross1_legacy.gno`)
   converts to an error test asserting the sentinel no longer
   resolves. Update `gnovm/adr/migration_guide.md` §16, which
   documents `cross1` as a supported long-tail shim (this removal
   revokes that promise), and add a removed-as-of note to
   `gnovm/adr/interrealm_v2.md`.

   Policy: `cross1` was a temporary migration sentinel; nobody should
   be using it, and its removal is not gated on compatibility. For
   awareness, the failure mode if a straggler exists: a stored
   `cross1` package fails the boot-time re-preprocess
   (`keeper.go:153-168`) and a `cross1` tx in replayed history fails
   type-checking. Cheap due diligence: grep any network state meant
   to survive before rollout; a straggler migrates (`cross1` →
   `cross(rlm)`) or the network relaunches. The in-repo tree —
   examples, genesis fixtures, testdata, stdlibs — is verified
   cross1-free. Batch activation with step 1's additive `Sub` change
   in one coordinated upgrade.

1. **VM change**: `.grealm` fields + constructor/carve-out plumbing,
   Sub() native (strict guard + caller-pkgpath check + validation),
   extended `realmIsCurrentOnMachine`, realm-interface + typechecker
   shim, `:` rejection at mempackage validation, derivation asserts,
   `chain.DerivePkgSubAddr`/`SplitPkgSubPath`, gas entry.
2. **unsafe parity**: extend `execctx.GetRealm` + test-stdlib mirror
   to the presented-identity chain.
3. **Banker guards**: forbid RealmIssue/OriginSend for sub-tokens.
4. **Tests**: filetests under `gnovm/tests/files/zsubrealm_*.gno`
   exercising:
   - Standard Sub() flow via cross(); callee observations
     (cur.Previous vs unsafe.* parity assertions, heights 0/1, both
     MsgCall and MsgRun entry).
   - Address derivation determinism; `DerivePkgSubAddr` equivalence.
   - Caller-pkgpath check rejection (all table rows, incl. nil
     m.Realm).
   - Nested Sub() rejection; regression test that IsCurrent(true) +
     Sub(panic) diverge for a live-parent sub-token.
   - Sub()-side strict-guard rejections: no live crossing frame
     panics; a stale or sibling primary cur (stashed in a global,
     read after its frame pops within the same tx) panics on Sub().
   - Subpath validation rules; derivation asserts.
   - IsCurrent semantics for sub-tokens; classification methods table.
   - Staleness: sub-token held after parent frame exits → cross()
     rejected; sibling-frame capture rejected.
   - Persistence refusal: storing a sub-token in realm state aborts;
     nil-prev-parent sub-token also refused (isOriginRealmHIV
     carve-out).
   - Legacy 3-field AST-persisted origin values flow through
     IsCurrent/cross/Sub paths without out-of-range panics (accessor
     guards).
   - `/p/` helper receiving a live cur mints a sub and drains via
     banker (documents footgun #1 as a test, not just prose).
   - Legitimate nesting across frames: a callee entered by crossing
     with a sub-token minting its own `cur.Sub(y)` (allowed — its cur
     is a fresh primary).
   - Sub() on an `/e/`-run cur panics (ephemeral hosts forbidden).
   - Banker: RealmSend with sub-token works; RealmIssue/OriginSend
     panic; treasury-style token wrapper accepts a sub-token.
   - Go unit test (not a filetest): `ValidateMemPackageAny` rejects
     pkgpaths containing `:`.
5. **commondao integration**: `TreasurySpendProposal` using
   `cur.Sub(subpathOf(daoID))` for native and GRC20 spends;
   `DAOAddress()` helper. Note the `/r/` realm is currently quarantined (see
   Verification).
6. **Docs**: update `gno-interrealm-v2.md` (Sub section + the
   identity-chain semantics of `unsafe.*`), tooling docs
   (`SplitPkgSubPath`), CONSTITUTION.md cross-reference. Schedule
   RealmSubTeller migration discussion separately.

## Verification

For the VM change:

1. `go test ./gnovm/...` — VM unit tests pass.
2. New filetests per Sequencing step 4.
3. Existing filetests pass without modification (backwards
   compatibility).
4. Per AGENTS.md, before declaring done:
   `go test ./gno.land/pkg/sdk/vm/ -run Gas`,
   `go test ./gno.land/pkg/integration/ -run TestTestdata`,
   `go test ./gnovm/pkg/gnolang/ -run Files -test.short`.

For commondao integration — note `/r/nt/commondao/v0` currently lives
at `examples/quarantined/gno.land/r/nt/commondao/v0` (quarantined);
either un-quarantine as part of the integration or target the
quarantined path:

1. `gno test ./examples/gno.land/p/nt/commondao/v0/...` passes
   (examples are `.gno`; `go test` matches nothing there).
2. Tests for the (un)quarantined `/r/` realm pass.
3. New filetest: full treasury flow (deposit to DAO sub-address; pass
   TreasurySpendProposal; assert funds arrive at destination).
4. New filetest: DAO as board admin (grant role to
   `DAOAddress(daoID)`; DAO acts via `cur.Sub`).
5. New filetest: attacker realm holding a captured sub-token attempts
   `cross()` after the parent's frame has exited — assert
   realmIsCurrentOnMachine rejects.

## Out of scope

- Per-DAO state isolation under separate PkgID (virtual realm
  instances).
- Constitution-level governance features (early termination, ancestor
  governance, treasury rules, Bylaws / Mandates).
- DAO-as-multisig account model.
- RealmSubTeller migration onto `Sub()` (tracked as follow-up; see
  Prior art).
- Sub-level event attribution (see Event attribution).

Each can be pursued separately. None block this proposal, and this
proposal does not preclude them.

## Alternatives considered

**`crosssub(...)` fused call form** (the CROSSUB.md draft): a second
special form performing mint+cross in one step at the call site, e.g.
`target.Foo(crosssub("dao/42"), ...)`. Rejected because it produces no
first-class `realm` value: non-crossing token-style consumers —
`banker.NewBanker`, treasury `Send(_ int, rlm realm, ...)`, GRC20
tellers — take a realm *value*, and the flagship
zero-native-banker-change result depends on sub-identities being such
values. A fused form would require new API entries in every such
consumer (or a token form anyway), plus a second special form in the
parser/typechecker alongside `cross`. The honest trade-off: the fused
form would have needed almost none of the token-holding security
machinery (no staleness surface, no persistence refusal, no IsCurrent
relaxation, no sub-of-sub guard) — roughly half this document exists
to make the held token safe. The token form was chosen because the
ecosystem's non-crossing APIs are where sub-identities are most
useful; the machinery is the price, bounded by the strict entry guard
and HIV checks.

**`Previous() = parent cur`** (host inserted as a chain step): would
make the host visible to `Previous().Previous()` walks, at the cost of
changing chain depth at every sub-crossing and breaking callees that
assume fixed depth. Rejected: depth consistency wins; the host stays
recoverable via the synthesized pkgpath (`SplitPkgSubPath`).

**Documented `unsafe.*` divergence** (instead of extending the
walker): keep `execctx.GetRealm` on the crossing-context chain, so
`unsafe.PreviousRealm()` reports the host while `cur.Previous()`
reports the sub. Defensible — `unsafe.*` is debug-only, and "context
chain vs identity chain" is a coherent split — but rejected because a
debug API that disagrees with what the callee's auth logic observed is
a forensics trap, and the fix is small (the identity chain already
sits on the frames as `Cur`/prev).

**Distinct sub-realm type** (instead of returning `realm`): a separate
type (or an added `Subpath()`/`IsSub()` accessor) would make sub-ness
visible in signatures, but every existing consumer takes `realm`;
drop-in compatibility is the point. Sub-ness remains detectable via
`SplitPkgSubPath(rlm.PkgPath())`. The classification-method table
covers the behavioral questions a distinct type would have answered
structurally.

## Why this is the right design

- **Smallest surface that solves the actual problem.** Each piece is
  justified by a concrete use case.
- **Compatible with both crossing-function and non-crossing API
  patterns.** The token-style API works wherever a `realm` value can
  be supplied; the synthesized pkgpath ensures existing checks remain
  precise.
- **No silent semantic shifts for exact-match auth.** Existing
  exact-match pkgpath checks behave as today. Sub-token acceptance is
  opt-in via the anchored idiom — with the one caveat that deployed
  bare-prefix checks must be audited (see Migration).
- **Footguns are bounded and have clear mitigations.** The passed-cur
  trust footgun is the only new exposure; it has the same mitigation
  as the existing `/p/`-trust model, with the blast radius stated
  honestly (entire enumerable sub-namespace).
- **Constitutional fit.** DAOs (and other sub-actors) get real
  chain-derived addresses they control via their host realm. Receiving
  coins is automatic; spending is gated by the host realm's proposal
  machinery.

The synthesized pkgpath is the load-bearing improvement; the token
form is the load-bearing compatibility win; the rest are bounded
protections.
