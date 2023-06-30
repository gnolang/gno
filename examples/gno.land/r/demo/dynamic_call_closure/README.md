# Pattern: Dynamic call with anonymous functions

The concept is to simulate dynamic calls while preserving full type safety and
hardcoded imports.

To achieve this, we use the fact that `std.PrevRealm()` is looking for the
previous realm context in the stack, and that a realm can receive as function
parameter, a closure created from another contract, but this closure will be
executed from the context of the current proxy realm.

## Pseudo code

```go
// r/myproxy/proxy.gno

// nothing to import.
var approvedCallers = []std.Address{myOtherContract, myAddress, ...}
func UpdateApprovedCallers() {...}

func ExecAsMe(callback func()) {
    assertIsApprovedCaller()
    callback()
}
```

```go
// r/mylogic/logic.gno
import "gno.land/r/myproxy"
import "gno.land/r/foo20"

func SendFoo20ThroughProxy(to std.Address, amount int) {
    assertConditionsAreMet()

    // HERE, from foo20's PoV, std.PrevRealm is r/mylogic
    myproxy.ExecAsMe(func() {
        // HERE, from foo20's PoV, std.PrevRealm is r/myproxy
        foo20.TransferTo(to, amount)
    })
}

```

## Use cases

A contract-based DAO, similar to Gnosis on EVM.

Users are interacting with the contract, and when enough approvals are given,
the DAO will perform a call to another contract.

The idea is to make those contract-based DAO dynamic and not limited to
transfering native tokens with the `banker` module.

    +------------------+     +------------------+     +---------------------------+    +------------------+
    |Users interacting |     |DAO logic waiting |     |    Proxy contract with    |    | A realm checking |
    |    with a DAO    |---->|for all conditions|---->|    `ExecAsMe(func())`     |--->|for std.PrevRealm,|
    |     contract     |     |    to be met     |     |   {assertIsWhitelisted}   |    |   i.e., foo20    |
    +------------------+     +------------------+     +---------------------------+    +------------------+

Users are interacting with a frontend Realm V1, and later to Realm V2, those
contracts are focused on logic, while they are not intended to store assets, for
this need, they pass through a proxy contract that will own the assets
independently from the frontend version.

    +------------------+     +------------------+     +---------------------------+    +------------------+
    |Users interacting |  +->|     Realm V1     |--+  |    Proxy contract with    |    | A realm checking |
    |with a versionned |--+  +------------------+  +->|    `ExecAsMe(func())`     |--->|for std.PrevRealm,|
    |      realm       |  +->|     Realm V2     |--+  |   {assertIsWhitelisted}   |    |   i.e., foo20    |
    +------------------+     +------------------+     +---------------------------+    +------------------+
