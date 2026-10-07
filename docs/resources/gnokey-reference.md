# `gnokey` command reference

`gnokey` is the official command-line client for Gno.land. This reference lists
every `gnokey` command and covers deploying packages, calling and scripting
realms, and reading chain state.

`gnokey` is a production tool and stays deliberately minimal: signing keys and
talking to a live network, nothing more. Developer conveniences, for example a
transaction template system or realm scaffolding, would belong in a development
tool that could come in the future, the same way
[`gnodev`](./gnodev.md) exists so you can develop locally without
driving a full node through the `gnoland` toolchain.

For everyday wallet use, see
[Using the `gnokey` wallet](../users/using-gnokey.md). For a guided first deploy,
follow [Getting started](../builders/getting-started.md). If you don't have
`gnokey` yet, see [Installation](../builders/install.md).

## Commands

| Command | What it does |
|---------|-------------|
| [`add`](../users/using-gnokey.md#managing-key-pairs) | create or import a key pair |
| [`add bech32`](../../gno.land/cmd/gnokey/README.md#multisig-k-of-n) | add a watch-only key from a bech32 public key |
| [`add multisig`](../../gno.land/cmd/gnokey/README.md#multisig-k-of-n) | create a multisig key from member keys |
| `add ledger` | add a key from a Ledger device |
| [`list`](../users/using-gnokey.md#managing-key-pairs) | list keys in a keybase |
| `delete` | delete a key |
| [`export`](../../gno.land/cmd/gnokey/README.md#exporting-and-importing-keys) | export a private key as encrypted armor |
| [`import`](../../gno.land/cmd/gnokey/README.md#exporting-and-importing-keys) | import an encrypted private key |
| `generate` | generate a BIP39 mnemonic |
| `rotate` | change a key's keybase password |
| [`maketx`](#making-transactions) | build, sign, and broadcast transactions |
| [`maketx enablepkg`](#enablepkg-and-rejectpkg) | activate a package awaiting approval |
| [`maketx rejectpkg`](#enablepkg-and-rejectpkg) | remove a package awaiting approval |
| [`maketx session`](../../gno.land/cmd/gnokey/README.md#session) | create, revoke, or revokeall session accounts |
| [`query`](#querying-a-gnoland-network) | read chain state without spending gas |
| [`sign`](../../gno.land/cmd/gnokey/README.md#airgapped-signing) | sign an unsigned transaction |
| [`broadcast`](../../gno.land/cmd/gnokey/README.md#airgapped-signing) | broadcast a signed transaction |
| [`verify`](../../gno.land/cmd/gnokey/README.md#verifying-a-signature) | verify a transaction signature |
| [`multisign`](../../gno.land/cmd/gnokey/README.md#multisig-k-of-n) | combine multisig signatures |
| `version` | print the `gnokey` binary version |

## Making transactions

Four message types change on-chain state:

| Message      | What it does                            |
|--------------|-----------------------------------------|
| `AddPackage` | upload new code to the chain            |
| `Call`       | call an exported realm function         |
| `Send`       | transfer coins between addresses        |
| `Run`        | execute a Gno script against the chain  |

Each `maketx` command sends one message, signed with a key from your keybase (see
[Using the `gnokey` wallet](../users/using-gnokey.md#managing-key-pairs)). Every
command takes the same base-configuration flags:

- `-gas-wanted` - the maximum gas units the transaction may consume (required)
- `-gas-fee` - the fee paid for the transaction, as `<amount>ugnot`
  (e.g. `1000000ugnot`; required)
- `-chainid` and `-remote` - the network to target; the two must match
- `-broadcast` - send the transaction to the chain (default `true`; set
  `-broadcast=false` to build the unsigned transaction without sending it, as
  in [Airgapped signing](../../gno.land/cmd/gnokey/README.md#airgapped-signing))
- `-memo` - arbitrary text attached to the transaction (optional)
- `-simulate` - simulation mode: `test` (default, simulate first, broadcast
  only on success), `skip` (broadcast without simulating), `only` (dry run,
  report gas used and exit without broadcasting)
- `-gas-fee-margin` - percentage added to the estimated gas fee (default `5`;
  only used with `-simulate only`)
- `-master` - the master account's key name or address, when signing with a
  session key (optional; see [Session](../../gno.land/cmd/gnokey/README.md#session))

`-gas-wanted` and `-gas-fee` together cap what you pay; `gnokey` never fills
them in for you. Run the transaction with `-simulate only` to get good values,
as shown in [Gas estimation](./gas-fees.md#gas-estimation). The default
`-simulate test` then guards them: a transaction that fails simulation is never
broadcast, and no fee is spent. Find `-chainid` and `-remote`
values per network in [Network configuration](./gnoland-networks.md).
State-changing calls cost gas paid in GNOT, so on testnets grab some from the
[Faucet Hub](https://faucet.gno.land) first.

Every successful transaction prints the same summary; a `Call` return value
prints above the `OK!` line:

```console
OK!
GAS WANTED: 200000
GAS USED:   117564
HEIGHT:     3990
EVENTS:     []
INFO:
TX HASH:    Ni8Oq5dP0leoT/IRkKUKT18iTv8KLL3bH8OFZiV79kM=
```

- `GAS WANTED` - the gas units you requested
- `GAS USED` - the gas actually consumed
- `HEIGHT` - the block the transaction landed in
- `EVENTS` - any [Gno events](./gno-stdlibs.md#events) the call emitted
- `INFO` - extra information from the message handler (usually empty)
- `TX HASH` - the transaction's hash

Transactions that change [storage deposits](./storage-deposit.md) add
`STORAGE DELTA`, `STORAGE FEE` (or `STORAGE REFUND`), and `TOTAL TX COST`
lines, and `addpkg` appends a `PKGPATH` line.

For an end-to-end deploy-and-call walkthrough, see
[Getting started](../builders/getting-started.md).

### `Send`

`Send` transfers coins between two addresses with `gnokey maketx send`. Its own
flags are:

- `-to` - the recipient's bech32 address
- `-send` - the amount to transfer, as `<amount><denom>` (e.g. `100ugnot`)

```bash
gnokey maketx send \
  -to g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 \
  -send 100000ugnot \
  -gas-fee 1000000ugnot -gas-wanted 2000000 \
  -chainid staging \
  -remote "https://rpc.staging.gno.land:443" \
  mykey
```

### `AddPackage`

`AddPackage` uploads new code to the chain with `gnokey maketx addpkg`. On top of
the base configuration, it takes flags of its own:

- `-pkgpath` - the on-chain path the code is published to
- `-pkgdir` - the local directory holding the code
- `-send` - coins to send to the realm with the deploy (optional)
- `-max-deposit` - cap on GNOT locked for [storage deposit](./storage-deposit.md) (optional)

Run it from the package directory, publishing to a path under a
[namespace](./users-and-teams.md) you own:

```bash
gnokey maketx addpkg \
  -pkgpath "gno.land/p/examplenamespace/hello_world" \
  -pkgdir "." \
  -gas-fee 1000000ugnot \
  -gas-wanted 20000000 \
  -chainid staging \
  -remote "https://rpc.staging.gno.land:443" \
  mykey
```

`-pkgpath` decides where the code lives on chain; on deploy it overwrites the
`module` declared in the package's `gnomod.toml`, so keep the two in sync. For
writing the package and declaring that path, see
[Getting started](../builders/getting-started.md) and
[Configuring Gno projects](./configuring-gno-projects.md#gnomodtoml).

### `Call`

`Call` invokes an exported realm function with `gnokey maketx call`. Its own flags
are:

- `-pkgpath` - the realm's on-chain path
- `-func` - the function to call
- `-args` - one argument (repeat the flag for more; see below)
- `-send` - coins to send with the call (optional)
- `-max-deposit` - cap on GNOT locked for [storage deposit](./storage-deposit.md) (optional)

`-func` must name an exported crossing function, one declared with a leading
`cur realm` parameter. Non-crossing functions are rejected; read them with
[`vm/qeval`](#vmqeval), or call them from [`Run`](#run).

For example, calling `Deposit()` on the `gno.land/r/gnoland/wugnot` realm to wrap
`1000ugnot` into the GRC20 token `wugnot`:

```bash
gnokey maketx call \
  -pkgpath "gno.land/r/gnoland/wugnot" \
  -func "Deposit" \
  -send "1000ugnot" \
  -gas-fee 10000000ugnot \
  -gas-wanted 2000000 \
  -chainid staging \
  -remote "https://rpc.staging.gno.land:443" \
  mykey
```

:::info `Call` always uses gas

`maketx call` spends gas even when the function only reads state. To read without
paying, use the [`vm/qeval`](#vmqeval) query instead.

:::

#### Variadic functions

Pass one `-args` flag per variadic element. Given:

```go
func Add(cur realm, nums ...int) int
```

call it with any number of arguments:

```bash
# Two variadic args (base flags omitted)
gnokey maketx call -pkgpath gno.land/r/tests/vm/variadic -func Add -args 10 -args 20 ...

# Zero variadic args (omit -args entirely)
gnokey maketx call -pkgpath gno.land/r/tests/vm/variadic -func Add ...
```

Slice expansion (passing `nums...`) is not supported; pass each element as its
own `-args`.

### `Run`

`Run` executes a Gno script against on-chain code with `gnokey maketx run`. Write a
`main` package; its `main()` function is detected and run, and any state changes
are applied. Its own flags are:

- `-send` - coins to send with the run (optional)
- `-max-deposit` - cap on GNOT locked for [storage deposit](./storage-deposit.md) (optional)

For example, calling `Increment()` on the
[Counter realm](https://staging.gno.land/r/demo/counter):

```go
package main

import "gno.land/r/demo/counter"

func main(cur realm) {
	println(counter.Increment(cross(cur)))
}
```

```bash
gnokey maketx run \
  -gas-fee 1000000ugnot \
  -gas-wanted 20000000 \
  -chainid staging \
  -remote "https://rpc.staging.gno.land:443" \
  mykey ./script.gno
```

`println` lets you see the return value: its output is surfaced only in `Run`
and in tests, and discarded in a `Call`.

#### When to use `Run` over `Call`

That example could just as easily have been a `maketx call`. `Run` earns its place
when a plain call can't express what you need:

1. Constructing composite arguments such as structs, maps, or slices, which
   `Call` cannot pass
2. Calling realm functions repeatedly in a loop
3. Calling methods on exported variables

**1. Composite arguments.** `-args` only carries primitive values: booleans,
numbers, strings, and base64-encoded `[]byte`. Structs, maps, and other slices
cannot be passed. `Run` is full Gno code, so it can build them directly:

```go
package main

import "gno.land/r/myrealm"

func main(cur realm) {
	post := myrealm.Post{
		Title: "Hello",
		Tags:  []string{"gno", "blockchain"},
	}
	myrealm.CreatePost(cross(cur), post)
}
```

**2. Looping over a realm function.** `Call` sends one transaction per call.
`Run` batches multiple calls into a single transaction, saving gas and keeping
the changes atomic. Using the [counter realm](https://staging.gno.land/r/demo/counter)
from the example above:

```go
package main

import "gno.land/r/demo/counter"

func main(cur realm) {
	for i := 0; i < 5; i++ {
		println(counter.Increment(cross(cur)))
	}
}
```

This increments the counter five times in one transaction, printing each new
value.

**3. Methods on exported variables.** `Call` only invokes exported functions;
`Run` can also call methods on exported variables. For example, if a realm
exposes `var Pool *Token` with a `Balance()` method:

```go
package main

import "gno.land/r/myrealm"

func main() {
	println(myrealm.Pool.Balance())
}
```

### `enablepkg` and `rejectpkg`

On a chain whose `code_submission_policy` parameter is `inert`, as mainnet's is,
`addpkg` parks the package rather than deploying it: the code is stored, but
nothing type-checks it, runs it, or can import it until an address listed in the
`pkg_approvers` parameter activates it. List what is waiting with
[`vm/qinertpaths`](#vmqinertpaths).

`gnokey maketx enablepkg` activates a parked package, and only an approver can
send it. Its own flags are:

- `-pkgpath` - the parked package's path (required)
- `-pkgdir` - a local copy of the source you reviewed, hashed so the approval
  names those exact bytes
- `-pkg-hash` - the content hash, when it was computed elsewhere; use instead of
  `-pkgdir`
- `-pkg-height` - the block the reviewed submission landed in, so any
  re-submission invalidates the approval (optional)

Hash your own reviewed copy, never one read from the chain: the submitter can
replace the parked bytes at any time, and a hash taken from the chain approves
whatever is parked at that moment.

`gnokey maketx rejectpkg -pkgpath <path>` removes a parked package. An approver
or the address that submitted it can send it, and the submission charge is not
refunded.

## Operator workflows

Airgapped signing, multisig, session accounts, and key export and import are
covered in the [`gnokey` README](../../gno.land/cmd/gnokey/README.md).

## Querying a Gno.land network

`gnokey query` sends ABCI queries, which read network state without spending gas.
Every query needs a `-remote` to read from; `-data` carries the query argument,
and `-height` reads the state at a specific block instead of the latest one. The
available queries:

- `auth/accounts/{ADDRESS}` - account information
- `auth/accounts/{ADDRESS}/sessions` - the [session](../../gno.land/cmd/gnokey/README.md#session) accounts of an address
- `auth/gasprice` - the current minimum gas price for transactions
- `bank/balances/{ADDRESS}` - account balances
- `bank/supply/{DENOM}` - the total supply of a denomination
- `params/{MODULE}:{SUBMODULE}:{NAME}` - a module parameter, e.g.
  `params/vm:gno.land/r/myrealm:foo`
- `vm/qfuncs` - the exported functions of a realm
- `vm/qfile` - the file list or file contents of a package path
- `vm/qdoc` - the documentation of a package path, as JSON
- `vm/qeval` - evaluate an expression in read-only mode
- `vm/qrender` - call a realm's `Render` function and return its output
- `vm/qpaths` - list existing package paths
- `vm/qinertpaths` - list package paths parked awaiting approval
- `vm/qstorage` - a realm's storage usage and locked deposit

For the machine-readable variants (`vm/qeval_json`, `vm/qobject_json`, and
friends), see [Query state API](../builders/query-state-api.md).

### `auth/accounts`

Returns information about an address:

```bash
gnokey query auth/accounts/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 -remote https://rpc.staging.gno.land:443
```

```console
height: 0
data: {
  "BaseAccount": {
    "address": "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
    "coins": "227984898927ugnot",
    "public_key": {
      "@type": "/tm.PubKeySecp256k1",
      "value": "A+FhNtsXHjLfSJk1lB8FbiL4mGPjc50Kt81J7EKDnJ2y"
    },
    "account_number": "0",
    "sequence": "12"
  },
  "attributes": "0"
}
```

`height` is currently always `0`; module queries leave it unset. The state
itself is read at the latest block, or at `-height` if given.

In `data`, the `BaseAccount` object is the TM2 struct for account data, and
gno.land adds an `attributes` field next to it:

- `address` - the account's address
- `coins` - the gas-denom coins the account owns, not the full balance:
  every other denom lives in its own keys and shows up only under
  [`bank/balances`](#bankbalances)
- `public_key` - the TM2 public key the address derives from
- `account_number` - a unique identifier for the account on chain
- `sequence` - a nonce, used to protect against replay attacks
- `attributes` - gno.land status flags as a bitset, such as frozen, validator
  account, or token-lock whitelisted; `"0"` for a plain account

### `bank/balances`

Returns the [coin](./gno-stdlibs.md#coin) balances of an address:

```bash
gnokey query bank/balances/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 -remote https://rpc.gno.land:443
```

```console
height: 0
data: "227984898927ugnot"
```

### `bank/supply`

Returns how much of a denomination exists across all accounts. A denomination
nobody holds reads `0`, and so does an unknown one:

```bash
gnokey query bank/supply/ugnot -remote https://rpc.gno.land:443
```

A realm-issued denomination is `/{PKGPATH}:{NAME}` and carries its own slashes,
so it goes straight into the path:

```bash
gnokey query bank/supply//gno.land/r/demo/foo:gold -remote https://rpc.gno.land:443
```

```console
height: 0
data: "1000000"
```

The amount comes back quoted, which is how `int64` renders on the wire.

### `auth/gasprice`

Returns the minimum gas price currently required for transactions, handy for
setting `-gas-fee`:

```bash
gnokey query auth/gasprice -remote https://rpc.gno.land:443
```

```console
height: 0
data: {
  "gas": "1000",
  "price": "100ugnot"
}
```

`data` holds a `GasPrice`: `gas` is the gas units and `price` is their cost as a
[coin](./gno-stdlibs.md#coin). The network adjusts the price after each block based
on demand; this query returns the value from the most recently completed block, the
minimum for new transactions. For a deeper explanation, see
[Gas Price](./gas-fees.md#gas-price).

### `vm/qfuncs`

Returns the exported functions of a realm path, given with `-data`; package
(`/p/`) paths are rejected. Crossing functions list their leading `cur realm`
parameter:

```bash
gnokey query vm/qfuncs --data "gno.land/r/gnoland/wugnot" -remote https://rpc.gno.land:443
```

```
height: 0
data: [
        {
          "FuncName": "Deposit",
          "Params": null,
          "Results": null
        },
        {
          "FuncName": "Withdraw",
          "Params": [
            {
            "Name": "amount",
            "Type": "int64",
            "Value": ""
            }
          ],
          "Results": null
        },
        // other functions
]
```

### `vm/qfile`

Returns the contents of a package path, given with `-data`. With only the package
path, it lists the files:

```bash
gnokey query vm/qfile -data "gno.land/r/gnoland/wugnot" -remote https://rpc.gno.land:443
```

```console
height: 0
data: gnomod.toml
wugnot.gno
z0_filetest.gno
```

With a file name appended to the path, it returns that file's source:

```bash
gnokey query vm/qfile -data "gno.land/r/gnoland/wugnot/wugnot.gno" -remote https://rpc.gno.land:443
```

```console
height: 0
data: package wugnot

import (
        "chain"
        "chain/banker"
        "chain/runtime"
        "strings"

        "gno.land/p/nt/grc20/v0"
        "gno.land/p/nt/ufmt/v0"
        "gno.land/r/nt/grc20reg/v0"
)
...
```

### `vm/qdoc`

Returns the documentation of a package path, given with `-data`, as JSON. It covers
the package itself and its functions, types, and values:

```bash
gnokey query vm/qdoc --data "gno.land/r/gnops/valopers" -remote https://rpc.gno.land:443
```

```
height: 0
data: {
  "package_path": "gno.land/r/gnops/valopers",
  "package_doc": "Package valopers is designed around the permissionless lifecycle of valoper profiles.\n",
  "funcs": [
    {
      "name": "GetByAddr",
      "signature": "func GetByAddr(addr address) Valoper",
      "doc": "GetByAddr fetches the valoper using the operator address, if present.\n",
      "params": [{ "name": "addr", "type": "address", "doc": "" }],
      "results": [{ "name": "", "type": "Valoper", "doc": "" }]
    }
    // other funcs
  ],
  "types": [
    {
      "name": "Valoper",
      "type": "struct { ... }",
      "doc": "Valoper represents a validator operator profile.\n"
    }
  ]
  // values omitted
}
```

### `vm/qeval`

Evaluates a Gno expression, typically a call to an exported function, in
read-only mode without paying gas:

```bash
gnokey query vm/qeval -remote https://rpc.gno.land:443 -data "gno.land/r/gnoland/wugnot.BalanceOf(\"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5\")"
```

This returns the `wugnot` balance of the address without a transaction.
Quotation marks around string arguments must be escaped, and arguments must be
literal expressions. Queries still run under an internal gas cap, so unbounded
evaluations can fail.

### `vm/qrender`

Evaluates `Render(<renderpath>)` on a package path, given as
`<pkgpath>:<renderpath>`. The colon is required; with nothing after it, the
realm renders its root path:

```bash
gnokey query vm/qrender --data "gno.land/r/gnoland/wugnot:" -remote https://rpc.staging.gno.land:443
```

```console
height: 0
data: # wrapped GNOT ($wugnot)

* **Decimals**: 0
* **Total supply**: 5012404
* **Known accounts**: 2
```

:::info Specifying a path to `Render()`

To render a specific path, use the `<pkgpath>:<renderpath>` syntax. For example,
the `wugnot` realm renders the balance of an address at:

```bash
gnokey query vm/qrender --data "gno.land/r/gnoland/wugnot:balance/g125em6arxsnj49vx35f0n0z34putv5ty3376fg5" -remote https://rpc.gno.land:443
```

The realm decides what paths it accepts; look at its `Render()` function to see
which ones it handles.

:::

### `vm/qpaths`

Lists existing package paths that start with the prefix given via `-data`. With
no prefix, it lists every known path, including those from `stdlibs`:

```bash
gnokey query vm/qpaths --data "gno.land/r/gnoland" -remote https://rpc.gno.land:443
```

```console
height: 0
data: gno.land/r/gnoland/blog
gno.land/r/gnoland/coins
gno.land/r/gnoland/events
gno.land/r/gnoland/home
gno.land/r/gnoland/pages
```

A prefix can also be a `@username`, which lists that user's `/p` and `/r`
sub-packages:

```bash
gnokey query vm/qpaths --data "@foo" -remote https://rpc.gno.land:443
```

Append `?limit=<x>` to cap the number of results (default `1000`; values above
`10000` are capped to `10000`). Quote the whole query string so the shell keeps
it in one piece:

```bash
gnokey query "vm/qpaths?limit=3" --data "gno.land/r/gnoland" -remote https://rpc.gno.land:443
```

### `vm/qinertpaths`

Lists the paths of packages parked awaiting approval (see
[`enablepkg` and `rejectpkg`](#enablepkg-and-rejectpkg)) that start with the
prefix given via `-data`. `vm/qpaths` never lists them, since a parked package
is not live. It takes the same `?limit=<x>` suffix:

```bash
gnokey query "vm/qinertpaths?limit=3" -remote https://rpc.gno.land:443
```

```console
height: 0
data: gno.land/r/g1n4pl5uc4yt5r96m9w6fmdznx3x0jyg8l6arhmt/bazaar/genesis
gno.land/r/g1n4pl5uc4yt5r96m9w6fmdznx3x0jyg8l6arhmt/bazaar/gnft
gno.land/r/g1n4pl5uc4yt5r96m9w6fmdznx3x0jyg8l6arhmt/zdex/v1
```

### `vm/qstorage`

Returns the current storage usage and deposit of a realm:

```bash
gnokey query vm/qstorage --data "gno.land/r/foo" -remote https://rpc.gno.land:443
```

```
height: 0
data: storage: 5025, deposit: 502500
```

`storage` is the total bytes used; `deposit` is the total GNOT locked by the realm.
Dividing the two gives the storage price (`502500/5025 = 100ugnot` per byte),
without querying the chain parameters.
