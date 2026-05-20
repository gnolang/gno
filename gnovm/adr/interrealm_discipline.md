# Interrealm Security Discipline — Library & Realm Author Guide

Operational guidance for writing safe `/p/` libraries and `/r/` realms
under the v3a (post-PR #5669) interrealm model. Assumes v3a Phase A
(machine ops + generalized primitive-recv anchor) has landed.

## What v3a gives you for free

Before any discipline, the language already defends against most
attack classes:

- **omarsy forgery**: closed by concrete `runtime.Realm` struct.
- **`.Title()`-injection via object receivers**: closed by Layer-2
  receiver-borrow (m.Realm shifts to receiver's PkgID).
- **`.Title()`-injection via primitive/nil receivers**: closed by
  generalized anchor (m.Realm shifts to receiver type's declaring
  pkg, which for attacker-authored impls equals attacker's realm).
- **Stack-walk tampering**: closed by transition-bounded
  `runtime.Caller()` walk.
- **Mutation of victim's data through pointer params**: closed by
  v2's allocation-time PkgID + readonly check at write sites.

## What discipline addresses

The residual surface that the language can't close structurally
without a marker model or capability typing (see
`interrealm_capability_levels.md`):

- Read-side data flows through interface boundaries (data is public
  on-chain, so usually a non-issue, but worth knowing).
- Capability leakage if the author hands a constructed capability to
  an untrusted impl through an interface.
- Interface contracts that re-introduce primitive-recv attack shapes
  by accepting /p/-declared mutable references.

These are addressed by **library design conventions**, not by VM
changes. The conventions below are short, lint-friendly, and concrete.

## The five rules

### Rule 1 — Interface contract design

**Don't accept `*your.MutableType` as a parameter in interface
methods.** Pass a value (copy) or a read-only wrapper instead.

```go
// BAD — interface methods take a mutable ref
type Voter interface {
    Vote(p *Proposal)            // attacker impl can mutate p
}

// GOOD — pass a value, body works on a copy
type Voter interface {
    Vote(p Proposal) Proposal    // returns the updated copy
}

// GOOD — pass a read-only wrapper
type Voter interface {
    Vote(p ProposalView)         // ProposalView exposes only getters
}
```

The fix is per-interface. Once an interface is designed with this
rule, attacker impls satisfying it have no mutable surface to attack.

### Rule 2 — Capability handling

**Don't pass capabilities through interface methods.** Obtain them
inside the trusted realm via `runtime.Caller()` / `runtime.Self()`.

```go
// BAD — caller hands a banker to an arbitrary impl
type Payer interface {
    Pay(b banker.Banker, amount uint64)
}

// GOOD — impl constructs its own banker inside its body
type Payer interface {
    Pay(amount uint64)          // body queries runtime.Caller() etc.
}
```

If the impl needs caller authority for a side effect, **caller wraps
the operation in a closure** and passes the closure (its provenance
carries caller's realm):

```go
type Payer interface {
    Pay(amount uint64, onSuccess func())
}

// /r/me caller
mypayer.Pay(100, func() {
    // closure's OriginRealm = /r/me
    // when impl invokes it, m.Realm borrows back to /r/me
    /r/me.recordPayment(...)
})
```

### Rule 3 — Allowlist gating at trust boundaries

**Type-assert interface values to a known concrete-type allowlist
before invoking methods.** Switch on the concrete type; reject the
default.

```go
func ProcessVoter(v Voter, p Proposal) Proposal {
    switch concrete := v.(type) {
    case *TrustedImpl1, *TrustedImpl2:
        return concrete.Vote(p)
    default:
        panic("ProcessVoter: unrecognized impl")
    }
}
```

This is a per-call-site review pattern, not a general rule applied
to all interface invocations. Use it specifically at boundaries
where you accept interface values from untrusted callers.

### Rule 4 — Top-level functions over methods for shared utilities

**`/p/` libraries should expose top-level functions instead of
methods when the operation needs caller authority.** Top-level
functions auto-borrow correctly; methods follow the anchor rule.

```go
// BAD — primitive-recv method that needs caller authority
type Weight uint32
func (w Weight) Apply(p *Proposal) { p.tally += uint32(w) }
// after generalized anchor: m.Realm shifts to /p/voting, write fails

// GOOD — top-level function
type Weight uint32
func Apply(w Weight, p *Proposal) { p.tally += uint32(w) }
// callers: voting.Apply(w, &p) — gets normal /p/ borrow, write succeeds
```

This applies specifically to **primitive/nil-receiver methods**.
Struct/pointer-to-struct receivers are unaffected (Layer-2 handles
them correctly because they carry per-value PkgID).

### Rule 5 — Don't write primitive-receiver methods that take /p/-mutable pointer params

**The attack shape is**: primitive receiver + a parameter that's a
pointer to a /p/-declared mutable type. Anywhere this pattern
appears, rewrite as a top-level function (Rule 4).

```go
// BAD — attack-shape signature
type Counter int
func (c Counter) Apply(d *DAO) { ... }       // /p/-declared *DAO param

// GOOD — top-level function in the same package
type Counter int
func ApplyCounter(c Counter, d *DAO) { ... }
```

After the generalized anchor lands, the BAD shape's body will still
panic at the readonly check when it tries to mutate `d` — but
authors should prefer not to write the shape at all, both to avoid
the runtime surprise and because the rewrite is mechanically
equivalent.

## Lint coverage

Rules 1, 4, 5 are mechanically lintable:

- **Lint rule for Rule 1**: scan interface declarations; flag method
  specs whose parameters include pointers to mutable types declared
  in `/p/` packages.
- **Lint rule for Rule 4 + 5**: scan method declarations; flag
  methods with primitive-underlying receivers that have pointer or
  interface parameters. Suggest top-level-function rewrite.

Rule 2 is a security-guide convention; no static lint applies but
code review patterns can catch obvious violations (any interface
method accepting `banker.Banker`, etc.).

Rule 3 is per-callsite review work; document the allowlist-gating
pattern in the security guide so authors recognize the idiom.

## Summary table

| Rule | Lintable? | Closes |
|---|---|---|
| 1. No /p/-mutable refs in interface params | ✅ yes | The downstream re-opening of attack-H shape |
| 2. Don't pass capabilities through interfaces | ⚠️ partial | Capability leakage to attacker impls |
| 3. Allowlist gating at boundaries | ❌ pattern-only | Untrusted-impl substitution |
| 4. Top-level fns for caller-authority utilities | ✅ yes | Anchor friction on legit primitive-recv methods |
| 5. No primitive-recv methods with /p/-mutable params | ✅ yes | The attack-shape construct itself |

## Reference contracts

These contracts demonstrate the discipline in practice:

- `/p/grc20`, `/p/grc721` — fungible/non-fungible token primitives.
  All public methods take primitive or /r/-declared args only.
- `/r/wugnot` — wrapped native token.
- `/r/grcfactory` — token factory.
- `/r/grcreg` — token registry.
- `/r/atomicswap` — atomic-swap protocol.

If you're designing a new contract, copy the patterns from these.
If you're auditing existing code, check it against the rules above
and against these examples.

## Relation to v3a, v3b, and the marker model

- **v3a alone** + this discipline closes the practical attack surface
  for well-designed contracts. This is the deferred path.
- **Marker model (level 2)** would convert Rule 1 into a structural
  invariant — the language refuses non-borrow impls from satisfying
  borrow interfaces. Lands if discipline isn't holding in the field.
- **Capability typing (level 2.7)** would extend the marker to read/
  write granularity. Lands if the marker isn't enough.

See `interrealm_capability_levels.md` for the full design space.

## `revive()` vs `defer recover()` after migration

When migrating a `/p/` helper from the v2 `(_ int, rlm realm, ...)` shape
to a v3a plain signature, callers may also need to switch how they
catch panics from that helper.

- **`revive(fn)`** catches **cross-realm aborts** (a runtime event:
  panic crossing a frozen-Realm or /r/-to-/r/ boundary). Tied to the
  VM-level transition, not source syntax. Persists into Phase C
  unchanged.
- **`defer recover()`** catches **same-realm panics** (normal Go
  semantics, no realm transition involved).

After v3a migration:
- A `/p/` method called *directly* from another /p/ frame (no realm
  transition) → use `defer recover()`. The v2-era `revive()` no longer
  applies.
- A `/p/` method called *across a realm boundary* (e.g., the test
  harness's `cross()` wrapping still in place, or invocation from a
  different /r/) → keep `revive()`.

Concrete instance in this PR: `/p/agherasie/forms.SubmitForm` was
called via `cross2(cur)` wrapper in v2 (cross-realm abort path,
caught by `revive`). After migration, the test calls SubmitForm
directly (same /p/ frame), so the panic is regular and `defer recover()`
is appropriate. The wrapper was dropped; `revive()` was replaced.

Dropping the `cross` keyword in Phase C does **not** affect `revive()`
— `revive` operates on VM-level events (frame's `IsRevive` flag,
unwind on cross-realm panic), which persist regardless of source syntax.

## Where to enforce

| Layer | Mechanism |
|---|---|
| Author awareness | Security guide + this document |
| Pre-commit | Lint rules for Rules 1, 4, 5 |
| Code review | Pattern recognition for Rule 3 |
| Runtime | v3a's generalized anchor + readonly check (defense-in-depth) |
| Test corpus | Reference contracts as canonical templates |

Defense layers compose. Author awareness + lint catches most
violations at design time; runtime checks catch slips that escape
into deployed code.
