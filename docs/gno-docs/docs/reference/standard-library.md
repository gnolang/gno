---
id: standard-library
---

# Gno Standard Library

When developing a realm in Gnolang, developers may utilize libraries in [stdlibs](https://github.com/gnolang/gno/tree/master/stdlibs). These are the core standard packages provided for Gnolang [Realms ](../explanation/realms.md)& [Packages](../explanation/packages.md).

Libraries can be imported in a manner similar to how libraries are imported in Golang.

An example of importing a `std` library in Gnolang is demonstrated in the following command:

```go
import "std"
```

Let's explore some of the most commonly used modules in the library.

## `stdshim`

### `banker.gno`

A library for manipulating `Coins`. Interfaces that must be implemented when using this library are as follows:

```go
// returns the list of coins owned by the address
GetCoins(addr Address) (dst Coins)

// sends coins from one address to another
SendCoins(from, to Address, amt Coins)

// returns the total supply of the coin
TotalCoin(denom string) int64

// issues coins to the address
IssueCoin(addr Address, denom string, amount int64)

// burns coins from the address
RemoveCoin(addr Address, denom string, amount int64)
```

### `coins.gno`

A library that declares structs for expressing `Coins`. The struct looks like the following:

```go
type Coin struct {
    Denom    string   `json:"denom"`     // the symbol of the coin
    Amount   int64    `json:"amount"`    // the quantity of the coin
}
```

### `testing`

A library that declares `*testing`, which is a tool used for the creation and execution of test cases during the development and testing phase of realms utilizing the `gno` CLI tool with the `test` option.

There are 3 types of testing in `gno`.

* Type `T`
  * Type passed to Test functions to manage test state and support formatted test logs.
* Type `B`
  * Type passed to Benchmark functions.
    * Manage benchmark timing.
    * Specify the number of iterations to run.
* Type `PB`
  * Used by `RunParallel` for running parallel benchmarks.
