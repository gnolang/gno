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
import "std"

func Foobar() {
	caller := std.PreviousRealm().Address()
	if caller != "g1xxxxx" {
		panic("permission denied")
	}
	// ...
}
```

<!-- TODO: suggest MustXXX and AssertXXX flows in p/ -->

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

func init() {
	registry.Register(cross, "myID", myCallback)
}

func myCallback(a, b string) { /* ... */ }
```

A common use case could be to set the "admin" as the caller uploading the
package.

```go
import (
	"std"
	"time"
)

var (
	created time.Time
	admin   std.Address
	list	= []string{"foo", "bar", time.Now().Format("15:04:05")}
)

func init() {
	created = time.Now()
	// std.OriginCaller in the context of realm initialisation is,
	// of course, the publisher of the realm :)
	// This can be better than hardcoding an admin address as a constant.
	admin = std.OriginCaller()
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

Here's an example from [grc20](https://gno.land/p/demo/tokens/grc20$source&file=types.gno)
to illustrate the concept:

```go
// Teller interface defines the methods that a GRC20 token must implement. It
// extends the TokenMetadata interface to include methods for managing token
// transfers, allowances, and querying balances.
//
// The Teller interface is designed to ensure that any token adhering to this
// standard provides a consistent API for interacting with fungible tokens.
type Teller interface {
	exts.TokenMetadata

	// Returns the amount of tokens in existence.
	TotalSupply() uint64

	// Returns the amount of tokens owned by `account`.
	BalanceOf(account std.Address) uint64

	// Moves `amount` tokens from the caller's account to `to`.
	//
	// Returns an error if the operation failed.
	Transfer(to std.Address, amount uint64) error

	// Returns the remaining number of tokens that `spender` will be
	// allowed to spend on behalf of `owner` through {transferFrom}. This is
	// zero by default.
	//
	// This value changes when {approve} or {transferFrom} are called.
	Allowance(owner, spender std.Address) uint64

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
	Approve(spender std.Address, amount uint64) error

	// Moves `amount` tokens from `from` to `to` using the
	// allowance mechanism. `amount` is then deducted from the caller's
	// allowance.
	//
	// Returns an error if the operation failed.
	TransferFrom(from, to std.Address, amount uint64) error
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

Your package name should match the folder name. This helps to prevent having
named imports, which can make your code more difficult to understand and
maintain. By matching the package name with the folder name, you can ensure that
your imports are clear and intuitive.

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
in between, like `gno.land/p/nt/seqid/internal`, or
`gno.land/p/nt/seqid/internal/base32`) can only be imported by packages
sharing the same root as the `internal` package. That is, given a package
structure as follows:

```
gno.land/p/nt/seqid
├── generator
└── internal
	├── base32
	└── cford32
```

The `seqid/internal`, `seqid/internal/base32`, and `seqid/internal/cford32`
packages can only be imported by `seqid` and `seqid/generator`.

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
import "std"

func PublicMethod(nb int) {
	caller := std.PreviousRealm().Address()
	privateMethod(caller, nb)
}

func privateMethod(caller std.Address, nb int) { /* ... */ }
```

In this example, `PublicMethod` is a public function that can be called by other
realms. It retrieves the caller's address using `std.PreviousRealm().Address()`, and
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
They are emitted with the `Emit()` function, contained in the `std` package in
the Gno standard library:

```go
package events

import (
	"std"
)

var owner std.Address

func init() {
	owner = std.PreviousRealm().Address()
}

func ChangeOwner(_ realm, newOwner std.Address) {
	caller := std.PreviousRealm().Address()

	if caller != owner {
		panic("access denied")
	}

	owner = newOwner
	std.Emit("OwnershipChange", "newOwner", newOwner.String())
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
	  "pkg_path": "gno.",
	  "func": "ChangeOwner",
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
(`std.Address`) in a variable, and then create helper functions to update the
owners. These helper functions should check if the caller of a function is
whitelisted or not.

Let's deep dive into the different access control mechanisms we can use:

One strategy is to look at the caller with `std.PreviousRealm()`, which could be the
EOA (Externally Owned Account), or the preceding realm in the call stack.

Another approach is to look specifically at the EOA. For this, you should call
`std.OriginCaller()`, which returns the public address of the account that
signed the transaction.

TODO: explain when to use `std.OriginCaller`.

Internally, this call will look at the frame stack, which is basically the stack
of callers, including all the functions, anonymous functions, other realms, and
take the initial caller. This allows you to identify the original caller and
implement access control based on their address.

Here's an example:

```go
import "std"

var admin std.Address = "g1xxxxx"

func AdminOnlyFunction(_ realm) {
	caller := std.PreviousRealm().Address()
	if caller != admin {
		panic("permission denied")
	}
	// ...
}

// func UpdateAdminAddress(_ realm, newAddr std.Address) { /* ... */ }
```

In this example, `AdminOnlyFunction` is a function that can only be called by
the admin. It retrieves the caller's address using `std.PreviousRealm().Address()`,
this can be either another realm contract, or the calling user if there is no
other intermediary realm. and then checks if the caller is the admin. If not, it
panics and stops the execution.

The goal of this approach is to allow a contract to own assets (like grc20 or
coins), so that you can create contracts that can be called by another
contract, reducing the risk of stealing money from the original caller. This is
the behavior of the default grc20 implementation.

Here's an example:

```go
import "std"

func TransferTokens(_ realm, to std.Address, amount int64) {
	caller := std.PreviousRealm().Address()
	if caller != admin {
		panic("permission denied")
	}
	// ...
}
```

In this example, `TransferTokens` is a function that can only be called by the
admin. It retrieves the caller's address using `std.PreviousRealm().Address()`, and
then checks if the caller is the admin. If not, the function panics and execution is stopped.

By using these access control mechanisms, you can ensure that your contract's
functionality is accessible only to the intended users, providing a secure and
reliable way to manage access to your contract.

### Prefer avl.Tree over map for scalable storage

An `avl.Tree` works like a `map` for storing key-value pairs. `maps` store all
entries in one object (accessing any value loads everything), while AVL trees
store each node separately (accessing a value loads only the search path).
This makes `avl.Tree` far more gas-efficient for large or growing datasets.

**Key differences**:

- **AVL Trees**: O(log n) lookup, lazy loading, iterate in **sorted key order**.
- **Maps**: O(1) lookup, type safety, iterate in **unspecified order**.

**Use `avl.Tree` when you need**:

- Lazy loading (gas-efficient for large datasets - only loads the search path)
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
	user := value.(*User) // Type assertion required - values are interface{}
	return false // return true to stop iteration
})

// Range query: get users from "bob" to "charlie" (inclusive)
// This is O(log n + k) where k = results in range
users.Iterate("bob", "charlie", func(name string, value any) bool {
	// Only visits: bob, charlie
	user := value.(*User) 
	return false
})

// Get a specific user (O(log n))
value, exists := users.Get("alice")
if !exists {
	return nil
}
return value.(*User)

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

For a detailed explanation of how AVL trees are stored in Gno's object store, see the [avl package README](../../examples/gno.land/p/nt/avl/README.md).

### Construct "safe" objects

A safe object in Gno is an object that is designed to be tamper-proof and
secure. It's created with the intent of preventing unauthorized access and
modifications. This follows the same principle of making a package an API, but
for a Gno object that can be directly referenced by other realms.

The goal is to create an object which, once instantiated, can be linked and its
pointer can be "stored" by other realms without issue, because it protects its
usage completely.

```go
type MySafeStruct {
	counter nb
	admin std.Address
}

func NewSafeStruct() *MySafeStruct {
	caller := std.PreviousRealm().Address()
	return &MySafeStruct{
		counter: 0,
		admin: caller,
	}
}

func (s *MySafeStruct) Counter() int { return s.counter }
func (s *MySafeStruct) Inc(_ realm) {
	caller := std.PreviousRealm().Address()
	if caller != s.admin {
		panic("permission denied")
	}
	s.counter++
}
```

Then, you can register this object in one or more other realms so that they can access it, while still following your own rules.

```go
import "gno.land/r/otherrealm"

func init() {
	mySafeObj := NewSafeStruct()
	otherrealm.Register(mySafeObject)
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

var fooToken = grc20.NewBanker("Foo Token", "FOO", 4)

func MyBalance(_ realm) uint64 {
	caller := std.PreviousRealm().Address()
	return fooToken.BalanceOf(caller)
}
```

See also: https://gno.land/r/demo/defi/foo20

#### Wrapping Coins

Want the best of both worlds? Consider wrapping your Coins. This gives
your coins the flexibility of GRC20 while keeping the security of Coins.
It's a bit more complex, but it's a powerful option that offers great
versatility.

See also: https://github.com/gnolang/gno/tree/master/examples/gno.land/r/gnoland/wugnot

<!-- TODO:

- suggested file names: types.gno, render.gno, admin.gno, etc.
- packages and realms versioning
- code generation
- unit tests, fuzzing tests, example tests, txtar
- shipping non-contract stuff with the realm: client, documentation, assets
- unoptimized / gas inefficient code
- optimized data structures
- using state machines (gaming example)
- TDD and local dev
- contract-contract pattern
- upgrade pattern
- pausable pattern
- flexible DAO pattern
- maketx run to use go as shell script
- when to launch a local testnet, a full node, gnodev, or using testnets,
  staging, etc
- go std vs gno std
- use rand
- use time
- use oracles
- subscription model
- forking contracts
- finished packages
- packages for developers, realms for users (NPM vs App Store)
- cross-realm ownership: function pointer, callback, inline function ownership
- advanced usages of the frame stack
-->
