# Effective Gno

This guide covers the practices that make Gno code safe and easy to read. Gno
is close to Go, so read ["Effective Go"](https://go.dev/doc/effective_go) first
if you haven't.

## Counter-intuitive good practices

These practices may surprise you if you come from Go.

### Embrace global variables in realms

In a realm, global variables are the state. The GnoVM saves them after each
transaction and loads them for the next, so you need no database. Keep them
unexported, and change them only through functions you write.

This applies to realms only. A `p/` package's globals freeze once its `init`
has run, so a package keeps no state.

An exported function can write your state even without a `realm` parameter,
and then it cannot tell who called. So give every exported function that writes
state a `realm` as its first parameter. That makes it a
[crossing function](./gno-interrealm-v2.md#5-crossing-functions-and-crossing-methods):
the call enters your realm, and `cur.Previous()` tells you who called. Name the
parameter `_` when you don't read it.

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

In Gno, `panic` is the normal way to refuse a call. A panic that nothing
recovers aborts the transaction and reverts every change it made. A returned
`error` does neither: it is plain data, and the changes stay.

So a realm panics on bad input or a failed check, where Go would return an
error. A `p/` package returns an `error` and lets its caller decide. It panics
only in a wrapper named `Must*` or `Assert*`, the way
[`gno.land/p/onbloc/uint256/v0`](../../examples/gno.land/p/onbloc/uint256/v0/uint256.gno)
pairs `FromDecimal` with `MustFromDecimal`.

A caller [cannot recover](./gno-interrealm-v2.md#9-panic-and-cross-realm-boundary)
a panic raised in your realm. The whole transaction aborts.

### Understand the importance of `init()`

`init()` runs once, when the realm is deployed. It is the realm's constructor.
Use it for what only the deployment knows, such as who deployed the realm, and
to register the realm with another one. Set a value you know in advance where
its global is declared.

```go
var admin address

func init(cur realm) {
	// At deployment, the previous realm is the account publishing this one.
	admin = cur.Previous().Address()
}
```

### A little dependency is better than a little copying

Go says the opposite: ["A little copying is better than a little
dependency"](https://www.youtube.com/watch?v=PAAkCSZUG1c&t=568s). On gno.land,
a package others can import can never be changed by its author, and its source
is public. So import a small, widely used, audited `p/` package rather than
copying it. Users reading your realm then see parts they may already trust.
Still vet what you import: [supply-chain attacks][sc-attack] happen in every
ecosystem.

[sc-attack]: https://en.wikipedia.org/wiki/Supply_chain_attack

## When Gno takes Go practices to the next level

### Documentation is for users

Go's proverb is ["Documentation is for
users"](https://www.youtube.com/watch?v=PAAkCSZUG1c&t=1147s). In Gno, that user
may be someone deciding whether to call your realm, not a developer. Write each
exported function's comment for them: say what the call does for them.
gnoweb's `$help` page shows that comment next to the function's call form, so
it is what a user reads before calling. From
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

Gno has no `reflect` package, as the
[standard library table](./go-gno-compatibility.md#stdlibs) shows. Write
explicit code instead. Users read on-chain code before calling it, and explicit
code is easier for them to check.

## Gno good practices

### Contract-level access control

Check who called before doing anything privileged. In a crossing function,
`cur.Previous().Address()` is the caller: the user's address when a user called
you directly, or the calling realm's address when another realm did. It plays
the role of Solidity's `msg.sender`.

```go
var admin address // set in init(), as above

func AdminOnlyFunction(cur realm) {
	if cur.Previous().Address() != admin {
		panic("permission denied")
	}
	// ...
}
```

- Never authorize with the signer,
  [`unsafe.OriginCaller()`](./gno-stdlibs.md#origincaller). Any realm the user
  calls could then act as that user towards yours, like Solidity's `tx.origin`.
  Use it only to record who signed. Checking the direct caller is also what
  lets a realm hold tokens and be called by other realms, as grc20 does.
- To accept only a user's direct call, with no realm in between, call
  [`runtime.AssertOriginCall()`](./gno-stdlibs.md#assertorigincall). It panics
  when the call comes through another named function or from a `maketx run`
  script.
- A session key passes the same check as its master account. To refuse session
  keys on an endpoint, check `runtime.GetSessionInfo()`.
- For an owner or a pause switch, reuse the
  [access-control helpers](./community-packages.md#access-control-helpers)
  rather than writing your own.

### Design your realm as a public API

Every exported function of a realm is an endpoint that any user or realm can
call. Read the caller there, then pass it to unexported functions that do the
work.

```go
func PublicMethod(cur realm, nb int) {
	caller := cur.Previous().Address()
	privateMethod(caller, nb)
}

func privateMethod(caller address, nb int) { /* ... */ }
```

No other realm can call `privateMethod`, so it can trust the address it gets.

Never return a pointer to stored state, such as an `*avl.Tree`. Every caller
could then use its methods to change your state. Return values instead, as the
[security guide](./gno-security-guide.md#51-exposing-a-pointer-to-mutable-state)
explains.

### Never call a caller-supplied function under your own authority

A top-level function from a `/p/` package has no realm of its own. When your
realm calls one, it runs as your realm and can change your state. A caller can
pass you exactly such a function. So never call a callback or interface value
a caller gives you. Return the result instead, and let the caller act on it.
The [security guide](./gno-security-guide.md#53-accepting-an-attacker-callback-under-your-own-authority)
covers this in full.

The same holds the other way. Pass `cur` only to code you trust: with it, the
callee can build a banker over your realm's address and keep it, which lets it
spend your coins later.

### Construct "safe" objects

A safe object is one other realms can hold and pass around without being able
to misuse it. Each method that changes it takes the calling realm as `rlm` and
checks two things: that the call comes from the realm that created the object,
and that the admin called that realm.

```go
type MySafeStruct struct {
	counter int
	admin   address
	home    string // path of the realm that created the object
}

// This is /p/ code, which cannot declare crossing functions, so the caller's
// realm comes after a placeholder `_ int` that callers pass as 0.
func NewSafeStruct(_ int, rlm realm) *MySafeStruct {
	if !rlm.IsCurrent() {
		panic("realm handle is not the live caller")
	}
	return &MySafeStruct{admin: rlm.Previous().Address(), home: rlm.PkgPath()}
}

func (s *MySafeStruct) Counter() int { return s.counter }

func (s *MySafeStruct) Inc(_ int, rlm realm) {
	if !rlm.IsCurrent() {
		panic("realm handle is not the live caller")
	}
	if rlm.PkgPath() != s.home {
		panic("only the creating realm can change this")
	}
	if rlm.Previous().Address() != s.admin {
		panic("permission denied")
	}
	s.counter++
}
```

Without the `home` check, any realm the admin calls could change the object.
grc20's tellers make the same check. A realm creates the object in its `init`
with `NewSafeStruct(0, cur)`. Other realms can then read it and pass it on.

### Choosing between Coins and GRC20 tokens

Coins are simple, and the chain sets their rules. GRC20 tokens are contracts,
so they follow whatever rules their code defines. Pick the one your realm
needs.

#### Coins

Coins are balances the chain keeps outside the GnoVM. Sending them or reading a
balance runs no contract code. A realm can issue its own denom through a
[banker](./gno-stdlibs.md#banker), and can [burn](./gno-stdlibs.md#removecoin)
that denom from any holder. A denom listed in the chain's `restricted_denoms`
parameter can be sent only from whitelisted accounts.

#### Verifying inbound Coin payments

Gno has no `payable` keyword. A `maketx call` that sends coins fails unless the
realm reads them with `unsafe.OriginSend()`.

`OriginSend()` returns the coins attached to the transaction, not the coins
your realm received. So pair it with `cur.Previous().IsUserCall()`. That check
passes only when a user's `maketx call` reached your realm directly, which is
when the coins landed at your address.

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

- Don't use `IsUser()`. It also accepts a `maketx run` script, which runs as the
  realm `gno.land/e/<addr>/run` and whose coins never leave the user's address.
  The amount check would pass with nothing paid.
- Keep both checks. Without the caller check, `OriginSend()` can report coins
  you never got. Without the amount check, a user pays nothing.
- A realm can't pull coins from its caller: `banker.SendCoins` only sends from
  the realm's own address. So this pair is the only way to take a payment.
- For the strictest form, also call `runtime.AssertOriginCall()`, as `wugnot`'s
  `Deposit` does.
- Creating a `BankerTypeOriginSend` banker runs the same `IsUserCall()` check,
  and panics when it fails.

#### GRC20 tokens

GRC20 is Gno's version of ERC20, with `Approve` and `TransferFrom`. Every
transfer runs the token's contract, which keeps control of the rules. That lets
a token gate access, sit in a vault, or count votes in a DAO.

`Approve` replaces the allowance rather than adding to it. To change one that
isn't zero, set it to 0 first. Otherwise the spender can spend both the old and
the new amount.

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

A realm can wrap a coin in a GRC20 token. A user deposits the coin and gets the
token, which then follows the contract's rules. Withdrawing burns the token and
returns the coin.
[`gno.land/r/gnoland/wugnot`](../../examples/gno.land/r/gnoland/wugnot) wraps
`ugnot` this way.

### Do not trust on-chain randomness

The top-level `math/rand` functions use a
[fixed seed](../../gnovm/stdlibs/math/rand/rand.gno). Any seed a contract
computes is public too. So a lottery with real stakes needs commit-reveal or an
outside source of randomness.

### Bring off-chain data on-chain with oracles

An oracle is a set of off-chain agents you choose to trust. The chain never
checks the outside fact, only that one of your agents reported it, so pick them
with care. [gnorkle](../../examples/gno.land/p/demo/gnorkle/README.md) manages
the agent list and the requests.

gno.land has no built-in price feed. A realm that moves funds based on a fed
value is only as safe as the feed: an attacker needs no bug in your code, just
a bad number.

### Test the attacker, not just the happy path

[The testing guide](./gno-testing.md) covers the kinds of tests. In a realm,
the `testing` package also lets you call your realm as someone else:

- `testing.NewUserRealm` and `testing.NewCodeRealm` stand in for a user or
  another contract.
- `testing.SetRealm` and `testing.SetOriginCaller` set the caller.
- `testing.IssueCoins` and `testing.SkipHeights` set up funds and time.
- `testing.SetOriginSend` sets what `OriginSend()` reports. It moves no coins.

Use them to attack your own realm. For example, call it through a fake
intermediary contract and prove your
[payment check](#verifying-inbound-coin-payments) holds.

### Choose storage types by access pattern

A `map` or slice is stored as one object. Reading or changing one element loads
or rewrites the whole thing. A tree stores its nodes separately, so touching
one key loads only the path to it. Use maps and slices for small, bounded
state, and a tree for anything that grows.
[Gno data structures](./gno-data-structures.md#tree-backed-indexes) compares
the tree types.

### Define types and interfaces in pure packages (p/)

Put shared types and interfaces in `p/NAMESPACE/DAPP`, and the realm that uses
them in `r/NAMESPACE/DAPP`. A `p/` package can't import an `r/`, but an `r/`
can import both. So only a definition in `p/` can become a standard that others
build on. A realm with nothing reusable needs no `p/`.

### Emit Gno events to make life off-chain easier

Emit an event after each action an indexer would care about, such as a
transfer or a change of owner. [`chain.Emit`](./gno-stdlibs.md#events) writes
a type and `key`, `value` pairs into the block results, where off-chain
services can search them.

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

- Name the package after the last element of its
  [path](./gno-packages.md#package-path-structure), ignoring a `/vN` suffix.
  Imports then need no alias. Keep the name short.
- A project usually lives at `r/NAMESPACE/DAPP`. Each realm under it is
  separate and shares no data, so split a project only for a reason, such as
  one `r/NAMESPACE/auth` that all your realms import.
- Publish a top-level `p/NAMESPACE/DAPP` only for code others should use. Keep
  your own helpers in subdirectories under it.
- A path with an `internal` element can be imported only from inside the
  directory above `internal`. `mypackage/internal/helpers` is open to
  `mypackage` and `mypackage/utils`, and to nothing outside `mypackage`.

### Suggested file names and layout

Anyone can browse a realm's source on-chain, so its file names are part of its
public face. Name each file for the one thing it holds: `<realm>.gno` for the
main entry points, `render.gno` for `Render()`, `admin.gno` for privileged
endpoints, then `types.gno` and `errors.gno` once those grow.
[`gno.land/r/sys/cla`](../../examples/gno.land/r/sys/cla) splits into
`cla.gno`, `admin.gno` and `render.gno`, with tests beside them.

### Versioning and upgrades

You upgrade by deploying a new version at a new
[`/vN` path](./gno-packages.md#version-suffixes), next to the old one. Only a
[`private`](./configuring-gno-projects.md#private) realm, which nothing can
import, can be re-uploaded in place.

If callers need one stable address, the realm can forward calls to an
implementation it can swap, the way
[`gno.land/r/gov/dao`](../../examples/gno.land/r/gov/dao/proxy.gno) keeps the
current implementation in a variable. A simpler realm keeps its versions side
by side.

### Ship more than code

Anyone can read a realm's source and its `Render()` output on gnoweb. Treat
them as your storefront: give strangers a `Render()` they can navigate, a
`README.md`, and doc comments [written for users](#documentation-is-for-users).
Plan for the client, for the indexers your
[events](#emit-gno-events-to-make-life-off-chain-easier) feed, and for the docs
they need. A `p/` package is a library for developers; a realm is an app for
end users. If a realm needs pages of outside docs to be usable, put that text
in its `Render()`.

`Render()`'s path comes from the caller. Pass path segments and stored user
strings through
[`sanitize.InlineText`](../../examples/gno.land/p/nt/markdown/sanitize/v0)
before they reach the markdown.

### Treat forking as a feature

Every realm and package is published source. If its license allows, deploy a
modified copy at a path you control. That gives users somewhere to go when a
contract is abandoned or hostile. Keep the original license file and say where
the code came from: `gno.land/p/onbloc/uint256/v0` ports
[holiman/uint256](https://github.com/holiman/uint256) and ships its BSD license
beside the code.

### Generate repetitive code off-chain

Gno has [no generics](./go-gno-compatibility.md#reserved-keywords) and no
on-chain `go generate`. Nothing stops you from generating `.gno` source before
you deploy and committing the result.
[`gno.land/p/moul/xmath`](../../examples/quarantined/gno.land/p/moul/xmath)
does this: `generator.go` writes the same helpers once per number type into
`xmath.gen.gno`. Readers on-chain see the final source, so generated code
stays easy to audit.
