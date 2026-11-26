# Interrealm: a whitepaper

How realms protect their state, built up from a few basic facts. Sections 1–3 say why a boundary is needed;
§4 states the basic facts; §5 the four rules; §6–7 the patterns, first one realm on its own, then two realms talking to each other;
§8 is a checklist and §9 the map to the code. Every example names the test that pins it. The teaching page is
[gno-interrealm.md](./gno-interrealm.md); the decision record is
[`gnovm/adr/pr6495_cross_to_write.md`](../../gnovm/adr/pr6495_cross_to_write.md).

## 1. A realm is a Go package with state

A realm (`gno.land/r/...`) is a Go package whose package-level variables persist between transactions. It also
has an address and can hold coins. A library (`gno.land/p/...`) is a Go package without state: after `init`
its variables never change.

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

That is fine inside one program with one author. On a chain, every package has a different author and some of
them are hostile. Solidity answers with a wall around each contract's storage and one kind of door:

```solidity
contract Token {
    mapping(address => uint) balances;              // nobody outside can touch it
    function transfer(address to, uint n) public {  // the door
        balances[msg.sender] -= n;                  // the door knows who knocked
        balances[to] += n;
    }
}
```

Gno keeps Go's packages and adds the same two things, a wall and doors that know who knocked, with one
difference: realms share live object references. A realm may return a pointer to its own data, and a callee
may receive an argument that points into the caller's state. Solidity never hands storage out, so it has no
rule for this; Gno needs one.

## 3. Fences in an open yard

Go's packages are an open yard: any code walks anywhere. The interrealm model puts a fence around each realm's
state and one kind of gate in it, the **crossing function**:

```go
func Transfer(cur realm, to address, n int64) { … } // first parameter `cur realm`
token.Transfer(cross(cur), bob, 10)                 // walking through the gate
```

Everyone may look over the fence. Only code that came in through the gate may change what is inside. The gate
does two jobs: it names the realm that may write, and it records who came through it, so the realm can
check the caller with `cur.Previous()`.

## 4. Basic facts

Five facts. Everything after this section follows from them plus the rules in §5.

**F1. Every stored object has one owner.** An object here is anything Go puts behind a pointer or a reference: a
struct or array you take the address of, a slice's backing array, a map, a closure. The owner is the realm whose
state the object was saved into. A library owns nothing it can change after `init`.

**F2. One realm is current.** A transaction starts with one realm current: the realm the user called, or the
run script's own realm. A **crossing call** `F(cross(cur))` into a **crossing function** `func F(cur realm, …)`
makes `F`'s realm current until `F` returns. Every other call (a plain call, a method call, a call through a
function value) leaves the current realm as it was.

**F3. Reading is free.** Any code may read any object it can reach.

**F4. A write needs two things.** A write to a stored object succeeds only if

1. the object's owner is the current realm, and
2. the code doing the write **acts for** the current realm.

R1 and R2 in §5 decide point 2: which code acts for the current realm.

**F5. Code that does not act for the current realm runs read-only.** It may read anything, make new objects and
change the new objects it made (they are its scratch). It changes no stored object. A refused write is a panic
and aborts the transaction unless something recovers it.

```go
// realm x, current; this code runs read-only
n := x.Count          // reading: fine (F3)
s := []int{1, 2}      // a new object
s[0] = 9              // fine: scratch (F5)
x.Count++             // refused (F4)
```

## 5. The rules

Four rules. The first two decide which code acts for the current realm: R1 for code called by its name, R2 for
code that arrives as a value. The last two decide what may leave a realm: R3 its values, R4 its ticket `cur`.
Each rule has one small example and the reason. §9 maps each rule to the code that enforces it and lists the
error texts.

### R1 Cross to write: code called by name

**A realm's own code acts for it only inside a crossing call into that realm.** Called any other way — by
another realm, by a run script, without `cross` — the realm's code runs read-only. **Library code has no realm of
its own: it acts for whichever realm is current.**

```go
// realm ctw
var Count int
func Add(cur realm, n int) { Count += n }   // a crossing function: a gate
func Bump()                { Count++ }      // exported, but not a gate
type Local struct{ N int }
func (l *Local) Set(cur realm, n int) { l.N = n } // a method is called by name too: a gate
func (l *Local) Inc()      { l.N++ }        // not a gate
func (l Local) Reset(cur realm) { l.N = 0; Count = 0 } // a gate with a value receiver
var Saved = &Local{N: 1}
var Kept  = Local{N: 1}     // a value, not a pointer

// another realm
ctw.Add(cross(cur), 2)        // ctw is current, ctw's code acts for it: writes
ctw.Bump()                    // the caller is current, ctw's code runs read-only: refused (master: writes)
ctw.Saved.Set(cross(cur), 5)  // the same for a method: writes
ctw.Saved.Inc()               // refused, whether the receiver is a pointer or a value
ctw.Kept.Set(cross(cur), 5)   // a pointer receiver on a value variable: writes Kept in place
ctw.Kept.Reset(cross(cur))    // a value receiver: Count = 0 writes, l.N = 0 changes only the copy
```

The call is decided by the method, not by the variable: a pointer receiver writes the variable whether it
holds a pointer or a value; a value receiver gets a copy, so writing it changes nothing saved, while the
method can still write the realm's other state and whatever the receiver points to.
`zrealm_wp_2.gno` shows `Kept.Set` and `Kept.Reset` on a value variable.

A plain call on master (its "declaring-realm borrow",
[gno-interrealm.md](https://github.com/gnolang/gno/blob/d25b308c1/docs/resources/gno-interrealm.md)):

```go
// realm ctw
func Bump() {
	Count++                              // master: saved to ctw
	println(runtime.CurrentRealm())      // master: the caller, not ctw; there is no cur to ask who called
}
var c *Config
func Conf() *Config { if c == nil { c = load() }; return c }   // fills a cache on first use

// another realm
ctw.Bump()    // master: writes, for anyone; today: refused, call ctw.Add(cross(cur), 1)
ctw.Conf()    // today: refused while c is nil (it writes c); fill c in init or in a gate
```

Library code (a `/p/` package) has no realm of its own, so its methods may write only objects the current realm
owns. A realm's exported library object is therefore readable by everyone and writable only from that realm's
own crossing calls:

```go
// realm r
var Tree avl.Tree                          // exported; Set and Get are library (/p/) methods
func Put(cur realm, k, v string) { Tree.Set(k, v) }   // r is current, owns Tree: writes

// another realm
r.Put(cross(cur), "k", "v")   // writes
r.Tree.Get("k")               // reading: fine
r.Tree.Set("k", "pwnd")       // the caller is current and does not own Tree: refused
```

On master that last line writes `r`'s tree: a library method on another realm's object runs with the owner's
rights. Master's security guide lists it as a known trap
([§5.1a](https://github.com/gnolang/gno/blob/d25b308c1/docs/resources/gno-security-guide.md#51a-p-type-with-unexported-fields-but-exported-mutation-methods)).

### R2 Code that arrives as a value

Code can also reach a realm as a value instead of a name: a **function value** (a closure
`func(s string){ … }`, a method value `obj.Set` not yet called, or a declared function used as a value
`f := SetOwner`) or an **interface value**, whose methods are chosen by whoever made it. Such a value can be
made in one call and run in another, made by one realm and run by another, or picked by the caller. So when it
runs, the question is: may it write? R2 answers in three parts: R2a by how it is called, R2b by whose hands
it passed, R2c by which package its code comes from.

#### R2a A plain function value writes in three places only

The idea: R1 makes the gate the only way in by name, but a function value can carry a realm's writer past the
gate. So a plain function value (one without a `cur realm` parameter) counts as the realm's own intent only
where the code that runs it can be read at the call line.

**A plain function writes its realm's stored state only when called by name by code that may already write, or
as a callback lent for that one call.** In code: a plain function value writes only (1) as the realm's own code
called by name, or as a local helper; (2) as a literal written in place as a call argument, while that call runs, lent only to the realm's
own functions or to a `/p/` library function reached by name, never to another realm's function; (3) as a literal
called in the expression that writes it. Reached by name means a direct call of a `/p/` function or method,
including a method called on an interface value that holds a `/p/` type the realm trusts (R2c), or a call through a
local set once to a declared function and only ever called (`each := lib.Each; each(xs, f)`). A method value is
not. Anything else runs read-only: a stored value, a returned one, a declared
function passed as a value, a method value, a literal the callee keeps and calls later.

```go
// realm r, inside a crossing call
f := func() { n++ }; f()                        // (1) local helper: writes
lib.Each(xs, func(x int) { n += x })            // (2) lent in place to a library called by name: writes
func() { n++ }()                                // (3) called where it is written: writes
Reg = append(Reg, func() { n++ }); Reg[0]()     // stored: refused, here and in every later call
```

The loan in (2) ends when the callee returns, stores the value or returns it, and it is never given to a
deferred or variadic argument. A callee parameter that is only called or passed on keeps it, down a chain of
such functions in the realm or a library, recursion included; passing it on converted to a named func type,
inside a lent literal, or to the realm's own interface method keeps it too. A function on the chain that stores
it, returns it, defers a call with it, or hands it to something that keeps it ends the loan. A callee reached as
a value (a reassigned local, a package variable) runs read-only, as does a new crossing call into the same realm
(`zrealm_lend_chain_ok.gno`, `zrealm_lend_chain_refused.gno`, `zrealm_lend_chain_crossing.gno`,
`zrealm_lend_nonesc_earlykeep.gno`). The marks are
decided when the code is prepared, so nothing about them is stored. Tests: `zrealm_lend_each.gno`,
`zrealm_lend_helper.gno`, `zrealm_lend_store.gno`, `zrealm_lend_values.gno`; `zrealm_wp_1.gno` (a method called
on an interface, a deferred call, a variadic slot).

```go
// realm cbvictim
func SetOwner(s string) { Owner = s }                 // a writer, meant only for cbvictim's own use
func GetDecl() func(string) { return SetOwner }       // hands it out as a value
func RunS(cur realm, f func(string)) { f("picked") }  // a gate that runs a callback it is given

// attacker: calling SetOwner by name is refused (R1), so it gets cbvictim's own gate to run it
cbvictim.RunS(cross(cur), cbvictim.GetDecl())   // a declared function as a value: refused (master: writes)
```

What changes for callback authors: a plain closure handed to another realm runs read-only there. To write, it
must take `cur realm` and be called with `cross`, which enters its realm; this is the default for callbacks:

```go
// realm ctw
func ForEach(cur realm, f func(n int))             { f(1) }
func ForEachX(cur realm, f func(cur realm, n int)) { f(cross(cur), 1) }

// realm b, which keeps a total in its own variable sum
ctw.ForEach(cross(cur), func(n int) { sum += n })              // another realm's function: no loan, refused (master: writes)
ctw.ForEachX(cross(cur), func(cur realm, n int) { sum += n })  // f enters b: writes
```

Known limits: a callee not reached by name (a reassigned local, a package variable, a method expression copied to
a local, or a literal kept in a variable) lends nothing; a method value never writes; `lib.Default.Each(func(int){ count++ })` is refused, since a method of
a variable the library owns runs as the library; and a local helper used inside a literal lent in place is refused
(call it outside the literal, or write its body in the literal). Tests: `zrealm_ctw_owncb.gno`,
`zrealm_lend_lib_shapes.gno`, `zrealm_lend_reg_plibeach.gno`, `zrealm_wp_1.gno`.

**The save check (R2a-save).** Read-only code may still fill in an object made during the call, since that is
how a getter builds its result. The gap: a read-only function value fills in the realm's new object, and the
realm then saves it. The write never touched saved state, yet it ends up there:

```go
// realm s3urv
var W = func(e *Entry) { e.Owner = "stamped" }   // stored function value: runs read-only
func AddRelay(cur realm, f func(*Entry)) {        // a gate that runs a hook on each new entry
	e := &Entry{Owner: "victim"}                  // new, not saved yet
	entries = append(entries, e)                  // e will be saved when the call ends
	f(e)                                          // f is read-only, but e is new: the write goes through
}

// attacker
s3urv.AddRelay(cross(cur), s3urv.W)   // refused when the call ends, because e would be saved (master: saves "stamped")
```

So the write is allowed and marked, and the save refuses a marked object. The check cannot come at the write:
while the call runs, the VM cannot tell whether `e` will be saved or is only a temporary value. (The
attacker's own closure is refused earlier, at the write: its realm was not entered, R1.)

#### R2b A function value from another realm runs read-only

**A plain function value that passes through another realm's hands — returned by it, stored in its variable,
struct or slice, or handed back by it — runs read-only.** Crossing function values (`func(cur realm, …)`) are
not touched: they enter their realm with `cross`.

```go
// realm x
func Greet(cur realm) {
	f := func(s string) { Owner = s }   // x's writer
	fmtr := y.Formatter(f)              // x asks realm y for a formatter, passing f along
	fmtr("hello")                       // y handed back f itself: refused (master: writes)
}
```

Why: whatever comes back from y may be x's own writer, swapped in where x expected something else; x cannot
tell. (Test: `zrealm_fv_echo_same_call.gno`.)

#### R2c Only imported library code writes

The problem first. By R1, library code acts for whichever realm is current. Inside realm `x`'s crossing call,
`x` is current, so any library code running there may change `x`'s state. When `x` calls a library by name,
`x` chose that code:

```go
// realm x
import "gno.land/p/nt/avl"
func Add(cur realm, k string) { tree.Set(k, 1) }   // x picked avl.Tree.Set by importing avl: writes
```

But Go has two ways for the *caller* to choose the code instead: an interface value and a function value. A
function value, a method value included, follows R2a and never writes outside its three places, whatever its
package (`zrealm_wp_3`). For an interface value, `x` writes `r.Run(obj)`, and which `Run` runs depends on what
the caller passed in:

```go
// library p/evil, written by the attacker
type T struct{}
func (T) Run(o *Obj) { o.Owner = "attacker" }

// realm x, which never imports p/evil
type Runner interface{ Run(o *Obj) }
func Use(cur realm, r Runner) { r.Run(myObj) }   // x is current: without R2c, evil's Run would write myObj

// the attacker's realm
x.Use(cross(cur), evil.T{})   // refused (master: writes)
```

**So: library code that `x` reaches through an interface acts for `x` only if `x`, or the package whose code
makes the call, imports that library directly.** Otherwise it runs read-only. Above,
`x` never imports `p/evil`, so `evil.T.Run` runs read-only and the write to `myObj` is refused.

Importing a package is how a realm says "I trust this code". A call by name always passes, because a realm can
only name packages it imports; the rule only matters when the caller picks the code. A standard-library package
is trusted anywhere in the realm's import tree; a method promoted from an embedded type counts as the embedded
type's package; a realm's test files do not widen the trust.

Test: `zrealm_entry_iface_val.gno`; `zrealm_trust_direct_*.gno` (direct imports, the calling package, promoted
methods, the standard library).

### R3 Types stay home: only pointers cross

**A struct or array value of a type a realm declares is made and copied only by that realm's code.** Other
realms hold a pointer to it.

R1 and R2 decide which code may write a realm's state. R3 is a separate question: which values of a realm's
types may exist. A value of `T` made or copied outside `T`'s realm is a forgery: it looks like the realm's
own and has the same fields, but the realm never made it and does not know it exists. A realm's code often
trusts a `*T` it receives because only its own code makes a `T`; R3 keeps that trust true, since every `T`
outside its realm is reached through a pointer the realm handed out.

What another realm may still do is hold the pointer and use it:

```go
p := fvx.New(1)   // fine: a pointer to fvx's value
p.Val()           // fine: call its methods through the pointer (R1 decides if they write)
```

Everything else that makes or copies an `fvx.T` outside `fvx` is refused:

```go
c := *p; fvx.Read(cross(cur), &c)  // a copy handed back as fvx's own (master: fvx reads it as its own)
c := *p; c.N = 999                 // a copy changed (master: copies; refused only at the write)
var t fvx.T; t.N = 5               // a new fvx.T made outside fvx (master: makes it; refused only at the write)
f := p.Val; f()                    // a method value keeps a copy; so does `defer p.Val()` (master: runs)
err := fvx.Fail()                  // a struct error copied out of fvx (master: works)
q := (*fvx.T)(&struct{ N int }{5}) // a forged pointer (master: refused too, "illegal conversion")
```

Master already refuses most writes to such values; R3 refuses the forgery itself, so a value the owner did
not make never exists outside it (§7.4 has more cases).

In practice: return `*T`, return errors as pointers (`&E{…}`), and defer a closure rather than a value-receiver
method. A struct with no type name (`struct{ N int }`) belongs to no realm, so it may be passed by value when
all its fields are exported.

### R4 The ticket stays in its call

`cur` is a ticket: code holding it can cross as this realm and spend its coins. **`cur` lives only in the call
that received it: it cannot be stored, replaced, have its address taken, or be faked; coins and sub-tickets come
only from the live `cur`.** A function literal whose result type holds a realm value is refused when the package
is deployed (`zrealm_cur_escape_lit*.gno`).

```go
stashed = cur                                    // refused at save: cannot persist realm value
cur = other                                      // refused: cannot reassign the crossing `cur` parameter
var _ realm = fooRealm{}                         // refused: only the runtime's type is a realm
banker.NewBanker(banker.BankerTypeRealmSend, cur.Previous())   // refused: not the current realm
```

Why: a stored or forged ticket would let code cross as this realm long after, or without, a real call.

Tests: §6.2 (store, replace, fake), §6.5 (sub-tickets), §6.6 (banker).

## 6. Patterns: one realm on its own

§5 says why each rule exists. §6–7 say what to write: each pattern is one task, the code for it, the rule that
decides it and the tests that pin it. Marks in the code are the same as in §5: `(master: …)` is what master
(d25b308c1) does where it differs from this branch, and "refused" means the write panics and aborts the
transaction unless something recovers it.

### 6.1 Check who called: `cur.Previous()`

In a crossing function, `cur` is this realm and `cur.Previous()` is whoever crossed in: a user, a run script, or
another realm. A chain of crossing calls is a chain of `Previous()` values; past the user, `Previous()` panics.

```go
func A(cur realm) {                 // realm curchain_a, called by curchaintest
	cur.PkgPath()                   // gno.land/r/tests/vm/curchain_a
	cur.Previous().PkgPath()        // gno.land/r/curchaintest
	curchain_b.B(cross(cur))        // in B: Previous() is curchain_a, then curchaintest, then "" (the user)
}
```

To accept a payment sent with the transaction, check `cur.Previous().IsUserCall()`, not `IsUser()`:
`IsUser()` also accepts a run script, which could spend the payment before calling you.

```go
prev.IsUserCall()   // true for the user calling main
cur.IsUserCall()    // false: cur is a code realm
```

Tests: `zrealm_cur_chain.gno` (Output `B Previous().Previous().PkgPath: gno.land/r/curchaintest`,
`B Previous()^4 panics past origin: true`); `zrealm_cur_classify.gno` (`prev.IsUserCall: true`,
`cur.IsUserCall: false`).

### 6.2 Use `cur` only in the call that received it

`cur` is the realm's ticket (R4). Use it in the call; do not keep it, and never hand it to code you did not
write. Each way of keeping it is refused:

```go
var stashed realm
stashed = cur                             // refused at save: cannot persist realm value
H = &Holder{R: cur}                       // the same, inside a struct
fn = func() string { return c.PkgPath() } // the same, inside a closure
p := &cur                                 // refused: cannot take the address of a realm-typed `cur`
cur = other                               // refused: cannot reassign the crossing `cur` parameter
```

Nor can anyone make a fake ticket: only the runtime's own type satisfies `realm`.

```go
type fooRealm struct{ realm; forged address }
var _ realm = fooRealm{}   // refused: does not implement realm (missing method .seal)
```

Rule: R4. Tests: `zrealm_cur_persist_var.gno`, `_struct`, `_closure` (Error `cannot persist realm value: realm
values are ephemeral and tied to a call frame`); `zrealm_cur_addr.gno`, `zrealm_cur_reassign.gno`,
`zrealm_cur_range.gno`;
`zrealm_seal_embed.gno` (Error `…fooRealm does not implement .uverse.realm (missing method .seal)`),
`zrealm_seal_embed_runtime.gno` (Output `assertion ok: false`).

### 6.3 Set up state at deploy: `init(cur realm)`

`init` may take `cur`. It is a gate like any crossing function: it runs at deploy with the realm current and the
deployer as `Previous()`, so it may write and cross out.

```go
func init(cur realm) {
	initCur  = cur.PkgPath()             // gno.land/r/curinit
	initPrev = cur.Previous().PkgPath()  // "": the deployer
}
```

Rule: R1. Tests: `zrealm_cur_init.gno` (Output `init cur.PkgPath: gno.land/r/curinit`, `init prev.PkgPath empty
(EOA): true`); `zrealm_crossrealm15.gno` (`init` crossing out); `zrealm_initctx.gno`.

### 6.4 The ways into a realm: MsgCall, run scripts, `Render`

- **MsgCall** calls a crossing function as a crossing call from the user.
- **A run script** (MsgRun) is its own realm, `gno.land/e/<addr>/run`. Its `main` changes another realm's state
  only by crossing, like any realm.
- **`Render(cur realm, path string)`**: a render query passes `cur`, so `Render` may cross into its own
  helpers. The older `Render(path string)` stays read-only.

```go
borrowrealm.Mutate()                     // from a run script or a MsgCall: refused, entered without crossing
borrowrealm.MutateCrossing(cross(cur))   // writes
```

Rule: R1. Tests: `interrealm_mix_run.txtar`, `interrealm_mix_call.txtar`, `render_crossing.txtar`
(in `gno.land/pkg/integration/testdata/`).

### 6.5 Cross as a part of the realm: `cur.Sub`

`cur.Sub("dao/1")` makes a sub-ticket: a realm value named `<path>#dao/1` with its own address. It crosses like
`cur`, and the callee sees it as `cur.Previous()`. Only the live `cur` can make one; a sub-ticket cannot make
another, be stored, or spend coins of its own.

```go
sub := cur.Sub("dao/1")
callee(cross(sub))   // in callee: cur.Previous().PkgPath() == "gno.land/r/demo/subrealm#dao/1"
```

Rule: R4. Tests: `zrealm_sub1.gno` (Output `prev.PkgPath: gno.land/r/demo/subrealm#dao/1`);
`zrealm_sub_nested.gno`, `zrealm_sub_stale_cross.gno`, `zrealm_sub_persist.gno`, `zrealm_sub_banker.gno`.

### 6.6 Spend the realm's coins: a banker over `cur`

Build the banker over the live `cur`; then only the realm's own crossing call can spend its coins.

```go
banker.NewBanker(banker.BankerTypeRealmSend, cur)              // fine
banker.NewBanker(banker.BankerTypeRealmSend, cur.Previous())   // refused
```

Rule: R4. Test: `zrealm_banker_previous.gno` (Error `banker can only be instantiated for the current realm`).

### 6.7 A library's variables never change after `init`

A library has no state (§1): its package variables are fixed after `init`, whoever calls.

```go
p := tests.SomeValue3   // a library's package variable
p.Modify()              // refused: package is immutable post-init
```

Rule: F1. Tests: `zrealm_readonly_recv_method.gno` (Error `cannot mutate gno.land/p/demo/tests/v0: package is
immutable post-init`); `object_pointer_pure.txtar`.

## 7. Patterns: two realms

Each pattern is one way two realms meet. The code is labelled by package: `realm a` owns the state, `realm b`
is the caller, `p:` is a library.

### 7.1 Let other realms change your state: write a gate

Only a gate writes. An exported function without `cur realm` is readable by anyone and writes for no one.

```go
// realm a: gno.land/r/tests/vm/ctw
var Count int
var Saved = &Local{N: 1}
type Local struct{ N int }
func (l *Local) Inc() { l.N++ }
func Add(cur realm, n int) { Count += n; bump(); Saved.Inc() } // a gate: three writes
func Bump() { Count++ }                                        // exported, not a gate
func bump() { Count++ }                                        // a helper called from Add

// realm b
ctw.Add(cross(cur), 2)   // Count 0 -> 3, Saved.N 1 -> 2
ctw.Bump()               // refused: "gno.land/r/tests/vm/ctw's code was entered without crossing …" (master: writes)
```

Inside the gate, everything `ctw`'s code calls by name, its helper and its own method, acts for `ctw`.

The gate is narrow on purpose: `cross` is only ever a call's first argument; a crossing call cannot be deferred
directly, so wrap it in a closure; the old `cross1` and `mutable` builtins are gone.

Rule: R1. Tests: `zrealm_ctw_cross.gno` (Output `3 2`), `zrealm_ctw_bump.gno`; `zrealm_cross_defer.gno`,
`zrealm_cross_deferwrap.gno`, `zrealm_cross1_removed.gno`, `zrealm_entry_no_mutable.gno`.

### 7.2 Let holders of your object change it: a crossing method

Many realms can hold the same object. A crossing method is a gate into the object's owner, for anyone holding
the object; a method that is not a gate runs read-only when another realm calls it.

```go
// realm a: gno.land/r/crossrealm (owns f)
func (f *fooer) Foo(cur realm) { f.s = "Foo" }   // a crossing method: a gate
func (f *fooer) Bar()          { f.s = "Bar" }   // not a gate

// realm b holds f as an interface
f.Foo(cross(cur))   // writes, as a
f.Bar()             // refused: "…crossrealm's code was entered without crossing …: f<VPBlock(1,0)>.s" (master: writes)
```

Since anyone holding the object can call it, a crossing method checks `cur.Previous()` like any gate (§6.1).

Rule: R1. Test: `zrealm_crossrealm20.gno`.

### 7.3 Hand out a pointer for reading

A pointer to your object lets other realms read it, not change it, even through a library method. The same
library method writes inside your gate and is refused outside it: what differs is which realm is current.

```go
// p: modelledger
func (l *Ledger) Move(from, to address, n int64) { l.bal[from] -= n; l.bal[to] += n }
func (l *Ledger) Balance(a address) int64       { return l.bal[a] }

// realm a: modeltoken
var l = modelledger.New()
func Ledger() *modelledger.Ledger { return l }
func Transfer(cur realm, to address, n int64) { l.Move(cur.Previous().Address(), to, n) }

// realm b: thief
modeltoken.Transfer(cross(cur), bob, 10)               // the thief's own tokens: writes
modeltoken.Ledger().Balance(bob)                       // reading: 10
modeltoken.Ledger().Move(alice, "g1thief", 1000000000) // refused: "cannot modify object owned by another realm from library code …" (master: writes)
```

Rule: R1. Test: `zrealm_model_token.gno` (Output `bob: 10`, then the Error).

### 7.4 Return pointers, not values

Hand out `*T`, and errors as `&E{…}`. Another realm cannot make or copy a value of your type (R3), so a value
you return to it is refused:

```go
// realm a: fvx
type T struct{ N int }
type E struct{ Msg string }
func (e E) Error() string { return e.Msg }
func New(n int) *T  { return &T{n} }
func Fail() error   { return E{"x"} }    // an E value boxed in an error: a copy

// realm b
p := fvx.New(1)
c := *p                  // refused: a copy of fvx.T outside fvx
c.N = 999                //   (master: copies, then refuses this write: "readonly tainted object")
fvx.Read(cross(cur), &c)
err := fvx.Fail()        // refused: a copy of fvx.E; return &E{…} instead
defer p.Val()            // refused: defer keeps a copy of *p; write defer func() { p.Val() }()
n := p.Val()             // fine: a direct call copies only for the call
```

Master refuses changing the copy, but an unchanged copy handed back is still a value the owner never made and
trusts as its own. A struct with no type name (`struct{ N int }`) belongs to no realm and passes by value.

Rule: R3. Tests: `zrealm_attack_copy_handback.gno`, `zrealm_fv_copy_assign.gno`, `zrealm_fv_err_define.gno`,
`zrealm_fv_defer_val.gno`, `zrealm_fv_forge_ro.gno`; `zrealm_fv_anon_arg.gno` (the unnamed struct passes).

### 7.5 Copy a library's value: the copy is yours

A library's value is not a realm's, so R3 does not apply to it: a copy of it belongs to the realm that made the
copy, and changing the copy changes nothing shared.

```go
stolen := launderpkg.PInitData        // a copy, owned by this realm
stolen.Field = "pwnd"                 // writes the copy only
```

Writing *through* a pointer into the library's variable is refused (§6.7).

Rule: R3, F1. Tests: `zrealm_launder_pdata_value_copy.gno`, `zrealm_launder_pdata_ptr_deref.gno`.

### 7.6 A pointer outlives the object it points to

Deleting an object does not take back the pointers others hold. A copy taken after the delete still cannot make
the owner write: the function value inside it is a stored value, so it runs read-only (R2a, R2b).

```go
p := rv.GetP()               // a pointer for reading (§7.3)
rv.Del(cross(cur))           // rv deletes the object
o := *p                      // the copy
rv.RunOuter(cross(cur), o)   // o.Inner.F runs read-only: refused (master: writes Owner)
```

Rule: R3, R2a. Test: `home_deleted_ref.txtar` (qeval `Owner` stays `victim`).

### 7.7 Use a library on your own state

Library code acts for the current realm (R1). Inside realm `a`'s gate, a library that calls back into `a`'s
closures, walks `a`'s objects or uses its own tables writes `a`'s state:

```go
// realm a: ctw
func ViaTable(cur realm) {
	libtable.Table["run"](
		func() { Count++ },
	)
}
func ViaDispatch(cur realm) { libtable.Dispatch("inc", Boxes[0]) }

// realm b
ctw.ViaTable(cross(cur))       // Count +1
ctw.ViaDispatch(cross(cur))    // Boxes[0].N +1
```

`ViaApply`, which passes its literal to `libtable.Apply`, a literal kept in a library variable, is refused: a
callee kept in a variable lends nothing (§7.14). `ViaAlias` writes: `each := libtable.Each` is set once to a
declared function and only called, so it is that function by name (`zrealm_lend_alias_ok.gno`).

The same library code, called by realm `b` on `a`'s objects, is refused (§7.3). Library code that a caller
passes in through an interface is decided by R2c (§7.10).

Rule: R1, R2a (a literal lent in place to a library function called by name). Tests: `zrealm_ctw_owncb.gno`;
`zrealm_launder_d_pattack_direct.gno` (a library function writing another realm's object: refused).

### 7.8 A callback that writes its own realm: take `cur realm`

A plain closure handed to another realm gets no loan, so it runs read-only (R2a). A
crossing closure, `func(cur realm, …)` called with `cross`, enters its own realm like any gate and writes.

```go
// realm a: ctw
func ForEach(cur realm, f func(n int))             { f(1) }
func ForEachX(cur realm, f func(cur realm, n int)) { f(cross(cur), 1) }

// realm b
var sum int
ctw.ForEach(cross(cur), func(n int) { sum += n })              // refused: "gno.land/r/ctw_callback's code was entered without crossing …" (master: writes)
ctw.ForEachX(cross(cur), func(cur realm, n int) { sum += n })  // writes: sum 1
```

Inside the crossing closure, `cur.Previous()` is `ctw`. A real library does the same: `p/nt/mux` stores
realm-aware handlers as `func(cur realm, res, req)` and calls them with `cross(rlm)` (`router.gno:127`); GovDAO's
`Render` uses it.

Rule: R2a. Tests: `zrealm_ctw_callback.gno`, `zrealm_ctw_callbackx.gno` (Output `1`); `render_crossing.txtar`.

### 7.9 Function values kept for later

A plain function value a realm stores, returns or hands out runs read-only wherever it is later called (R2a),
and a value that passed through another realm runs read-only too (R2b). Seen from the owner's side:

```go
// realm a: cbvictim
var Owner = "victim"
func SetOwner(s string) { Owner = s }
func GetDecl() func(string) { return SetOwner }               // a declared function as a value
func RunS(cur realm, f func(string)) { f("picked") }

// realm b
cbvictim.RunS(cross(cur), cbvictim.GetDecl())   // refused: "…cbvictim's code is read-only here: …" (master: writes)
```

What still writes:

```go
func Bump(cur realm) { f := func() { n++ }; f() }                 // a local helper
func Sum(cur realm)  { lib.Each(xs, func(x int) { n += x }) }     // a literal lent in place
```

A declared function passed as a value is refused even to a library called by name: pass a literal that calls
it instead (`zrealm_maker_ownfunc.gno`: `declared` refused, `literal` writes). A function value a realm keeps
for later is a crossing closure (§7.8).

Rule: R2a, R2b. Tests: `zrealm_maker_getdecl.gno`, `zrealm_maker_ownfunc.gno`, `zrealm_home_gate_result.gno`,
`zrealm_saved_cross_helper.gno`, `zrealm_saved_same_tx.gno`, `zrealm_lend_store.gno`, `zrealm_lend_each.gno`,
`zrealm_lend_helper.gno`, `zrealm_lend_values.gno`; R2a-save: `zrealm_s3u_rv2_1.gno`.

### 7.10 Accept interfaces and plugins

A realm that calls a method on a caller-supplied interface runs code the caller chose. That code writes only if
the realm imports its package (R2c).

```go
// p: ifaceattack (the caller's)
func (EvilV) Run(x *launderpkg.Immutable) { x.Field = "pwnd-via-iface" }

// realm a: pfuncvictim (never imports ifaceattack)
func Use(cur realm, r Runner) { r.Run(gObj) }

// realm b
pfuncvictim.Use(cross(cur), ifaceattack.EvilV{})   // refused: "cannot modify object owned by another realm from library code …" (master: writes)
```

The standard library is trusted anywhere in the realm's import tree: a realm importing `bytes` accepts a
`bytes.Reader` writing its buffer, while a realm whose tree holds only `io` refuses the same reader.

```go
control.Fill(cross(cur), bytes.NewReader([]byte("efgh")))     // control imports bytes: writes
impvictim.Fill(cross(cur), bytes.NewReader([]byte("abcd")))   // impvictim imports io only: refused
```

Still, type-assert caller-supplied interface values to the types you expect. A method value taken from such a
value, `f := r.Run; f(obj)`, never writes, even from an imported package (R2a; `zrealm_entry_iface_mv`,
`zrealm_wp_3`).

Rule: R2c. Tests: `zrealm_entry_iface_val.gno`, `zrealm_entry_iface_ptr.gno`, `zrealm_entry_iface_gen.gno`,
`zrealm_import_std_out.gno` (Output `control: efgh`).

### 7.11 Let another realm spend for you: hand it a crossing closure

Hand over a crossing closure that spends, not the sender object: the closure enters your realm, while the object
handed over is only readable to the other realm (§7.3).

```go
// realm g: granter
al := allowancesender.New(banker.NewBanker(banker.BankerTypeRealmSend, cur), …, cap)
defer al.Close()
callee.StoreFn(cross(cur), func(cur realm, to address, amt chain.Coins) { al.Send(to, amt) })

// realm c: callee
fn(cross(cur), cur.Address(), coins)   // spends within the cap until g closes it
al.Send(…)                             // on the object g passed in: refused, cannot modify
```

Rule: R1. Test: `allowancesender_handoff.txtar` (stderr `closed`, `cap exceeded`, `cannot modify`).

### 7.12 Events belong to the realm that emits them

An event belongs to the realm whose code emits it, including a closure that realm runs in its own crossing call.

Rule: R1. Test: `event_callback.txtar` (both events `"pkg_path":"gno.land/r/demo/cbee"`).

### 7.13 Recover a refused write

A refused write is a panic. Running read-only does not cross into another realm, so the caller may recover the
panic and keep writing its own state.

```go
func() {
	defer func() { println("recovered:", recover()) }()
	ctw.Boom()           // recovered: boom
}()
Count++                  // still writes: count: 1
```

Rule: R1. Tests: `zrealm_ctw_recover.gno` (Output `recovered: boom`, `count: 1`);
`zrealm_launder_rdata_pdr_panic_defer.gno`.

### 7.14 Known limits

- A fresh function a realm stores and then runs in the same call is read-only; run it first, or make it a
  crossing function (`zrealm_saved_same_tx`, `zrealm_fv_wrap_ptr`).
- A library table entry called in place (the entry `lib.Table[k]` called directly) acts for the current realm. A
  table entry or a literal copied to a local is a value and runs read-only (`zrealm_ctw_owncb`'s `ViaTable`,
  `zrealm_saved_libtable_copy`); only a local set once to a declared function is that function by name (§7.7).
- A library imported directly, by the realm or by the calling package, is trusted for interface calls (R2c); a
  method value still never writes (R2a, `zrealm_wp_3`).
- A literal is lent only to a callee reached by name: through a reassigned local, a package variable or a function
  kept in a variable it runs read-only (`zrealm_ctw_owncb`, `zrealm_lend_alias_refused`); another realm's function
  gets no loan (`zrealm_lend_other_plain`); a method value never writes (`zrealm_lend_lib_shapes`).
- A literal passed to a method of a variable a library owns, `lib.Default.Each(func(int) { n++ })`, is refused:
  the method runs as the library (`zrealm_lend_reg_plibeach`). Call a library function by name, or pass a
  crossing closure.
- A local helper used inside a lent literal is refused; call it outside the literal or write its body there
  (`zrealm_wp_1`).
- A lent literal captured, at any depth, by a function value that is stored, returned or kept loses its loan at
  once, also when that wrapper was itself lent (`zrealm_lend_nested_two`, `zrealm_lend_captures`).
- Replaying past blocks with this binary computes other results (a function value read back from saved state
  no longer writes; gas differs), so it ships on a new chain or behind a gated upgrade (ADR, "Upgrade notes").

Tests: `zrealm_home_known_relay_table.gno`, `zrealm_home_known_ro_relay.gno`, `zrealm_home_ro_relay.gno`,
`zrealm_s3u_rv3_2.gno`, `zrealm_s3u_rv4_1.gno`.

## 8. Writing a realm

- Every function that changes state takes `cur realm` first and checks `cur.Previous()` if not everyone may
  call it (R1).
- Exported plain functions are for reading, constructors and helpers; never for writes (R1).
- Return pointers, and errors as pointers; copy only your own values (R3).
- A callback that must change its own realm is `func(cur realm, …)`, called with `f(cross(cur), …)`; a stored
  or returned function value is read-only (R2a, R2b).
- A plain function writes only when called by name by code that may already write, or as a callback lent for
  that one call; do not count on it elsewhere.
- Reach code through packages you import; type-assert caller-supplied interface values (R2c).
- Never hand `cur` to code you did not write, and never store it (R4, §6.2).
- A crossing method runs as its realm for anyone holding the object: check the caller there too (§7.2).
- To keep data in another realm's type, cross into that realm; otherwise use your own or a library type.

## 9. Where it lives (for reviewers)

The table maps each rule to the code that enforces it and the filetest family that pins it. The count command
for every row is printed under the table.

| Rule | Enforcing code | Fixture family | Files |
|---|---|---|---|
| R1 Cross to write | `gnovm/pkg/gnolang/machine.go:readonlyAccessPanic`, `isReadonly` | `zrealm_ctw_*` (14) + `zrealm_crossrealm*` (56) | 70 |
| R2a A plain function value writes in three places only | `gnovm/pkg/gnolang/crossingfn.go:calledByValue`, `callstart.go:byValueReadOnly`; `birthrule.go:markLocalHelpers` (local helpers); `lend.go:markNonEscParams`, `lendPlace`, `lendArgs`, `lendLive` (the loan) | `zrealm_lend_*` (125) + `zrealm_saved_*` (16) | 141 |
| R2b A function value from another realm runs read-only | `lend.go:lendLive` (a loan holds only in its own realm, uncrossed) and R1's read-only entry | `zrealm_maker_*` (15) + `zrealm_saved_*` (16) | 31 |
| R2c Only imported library code writes | `gnovm/pkg/gnolang/machine.go:foreignLib`, `trustRoot` | `zrealm_entry_iface*` (27, includes `zrealm_entry_iface.gno`) + `zrealm_import_*` (6) | 33 |
| R2a-save The refusal at save | `gnovm/pkg/gnolang/machine.go:unstoredWrite`, `markViewWrite`; `gnovm/pkg/gnolang/realm.go:assertNoUncrossedWrite` | `zrealm_s3u_*` | 38 |
| R3 Types stay home: only pointers cross | `gnovm/pkg/gnolang/alloc.go:checkCopy`, `panicForeignAlloc`; `gnovm/pkg/gnolang/realmvalue.go:checkRealmValues` | `zrealm_fv_*` (71) + `zrealm_value_ban*` (2) | 73 |
| R4 The ticket stays in its call | `gnovm/pkg/gnolang/uverse.go:refusePersistRealmHIV`; `preprocess.go` (address of `cur`, reassign); `gnovm/stdlibs/chain/banker/banker.gno` | `zrealm_cur_*` (47) + `zrealm_seal*` (6) + `zrealm_sub*` (12) | 65 |
| Laundering attempts | the same checks, from the offender side | `zrealm_launder_*` | 83 |

```
ls gnovm/tests/files/zrealm_ctw_* | wc -l          # 14
ls gnovm/tests/files/zrealm_crossrealm* | wc -l    # 56
ls gnovm/tests/files/zrealm_fv_* | wc -l           # 71
ls gnovm/tests/files/zrealm_value_ban* | wc -l     # 2
ls gnovm/tests/files/zrealm_maker_* | wc -l        # 15
ls gnovm/tests/files/zrealm_home_* | wc -l         # 27
ls gnovm/tests/files/zrealm_entry_pfunc_* | wc -l  # 14
ls gnovm/tests/files/zrealm_entry_iface* | wc -l   # 27 (includes zrealm_entry_iface.gno)
ls gnovm/tests/files/zrealm_import_* | wc -l       # 6
ls gnovm/tests/files/zrealm_cur_* | wc -l          # 47
ls gnovm/tests/files/zrealm_seal* | wc -l          # 6
ls gnovm/tests/files/zrealm_sub* | wc -l           # 12
ls gnovm/tests/files/zrealm_launder_* | wc -l      # 83
ls gnovm/tests/files/zrealm_saved_* | wc -l        # 16
ls gnovm/tests/files/zrealm_s3u_* | wc -l          # 38
ls gnovm/tests/files/zrealm_lend_* | wc -l         # 125
ls gnovm/tests/files/zrealm_* | wc -l              # 713
```

R2a's check and its save check are both enforced here. The counts
above are what `ls` prints, so a reviewer can re-run them.

### Error texts

| Rule | Layer | Text |
|---|---|---|
| R1 | run time | E2 `<path>'s code was entered without crossing and is read-only (call its crossing function): expr` |
| R1 | run time | E1 `cannot modify object owned by another realm from library code (call a crossing function of the owning realm): expr` |
| R1 | run time | E3 `cannot directly modify readonly tainted object (write it in a named function of the realm that owns it, during a crossing call into that realm): expr` |
| R3 | type check | `cannot hold a value of <T> outside its realm (use a pointer)` (`realmvalue.go`), e.g. `c := *p` |
| R3 | run time | `cannot copy <T> outside realm <R> (use a pointer)` (`alloc.go:checkCopy`), the only refusal for `err := fvx.Fail()` and `defer p.Val()` |
| R3 | run time | `cannot allocate <T> in realm <R>` (`alloc.go:panicForeignAlloc`) |
| R2a, R2b | run time | E4 `<path>'s code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): expr` |
| R2a-save | save | `<realm>: an object written by read-only code would be stored (write it inside a crossing call of <realm>)` |

R2a's marks are decided when the code is prepared: `ATTR_LOCAL_HELPER` on a local helper's calls,
`ATTR_NAMED_CALLEE` on a call by name, `ATTR_NONESC_PARAMS` on a function whose parameter is only called or
passed on. The loan is held in memory for the length of the call (`Machine.lends`, a token per frame) and ends
on return, on escape, and in a variadic slice; nothing is stored, so sizes and encoding are unchanged. Following
captures below the first level, and walking a variadic slice while a loan is live, is charged CPU gas
(`walkCaptures`, `dropSliceLends`).
R2a-save's mark is `ObjectInfo.s3Written`; `markViewWrite` marks a written unreal heap item the same way.

### The write check, step by step

A write is an assignment, or a builtin that mutates a slice or map. The guard is `isReadonly`
(`gnovm/pkg/gnolang/machine.go`), called by every mutation site: assignments through `PopAsPointer2`,
and the builtins `append` (`uverse.go`), `copy` and `delete`.

`isReadonly` decides each write, in this order:

1. **R2a first.** A non-crossing function value reached by value, neither a local helper nor holding a live
   loan, enters read-only (`enterReadOnlyByS3`, `machine.go`): `viewOnly` and `viewS3` set, `viewRealm` is the
   realm current at the call, `m.Realm` cleared.
2. **Allowed**, first match wins: single-user mode, `m.Realm == nil` and not view-only; scratch
   of this entry; a `RefValue` naming the writing package; R2a-save, the read-only realm's
   own unsaved object, only when `viewS3`; then `IsReadonlyBy`: the current realm's objects
   and the executing stdlib package's own stamp. An unreal heap item is marked first.
3. **Refused**, text chosen by the writing code: a builtin gives `cannot append to / copy to / delete from
   readonly tainted …` (`uverse.go`); `/p/` code gives E1; an R2a stretch gives
   E4; a realm entered read-only gives E2; anything else gives E3.
4. **Marked, refused at save**: an R2a-save write, or after any read-only entry a write to an unreal heap item it did
   not make (`markViewWrite`); `assertNoUncrossedWrite` (`realm.go`) refuses the save.

A name is resolved first by `isExternalRealm`, which exempts scratch, a transient with a zero
object ID, the R2a-save own unsaved object, an unreal heap item and the stdlib
own stamp.

R2a sits above the ownership check because it decides whether the call runs with write rights at all: a
non-crossing value reached by value is writable only as a local helper or while lent, and otherwise the
callee enters read-only mode with `viewOnly` and `viewS3` set (`enterReadOnlyByS3`), so
every write goes through `isReadonly` in the order shown. For the realm's own code the first write gets the
R2a text (E4, `machine.go`); a `/p/` value reached by value gets the library text
(E1, `readonlyAccessPanic`, `machine.go`), as `zrealm_saved_libtable_copy` shows. R2a-save sits under both: the
write goes through, the object is marked, and the refusal arrives at save (`assertNoUncrossedWrite`,
`realm.go`).

`isReadonly` compares the target's package ID with `m.Realm.ID` and the executing package's own stamp
(`ownership.go`, the package ID an object carries). When they differ and nothing exempts the target, the
write is refused and `readonlyAccessPanic` (`machine.go`) picks the text from the **writing code**, not
from the call stack:

- `/p/` code (an immutable, non-stdlib package): the library text, E1 (`readonlyAccessPanic`,
  `machine.go`).
  `zrealm_model_token.gno` shows it: the token's ledger is written by library code (`modelledger.Move`) and no
  crossing into the token realm is on the stack.
- A realm's code entered read-only (`m.viewRealm != nil`): the "entered without crossing" text, E2
  (`machine.go`). It names the realm whose code runs read-only, not the owner of the target.
  `zrealm_ctw_callback.gno` shows it while a crossing call into `ctw` is on the stack, because the refused
  write is by the callback's home realm, `ctw_callback`. `zrealm_launder_rdata_pdr_panic_defer.gno` shows it
  naming the attacker realm that writes the victim's object during panic unwinding.
- An R2a-entered stretch (`m.viewRealm != nil && m.viewS3`): the R2a text, E4 (`machine.go`):
  `<path>'s code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross)`. The
  callee is a function value outside R2a's three places, so there is no crossing function to call; make the
  callback crossing (`func(cur realm, …)`, called with `cross(cur)`) or pass a literal in place. `zrealm_maker_getdecl.gno` and `zrealm_saved_getter.gno` show it.
- Anything else: the fallback text, E3 (`machine.go`), when the writing code is not `/p/` and `viewRealm`
  is nil. `zrealm_crossrealm25a.gno` shows it: a crossing call into `crossrealm_b` gives that realm full
  rights, and its code writes an object owned by `crossrealm`.

The exemptions, each with the check that grants it:

- Single-user mode: `m.Realm == nil` and not view-only, so nothing is read-only (`machine.go`). This
  is the only nil-realm exemption: `setViewOnly` (`machine.go`) sets `m.Realm` nil *and* `viewOnly`
  true, so with `rid` left zero every stored object is refused.
- An object allocated since the read-only stretch began is scratch, not stored state (`scratchOwn` in
  `isReadonly`, `machine.go`).
- The read-only realm's own object that is not saved yet: R2a-save allows the write and marks it.
- A `RefValue` (a stored reference to a package, named by its path) whose package path is the writing package's own is writable; a foreign package path is not
  (`machine.go`).
- For a name, `isExternalRealm` (`machine.go`) exempts scratch, an object with a zero object
  ID, an unsaved transient (an unreal heap item is marked first when read-only, then the zero ID returns), and an
  unreal `HeapItemValue`, a heap wrapper for an escaping local.
- Stdlib code may write its own stdlib-stamped package global. `IsReadonly` grants the own stamp only to
  stdlib (`machine.go`), and `isExternalRealm` mirrors it.
  `zrealm_stdlib_global_write.gno` shows it: a stdlib function's `counter++` on its own package global, run
  from a realm, prints `1` then `2`.
