---
id: gno-tooling-gnodev
---

# gnodev

Gnodev allows for quick and efficient development of Gno code.

By watching your development directory, gnodev detects changes in your Gno
code, reflecting them in the state of the node immediately. Gnodev also runs a
local instance of `gnoweb`, allowing you to see the rendering of your Gno code instantly. 

## Features
- **In-Memory Node**: Gnodev starts an in-memory node, and automatically loads
  the **examples** folder and any user-specified paths.
- **Web Interface Server**: Gnodev automatically starts a `gnoweb` server on
[`localhost:8888`](https://localhost:8888).
- **Balances and Keybase Customization**: Users can set account balances, load them from a file, or add new
  accounts via a flag.
- **Hot Reload**: Gnodev monitors the **examples** folder, as well as any folder specified as an argument for
  file changes, reloading and automatically restarting the node as needed.
- **State Maintenance**: Gnodev replays all transactions in between reloads,
  ensuring the previous node state is preserved.

## Installation

Gnodev can be found in the `contribs` folder in the monorepo.
To install `gnodev`, run `make install`.

## Usage
Gnodev can be run from anywhere on the machine it was installed on, and it will
automatically load the examples folder, providing all the packages and realms found in it for use.

![gnodev_usage](../../assets/gno-tooling/gnodev/gnodev.gif)

For hot reloading, `gnodev` watches the examples folder, as well as any specified folder:
```
gnodev ./myrealm
```

## Keybase and Balance

Gnodev will, by default, load the keybase located in your GNOHOME directory, pre-mining `10e12` amount of
ugnot to all of them. This way, users can interact with Gnodev's in-memory node out of the box. The addresses
and their respective balance can be shown at runtime by pressing `A` to display accounts interactively.

### Adding or Updating Accounts

Utilize the `--add-account` flag to add a new account or update an existing one in your local Keybase,
following the format `<bech32/name>[:<amount>]`. The `<bech32/name>` represents the specific key name or
address, and `<amount>` is an optional limitation on the account.

Example of use:

```
gnodev --add-account <bech32/name1>[:<amount1>] --add-account <bech32/name2>[:<amount2>] ...
```

Please note: If the address exists in your local Keybase, the `--add-account` flag will only update its amount,
instead of creating a duplicate.

### Balance file

You can specify a balance file using `--balance-file`. The file should contain a
list of Bech32 addresses with their respective amounts:

```
# Accounts:
g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5=10000000000000ugnot # test1
g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj=10000000000000ugnot # test2

# ...
```

### Deploy

All realms and packages will be deployed to the in-memory node by the address passed in with the
`--deploy-key` flag. The `deploy-key` address can be changed for a specific package or realm by passing in
the desired address (or a known key name) using with the following pattern:

```
gnodev ./myrealm?creator=g1....
```

A specific deposit amount can also be set with the following pattern:

```
gnodev ./myrealm?deposit=42ugnot
```

This patten can be expanded to accommodate both options:

```
gnodev ./myrealm?creator=<addr>&deposit=<amount>
```

## Interactive Usage

While `gnodev` is running, the following shortcuts are available:
- To see help, press `H`.
- To display accounts balances, press `A`.
- To reload manually, press `R`.
- To reset the state of the node, press `CMD+R`.
- To stop `gnodev`, press `CMD+C`.

### Options

| Flag                | Effect                                                     |
|---------------------|------------------------------------------------------------|
| --minimal           | Start `gnodev` without loading the examples folder.        |
| --no-watch          | Disable hot reload.                                        |
| --add-account       | Pre-add account(s) in the form `<bech32>[=<amount>]`       |
| --balances-file     | Load a balance for the user(s) from a balance file.        |
| --chain-id          | Set node ChainID                                           |
| --deploy-key        | Default key name or Bech32 address for uploading packages. |
| --home              | Set the path to load user's Keybase.                       |
| --max-gas           | Set the maximum gas per block                              |
| --no-replay         | Do not replay previous transactions upon reload            |
| --node-rpc-listener | listening address for GnoLand RPC node                     |
| --root              | gno root directory                                         |
| --server-mode       | disable interaction, and adjust logging for server use.    |
| --verbose           | enable verbose output for development                      |
| --web-listener      | web server listening address                               |
