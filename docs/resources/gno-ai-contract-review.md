# Gno Contract Review Guide for AI Agents

A concise reference for AI agents performing security review of `.gno` realm code.
For the full threat model and worked examples, see [`gno-security-guide.md`](./gno-security-guide.md).

---

## Quick Checks

These are the highest-yield issues to look for in any realm:

### 1. Caller identity — use `cur realm`, not `address` parameters

```go
// WRONG: address parameter is attacker-controlled
func AdminAction(caller address) { ... }

// RIGHT: derive identity from the live crossing frame
func AdminAction(cur realm) {
    addr := cur.Previous().Address()
    ...
}
```

A realm's stored state changes only during a crossing call into it, so every
writer another caller reaches must be crossing. An exported non-crossing
function or method is read-only for other callers; flag one that is meant to
write. No `cur.IsCurrent()` guard is needed in a crossing function.

### 2. Payment guards — `IsUserCall()`, not `IsUser()`

```go
// WRONG: MsgRun ephemeral realms pass IsUser()
if !IsUser() { panic("not a user") }

// RIGHT
if !cur.Previous().IsUserCall() { panic("not a direct user call") }
```

### 3. No exported pointers to mutable state

```go
// WRONG: aliases internal state; foreign writes are refused, but every
// method of the type is published and later changes widen it
func GetAccount() *Account { return gAccount }

// RIGHT: return a copy, or expose read-only accessors
func GetBalance() int { return gAccount.balance }
```

### 4. No caller-supplied callbacks invoked with realm authority

```go
// WRONG: fn runs view-only or read-only, so it cannot write your
// state, but it runs inside your call and can read or abort it
func ApplyHook(fn func()) { fn() }

// RIGHT: type the callback with your own /r/-declared type so
// /p/ code can't supply a matching implementation
func ApplyHook(fn func(*MyState)) { fn(gState) }
```

A plain callback supplied by another realm runs read-only. If the callback
must write its own realm, it is a crossing function the other realm calls with
`cross`; a registry of crossing functions stores `func(cur realm, ...)` values,
not plain functions.

### 5. Interface parameters need canonical-type assertion

```go
// WRONG: Evil{Teller} embedding bypasses interface checks
func DoBanking(t grc20.Teller) { t.Transfer(...) }

// RIGHT: assert the concrete type before dispatch
func DoBanking(t grc20.Teller) {
    if !grc20.IsCanonicalTeller(t) { panic("not a canonical Teller") }
    t.Transfer(...)
}
```

### 6. Do not store `realm` values

`realm` values are ephemeral — store `Address()` or `PkgPath()` strings instead.

```go
// WRONG: panics at attach time
var savedRealm realm

// RIGHT
var savedAddr address
func Save(cur realm) { savedAddr = cur.Previous().Address() }
```

### 7. `/p/`-type with callback iterators

If a realm field is a `/p/`-type with methods like `Iterate(cb func(*Node) bool)`,
attackers can supply a top-level `/p/`-function as the callback. A function picked by
the caller runs view-only, so its writes are refused, but it still reads every node.
Keep such fields unexported **and** do not return aliased pointers to them.

### 8. `/p/`-type with mutation methods returned as pointer

This is the subtlest case. A `/p/` library type whose fields are all **unexported**
can still have exported mutation methods. A pointer you return to an instance
stored in your realm is a view: a `/p/` method runs with the caller's authority,
so the caller's write to your stored object is refused with
`cannot modify object owned by another realm from library code (call a function of the owning realm)`.

```go
// avl.Tree fields are all unexported — looks safe.
// But Tree has exported mutation methods: Set, Remove, etc.
var store = avl.NewTree()

// The caller's store.Set(...) on the returned pointer is refused.
func GetStore() *avl.Tree { return store }

// RIGHT: never return the tree pointer. Expose only what you control.
func GetValue(key string) (any, bool) { return store.Get(key) }
```

The rule: a reference that leaves your realm is a view. To let another realm
write the tree, expose a crossing function
(`func Set(cur realm, k string, v any) { store.Set(k, v) }`). Fresh-object
corner: a `/p/` object your function builds and returns carries your realm's
stamp, so the caller cannot write it either. Never write through a reference
received as an argument: a caller can pass the realm's own object back, and your
function would write whichever of your objects the caller picked.

#### Sub-case: exported pointer fields

The same exposure exists through exported pointer fields of `/p/` structs:

```go
// p/mylib
type Container struct {
    Items *avl.Tree   // exported pointer field
}

// r/V
var c = &Container{Items: avl.NewTree()}
func GetContainer() *Container { return c }
```

An exported pointer field exposes the pointed-to type's method set. A foreign
caller's write to your stored object is refused, but the field still widens what
you publish; keep it unexported and expose the reads you mean to.

**Rule**: treat every exported pointer field of a `/p/` type as part of the API
you publish.

### 9. `unsafe.PreviousRealm()` — old API, skips frame verification

`chain/runtime/unsafe.PreviousRealm()` walks the frame stack and answers from
wherever it is called; `cur.Previous()` names the realm that was current when `cur`
was minted. It should never appear alongside a `cur realm` parameter.

```go
// WRONG: cur is accepted but ignored
import "chain/runtime/unsafe"
func Set(cur realm, key, value string) {
    caller := unsafe.PreviousRealm().Address()
    ...
}

// RIGHT
func Set(cur realm, key, value string) {
    caller := cur.Previous().Address()
    ...
}
```

Flag any import of `chain/runtime/unsafe` in a realm that also has `cur realm` parameters.

### 10. Unsanitized user input in `Render`

`Render(path string)` receives attacker-controlled input. Writing path segments, keys,
or user-supplied values directly into markdown output enables injection (broken table
cells, injected links, heading overrides).

```go
// WRONG: path, keys, and values written raw
func Render(path string) string {
    return "# Vault: " + path + "\n"          // heading injection
}

// ALSO WRONG: table cell content not escaped
b.WriteString("| " + key + " | " + val + " |\n")  // | in key breaks table

// RIGHT: escape pipe characters at minimum; use sanitize.InlineText for
// full inline markdown escaping
import "gno.land/p/nt/markdown/sanitize/v0"
b.WriteString("| " + sanitize.InlineText(key) + " | " + sanitize.InlineText(val) + " |\n")
```

### 11. `GetCoins` to read one balance — attacker-influenced cost

Any realm can mint an arbitrary denom to any address without the holder's consent, and
nothing bounds how many distinct denoms an address accumulates. `banker.GetCoins(addr)`
reads every one of them, so its cost is set by whoever last sent that address a coin —
not by the realm. When `addr` comes from the caller, a third party can make the function
run out of gas, permanently.

```go
// WRONG: cost grows with denoms the address happens to hold
if banker.NewReadonlyBanker().GetCoins(addr).AmountOf("ugnot") < price {
    panic("insufficient balance")
}

// WRONG AND QUADRATIC: a full read per iteration
for _, coin := range coins {
    if bnk.GetCoins(realmAddr).AmountOf(coin.Denom) < coin.Amount { ... }
}

// RIGHT: one denom, one store read
if banker.NewReadonlyBanker().GetCoin(addr, "ugnot") < price {
    panic("insufficient balance")
}
```

Reserve `GetCoins` for cases that genuinely need every balance, and treat it as
unbounded when the address is caller-supplied. Hoisting the call out of a
loop helps but is not enough: a second `GetCoins` on the same address is much cheaper —
its per-key reads are cache hits, and a cache hit costs no gas — but the iterator walk
over the address's balance keys is charged again, so the part that scales with the denom
count survives. Measured on an address holding 64 unsolicited denoms, a second
`GetCoins` cost about a quarter of the first, while a second single-denom `GetCoin` cost
nothing at all.

Two cases where the swap is **wrong**, both found by making it:

- **The denom is caller-supplied.** `GetCoin` panics on a malformed denom where
  `AmountOf` returns zero. In `Render`, the denom is often a path segment or query
  parameter, so the swap hands any visitor a way to break the page.

  ```go
  // WRONG in Render: denom comes from the URL, and a malformed one now panics
  denom := req.Query.Get("coin")
  amount := bnk.GetCoin(addr, denom)
  ```

- **You already hold the full set.** If `GetCoins` was already called for another
  reason, `AmountOf` on the result is free and `GetCoin` is a second store read. Read
  the surrounding function before swapping a call in it.

---

## Review Checklist

- [ ] Every writer other callers reach is crossing (`cur realm`) and authenticates with `cur.Previous()`; no `cur.IsCurrent()` guard
- [ ] No import of `chain/runtime/unsafe` alongside `cur realm` parameters
- [ ] Payment-guarded functions use `cur.Previous().IsUserCall()`
- [ ] No exported function returns a pointer to internal mutable state
- [ ] No exported function returns a `/p/`-type pointer whose type has mutation methods
- [ ] No exported `/p/`-struct field is itself a pointer to a type with mutation methods
- [ ] No method accepts a `func(...)` callback with a `/p/`-typed parameter and invokes it
- [ ] Errors returned to other realms are pointer errors (`*MyError`), never struct values
- [ ] A closure or function value I hand out carries no write power for another realm; writers are crossing functions or crossing closures
- [ ] A method value of a `/p/` type bound on my object writes nothing, wherever it runs (see [the four rules (R2a)](./gno-interrealm.md)); my crossing function performs the write
- [ ] A registry of crossing functions stores `func(cur realm, ...)` values and calls them with `cross`
- [ ] Interface parameters from external callers are guarded with canonical-type asserts
- [ ] No `realm`-typed value in package-level vars, struct fields, or closure captures
- [ ] `/p/`-type fields with callback iterators are unexported
- [ ] Data types holding sensitive state are declared in this realm (`/r/`), not in shared `/p/`
- [ ] `Render` sanitizes path segments, keys, and user-supplied values before writing to output
- [ ] Single-denom balance checks use `GetCoin(addr, denom)`, not `GetCoins(addr).AmountOf(denom)` — unless the denom is unvalidated caller input or the full set is already read (case 11)

---

## Relationship to Other Docs

| Resource | Purpose |
|----------|---------|
| [`gno-security-guide.md`](./gno-security-guide.md) | Deep technical explanation of the threat model, write rules, and anti-patterns |
| [`gno-security.md`](./gno-security.md) | Numbered threat-class taxonomy |
| [`gno-interrealm.md`](./gno-interrealm.md) | Cross-realm call mechanics (`cur realm`, `IsCurrent()`, the four rules) |
| [`effective-gno.md`](./effective-gno.md) | Idiomatic Gno patterns including payment guards |
| `misc/audit-pattern-harness/` | Automated pattern detection tooling with sanitized fixtures |

This guide distills the above into the shortest checklist that catches the most critical issues.
