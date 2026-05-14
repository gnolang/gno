# Getting started with Gno.land

Gno.land is a Layer 1 blockchain where smart contracts are written in
**Gno**, a deterministic variant of Go. If you can write Go, you can
write Gno.

Realms (`r/`) hold on-chain state, packages (`p/`) provide stateless
libraries, and the GnoVM interprets everything. See
[What is Gno.land?](./what-is-gnolang.md) for the full picture.

This page walks you from zero to a working local chain and your first
on-chain transaction. Just want the commands? See [Quick Start](./quickstart.md).

:::tip
Try the **[Playground](https://play.gno.land)** to write Gno in your browser.
:::

## Install

The toolchain has three binaries:

| Binary    | What it is                                              |
|-----------|---------------------------------------------------------|
| `gno`     | the Gno language toolchain (format, test, run, mod init)|
| `gnokey`  | key management for interacting with the network         |
| `gnodev`  | local development environment with hot reload + web UI  |

```sh
curl -fsSL https://raw.githubusercontent.com/gnolang/gno/master/misc/install.sh | sh
```

Binaries land in `$HOME/.gno/bin`. The script is bash-only; Windows users should use
WSL or see [Other methods](./install.md) for source builds and Docker.

:::tip
`gno: command not found`? Add `$HOME/.gno/bin` to your `$PATH`:
`export PATH="$HOME/.gno/bin:$PATH"`.
:::

## Run locally with gnodev

This section creates a realm, runs a local chain with `gnodev`, and
opens it in a browser in `gnoweb`.

### Declare the module path

```sh
gno mod init gno.land/r/myname/myrealm
```

This writes a `gnomod.toml` in the current directory, declaring the
package's on-chain path. Use `gno.land/r/…` for realms (stateful) or
`gno.land/p/…` for packages (stateless — note that pure packages cannot
import realms; see [import rules](../resources/gno-packages.md#import-rules)).

### Write Gno code

Add `.gno` files next to the freshly created `gnomod.toml`. We'll build
a counter, a realm that stores a number and exposes a function to
increment it:

```gno
package myrealm

import "strconv"

var count int

func Increment(_ realm, n int) int {
	count += n
	return count
}

func Render(path string) string {
	return "Count: " + strconv.Itoa(count)
}
```

- **`var count int`** — top-level variables are automatically persisted
  to chain state after each transaction.
- **`Increment`** — the `_ realm` parameter makes the function
  ["crossing"](../resources/gno-interrealm.md), which is required for
  any function that modifies realm state. Callers pass `cross` as the
  first argument.
- **`Render`** — gnoweb calls this to display your realm in the
  browser. The signature must be `func Render(path string) string`.

Add a test file alongside it, `myrealm_test.gno`:

```gno
package myrealm

import "testing"

func TestIncrement(t *testing.T) {
	val := Increment(cross, 5)
	if val != 5 {
		t.Fatalf("expected 5, got %d", val)
	}
}
```

For deeper test patterns see the [Testing guide](../resources/gno-testing.md); a
fuller version of this counter lives at
[`examples/gno.land/r/demo/counter`](https://github.com/gnolang/gno/tree/master/examples/gno.land/r/demo/counter).

### Format, lint, and test

```sh
gno fmt ./...     # rewrite .gno files in canonical style
gno lint ./...    # static checks for common mistakes
gno test ./...    # run _test.gno files
```

### Run a local chain

```sh
gnodev .
```

Open http://localhost:8888 — gnoweb shows your realm under its
package path. Click into it to see the `Render` output ("Count: 0"),
browse exported functions and source code, and view prefunded account
balances. Every key in your local `gnokey` keybase is auto-funded at
startup, so no faucet is needed.

Save a `.gno` file and the chain reloads automatically. Pass multiple
directories to load several packages at once; the bundled `examples/`
are loaded by default including
[`r/docs`](http://localhost:8888/r/docs), an on-chain guided tour you
can modify and browse locally.

### Create a key

Every transaction (deploy or function call) is signed by a key.
Create one with `gnokey`:

```sh
gnokey add mykey
```

It prompts for an encryption password and prints a 24-word mnemonic
— store it somewhere safe to recover the key later. List your keys
to see the derived `g1...` address:

```sh
gnokey list
```

```text
0. mykey (local) - addr: g1abc...xyz pub: gpub1pgf..., path: <nil>
```

That `g1...` address is your on-chain identity. It owns funds, signs
transactions, and forms the base of your address-based namespace
when you deploy to a shared network.

:::warning
Keys created this way are **development-only**. Do not reuse the
mnemonic for real funds.
:::

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
  -func "Increment" -args "5" \
  -gas-fee 1000000ugnot -gas-wanted 1000000000 \
  -chainid dev -remote http://localhost:26657 \
  mykey
```

`-gas-wanted` is the maximum units the transaction may consume;
`-gas-fee` is the price per unit (in `ugnot`, the smallest GNOT
denomination). Together they cap what you'll pay — see
[Gas fees](../resources/gas-fees.md) for estimation and tuning.

The signer at the end is the `mykey` you just created. gnodev
auto-funds every key in your local keybase on startup, so it's
ready to spend with no faucet — and you'll reuse the same key for
the staging and testnet sections later.

`gnodev` also prints a `devtest` seed at startup. That's a well-known
test account — the seed is public and the same on every machine, so
it's only useful for shared dev fixtures. Treat it as throwaway and
never reuse the mnemonic elsewhere.

On success you'll see:

```text
(5 int)
OK!
GAS WANTED: 1000000000
GAS USED:   234567
HEIGHT:     42
EVENTS:     []
TX HASH:    gQP9fJYrZMTK3GgRiio3/V35smzg/jJ62q7t4TLpdV4=
```

The leading `(5 int)` is `Increment`'s return value. Reload the realm
page and `Render` flips from "Count: 0" to "Count: 5"; re-run to keep
incrementing.

For more options, see
[Running a local dev node](./local-dev-with-gnodev.md).

## Deploy to a shared network

Publish your package to a live testnet. Two things change compared
to `gnodev`: keys aren't auto-funded — you'll need test `ugnot` from
a faucet — and each deploy is one explicit `addpkg` transaction
instead of hot-reload on file save.

Every command that hits the chain (faucet, query, deploy, call) must
target the same network. Pick one up front and keep `-remote` (and
`-chainid`) consistent throughout:

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
`gno.land/r/alice/…` are not available yet and will require registration.

**CLA.** Publishing code on gno.land may require acknowledging a
[Contributor License Agreement](https://github.com/gnolang/gno/blob/master/CLA.md),
read it first. If `addpkg` fails with `has not signed the required CLA`,
fetch the current hash from [`r/sys/cla`](https://gno.land/r/sys/cla) and
sign once to acknowledge, then retry the deploy:

```sh
gnokey maketx call -pkgpath gno.land/r/sys/cla -func Sign \
  -args "<current-hash>" -gas-fee 100000ugnot -gas-wanted 2000000 \
  -chainid staging -remote https://rpc.staging.gno.land:443 mykey
```

### 4. Deploy your package

`addpkg` uploads a directory of `.gno` files (with its `gnomod.toml`)
as a single package on-chain.

```sh
gnokey maketx addpkg \
  -pkgpath "gno.land/r/<your-g1-addr>/myrealm" \
  -pkgdir . \
  -gas-fee 1000000ugnot -gas-wanted 20000000 \
  -broadcast \
  -chainid staging -remote https://rpc.staging.gno.land:443 \
  mykey
```

The `mykey` at the end is the key you created earlier. On success you'll see:

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

After deploying, call `Increment` to change on-chain state:

```sh
gnokey maketx call \
  -pkgpath "gno.land/r/<your-g1-addr>/myrealm" \
  -func "Increment" -args "5" \
  -gas-fee 1000000ugnot -gas-wanted 2000000 \
  -broadcast \
  -chainid staging -remote https://rpc.staging.gno.land:443 \
  mykey
```

On success the response leads with the return value, then the tx
receipt:

```text
(5 int)
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

This returns the `Render` output ("Count: 5") — a free, read-only
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
