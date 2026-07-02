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
if !IsUser() { panic("not a user") }

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
// WRONG: if fn is a top-level /p/-declared function, it inherits
// the caller's m.Realm and can write to your state
func ApplyHook(fn func()) { fn() }

// RIGHT: type the callback with your own /r/-declared type so
// /p/ code can't supply a matching implementation
func ApplyHook(fn func(*MyState)) { fn(gState) }
```

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

### 7. `/p/`-embedded types with callback iterators

If a realm field is a `/p/`-type with methods like `Iterate(cb func(*Node) bool)`,
attackers can supply a top-level `/p/`-function that runs with your realm's authority.
Keep such fields unexported **and** do not return aliased pointers to them.

---

## Review Checklist

- [ ] Authenticated mutators take `cur realm` and call `cur.IsCurrent()`
- [ ] Payment-guarded functions use `cur.Previous().IsUserCall()`
- [ ] No exported function returns a pointer to internal mutable state
- [ ] No method accepts a `func(...)` callback with a `/p/`-typed parameter and invokes it
- [ ] Interface parameters from external callers are guarded with canonical-type asserts
- [ ] No `realm`-typed value in package-level vars, struct fields, or closure captures
- [ ] `/p/`-type fields with callback iterators are unexported
- [ ] Data types holding sensitive state are declared in this realm (`/r/`), not in shared `/p/`

---

## Relationship to Other Docs

| Resource | Purpose |
|----------|---------|
| [`gno-security-guide.md`](./gno-security-guide.md) | Deep technical explanation of the threat model, borrow rules, and anti-patterns |
| [`gno-security.md`](./gno-security.md) | Numbered threat-class taxonomy |
| [`gno-interrealm.md`](./gno-interrealm.md) | Cross-realm call mechanics (`cur realm`, `IsCurrent()`, borrow rules) |
| [`effective-gno.md`](./effective-gno.md) | Idiomatic Gno patterns including payment guards |
| `misc/audit-pattern-harness/` | Automated pattern detection tooling with sanitized fixtures |

This guide distills the above into the shortest checklist that catches the most critical issues.
