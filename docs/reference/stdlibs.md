# Standard Libraries

Gno comes with a set of standard libraries which are included to ease development
and provide extended functionality to the language. These include:
- standard libraries as we know them in classic Go, i.e. `strings`, `testing`, etc.
- a special `std` package, which contains types, interfaces, and APIs created to 
handle blockchain-related functionality, such as fetching the last caller, 
fetching coins sent along with a transaction, getting the block timestamp and height, and more. 

Standard libraries differ from on-chain packages in terms of their import path structure.
Unlike on-chain [packages](../packages.md), standard libraries do not incorporate
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
issued by the native Gno [Banker](banker.md).  

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

Read more about coins in the [Effective Gno](../../misc/effective-gno.md) section. 

The Coin(s) API can be found in under the `std` package [reference](../../reference/std.md#coin).


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

The Banker API can be found under the `std` package [reference](../../reference/std.md#banker).

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
[Effective Gno](../../misc/effective-gno.md#emit-gno-events-to-make-life-off-chain-easier),
and the `Emit()` reference [here](../../reference/std.md#emit).

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
