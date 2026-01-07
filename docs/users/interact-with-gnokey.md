# Interacting with Gno.land using gnokey

`gnokey` is the official command-line wallet and utility for interacting with
Gno.land networks. It allows to manage keys, query the blockchain, send
transactions, and deploy smart contracts. This guide will help you get started
with the essential operations.

## Installing gnokey

To build and install from source, you'll need:
- Git
- Go 1.22+
- Make

```bash
# Clone the repository
git clone https://github.com/gnolang/gno.git
cd gno

# Install gnokey
make install
```

## Managing key pairs

A key pair is required to send transactions to the blockchain, including 
deploying code, interacting with existing applications, and transferring coins.

## A word about key pairs

Key pairs are the foundation of blockchain interactions. A 12-word or 24-word 
[mnemonic phrase](https://www.zimperium.com/glossary/mnemonic-seed/) generates 
a private and public key. Your public key derives your address (your unique 
identifier), while your private key signs transactions, proving ownership.

## Generating a key pair

Generate a new key pair locally:

```bash
gnokey add MyKey
```

You'll be prompted for a password to encrypt the key pair. The output shows:
- Your public key and Gno address (starting with `g1`)
- Your 12-word mnemonic phrase

:::warning Safeguard your mnemonic phrase!

Your **mnemonic phrase** can regenerate your key pairs. Store it safely offline 
(write it down on paper). **If lost, it cannot be recovered.**

:::

Key pairs are stored in a keybase directory (see `-home` flag).

### Gno addresses

Your **Gno address** (starting with `g1`) is your unique identifier on the network. 
It's visible in transactions and used to receive [coins](../resources/gno-stdlibs.md#coin).

## Making transactions

Four message types can change on-chain state:
- `AddPackage` - adds new code to the chain
- `Call` - calls a realm function
- `Send` - sends coins between addresses
- `Run` - executes a Gno script

Each transaction requires:
- Base configuration (`gas-fee`, `gas-wanted`, etc.)
- One or more messages to execute

`gnokey` supports single-message transactions. For multiple-message transactions, 
use [gnoclient](https://github.com/gnolang/gno/tree/master/gno.land/pkg/gnoclient) in Go programs.

:::info Getting testnet tokens

Visit [Faucet Hub](https://faucet.gno.land) to get GNOTs for testnets.

:::

## `AddPackage`

Upload new code to the chain:

```bash
gnokey maketx addpkg
```

Let's create a simple "Hello world" [pure package](../resources/gno-packages.md):

```bash
mkdir -p example/p
cd example/p
touch hello_world.gno
```

In `hello_world.gno`:

```go
package hello_world

func Hello() string {
  return "Hello, world!"
}
```

Create the required `gnomod.toml` file:

```bash
gno mod init "gno.land/p/<your_namespace>/hello_world"
```

The module path must match the `-pkgpath` flag used when uploading.

:::info About `gnomod.toml`

This manifest file defines the module path for imports and package resolution. 
Required for all packages and realms.

:::

Key flags for `addpkg`:
- `-pkgpath` - on-chain path for your package
- `-pkgdir` - local directory containing your code
- `-broadcast` - broadcast transaction to chain
- `-gas-wanted` / `-gas-fee` - gas configuration (see [Gas Fees](../resources/gas-fees.md))
- `-chainid` / `-remote` - network configuration

Upload the package to [Staging](../resources/gnoland-networks.md):

```bash
gnokey maketx addpkg \
-pkgpath "gno.land/p/<your_namespace>/hello_world" \
-pkgdir "." \
-gas-fee 10000000ugnot \
-gas-wanted 200000 \
-broadcast \
-chainid staging \
-remote "https://rpc.gno.land:443" \
mykey
```

Replace `<your_namespace>` with your [namespace](../resources/users-and-teams.md).

Output:

```console
OK!
GAS WANTED: 200000
GAS USED:   117564
HEIGHT:     3990
EVENTS:     []
TX HASH:    Ni8Oq5dP0leoT/IRkKUKT18iTv8KLL3bH8OFZiV79kM=
```

Let's analyze the output, which is standard for any `gnokey` transaction:
- `GAS WANTED: 200000` - the original amount of gas specified for the transaction
- `GAS USED:   117564` - the gas used to execute the transaction
- `HEIGHT:     3990` - the block number at which the transaction was executed at
- `EVENTS:     []` - [Gno events](../resources/gno-stdlibs.md#events) emitted by the transaction, in this case, none
- `TX HASH:    Ni8Oq5dP0leoT/IRkKUKT18iTv8KLL3bH8OFZiV79kM=` - the hash of the transaction

Congratulations! You have just uploaded a pure package to the Staging network.
If you wish to deploy to a different network, find the list of all network
configurations in the [Network Configuration](../resources/gnoland-networks.md) section.

## `Call`

The `Call` message type is used to call any exported realm function.
You can send a `Call` transaction with `gnokey` using the following command:

```bash
gnokey maketx call
```

:::info Gas-free queries

`Call` uses gas even for read-only functions. Use `vm/qeval` [queries](#vmqeval) 
for gas-free reads.

:::

Example - wrapping GNOTs using the `wugnot` realm:

```bash
gnokey maketx call \
-pkgpath "gno.land/r/gnoland/wugnot" \
-func "Deposit" \
-send "1000ugnot" \
-gas-fee 10000000ugnot \
-gas-wanted 2000000 \
-broadcast \
-chainid staging \
-remote "https://rpc.gno.land:443" \
mykey
```

The output shows an [event](../resources/gno-stdlibs.md#events) emitted by the `Deposit()` function.

Verify the balance using a gas-free query:

```bash
gnokey query vm/qeval -remote "https://rpc.gno.land:443" -data "gno.land/r/gnoland/wugnot.BalanceOf(\"<your_address>\")"
```

## `Send`

Transfer coins between addresses:
```bash
gnokey maketx send \
-to g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 \
-send 100ugnot \
-gas-fee 10000000ugnot \
-gas-wanted 2000000 \
-broadcast \
-chainid staging \
-remote "https://rpc.gno.land:443" \
mykey
```

Check balances using the [`bank/balances`](#bankbalances) query.

## `Run`

Execute Gno scripts against on-chain code. Example using the [Userbook realm](https://gno.land/r/demo/userbook):

Create `script.gno`:

```go
package main

import "gno.land/r/demo/userbook"

func main() {
  println(userbook.SignUp())
}
```

Run the script:

```bash
gnokey maketx run \
-gas-fee 1000000ugnot \
-gas-wanted 20000000 \
-broadcast \
-chainid staging \
-remote "https://rpc.gno.land:443" \
mykey ./script.gno
```

The chain executes the script and applies state changes. Using `println` (available in 
`Run` and testing contexts) displays function return values.

### Advanced `Run` capabilities

`Run` excels in three scenarios:

**1. Loop multiple calls:**

```go
package main

import "gno.land/r/docs/examples/foo"

func main() {
  for i := 0; i < 5; i++ {
    println(foo.Render(""))
  }
}
```

**2. Non-primitive arguments** (`Call` only supports primitives like strings, numbers, booleans):

```go
package main

import (
  "strconv"
  "gno.land/r/docs/examples/foo"
)

func main() {
  var multipleFoos []*foo.Foo
  for i := 0; i < 5; i++ {
    multipleFoos = append(multipleFoos, foo.NewFoo("bar"+strconv.Itoa(i), i))
  }
  foo.AddFoos(multipleFoos)
}
```

**3. Call methods on exported variables:**

```go
package main

import "gno.land/r/docs/examples/foo"

func main() {
  println(foo.MainFoo.String())
}
```

## Making an airgapped transaction

Create, sign, and broadcast transactions securely using separate online/offline machines. 
This provides maximum security by keeping private keys offline ([airgap](https://en.wikipedia.org/wiki/Air_gap_(networking))).

Workflow:
1. **Online machine**: Fetch account information
2. **Offline machine**: Create and sign transaction
3. **Online machine**: Broadcast transaction

### 1. Fetch account information (online)

```bash
gnokey query auth/accounts/<your_address> -remote "https://rpc.gno.land:443"
```

We need to extract the account number and sequence from the output:

```bash
height: 0
data: {
  "BaseAccount": {
    "address": "g1zzqd6phlfx0a809vhmykg5c6m44ap9756s7cjj",
    "coins": "10000000ugnot",
    "public_key": null,
    "account_number": "468",
    "sequence": "0"
  }
}
```

Extract `account_number` (`468`) and `sequence` (`0`) for signing.

### 2. Create unsigned transaction (offline)

```bash
gnokey maketx call \
-pkgpath "gno.land/r/demo/userbook" \
-func "SignUp" \
-gas-fee 1000000ugnot \
-gas-wanted 2000000 \
mykey > userbook.tx
```

Creates `userbook.tx` with null signature. Note: no `-broadcast` flag, so the transaction 
is not sent to the chain.

### 3. Sign transaction (offline)

Sign using account number and sequence from step 1:

```bash
gnokey sign \
-tx-path userbook.tx \
-chainid "staging" \
-account-number 468 \
-account-sequence 0 \
mykey
```

After entering the password, the signature field is populated.

### 4. Broadcast transaction (online)

```bash
gnokey broadcast -remote "https://rpc.gno.land:443" userbook.tx
```

### Verify transaction signature

Verify signature correctness (signature must be in `hex` format):

```bash
gnokey verify -docpath userbook.tx mykey <signature>
```

## Querying a Gno.land network

Use ABCI queries to read blockchain state without spending gas. Send queries using 
`gnokey query` with the appropriate subcommand.

Below is a list of queries a user can make with `gnokey`:
- `auth/accounts/{ADDRESS}` - account information
- `auth/gasprice` - current minimum gas price
- `bank/balances/{ADDRESS}` - account balances
- `vm/qfuncs` - exported functions for a package
- `vm/qfile` - package file contents
- `vm/qdoc` - package documentation
- `vm/qeval` - evaluate expressions in read-only mode
- `vm/qrender` - render output for a package
- `vm/qpaths` - list existing package paths
- `vm/qstorage` - storage usage and deposit

## `auth/accounts`

Get information about a specific address:

```bash
gnokey query auth/accounts/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 -remote https://rpc.gno.land:443
```

Output:

```bash
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
  }
}
```

The return data contains:
- `height` - query execution height (currently `0` by default)
- `data` - query result

The `BaseAccount` struct contains:
- `address` - the address of the account
- `coins` - the list of coins the account owns
- `public_key` - the TM2 public key of the account, from which the address is derived
- `account_number` - a unique identifier for the account on the Gno.land chain
- `sequence` - a nonce, used for protection against replay attacks

## `bank/balances`

Fetch [coin](../resources/gno-stdlibs.md#coin) balances of an account:

```bash
gnokey query bank/balances/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 -remote https://rpc.gno.land:443
```

Output:

```bash
height: 0
data: "227984898927ugnot"
```

## `auth/gasprice`

Fetch the minimum gas price required for transactions:

```bash
gnokey query auth/gasprice -remote https://rpc.gno.land:443
```

Output:

```bash
height: 0
data: {
  "gas": 1000,
  "price": "100ugnot"
}
```

The `GasPrice` object contains:
- `gas` - the gas units
- `price` - the price for those gas units in the form of a [coin](../resources/gno-stdlibs.md#coin)

The network adjusts the gas price after each block based on demand. This query returns
the minimum gas price required for new transactions.
For more details, see [Gas Price](../resources/gas-fees.md#gas-price).

## `vm/qfuncs`

Fetch exported functions from a package path using the `-data` flag:

```bash
gnokey query vm/qfuncs --data "gno.land/r/gnoland/wugnot" -remote https://rpc.gno.land:443
```

Returns all exported functions for the package.

```json
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
            "Type": "uint64",
            "Value": ""
            }
          ],
          "Results": null
        },
        // other functions
]
```

## `vm/qfile`

Fetch files and their content from a package path:

```bash
gnokey query vm/qfile -data "gno.land/r/gnoland/wugnot" -remote https://rpc.gno.land:443
```

Lists all files in the package. To retrieve a specific file's source code, add the 
filename to the path:

```bash
height: 0
data: gnomod.toml
wugnot.gno
z0_filetest.gno
```

```bash
gnokey query vm/qfile -data "gno.land/r/gnoland/wugnot/wugnot.gno" -remote https://rpc.gno.land:443
```

## `vm/qdoc`

Fetch documentation for functions, types, and variables from a package:

```bash
gnokey query vm/qdoc --data "gno.land/r/gnoland/valopers/v2" -remote https://rpc.gno.land:443
```

Returns JSON with package documentation, functions, types, and values.

```json
height: 0
data: {
  "package_path": "gno.land/r/gnoland/valopers/v2",
  "package_line": "package valopers // import \"valopers\"",
  "package_doc": "Package valopers is designed around the permissionless lifecycle of valoper profiles. It also includes parts designed for govdao to propose valset changes based on registered valopers.\n",
  "values": [
    {
      "name": "valopers",
      "doc": "// Address -> Valoper\n",
      "type": "*avl.Tree"
    }
    // other values
  ],
  "funcs": [
    {
      "type": "",
      "name": "GetByAddr",
      "signature": "func GetByAddr(address std.Address) Valoper",
      "doc": "GetByAddr fetches the valoper using the address, if present\n",
      "params": [
        {
          "Name": "address",
          "Type": "std.Address"
        }
      ],
      "results": [
        {
          "Name": "",
          "Type": "Valoper"
        }
      ]
    }
    // other funcs
    {
      "type": "Valoper",
      "name": "Render",
      "signature": "func (v Valoper) Render() string",
      "doc": "Render renders a single valoper with their information\n",
      "params": [],
      "results": [
        {
          "Name": "",
          "Type": "string"
        }
      ]
    }
    // other methods (in this case of the Valoper type)
  ],
  "types": [
    {
      "name": "Valoper",
      "signature": "type Valoper struct {\n\tName        string // the display name of the valoper\n\tMoniker     string // the moniker of the valoper\n\tDescription string // the description of the valoper\n\n\tAddress      std.Address // The bech32 gno address of the validator\n\tPubKey       string      // the bech32 public key of the validator\n\tP2PAddresses []string    // the publicly reachable P2P addresses of the validator\n\tActive       bool        // flag indicating if the valoper is active\n}",
      "doc": "Valoper represents a validator operator profile\n"
    }
  ]
}
```

## `vm/qeval`

Evaluate exported functions in read-only mode without using gas:

```bash
gnokey query vm/qeval -remote https://rpc.gno.land:443 -data "gno.land/r/gnoland/wugnot.BalanceOf(\"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5\")"
```

Escape quotation marks for string arguments. Only supports primitive types.

## `vm/qrender`

Render a package's output (shorthand for `vm/qeval` on the `Render("")` function):

```bash
gnokey query vm/qrender --data "gno.land/r/gnoland/wugnot:" -remote https://rpc.gno.land:443
```

Displays the current `Render()` output of the realm.

```bash
height: 0
data: # wrapped GNOT ($wugnot)

* **Decimals**: 0
* **Total supply**: 5012404
* **Known accounts**: 2
```

:::info Specifying a path to `Render()`

Use `<pkgpath>:<renderpath>` syntax to call `Render()` with a specific path:

```bash
gnokey query vm/qrender --data "gno.land/r/gnoland/wugnot:balance/g125em6arxsnj49vx35f0n0z34putv5ty3376fg5" -remote https://rpc.gno.land:443
```

:::

## `vm/qpaths`

List package paths with a specified prefix:
```bash
gnokey query vm/qpaths --data "gno.land/r/gnoland"
```

Without a prefix, lists all paths including `stdlibs`. Limit results using 
`<path>?limit=<x>` (default: 1000, max: 10000):
```bash
height: 0
data: gno.land/r/gnoland/blog
gno.land/r/gnoland/coins
gno.land/r/gnoland/events
gno.land/r/gnoland/home
gno.land/r/gnoland/pages
gno.land/r/gnoland/users
gno.land/r/gnoland/users/v1
```

```bash
gnokey query "vm/qpaths?limit=3" --data "gno.land/r/gnoland"
```

Use `@username` to list packages under both `/p` and `/r`:
```bash
gnokey query vm/qpaths --data "@foo"
```

```bash
height: 0
data: gno.land/r/foo
gno.land/r/foo/art/gnoface
gno.land/r/foo/art/millipede
gno.land/p/foo/ui
gno.land/p/foo/svg
```

## `vm/qstorage`

Inspect storage usage and deposit in a realm:

```bash
gnokey query vm/qstorage --data "gno.land/r/foo"
```

Output shows total bytes used (`storage`) and GNOT locked (`deposit`):

```
storage: 5025, deposit: 502500
```

Calculate storage price: `deposit / storage` (e.g., `502500/5025 = 100ugnot`).

### Gas parameters

When using `gnokey` to send transactions, you'll need to specify gas parameters:

```bash
gnokey maketx call \
  --pkgpath "gno.land/r/demo/boards" \
  --func "CreateBoard" \
  --args "MyBoard" "Board description" \
  --gas-fee 1000000ugnot \
  --gas-wanted 2000000 \
  --remote https://rpc.gno.land:443 \
  --chainid staging \
  YOUR_KEY_NAME
```

For detailed information about gas fees, including recommended values and
optimization strategies, see the [Gas Fees documentation](../resources/gas-fees.md).
