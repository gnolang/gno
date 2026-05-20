# Local development with `gnodev`

`gnodev` is a local Gno.land node bundled with [`gnoweb`](../users/explore-with-gnoweb.md),
designed for a fast edit → save → reload loop. Use it instead of deploying
to a network while you're still iterating on code.

For a first-realm walkthrough, see [Getting started](../builders/getting-started.md).

## Quick start

```sh
gnodev .
```

You should see output along these lines:

```
Loader      ┃ I guessing directory path path=gno.land/r/dev/counter dir={your_pwd}
Accounts    ┃ I default address imported name=devtest addr=g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
Node        ┃ I packages paths=[gno.land/r/dev/counter]
GnoWeb      ┃ I gnoweb started lisn=http://127.0.0.1:8888
--- READY   ┃ I for commands and help, press `h` took=1.391020125s
```

Open <http://localhost:8888> to browse your realm via the built-in
[gnoweb](../users/explore-with-gnoweb.md). The `devtest` account is preloaded
with funds, so no faucet is needed. Press `h` at any time for the in-terminal
help menu (see [Interactive controls](#interactive-controls)).

## Features

### Automatic deployment

`gnodev` deploys the working directory's package to the built-in node
automatically — no `addpkg` step. It also deploys every package and
realm in the monorepo's [`examples/`](https://github.com/gnolang/gno/tree/master/examples)
folder so they're importable.

Package path resolution:
- If a `gnomod.toml` is present, the path inside is used.
- Otherwise, `gnodev` reads the first `.gno` file's `package` declaration
  and deploys it under `gno.land/r/dev/<pkgname>`.

See [Configuring Gno projects](./configuring-gno-projects.md) for `gnomod.toml`
details. The default deployer is `devtest`[^1]; override with `-deploy-key`.

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

### Hot reload

`gnodev` watches the working directory and reloads the node on every `.gno`
save, replaying prior transactions to preserve state across reloads.

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
  -func "Increment" -args "42" \
  -gas-fee 1000000ugnot -gas-wanted 20000000 \
  -broadcast \
  -chainid dev -remote 127.0.0.1:26657 \
  devtest
```

Response:

```text
(42 int)
OK!
GAS WANTED: 20000000
GAS USED:   126933
HEIGHT:     203
EVENTS:     []
TX HASH:    k+WuKgPpoAg+EcR2EnzqxeWqUXB4KhOhg3l6zthSy0I=
```

Refresh <http://localhost:8888> to see the updated `Render()` output. The
`devtest` key works out of the box because it's premined (see above); swap it
for any other key in your keybase.

## Interactive controls

While `gnodev` is running, the following keys are bound (case-insensitive):

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
| `-resolver remote=<rpc>` | Resolve missing dependencies from a remote testnet (e.g. `https://rpc.staging.gno.land:443`) |
| `-genesis <file>` | Load a custom genesis file at startup |
| `-no-watch` | Disable file watching |
| `-no-replay` | Skip transaction replay across reloads |

Run `gnodev --help` for the full flag list.

## See also

- [Getting started](../builders/getting-started.md) — first-realm walkthrough
- [Editor setup](../builders/editor-setup.md) — LSP integration with `gnopls`

[^1]: `devtest` corresponds to address `g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5`.
Its mnemonic is **publicly known** — never use it for real funds.
