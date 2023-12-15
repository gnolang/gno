---
id: 'effective-gno'
---

# Effective Gno

This document provides advices and guidelines for writing effective Gno code.

First, Gno shares several similarities with Go. Therefore, please read ["Effective Go"](https://go.dev/doc/effective_go) first.

## Counter-Intuitive Good Practices

In this section, we're listing Gno good practices that could sound as bad practices if you come from Go.

### Global Variable is Good in Gno.

It's not just a good practice, in Gno, this is the way to have persisted states. Automatically.

In Go, you would write your logic in Go, and you could have some state in memory, but as soon as you would want to persist the state to survive a restart, then you would need to use a store (plan file, custom file structure, a database, a key-value store, an API, ...).

In Gno, you declare global variables; then, GnoVM will automatically persist and restore them when needed between each runs.

Take care of not exporting your global variables, because then they will be available for everyone not only to read, but also to write.

An ideal pattern could this one:

```go
var counter int

func GetCounter() int {
    return counter
}

func IncCounter() {
    counter++
}
```


## TODO

- panic is good
- init() is good
- global variable is good
- NPM-style small and focused libraries are good
- versionning is different
- exporting a variable is unsafe; instead, you should create getters and setters checking for permission to udpate
- safe objects are possible
- exporting an object can be done, but the object needs to be made "securely smart"
- code gen -> is not yet a thing but will be
- unoptimized / gas ineffecient examples
- optimized data structures
- state machines
- patterns to set an initial owner
- TDD
- ship related code to help the review
- write doc for users, not only for developers
- export/unexport things for different reasons
- contract contract pattern: https://github.com/gnolang/gno/pull/1262/files#diff-115a8376223d0de272e687826e128df17aca57257b143b76b476e9fb39eb4b23R18
- upgrade pattern: speak about the future + link to contract-contract pattern
- DAO pattern: contract-contract pattern
- gno run for customization instead of contracts everywhere
- using multiple avl trees as an alternative to sql indexes
- r/NAME/home
- p/NAME/home/{foo, bar}[/v0-9]
- when to launch a local testnet, a full node, gnodev, etc, personal portal loop, etc
- package name should match the folder
- use demo/ folder for most things
- package names should be short and clear
- VERSIONING: TODO
- using p/ for interfaces, r/ for implementation
