---
id: 'effective-gno'
---

# Effective Gno

Welcome to the guide for writing effective Gno code. This document is designed
to help you understand the nuances of Gno and how to use it effectively.

Before we dive in, it's important to note that Gno shares several similarities
with Go. Therefore, if you haven't already, we highly recommend reading
["Effective Go"](https://go.dev/doc/effective_go) as a primer.

## Counter-Intuitive Good Practices

This section highlights some Gno good practices that might seem
counter-intuitive, especially if you're coming from a Go background.

### Embrace Global Variables in Gno Realms

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

### Embrace Panic in Gno

In Gno, it's important to know when to return an `error` and when to use
`panic()`. Each does something different to your code and data.

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
    if caller != "g1234567890123456789012345678912345678" {
        panic("permission denied")
    }
    // ...
}
```

- TODO: suggest MustXXX and AssertXXX flows in p/.

### Understand the Importance of `init()`

In Gno, the `init()` function isn't just a function, it's a cornerstone. It's
automatically triggered when a new realm is added onchain, making it a one-time
setup tool for the lifetime of a realm. In essence, `init()` acts as a
constructor for your realm.

Unlike Go, where `init()` is used for tasks like setting up database
connections, configuring logging, or initializing global variables every time
you start a program, in Gno, `init()` is executed once per realm's lifetime.

In Gno, `init()` primarily serves two purposes:
1. It registers your new realm on a new realm. This is typically done using the
   registry pattern. This means you import another realm and call a method.
2. It configures the initial state, i.e., global variables.

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

## Gno Good Practices

### Design Your Realm as a Public API

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

### Contract-Level Access Control

In Gno, it's a good practice to design your contract as an application with its
own access control. This means that different endpoints of your contract should
be accessible to different types of users, such as the public, admins, or
moderators.

The goal is usually to store the admin address or a list of addresses
(`std.Address`) in a variable, and then create helper functions to update the
owners. These helper functions should check if the caller of a function is
whitelisted or not.

Let's deep dive into the different access control mechanisms we can use:

#### Using the Original Caller Address

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

var admin std.Address = "g1......"

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

#### Using the Previous Realm Address

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

### Construct "Safe" Objects

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

- Packages vs realms, subpackages, subrealms, internal
- Elaborate on the benefits of global variables
- Discuss the advantages of NPM-style small and focused libraries
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
- std.GetOrigCaller vs std.PrevRealm().Addr(), etc
- go std vs gno std
- use rand
- use time
- use oracles
- subscription model
