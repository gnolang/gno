# Pattern: Dynamic Call with Anonymous Functions

The concept of this pattern is to simulate dynamic calls while maintaining full
type safety and hardcoded imports.

To achieve this, we leverage the functionality of `std.PrevRealm()`, which
searches for the previous realm context in the stack. Additionally, a realm can
receive a closure created from another contract as a function parameter.
However, this closure will be executed within the context of the current proxy
realm.

## Pseudo Code

```go
// r/myproxy/proxy.gno

// No imports needed.
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

    // Here, from foo20's perspective, std.PrevRealm is r/mylogic
    myproxy.ExecAsMe(func() {
        // Here, from foo20's perspective, std.PrevRealm is r/myproxy
        foo20.TransferTo(to, amount)
    })
}
```

## Use Cases

1. Contract-based DAO, similar to Gnosis on EVM:
   Users interact with the contract, and once enough approvals are given, the
   DAO performs a call to another contract. This pattern allows for dynamic
   contract-based DAOs, not limited to transferring native tokens with the
   `banker` module.

   ```
   +------------------+     +------------------+     +---------------------------+    +------------------+
   | Users interacting|     | DAO logic waiting|     |    Proxy contract with    |    | A realm checking |
   |  with a DAO      |---->| for all conditions|---->|    `ExecAsMe(func())`     |--->| for std.PrevRealm,|
   |    contract      |     |    to be met     |     |   {assertIsWhitelisted}   |    |   i.e., foo20    |
   +------------------+     +------------------+     +---------------------------+    +------------------+
   ```

2. Users interacting with a versioned Realm V1 and later Realm V2:
   In this case, users interact with a frontend Realm V1 and later transition to
   Realm V2. The contracts focus on logic and are not intended to store assets.
   Instead, they pass through a proxy contract that independently owns the
   assets regardless of the frontend version.

   ```
   +------------------+     +------------------+     +---------------------------+    +------------------+
   |Users interacting |  +->|     Realm V1     |--+  |    Proxy contract with    |    | A realm checking |
   |with a versioned  |--+  +------------------+  +->|    `ExecAsMe(func())`     |--->| for std.PrevRealm,|
   |      realm       |  +->|     Realm V2     |--+  |   {assertIsWhitelisted}   |    |   i.e., foo20    |
   +------------------+     +------------------+     +---------------------------+    +------------------+
   ```
