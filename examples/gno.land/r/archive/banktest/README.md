This is a simple test realm contract that demonstrates how to use the banker.

See [gno.land/r/demo/banktest/banktest.go](/r/demo/banktest/banktest.go) to see the original contract code.

This article will go through each line to explain how it works.

```go
package banktest
```

This package is locally named "banktest" (could be anything).

```go
import (
    "std"
)
```

The "std" package is defined by the gno code in stdlibs/std/. </br> Self explanatory; and you'll see more usage from std later.

```go
type activity struct {
    caller   address
    sent     std.Coins
    returned std.Coins
    time     time.Time
}

func (act *activity) String() string {
    return act.caller.String() + " " +
        act.sent.String() + " sent, " +
        act.returned.String() + " returned, at " +
        act.time.Format("2006-01-02 3:04pm MST")
}

var latest [10]*activity
```

This is just maintaining a list of recent activity to this contract. Notice that the "latest" variable is defined "globally" within the context of the realm with path "gno.land/r/demo/banktest".

This means that calls to functions defined within this package are encapsulated within this "data realm", where the data is mutated based on transactions that can potentially cross many realm and non-realm package boundaries (in the call stack).

```go
// Deposit will take the coins (to the realm's pkgaddr) or return them to user.
func Deposit(returnDenom string, returnAmount int64) string {
    std.AssertOriginCall()
    caller := std.OriginCaller()
    send := std.Coins{{returnDenom, returnAmount}}
```

This is the beginning of the definition of the contract function named "Deposit". `std.AssertOriginCall() asserts that this function was called by a gno transactional Message. The caller is the user who signed off on this transactional message. Send is the amount of deposit sent along with this message.

```go
    // record activity
    act := &activity{
        caller:   caller,
        sent:     std.OriginSend(),
        returned: send,
        time:     time.Now(),
    }
    for i := len(latest) - 2; i >= 0; i-- {
        latest[i+1] = latest[i] // shift by +1.
    }
    latest[0] = act
```

Updating the "latest" array for viewing at gno.land/r/demo/banktest: (w/ trailing colon).

```go
    // return if any.
    if returnAmount > 0 {
```

If the user requested the return of coins...

```go
        banker := std.NewBanker(std.BankerTypeOriginSend)
```

use a std.Banker instance to return any deposited coins to the original sender.

```go
        pkgaddr := std.CurrentRealm().Address()
        // TODO: use std.Coins constructors, this isn't generally safe.
        banker.SendCoins(pkgaddr, caller, send)
        return "returned!"
```

Notice that each realm package has an associated Cosmos address.

Finally, the results are rendered via an ABCI query call when you visit [/r/demo/banktest:](/r/demo/banktest:).

```go
func Render(path string) string {
    // get realm coins.
    banker := std.NewBanker(std.BankerTypeReadonly)
    coins := banker.GetCoins(std.CurrentRealm().Address())

    // render
    res := ""
    res += "## recent activity\n"
    res += "\n"
    for _, act := range latest {
        if act == nil {
            break
        }
        res += " * " + act.String() + "\n"
    }
    res += "\n"
    res += "## total deposits\n"
    res += coins.String()
    return res
}
```

You can call this contract yourself, by vistiing [/r/demo/banktest](/r/demo/banktest) and the [quickstart guide](/r/demo/boards:gnolang/4).
