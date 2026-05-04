# Getting started with Gno.land

Gno.land is a Layer 1 blockchain where smart contracts are written in
**Gno**, a deterministic variant of Go. If you can write Go, you can
write Gno.

This page is the shortest path from zero to a working local chain and
your first on-chain query.

> Try the **[Playground](https://play.gno.land)** to write Gno in your browser

## TL;DR - [Cheatsheet](./cheatsheet.md)

**Local** — install + run a chain on your machine:

```sh
# 1. Install the toolchain (gno, gnokey, gnodev)
curl -fsSL https://raw.githubusercontent.com/gnolang/gno/master/misc/install.sh | sh

# 2. Declare a module path (writes gnomod.toml in current dir)
gno mod init gno.land/r/myname/myrealm

# 3. Add .gno files in the same directory, then run a local chain (hot reload)
gnodev .
# → open http://localhost:8888
```

**Live network** — key, faucet, query:

```sh
# 4. Create a key, then fund it at https://faucet.gno.land
gnokey add dev
gnokey list   # copy the g1... address

# 5. Query your balance
gnokey query bank/balances/g1... -remote https://rpc.gno.land:443
```

- [r/docs](https://gno.land/r/docs/home) — on-chain docs
- [Networks](../resources/gnoland-networks.md) — chain IDs + RPC endpoints
- [Anatomy of a Gno package](./anatomy-of-a-gno-package.md) — realm structure

## What is Gno.land?

Gno.land runs smart contracts written in Gno, an interpreted Go-like
language built for deterministic execution. Realms (`r/`) hold on-chain state, packages
(`p/`) are stateless libraries, and everything is interpreted by the
GnoVM. For the full picture, see
[What is Gno.land?](./what-is-gnolang.md).

## Install - [Other methods](./install.md)

One-liner:

```sh
# Installs gno, gnokey, gnodev to ~/.gno/bin
curl -fsSL https://raw.githubusercontent.com/gnolang/gno/master/misc/install.sh | sh
```

After installing, `gno`, `gnokey`, and `gnodev` should be on your `PATH`.

:::info Troubleshooting
- **`command not found`** — install dir not on `PATH`. Installer uses `$HOME/.gno/bin`; `make install` from source uses `$(go env GOPATH)/bin`. Add to `~/.bashrc` or `~/.zshrc`:
  ```sh
  export PATH="$PATH:$HOME/.gno/bin:$(go env GOPATH)/bin"
  ```
- **Windows** — use [WSL2](https://learn.microsoft.com/windows/wsl/install) and run from inside Linux.
:::

## Build locally with gnodev

Declare a module path, then run a local chain against the current dir:

```sh
# Write gnomod.toml in current dir, declaring the module path
gno mod init gno.land/r/myname/myrealm

# Run a local chain that loads the current dir; hot reloads on edits
gnodev .
# open http://localhost:8888 — Ctrl+C to stop
```

Add your own `.gno` files next to `gnomod.toml`; `gnodev` reloads on
save. Pass directories to load multiple realms; with no arguments it
loads the bundled `examples/`.

See [Running a local dev node](./local-dev-with-gnodev.md) for
genesis, resolvers, and multi-realm setups.

## Deploy to a shared network

Once local iteration works, graduate to a shared network: create a key,
fund it, and query on-chain.

### 1. Create a key

```sh
# Create a new dev key named "dev"; prompts for password, prints mnemonic
gnokey add dev
```

You'll be asked for a password, then shown a 24-word mnemonic.
`gnokey list` prints the address — a `g1…` string used to identify you
on-chain.

:::warning
Keys created this way are **development-only**. Do not reuse the
mnemonic for real funds; faucet tokens are testnet-only.
:::

### 2. Get test tokens

Go to **[faucet.gno.land](https://faucet.gno.land)**, paste your `g1…`
address, pick a network, and submit. Tokens arrive in seconds.

### 3. Query on-chain

```sh
# Query the bank module for the balance of a g1... address
gnokey query bank/balances/g1... -remote https://rpc.gno.land:443
```

The response shows your balance as `<amount>ugnot`. See
[Networks](../resources/gnoland-networks.md) for all chain IDs and RPC
endpoints.

:::info Betanet
Current live network: `-chainid gnoland1 -remote https://rpc.gno.land:443`.
:::

### 4. Before you deploy

**Namespaces.** Anyone can deploy under their address-based namespace
today. Username-based namespaces like `gno.land/r/alice/…` are not
available yet and will require registration.

**CLA.** Publishing code on gno.land may require signing a
[Contributor License Agreement](https://github.com/gnolang/gno/blob/master/CLA.md).
If `gnokey maketx addpkg` fails with `has not signed the required CLA`,
fetch the current hash from [`r/sys/cla`](https://gno.land/r/sys/cla)
and sign once:

```sh
# Sign the CLA on-chain with the "dev" key
gnokey maketx call -pkgpath gno.land/r/sys/cla -func Sign \
  -args "<current-hash>" -gas-fee 100000ugnot -gas-wanted 2000000 \
  -chainid gnoland1 -remote https://rpc.gno.land:443 dev
```

## Next steps

1. [r/docs](https://gno.land/r/docs) — on-chain tour
2. [Anatomy of a Gno package](./anatomy-of-a-gno-package.md) — realm structure via Counter
3. [Running a local dev node](./local-dev-with-gnodev.md) — `gnodev` reference
4. [Example: the `minisocial` dApp](./example-minisocial-dapp.md) — end-to-end with deploy

## Getting help

- **[Gno Forum](https://gno.land/r/gnoland/boards2/v1)** — long-form
  questions and proposals, on-chain.
- **[GitHub issues](https://github.com/gnolang/gno/issues)** — bugs,
  feature requests, roadmap.
- **[@_gnoland on X](https://twitter.com/_gnoland)** — announcements.
