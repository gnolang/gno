# Standard Libraries

Gno comes with a set of standard libraries which are included to ease development
and provide extended functionality to the language. These include:
- standard libraries as we know them in classic Go, i.e. `strings`, `testing`, etc.
- a special `std` package, which contains types, interfaces, and APIs created to
handle blockchain-related functionality, such as fetching the last caller,
fetching coins sent along with a transaction, getting the block timestamp and height, and more.

Standard libraries differ from on-chain packages in terms of their import path structure.
Unlike on-chain [packages](./gno-packages.md), standard libraries do not incorporate
a domain-like format at the beginning of their import path. For example:
- `import "strings"` refers to a standard library
- `import "gno.land/p/demo/avl"` refers to an on-chain package.

To see concrete implementation details & API references of the `std` pacakge,
see the reference section.

## Accessing documentation

Apart from the official documentation you are currently reading, you can also
access documentation for the standard libraries in several other different ways.
You can obtain a list of all the available standard libraries with the
following commands:

```console
$ cd gnovm/stdlibs # go to correct directory

$ find -type d
./testing
./math
./crypto
./crypto/chacha20
./crypto/chacha20/chacha
./crypto/chacha20/rand
./crypto/sha256
./crypto/cipher
...
```

All the packages have automatically generated documentation through the use of the
`gno doc` command, which has similar functionality and features to `go doc`:

```console
$ gno doc encoding/binary
package binary // import "encoding/binary"

Package binary implements simple translation between numbers and byte sequences
and encoding and decoding of varints.

[...]

var BigEndian bigEndian
var LittleEndian littleEndian
type AppendByteOrder interface{ ... }
type ByteOrder interface{ ... }
$ gno doc -u -src encoding/binary littleEndian.AppendUint16
package binary // import "encoding/binary"

func (littleEndian) AppendUint16(b []byte, v uint16) []byte {
        return append(b,
                byte(v),
                byte(v>>8),
        )
}
```

`gno doc` will work automatically when used within the Gno repository or any
repository which has a `go.mod` dependency on `github.com/gnolang/gno`.

Another alternative is setting your environment variable `GNOROOT` to point to
where you cloned the Gno repository.

```sh
export GNOROOT=$HOME/gno
```

## Coin

A Coin is a native Gno type that has a denomination and an amount. Coins can be
issued by the native Gno Banker.

A coin is defined by the following:

```go
type Coin struct {
	Denom  string `json:"denom"`
	Amount int64  `json:"amount"`
}
```

`Denom` is the denomination of the coin, i.e. `ugnot`, and `Amount` is a
non-negative amount of the coin.

Multiple coins can be bundled together into a `Coins` slice:

```go
type Coins []Coin
```

This slice behaves like a mathematical set - it cannot contain duplicate `Coin` instances.

The `Coins` slice can be included in a transaction made by a user addresses or a realm.
Coins in this set are then available for access by specific types of Bankers,
which can manipulate them depending on access rights.

Read more about coins in the [Effective Gno](./effective-gno.md) section.

The Coin(s) API can be found in the `std` package.

## Banker

The Banker's main purpose is to handle balance changes of [native coins](coin.md)
within Gno chains. This includes issuance, transfers, and burning of coins.

The Banker module can be cast into 4 subtypes of bankers that expose different
functionalities and safety features within your packages and realms.

### Banker Types

1. `BankerTypeReadonly` - read-only access to coin balances
2. `BankerTypeOriginSend` - full access to coins sent with the transaction that called the banker
3. `BankerTypeRealmSend` - full access to coins that the realm itself owns, including the ones sent with the transaction
4. `BankerTypeRealmIssue` - able to issue new coins

## Events

Events in Gno are a fundamental aspect of interacting with and monitoring
on-chain applications. They serve as a bridge between the on-chain environment
and off-chain services, making it simpler for developers, analytics tools, and
monitoring services to track and respond to activities happening in gno.land.

Gno events are pieces of data that log specific activities or changes occurring
within the state of an on-chain app. These activities are user-defined; they might
be token transfers, changes in ownership, updates in user profiles, and more.
Each event is recorded in the ABCI results of each block, ensuring that action
that happened is verifiable and accessible to off-chain services.

To emit an event, you can use the `Emit()` function from the `std` package
provided in the Gno standard library. The `Emit()` function takes in a string
representing the type of event, and an even number of arguments after representing
`key:value` pairs.

Read more about events & `Emit()` in
[Effective Gno](./effective-gno.md#emit-gno-events-to-make-life-off-chain-easier).

An event contained in an ABCI response of a block will include the following
data:

``` json
{
    "@type": "/tm.gnoEvent", // TM2 type
    "type": "OwnershipChange", // Type/name of event defined in Gno
    "pkg_path": "gno.land/r/demo/example", // Path of the emitter
    "func": "ChangeOwner", // Gno function that emitted the event
    "attrs": [ // Slice of key:value pairs emitted
        {
            "key": "oldOwner",
            "value": "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
        },
        {
            "key": "newOwner",
            "value": "g1zzqd6phlfx0a809vhmykg5c6m44ap9756s7cjj"
        }
    ]
}
```

You can fetch the ABCI response of a specific block by using the `/block_results`
RPC endpoint.

<!-- XXX: remove everything after this and use automatically generated package doc -->

## Package `std`

[//]: <> (todo: autogenerate from godoc.)

This is the reference page for the special `std` package found in Gno, containing
critical functionality for managing realms, addresses, the Banker module, etc.

## Address
Native address type in Gno, conforming to the Bech32 format.

```go
type Address string
func (a Address) IsValid() bool {...}
func (a Address) String()  string {...}
```

### IsValid
Check if **Address** is of a valid length, and conforms to the bech32 format.

##### Usage
```go
if !address.IsValid() {...}
```

---

### String
Get **string** representation of **Address**.

##### Usage
```go
stringAddr := addr.String()
```

---

## Banker

```go
type BankerType uint8

const (
    BankerTypeReadonly BankerType = iota
    BankerTypeOriginSend
    BankerTypeRealmSend
    BankerTypeRealmIssue
)

type Banker interface {
    GetCoins(addr Address) (dst Coins)
    SendCoins(from, to Address, coins Coins)
    IssueCoin(addr Address, denom string, amount int64)
    RemoveCoin(addr Address, denom string, amount int64)
}
```

### NewBanker
Returns `Banker` of the specified type.

##### Parameters
- `BankerType` - type of Banker to get:
    - `BankerTypeReadonly` - read-only access to coin balances
    - `BankerTypeOrigSend` - full access to coins sent with the transaction that calls the banker
    - `BankerTypeRealmSend` - full access to coins that the realm itself owns, including the ones sent with the transaction
    - `BankerTypeRealmIssue` - able to issue new coins

##### Usage

```go
banker := std.NewBanker(std.<BankerType>)
```
---

### GetCoins
Returns `Coins` owned by `Address`.

##### Parameters
- `addr` **Address** to fetch balances for

##### Usage

```go
coins := banker.GetCoins(addr)
```
---

### SendCoins
Sends `coins` from address `from` to address `to`. `coins` needs to be a well-defined
`Coins` slice.

##### Parameters
- `from` **Address** to send from
- `to` **Address** to send to
- `coins` **Coins** to send

##### Usage
```go
banker.SendCoins(from, to, coins)
```
---

### IssueCoin
Issues `amount` of coin with a denomination `denom` to address `addr`.

##### Parameters
- `addr` **Address** to issue coins to
- `denom` **string** denomination of coin to issue
- `amount` **int64** amount of coin to issue

##### Usage
```go
banker.IssueCoin(addr, denom, amount)
```

:::info Coin denominations

`Banker` methods expect qualified denomination of the coins. Read more [here](#coindenom).

:::

---


### RemoveCoin
Removes (burns) `amount` of coin with a denomination `denom` from address `addr`.

##### Parameters
- `addr` **Address** to remove coins from
- `denom` **string** denomination of coin to remove
- `amount` **int64** amount of coin to remove

##### Usage
```go
banker.RemoveCoin(addr, denom, amount)
```

---

## Chain-related

### AssertOriginCall
```go
func AssertOriginCall()
```
Panics if caller of function is not an EOA.

##### Usage
```go
std.AssertOriginCall()
```
---

### ChainDomain
```go
func ChainDomain() string
```
Returns the chain domain. Currently only `gno.land` is supported.

##### Usage
```go
domain := std.ChainDomain() // gno.land
```
---

### Emit
```go
func Emit(typ string, attrs ...string)
```
Emits a Gno event. Takes in a **string** type (event identifier), and an even number of string
arguments acting as key-value pairs to be included in the emitted event.

##### Usage
```go
std.Emit("MyEvent", "myKey1", "myValue1", "myKey2", "myValue2")
```
---

### ChainID
```go
func ChainID() string
```
Returns the chain ID.

##### Usage
```go
chainID := std.ChainID() // dev | test5 | main ...
```
---

### ChainHeight
```go
func ChainHeight() int64
```
Returns the current block number (height).

##### Usage
```go
height := std.ChainHeight()
```
---

### OriginSend
```go
func OriginSend() Coins
```
Returns the `Coins` that were sent along with the calling transaction.

##### Usage
```go
coinsSent := std.OriginSend()
```
---

### OriginCaller
```go
func OriginCaller() Address
```
Returns the original signer of the transaction.

##### Usage
```go
caller := std.OriginCaller()
```
---

### OriginPkgAddress
```go
func OriginPkgAddress() Address
```
Returns the address of the first (entry point) realm/package in a sequence of realm/package calls.

##### Usage
```go
addr := std.OriginPkgAddress()
```
---

### CurrentRealm
```go
func CurrentRealm() Realm
```
Returns current [Realm](./realms.md) object.

##### Usage
```go
currentRealm := std.CurrentRealm()
```
---

### PrevRealm
```go
func PreviousRealm() Realm
```
Returns the previous caller [realm](./realms.md) (can be code or user realm). If caller is a
user realm, `pkgpath` will be empty.

##### Usage
```go
prevRealm := std.PreviousRealm()
```
---

### CallerAt
```go
func CallerAt(n int) Address
```
Returns the n-th caller of the function, going back in the call trace.
Includes calls to pure packages.

##### Usage
```go
currentRealm := std.CallerAt(1)      // returns address of current realm
previousRealm := std.CallerAt(2)     // returns address of previous realm/caller
std.CallerAt(0)                      // error, n must be > 0
```
---

### DerivePkgAddr
```go
func DerivePkgAddr(pkgPath string) Address
```
Derives the Realm address from its `pkgpath` parameter.

##### Usage
```go
realmAddr := std.DerivePkgAddr("gno.land/r/demo/tamagotchi") //  g1a3tu874agjlkrpzt9x90xv3uzncapcn959yte4
```

---

### CoinDenom
```go
func CoinDenom(pkgPath, coinName string) string
```
Composes a qualified denomination string from the realm's `pkgPath` and the
provided coin name, e.g. `/gno.land/r/demo/blog:blgcoin`. This method should be
used to get fully qualified denominations of coins when interacting with the
`Banker` module. It can also be used as a method of the `Realm` object.
Read more[here](#coindenom-1).

#### Parameters
- `pkgPath` **string** - package path of the realm
- `coinName` **string** - The coin name used to build the qualified denomination.  Must start with a lowercase letter, followed by 2–15 lowercase letters or digits.

#### Usage
```go
denom := std.CoinDenom("gno.land/r/demo/blog", "blgcoin") // /gno.land/r/demo/blog:blgcoin
```

---

## Coin

```go
type Coin struct {
	Denom  string `json:"denom"`
	Amount int64  `json:"amount"`
}

func NewCoin(denom string, amount int64) Coin {...}
func (c Coin) String() string {...}
func (c Coin) IsGTE(other Coin) bool {...}
func (c Coin) IsLT(other Coin) bool {...}
func (c Coin) IsEqual(other Coin) bool {...}
func (c Coin) Add(other Coin) Coin {...}
func (c Coin) Sub(other Coin) Coin {...}
func (c Coin) IsPositive() bool {...}
func (c Coin) IsNegative() bool {...}
func (c Coin) IsZero() bool {...}
```

### NewCoin
Returns a new Coin with a specific denomination and amount.

##### Usage
```go
coin := std.NewCoin("ugnot", 100)
```
---

### String
Returns a string representation of the `Coin` it was called upon.

##### Usage
```go
coin := std.NewCoin("ugnot", 100)
coin.String() // 100ugnot
```
---

### IsGTE
Checks if the amount of `other` Coin is greater than or equal than amount of
Coin `c` it was called upon. If coins compared are not of the same denomination,
`IsGTE` will panic.

##### Parameters
- `other` **Coin** to compare with

##### Usage
```go
coin1 := std.NewCoin("ugnot", 150)
coin2 := std.NewCoin("ugnot", 100)

coin1.IsGTE(coin2) // true
coin2.IsGTE(coin1) // false
```
---

### IsLT
Checks if the amount of `other` Coin is less than the amount of Coin `c` it was
called upon. If coins compared are not of the same denomination, `IsLT` will
panic.

##### Parameters
- `other` **Coin** to compare with

##### Usage
```go
coin := std.NewCoin("ugnot", 150)
coin := std.NewCoin("ugnot", 100)

coin1.IsLT(coin2) // false
coin2.IsLT(coin1) // true
```
---

### IsEqual
Checks if the amount of `other` Coin is equal to the amount of Coin `c` it was
called upon. If coins compared are not of the same denomination, `IsEqual` will
panic.

##### Parameters
- `other` **Coin** to compare with

##### Usage
```go
coin1 := std.NewCoin("ugnot", 150)
coin2 := std.NewCoin("ugnot", 100)
coin3 := std.NewCoin("ugnot", 100)

coin1.IsEqual(coin2) // false
coin2.IsEqual(coin1) // false
coin2.IsEqual(coin3) // true
```
---

### Add
Adds two coins of the same denomination. If coins are not of the same
denomination, `Add` will panic. If final amount is larger than the maximum size
of `int64`, `Add` will panic with an overflow error. Adding a negative amount
will result in subtraction.

##### Parameters
- `other` **Coin** to add

##### Usage
```go
coin1 := std.NewCoin("ugnot", 150)
coin2 := std.NewCoin("ugnot", 100)

coin3 := coin1.Add(coin2)
coin3.String() // 250ugnot
```
---

### Sub
Subtracts two coins of the same denomination. If coins are not of the same
denomination, `Sub` will panic. If final amount is smaller than the minimum size
of `int64`, `Sub` will panic with an underflow error. Subtracting a negative amount
will result in addition.

##### Parameters
- `other` **Coin** to subtract

##### Usage
```go
coin1 := std.NewCoin("ugnot", 150)
coin2 := std.NewCoin("ugnot", 100)

coin3 := coin1.Sub(coin2)
coin3.String() // 50ugnot
```
---

### IsPositive
Checks if a coin amount is positive.

##### Usage
```go
coin1 := std.NewCoin("ugnot", 150)
coin2 := std.NewCoin("ugnot", -150)

coin1.IsPositive() // true
coin2.IsPositive() // false
```
---

### IsNegative
Checks if a coin amount is negative.

##### Usage
```go
coin1 := std.NewCoin("ugnot", 150)
coin2 := std.NewCoin("ugnot", -150)

coin1.IsNegative() // false
coin2.IsNegative() // true
```
---

### IsZero
Checks if a coin amount is zero.

##### Usage
```go
coin1 := std.NewCoin("ugnot", 150)
coin2 := std.NewCoin("ugnot", 0)

coin1.IsZero() // false
coin2.IsZero() // true
```

---

## Coins

`Coins` is a set of `Coin`, one per denomination.

```go
type Coins []Coin

func NewCoins(coins ...Coin) Coins {...}
func (c Coins) String() string {...}
func (c Coins) AmountOf(denom string) int64 {...}
func (c Coins) Add(other Coins) Coins {...}
```

### NewCoins
Returns a new set of `Coins` given one or more `Coin`. Consolidates any denom
duplicates into one, keeping the properties of a mathematical set.

##### Usage
```go
coin1 := std.NewCoin("ugnot", 150)
coin2 := std.NewCoin("example", 100)
coin3 := std.NewCoin("ugnot", 100)

coins := std.NewCoins(coin1, coin2, coin3)
coins.String() // 250ugnot, 100example
```
---

#### String
Returns a string representation of the `Coins` set it was called upon.

##### Usage
```go
coins := std.Coins{std.Coin{"ugnot", 100}, std.Coin{"foo", 150}, std.Coin{"bar", 200}}
coins.String() // 100ugnot,150foo,200bar
```
---

### AmountOf
Returns **int64** amount of specified coin within the `Coins` set it was called upon. Returns `0` if the specified coin does not exist in the set.

#### Parameters
- `denom` **string** denomination of specified coin

#### Usage
```go
coins := std.Coins{std.Coin{"ugnot", 100}, std.Coin{"foo", 150}, std.Coin{"bar", 200}}
coins.AmountOf("foo") // 150
```
---

### Add
Adds (or updates) the amount of specified coins in the `Coins` set. If the specified coin does not exist, `Add` adds it to the set.

#### Parameters
- `other` **Coins** to add to `Coins` set

#### Usage
```go
coins := // ...
otherCoins := // ...
coins.Add(otherCoins)
```

## Realm

`Realm` is the structure representing a realm in Gno. See our [realm documentation](./realms.md) for more details.

```go
type Realm struct {
    addr    Address
    pkgPath string
}

func (r Realm) Address() Address {...}
func (r Realm) PkgPath() string {...}
func (r Realm) IsUser() bool {...}
func (r Realm) CoinDenom(coinName string) string {...}
```

### Addr
Returns the **Address** field of the realm it was called upon.

##### Usage
```go
realmAddr := r.Address() // eg. g1n2j0gdyv45aem9p0qsfk5d2gqjupv5z536na3d
```
---
### PkgPath
Returns the **string** package path of the realm it was called upon.

##### Usage
```go
realmPath := r.PkgPath() // eg. gno.land/r/gnoland/blog
```
---
### IsUser
Checks if the realm it was called upon is a user realm.

##### Usage
```go
if r.IsUser() {...}
```

---

### CoinDenom

Composes a qualified denomination string from the realm's `pkgPath` and the
provided coin name, e.g. `/gno.land/r/demo/blog:blgcoin`. This method should be
used to get fully qualified denominations of coins when interacting with the
`Banker` module.

#### Parameters
- `coinName` **string** - The coin name used to build the qualified denomination.
Must start with a lowercase letter, followed by 2–15 lowercase letters or digits.

#### Usage
```go
// in "gno.land/r/gnoland/blog"
denom := r.CoinDenom("blgcoin") // /gno.land/r/gnoland/blog:blgcoin
```

---

## Testing

```go
func TestSkipHeights(count int64)
func TestSetOriginCaller(addr Address)
func TestSetOriginPkgAddress(addr Address)
func TestSetOriginSend(sent, spent Coins)
func TestIssueCoins(addr Address, coins Coins)
func TestSetRealm(realm Realm)
func NewUserRealm(address Address) Realm
func NewCodeRealm(pkgPath string) Realm
```

### TestSkipHeights

```go
func TestSkipHeights(count int64)
```
Modifies the block height variable by skipping **count** blocks.

It also increases block timestamp by 5 seconds for every single count

#### Usage
```go
std.TestSkipHeights(100)
```
---

### TestSetOriginCaller

```go
func TestSetOriginCaller(addr Address)
```
Sets the current caller of the transaction to **addr**.

#### Usage
```go
std.TestSetOriginCaller(std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"))
```
---

### TestSetOriginPkgAddress

```go
func TestSetOriginPkgAddress(addr Address)
```
Sets the call entry realm address to **addr**.

#### Usage
```go
std.TestSetOriginPkgAddress(std.Address("g1ecely4gjy0yl6s9kt409ll330q9hk2lj9ls3ec"))
```

---

### TestSetOriginSend

```go
func TestSetOriginSend(sent, spent Coins)
```
Sets the sent & spent coins for the current context.

#### Usage
```go
std.TestSetOriginSend(sent, spent Coins)
```
---

### TestIssueCoins

```go
func TestIssueCoins(addr Address, coins Coins)
```

Issues testing context **coins** to **addr**.

#### Usage

```go
issue := std.Coins{{"coin1", 100}, {"coin2", 200}}
addr := std.Address("g1ecely4gjy0yl6s9kt409ll330q9hk2lj9ls3ec")
std.TestIssueCoins(addr, issue)
```

---

### TestSetRealm

```go
func TestSetRealm(rlm Realm)
```

Sets the realm for the current frame. After calling `TestSetRealm()`, calling
[`CurrentRealm()`](#currentrealm) in the same test function will yield the value of `rlm`, and
any `PreviousRealm()` called from a function used after TestSetRealm will yield `rlm`.

Should be used in combination with [`NewUserRealm`](#newuserrealm) &
[`NewCodeRealm`](#newcoderealm).

#### Usage
```go
addr := std.Address("g1ecely4gjy0yl6s9kt409ll330q9hk2lj9ls3ec")
std.TestSetRealm(std.NewUserRealm(""))
// or
std.TestSetRealm(std.NewCodeRealm("gno.land/r/demo/users"))
```

---

### NewUserRealm

```go
func NewUserRealm(address Address) Realm
```

Creates a new user realm for testing purposes.

#### Usage
```go
addr := std.Address("g1ecely4gjy0yl6s9kt409ll330q9hk2lj9ls3ec")
userRealm := std.NewUserRealm(addr)
```

---

### NewCodeRealm

```go
func NewCodeRealm(pkgPath string) Realm
```

Creates a new code realm for testing purposes.

#### Usage
```go
path := "gno.land/r/demo/boards"
codeRealm := std.NewCodeRealm(path)
```
