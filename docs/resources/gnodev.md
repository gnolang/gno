# Local development with `gnodev`

`gnodev` is a local Gno.land node bundled with [`gnoweb`](../users/explore-with-gnoweb.md),
designed for a fast edit → save → reload loop. Use it instead of deploying
to a network while you're still iterating on code.

For a first-realm walkthrough, see [Getting started](../builders/getting-started.md).

`gnodev` ships with the Gno toolchain. If `gnodev --help` fails, install it from
the [installation guide](../builders/install.md).

## Quick start

Run `gnodev` from a package directory that has a `gnomod.toml`. The
`gnomod.toml` declares the path the package deploys under:

```toml
module = "gno.land/r/dev/counter"
gno = "0.9"
```

Then start the node:

```sh
gnodev .
```

You should see output along these lines:

```
Loader      ┃ I workspace detected root={your_pwd}
Accounts    ┃ I default address imported name=devtest addr=g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
Proxy       ┃ I lazy loading is enabled. packages will be loaded only upon a request via a query or transaction. loader=native
Node        ┃ I packages paths=[gno.land/r/dev/counter]
GnoWeb      ┃ I gnoweb started lisn=http://127.0.0.1:8888
--- READY   ┃ I for commands and help, press `h` took=1.4s
```

Open `http://localhost:8888` to browse your realm via the built-in
[gnoweb](../users/explore-with-gnoweb.md) (change the address with `-web-listener`).
The `devtest` account is preloaded with funds, so no faucet is needed. Press `h`
at any time for the in-terminal help menu (see [Interactive controls](#interactive-controls)).

## Modes

`gnodev` runs in one of two modes, picked by subcommand:

- `gnodev local` is the default when you run `gnodev` with no subcommand. It is
  tuned for iterating locally: interactive controls, the unsafe `/reset` and
  `/reload` endpoints, and console logs. The workspace and every `-extra-root`
  load up front; `examples/` packages resolve on demand the first time they're
  referenced.
- `gnodev staging` is tuned for server use: no interactive mode, no unsafe API,
  and JSON logs. On top of the workspace and `-extra-root`, all of `examples/`
  is eager-loaded at startup (use `-no-examples` to skip it).

## Features

### Automatic deployment

`gnodev` deploys the working directory's package to the built-in node
automatically, with no `addpkg` step. Pass several directories
(`gnodev ./realmA ./realmB`) to deploy and hot-reload them together.
Packages and realms in the monorepo's
[`examples/`](https://github.com/gnolang/gno/tree/master/examples) folder are
resolved on demand, the first time a query or transaction references them, so
they're importable without deploying them yourself.

To make packages outside the working directory and `examples/` resolvable, add
their root with `-extra-root <dir>` (repeatable). Every package under an extra
root is loaded, so they're available without deploying them by hand.

Package path resolution:
- If a `gnomod.toml` is present, the path inside is used.
- Otherwise, `gnodev` derives the path from the directory name and deploys
  under `gno.land/r/dev/<dirname>`, where `<dirname>` is sanitized to a valid
  path segment: lowercased, non-alphanumerics collapsed to underscores. A name
  with no letters falls back to `app`, and a leading digit is prefixed with `d`
  (the path must start with a letter).

See [Configuring Gno projects](./configuring-gno-projects.md) for `gnomod.toml`
details. The default deployer is `devtest`[^1]; override with `-deploy-key`.

The `devtest` account is automatically imported and pre-funded with locally usable GNOT.
It is available for immediate use in development and testing.

### Premining

Every key in your local `gnokey` keybase is prefunded on the built-in node.
Press `A` while `gnodev` is running to list balances:

```
Accounts    ┃ I (2) known keys
            ┃   │ KeyName  Address                                   Balance
            ┃   │ devtest  g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5  10000000000000ugnot
            ┃   │ MyKey    g1q4q3uegdnq9rsvf3xgxydr3yqd2v6w2tww5920  10000000000000ugnot
```

This is local-only; on remote networks you still need the [faucet](https://faucet.gno.land).

To fund an address you don't hold a key for, or to set a specific balance, use
`-add-account <name|addr>[=<amount>]` (repeatable), or seed many at once from a
file with `-balance-file <file>`. Replay a set of genesis transactions at startup
with `-txs-file <file>`; its signers are auto-premined and the packages it
references are auto-loaded. The `-balance-file` and `-txs-file` flags cannot be
combined with `-genesis`.

The `-balance-file` format is one `<address>=<amount>ugnot` entry per line, with
`#` comments allowed:

```text
g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5=10000000000000ugnot # test1
g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj=10000000000000ugnot # test2
```

The `-txs-file` format is one JSON transaction per line, each wrapped in a `tx`
field. Build the inner transaction with the first steps of [making an airgapped
transaction](../users/interact-with-gnokey.md#making-an-airgapped-transaction):

```json
{"tx": {"msg":[{"@type":"/vm.m_call","caller":"g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj","send":"","pkg_path":"gno.land/r/demo/counter","func":"Increment","args":[]}],"fee":{"gas_wanted":"2000000","gas_fee":"1000000ugnot"},"signatures":[{"pub_key":{"@type":"/tm.PubKeySecp256k1","value":"AmG6kzznyo1uNqWPAYU6wDpsmzQKDaEOrVRaZ08vOyX0"},"signature":""}],"memo":""}}
```

### Hot reload

`gnodev` watches the working directory and reloads the node on every `.gno`
save, replaying prior transactions to preserve state across reloads.

### Gas profiling

The built-in node exposes the `.app/profiletx` gas-profiler query, so you can
profile any transaction's gas usage against your dev node and view it with
`go tool pprof`. See
[Profiling a transaction](./gno-testing.md#profiling-a-transaction). (This query
is a local-development feature and is off by default on real nodes.)

### Genesis and node tuning

`gnodev` can also load a genesis file with `-genesis <file>` to seed the
built-in node with a custom state, and exposes flags to tune node behavior
(RPC listener, web listener, chain ID, …). See `gnodev --help` for the full
list.

## Calling your realm

Once `gnodev` is running, you can drive your realm from another terminal using
`gnokey` against the local node. The defaults are `-remote 127.0.0.1:26657` and
`-chainid dev`:

```sh
gnokey maketx call \
  -pkgpath "gno.land/r/dev/counter" \
  -func "Increment" \
  -gas-fee 1000000ugnot -gas-wanted 20000000 \
  -broadcast \
  -chainid dev -remote 127.0.0.1:26657 \
  devtest
```

Response (`Increment` takes no argument and returns the new count):

```text
(1 int)
OK!
GAS WANTED: 20000000
GAS USED:   126933
HEIGHT:     203
EVENTS:     []
INFO:
TX HASH:    k+WuKgPpoAg+EcR2EnzqxeWqUXB4KhOhg3l6zthSy0I=
```

Refresh `http://localhost:8888` to see the updated `Render()` output. The
`devtest` key works out of the box because it's premined (see above); swap it
for any other key in your keybase.

If you start `gnodev` on a non-default RPC port, point `-remote` at the same
address. For example, started with:

```sh
gnodev -node-rpc-listener 127.0.0.1:36657
```

the call above must use `-remote 127.0.0.1:36657`.

## Interactive controls

These keys work while `gnodev` is running in interactive mode (case-insensitive).
Interactive mode is on by default in `gnodev local` when standard output is a
terminal; it turns off when output is piped or redirected, and in 
`gnodev staging`. Pass `-interactive` to force it:

| Key | Action |
|---|---|
| `H` | Show the in-terminal help menu |
| `A` | List known accounts and balances |
| `R` | Reload all packages |
| `N` / `P` | Step to the next / previous transaction |
| `E` | Export the current state as a genesis doc |
| `Ctrl+S` | Save the current state |
| `Ctrl+R` | Reset to the initial or last-saved state |
| `Ctrl+C` | Exit |

## Useful flags

| Flag | Purpose |
|---|---|
| `-deploy-key <name>` | Override the default deployer (`devtest`) |
| `-extra-root <dir>` | Add another directory tree whose packages are resolvable, alongside the working directory and `examples/` (repeatable) |
| `-remote <domain>=<rpc>` | Fetch missing packages for a chain domain from its RPC, as `<domain>=<rpc>` (e.g. `gno.land=https://rpc.staging.gno.land:443`). Only domains given an entry are fetched; with no `-remote`, gnodev never reaches the network for packages |
| `-paths <paths>` | Preload extra package paths, comma-separated (e.g. `gno.land/r/my/realm`) |
| `-no-examples` | Skip loading `$GNOROOT/examples` entirely |
| `-add-account <name\|addr>[=<amount>]` | Premine or set the balance of an account (repeatable) |
| `-balance-file <file>` | Seed account balances from a file (cannot be combined with `-genesis`) |
| `-txs-file <file>` | Replay genesis transactions at startup; signers are auto-premined and referenced packages auto-loaded (cannot be combined with `-genesis`) |
| `-genesis <file>` | Load a custom genesis file at startup |
| `-node-rpc-listener <addr>` | Node RPC listen address (default `127.0.0.1:26657`); `gnokey -remote` must match it |
| `-web-listener <addr>` | gnoweb listen address (default `127.0.0.1:8888`) |
| `-no-web` | Run without gnoweb |
| `-unsafe-api` | Expose the `/reset` and `/reload` HTTP endpoints (on in local, off in staging) |
| `-interactive` | Force interactive controls when stdout is not a terminal (on by default in local at a terminal, off in staging) |
| `-no-watch` | Disable file watching |
| `-no-replay` | Skip transaction replay across reloads |

Run `gnodev --help` for the full flag list.

## See also

- [Getting started](../builders/getting-started.md): first-realm walkthrough
- [Editor setup](../builders/editor-setup.md): LSP integration with `gnopls`

[^1]: `devtest` corresponds to address `g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5`.
Its mnemonic is **publicly known**. Never use it on production networks or for real funds.
