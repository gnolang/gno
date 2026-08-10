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

The first `cur realm` of a crossing function is guaranteed current by the runtime,
so an `IsCurrent()` check on it is dead code (see `gnovm/adr/interrealm_v2.md` and
the filetest `gnovm/tests/files/zrealm_iscurrent.gno`). The check is load-bearing
on every realm value a function is handed other than that first `cur`: a secondary
`rlm realm` parameter — on a non-crossing helper, or alongside a crossing
function's own `cur` — can be filled with a forwarded or derived realm value,
typically `cur.Previous()`:

```go
// Non-crossing helper receiving the acting realm as a plain value.
// `_ int` keeps `rlm` out of first position — a realm-typed first
// parameter would make this a crossing function.
func helper(_ int, rlm realm) {
    if !rlm.IsCurrent() { panic("rlm is not the current realm") }
    addr := rlm.Previous().Address()
    ...
}
```

Passing a realm value is also a capability transfer, not just an identity claim:
the callee can mint `banker.NewBanker(banker.BankerTypeRealmSend, rlm)` and retain
it across transactions (see the SECURITY note in
`gnovm/stdlibs/chain/banker/banker.gno`). Only hand a realm value to code you
would trust with permanent spend authority over your realm's address.

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
// WRONG: a top-level /p/-declared fn triggers no borrow rule, so your
// realm's m.Realm stays in effect for its body and it can write your state
func ApplyHook(fn func()) { fn() }

// RIGHT: type the callback with your own /r/-declared type so
// /p/ code can't supply a matching implementation
func ApplyHook(fn func(*MyState)) { fn(gState) }
```

### 5. Interface parameters need canonical-type assertion

```go
// WRONG: Evil{Teller} embedding bypasses interface checks
func DoBanking(cur realm, t grc20.Teller, to address, amount int64) {
    t.Transfer(0, cur, to, amount)   // who is t? could be Evil{Teller}
}

// RIGHT: assert the concrete type before dispatch
func DoBanking(cur realm, t grc20.Teller, to address, amount int64) {
    if !grc20.IsCanonicalTeller(t) { panic("not a canonical Teller") }
    t.Transfer(0, cur, to, amount)
}
```

### 6. Do not store `realm` values

`realm` values are ephemeral — store `Address()` or `PkgPath()` strings instead.

```go
// WRONG: panics when the value is attached, or at finalize
var savedRealm realm
func Save(cur realm) { savedRealm = cur }

// RIGHT
var savedAddr address
func Save(cur realm) { savedAddr = cur.Previous().Address() }
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
func GetValue(key string) any  { return store.Get(key) }
func HasValue(key string) bool { return store.Has(key) }
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

Using `chain/runtime/unsafe.PreviousRealm()` directly bypasses the frame verification
that a runtime-current `cur realm` provides. It should never appear alongside a
`cur realm` parameter.

```go
// WRONG: cur is accepted but ignored; identity comes from a stack walk
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

Inside a crossing function the two calls return the same realm, so this shape is
a greppability defect rather than an exploit. The exploitable shape is a
non-crossing helper: there `unsafe.PreviousRealm()` returns whichever realm was
outermost-crossing in the chain, not the immediate caller, so
`unsafe.PreviousRealm().PkgPath() == "gno.land/r/admin"` is an authentication
bypass. Accept `_ int, rlm realm` and check `rlm.IsCurrent()` instead (case 1).

Flag any import of `chain/runtime/unsafe` in a realm that also has `cur realm`
parameters, and any `unsafe.PreviousRealm()`-based authentication in a
non-crossing function.

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

// RIGHT: TableCell for table cells — it adds `|` escaping on top of
// InlineText, which leaves `|` literal by design. InlineText is for
// headings and other inline text.
import "gno.land/p/nt/markdown/sanitize/v0"
return "# Vault: " + sanitize.InlineText(path) + "\n"
b.WriteString("| " + sanitize.TableCell(key) + " | " + sanitize.TableCell(val) + " |\n")
```

### 11. `GetCoins` to read one balance — attacker-influenced cost

Any realm can mint denoms in its own namespace (`/<pkgpath>:<base>`) to any address
without the holder's consent, and since anyone can deploy a realm, nothing bounds
how many distinct denoms an address accumulates. `banker.GetCoins(addr)`
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

- [ ] Authenticated mutators take `cur realm` and derive the caller from `cur.Previous()`
- [ ] Every realm value other than the frame's own first `cur` — any secondary `rlm realm` parameter, in a helper or in a crossing function — is checked with `rlm.IsCurrent()` before `Previous()`/`Address()`/`PkgPath()` is trusted
- [ ] No realm value is passed to code outside this realm's trust boundary (the callee can mint and retain a `banker.NewBanker(BankerTypeRealmSend, rlm)` — permanent spend authority)
- [ ] No import of `chain/runtime/unsafe` alongside `cur realm` parameters
- [ ] Payment-guarded functions use `cur.Previous().IsUserCall()`
- [ ] No exported function returns a pointer to internal mutable state
- [ ] No exported function returns a `/p/`-type pointer whose type has mutation methods
- [ ] No exported `/p/`-struct field is itself a pointer to a type with mutation methods
- [ ] No exported function or method invokes a caller-supplied function or interface value (including a bare `func()`) while holding its own realm authority
- [ ] Interface parameters from external callers are guarded with canonical-type asserts
- [ ] No `realm`-typed value in package-level vars, struct fields, or closure captures
- [ ] `/p/`-type fields with callback iterators are unexported
- [ ] Data types holding sensitive state are declared in this realm (`/r/`), not in shared `/p/` (see [`gno-security-guide.md`](./gno-security-guide.md), author checklist)
- [ ] `Render` sanitizes path segments, keys, and user-supplied values before writing to output
- [ ] Single-denom balance checks use `GetCoin(addr, denom)`, not `GetCoins(addr).AmountOf(denom)` — unless the denom is unvalidated caller input or the full set is already read (case 11)

---

## Relationship to Other Docs

| Resource | Purpose |
|----------|---------|
| [`gno-security-guide.md`](./gno-security-guide.md) | Deep technical explanation of the threat model, borrow rules, and anti-patterns |
| [`gno-security.md`](./gno-security.md) | Numbered threat-class taxonomy |
| [`gno-interrealm-v2.md`](./gno-interrealm-v2.md) | Current cross-realm call mechanics (`cur realm`, `IsCurrent()`, sub-realms) |
| [`gnovm/adr/interrealm_v2.md`](../../gnovm/adr/interrealm_v2.md) | The v2 spec and the three borrow rules |
| [`effective-gno.md`](./effective-gno.md) | Idiomatic Gno patterns including payment guards |
| `gnovm/tests/files/zrealm_launder_*.gno` | Exploit-attempt filetest corpus, each annotated with the attack mechanism and outcome |

This guide distills the above into the shortest checklist that catches the most critical issues.
