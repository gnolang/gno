# Interrealm: a whitepaper

A short statement of how realms protect their state. The mechanism and the
full call-by-call table are in [gno-interrealm.md](./gno-interrealm.md); the
decision record is `gnovm/adr/pr6495_cross_to_write.md`.

## 1. A realm is a Go package with state

A realm (`gno.land/r/...`) is a Go package whose package-level variables
persist between transactions. It also has an address and can hold coins. A
library (`gno.land/p/...`) is a Go package without state: after `init` its
variables never change.

## 2. Why realms need a boundary

Take a token. In Go, packages trust each other completely:

```go
// package token
var Balances = map[address]int64{}
func Transfer(from, to address, n int64) { Balances[from] -= n; Balances[to] += n }

// any importer
token.Balances[me] = 1_000_000_000       // nothing stops it
token.Transfer(alice, me, 100)           // and nothing says who is calling
```

That is fine inside one program with one author. On a chain, every package
has a different author and some of them are hostile. Solidity answers with a
wall around each contract's storage and one kind of door:

```solidity
contract Token {
    mapping(address => uint) balances;              // nobody outside can touch it
    function transfer(address to, uint n) public {  // the door
        balances[msg.sender] -= n;                  // the door knows who knocked
        balances[to] += n;
    }
}
```

Gno keeps Go's packages and adds the same two things, a wall and doors that
know who knocked, with one difference: realms share live object references. A
realm may return a pointer to its own data, and a callee may receive an
argument that points into the caller's state. Solidity never hands storage
out, so it has no rule for this; Gno needs one.

## 3. Fences in an open yard

Go's packages are an open yard: any code walks anywhere. The interrealm model
puts a fence around each realm's state and one kind of gate in it, the
**crossing function**:

```go
func Transfer(cur realm, to address, n int64) { … } // first parameter `cur realm`
token.Transfer(cross(cur), bob, 10)                 // walking through the gate
```

Everyone may look over the fence. Only code that came in through the gate may
change what is inside.

## 4. The rules

**R1. Every object belongs to one realm:** the realm that declares its type, or
else the realm whose code was running when it was created. Libraries own no
changeable objects.

**R2. Reading needs no gate.** Any code may read any object it can reach.

**R3. A realm's state changes only during a crossing call into it.**

**R4. Inside that call, code acts for the realm if the realm's source names it:**
the realm's own functions and methods, and library functions and methods it
calls by name or through types from packages it imports. Code someone else
handed in — a function value, a closure, a method value, an implementation of
an interface from a package the realm does not import — acts for nobody and
can change nothing.

**R5. A realm's code reached any other way is read-only.** Called by another
realm, a run script or a library without crossing, it may read, build new
objects and change the temporaries it builds; it changes no stored object.

**R6. Identity comes only from crossing.** In a crossing function, `cur` is the
realm and `cur.Previous()` is whoever came through the gate. That is the one
check a realm needs.

A refused write says which gate to use:

```
gno.land/r/x's code was entered without crossing and is read-only (call its crossing function)
cannot modify object owned by another realm from library code (call a function of the owning realm)
```

## 5. Scenarios

Each scenario lists the library (`p`), the realm that owns the data (`victim`)
and the realm that calls it (`caller`).

**The token, done right.**
```go
// p: gno.land/p/demo/ledger
type Ledger struct{ bal map[address]int64 }
func (l *Ledger) Move(from, to address, n int64) { l.bal[from] -= n; l.bal[to] += n }
func (l *Ledger) Balance(a address) int64       { return l.bal[a] }

// victim: gno.land/r/demo/token
var l = ledger.New()
func Transfer(cur realm, to address, n int64) { l.Move(cur.Previous().Address(), to, n) }
func Ledger() *ledger.Ledger                  { return l }

// caller: gno.land/r/demo/thief
token.Transfer(cross(cur), bob, 10)         // moves the thief's own tokens (R3, R6)
token.Ledger().Balance(alice)               // reading is fine (R2)
token.Ledger().Move(alice, thief, 1e9)      // refused: library code on token's object (R1, R4)
```

**An exported helper is not a gate.**
```go
// victim
func SetOwner(a address) { owner = a }                      // no `cur`: not a gate
func ChangeOwner(cur realm, a address) {                    // a gate with a check
	if cur.Previous().Address() != owner { panic("unauthorized") }
	owner = a
}
// caller
x.SetOwner(attacker)              // refused: x's code entered without crossing (R5)
x.ChangeOwner(cross(cur), me)     // runs, and the check decides (R3, R6)
```

**A plugin cannot borrow the host's authority.**
```go
// p: gno.land/p/evil
type T struct{}
func (T) Run(o *x.Obj) { o.Field = "pwnd" }
// victim
type Runner interface{ Run(*Obj) }
func Use(cur realm, r Runner) { r.Run(gObj) }
// caller
x.Use(cross(cur), evil.T{})       // refused: x does not import p/evil (R4)
```

**A callback that must change its own realm is a crossing closure.**
```go
// callee realm
func ForEach(cur realm, f func(cur realm, v int)) { f(cross(cur), 1) }
func ForEach2(cur realm, f func(v int))           { f(1) }
// caller realm
var sum int
callee.ForEach(cross(cur), func(cur realm, v int) { sum += v }) // runs as the caller (R3)
callee.ForEach2(cross(cur), func(v int) { sum += v })           // a plain closure: read-only (R5)
```

**Handing objects out.**
```go
// victim
func NewVal() counter.Counter { return counter.Counter{} } // a copy: the caller's own
func NewPtr() *counter.Counter { return &counter.Counter{} } // x's object: a view
// caller
v := x.NewVal(); v.Inc()          // fine: the caller's copy (R1)
p := x.NewPtr(); p.Inc()          // refused: x's object (R1, R4)
```

## 6. Writing a realm

- Every function that changes state takes `cur realm` first and checks
  `cur.Previous()` if not everyone may call it.
- A crossing method runs as its realm for anyone who holds the object: check
  `cur.Previous()` there too.
- Return values for callers to keep; return pointers only as read-only views.
- A callback that must change its own realm is `func(cur realm, …)`, called
  with `f(cross(cur), …)`.
- To keep data in another realm's type, cross into that realm; otherwise use
  your own or a library type.
