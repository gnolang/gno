---
id: 'effective-gno'
---

# Effective Gno

Welcome to the guide for writing effective Gno code. This document is designed
to help you understand the nuances of Gno and how to use it effectively.

Before we dive in, it's important to note that Gno shares several similarities
with Go. Therefore, if you haven't already, we highly recommend reading
["Effective Go"](https://go.dev/doc/effective_go) as a primer.

## Counter-intuitive good practices

This section highlights some Gno good practices that might seem
counter-intuitive, especially if you're coming from a Go background.

### Embrace global variables in realms

In Gno, using global variables is not only acceptable, but it's also encouraged,
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
func IncCounter() {
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

Panic in Gno is not just for critical errors or programming mistakes as it is in
Go. Instead, it's used as a control flow mechanism to stop the execution of a
contract when something goes wrong. This could be due to an invalid input, a
failed precondition, or any other situation where it's not possible or desirable
to continue executing the contract.

So, while in Go, you should avoid `panic` and handle `error`s gracefully, in Gno,
don't be afraid to use `panic` to enforce contract rules and protect the integrity
of your contract's state. Remember, a well-placed panic can save your contract
from a lot of trouble.

When you return an `error` in Gno, it's like giving back any other piece of data.
It tells you something went wrong, but it doesn't stop your code or undo any
changes you made.

But, when you use `panic` in Gno, it stops your code right away, says it failed,
and doesn't save any changes you made. This is safer when you want to stop
everything and not save wrong changes.

In general, it's good to use `panic()` in realms. In reusable packages, you can
use either `panic` or `error`, depending on what you need.

```go
import "std"

func Foobar() {
    caller := std.GetOrigCaller()
    if caller != "g1xxxxx" {
        panic("permission denied")
    }
    // ...
}
```

- TODO: suggest MustXXX and AssertXXX flows in p/.

### Understand the importance of `init()`

In Gno, the `init()` function isn't just a function, it's a cornerstone. It's
automatically triggered when a new realm is added onchain, making it a one-time
setup tool for the lifetime of a realm. In essence, `init()` acts as a
constructor for your realm.

Unlike Go, where `init()` is used for tasks like setting up database
connections, configuring logging, or initializing global variables every time
you start a program, in Gno, `init()` is executed once per realm's lifetime.

In Gno, `init()` primarily serves two purposes:
1. It establishes the initial state, specifically, setting up global variables.
2. It communicates with another realm, for example, to register itself in a registry.

```go
import "gno.land/r/some/registry"

func init() {
    registry.Register("myID", myCallback)
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
    list    []string
)

func init() {
    created = time.Now()
    admin = std.GetOrigCaller()
    list = append(list, "foo", "bar")
}
```

In essence, `init()` in Gno is your go-to function for setting up and
registering realms. It's a powerful tool that helps keep your realms organized
and properly configured from the get-go. Acting as a constructor, it sets the
stage for the rest of your realm's lifecycle.

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

A Gno contract is not just its lines of code, but also the imports it uses. And
importantly, Gno contracts are not just for developers. For the first time, it
makes sense for users to check out what they are executing too. Code simplicity,
explicitness, and trustability are paramount.

Another good reason for creating simple, focused libraries is the composability
of Go and Gno. Essentially, you can think of each `p/` package as a Lego brick
in an ever-growing collection, giving more power to users. `p/` in Gno is
basically a way to extend the standard libraries in a community-driven manner.

Unlike other compiled languages where dependencies are not always well-known and
clear metrics are lacking, Gno allows for a reputation system not only for the
called contracts, but also for the dependencies.

For example, you might choose to use well-crafted `p/` packages that have been
reviewed, audited, and have billions of transactions under their belt, boasting
super high stability. This approach can make your code smaller and more
reliable.

In other platforms, an audit usually involves auditing everything, including the
dependencies. However, in Gno, we can expect that over time, contracts will
become smaller, more powerful, and partially audited by default, thanks to this
enforced open-source system.

So, while you can still adhere to the original philosophy of minimizing
dependencies, ultimately, try to use and write super stable, simple, tested,
focused `p/` small libraries. This approach can lead to more reliable,
efficient, and trustworthy Gno contracts.

##  When Gno takes Go practices to the next level

### Documentation is for users

One of the well-known proverbs in Go is: ["Documentation is for
users"](https://www.youtube.com/watch?v=PAAkCSZUG1c&t=1147s), as stated by Rob
Pike. In Go, documentation is for users, but users are often developers. In Gno,
documentation is for users, but users can be another developer or the end users.

In Go, we usually have well-written documentation for other developers to
maintain and use our code as a library. Then, we often have another layer of
documentation on our API, sometimes with OpenAPI Specs, Protobuf, or even user
documentation.

In Gno, the focus shifts towards writing documentation for the end user. You can
even consider that the main reader is an end user, who is not so interested in
technical details, but mostly interested in how and why they should use a
particular endpoint. Comments will be used for code source reading, but also to
generate documentation and even for smart wallets that need to understand what
to do.

Inline comments have the same goal: to guide users (developers or end users)
through the code. While comments are still important for maintainability, their
main purpose in Gno is for discoverability. This shift towards user-centric
documentation reflects the broader shift in Gno towards making code more
accessible and understandable for all users, not just developers.

TODO: `func ExampleXXX`.

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
or not, or perhaps in a privileged mode for very rare libraries. But for now,
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

Packages and realms can be organized into subfolders. However, consider that the
best place for your main project will likely be `r/NAMESPACE/PROJECT`, similar
to how repositories are organized on GitHub.

If you have multiple sublevels of realms, remember that they are actually
independent realms and won't share data. A good usage could be to have an
ecosystem of realms, where one realm is about storing the state, another one
about configuration, etc. But in general, a single realm makes sense.

You can also create small realms to create your ecosystem. For example, you
could centralize all the authentication for all your company/organization in
`r/NAMESPACE/auth`, and then import it from all your contracts.

The `p/` prefix is different. In general, you should use top-level `p/` like
`p/NAMESPACE/PROJECT` only for things you expect people to use. If your goal is
just to have internal libraries that you created to centralize your helpers and
don't expect that other people will use your helpers, then you should probably
use subfolders like `p/NAMESPACE/PROJECT/foo/bar/baz`.

TODO: link to the versionning section

### Design your realm as a public API

In Go, all your packages, including your dependencies, are typically treated as
part of your safe zone, similar to a secure perimeter. The boundary is drawn
between your program and the rest of the world, which means you secure the API
itself, potentially with authentication middlewares.

However, in Gno, your package is the public API. It's exposed to the outside
world and can be accessed by other realms. Therefore, it's crucial to design
your realm with the same level of care and security considerations as you would
a public API.

One approach is to simulate a secure perimeter within your realm by having
private functions for the logic and then writing your API layer by adding some
front-facing API with authentication. This way, you can control access to your
realm's functionality and ensure that only authorized callers can execute
certain operations.

```go
import "std"

func PublicMethod(nb int) {
    caller := std.GetOrigCaller()
    privateMethod(caller, nb)
}

func privateMethod(caller std.Address, nb int) { /* ... */ }
```

In this example, `PublicMethod` is a public function that can be called by other
realms. It retrieves the caller's address using `std.GetOrigCaller()`, and then
passes it to `privateMethod`, which is a private function that performs the
actual logic. This way, `privateMethod` can only be called from within the
realm, and it can use the caller's address for authentication or authorization
checks.

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

One approach is to look at the EOA (Externally Owned Account), which is the
original caller. For this, you should call `std.GetOrigCaller()`, which returns
the address of the wallet used to make the transaction.

Internally, this call will look at the frame stack, which is basically the stack
of callers including all the functions, anonymous functions, other realms, and
take the initial caller. This allows you to identify the original caller and
implement access control based on their address.

Here's an example:

```go
import "std"

var admin std.Address = "g1xxxxx"

func AdminOnlyFunction() {
    caller := std.GetOrigCaller()
    if caller != admin {
        panic("permission denied")
    }
    // ...
}

// func UpdateAdminAddress(newAddr std.Address) { /* ... */ }
```

In this example, `AdminOnlyFunction` is a function that can only be called by
the admin. It retrieves the caller's address using `std.GetOrigCaller()`, and
then checks if the caller is the admin. If not, it panics and stops the
execution.

Another approach is to use `std.PrevRealm().Addr()`, which returns the previous
realm. This can be either another realm contract, or the calling user if there
is no other intermediary realm.

The goal of this approach is to allow a contract to own assets (like grc20 or
native tokens), so that you can create contracts that can be called by another
contract, reducing the risk of stealing money from the original caller. This is
the behavior of the default grc20 implementation.

Here's an example:

```go
import "std"

func TransferTokens(to std.Address, amount int64) {
    caller := std.PrevRealm().Addr()
    if caller != admin {
        panic("permission denied")
    }
    // ...
}
```

In this example, `TransferTokens` is a function that can only be called by the
admin. It retrieves the caller's address using `std.PrevRealm().Addr()`, and
then checks if the caller is the admin. If not, it panics and stops the
execution.

By using these access control mechanisms, you can ensure that your contract's
functionality is accessible only to the intended users, providing a secure and
reliable way to manage access to your contract.

### Using avl.Tree for efficient data retrieval

In Gno, the `avl.Tree` data structure is a powerful tool for optimizing data
retrieval. It works by lazily resolving information, which means it only loads
the data you need when you need it. This allows you to scale your application
and pay less gas for data retrieval.

The `avl.Tree` can be used like a map, where you can store key-value pairs and
retrieve an entry with a simple key. However, unlike a traditional map, the
`avl.Tree` doesn't load unnecessary data. This makes it particularly efficient
for large data sets where you only need to access a small subset of the data at
a time.

Here's an example of how you can use `avl.Tree`:

```go
import "avl"

var tree avl.Tree

func GetPost(id string) *Post {
    return tree.Get(id).(*Post)
}

func AddPost(id string, post *Post) {
    tree.Set(id, post)
}
```

In this example, `GetPost` is a function that retrieves a post from the
`avl.Tree` using an ID. It only loads the post with the specified ID, without
loading any other posts.

In the future, we plan to add internal "map" support that will be as efficient
as an `avl.Tree` but with a more idiomatic API. For now, consider storing things
in slices when you know the structure will stay small, and consider using
`avl.Tree` each time you can make direct access.

You can also create SQL-like indexes by having multiple `avl.Tree` instances for
different fields. For example, you can have an `avl.Tree` for ID to *post, then
an `avl.Tree` for Tags, etc. Then, you can create reader methods that will just
retrieve what you need, similar to SQL indexes.

By using `avl.Tree` and other efficient data structures, you can optimize your
Gno code for performance and cost-effectiveness, making your applications more
scalable and efficient.

TODO: multi-indices example

### Construct "safe" objects

A safe object in Gno is an object that is designed to be tamper-proof and
secure. It's created with the intent of preventing unauthorized access and
modifications. This follows the same principle of making a package an API, but
for a Go object that can be directly referenced by other realms.

The goal is to create an object which, once instantiated, can be linked and its
pointer can be "stored" by other realms without issue, because it protects its
usage completely.

```go
type MySafeStruct {
    counter nb
    admin std.Address
}

func NewSafeStruct() *MySafeStruct {
    caller := std.GetOrigCaller()
    return &MySafeStruct{
        counter: 0,
        admin: caller,
    }
}

func (s *MySafeStruct) Counter() int { return s.counter }
func (s *MySafeStruct) Inc() {
    caller := std.GetOrigCaller()
    if caller != s.admin {
        panic("permission denied")
    }
    s.counter++
}
```

Then, you can register this object in another or several other realms so other
realms can access the object, but still following your own rules.

```go 
import "gno.land/r/otherrealm"

func init() {
    mySafeObj := NewSafeStruct()
    otherrealm.Register(mySafeObject)
}

// then, otherrealm can call the public functions but won't be the "owner" of
// the object.
```

## TODO

- Describe how versioning is different in Gno
- Explain why exporting a variable is unsafe; instead, suggest creating getters and setters that check for permission to update
- Explain how to export an object securely
- Discuss the future of code generation
- Provide examples of unoptimized / gas inefficient code
- Discuss optimized data structures
- Explain the use of state machines
- Share patterns for setting an initial owner
- Discuss Test-Driven Development (TDD)
- write tests efficiently (mixing unit tests, and other tests)
- Suggest shipping related code to aid review
- Encourage writing documentation for users, not just developers
- Discuss the different reasons for exporting/unexporting things
- Introduce the contract contract pattern
- Discuss the upgrade pattern and its future
- Introduce the DAO pattern
- Discuss the use of gno run for customization instead of contracts everywhere
- Suggest using multiple AVL trees as an alternative to SQL indexes
- Discuss the use of r/NAME/home and p/NAME/home/{foo, bar}[/v0-9]
- Provide guidance on when to launch a local testnet, a full node, gnodev, etc.
- Suggest matching the package name with the folder
- Recommend using the demo/ folder for most things
- Suggest keeping package names short and clear
- Discuss VERSIONING
- Suggest using p/ for interfaces and r/ for implementation
- make your contract mixing onchain, unittest, and run, and eventually client.
- go std vs gno std
- use rand
- use time
- use oracles
- subscription model
- forking contracts
- finish contracts
- pausable contracts
- more go than go: everything in code; use go comments; exception: readme.md
