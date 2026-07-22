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
    if !cur.IsCurrent() { panic("spoofed realm") }
    addr := cur.Previous().Address()
    ...
}
```

### 2. Payment guards — `IsUserCall()`, not `IsUser()`

```go
// WRONG: MsgRun ephemeral realms pass IsUser()
if !cur.Previous().IsUser() { panic("not a user") }

// RIGHT
if !cur.Previous().IsUserCall() { panic("not a direct user call") }
```

### 3. No exported pointers to mutable state

```go
// WRONG: attacker can call mutator methods on returned pointer
func GetAccount() *Account { return gAccount }

// RIGHT: return a copy, or expose read-only accessors
func GetBalance() int { return gAccount.balance }
```

### 4. No caller-supplied callbacks invoked with realm authority

```go
// WRONG: the parameter type is /p/-declared, so a top-level /p/
// function can match it and run under your realm's authority.
// Same hole for any caller-supplied func, down to plain func().
func ApplyHook(fn func(*somelib.Node)) { fn(gNode) }

// RIGHT: type the callback with your own /r/-declared type so
// /p/ code can't supply a matching implementation
func ApplyHook(fn func(*MyState)) { fn(gState) }
```

### 5. Interface parameters need canonical-type assertion

```go
// WRONG: no check. Evil{Teller} satisfies grc20.Teller, runs its own Transfer.
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
// WRONG: storing a live realm value panics at tx finalize
var savedRealm realm
func Remember(cur realm) { savedRealm = cur }

// RIGHT
var savedAddr address
func Save(cur realm) {
    if !cur.IsCurrent() { panic("spoofed realm") }
    savedAddr = cur.Previous().Address()
}
```

### 7. `/p/`-type with callback iterators

If a realm field is a `/p/`-type with methods like `Iterate(cb func(*Node) bool)`,
attackers can supply a top-level `/p/`-function that runs with your realm's authority.
Keep such fields unexported **and** do not return aliased pointers to them.

### 8. `/p/`-type with mutation methods returned as pointer

This is the subtlest case. A `/p/` library type whose fields are all **unexported** can
still be a write-authority leak if it has exported mutation methods and you return a pointer
to an instance stored in your realm.

```go
// avl.Tree fields are all unexported — looks safe.
// But Tree has exported mutation methods: Set, Remove, etc.
var store = avl.NewTree()

// WRONG: attacker calls store.Set(key, value) on the returned pointer.
// Borrow rule #2 fires (tree was allocated in this realm) → m.Realm = /r/V
// for the method body → the write inside Set commits under your authority.
func GetStore() *avl.Tree { return store }

// RIGHT: never return the tree pointer. Expose only what you control.
func GetValue(key string) (any, bool) { return store.Get(key) }
```

The rule: **any exported method on a `/p/` type that writes to its receiver is a
mutator**. If you return a pointer to an instance, that mutator is now callable by
anyone with the authority of your realm.

#### Sub-case: exported pointer fields

The same path exists through exported pointer fields of `/p/` structs:

```go
// p/mylib
type Container struct {
    Items *avl.Tree   // exported pointer field
}

// r/V
var c = &Container{Items: avl.NewTree()}
func GetContainer() *Container { return c }

// Attacker: c.Items.Set(key, value) → borrow rule #2 on Items
// (Items was allocated in /r/V) → write commits.
```

Readonly taint on `c` does NOT block this: method dispatch is not a write operation,
so the taint check does not fire. Borrow rule #2 fires first on method entry and
authorizes the writes inside the method body.

**Rule**: treat every exported pointer field of a `/p/` type as if it were a direct
pointer to mutable state. If the pointed-to type has any mutation method, it is a
live mutator handle. Never return the containing struct as a pointer.

### 9. `unsafe.PreviousRealm()` — old API, skips frame verification

Using `chain/runtime/unsafe.PreviousRealm()` directly bypasses the `cur.IsCurrent()`
safety check. It should never appear alongside a `cur realm` parameter.

```go
// WRONG: cur is accepted but ignored; no IsCurrent() guard
import "chain/runtime/unsafe"
func Set(cur realm, key, value string) {
    caller := unsafe.PreviousRealm().Address()
    ...
}

// RIGHT
func Set(cur realm, key, value string) {
    if !cur.IsCurrent() { panic("spoofed realm") }
    caller := cur.Previous().Address()
    ...
}
```

Flag `unsafe.PreviousRealm()` or `unsafe.CurrentRealm()` used for caller identity in a crossing function. The `unsafe.OriginCaller()` and `unsafe.OriginSend()` tx-origin primitives have no `cur` substitute and are legitimate.

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

---

## Review Checklist

- [ ] Authenticated mutators take `cur realm` and panic unless `cur.IsCurrent()`
- [ ] No `unsafe.PreviousRealm()` or `unsafe.CurrentRealm()` used for caller identity in a crossing function
- [ ] Payment-guarded functions use `cur.Previous().IsUserCall()`
- [ ] No exported function returns a pointer, slice, or map aliasing internal mutable state
- [ ] No exported function returns a `/p/`-type pointer whose type has mutation methods
- [ ] No exported `/p/`-struct field is itself a pointer to a type with mutation methods
- [ ] No method invokes a caller-supplied func or interface value whose signature `/p/` code can satisfy, including plain `func()`
- [ ] Interface parameters from external callers are guarded with canonical-type asserts
- [ ] No `realm`-typed value in package-level vars, struct fields, map values, slice elements, or closure captures
- [ ] `/p/`-type fields with callback iterators are unexported and not reachable via a returned or promoted method
- [ ] Sensitive state in a `/p/`-declared type (e.g. `grc20.Token`) is stored in unexported vars with no leaked pointers
- [ ] `Render` sanitizes path segments, keys, and user-supplied values before writing to output

---

## Relationship to Other Docs

| Resource | Purpose |
|----------|---------|
| [`gno-security-guide.md`](./gno-security-guide.md) | Deep technical explanation of the threat model, borrow rules, and anti-patterns |
| [`gno-security.md`](./gno-security.md) | Numbered threat-class taxonomy |
| [`gno-interrealm-v2.md`](./gno-interrealm-v2.md) | Cross-realm call mechanics (`cur realm`, `IsCurrent()`, borrow rules) |
| [`effective-gno.md`](./effective-gno.md) | Idiomatic Gno patterns including payment guards |
| `misc/audit-pattern-harness/` | Automated pattern detection tooling with sanitized fixtures |

This guide distills the above into the shortest checklist that catches the most critical issues.
