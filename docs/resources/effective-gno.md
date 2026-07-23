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
realms due to their persistent characteristics. In other Gno code, such as
packages, the use of global variables is actually discouraged and may even be
completely disabled in the future. Instead, packages should use global
constants, which provide a safe and reliable way to define values that don't
change.

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
func Foobar(cur realm) {
	if !cur.IsCurrent() {
		panic("spoofed realm")
	}
	caller := cur.Previous().Address()
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
// Returns an error, so callers decide how to handle it.
func ParseAddress(s string) (address, error) {
	// ...
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
call site. A realm can call `MustParseAddress` to abort the transaction on bad
input, while a library can call `ParseAddress` and handle the error inline.

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
	- Note: global variables can often be set up just by assigning their initial value when you're declaring them. See below for an example! \
	  Deciding when to initialise the variable directly, and when to set it up in `init` can be non-straightforward. As a rule of thumb, though, `init` visually marks the code as executing only when the realm is started, while assigning the variables can be less straightforward.
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
	list	= []string{"foo", "bar", time.Now().Format("15:04:05")}
)

func init(cur realm) {
	created = time.Now()
	// cur.Previous() in the context of realm initialisation is,
	// of course, the publisher of the realm :)
	// This can be better than hardcoding an admin address as a constant.
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
reviewed, audited, and have billions of transactions under their belt, boasting
super high stability. This approach can make your code footprint smaller and more
reliable.

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
	"gno.land/p/finance/tokens"
	"gno.land/p/finance/exchange"
	"gno.land/p/finance/wallet"
	"gno.land/p/utils/permissions"
)

var (
	myWallet wallet.Wallet
	myToken tokens.Token
	myExchange exchange.Exchange
)

func init() {
	myWallet = wallet.NewWallet()
	myToken = tokens.NewToken("MyToken", "MTK")
	myExchange = exchange.NewExchange(myToken)
}

func BuyTokens(_ realm, amount int) {
	caller := permissions.GetCaller()
	permissions.CheckPermission(caller, "buy")
	myWallet.Debit(caller, amount)
	myExchange.Buy(caller, amount)
}

func SellTokens(_ realm, amount int) {
	caller := permissions.GetCaller()
	permissions.CheckPermission(caller, "sell")
	myWallet.Credit(caller, amount)
	myExchange.Sell(caller, amount)
}
```

##  When Gno takes Go practices to the next level

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

Here's an example from [grc20](https://staging.gno.land/p/demo/tokens/grc20$source&file=types.gno)
to illustrate the concept:

```go
// Teller interface defines the methods that a GRC20 token must implement. It
// extends the TokenMetadata interface to include methods for managing token
// transfers, allowances, and querying balances.
//
// The Teller interface is designed to ensure that any token adhering to this
// standard provides a consistent API for interacting with fungible tokens.
type Teller interface {
	// Returns the name of the token.
	GetName() string

	// Returns the symbol of the token, usually a shorter version of the
	// name.
	GetSymbol() string

	// Returns the decimals places of the token.
	GetDecimals() int

	// Returns the amount of tokens in existence.
	TotalSupply() int64

	// Returns the amount of tokens owned by `account`.
	BalanceOf(account address) int64

	// Moves `amount` tokens from the caller's account to `to`.
	//
	// Returns an error if the operation failed.
	Transfer(to address, amount int64) error

	// Returns the remaining number of tokens that `spender` will be
	// allowed to spend on behalf of `owner` through {transferFrom}. This is
	// zero by default.
	//
	// This value changes when {approve} or {transferFrom} are called.
	Allowance(owner, spender address) int64

	// Sets `amount` as the allowance of `spender` over the caller's tokens.
	//
	// Returns an error if the operation failed.
	//
	// IMPORTANT: Beware that changing an allowance with this method brings
	// the risk that someone may use both the old and the new allowance by
	// unfortunate transaction ordering. One possible solution to mitigate
	// this race condition is to first reduce the spender's allowance to 0
	// and set the desired value afterwards:
	// https://github.com/ethereum/EIPs/issues/20#issuecomment-263524729
	Approve(spender address, amount int64) error

	// Moves `amount` tokens from `from` to `to` using the
	// allowance mechanism. `amount` is then deducted from the caller's
	// allowance.
	//
	// Returns an error if the operation failed.
	TransferFrom(from, to address, amount int64) error
}
```

### Reflection is never clear

In Go, there's a well-known saying by Rob Pike: ["Reflection is never
clear."](https://www.youtube.com/watch?v=PAAkCSZUG1c&t=15m22s) This statement
emphasizes the complexity and potential pitfalls of using reflection in Go.

In Gno, reflection does not exist (yet). There are technical reasons for this,
but also a desire to create a Go alternative that is explicitly safer to use
than Go, with a smaller cognitive difficulty to read, discover, and understand.

The absence of reflection in Gno is not just about simplicity, but also about
safety. Reflection can be powerful, but it can also lead to code that is hard to
understand, hard to debug, and prone to runtime errors. By not supporting
reflection, Gno encourages you to write code that is explicit, clear, and easy
to understand.

We're currently in the process of considering whether to add reflection support
or not, or perhaps add it in a privileged mode for very few libraries. But for now,
when you're writing Gno code, remember: explicit is better than implicit, and
clear code is better than clever code.

## Gno good practices

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

### Define types and interfaces in pure packages (p/)

In Gno, it's common to create `p/NAMESPACE/DAPP` for defining types and
interfaces, and `r/NAMESPACE/DAPP` for the runtime, especially when the goal
for the realm is to become a standard that could be imported by `p/`.

The reason for this is that `p/` can only import `p/`, while `r/` can import
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
	if !cur.IsCurrent() {
		panic("spoofed realm")
	}
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
package events

import "chain"

var owner address

func init(cur realm) {
	owner = cur.Previous().Address()
}

func ChangeOwner(cur realm, newOwner address) {
	if !cur.IsCurrent() {
		panic("spoofed realm")
	}
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
	  "@type": "/tm.gnoEvent",
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

Another approach is to look specifically at the EOA. For this, you can call
[`unsafe.OriginCaller()`](./gno-stdlibs.md#origincaller), which returns the
public address of the account that signed the transaction. Internally, this
call walks the frame stack, the stack of callers including all the functions,
anonymous functions, and other realms, and takes the initial caller.

Do not use `unsafe.OriginCaller()` for access control. It is gno's
`tx.origin`: a malicious realm called by the EOA can act as the EOA towards
your realm. Reserve it for cases that intentionally want tx-level identity,
such as event emission or fee attribution, and pair it with
[`runtime.AssertOriginCall()`](./gno-stdlibs.md#assertorigincall) so misuse
panics.

Here's an example:

```go
var admin address = "g1xxxxx"

func AdminOnlyFunction(cur realm) {
	if !cur.IsCurrent() {
		panic("spoofed realm")
	}
	caller := cur.Previous().Address()
	if caller != admin {
		panic("permission denied")
	}
	// ...
}

// func UpdateAdminAddress(cur realm, newAddr address) { /* ... */ }
```

In this example, `AdminOnlyFunction` is a function that can only be called by
the admin. It retrieves the caller's address using `cur.Previous().Address()`,
this can be either another realm contract, or the calling user if there is no
other intermediary realm. and then checks if the caller is the admin. If not, it
panics and stops the execution.

The goal of this approach is to allow a contract to own assets (like grc20 or
coins), so that you can create contracts that can be called by another
contract, reducing the risk of stealing money from the original caller. This is
the behavior of the default grc20 implementation.

Here's an example:

```go
func TransferTokens(cur realm, to address, amount int64) {
	if !cur.IsCurrent() {
		panic("spoofed realm")
	}
	caller := cur.Previous().Address()
	if caller != admin {
		panic("permission denied")
	}
	// ...
}
```

In this example, `TransferTokens` is a function that can only be called by the
admin. It retrieves the caller's address using `cur.Previous().Address()`, and
then checks if the caller is the admin. If not, the function panics and execution is stopped.

By using these access control mechanisms, you can ensure that your contract's
functionality is accessible only to the intended users, providing a secure and
reliable way to manage access to your contract.

### Prefer avl.Tree over map for scalable storage

An `avl.Tree` works like a `map` for storing key-value pairs. `maps` store all
entries in one object (accessing any value loads everything), while AVL trees
store each node separately (accessing a value loads only the search path).
This makes `avl.Tree` significantly more efficient in both gas usage and
runtime performance for large or growing datasets.

**Key differences**:

- **AVL Trees**: O(log n) lookup, lazy loading, iterate in **sorted key order**.
- **Maps**: O(1) lookup, type safety, iterate in **unspecified order**.

**Use `avl.Tree` when you need**:

- Lazy loading (efficient for large datasets - only loads the search path)
- Efficient range queries (find all keys between "bob" and "charlie")
- Data that grows over time (user registries, leaderboards)
- Sorted iteration by key value

**Use `map` when you need**:

- O(1) fast lookups
- Small bounded datasets (e.g.: configuration values)
- Type safety (AVL values are `any` and require type assertions)

```go
// Map example
users := make(map[string]User)
users["bob"] = User{}
users["alice"] = User{}
for name := range users { // unspecified order
	// ...
}
user := users["alice"] // O(1) direct access

// AVL example
var users avl.Tree
users.Set("bob", &User{})
users.Set("alice", &User{})
users.Set("charlie", &User{})

// Iterate all users (sorted alphabetically)
users.Iterate("", "", func(name string, value any) bool {
	// Order: alice, bob, charlie (sorted by key)
	user := value.(*User) // Type assertion required - values are any
	return false // return true to stop iteration
})

// Range query: get users from "bob" (inclusive) to "charlie" (exclusive)
// This is O(log n + k) where k = results in range
users.Iterate("bob", "charlie", func(name string, value any) bool {
	// Only visits: bob (end is exclusive)
	user := value.(*User) 
	return false
})

// Get a specific user (O(log n))
// Get returns nil if the key does not exist
value := users.Get("alice")
if value == nil {
	return nil
}
return value.(*User)

// Check if a key exists without retrieving the value
if users.Has("alice") {
	// key exists
}

// Multi-index example - search the same data in different ways
var (
	usersById   avl.Tree // Find user by ID
	usersByName avl.Tree // Find user by name
)

func AddUser(id, name string) {
	usersById.Set(id, name)     // Can search by ID
	usersByName.Set(name, id)   // Can search by name
}
```

For a detailed explanation of how AVL trees are stored in Gno's object store, see the [avl package README](../../examples/gno.land/p/nt/avl/v0/README.md).

### Construct "safe" objects

A safe object in Gno is an object that is designed to be tamper-proof and
secure. It's created with the intent of preventing unauthorized access and
modifications. This follows the same principle of making a package an API, but
for a Gno object that can be directly referenced by other realms.

The goal is to create an object which, once instantiated, can be linked and its
pointer can be "stored" by other realms without issue, because it protects its
usage completely.

```go
type MySafeStruct struct {
	counter int
	admin address
}

func NewSafeStruct(cur realm) *MySafeStruct {
	if !cur.IsCurrent() {
		panic("spoofed realm")
	}
	caller := cur.Previous().Address()
	return &MySafeStruct{
		counter: 0,
		admin: caller,
	}
}

func (s *MySafeStruct) Counter() int { return s.counter }
func (s *MySafeStruct) Inc(cur realm) {
	if !cur.IsCurrent() {
		panic("spoofed realm")
	}
	caller := cur.Previous().Address()
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
	mySafeObj := NewSafeStruct(cur)
	otherrealm.Register(cross(cur), mySafeObj)
}

// then, other realm can call the public functions but won't be the "owner" of
// the object.
```

### Choosing between Coins and GRC20 tokens

In Gno, you've got two primary options: Coins or GRC20. Each option
has its unique advantages and disadvantages, and the ideal choice varies based
on individual requirements.

#### Coins

Coins are managed by the banker module, separate from GnoVM. They're
simple, strict, and secure. You can create, transfer, and check balances with an
RPC call, no GnoVM needed.

For example, if you're creating a coin for cross-chain transfers, Coins
are your best bet. They're IBC-ready and their strict rules offer top-notch
security.

Read about how to use the Banker module [here](./gno-stdlibs.md#banker).

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
  intermediaries, no MsgRun". Correct, but stricter than most realms want:
  rejects `testing.NewUserRealm`-based unit tests in some configurations and
  blocks all `maketx run` usage. Use it when you want to forbid MsgRun entirely
  (e.g. governance-only functions).

- **`banker.NewBanker(banker.BankerTypeOriginSend, cur)`** — creating this
  banker requires `cur.Previous().IsUserCall()`, so it implicitly asserts EOA.
  But it's a side-effectful assertion; if you don't need the banker itself,
  `IsUserCall()` is clearer.

- **Pulling coins from the caller** — **not possible** in current gno. Every
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

Remember, GRC20 tokens are more gas-intensive and aren't IBC-ready yet. They
also come with shared ownership, meaning the contract retains some control.

In the end, your choice depends on your needs: simplicity and security with
Coins, or flexibility and control with GRC20 tokens. And if you want the
best of both worlds, you can wrap a Coins into a GRC20 compatible token.

```go
import "gno.land/p/demo/tokens/grc20"

var (
	Token         *grc20.Token
	privateLedger *grc20.PrivateLedger
	UserTeller    grc20.Teller
)

func init(cur realm) {
	// This realm only ever creates this one token, so id 0 can't collide.
	Token, privateLedger = grc20.NewToken("Foo Token", "FOO", 4, 0, cur)
	UserTeller = Token.CallerTeller()
}

func MyBalance(cur realm) int64 {
	caller := cur.Previous().Address()
	return UserTeller.BalanceOf(caller)
}
```

See also: https://staging.gno.land/r/demo/defi/foo20

#### Wrapping Coins

Want the best of both worlds? Consider wrapping your Coins. This gives
your coins the flexibility of GRC20 while keeping the security of Coins.
It's a bit more complex, but it's a powerful option that offers great
versatility.

See also: https://github.com/gnolang/gno/tree/master/examples/gno.land/r/gnoland/wugnot

### Suggested file names and layout

Split a realm into predictably named files. The source is published and
browsable on-chain, so file names are part of your public interface. There is
no enforced layout, but a few names recur across the standard examples and are
worth adopting:

- `<realm>.gno`: the package declaration and main entrypoints, named after the
  realm itself (the counter realm uses `counter.gno`).
- `types.gno`: type and interface definitions.
- `render.gno`: the `Render()` function and its formatting helpers.
- `admin.gno`: ownership and privileged endpoints.
- `errors.gno`: sentinel errors and error constructors.
- `doc.gno`: a package-level doc comment when the overview is long.
- `gnomod.toml`: module metadata, always present.
- `README.md`: optional, rendered alongside the source.

For example, [`gno.land/r/gnoland/blog`](../../examples/gno.land/r/gnoland/blog)
splits its logic into `gnoblog.gno` (the
package and main API), `admin.gno` (permissions), and `util.gno`, with tests
beside the files they cover. Keep each file focused on one concern: a reader
should be able to guess a file's contents from its name alone.

### Versioning and upgrades

A published package or realm is immutable: the blockchain will not let anyone
overwrite it. This is a feature, not a limitation. Importers get a permanent,
auditable target, and a dependency cannot change under you. The flip side is
that you cannot edit a deployed contract in place; instead, you publish a new
version at a new path.

The convention is a `/vN` suffix on the package path, like
`gno.land/p/nt/avl/v0` or `gno.land/r/sys/validators/v2` and `.../v3`. The
version is part of the path, not a separate field in `gnomod.toml`, and the
package name ignores the trailing `/vN` (so `.../validators/v3` is still
`package validators`). Versions coexist: old importers keep working against the
exact version they pinned, while new code imports the newer one.

Because a realm cannot be mutated, "upgrading" means deploying a new version and
moving callers to it. A common pattern is a small proxy or registry realm at a
stable path that forwards to the current implementation and can be repointed by
an admin or DAO, keeping previous versions live for rollback. `gno.land/r/gov/dao`
works this way: it delegates to a versioned implementation and tracks a set of
allowed DAOs so a buggy upgrade can be rolled back to a prior version. For a
simpler case, `gno.land/r/sys/validators` simply keeps `v2` and `v3` side by
side.

Not every package needs this. A small, focused, well-tested `p/` library can
reach a "finished" state where it simply does its job and never needs a new
version. Aim for that: stable bricks others can build on without surprises.

### Call other realms with `cross`

Realms are composable: one realm can import another and call its functions, the
same way you import a `p/` package. The difference is the realm boundary. A
function meant to be invoked from another realm is a *crossing function*,
declared with a leading `cur realm` parameter:

```go
// in gno.land/r/some/registry
func Register(cur realm, id string, h Handler) { /* ... */ }
```

To actually cross into that realm, wrap your own `cur` in `cross(...)` and
pass it as the first argument; the
[`init()` example above](#understand-the-importance-of-init) does exactly this
with `registry.Register(cross(cur), "myID", myCallback)`.

When you cross, `cur.Previous()` inside the callee is *your* realm, not the
original user. That shift is what lets a contract act on a caller's behalf,
and it is the basis for realms that hold and move assets safely. Before
trusting `cur.Previous()`, check `cur.IsCurrent()`: it verifies the handle is
the live crossing frame's own `cur` and not a stale or smuggled one. Within
your own realm, passing `cur` directly instead of `cross(cur)` runs the
function without crossing, so the realm context does not shift.
Getting this distinction right is essential for access control, so read
[the interrealm specification](./gno-interrealm.md) before writing cross-realm
code: do not reason about callers using Solidity's `msg.sender` intuition.

### Reuse access control instead of rolling your own

The [contract-level access control](#contract-level-access-control) pattern above
is easy to get subtly wrong, and getting it wrong usually means anyone can call
your admin functions. For common needs, prefer the shared building blocks in
`p/` rather than reimplementing them:

- `gno.land/p/nt/ownable/v0`: single-owner access control, with transferable and
  droppable ownership.
- `gno.land/p/moul/authz`: flexible authorization: member lists, contract- or
  DAO-backed authority, and auto-accept strategies behind one interface.
- `gno.land/p/nt/commondao/v0`: a minimal framework for DAOs, with proposals,
  voting, and member groups, for when a single admin is not enough.

For an emergency stop, build a pausable switch on top of `ownable`: an
owner-gated `paused` flag checked at the top of sensitive endpoints.

These packages get the caller check right for you: the methods that change who
is in control verify the caller's realm themselves, so an attacker cannot pose
as the owner. Reusing them means less code to audit and one fewer place to
make a security mistake. See each package's doc comments and tests for the
exact calling convention.

### Respect determinism: time, randomness, and ordering

Every validator must execute your code and reach exactly the same result, so the
GnoVM is strictly deterministic. A few Go habits change as a result.

**Time.** `time.Now()` returns the *block* time, not the machine's wall clock,
and every call within the same block returns the same instant. It is safe for
timestamps and deadlines, but expect no real-world precision or sub-block
resolution. Timers and tickers do not exist.

**Randomness.** There is no secure source of randomness on-chain. `math/rand` is
available, but it is a deterministic pseudo-random generator: its own
documentation warns that its outputs may be easily predictable regardless of how
it is seeded. Any seed you can compute inside a contract, block time, height,
or caller addresses, is equally known to validators and users. They can
predict the outcome, or manipulate it by ordering transactions.
Use on-chain randomness only for cosmetic or non-adversarial purposes.
For anything an attacker could profit from biasing, such as a lottery with real
stakes, use a commit-reveal scheme or an external source instead.

**Ordering.** Unlike Go, where map iteration order is randomized, Gno iterates
maps in insertion order. This is deterministic, but do not depend on it for
correctness: Go does not promise it, and insertion order is rarely the order
you actually want. When you need ordered
iteration, use an [`avl.Tree`](#prefer-avltree-over-map-for-scalable-storage),
which iterates in sorted key order. And of course there are no goroutines or
channels: a transaction runs as a single deterministic thread.

### Know what the Gno standard library gives you

Gno ships a curated subset of Go's standard library, plus a few chain-specific
packages of its own. Knowing the boundary saves you from reaching for something
that isn't there.

Most pure, deterministic Go packages are available and behave as you would
expect: `strings`, `strconv`, `bytes`, `bufio`, `sort`, `errors`, `math` and
`math/bits`, `regexp`, `unicode`, `html`, `path`, `net/url`, and the common
`encoding/*` codecs like `base64`, `hex`, and `binary`. `time` is present but
returns block time, as described above.

Anything that touches the outside world or breaks determinism is deliberately
absent: `os`, `net/http`, `syscall`, `sync`, `unsafe`, `crypto/rand`, and
`reflect` (see [Reflection is never
clear](#reflection-is-never-clear)). `fmt` is absent too; use
`gno.land/p/nt/ufmt/v0` for formatting.

In their place, Gno adds the `chain` family, which has no Go equivalent:

- `chain`: emit events with `Emit`, plus the `Coin`/`Coins` types and address
  helpers.
- `chain/runtime`: chain context: chain ID and domain, block height, the
  `Realm` type, origin-call assertion, and session info.
- `chain/banker`: create, send, and inspect native coin balances.
- `chain/params`: set realm-scoped on-chain parameters (set-only).

These are what the older `std` package became. You can explore any of them from
the command line with `gno doc <pkg>`, and they are documented in
[the standard library reference](./gno-stdlibs.md).

### Test your realms

Gno mirrors Go's testing model and adds a few chain-aware tools. There are four
kinds of test worth knowing:

- **Unit tests** in `*_test.gno`, with the usual `func TestXxx(t *testing.T)`.
  Run them with `gno test ./...`.
- **Example tests**, `func ExampleXxx()` with an `// Output:` comment. They
  double as documentation, since the expected output lives next to the code.
- **Filetests** in `*_filetest.gno`: a `main` program checked against directives
  like `// Output:`, `// Error:`, `// Events:`, and `// Realm:`. They are ideal
  for asserting exact output, a specific panic, or emitted events for a whole
  program.
- **txtar integration tests** under `gno.land/pkg/integration/testdata/`, which
  run end to end against a real node, driving `gnoland` and `gnokey` and
  asserting on their output.

For realm logic, the `testing` package adds helpers that shape the execution
context: `testing.NewUserRealm` and `testing.NewCodeRealm` to stand in for an EOA
or another contract, `testing.SetOriginCaller` and `testing.SetRealm` to control
who is calling, `testing.IssueCoins` to fund an address, and `testing.SkipHeights`
to advance blocks. Use them to cover access control and payment paths: for
example, simulate an intermediary realm with `testing.NewCodeRealm` to prove
your [origin-send payment check](#verifying-inbound-coin-payments) cannot be
bypassed.

One gap to be aware of: `gno test` does not yet discover `BenchmarkXxx` or
`FuzzXxx` functions. The `testing` package ships a small fuzzing helper
(`testing.F`) you can drive from a regular unit test, but there is no
`go test -fuzz`-style runner.

### Develop locally with gnodev

For the inner development loop, use `gnodev` (in `contribs/gnodev`). It runs
an in-memory gno.land node and web frontend with hot reload. Point it at your
package directory: it deploys your code alongside the standard `examples/`,
pre-funds a local key, and redeploys on every save, replaying previous
transactions. Edit, save, refresh, repeat, with no manual node setup.

```bash
gnodev ./path/to/your/realm
```

Where each tool fits:

- `gno test ./...` is your fastest feedback and should drive day-to-day
  development. Keep tests beside your code and run them constantly.
- `gnoland start` runs a full standalone node, useful when you need real genesis,
  persistence, or multi-node behavior.
- A shared **testnet** lets others interact with your realm before mainnet. The
  `staging` network and the rolling `testN` testnets are reachable over RPC and
  have a faucet for test coins.

### Script transactions with `maketx run`

A `maketx call` invokes a single function. When you need more, `gnokey maketx
run` executes an entire Gno script, a `package main` with a `main` function, as
one transaction. Think of it as using Gno like a shell script against the chain:
you can import realms, make several calls, loop, and compute intermediate
values, all atomically.

```go
package main

import "gno.land/r/demo/counter"

func main(cur realm) {
	for i := 0; i < 5; i++ {
		counter.Increment(cross(cur))
	}
}
```

```bash
gnokey maketx run -gas-fee 1000000ugnot -gas-wanted 20000000 mykey ./script.gno
```

This is handy for multi-step admin tasks, data migrations, or composing several
realm calls into one transaction. One caveat: a `run` script does not execute as
a plain account. It runs in an ephemeral realm at `gno.land/e/{youraddr}/run`, so
a realm that charges for a function must guard with `IsUserCall()` rather than
`IsUser()`, as explained in [Verifying inbound Coin
payments](#verifying-inbound-coin-payments). The ephemeral realm can otherwise
import and call any realm just like regular contract code.

### Ship more than code

A realm is the on-chain core of your project, but a finished product usually
ships more around it. Because the source and `Render()` output are public and
user-facing, treat the realm itself as part of the product: write a `Render()`
for end users, include a `README.md`, and keep doc comments aimed at readers, as
covered in [Documentation is for users](#documentation-is-for-users).

Off-chain, you will often pair the realm with a client that calls it (a CLI or a
web frontend), indexers that consume the
[events](#emit-gno-events-to-make-life-off-chain-easier) you emit, and static
assets. These live outside the contract but are part of how people actually use
it.

Keep the two audiences distinct. A `p/` package is for developers, much like
an NPM library: give it a clean API, stable versions, and documentation for
integrators. A realm is for end users, more like an app in an app store: give
it clear endpoints and a `Render()` that explains itself.

### Write gas-conscious code

Every statement your realm executes and every byte it stores costs gas, and
storage dominates: a write costs 2,000 gas plus 30 gas per byte, ten times the
per-byte price of a read. The most effective optimization is almost always to
store less, not to compute less.

Habits that keep gas down:

- Store only what the contract needs to enforce its rules. Derived values,
  formatted strings, and anything a client can recompute belong off-chain or
  in `Render()`.
- Avoid unbounded loops over growing state. A function that walks every user
  gets more expensive every day, and one day it no longer fits in the block
  gas limit. Iterate over bounded ranges instead, using
  [avl.Tree](#prefer-avltree-over-map-for-scalable-storage) range iteration
  and pagination.
- Build strings with `strings.Builder` instead of `+=` in a loop: every
  concatenation allocates a new string.
- Precompute off-chain whatever can be passed in as an argument.

Execution gas is not the whole storage price: bytes that stay persisted also
lock a refundable GNOT [storage deposit](./storage-deposit.md), returned when
the data is deleted. Deleting state you no longer need literally pays.

Measure before optimizing: `gnokey maketx ... -simulate only` reports the gas
a call would use, and [gas fees](./gas-fees.md) documents pricing, estimation,
and the storage price split in detail.

### Choose data structures by what they touch in storage

In a realm, a data structure's real cost is how many stored objects an
operation touches, because modified objects are written back to storage at the
end of the transaction. That is why
[avl.Tree beats map](#prefer-avltree-over-map-for-scalable-storage) for large
collections: a `Set` rewrites one path of nodes, while a map or slice is
persisted as one big object. A few more building blocks are worth knowing:

- `gno.land/p/nt/bptree/v0` implements the same interface as `avl.Tree` with a
  B+ tree: configurable fanout, fewer nodes touched per operation, and cheaper
  in-order iteration over large datasets.
- `gno.land/p/nt/seqid/v0` produces sequential IDs that sort correctly as
  `avl.Tree` keys, which gives you insertion-ordered listings and clean
  pagination for free.
- Small, bounded collections do not need any of this: below a few dozen
  entries a plain slice or map costs less than the tree machinery.

[Gno data structures](./gno-data-structures.md) covers the basics of each
container type.

### Model workflows as state machines

When a realm walks users through a process, whether a game, an auction, or a
vesting schedule, model the process as an explicit state machine: one enum
field holds the current phase, and every entrypoint first asserts the phase it
expects.

```go
type Phase byte

const (
	PhaseLobby Phase = iota
	PhasePlaying
	PhaseFinished
)

type Game struct {
	phase   Phase
	players [2]address
	moves   int
}

func Play(cur realm, gameID string, cell int) {
	if !cur.IsCurrent() {
		panic("spoofed realm")
	}
	g := mustGetGame(gameID)
	if g.phase != PhasePlaying {
		panic("game is not in progress")
	}
	caller := cur.Previous().Address()
	if caller != g.players[g.moves%2] {
		panic("not your turn")
	}
	// ...apply the move, then maybe transition:
	// g.phase = PhaseFinished
}
```

The single phase field is the source of truth: transitions happen only inside
entrypoints, after validation, so an invalid sequence of calls can never leave
the state half-updated. Derive what you can instead of storing it, like whose
turn it is from the move count above, so no second copy can drift out of sync.
Terminal phases reject every mutating call.
[`gno.land/r/demo/defi/atomicswap`](../../examples/gno.land/r/demo/defi/atomicswap)
derives a swap's status from its deadline and flags the same way.

### Keep governance flexible

Hardcoding one admin address works for a prototype and then hurts: the realm
cannot change operators, share control, or decentralize without a redeploy.
Put "who decides" behind a value you can swap instead of a constant you
cannot.

`gno.land/p/moul/authz` models exactly this: your realm asks an `Authorizer`
whether an action is allowed, and the authority behind it can grow from a
single admin to a member list to a full DAO vote without touching the rest of
the code. When you need real proposals and voting, reach for `commondao`,
introduced above. And at the chain level, `gno.land/r/gov/dao` shows the same
idea applied to governance itself, as described in
[Versioning and upgrades](#versioning-and-upgrades): a stable proxy realm
whose implementation can be voted in and rolled back.

Start with the smallest authority that works, but reach for these packages the
moment more than one person should hold the keys.

### Charge for access with a subscription

A subscription is a small pattern on top of
[payment verification](#verifying-inbound-coin-payments): charge once, record
when the payment expires, and gate paid endpoints on that record.

```go
var subs avl.Tree // address -> expiry (time.Time)

func Subscribe(cur realm) {
	if !cur.IsCurrent() {
		panic("spoofed realm")
	}
	// Verify the sent coins as shown in "Verifying inbound Coin payments".
	caller := cur.Previous().Address()
	subs.Set(caller.String(), time.Now().Add(30*24*time.Hour))
}

func assertSubscribed(caller address) {
	v := subs.Get(caller.String())
	if v == nil || time.Now().After(v.(time.Time)) {
		panic("no active subscription")
	}
}
```

Every paid entrypoint calls `assertSubscribed(cur.Previous().Address())` first.
Variations fall out naturally: store a bool instead of an expiry for a lifetime
subscription, extend the current expiry instead of replacing it for renewals,
or let a caller pay for a different address to gift a subscription. Block time
is all you need for expiry checks, so the whole pattern stays a few lines.

### Bring off-chain data on-chain with oracles

A realm cannot fetch anything: no HTTP, no files, no external reads. Off-chain
data enters the chain only when someone sends a transaction carrying it, so an
oracle in gno.land is an agreement between a realm and off-chain agents it
chooses to trust.

The
[gnorkle](https://github.com/gnolang/gno/tree/master/examples/gno.land/p/demo/gnorkle)
framework (`gno.land/p/demo/gnorkle/gnorkle`) structures that agreement. A
realm registers *feeds*, each describing *tasks* for agents to perform. An
agent polls the realm for pending tasks, does the off-chain work, and pushes
the result back through an entrypoint; an *ingester* validates and commits the
value, and a whitelist controls which agents may provide it.

[ghverify](https://github.com/gnolang/gno/tree/master/examples/gno.land/r/gnoland/ghverify)
(`gno.land/r/gnoland/ghverify`) is a complete deployed example: a user
requests verification of their GitHub handle, and an off-chain agent checks
that the handle controls a repository containing the user's address and pushes
the result back. The realm then serves the verified handle-to-address mapping
on-chain.

The trust model is explicit: the chain never verifies the off-chain fact, only
that a whitelisted agent attested to it. Choose your agents accordingly, and
remember there is no built-in price feed; for anything financial, the oracle
is the weakest link.

### Treat forking as a feature

Every realm and package is published source, and the
[license](https://gno.land/license) plus the
[project constitution](../CONSTITUTION.md) explicitly
allow copying and redeploying it. If a contract no longer serves its users, or
you need a variant its author will not merge, deploy your modified copy at a
path you control. Since deployed code is immutable, forking is the escape
hatch: users are never trapped in an abandoned or hostile contract.

The examples tree already works this way. `gno.land/p/onbloc/uint256` is a
port of the [holiman/uint256](https://github.com/holiman/uint256) Go library
and ships the upstream BSD license alongside the code;
`gno.land/p/nt/cford32/v0` adapts
Go's `encoding/base32`. Do the same when you fork: keep the upstream license
file and state where the code came from, so provenance stays auditable
on-chain.

The flip side was stated at the top of this document: design your packages to
be flexible so users are not forced to fork them. A fork that exists because
your API was too rigid is a bug report you can no longer merge.

### Generate repetitive code off-chain

Gno does not have generics yet, and there is no on-chain `go generate`: what
you deploy is plain `.gno` source. Nothing stops you from generating that
source before deployment, though. If you catch yourself writing the same typed
wrapper around `avl.Tree` for the third time, write a small Go program that
emits it and commit the output like any other file.

The repository already works this way in places: `gno.land/p/onbloc/uint256`
ships lookup tables generated by Go tooling, and the documentation uses
[embedmd](https://github.com/campoy/embedmd) (`make -C docs generate`) to copy
real, compiling source files into markdown instead of hand-maintaining
snippets. Generated files stay reviewable
and auditable on-chain, since readers see the final source, not the generator.

### Pass functions across realms

Function values cross realm boundaries freely: a realm can accept a callback,
store it, and call it later. What does not travel with them is authority.
A stored function writes state under its home realm: the realm that declared
it, or, for a closure built by `p/` code, the realm that created it. It never
writes under the realm that happens to invoke it. So holding another realm's
callback does not let you write its state, and a callback smuggled into your
realm cannot write yours. The runtime enforces this with the borrowing rules
described in [the interrealm specification](./gno-interrealm.md).

Invoking a stored callback without `cross` also leaves the realm context
alone: your realm stays the current realm, and the callee cannot tell the
difference. The dangerous move is the opposite one: cross-calling a
caller-supplied function value with `cross(cur)` hands the callee your agency,
because it will see your realm as its caller. Never cross-call an arbitrary
submitted function; accept only callbacks you have reviewed or whitelisted.

Two practical rules follow. Ownership handles are addresses, not functions:
`ownable` stores the owner's address. Storing a realm value is not even an
option, since realm values cannot be persisted at all. And in a realm, never
expose a function-typed package variable whose type takes a `realm` parameter;
it invites exactly the cross-call above.

### Know what the frame stack can tell you

`cur.Previous()` on a crossing function's `cur realm` parameter is the everyday
tool: the unforgeable identity of whoever crossed into you. The frame stack
offers a few sharper instruments for specific jobs:

- `runtime.AssertOriginCall()` panics unless the call is a direct, top-level
  `maketx call`. Use it for endpoints that must be invoked by a signing user
  directly, never through another contract or a `maketx run` script.
- `cur.Previous().IsUserCall()` and `IsUserRun()` distinguish a plain user
  call from a `maketx run` ephemeral realm; the payment rules in
  [Verifying inbound Coin payments](#verifying-inbound-coin-payments) hinge on
  this distinction.
- `runtime.GetSessionInfo()` reports whether the transaction was signed with a
  session key, so a realm can apply tighter limits to delegated sessions.
- `chain/runtime/unsafe` holds the raw stack walkers: `CurrentRealm()`,
  `PreviousRealm()`, `OriginCaller()`, `OriginSend()`. The package is named
  `unsafe` deliberately: called from a non-crossing helper,
  `unsafe.PreviousRealm()` answers relative to the last crossing, not the
  helper's own caller, and `OriginCaller()` is the `tx.origin` covered in
  [contract-level access control](#contract-level-access-control).

If a check matters for security, express it through `cur`; reach into the
frame stack only when the question is genuinely about the transaction, not
the caller.
