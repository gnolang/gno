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
- **Balances and Keybase Customization**: Users can set balances, load from a file, or add users via a flag.
- **Hot Reload**: Gnodev monitors the **examples** folder and any specified for
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

Gnodev will, by default, load your keybase located in the GNOHOME directory, giving all your keys (almost) unlimited fund. 

All realms will be by the `-genesis-creator`, but You can also pass query options to the realm path to load a
realm with specific creator and deposit:
```
gnodev ./myrealm?creator=foo&deposit=42ugnot``
```

### Additional Account
Use the `-add-account` flag with the format `<bech32/name>[:<amount>]` to add a specific address or key name
from your local keybase. You can set an optional amount for this address. Repeat this command to add multiple
accounts. The addresses will be shown during runtime or by pressing `A` to display accounts interactively.

## Interactive Usage

While `gnodev` is running, the following shortcuts are available:
- To see help, press `H`.
- To display accounts balances, press `A`.
- To reload manually, press `R`.
- To reset the state of the node, press `CMD+R`.
- To stop `gnodev`, press `CMD+C`.

### Options

| Flag                | Effect                                                  |
|---------------------|---------------------------------------------------------|
| --minimal           | Start `gnodev` without loading the examples folder.     |
| --no-watch          | Disable hot reload.                                     |
| --add-account       | Pre-add account(s) in the form `<bech32>[:<amount>]`    |
| --balances-file     | Load a balance for the user(s) from a balance file.     |
| --chain-id          | Set node ChainID                                        |
| --genesis-creator   | Name or bech32 address of the genesis creator           |
| --home              | Set the path to load user's Keybase.                    |
| --max-gas           | Set the maximum gas per block                           |
| --no-replay         | Do not replay previous transactions upon reload         |
| --node-rpc-listener | listening address for GnoLand RPC node                  |
| --root              | gno root directory                                      |
| --server-mode       | disable interaction, and adjust logging for server use. |
| --verbose           | enable verbose output for development                   |
| --web-listener      | web server listening address                            |
