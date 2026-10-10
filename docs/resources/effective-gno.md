# Effective Gno

Welcome to the guide for writing effective Gno code. This document is designed
to help you understand the nuances of Gno and how to use it effectively.

Before we dive in, it's important to note that Gno shares several similarities
with Go. Therefore, if you haven't already, we highly recommend reading
["Effective Go"](https://go.dev/doc/effective_go) as a primer.

## Disclaimer

Gno is a young language. The practices we've identified are based on its current
state. As Gno evolves, new practices will emerge, and some current ones may
become obsolete. We welcome your contributions and feedback. Stay updated and
help shape Gno's future!

## Counter-intuitive good practices

This section highlights some Gno good practices that might seem
counter-intuitive, especially if you're coming from a Go background.

### Embrace global variables in realms

A realm's global variables are its state: the GnoVM saves them after each
transaction and restores them for the next, with no database to set up. Keep
them unexported, and let callers read or change them only through functions you
write. This holds for realms alone: a `p/` package's globals are frozen once its
`init` has run, so a package keeps no state of its own.

A function other realms or users call to change your state takes a `realm` as
its first parameter, which makes it a
[crossing function](./gno-interrealm.md#crossing-functions-and-crossing-methods):
the call enters your realm, and `cur.Previous()` on that parameter names the
caller. Write `_` instead of `cur` when the function never reads it.

```go
var counter int

func GetCounter() int {
	return counter
}

func IncCounter(_ realm) {
	counter++
}
```

### Embrace `panic`

An unrecovered `panic` aborts the transaction and discards every change it
made, while a returned `error` is ordinary data that stops nothing and undoes
nothing. So a realm panics on invalid input or a failed precondition, where Go
code would return an error. A reusable `p/` package returns an `error` and lets
its caller decide, and panics only in a `Must*` or `Assert*` wrapper whose name
says so, the way
[`gno.land/p/onbloc/uint256/v0`](../../examples/gno.land/p/onbloc/uint256/v0/uint256.gno)
pairs `FromDecimal` with `MustFromDecimal`. How a panic crosses realms is in
[the interrealm spec](./gno-interrealm.md#panic-and-revivefn).

### Understand the importance of `init()`

`init()` runs once, when the realm is deployed, so it is the realm's
constructor rather than a setup step run at every start as in Go. Use it for
what only the deployment knows, such as the deployer, and to register the realm
with another one. A value known in advance is set where its global is declared.

```go
var admin address

func init(cur realm) {
	// At deployment, the previous realm is the account publishing this one.
	admin = cur.Previous().Address()
}
```

### A little dependency is better than a little copying

Go's proverb runs the other way: ["A little copying is better than a little
dependency"](https://www.youtube.com/watch?v=PAAkCSZUG1c&t=568s). On gno.land a
published package that others can import can never be overwritten by its author,
since the chain refuses it, and its source is public. So importing a small,
widely used and audited `p/` package is safer than maintaining a copy, and a
user reading your realm sees parts they may already trust. Dependency code still
needs vetting, [whatever the ecosystem][sc-attack].

[sc-attack]: https://en.wikipedia.org/wiki/Supply_chain_attack

## When Gno takes Go practices to the next level

### Documentation is for users

Go's ["Documentation is for
users"](https://www.youtube.com/watch?v=PAAkCSZUG1c&t=1147s) goes further in
Gno, where the reader of a comment can be an end user deciding whether to call
your realm, not only a developer. Write each exported function's comment for
that reader: what calling it does for them. Here's an excerpt from
[grc20](../../examples/gno.land/p/nt/grc20/v0/types.gno):

```go
// Teller interface defines the methods that a GRC20 token must implement.
type Teller interface {
	// Returns the amount of tokens in existence.
	TotalSupply() int64

	// Returns the amount of tokens owned by `account`.
	BalanceOf(account address) int64

	// ...
}
```

### Reflection is never clear

Gno has no `reflect` package, as the [standard library
table](./go-gno-compatibility.md#stdlibs) shows. Write the explicit code
instead: users read on-chain code before calling it, and code
without reflection is easier for them to read and audit.

## Gno good practices

### Contract-level access control

Put each privileged endpoint behind a check of its caller, read with
`cur.Previous().Address()` on the `cur realm` parameter of a crossing function:
another realm if one called it, otherwise the user, much like Solidity's
`msg.sender`.

```go
var admin address // set in init(), as above

func AdminOnlyFunction(cur realm) {
	if cur.Previous().Address() != admin {
		panic("permission denied")
	}
	// ...
}
```

Checking the immediate caller rather than the signer lets a contract own
tokens or coins and be called by other contracts without putting the signer's
funds at risk. The default grc20 implementation works this way.

Never use [`unsafe.OriginCaller()`](./gno-stdlibs.md#origincaller) for access
control. It is Gno's `tx.origin`: a malicious realm the signer calls can act as
the signer towards yours, so keep it for recording who signed. A function that
must run only as a signer's direct call can enforce that with
[`runtime.AssertOriginCall()`](./gno-stdlibs.md#assertorigincall). It panics
unless the signer's `gnokey maketx call` entered that function directly: a call
through another named function, or from a `maketx run` script, panics.

For common needs, reuse the
[access-control helpers](./community-packages.md#access-control-helpers).
A session key signs as its master account, so it passes every
`cur.Previous().Address()` check that account passes. An endpoint that must
refuse session keys checks `runtime.GetSessionInfo()`.

### Design your realm as a public API

In Go, the boundary sits around your whole program. In Gno, every exported
function of a realm is an endpoint any user or realm can call, so check the
caller there and keep the logic in unexported functions.

```go
var admin address // set in init()

func PublicMethod(cur realm, nb int) {
	caller := cur.Previous().Address()
	if caller != admin {
		panic("caller is not the admin")
	}
	privateMethod(caller, nb)
}

func privateMethod(caller address, nb int) { /* ... */ }
```

`privateMethod` is unexported, so no other realm can call it, and it can trust
the address it receives. An exported function returning a pointer to stored
state, such as an `*avl.Tree`, hands its mutating methods to every caller, so
return values instead, per the
[security guide](./gno-security-guide.md#51-exposing-a-pointer-to-mutable-state).

### Never call a caller-supplied function under your own authority

A function declared in a realm runs under that realm's authority, and a
closure under the authority of the realm that created it, wherever either is
called from. A top-level function declared in a `/p/` package has no realm of
its own, so when your realm calls it, it runs as your realm and can rewrite
your state. A caller can hand you exactly such a function. So never invoke a
callback or interface value a caller supplies: return the result and let the
caller act on it.
The [security guide](./gno-security-guide.md#53-accepting-an-attacker-callback-under-your-own-authority)
covers the vector in full.

### Construct "safe" objects

A safe object is one other realms can hold and pass around, even by pointer,
without being able to misuse it. Each mutating method takes the calling realm
as `rlm` and checks who called that realm, so any realm the admin is calling
can change it. To keep writes in the realm that created the object, also
compare `rlm.PkgPath()` with that realm, as grc20's tellers do.

```go
type MySafeStruct struct {
	counter int
	admin   address
}

// A /p/ package cannot declare crossing functions, so the realm comes after a
// placeholder `_ int` that callers pass 0 for, as p/nt/ownable does.
func NewSafeStruct(_ int, rlm realm) *MySafeStruct {
	if !rlm.IsCurrent() {
		panic("realm handle is not the live caller")
	}
	return &MySafeStruct{admin: rlm.Previous().Address()}
}

func (s *MySafeStruct) Counter() int { return s.counter }

func (s *MySafeStruct) Inc(_ int, rlm realm) {
	if !rlm.IsCurrent() {
		panic("realm handle is not the live caller")
	}
	if rlm.Previous().Address() != s.admin {
		panic("permission denied")
	}
	s.counter++
}
```

A realm creates one in its `init` with `NewSafeStruct(0, cur)` and hands it to
other realms, which can read it and pass it on.

### Choosing between Coins and GRC20 tokens

Coins are simple, and the chain fixes their rules. GRC20 tokens run contract
code, so they can do whatever their contract allows. Choose by which of the two
you need.

#### Coins

Coins are balances the chain keeps outside the GnoVM. A realm issues its own
denom through a [banker](./gno-stdlibs.md#banker), which can also
[burn](./gno-stdlibs.md#removecoin) that denom from any holder's balance. A
plain bank transaction moves coins and a bank query reads a balance; neither
runs contract code. A denom listed in the chain's `restricted_denoms` parameter
can be sent only from accounts the chain has whitelisted.

Read one balance with `GetCoin(addr, denom)` rather than `GetCoins(addr)`,
which reads every denom the address holds. Anyone can send any address a new
denom, so `GetCoins` on a caller-supplied address costs what a third party
chose, and enough denoms make your function impossible to call. `GetCoin`
panics on a malformed denom, so validate one you do not control first.

#### Verifying inbound Coin payments

Gno has no `payable` keyword: a `maketx call` that attaches coins fails unless
the called realm reads them with `unsafe.OriginSend()`. That returns the coins
attached to the transaction, not the coins your realm received. Pair it with
`cur.Previous().IsUserCall()`, which holds only when a user's `maketx call`
entered your realm directly, so the `-send` coins are known to have landed at
your realm's address:

```go
import "chain/runtime/unsafe"

func BuyThing(cur realm) {
	if !cur.Previous().IsUserCall() { // not IsUser()
		panic("must be called directly by an EOA (maketx call)")
	}
	if unsafe.OriginSend().AmountOf("ugnot") != price {
		panic("wrong payment amount")
	}
	// ... do the thing ...
}
```

`IsUser()` is not enough: it also accepts a `maketx run` script, whose `-send`
coins never leave the signer's address, so the amount check passes while nothing
reached your realm. Keep both checks: without the guard, `OriginSend()`
overstates what arrived, and without the amount check, a user pays nothing. A
realm cannot pull coins from its caller, since `banker.SendCoins` sends only
from the realm's own address, so this pair is the only way to take a payment.
For the strictest form, add `runtime.AssertOriginCall()`, as `wugnot`'s
`Deposit` does.

#### GRC20 tokens

GRC20 is Gno's ERC20, `Approve` and `TransferFrom` included. Every transfer runs
the contract that defines the token, which keeps control over its rules, so a
token can gate access, sit in a vault or count votes in a DAO.

```go
import "gno.land/p/nt/grc20/v0"

var (
	Token         *grc20.Token
	privateLedger *grc20.PrivateLedger
	// Keep the teller unexported: it is a spend capability. It is built from
	// the ledger, which never leaves this realm, and it only works here.
	userTeller grc20.Teller
)

func init(cur realm) {
	Token, privateLedger = grc20.NewToken("Foo Token", "FOO", 4, 0, cur)
	userTeller = privateLedger.CallerTeller()
}

// Transfer moves the caller's own tokens (userTeller debits cur.Previous()).
func Transfer(cur realm, to address, amount int64) {
	if err := userTeller.Transfer(0, cur, to, amount); err != nil {
		panic(err)
	}
}
```

See also [`gno.land/r/demo/defi/foo20`](../../examples/gno.land/r/demo/defi/foo20).

#### Wrapping Coins

A realm can wrap a coin in a GRC20 token: a user deposits the coin and receives
the token, which then moves under the contract's rules, and withdrawing burns
the token and returns the coin.
[`gno.land/r/gnoland/wugnot`](../../examples/gno.land/r/gnoland/wugnot) wraps
`ugnot` this way.

### Do not trust on-chain randomness

The top-level `math/rand` functions draw from a
[fixed seed](../../gnovm/stdlibs/math/rand/rand.gno), and any seed a contract
computes instead is public. A lottery with real stakes therefore needs
commit-reveal or an external source.

### Bring off-chain data on-chain with oracles

An oracle is an agreement with off-chain agents you choose to trust. The
chain never verifies the off-chain fact, only that a whitelisted agent
attested to it, so choose your agents accordingly.
[gnorkle](../../examples/gno.land/p/demo/gnorkle/README.md) already manages
the agent whitelists and requests.
gno.land has no built-in price feed, so a realm that moves funds based on a
fed value is only as secure as whoever provides that value: an attacker does
not need a bug in your code, only a bad number in the feed.

### Test the attacker, not just the happy path

The test kinds are covered in [the testing guide](./gno-testing.md). What
realm tests add is the execution context: the `testing` package lets you call
your realm as someone else.

- `testing.NewUserRealm` and `testing.NewCodeRealm` stand in for a user or
  another contract.
- `testing.SetRealm` and `testing.SetOriginCaller` set the caller.
- `testing.IssueCoins` and `testing.SkipHeights` set up funds and time.
  `testing.SetOriginSend` only sets what `OriginSend()` reports and moves no
  coins.

Use them to attack your own realm: simulate an intermediary contract and prove
your [payment check](#verifying-inbound-coin-payments) cannot be bypassed.

### Choose storage types by access pattern

A `map` or slice is stored as one object, so reading or updating one element
loads or rewrites all of it, while a tree stores its nodes or leaf pages
separately, so touching one key loads only the path to it.
Keep maps and slices for small, bounded state, and put anything that grows in
a tree; [Gno data structures](./gno-data-structures.md#tree-backed-indexes)
compares the tree types.

### Define types and interfaces in pure packages (p/)

Put the types and interfaces others should share in `p/NAMESPACE/DAPP`, and the
realm using them in `r/NAMESPACE/DAPP`. A `p/` package cannot import an `r/`,
while an `r/` can import both, so only a definition in `p/` can become a
standard that other packages build on. A realm introducing nothing reusable
needs no `p/` at all.

### Emit Gno events to make life off-chain easier

Emit an event after each action an indexer would want to know about, such as a
transfer or an ownership change. [`chain.Emit`](./gno-stdlibs.md#events) writes
a type and `key`, `value` pairs into the block's results, where off-chain
services filter and search them.

```go
import "chain"

var owner address

func ChangeOwner(cur realm, newOwner address) {
	if cur.Previous().Address() != owner {
		panic("access denied")
	}
	owner = newOwner
	chain.Emit("OwnershipChange", "newOwner", newOwner.String())
}
```

### Package naming and organization

A package's name matches the last element of its
[path](./gno-packages.md#package-path-structure), ignoring a `/vN` suffix, so
imports need no alias. Keep it short and readable. A project usually lives at
`r/NAMESPACE/DAPP`, and each realm under it is independent and shares no data,
so split a project only for a reason, such as one `r/NAMESPACE/auth` every
contract of an organization imports. Publish a top-level `p/NAMESPACE/DAPP`
only for what others should use, and keep your own helpers in subdirectories
under it. A path containing an `internal` element can be imported only from
under the directory holding `internal`: `mypackage/internal/helpers` is open to
`mypackage` and `mypackage/utils`, and to nothing outside `mypackage`.

### Suggested file names and layout

A realm's source is browsable on-chain, so its file names are part of its
public interface. Name each file for the one concern it holds, so a reader can
guess its contents: `<realm>.gno` for the main entrypoints, `render.gno` for
`Render()`, `admin.gno` for privileged endpoints, `types.gno` and `errors.gno`
once those grow. [`gno.land/r/sys/cla`](../../examples/gno.land/r/sys/cla)
splits into `cla.gno`, `admin.gno` and `render.gno`, with tests beside them.

### Versioning and upgrades

An upgrade is a new deployment at a new
[`/vN` path](./gno-packages.md#version-suffixes), next to the old one; only a
[`private`](./configuring-gno-projects.md#private) realm, which nothing can
import, is re-uploaded in place. A realm
whose callers need one stable address can forward to an implementation it
swaps, the way
[`gno.land/r/gov/dao`](../../examples/gno.land/r/gov/dao/proxy.gno) keeps the
current implementation in a variable. A simpler realm keeps versions side by
side, as `gno.land/r/sys/validators` keeps `v0` and `v2`. A small, well-tested
`p/` library may never need a second version at all.

### Ship more than code

A realm's source and `Render()` output are public on gnoweb, so treat them as
your storefront: a `Render()` a stranger can navigate, a `README.md`, and doc
comments [written for users](#documentation-is-for-users). Around the realm,
plan for the client, the indexers fed by your
[events](#emit-gno-events-to-make-life-off-chain-easier), and the docs they
need. A `p/` package is a library for developers and a realm is an app for
end users: if a realm needs a wall of external docs to be usable, that text
belongs in its `Render()`. Its path is caller input, so pass path segments and
stored user strings through
[`sanitize.InlineText`](../../examples/gno.land/p/nt/markdown/sanitize/v0)
before they reach the markdown.

### Treat forking as a feature

Every realm and package is published source. Where its license allows, a
modified copy deployed at a path you control gives users somewhere to go when
a contract is abandoned or hostile. Keep the
upstream license file and say where the code came from:
`gno.land/p/onbloc/uint256/v0` ports
[holiman/uint256](https://github.com/holiman/uint256) and ships its BSD
license beside the code.

### Generate repetitive code off-chain

Gno [has no generics](./go-gno-compatibility.md#reserved-keywords) and no on-chain
`go generate`, but nothing stops you from generating `.gno` source before
deployment and committing the output.
[`gno.land/p/moul/xmath`](../../examples/quarantined/gno.land/p/moul/xmath)
works this way:
`generator.go` writes the same helpers once per numeric type into
`xmath.gen.gno`, standing in for generics. Readers on-chain see the final
source, so generated files stay auditable.
