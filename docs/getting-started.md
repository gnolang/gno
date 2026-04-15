# Getting started with Gno.land

Gno.land is a Layer 1 blockchain where smart contracts are written in **Gno**,
a deterministic variant of Go. If you can write Go, you can write Gno.

This page takes you from nothing installed to a working toolchain, a key,
test tokens, and your first query against a live network. Plan around
15 minutes.

:::tip Just want to try Gno?
You don't have to install anything to play with Gno code. Open the
**[Gno Playground](https://play.gno.land)**, write Gno in your browser,
and run it instantly. Come back here when you're ready to build locally.
:::

## TL;DR

```sh
git clone https://github.com/gnolang/gno.git && cd gno && make install
gnokey add dev
# → open https://faucet.gno.land and paste your g1… address
gnokey query bank/balances/<your-g1-address> -remote https://rpc.gno.land:443
```

Then read [Writing Gno code](./builders/anatomy-of-a-gno-package.md).

## Prerequisites

- **Git**
- **Go 1.24+** — required by the repository's `go.mod`
- **Make**
- Linux and macOS are the primary supported platforms. Windows users
  should develop inside [WSL2](https://learn.microsoft.com/windows/wsl/install).

## Install from source

```sh
git clone https://github.com/gnolang/gno.git
cd gno
make install
```

The first build takes a few minutes while Go downloads modules. When it
finishes, you should see three confirmation lines:

```
[+] 'gnokey' has been installed. Read more in ./gno.land/
[+] 'gno' has been installed. Read more in ./gnovm/
[+] 'gnodev' has been installed. Read more in ./contribs/gnodev/
```

Three binaries are installed into `$(go env GOPATH)/bin` — make sure
that directory is on your `PATH`:

| Binary   | Purpose                                                  |
|----------|----------------------------------------------------------|
| `gno`    | Gno language CLI — build, test, format, lint `.gno` code |
| `gnokey` | Wallet and transaction CLI — keys, queries, deploys      |
| `gnodev` | Local development node with hot reload                   |

### Verify the installation

```sh
gno version
gnokey version
gnodev --help    # gnodev has no `version` subcommand; --help confirms it runs
```

## Alternative: Docker

If you prefer not to build from source, official images are published to
GitHub Container Registry:

```sh
docker pull ghcr.io/gnolang/gno            # the gno CLI
docker pull ghcr.io/gnolang/gno/gnokey     # wallet / tx CLI
docker pull ghcr.io/gnolang/gno/gnodev     # local dev node
docker pull ghcr.io/gnolang/gno/gnoland    # full node
docker pull ghcr.io/gnolang/gno/gnoweb     # web frontend
```

## Create your first key

`gnokey` stores keys locally. Create a development key named `dev`:

```sh
gnokey add dev
```

You will be asked for a password, then shown a 24-word mnemonic.

:::warning
Treat any key you create this way as **development-only**. Do not reuse
the mnemonic for real funds. Tokens distributed by the faucet are
testnet tokens.
:::

List your keys to see the address that was generated:

```sh
gnokey list
```

The address starts with `g1…` — this is how you'll be identified on-chain.

## Get test tokens

Test transactions require test tokens (GNOT). The community faucet
distributes them:

→ **[faucet.gno.land](https://faucet.gno.land)**

Paste the `g1…` address from `gnokey list`, select a network, and submit.
Tokens arrive in a few seconds.

## Query a live network

Confirm your setup reaches a real network with a free read-only ABCI query:

```sh
gnokey query bank/balances/<your-g1-address> \
  -remote <rpc-endpoint>
```

For Betanet: `-remote https://rpc.gno.land:443` — use the endpoint matching
the network you picked at the faucet. On success, the response shows your
balance in the `data` field (e.g. `"10000000ugnot"`).

The main public networks are:

| Network  | RPC endpoint                               | Chain ID   |
|----------|--------------------------------------------|------------|
| Betanet  | `https://rpc.gno.land:443`                 | `gnoland1` |
| Staging  | `https://rpc.staging.gno.land:443`         | `staging`  |
| Test11   | `https://rpc.test11.testnets.gno.land:443` | `test11`   |

See [Networks](./resources/gnoland-networks.md) for the full reference.

## Before you deploy

### Namespaces

Today, anyone can deploy under their own address-based namespace without
registration:

```
gno.land/r/g1<your-address>/...   # realm (stateful)
gno.land/p/g1<your-address>/...   # package (stateless)
```

Username-based namespaces like `gno.land/r/alice/...` are **not available
yet**. They are planned to be introduced via GovDAO governance. For now,
use the address form. See [Users and Teams](./resources/users-and-teams.md).

### License agreement

Deploying code on gno.land is a public act: your code runs on a shared
network and others may build on top of it. To make the terms clear,
gno.land may require you to accept a **Contributor License Agreement (CLA)**
before you can publish packages.

If your first `gnokey maketx addpkg` is rejected with
`has not signed the required CLA`, sign once by calling the on-chain
`r/sys/cla` realm with the current hash:

```sh
gnokey maketx call \
  -pkgpath gno.land/r/sys/cla -func Sign -args "<current-hash>" \
  -gas-fee 100000ugnot -gas-wanted 2000000 \
  -broadcast -chainid <chain-id> -remote <rpc-endpoint> \
  dev
```

For Betanet: `-chainid gnoland1 -remote https://rpc.gno.land:443`.

Get the current `<hash>` at [`gno.land/r/sys/cla`](https://gno.land/r/sys/cla).
Canonical text: [`CLA.md`](https://github.com/gnolang/gno/blob/master/CLA.md).

## Next steps

- **[Writing Gno code](./builders/anatomy-of-a-gno-package.md)** — the
  language basics, through a Counter realm.
- **[Running a local dev node](./builders/local-dev-with-gnodev.md)** — use
  `gnodev` to iterate with hot reload before touching a shared network.
- **[Deploying Gno packages](./builders/deploy-packages.md)** — publish
  your first realm or package with `gnokey maketx addpkg`.
- **[Example: the `minisocial` dApp](./builders/example-minisocial-dapp.md)** —
  a full end-to-end walkthrough.

**Reading material:**

- **[Go–Gno compatibility](./resources/go-gno-compatibility.md)** — what's
  supported, what isn't, and which Go habits don't translate.
- **[Effective Gno](./resources/effective-gno.md)** — idiomatic patterns
  for writing Gno well.
- **[Discover Gno.land](./users/discover-gnoland.md)** — browse the
  on-chain ecosystem without any tools.

## Getting help

Stuck or want to talk to other builders?

- **[Gno Forum](https://gno.land/r/gnoland/boards2/v1)** — long-form
  questions and proposals, posted on-chain.
- **[GitHub issues](https://github.com/gnolang/gno/issues)** — bug
  reports, feature requests, and public roadmap.
- **[@_gnoland on X](https://twitter.com/_gnoland)** — announcements and
  short-form updates.

## Other components

The root `make install` covers the day-to-day developer toolchain. The
repository ships additional components for specific use cases:

```sh
# Full node and web frontend (operators)
cd gno.land && make install          # installs gnoland, gnoweb, gnokey

# Contrib tools (faucet, kms, migrate, bro, …)
cd contribs && make install_all
```

See [`contribs/`](https://github.com/gnolang/gno/tree/master/contribs) for
the full list.

## Troubleshooting

**`command not found: gno` (or `gnokey`, `gnodev`)**
Your `$(go env GOPATH)/bin` is not on `PATH`. Add it to your shell profile:

```sh
export PATH="$PATH:$(go env GOPATH)/bin"
```

**`unable to determine GNOROOT`**
Gno normally resolves `GNOROOT` automatically. If you see this error — for
example after moving binaries away from the source tree — set it
explicitly:

```sh
export GNOROOT=/path/to/gno
```

**Go version error during `make install`**
The repository requires Go 1.24 or later. Check with `go version` and
upgrade if needed.

**Windows build fails**
Use WSL2 and run the install from inside the Linux environment.
