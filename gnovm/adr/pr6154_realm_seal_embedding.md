# Seal markers must be directly declared, not promoted through embedding

## Context

The uverse `realm` type is an interface, but only one concrete type is meant to
satisfy it: the runtime realm value `.grealm`. GnoVM enforces this with a *seal*:
a hidden marker method named `.seal` is added to the `realm` interface
(`gnovm/pkg/gnolang/uverse.go`, `gRealmType`) and declared natively on `.grealm`
(`defNativePtrMethod(".grealm", ".seal", …)`). Because the Gno parser rejects
identifiers beginning with `.`, user code cannot declare a `.seal` method, so a
hand-written struct cannot structurally satisfy `realm`. This is what the
existing `gnovm/tests/files/zrealm_seal_realm.gno` filetest asserts.

The seal underpins an authorization invariant: code that receives a `realm`
value trusts that its *liveness* (`IsCurrent()`) and its *identity*
(`Address()`, `PkgPath()`) come from the same concrete `.grealm`. For example,
`chain/banker.NewBanker` authorizes a `RealmSend`/`RealmIssue` banker by reading
`rlm.IsCurrent()` and then binding the banker to `rlm.Address()` — all through
virtual interface dispatch.

## Problem

The seal blocked *declaration* of `.seal` but not *promotion* of it. Go method
promotion means a struct that embeds the `realm` interface inherits the interface's
entire method set — including `.seal` — into its own method set:

```go
type fooRealm struct {
	realm            // embeds a live realm value; promotes .seal, IsCurrent, …
	forged address
}

func (r fooRealm) Address() address { return r.forged } // shadows the promoted Address
func (r fooRealm) PkgPath() string  { return "…" }
```

`fooRealm` satisfies `realm` (it has the promoted `.seal`), yet its own
`Address()`/`PkgPath()` methods *shadow* the promoted ones while `IsCurrent()`
still dispatches to the embedded genuine realm. Identity and liveness now
originate from two different concrete values: liveness from the real embedded
realm, identity from attacker-chosen overrides.

This breaks every authorizer that trusts virtual dispatch on `realm`.
`NewBanker` is the concrete case: `IsCurrent()` returns true (embedded realm),
`Address()` returns an arbitrary address (override), so the banker binds to an
address the caller does not control. The downstream `pkgAddr == from` gate and
the native send path perform no further realm authentication, so a realm can
move the spendable balance of an arbitrary address — including system realms
whose package address is derivable from their path (e.g. draining a wrapped-token
reserve while its ledger reads unchanged).

`cross()` and `Sub()` are unaffected because the VM authenticates them against
the concrete realm value, not virtual calls.

The type-checker split compounds the issue: the Go type-checker's `realm`
definition (`gnoBuiltinsGno0p9` in `gotypecheck.go`) lists the public methods but
has no `.seal` (it is undeclarable), so the forgery passes go/types via
promotion; the VM's method-set check is the only layer that can enforce the seal.

## Decision

Enforce the seal in `InterfaceType.VerifyImplementedBy` — the single
interface-satisfaction chokepoint, shared by static assignment
(`type_check.go`), type assertions and conversions (`op_expressions.go`), and
type switches (`op_exec.go`).

A **sealed marker method** (any method whose name begins with `.`; undeclarable
in user source, so only the runtime declares one) must be satisfied by a method
**declared directly on the concrete type**, never one **promoted through an
embedded field**. Resolution distinguishes the two by the length of the
`findEmbeddedFieldType` trail: a directly-declared method resolves as a lone
method value-path (a single-element trail), whereas a promoted one prepends at
least one embedded-field hop, so its trail is longer than one element. When a
marker resolves through a field hop, satisfaction is rejected with the same
`missing method` result as an absent method. (Testing the length rather than the
trail head's `VPType` also covers the pointer-method head forms —
`VPSubrefField` / deref-method — that `applyPointerDeref` can produce.)

This restores the original invariant — only `.grealm`, which declares `.seal`
natively, satisfies `realm` — and closes the class at the type layer rather than
patching individual authorizers: a forged type no longer satisfies `realm`
anywhere, so it cannot be assigned, asserted, or switched to a `realm`, and a
realm using it fails to type-check (and, on-chain, fails to deploy).

## Consequences

- **No legitimate breakage.** `.seal` is the only dot-named marker method in the
  tree, and no legitimate type embeds `realm` to satisfy it; the genuine
  `.grealm` declares `.seal` directly and is unaffected. The rule generalizes to
  any future sealed interface for free.
- **Minimal surface.** ~40 lines in one function plus two small helpers
  (`isSealedMarkerName`, `trailPromotedThroughField`); no change to the resolver,
  the parser, or the go/types layer.
- **Defense-in-depth left open (not in this change).** Authorizers such as
  `NewBanker` could additionally anchor to the concrete `.grealm` value the way
  `cross()`/`Sub()` do, rather than trusting virtual dispatch. Worth a follow-up
  audit of any code that authorizes off a virtual `IsCurrent()` together with
  `Address()`/`PkgPath()`.

## Testing

- `gnovm/tests/files/zrealm_seal_embed.gno` — the embedding-promotion path is
  rejected with `does not implement .uverse.realm (missing method .seal)`,
  mirroring the structural-fake `zrealm_seal_realm.gno`.
- `gno.land/pkg/integration/testdata/banker_realm_embed.txtar` and
  `banker_realm_embed_wugnot.txtar` — end-to-end: a realm that embeds a live
  realm and overrides `Address()` fails to deploy, so no banker is constructed
  and balances/reserves are untouched.
- Regression: `banker_security`, `banker_persistence`, the existing seal
  filetest, and the full `gnovm/pkg/gnolang` suite pass.
