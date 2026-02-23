# Standard Library Usage: Go std vs Gno std

## Introduction

Gno uses a standard library that looks similar to Go, but it is not the same.
Some Go packages exist, some do not, and Gno adds blockchain-specific packages.

**This guide explains:**
- how imports work in Gno
- which Go std packages are available
- which ones are not
- how to use Gno blockchain libraries correctly

## Import Path Differences

The most visible difference between Go and Gno standard libraries is their import path structure:

```go
// Go standard library
import "strings"
import "encoding/json"
import "time"

// Gno standard library
import "strings"
import "encoding/binary"

// On-chain packages
import "gno.land/p/demo/avl"
import "gno.land/r/demo/users"
```

**Rule:**
- If the import has no domain → standard library
- If the import has a domain (gno.land/...) → on-chain package

## Go Standard Library in Gno

Gno includes many familiar Go standard libraries, but not all of them. Only deterministic and safe packages are included.

### Available Go std Libraries

- **`strings`** for string manipulation functions
- **`strconv`** for string conversions
- **`encoding/binary`** for binary encoding
- **`encoding/base64`** for base64 encoding
- **`math`** for mathematical functions
- **`testing`** for testing framework (with Gno extensions)
- **`errors`** for error handling
- **`fmt`** for formatted I/O

### Example:
```go
package example

import (
    "strings"
    "encoding/base64"
)

func Process(input string) string {
    upper := strings.ToUpper(input)
    return base64.StdEncoding.EncodeToString([]byte(upper))
}
```

### Unavailable Go std Libraries

Some Go standard libraries are **not available** in Gno due to blockchain constraints:

- **`net/http`** → No HTTP networking
- **`os`** → No file system access
- **`io/ioutil`** → No I/O operations
- **`database/sql`** → No external databases
- **`time.Now()`** → Non-deterministic → must use blockchain time instead

## Gno-Specific Standard Libraries

Gno extends the standard library with blockchain-specific packages under the special `chain` package and its subpackages. The `chain` package provides blockchain-specific functionality not found in Go:

### 1. `chain` - Core blockchain types

```go
import "chain"

// Emit blockchain events
chain.Emit("Transfer", "from", sender, "to", receiver, "amount", "100")

// Work with coins
coin := chain.NewCoin("ugnot", 1000)
coins := chain.NewCoins(coin)

// Get package addresses
addr := chain.PackageAddress("gno.land/r/demo/users")
```

### 2. `chain/runtime` - Execution context

```go
package example

import "chain/runtime" // Special Gno package providing access to the caller

caller := runtime.OriginCaller() // Know who called
height := runtime.ChainHeight()  // Get block height
realm  := runtime.CurrentRealm() // Access realm info
```

### 3. `chain/banker` - Native coin operations

```go
import "chain/banker"

// Get different banker types
readBanker := banker.NewBanker(banker.BankerTypeReadonly)
sendBanker := banker.NewBanker(banker.BankerTypeRealmSend)
issueBanker := banker.NewBanker(banker.BankerTypeRealmIssue)

// Check balances
coins := readBanker.GetCoins(addr)

// Transfer coins
sendBanker.SendCoins(from, to, coins)

// Issue new coins (requires RealmIssue banker)
denom := runtime.CurrentRealm().CoinDenom("mycoin")
issueBanker.IssueCoin(addr, denom, 1000)
```

### 4. `testing` - Testing utilities

Gno extends Go's testing package with blockchain-specific functions:

```go
import "testing"

import "testing"

func TestExample(t *testing.T) {
    caller := address("g1...")
    testing.SetOriginCaller(caller)

    testing.IssueCoins(caller, chain.NewCoins(chain.NewCoin("ugnot", 100)))
}
```

## Practical Usage Patterns

### String Processing

Use Go's `strings` library as normal:

```go
import "strings"

func ValidateUsername(username string) bool {
    username = strings.TrimSpace(username) // Use familiar Go string functions
    username = strings.ToLower(username)
    
    if len(username) < 3 || len(username) > 20 {
        return false
    }
    
    return !strings.Contains(username, " ")
}
```

### Blockchain Time

Don't use `time.Now()`. Use blockchain time instead **use block height as a time reference.**

```go
import (
    "chain/runtime"
    "time"
)

// This is wrong → Non-deterministic
func GetCurrentTime() time.Time {
    return time.Now()
}

// This is correct → Use block height
func GetBlockHeight() int64 {
    return runtime.ChainHeight()
}
```

### Event Emission

Use `chain.Emit()` for off-chain indexing:

```go
import "chain"

chain.Emit("UserCreated",
    "user", caller.String(),
    "height", strconv.Itoa(int(runtime.ChainHeight())),
)
```

**Always use key/value pairs.**

## Discovering Available Libraries

### Browse the Repository

```bash
cd gnovm/stdlibs
find .
```

### Use `gno doc`

```bash
# List package contents
gno doc strings
gno doc encoding/binary
```

## Conclusion

That's it 🎉

Understanding the difference between Go standard libraries and Gno's blockchain libraries is key to writing correct Gno code.
Use Go std packages for pure computation, and `chain/*` packages for all blockchain logic.           

**Remember:** Gno runs in a deterministic environment.
