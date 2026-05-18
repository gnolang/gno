# Getting started with Gno.land

Gno.land is a Layer 1 blockchain where smart contracts are written in
**Gno**, a deterministic variant of Go. If you can write Go, you can
write Gno. See [What is Gno.land?](./what-is-gnolang.md) for the full
picture.

This page walks you from zero to a working local chain and your first
on-chain transaction. Just want the commands? See [Quick Start](./quickstart.md).

:::tip
Try the **[Playground](https://play.gno.land)** to write Gno in your browser.
:::

## Install

This tutorial uses three of the toolchain binaries:

| Binary    | What it is                                              |
|-----------|---------------------------------------------------------|
| `gno`     | the Gno language toolchain (format, test, run, mod init)|
| `gnokey`  | key management for interacting with the network         |
| `gnodev`  | local development environment with hot reload + web UI  |

```sh
curl -fsSL https://raw.githubusercontent.com/gnolang/gno/master/misc/install.sh | sh
```

Binaries land in `$HOME/.gno/bin`. The script is bash-only; Windows users should use
WSL or see [Installation page](./install.md) for source builds, Docker, version pinning (`--version <tag>`), or `--full` (validator node).

Verify the toolchain is on your `$PATH`:

```sh
gno version && gnokey version && gnodev --help
```

:::tip
`gno: command not found`? Add `$HOME/.gno/bin` to your `$PATH`:
`export PATH="$PATH:$HOME/.gno/bin"`.
:::

## Run locally with gnodev

The fastest way to learn Gno is to run a chain on your own machine.
`gnodev` boots a single-node devnet with hot reload, a built-in
gnoweb UI, and a keybase whose accounts are pre-funded — no faucet,
no genesis file, no peers to configure.

You'll write a small realm, run it through the local toolchain
(`gno fmt`/`lint`/`test`), boot the chain, then sign a real
transaction against it. Everything in this section runs offline.

### Declare the module path

```sh
gno mod init gno.land/r/myname/myrealm
```

This writes a `gnomod.toml` in the current directory, declaring the
package's on-chain path. Use `gno.land/r/…` for realms (stateful) or
`gno.land/p/…` for pure packages (stateless — note that pure packages cannot
import realms; see [import rules](./anatomy-of-a-gno-package.md#import-rules)).

### Write Gno code

Then, add `.gno` files next to the freshly created `gnomod.toml`. We'll
build a counter — a realm that stores a number and exposes a function to
increment it:

```gno
package myrealm

import "strconv"

var count int

func Increment(_ realm) int {
	count++
	return count
}

func Render(path string) string {
	return "Count: " + strconv.Itoa(count)
}
```

- **`var count int`** — top-level variables are automatically persisted
  to chain state after each transaction.
- **`Increment`** — the `_ realm` parameter makes the function
  [crossing function](../resources/gno-interrealm.md), which is required for
  any function that modifies realm state. Callers pass `cross` as the
  first argument.
- **`Render`** — gnoweb calls this to display your realm in the
  browser. The signature must be `func Render(path string) string`.

Add a test file alongside it, `myrealm_test.gno`:

```gno
package myrealm

import "testing"

func TestIncrement(t *testing.T) {
	val := Increment(cross)
	if val != 1 {
		t.Fatalf("expected 1, got %d", val)
	}
}
```

For deeper test patterns see the [Testing guide](../resources/gno-testing.md); a
fuller version of this counter lives at
[`examples/gno.land/r/demo/counter`](https://github.com/gnolang/gno/tree/master/examples/gno.land/r/demo/counter).

### Format, lint, and test

The `gno` CLI ships the same toolchain you'd expect from Go: a
formatter that rewrites code in canonical style, a linter that
catches common mistakes (unused imports, dead code, misuse of
`cross`/`realm`), and a test runner that executes `_test.gno`
files. Run them from the package directory before every commit:

```sh
gno fmt ./...     # rewrite .gno files in canonical style
gno lint ./...    # static checks for common mistakes
gno test ./...    # run _test.gno files
```

### Create a key

A **key** is a private/public keypair managed locally by `gnokey`.
The private side signs your transactions; the public side derives the
`g1…` address that identifies you on chain. You'll need one to do
anything that writes state — a deploy, a function call, a transfer.

Create one:

```sh
gnokey add alice
```

It prompts for an encryption password and prints a 24-word mnemonic
— store it somewhere safe to recover the key later. List your keys
to see the derived `g1...` address:

```sh
gnokey list
```

```text
0. alice (local) - addr: g1abc...xyz pub: gpub1pgf..., path: <nil>
```

That `g1...` address is your on-chain identity. It owns funds, signs
transactions, and forms the base of your address-based namespace
when you deploy to a shared network.

:::warning
Keys created this way are **development-only**. Do not reuse the
mnemonic for real funds.
:::

For key import, derivation, and the full keybase reference, see
[Interact with gnokey](../users/interact-with-gnokey.md#managing-key-pairs).

### Run a local chain

Once the code passes tests, boot a devnet from the package directory.
`gnodev` starts a single-node chain, loads the realm at its declared
package path, and serves gnoweb in the foreground:

```sh
gnodev .
```

Open http://localhost:8888 — gnoweb shows your realm under its
package path. Click into it to see the `Render` output ("Count: 0"),
browse exported functions and source code, and view prefunded account
balances. Every key in your local `gnokey` keybase is auto-funded at
startup, so `alice` already has GNOT to spend — no faucet needed.

Save a `.gno` file and the chain reloads automatically. Pass
additional directories on the command line to load several packages
at once.

### Call Increment

Every realm page in gnoweb has three tabs in the top header:
**Content** (the `Render` output you've already seen), **Source**
(the `.gno` files), and **Actions**. The Actions tab introspects the
realm's exported crossing functions and, for each one, gives you both
a form to call it from the browser via a connected wallet (e.g.
[Adena](../users/third-party-wallets.md)) and the equivalent `gnokey`
command pre-filled with the values you've typed, copy-pasteable and
ready to run.

For `Increment`, the command looks like this:

```sh
gnokey maketx call \
  -pkgpath "gno.land/r/myname/myrealm" \
  -func "Increment" \
  -gas-fee 1000000ugnot -gas-wanted 1000000000 \
  -chainid dev -remote http://localhost:26657 \
  alice
```

`-pkgpath` is the realm's on-chain path — the same one you passed to
`gno mod init`. `-gas-wanted` is the maximum units the transaction
may consume; `-gas-fee` is the price per unit (in `ugnot`, the
smallest GNOT denomination). Together they cap what you'll pay — see
[Gas fees](../resources/gas-fees.md) for estimation and tuning.

The signer at the end is the `alice` key you just created. You'll
reuse it in the staging and testnet sections below.

On success you'll see:

```text
(1 int)
OK!
GAS WANTED: 1000000000
GAS USED:   234567
HEIGHT:     42
EVENTS:     []
TX HASH:    gQP9fJYrZMTK3GgRiio3/V35smzg/jJ62q7t4TLpdV4=
```

The leading `(1 int)` is `Increment`'s return value. Reload the realm
page and `Render` flips from "Count: 0" to "Count: 1"; re-run to keep
incrementing.

For more options, see
[Running a local dev node](./local-dev-with-gnodev.md).

## Deploy to a shared network

Publish your package to a live testnet. Two things change compared
to `gnodev`: keys aren't auto-funded — you'll need test `ugnot` from
a faucet — and each deploy is one explicit `addpkg` transaction
instead of hot-reload on file save.

Pick a target network now and use it consistently — the faucet's
network dropdown and every `gnokey` command's `-remote` and
`-chainid` flags must match:

| Network    | `-chainid` | `-remote`                                  |
|------------|------------|--------------------------------------------|
| Local      | `dev`      | `http://localhost:26657`                   |
| Staging    | `staging`  | `https://rpc.staging.gno.land:443`         |
| Testnet    | `testN`    | `https://rpc.<testN>.testnets.gno.land:443` |

Replace `<testN>` with the current testnet chainid (e.g. `test13`) — see
[Networks](../resources/gnoland-networks.md) for the full list, the
active testnet and mainnet status.

Examples below use **staging** because it resets on a short cadence —
fine for a throwaway first deploy. For anything you want to keep around,
use the current **testnet** instead; staging wipes regularly and your
realm will disappear with it.

### 1. Get test tokens

Deploys cost [gas](../resources/gas-fees.md), paid in `ugnot`. Get test tokens from the faucet:

Go to **[faucet.gno.land](https://faucet.gno.land)**, paste your `g1…`
address, pick a network, and submit. Tokens arrive in seconds. The
faucet is rate-limited per address; wait out the cooldown if a
re-request is rejected.

### 2. Query on-chain

Confirm the funds landed before spending them on a deploy:

```sh
gnokey query bank/balances/<your-g1-addr> -remote https://rpc.staging.gno.land:443
```

Response shows your balance as `<amount>ugnot` (1 GNOT = 1,000,000
ugnot). Read-only queries like this don't need a chainid or a key,
they hit the RPC endpoint directly.

### 3. Before you deploy

Two things to know before publishing your first package:

**Namespaces.** Anyone can deploy under their address-based namespace
(`gno.land/r/<your-g1-addr>/…`). Username-based namespaces like
`gno.land/r/alice/…` aren't supported yet — see
[Users and Teams](../resources/users-and-teams.md).

**CLA.** Some networks require contributors to acknowledge a
[Contributor License Agreement](https://github.com/gnolang/gno/blob/master/CLA.md)
before deploying. If `addpkg` fails with `has not signed the required CLA`,
sign once at [`r/sys/cla`](https://gno.land/r/sys/cla) and retry.

### 4. Deploy your package

`addpkg` uploads a directory of `.gno` files (with its `gnomod.toml`)
as a single package on-chain.

```sh
gnokey maketx addpkg \
  -pkgpath "gno.land/r/<your-g1-addr>/myrealm" \
  -pkgdir . \
  -gas-fee 1000000ugnot -gas-wanted 20000000 \
  -chainid staging -remote https://rpc.staging.gno.land:443 \
  alice
```

`-pkgdir` is the local directory to upload (`.` for the current
directory). 
The `alice` key at the end signs the transaction. On success you'll see:

```text
OK!
GAS WANTED: 20000000
GAS USED:   3456789
HEIGHT:     12345
EVENTS:     []
TX HASH:    Ni8Oq5dP0leoT/IRkKUKT18iTv8KLL3bH8OFZiV79kM=
```

The package is now live and browsable at
**`https://staging.gno.land/r/<your-g1-addr>/myrealm`**
(or `https://<testN>.testnets.gno.land/r/...` on the current testnet).

Two optional flags are worth knowing about:
- `-send <amount>ugnot` — transfer GNOT to the realm with the deploy.
- `-max-deposit <amount>ugnot` — cap the [storage deposit](../resources/storage-deposit.md)
  the chain may lock; the transaction fails if the cap is exceeded.

For the full flag list, see
[`addpkg` in Interact with gnokey](../users/interact-with-gnokey.md#addpackage).
You can also deploy via the [Playground](https://play.gno.land) with a browser
wallet like Adena.

### 5. Call your realm

Same shape as the local call earlier, with two changes: the package
path uses your address-based namespace, and `-gas-wanted` is tuned to
a realistic value.

```sh
gnokey maketx call \
  -pkgpath "gno.land/r/<your-g1-addr>/myrealm" \
  -func "Increment" \
  -gas-fee 1000000ugnot -gas-wanted 2000000 \
  -chainid staging -remote https://rpc.staging.gno.land:443 \
  alice
```

On success the response leads with the return value, then the tx
receipt:

```text
(1 int)
OK!
GAS WANTED: 2000000
GAS USED:   234567
HEIGHT:     12346
EVENTS:     []
TX HASH:    gQP9fJYrZMTK3GgRiio3/V35smzg/jJ62q7t4TLpdV4=
```

To read the state without spending gas, query the realm's render:

```sh
gnokey query vm/qrender \
  -pkgpath "gno.land/r/<your-g1-addr>/myrealm" -data "" \
  -remote https://rpc.staging.gno.land:443
```

This returns the `Render` output ("Count: 1") — a free, read-only
view of your realm's state. For the full `maketx call` and `gnokey`
reference, see [Interact with gnokey](../users/interact-with-gnokey.md).

## Next steps

1. [r/docs](https://gno.land/r/docs) — on-chain tour
2. [Effective Gno](../resources/effective-gno.md) — idiomatic patterns
3. [Example: the `minisocial` dApp](./example-minisocial-dapp.md) — end-to-end with deploy
4. [Gas fees](../resources/gas-fees.md) — pricing, estimation, and the "out of gas" fix
5. [Storage deposit](../resources/storage-deposit.md) — how on-chain storage is paid for, and how to cap it with `-max-deposit`

## Getting help

- **[Discord](https://discord.gg/vb4KVPFUKE)** — community chat.
- **[Gno Forum](https://gno.land/r/gnoland/boards2/v1)** — long-form
  questions and proposals, on-chain.
- **[GitHub issues](https://github.com/gnolang/gno/issues)** — bugs,
  feature requests, roadmap.
- **[@_gnoland on X](https://twitter.com/_gnoland)** — announcements.
