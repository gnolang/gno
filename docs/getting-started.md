# Getting started with Gno.land

Gno.land is a Layer 1 blockchain where smart contracts are written in
**Gno**, a deterministic variant of Go. If you can write Go, you can
write Gno.

This page is the shortest path from zero to a working local chain and
your first on-chain query. Plan around 10 minutes.

## Try it first (no install)

Before installing anything, you can explore Gno.land in the browser:

- **[Gno Playground](https://play.gno.land)** — write, run, and share
  Gno code instantly, no local setup.
- **[gno.land/r/docs](https://gno.land/r/docs)** — an on-chain guided
  tour of the language and the most common realms.

Come back here when you're ready to build locally.

## TL;DR

```sh
# 1. Install the toolchain (gno, gnokey, gnodev)
curl -fsSL https://raw.githubusercontent.com/gnolang/gno/master/misc/install.sh | sh

# 2. Run a local dev chain with hot reload
gnodev
# → open http://localhost:8888

# 3. Create a development key
gnokey add dev

# 4. Fund it at https://faucet.gno.land, then query your balance
gnokey query bank/balances/g1... -remote https://rpc.gno.land:443
```

Next: explore **[r/docs](https://gno.land/r/docs)**, read
[Anatomy of a Gno package](./builders/anatomy-of-a-gno-package.md), or
keep the [Cheatsheet](./builders/cheatsheet.md) open while you build.

## What is Gno.land?

Gno.land runs smart contracts written in Gno — Go, minus the
non-deterministic parts. Realms (`r/`) hold on-chain state, packages
(`p/`) are stateless libraries, and everything is interpreted by the
GnoVM. For the full picture, see
[What is Gno.land?](./builders/what-is-gnolang.md).

## Install

One-liner:

```sh
curl -fsSL https://raw.githubusercontent.com/gnolang/gno/master/misc/install.sh | sh
```

For prerequisites, building from source, Docker images, and per-binary
details, see **[Install](./builders/install.md)**.

After installing, `gno`, `gnokey`, and `gnodev` should be on your `PATH`.

## Build locally with gnodev

`gnodev` is the recommended way to iterate. It runs a preconfigured
local chain with premined test accounts, a web UI, and hot reload on any
`.gno` file you edit under `examples/`.

```sh
gnodev
# banner + http://localhost:8888
# Ctrl+C to stop
```

See [Running a local dev node](./builders/local-dev-with-gnodev.md) for
genesis config, resolvers, and multi-realm workflows.

## Deploy to a shared network

Once local iteration works, graduate to a shared network: create a key,
fund it, and query on-chain.

### 1. Create a key

```sh
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
gnokey query bank/balances/g1... -remote https://rpc.gno.land:443
```

The response shows your balance as `<amount>ugnot`. See
[Networks](./resources/gnoland-networks.md) for all chain IDs and RPC
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
gnokey maketx call -pkgpath gno.land/r/sys/cla -func Sign \
  -args "<current-hash>" -gas-fee 100000ugnot -gas-wanted 2000000 \
  -chainid gnoland1 -remote https://rpc.gno.land:443 dev
```

## Next steps

1. **[r/docs](https://gno.land/r/docs)** — the on-chain tour, read
   straight from the network.
2. **[Anatomy of a Gno package](./builders/anatomy-of-a-gno-package.md)**
   — how a realm is structured, through a Counter example.
3. **[Running a local dev node](./builders/local-dev-with-gnodev.md)**
   — deeper `gnodev` reference.
4. **[Example: the `minisocial` dApp](./builders/example-minisocial-dapp.md)**
   — a full end-to-end walkthrough, deploy included.
5. **[Cheatsheet](./builders/cheatsheet.md)** — the commands from this
   page, plus everything you'll use next, on one screen.

## Troubleshooting

- **`command not found: gno` (or `gnokey`, `gnodev`)** — the install
  directory isn't on `PATH`. The installer uses `$HOME/.gno/bin`;
  `make install` from source uses `$(go env GOPATH)/bin`. Add whichever
  applies:
  ```sh
  export PATH="$PATH:$HOME/.gno/bin:$(go env GOPATH)/bin"
  ```
- **`unable to determine GNOROOT`** — set it explicitly:
  `export GNOROOT=/path/to/gno`.
- **Go version error when building from source** — see the
  [Install](./builders/install.md) page for the current minimum.
- **Windows** — use [WSL2](https://learn.microsoft.com/windows/wsl/install)
  and run everything from inside Linux.

## Getting help

- **[Gno Forum](https://gno.land/r/gnoland/boards2/v1)** — long-form
  questions and proposals, on-chain.
- **[GitHub issues](https://github.com/gnolang/gno/issues)** — bugs,
  feature requests, roadmap.
- **[@_gnoland on X](https://twitter.com/_gnoland)** — announcements.
