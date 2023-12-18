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

### Embrace Global Variables in Gno

In Gno, using global variables isn't just acceptable - it's encouraged. This is
because global variables in Gno provide a way to have persisted states
automatically.

In Go, you would typically write your logic and maintain some state in memory.
However, to persist the state and ensure it survives a restart, you would need
to use a store (like a plain file, custom file structure, a database, a
key-value store, an API, etc.).

In contrast, Gno simplifies this process. When you declare global variables in
Gno, the GnoVM automatically persists and restores them as needed between each
run.

However, be mindful not to export your global variables. Doing so would make
them accessible for everyone to read and write.

Here's an ideal pattern to follow:

```go
var counter int

func GetCounter() int {
    return counter
}

func IncCounter() {
    counter++
}
```

### Embrace Panic in Gno

In Gno, it's important to know when to return an `error` and when to use `panic()`.
Each does something different to your code and data.

When you return an error in Gno, it's like giving back any other piece of data.
It tells you something went wrong, but it doesn't stop your code or undo any
changes you made.

But, when you use panic in Gno, it stops your code right away, says it failed,
and doesn't save any changes you made. This is safer when you want to stop
everything and not save wrong changes.

In general, it's good to use `panic()` in realms. In reusable packages, you can
use either panic or errors, depending on what you need.

- TODO: suggest MustXXX and AssertXXX flows in p/.
- TODO: code snippet.

### Understand the importance of `init()`

In Gno, the `init()` function isn't just a function, it's a cornerstone. It's
automatically triggered when a new realm is added onchain, making it a one-time
setup tool for the lifetime of a realm.

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
    admin std.Address
)

func init() {
    created = time.Now()
    admin = std.GetOrigCaller()
}
```

In essence, `init()` in Gno is your go-to function for setting up and
registering realms. It's a powerful tool that helps keep your realms organized
and properly configured from the get-go.

## TODO

- Packages vs realms, subpackages, subrealms, internal
- Elaborate on the benefits of global variables
- Discuss the advantages of NPM-style small and focused libraries
- Describe how versioning is different in Gno
- Explain why exporting a variable is unsafe; instead, suggest creating getters and setters that check for permission to update
- Discuss the possibility of safe objects
- Explain how to export an object securely
- Discuss the future of code generation
- Provide examples of unoptimized / gas inefficient code
- Discuss optimized data structures
- Explain the use of state machines
- Share patterns for setting an initial owner
- Discuss Test-Driven Development (TDD)
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
