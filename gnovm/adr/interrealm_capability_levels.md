# Interrealm Capability Levels — Design Space Beyond v3a's Marker Model

Tracking document for design alternatives that extend v3a's deferred
marker model toward finer-grained access control. Not a committed
plan; an exploration of what's available between "minimal language
change" and "Rust-style ownership."

Use this as a reference if v3a's library-discipline path proves
insufficient and structural mechanisms need to be considered.

## Context

v3a Phase A + generalized primitive-recv anchor closes most of the
attack surface jaekwon identified:

- omarsy (forgeable `realm` interface) → closed by concrete
  `runtime.Realm` struct.
- .Title()-injection class via object receivers → closed by v2's
  Layer-2 receiver-borrow.
- .Title()-injection class via primitive/nil receivers (Attack-H
  same-pkg case) → closed by generalized anchor.

The deferred-marker plan accepts that the residual class —
**attackers self-marking borrow on impls of borrow-accepting
interfaces** — is handled by:

1. Library-discipline: interface designers refuse borrow contracts
   for most interfaces.
2. Lint: scan for /p/-declared mutable refs in interface method
   parameters at trust boundaries.
3. Marker model (deferred): only if (1)+(2) prove insufficient.

This document explores **what's between the marker model and
Rust-style ownership** in case the deferral needs to be revisited.

## The capability spectrum

Ordered from minimal to maximal language change:

| Level | Mechanism | Closes | Cost |
|---|---|---|---|
| 1 | Marker on methods only | Documentation; partial defense | ~50 lines |
| 2 | Marker on methods + interfaces (level 2) | Self-marked borrow attack | ~150 lines |
| 2.5 | Region tags on parameters | Adds static authority tracking | ~300 lines |
| 2.7 | Capability-typed access (read/write/mut) | Bounds *what specifically* can be mutated | ~500 lines |
| 2.9 | Pony-style reference capabilities | Full structural authority + immutability | ~1000 lines |
| 3 | Effect types | Body verification against declared effects | Major language work |
| 4 | Rust-style ownership + borrow checker | Aliasing + authority + lifetime | Massive |

For Gno's on-chain sequential threat model, **2.7–2.9 is the realistic
sweet spot** if discipline + lint isn't enough. Level 3 and Level 4
are likely over-engineering.

## Level 1 — Method-only marker

```go
type Weight uint32
func (w Weight) borrow Apply(p *Proposal) { ... }
```

Unmarked methods anchor to declaring pkg (safe default). Marked
methods get caller authority (opt-in).

**Closes:** documents intent for code review. Audit surface narrowed
to grep-able `borrow` markers.

**Doesn't close:** attacker self-marks their malicious impl as
`borrow`, gets caller authority anyway. The author of the impl
controls the marker — attackers control their own impls. Self-marking
is unconstrained.

**Verdict:** insufficient alone. Documentation tool only.

## Level 2 — Marker + interface propagation

```go
type Voter interface { Vote(p *Proposal) }              // non-borrow
type Hasher interface { borrow Hash() []byte }          // explicitly borrow

type EvilWeight uint32
func (e EvilWeight) borrow Vote(p *Proposal) { ... }
// COMPILE ERROR: Voter requires non-borrow methods
```

Interface designers declare whether borrow impls are admitted. Type
checker enforces match at satisfaction site.

**Closes:** attacker can self-mark `borrow` but their impl won't fit
into a non-borrow interface slot. Designer-controlled trust boundary.

**Doesn't close:**
- Interfaces that *do* admit borrow are still attack surfaces — the
  designer accepts that arbitrary impl bodies run with caller
  authority for those interfaces.
- Read-exfiltration (returning sensitive data via the return value).
  Not relevant in public-blockchain context (state is already public).

**Verdict:** the deferred candidate. Real structural defense for the
self-marked-borrow class.

## Level 2.5 — Region tags on parameters

```go
func (e Evil) Apply(p *Proposal @caller) {
    // p is tagged with caller's realm
    // type system tracks that body's writes through p require caller authority
}
```

Parameters carry their realm in the type system as a static annotation
(Cyclone-style region typing). Mutations are gated by region match at
compile time, not just runtime readonly check.

**Closes:** static detection of "this method tries to mutate caller's
data" — caught at compile time instead of runtime panic.

**Doesn't close:** doesn't bound *what* can be mutated; if `@caller`
authority is granted, full mutation is allowed.

**Verdict:** mostly diagnostic improvement over runtime check. Not a
big leap in security.

## Level 2.7 — Capability-typed access

Each parameter declares what capabilities the method needs:

```go
func (w Weight) Apply(p *Proposal #write)
    // method requires write capability on p

func (e Evil) Inspect(p *Proposal #read)
    // method has only read capability on p
    // body cannot perform p.field = ... — type error
```

The caller decides at the call site what capabilities to grant:

```go
victim.RunInspector(handler, &p)
// victim passes &p with #read capability — handler can't write
```

**Closes:** even when the caller grants borrow authority, the method
can only do what its declared capability allows. Attacker handlers
declared `#read` cannot escalate to writes — type error.

**Doesn't close:** if the interface designer chose `#write`, attacker
impls can write through `p` freely. But the choice is now per-param,
not per-method — finer-grained.

**Verdict:** real semantic addition. Attackers can't widen their
declared access. Worth taking if level 2 isn't enough.

## Level 2.9 — Pony-style reference capabilities

A well-explored design from the Pony language. Each value reference
carries a capability tag:

| Cap | Meaning | Read | Write | Sendable |
|---|---|---|---|---|
| `iso` | unique, mutable, sendable across actors | ✅ | ✅ | ✅ |
| `val` | immutable, sendable | ✅ | ❌ | ✅ |
| `ref` | mutable, local only | ✅ | ✅ | ❌ |
| `box` | read-only view | ✅ | ❌ | ❌ |
| `tag` | identity only (methods OK, no field access) | ❌ | ❌ | ✅ |
| `trn` | write-only transitional | ❌ | ✅ | ❌ |

Methods declare what capability they require for each parameter; the
type system enforces.

```go
func (w Weight) Apply(p box *Proposal)
    // p is box — body can read p but not write
    // type checker rejects p.field = x

func (w Weight) Mutate(p ref *Proposal)
    // p is ref — body can write to p, but p cannot escape this method
```

**Closes:**
- Mutation through param: only with explicit `ref`/`iso` capability.
- Aliasing escape: `iso` and `ref` are non-sendable across realms.
- Immutability claims: `val` and `box` are statically immutable.
- Identity-only access: `tag` allows calling methods without reading
  fields.

**Doesn't close:** call patterns where the interface designer
chose unrestricted (`ref`/`iso`) for the params. But those become
visible audit targets.

**Verdict:** structurally strong defense. Pony has 20+ years of
deployment experience showing the system is workable. Cost is real
language complexity — the cap discipline pervades all parameter and
field declarations.

## Level 3 — Effect types

Methods declare what side effects they can have; the compiler reads
the body and verifies the annotation:

```go
func (t *Tree) Hash() []byte
    reads t.*             // only reads receiver
    no writes
    no allocs
    no calls

func (w Weight) Apply(p *Proposal)
    reads w
    writes p.tally        // ONLY p.tally, not p.passed or anything else
    no allocs
    no calls
```

**Closes:**
- Self-marked borrow with broader-than-declared effects → rejected.
- Read-exfiltration via captured-state calls → declared `no calls`
  blocks the exfiltration path.
- Allocation attribution attacks → declared `no allocs` enforced.
- Body verification → attacker can't lie about effects; compiler
  reads the body.

**Doesn't close:** still relies on interface designer choosing
restrictive effect signatures. Aliasing not tracked.

**Verdict:** strong but expensive. The body-verification phase is
nontrivial — needs interprocedural analysis or per-method local
verification with declared assumptions about callees. Multi-year
language work.

## Level 4 — Rust-style ownership + borrow checker

Each value has exactly one owner. References are either shared
(immutable, many) or unique (mutable, one). The borrow checker
statically tracks reference lifetimes and rejects aliasing violations.

For Gno's threat model, Rust-style adds:
- Aliasing control: prevents two methods writing through the same
  pointer simultaneously (irrelevant for sequential execution).
- Lifetime tracking: prevents use-after-free (irrelevant for GC'd
  language).

**Verdict:** over-engineering for Gno. The cost-benefit doesn't
favor Rust complexity in a single-threaded, garbage-collected,
on-chain runtime.

## What changes between levels for jaekwon's concern

Recall jaekwon's worry: attacker supplies an interface impl that has
caller authority and mutates caller's /p/-declared data.

| Level | jaekwon's scenario |
|---|---|
| v2 + anchor | ✅ closed for primitive-recv attack-H (anchor shifts m.Realm to attacker pkg) |
| Level 2 (marker) | ✅ additionally closes self-marked-borrow on non-borrow interfaces |
| Level 2.5 (regions) | ➕ static catch instead of runtime readonly panic |
| Level 2.7 (capabilities) | ✅ even when borrow is granted, attacker bound to declared access |
| Level 2.9 (Pony caps) | ✅ additionally closes aliasing/sendability escape |
| Level 3 (effect types) | ✅ additionally closes attacker bodies doing more than declared |
| Level 4 (Rust) | ➕ aliasing safety (not needed for Gno) |

After v2 + anchor, the residual is whatever level 2.7+ would close.

## Recommendation if marker isn't enough

If the library-discipline + lint path fails (corpus shows violations,
production exploit, community demand), the smallest escalation that
still closes the structural class is:

1. **Level 2 marker** as the first language addition. Closes
   self-marked-borrow. Smallest cost.
2. **Level 2.7 capability typing** if level 2 still leaves gaps —
   attacker impls that legitimately satisfy borrow interfaces but
   mutate beyond what the designer intended. Adds read/write
   distinction.

Skip 2.5 (mostly diagnostic) and 2.9 (good but Pony's full cap
system is heavier than needed). Skip 3 and 4 (over-engineering).

## Trigger conditions

Move from marker discipline to language-level enforcement when:

- The lint rule shows recurring violations in third-party `/p/`
  libraries that authors are unwilling or unable to refactor.
- A concrete exploit lands in production traceable to the gap that
  the marker would close.
- The community shifts toward wanting opt-in borrow contracts as a
  first-class language feature.
- Reference safe-template contracts (grc20, wugnot, etc.) cannot be
  made safe under the discipline rule — i.e., legitimate library
  designs require breaking the rule.

Move from marker (level 2) to capability typing (level 2.7) when:

- Level 2 lands and corpus shows interfaces that need to admit borrow
  but want bounded mutation (not full write authority).
- Recurring exploits within borrow-accepting interface boundaries.

## Summary

The marker model is the smallest structural defense. Capability typing
(level 2.7) is the next step if marker isn't enough. Effect types and
Rust are over-engineering. The library-discipline path may make all of
this unnecessary — that's the deferral bet v3a takes.

This document captures the design space so that if/when the marker
discussion is reopened, the alternatives are already laid out.
