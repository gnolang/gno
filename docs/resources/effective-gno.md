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

In Gno, using global variables is not only acceptable but also encouraged,
specifically when working with realms. This is due to the unique persistence
feature of realms.

In Go, you would typically write your logic and maintain some state in memory.
However, to persist the state and ensure it survives a restart, you would need
to use a store (like a plain file, custom file structure, a database, a
key-value store, an API, etc.).

In contrast, Gno simplifies this process. When you declare global variables in
Gno realms, the GnoVM automatically persists and restores them as needed between
each run. This means that the state of these variables is maintained across
different executions of the realm, providing a simple and efficient way to
manage state persistence.

However, it's important to note that this practice is not a blanket
recommendation for all Gno code. It's specifically beneficial in the context of
realms due to their persistent characteristics. A `p/` package's globals are
frozen once its `init` has run, so packages should use constants for values
that don't change.

Also, be mindful not to export your global variables. Doing so would make them
accessible for everyone to read and write, potentially leading to unintended
side effects. Instead, consider using getters and setters to control access to
these variables, as shown in the following pattern:

```go
// private global variable.
var counter int

// public getter endpoint.
func GetCounter() int {
	return counter
}

// public setter endpoint.
func IncCounter(_ realm) {
	counter++
}
```

In this example, `GetCounter` and `IncCounter` are used to read and increment
the `counter` variable, respectively. This allows you to control how the
`counter` variable is accessed and modified, ensuring that it's used correctly
and securely.

### Embrace `panic`

In Gno, we have a slightly different approach to handling errors compared to Go.
While the famous [quote by Rob
Pike](https://github.com/golang/go/wiki/CodeReviewComments#dont-panic) advises
Go developers "Don't panic.", in Gno, we actually embrace `panic`.

Panic in Gno is not just for critical errors or programming mistakes, as it is in
Go. Instead, it's used as a control flow mechanism to stop the execution of a
[realm](./realms.md) when something goes wrong. This could be due to an invalid input, a
failed precondition, or any other situation where it's not possible or desirable
to continue executing the contract.

So, while in Go, you should avoid `panic` and handle `error`s gracefully, in Gno,
don't be afraid to use `panic` to enforce contract rules and protect the integrity
of your contract's state. Remember, a well-placed panic can save your contract
from a lot of trouble.

When you return an `error` in Gno, it's like giving back any other piece of data.
It tells you something went wrong, but it doesn't stop your code or undo any
changes you made.

But when you use `panic` in Gno, it stops your code right away, says it failed,
and doesn't save any changes you made. This is safer when you want to stop
everything and not save wrong changes.

In Gno, the use of `panic()` and `error` should be context-dependent to ensure
clarity and proper error handling:
- Use `panic()` to immediately halt execution and roll back the transaction when
  encountering critical issues or invalid inputs that cannot be recovered from.
- Return an `error` when the situation allows for the possibility of recovery or
  when the caller should decide how to handle the error.

Consequently, reusable packages should avoid `panic()` except in assert-like
functions, such as `Must*` or `Assert*`, which are explicit about their
behavior. Packages should be designed to be flexible and not impose restrictions
that could lead to user frustration or the need to fork the code.

```go
func Foobar(_ int, rlm realm) {
	if !rlm.IsCurrent() {
		panic("realm handle is not the live caller")
	}
	caller := rlm.Previous().Address()
	if caller != "g1xxxxx" {
		panic("permission denied")
	}
	// ...
}
```

For reusable `p/` packages, a common way to offer both styles is to pair an
error-returning function with a thin `Must*` or `Assert*` wrapper that panics.
The plain function stays flexible for callers who want to recover; the wrapper
is for realm code that prefers to fail fast and roll back:

```go
import "errors"

// Returns an error, so callers decide how to handle it.
func ParseAddress(s string) (address, error) {
	addr := address(s)
	if !addr.IsValid() {
		return "", errors.New("invalid address: " + s)
	}
	return addr, nil
}

// Panics on failure.
func MustParseAddress(s string) address {
	addr, err := ParseAddress(s)
	if err != nil {
		panic(err)
	}
	return addr
}
```

Keep the `Must`/`Assert` prefix so the panic is obvious to anyone reading the
call site.
[`gno.land/p/onbloc/uint256/v0`](../../examples/gno.land/p/onbloc/uint256/v0/uint256.gno)
pairs `FromDecimal` with `MustFromDecimal` this way.

### Understand the importance of `init()`

In Gno, the `init()` function isn't just a function; it's a cornerstone. It's
automatically triggered when a new realm is added on-chain, making it a one-time
setup tool for the lifetime of a realm. In essence, `init()` acts as a
constructor for your realm.

Unlike Go, where `init()` is used for tasks like setting up database
connections, configuring logging, or initializing global variables every time
you start a program, in Gno, `init()` is executed once in a realm's lifetime.

In Gno, `init()` primarily serves two purposes:
1. It establishes the initial state, specifically, setting up global variables.
	- Note: a global can often be set where it is declared, as `list` is in
	  the second example below. Use `init` when the value comes from the
	  deployment, such as the deployer from `cur.Previous()`, as `admin` does.
2. It communicates with another realm, for example, to register itself in a registry.

```go
import "gno.land/r/some/registry"

func init(cur realm) {
	registry.Register(cross(cur), "myID", myCallback)
}

func myCallback(a, b string) { /* ... */ }
```

A common use case could be to set the "admin" as the caller uploading the
package.

```go
import "time"

var (
	created time.Time
	admin   address
	list    = []string{"foo", "bar", time.Now().Format("15:04:05")}
)

func init(cur realm) {
	created = time.Now()
	// During initialization the previous realm is the publisher of the realm
	// (the EOA that deployed it), so capturing it here is better than
	// hardcoding an admin address as a constant.
	admin = cur.Previous().Address()
	// list is already initialized, so it will already contain "foo", "bar" and
	// the current time as existing items.
	list = append(list, admin.String())
}
```

In essence, `init()` in Gno is your go-to function for setting up and
registering realms. It's a powerful tool that helps keep your realms organized
and properly configured from the get-go. Acting as a constructor, it sets the
stage for the rest of your realm's lifetime.

### A little dependency is better than a little copying

In Go, there's a well-known saying by Rob Pike: ["A little copying is better
than a little dependency"](https://www.youtube.com/watch?v=PAAkCSZUG1c&t=568s).
This philosophy encourages developers to minimize their dependencies and instead
copy small amounts of code where necessary. While this approach often makes
sense in Go, it's not always the best strategy in Gno.

In Gno, especially for `p/` packages, another philosophy prevails, one that is
more akin to the Node/NPM ecosystem. This philosophy encourages creating small
modules and leveraging multiple dependencies. The main reason for this shift is
code readability and trust.

A Gno contract is not just its lines of code, but also the imports it uses. More
importantly, Gno contracts are not just for developers. For the first time, it
makes sense for users to see what functionality they are executing too. Code simplicity, transparency,
explicitness, and trustability are paramount.

Another good reason for creating simple, focused libraries is the composability
of Go and Gno. Essentially, you can think of each `p/` package as a Lego brick
in an ever-growing collection, giving more power to users. `p/` in Gno is
basically a way to extend the standard libraries in a community-driven manner.

Unlike other compiled languages, where dependencies are not always well-known and
clear metrics are lacking, Gno allows for a reputation system not only for the
called contracts, but also for the dependencies.

For example, you might choose to use well-crafted `p/` packages that have been
reviewed, audited, and widely used. This approach can make your code footprint
smaller and more reliable.

In other platforms, an audit usually involves auditing everything, including the
dependencies. However, in Gno, we can expect that over time, contracts will
become smaller, more powerful, and partially audited by default, thanks to this
enforced open-source system.

One key difference between the Go and Gno ecosystems is the trust assumption when
adding a new dependency. Dependency code always needs to be vetted, [regardless
of what programming language or ecosystem you're using][sc-attack]. However, in
Gno, you can have the certainty that the author of a package cannot overwrite an
existing, published contract; as that is simply disallowed by the blockchain. In
other words, using existing and widely-used packages reinforces your security
rather than harming it.

[sc-attack]: https://en.wikipedia.org/wiki/Supply_chain_attack

So, while you can still adhere to the original philosophy of minimizing
dependencies, ultimately, try to use and write super stable, simple, tested,
and focused `p/` small libraries. This approach can lead to more reliable,
efficient, and trustworthy Gno contracts.

```go
import (
	"gno.land/p/finance/exchange"
	"gno.land/p/finance/tokens"
	"gno.land/p/finance/wallet"
	"gno.land/p/utils/permissions"
)

var (
	myWallet   wallet.Wallet
	myToken    tokens.Token
	myExchange exchange.Exchange
)

func init() {
	myWallet = wallet.NewWallet()
	myToken = tokens.NewToken("MyToken", "MTK")
	myExchange = exchange.NewExchange(myToken)
}

func BuyTokens(cur realm, amount int) {
	caller := cur.Previous().Address()
	permissions.CheckPermission(caller, "buy")
	myWallet.Debit(caller, amount)
	myExchange.Buy(caller, amount)
}

func SellTokens(cur realm, amount int) {
	caller := cur.Previous().Address()
	permissions.CheckPermission(caller, "sell")
	myWallet.Credit(caller, amount)
	myExchange.Sell(caller, amount)
}
```

## When Gno takes Go practices to the next level

### Documentation is for users

One of the well-known proverbs in Go is: ["Documentation is for
users"](https://www.youtube.com/watch?v=PAAkCSZUG1c&t=1147s), as stated by Rob
Pike. In Go, documentation is primarily for users, but users are often developers themselves. In Gno,
documentation is for users, and users can be other developers as well as end users.

In Go, we usually have well-written documentation for other developers to
maintain and use our code as a library. Then, we often have another layer of
documentation on our API, sometimes with OpenAPI Specs, Protobuf, or even user
documentation.

In Gno, the focus shifts towards writing documentation for the end user. You can
even consider that the main reader is an end user, who is not so interested in
technical details, but mostly interested in how and why they should use a
particular endpoint. Comments will be used to aid code source reading, but also to
generate documentation, and even for smart wallets that need to understand what
to do.

Inline comments have the same goal: to guide users (developers or end users)
through the code. While comments are still important for maintainability, their
main purpose in Gno is for discoverability. This shift towards user-centric
documentation reflects the broader shift in Gno towards making code more
accessible and understandable for all users, not just developers.

Here's an excerpt from
[grc20](../../examples/gno.land/p/nt/grc20/v0/types.gno), where each comment
tells the caller what a method does for them:

```go
// Teller interface defines the methods that a GRC20 token must implement.
type Teller interface {
	// ...

	// Returns the amount of tokens owned by `account`.
	BalanceOf(account address) int64

	// Moves `amount` tokens from the caller's account to `to`. rlm must
	// be the caller's own captured cur — verified via rlm.IsCurrent().
	//
	// Returns an error if the operation failed.
	Transfer(_ int, rlm realm, to address, amount int64) error

	// ...
}
```

### Reflection is never clear

In Go, there's a well-known saying by Rob Pike: ["Reflection is never
clear."](https://www.youtube.com/watch?v=PAAkCSZUG1c&t=15m22s) This statement
emphasizes the complexity and potential pitfalls of using reflection in Go.

Gno has no `reflect` package yet, which the [standard library
table](./go-gno-compatibility.md#stdlibs) lists as to be added. There are
technical reasons for this, but also a desire to create a Go alternative that is
explicitly safer to use than Go, with a smaller cognitive difficulty to read,
discover, and understand.

The absence of reflection in Gno is not just about simplicity, but also about
safety. Reflection can be powerful, but it can also lead to code that is hard to
understand, hard to debug, and prone to runtime errors. By not supporting
reflection, Gno encourages you to write code that is explicit, clear, and easy
to understand.

When you're writing Gno code, remember: explicit is better than implicit, and
clear code is better than clever code.

## Gno good practices

### Contract-level access control

In Gno, it's a good practice to design your contract as an application with its
own access control. This means that different endpoints of your contract should
be accessible to different types of users, such as the public, admins, or
moderators.

The goal is usually to store the admin address or a list of addresses
(`address`) in a variable, and then create helper functions to update the
owners. These helper functions should check if the caller of a function is
whitelisted or not.

Let's deep dive into the different access control mechanisms we can use:

One strategy is to look at the caller with `cur.Previous()` on the `cur realm`
parameter of a crossing function. The caller could be the EOA (Externally
Owned Account), or the preceding realm in the call stack.

The account that signed the transaction, the EOA, is also available through
[`unsafe.OriginCaller()`](./gno-stdlibs.md#origincaller), whatever realms the
call passed through on the way to yours. Do not use it for access control. It is
Gno's `tx.origin`: a malicious realm called by the EOA can act as the EOA
towards your realm. Reserve it for recording who signed, such as in an event. A
function that must run only as a signer's direct call can enforce that with
[`runtime.AssertOriginCall()`](./gno-stdlibs.md#assertorigincall), which panics
unless a `maketx call` entered that function directly, so another named function
in between, or a `maketx run` script, makes it panic.

Here's an example:

```go
var admin address = "g1xxxxx"

func AdminOnlyFunction(cur realm) {
	caller := cur.Previous().Address()
	if caller != admin {
		panic("permission denied")
	}
	// ...
}

// func UpdateAdminAddress(cur realm, newAddr address) { /* ... */ }
```

In this example, `AdminOnlyFunction` reads its caller with
`cur.Previous().Address()`: another realm if one called it, otherwise the
user. If the caller is not the admin, it panics and the transaction stops.

Checking the immediate caller rather than the signer lets a contract own
assets, such as grc20 tokens or coins, and be called by other contracts
without putting the original caller's funds at risk. The default grc20
implementation works this way.

By using these access control mechanisms, you can ensure that your contract's
functionality is accessible only to the intended users, providing a secure and
reliable way to manage access to your contract.

For common needs, reuse the shared helpers listed in
[Community packages](./community-packages.md#access-control-helpers) rather
than rolling your own. `runtime.GetSessionInfo()` tells whether a session key
signed the transaction, so the realm can apply tighter limits to it.

### Design your realm as a public API

In Go, all your packages, including your dependencies, are typically treated as
part of your safe zone, similar to a secure perimeter. The boundary is drawn
between your program and the rest of the world, which means you secure the API
itself, potentially with authentication middlewares.

However, in Gno, your realm is the public API. It's exposed to the outside
world and can be accessed by other realms. Therefore, it's crucial to design
your realm with the same level of care and security considerations as you would
a public API.

One approach is to simulate a secure perimeter within your realm by having
private functions for the logic, and then writing your API layer by adding some
front-facing API with authentication. This way, you can control access to your
realm's functionality and ensure that only authorized callers can execute
certain operations.

```go
func PublicMethod(cur realm, nb int) {
	caller := cur.Previous().Address()
	privateMethod(caller, nb)
}

func privateMethod(caller address, nb int) { /* ... */ }
```

In this example, `PublicMethod` is a public function that can be called by other
realms. It retrieves the caller's address using `cur.Previous().Address()`, and
then passes it to `privateMethod`, which is a private function that performs the
actual logic. This way, `privateMethod` can only be called from within the
realm, and it can use the caller's address for authentication or authorization
checks.

### Never call a caller-supplied function under your own authority

A function declared in a realm runs under that realm's authority, and a
closure under the authority of the realm that created it, wherever either is
called from. A top-level function declared in a `/p/` package has no realm of
its own, so when your realm calls it, it runs as your realm and can rewrite
your state. A caller can hand you exactly such a function. So either never
invoke a callback or interface value a caller supplies, or give the callback
a parameter of a type your realm declares, which no `/p/` package can name.
The [security guide](./gno-security-guide.md#53-accepting-an-attacker-callback-under-your-own-authority)
covers the vector in full.

### Construct "safe" objects

A safe object in Gno is an object that is designed to be tamper-proof and
secure. It's created with the intent of preventing unauthorized access and
modifications. This follows the same principle of making a package an API, but
for a Gno object that can be directly referenced by other realms.

The goal is an object that other realms can hold and pass around, even by
pointer, without risk: every mutating method checks its own caller, so the object
protects itself wherever it is stored.

```go
type MySafeStruct struct {
	counter int
	admin address
}

// A /p/ package cannot declare realm-first-arg crossing functions, so the
// caller's realm is threaded as a non-first argument, the way p/nt/ownable does.
func NewSafeStruct(_ int, rlm realm) *MySafeStruct {
	if !rlm.IsCurrent() {
		panic("realm handle is not the live caller")
	}
	caller := rlm.Previous().Address()
	return &MySafeStruct{
		counter: 0,
		admin: caller,
	}
}

func (s *MySafeStruct) Counter() int { return s.counter }
func (s *MySafeStruct) Inc(_ int, rlm realm) {
	if !rlm.IsCurrent() {
		panic("realm handle is not the live caller")
	}
	caller := rlm.Previous().Address()
	if caller != s.admin {
		panic("permission denied")
	}
	s.counter++
}
```

Then, you can register this object in one or more other realms so that they can access it, while still following your own rules.

```go
import "gno.land/r/otherrealm"

func init(cur realm) {
	mySafeObj := NewSafeStruct(0, cur)
	otherrealm.Register(mySafeObj)
}

// then, other realm can call the public functions but won't be the "owner" of
// the object.
```

### Choosing between Coins and GRC20 tokens

In Gno, you've got two primary options: Coins or GRC20. Each option
has its unique advantages and disadvantages, and the ideal choice varies based
on individual requirements.

#### Coins

Coins are balances the chain keeps outside the GnoVM. A realm issues its own
denom through a [banker](./gno-stdlibs.md#banker). A plain bank transaction
moves coins, unless the chain's `restricted_denoms` parameter locks that denom,
and a bank query reads a balance; neither runs contract code. Their rules are
fixed by the chain, which makes them simple and predictable.

When you only need one balance, ask for it: `GetCoin(addr, denom)` reads a single
store key, while `GetCoins(addr)` reads every denom the address holds. That
distinction is not just an optimization. Anyone can send any address a new denom
without its consent, so `GetCoins` on a caller-supplied address costs whatever a
third party decided it should — enough of them and your function can no longer be
called at all.

#### Verifying inbound Coin payments

A realm that wants to charge for a function typically attaches a payment check
like this:

```go
import "chain/runtime/unsafe"

func BuyThing(cur realm, ...) {
	if !cur.Previous().IsUser() {   // BAD
		panic("must be called by a user")
	}
	if unsafe.OriginSend().AmountOf("ugnot") != price {
		panic("wrong payment amount")
	}
	// ... do the thing ...
}
```

This is **subtly unsafe**. `unsafe.OriginSend()` returns the coins attached to
the *original transaction*, not the coins actually received by this realm. If
anything runs between the tx origin and this realm's function, those coins may
have been consumed by the intermediary. Two attacker shapes bypass the check:

1. **Intermediate code realm.** User calls `r/attacker/wrapper.DoIt()` with
   `-send 1000000ugnot`. The wrapper keeps the coins (via its own banker) and
   then calls `BuyThing(cross(cur), ...)` on your realm. Your realm sees
   `OriginSend() = 1000000ugnot`, the `IsUser()` check passes because... actually
   it doesn't — `IsUser()` rejects pure code realms. Which leads to:

2. **User-run ephemeral realm (`maketx run`).** The attacker writes a short
   script and broadcasts it via `gnokey maketx run -send 1000000ugnot ...`.
   That script runs in an ephemeral code realm at path
   `gno.land/e/{attacker}/run`. Inside main, the script consumes the origin-send
   envelope (via its own `BankerTypeOriginSend`) or simply does whatever it
   wants with the coins, then calls `BuyThing(cross(cur), ...)`. Your realm sees
   `OriginSend() = 1000000ugnot` in the envelope and `IsUser() = true` because
   **`IsUser()` accepts both `IsUserCall()` (pure EOA) AND `IsUserRun()` (user-run
   ephemeral realm)**. The check passes but no coins reached your realm.

The fix is to use `IsUserCall()` instead of `IsUser()`:

```go
import "chain/runtime/unsafe"

func BuyThing(cur realm, ...) {
	if !cur.Previous().IsUserCall() {  // GOOD
		panic("must be called directly by an EOA (maketx call)")
	}
	if unsafe.OriginSend().AmountOf("ugnot") != price {
		panic("wrong payment amount")
	}
	// ... do the thing ...
}
```

`IsUserCall()` returns true only when `cur.Previous().PkgPath() == ""`, i.e.
the caller is a pure EOA. In that case the `-send` coins are guaranteed to
have landed at this realm's address, so `OriginSend()` and receipt agree.

Why the pairing matters: removing either check alone reopens the bypass.
`OriginSend()` without the EOA guard is lying about receipt. The EOA guard
without the amount check lets users pay nothing. Keep them together, commented
as a pair, and ideally cover the bypass with a regression test using
`testing.NewCodeRealm()` to simulate an intermediate attacker realm.

Alternatives considered:

- **`runtime.AssertOriginCall()`** — strictly enforces "direct MsgCall, no
  intermediaries, no MsgRun". Correct, but stricter than most realms want: it
  blocks all `maketx run` usage. Use it when you want to forbid MsgRun entirely
  (e.g. governance-only functions).

- **`banker.NewBanker(banker.BankerTypeOriginSend, cur)`** — creating this
  banker requires `cur.Previous().IsUserCall()`, so it implicitly asserts EOA.
  But it's a side-effectful assertion; if you don't need the banker itself,
  `IsUserCall()` is clearer.

- **Pulling coins from the caller** — **not possible**. Every
  `banker.SendCoins(from, to, amt)` requires `from == pkgAddr` (your own realm's
  address); there is no ERC-20-style `transferFrom`. Payment flow is push-only
  via `-send`. The `OriginSend` amount check + `IsUserCall` guard is the only
  pattern available.

#### GRC20 tokens

GRC20 tokens, on the other hand, are like Ethereum's ERC20 or CosmWasm's CW20.
They're flexible, composable, and perfect for DeFi protocols and DAOs. They
offer more features like token-gating, vaults, and wrapping.

For instance, if you're creating a voting system for a DAO, GRC20 tokens are
ideal. They're programmable, can be embedded in safe Gno objects, and offer more
control.

Every GRC20 transfer runs contract code, and the contract that defines the
token keeps control over its rules.

In the end, your choice depends on your needs: simplicity with Coins, or
flexibility and control with GRC20 tokens.

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

func MyBalance(cur realm) int64 {
	caller := cur.Previous().Address()
	return Token.BalanceOf(caller)
}
```

See also [`gno.land/r/demo/defi/foo20`](../../examples/gno.land/r/demo/defi/foo20).

#### Wrapping Coins

Want the best of both worlds? Consider wrapping your Coins. This gives
your coins the flexibility of GRC20 while keeping the security of Coins.
It's a bit more complex, but it's a powerful option that offers great
versatility.

See also [`gno.land/r/gnoland/wugnot`](../../examples/gno.land/r/gnoland/wugnot).

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
the agent whitelists and requests; a static feed is the only feed type it
ships so far.
gno.land has no built-in price feed, so a realm that moves funds based on a
fed value is only as secure as whoever provides that value: an attacker does
not need a bug in your code, only a bad number in the feed.

### Test the attacker, not just the happy path

The test kinds are covered in [the testing guide](./gno-testing.md);
`gno test` does not run benchmarks or `FuzzXxx` functions yet. What
realm tests add is the execution context: the `testing` package lets you call
your realm as someone else. `testing.NewUserRealm` and `testing.NewCodeRealm`
stand in for a user or another contract, `testing.SetRealm` and
`testing.SetOriginCaller` set the caller, `testing.IssueCoins` and
`testing.SkipHeights` set up funds and time. Use them to attack your own
realm: simulate an intermediary contract and prove your
[payment check](#verifying-inbound-coin-payments) cannot be bypassed.

### Choose storage types by access pattern

A `map` or slice is stored as one object, so reading or updating one element
loads or rewrites all of it, while a tree stores its nodes or leaf pages
separately, so touching one key loads only the path to it.
Keep maps and slices for small, bounded state, and put anything that grows in
a tree; [Gno data structures](./gno-data-structures.md#tree-backed-indexes)
compares the tree types.

### Define types and interfaces in pure packages (p/)

In Gno, it's common to create `p/NAMESPACE/DAPP` for defining types and
interfaces, and `r/NAMESPACE/DAPP` for the runtime, especially when the goal
for the realm is to become a standard that could be imported by `p/`.

The reason for this is that `p/` cannot import `r/`, while `r/` can import
anything. This separation allows you to define standards in `p/` that can be
used across multiple realms and packages.

In general, you can just write your `r/` to be an app. But if for some reason
you introduce a concept that can be reused, it makes sense to have a
dedicated `p/` so that people can re-use your logic without depending on
your realm's data.

For instance, if you want to create a token type in a realm, you can use it, and
other realms can import the realm and compose it. But if you want to create a
`p/` helper that will create a pattern, then you need to have your interface and
types defined in `p/` so anything can import it.

By separating your types and interfaces into `p/` and your runtime into `r/`,
you can create more modular, reusable, and standardized code in Gno. This
approach allows you to leverage the composability of Gno to build more powerful
and flexible applications.

### Emit Gno events to make life off-chain easier

Gno provides users the ability to log specific occurrences that happened in their
on-chain apps. An `event` log is stored in the ABCI results of each block, and
these logs can be indexed, filtered, and searched by external services, allowing
them to monitor the behaviour of on-chain apps.

It is good practice to emit events when any major action in your code is
triggered. For example, good times to emit an event are after a balance transfer,
ownership change, profile created, etc. Alternatively, you can view event emission
as a way to include data for monitoring purposes, given the indexable nature of
events.

Events consist of a type and a slice of strings representing `key:value` pairs.
They are emitted with the `Emit()` function, contained in the `chain` package in
the Gno standard library:

```go
package example

import "chain"

var owner address

func init(cur realm) {
	owner = cur.Previous().Address()
}

func ChangeOwner(cur realm, newOwner address) {
	caller := cur.Previous().Address()

	if caller != owner {
		panic("access denied")
	}

	owner = newOwner
	chain.Emit("OwnershipChange", "newOwner", newOwner.String())
}
```
If `ChangeOwner()` was called in, for example, block #43, getting the `BlockResults`
of block #43 will contain the following data:

```json
{
  "Events": [
	{
	  "@type": "/tm.Event",
	  "type": "OwnershipChange",
	  "pkg_path": "gno.land/r/demo/example",
	  "attrs": [
		{
		  "key": "newOwner",
		  "value": "g1zzqd6phlfx0a809vhmykg5c6m44ap9756s7cjj"
		}
	  ]
	}
	// other events
  ]
}
```

Read more about events [here](./gno-stdlibs.md#events).

### Package naming and organization

Your package name must match the last element of the package path (ignoring a
trailing `/vN` version suffix). This keeps imports clear and intuitive, avoiding
the need for named imports.

Ideally, package names should be short and human-readable. This makes it easier
for other developers to understand what your package does at a glance. Avoid
using abbreviations or acronyms unless they are widely understood.

Packages and realms can be organized into subdirectories. However, consider that the
best place for your main project will likely be `r/NAMESPACE/DAPP`, similar
to how repositories are organized on GitHub.

If you have multiple sublevels of realms, remember that they are actually
independent realms and won't share data. A good usage could be to have an
ecosystem of realms, where one realm is about storing the state, another one
about configuration, etc. But in general, a single realm makes sense.

You can also create small realms to create your ecosystem. For example, you
could centralize all the authentication for your whole company/organization in
`r/NAMESPACE/auth`, and then import it in all your contracts.

The `p/` prefix is different. In general, you should use top-level `p/` like
`p/NAMESPACE/DAPP` only for things you expect people to use. If your goal is
just to have internal libraries that you created to centralize your helpers and
don't expect that other people will use your helpers, then you should probably
use subdirectories like `p/NAMESPACE/DAPP/foo/bar/baz`.

Packages which contain `internal` as an element of the path (ie. at the end, or
in between, like `gno.land/p/demo/mypackage/internal`, or
`gno.land/p/demo/mypackage/internal/helpers`) can only be imported by packages
sharing the same root as the `internal` package. That is, given a package
structure as follows:

```
gno.land/p/demo/mypackage
├── utils
└── internal
	├── helpers
	└── crypto
```

The `mypackage/internal`, `mypackage/internal/helpers`, and `mypackage/internal/crypto`
packages can only be imported by `mypackage` and `mypackage/utils`.

This works for both realms and packages, and can be used to create entirely
restricted packages and realms that are not meant for outside consumption.

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
belongs in its `Render()`.

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
[`gno.land/p/moul/xmath`](../../examples/quarantined/gno.land/p/moul/xmath),
kept under `examples/quarantined` outside the audited set, works this way:
`generator.go` writes the same helpers once per numeric type into
`xmath.gen.gno`, standing in for generics. Readers on-chain see the final
source, so generated files stay auditable.
