# Getting started with Gno.land

Gno.land is a Layer 1 blockchain where smart contracts are written in
**Gno**, a deterministic variant of Go. If you can write Go, you can
write Gno.

Realms (`r/`) hold on-chain state, packages (`p/`) provide stateless
libraries, and the GnoVM interprets everything. See
[What is Gno.land?](./what-is-gnolang.md) for the full picture.

This page is the shortest path from zero to a working local chain and
your first on-chain query.

> Try the **[Playground](https://play.gno.land)** to write Gno in your browser

## TL;DR - [Cheatsheet](./cheatsheet.md)

**Local** — install + run a chain on your machine:

```sh
# 1. Install the toolchain (gno, gnokey, gnodev)
curl -fsSL https://raw.githubusercontent.com/gnolang/gno/master/misc/install.sh | sh

# 2. Declare a module path (writes gnomod.toml in current directory)
gno mod init gno.land/r/myname/myrealm

# 3. Add .gno files in the same directory
# 4. Run a local chain (hot reload)
gnodev .
# → open http://localhost:8888
```

**Live network** — key, faucet, query:

```sh
# 5. Create a key, then fund it at https://faucet.gno.land
gnokey add dev
gnokey list   # copy the g1... address

# 6. Query your balance
gnokey query bank/balances/g1... -remote https://rpc.gno.land:443
```

- [r/docs](https://gno.land/r/docs/home) — on-chain docs
- [Networks](../resources/gnoland-networks.md) — chain IDs + RPC endpoints
- [Anatomy of a Gno package](./anatomy-of-a-gno-package.md) — realm structure

## Contents

- [Install](#install---other-methods) — toolchain
- [Build locally with gnodev](#build-locally-with-gnodev)
  - [Declare the module path](#declare-the-module-path)
  - [Write Gno code](#write-gno-code)
  - [Run a local chain](#run-a-local-chain)
- [Deploy to a shared network](#deploy-to-a-shared-network)
  - [1. Create a key](#1-create-a-key)
  - [2. Get test tokens](#2-get-test-tokens)
  - [3. Query on-chain](#3-query-on-chain)
  - [4. Deploy your package](#4-deploy-your-package)
- [Next steps](#next-steps)
- [Getting help](#getting-help)

## Install - [Other methods](./install.md)

The toolchain has three binaries:

- **`gno`** — language tool: format, test, run, init modules.
- **`gnokey`** — wallet + transaction CLI: keys, queries, deploys.
- **`gnodev`** — local dev chain with hot reload and a web UI.

One-liner installs all three:

```sh
# Installs gno, gnokey, gnodev to ~/.gno/bin
curl -fsSL https://raw.githubusercontent.com/gnolang/gno/master/misc/install.sh | sh
```

After installing, `gno`, `gnokey`, and `gnodev` should be available in
`$GOPATH/bin`. Verify with `gno --help`. If the shell can't find them,
see [Other methods](./install.md) for source builds, Docker, and PATH
fixes.

## Build locally with gnodev

`gnodev` is the local tool to build Gno packages. It runs a private
chain on your machine with test wallets already funded, a web UI, and
hot reload on every save. Iterate here before deploying to a real
chain ([Deploy](#deploy-to-a-shared-network)).

### Declare the module path

```sh
# Write gnomod.toml in current directory, declaring the module path
gno mod init gno.land/r/myname/myrealm
```

`gno mod init` writes a `gnomod.toml` in the current directory. The
path follows the gno.land convention:

- `gno.land/r/...` for **realms** — stateful smart contracts.
- `gno.land/p/...` for **packages** — stateless libraries.

Pick the realm path for your first project. The `myname` segment is
your namespace (use your `g1...` address for testnet deploys, see
[Deploy](#deploy-to-a-shared-network)).

### Write Gno code

Add `.gno` files next to `gnomod.toml`. A minimal realm:

```gno
package myrealm

func Hello() string {
    return "hello, gno.land"
}
```

See [Anatomy of a Gno package](./anatomy-of-a-gno-package.md) for the
canonical structure (state, exported functions, tests, render).

### Run a local chain

```sh
# Run a local chain that loads the current directory; hot reloads on edits
gnodev .
# open http://localhost:8888 — Ctrl+C to stop
```

`gnodev` boots an in-memory Gno blockchain with prefunded test
accounts and serves a web UI at port 8888. Save a `.gno` file and the
chain reloads automatically — no manual deploy needed.

Pass multiple directories to load several packages at once. With no
arguments it loads the bundled `examples/` so you can browse stdlib
realms.

See [Running a local dev node](./local-dev-with-gnodev.md) for
genesis tweaks, resolvers, and multi-realm setups.

## Deploy to a shared network

Publish your package to a live testnet. You'll need a key, some test
`ugnot` for gas, and one `addpkg` transaction.

:::info Betanet
Examples below target the current live network:
`-chainid gnoland1 -remote https://rpc.gno.land:443`. See
[Networks](../resources/gnoland-networks.md) for other chain IDs and
RPC endpoints.
:::

### 1. Create a key

Every transaction (deploy, function call) is signed by a key. Create
one with `gnokey`:

```sh
# Create a new dev key named "dev"; prompts for password, prints mnemonic
gnokey add dev
```

You'll be prompted for an encryption password and shown a 24-word
mnemonic — write it down to recover the key later. Then:

```sh
gnokey list   # prints the g1... address derived from the key
```

The `g1...` address is your on-chain identity: it owns funds, signs
deploys, and forms the base of your address-based namespace.

:::warning
Keys created this way are **development-only**. Do not reuse the
mnemonic for real funds; faucet tokens are testnet-only.
:::

### 2. Get test tokens

Deploys cost gas, paid in `ugnot`. Get test tokens from the faucet:

Go to **[faucet.gno.land](https://faucet.gno.land)**, paste your `g1…`
address, pick a network, and submit. Tokens arrive in seconds. The
faucet is rate-limited per address but you can re-request as needed
during development.

### 3. Query on-chain

Confirm the funds landed before spending them on a deploy:

```sh
# Query the bank module for the balance of a g1... address
gnokey query bank/balances/g1... -remote https://rpc.gno.land:443
```

The response shows your balance as `<amount>ugnot` (1 GNOT =
1,000,000 ugnot). Read-only queries like this don't need a chainid or
a key — they hit the RPC endpoint directly.

### 4. Deploy your package

`addpkg` uploads a directory of `.gno` files (with its `gnomod.toml`)
as a single package on-chain. The package path is permanent: pick it
carefully.

```sh
# Deploy the current directory as a package on the testnet
gnokey maketx addpkg \
  -pkgpath "gno.land/r/<your-g1-addr>/myrealm" \
  -pkgdir . \
  -gas-fee 1000000ugnot -gas-wanted 50000000 \
  -broadcast \
  -chainid gnoland1 -remote https://rpc.gno.land:443 \
  dev
```

Key flags:

- **`-pkgpath`** — on-chain location. Replace `<your-g1-addr>` with
  your `g1…` address. The realm name (`myrealm`) is yours to choose.
- **`-pkgdir`** — local directory to upload. `.` means current dir.
- **`-gas-wanted` / `-gas-fee`** — gas budget and price. Bump these if
  the tx fails with `out of gas`; the unused portion is refunded.
- **`-broadcast`** — actually send the tx (without it the CLI just
  signs and prints).
- The trailing `dev` is the key name from step 1.

On success the response includes the tx hash and the new package is
queryable at `gno.land/r/<your-g1-addr>/myrealm`.

**Namespaces.** Anyone can deploy under their address-based namespace
today. Username-based namespaces like `gno.land/r/alice/…` are not
available yet and will require registration.

**CLA.** Publishing code on gno.land may require signing a
[Contributor License Agreement](https://github.com/gnolang/gno/blob/master/CLA.md).
If `addpkg` fails with `has not signed the required CLA`, fetch the
current hash from [`r/sys/cla`](https://gno.land/r/sys/cla) and sign
once, then retry the deploy:

```sh
# Sign the CLA on-chain with the "dev" key
gnokey maketx call -pkgpath gno.land/r/sys/cla -func Sign \
  -args "<current-hash>" -gas-fee 100000ugnot -gas-wanted 2000000 \
  -chainid gnoland1 -remote https://rpc.gno.land:443 dev
```

## Next steps

1. [r/docs](https://gno.land/r/docs) — on-chain tour
2. [Anatomy of a Gno package](./anatomy-of-a-gno-package.md) — realm structure via Counter
3. [Effective Gno](../resources/effective-gno.md) — idiomatic patterns
4. [Running a local dev node](./local-dev-with-gnodev.md) — `gnodev` reference
5. [Example: the `minisocial` dApp](./example-minisocial-dapp.md) — end-to-end with deploy

## Getting help

- **[Gno Forum](https://gno.land/r/gnoland/boards2/v1)** — long-form
  questions and proposals, on-chain.
- **[GitHub issues](https://github.com/gnolang/gno/issues)** — bugs,
  feature requests, roadmap.
- **[@_gnoland on X](https://twitter.com/_gnoland)** — announcements.
