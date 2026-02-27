# `gnodev`: Your Gno Development Companion

`gnodev` is a robust tool designed to streamline your Gno package development process, enhancing productivity
by providing immediate feedback on code changes.

Please note that this is a quick overview. For a more detailed guide, refer to the official documentation at
[docs/builders/local-dev-with-gnodev.md](../../docs/builders/local-dev-with-gnodev.md).

## Synopsis

**gnodev** [**options**] [**PKG_PATH ...**]

## Features
-  **In-Memory Node**: Gnodev starts an in-memory node with automatic package discovery.
-  **Package Discovery**: Automatically detects packages via `gnomod.toml` files and workspaces via `gnowork.toml`.
-  **Flexible Loading Modes**: Three loading modes (`auto`, `lazy`, `full`) to balance startup time and convenience.
-  **Web Interface Server**: Gnodev starts a `gnoweb` server on [`localhost:8888`](https://localhost:8888).
-  **Balances and Keybase Customization**: Set account balances, load them from a file, or add new accounts via a flag.
-  **Hot Reload**: Monitors package directories for file changes, reloading the package and automatically
   restarting the node as needed.
-  **State Maintenance**: Ensures the previous node state is preserved by replaying all transactions.
-  **Transaction Manipulation**: Allows for interactive cancellation and redoing of transactions.
-  **State Export**: Export the current state at any time in a genesis doc format.

## Commands
While `gnodev` is running, trigger specific actions by pressing the following combinations:
-  **H**: Display help information.
-  **A**: Display account balances.
-  **R**: Reload the node manually.
-  **P**: Cancel the last action.
-  **N**: Redo the last cancelled action.
-  **Ctrl+S**: Save the current state.
-  **Ctrl+R**: Restore the saved state.
-  **E**: Export the current state to a genesis file.
-  **Cmd+R**: Reset the current node state.
-  **Cmd+C**: Exit `gnodev`.

## Usage
Run `gnodev` from a directory containing a `gnomod.toml` file, and the package will be automatically detected
and loaded. You can also pass package directories as arguments.

### Load Modes
Use the `-load` flag to control how packages are loaded:
- **auto** (default for local): Pre-loads current workspace/package only
- **lazy**: Loads all packages on-demand as they are accessed
- **full** (default for staging): Pre-loads all discovered packages

Example:
```
gnodev                                 # Auto-detect and pre-load current package
gnodev -load=lazy                      # Load packages on-demand only
gnodev -load=full                      # Pre-load all packages
```

### `gnodev -h`
```txt
USAGE
  gnodev <cmd> [flags]

The gnodev command starts an in-memory node and a gno.land
web interface, primarily for realm package development.

gnodev comes with two modes: <local> and <staging>. These commands
differ mainly by their default values - local mode is optimized for
development, while staging mode is oriented for server usage.

Package discovery is automatic via gnomod.toml and gnowork.toml files.
Use the -load flag to control loading behavior:
  - auto:  Pre-load current workspace/package (default for local)
  - lazy:  Load packages on-demand as accessed
  - full:  Pre-load all discovered packages (default for staging)

If no command is provided, gnodev will automatically start in <local> mode.

SUBCOMMANDS
  local    Start gnodev in local development mode (default)
  staging  Start gnodev in staging mode
```

### `gnodev local -h`
```txt
USAGE
  gnodev local [flags] [package_dir...]

LOCAL: Local mode configures the node for local development usage.
This mode is optimized for realm development, providing an interactive and flexible environment.
It enables features such as interactive mode, unsafe API access for testing, and auto loading mode.
The log format is set to console for easier readability, and the web interface is accessible locally.

Package discovery is automatic via gnomod.toml and gnowork.toml files.

FLAGS
  -C ...                      change directory context before running gnodev
  -add-account ...            add (or set) a premine account in the form `<bech32|name>[=<amount>]`
  -balance-file ...           load the provided balance file
  -chain-domain gno.land      set node ChainDomain
  -chain-id dev               set node ChainID
  -deploy-key ...             default key name or Bech32 address for deploying packages
  -empty-blocks=false         enable creation of empty blocks
  -empty-blocks-interval 1    set the interval for creating empty blocks (in seconds)
  -genesis ...                load the given genesis file
  -interactive=false          enable gnodev interactive mode
  -load auto                  package loading mode: auto, lazy, or full
  -log-format console         log output format: json or console
  -max-gas 10000000000        set the maximum gas per block
  -no-replay=false            do not replay previous transactions upon reload
  -no-watch=false             do not watch for file changes
  -no-web=false               disable gnoweb
  -node-rpc-listener ...      listening address for GnoLand RPC node
  -paths ...                  additional paths to preload (glob supported)
  -txs-file ...               load the provided transactions file
  -unsafe-api=true            enable /reset and /reload endpoints
  -v=false                    enable verbose output
  -web-home ...               set default home page
  -web-listener 127.0.0.1:8888  gnoweb: web server listener address
```

### `gnodev staging -h`
```txt
USAGE
  gnodev staging [flags] [package_dir...]

STAGING: Staging mode configures the node for server usage.
This mode is designed for stability and security, suitable for pre-deployment testing.
Interactive mode and unsafe API access are disabled to ensure a secure environment.
The log format is set to JSON, facilitating integration with logging systems.
Full loading mode is used by default, pre-loading all discovered packages.

FLAGS
  -add-account ...            add (or set) a premine account in the form `<bech32|name>[=<amount>]`
  -balance-file ...           load the provided balance file
  -chain-domain gno.land      set node ChainDomain
  -chain-id dev               set node ChainID
  -deploy-key ...             default key name or Bech32 address for deploying packages
  -empty-blocks=false         enable creation of empty blocks
  -empty-blocks-interval 1    set the interval for creating empty blocks (in seconds)
  -genesis ...                load the given genesis file
  -interactive=false          enable gnodev interactive mode
  -load full                  package loading mode: auto, lazy, or full
  -log-format json            log output format: json or console
  -max-gas 10000000000        set the maximum gas per block
  -no-replay=false            do not replay previous transactions upon reload
  -no-watch=false             do not watch for file changes
  -no-web=false               disable gnoweb
  -node-rpc-listener ...      listening address for GnoLand RPC node
  -paths ...                  additional paths to preload (glob supported)
  -txs-file ...               load the provided transactions file
  -unsafe-api=false           enable /reset and /reload endpoints
  -v=false                    enable verbose output
  -web-home :none:            set default home page
  -web-listener 127.0.0.1:8888  gnoweb: web server listener address
```

### Transaction file format

`gnodev` can sends genesis transactions to the local node using the `-txs-file` flag associated with the `-paths` and the `-add-account` flag.

The transactions file attached to the `-txs-file` is a set of transaction, each wrapped by `"tx"`. A transaction can be generated by following the step 1, 2 and 3 from [Making an airgapped transaction](https://docs.gno.land/users/interact-with-gnokey/#making-an-airgapped-transaction).
Here's an example of the format:
```
{"tx": {"msg":[{"@type":"/vm.m_call","caller":"g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj","send":"1000000ugnot","pkg_path":"gno.land/r/gnoland/users/v1","func":"Register","args":["administrator123"]}],"fee":{"gas_wanted":"2000000","gas_fee":"1000000ugnot"},"signatures":[{"pub_key":{"@type":"/tm.PubKeySecp256k1","value":"AmG6kzznyo1uNqWPAYU6wDpsmzQKDaEOrVRaZ08vOyX0"},"signature":""}],"memo":""}}
{"tx": {"msg":[{"@type":"/vm.m_call","caller":"g1qpymzwx4l4cy6cerdyajp9ksvjsf20rk5y9rtt","send":"1000000ugnot","pkg_path":"gno.land/r/gnoland/users/v1","func":"Register","args":["zoo_ma123"]}],"fee":{"gas_wanted":"2000000","gas_fee":"1000000ugnot"},"signatures":[{"pub_key":{"@type":"/tm.PubKeySecp256k1","value":"A6yg5/iiktruezVw5vZJwLlGwyrvw8RlqOToTRMWXkE2"},"signature":""}],"memo":""}}
{"tx": {"msg":[{"@type":"/vm.m_call","caller":"g1manfred47kzduec920z88wfr64ylksmdcedlf5","send":"1000000ugnot","pkg_path":"gno.land/r/gnoland/users/v1","func":"Register","args":["moul001"]}],"fee":{"gas_wanted":"2000000","gas_fee":"200000000ugnot"},"signatures":[{"pub_key":{"@type":"/tm.PubKeySecp256k1","value":"AnK+a6mcFDjY6b/v6p7r8QFW1M1PgIoQxBgrwOoyY7v3"},"signature":""}],"memo":""}}
```


## Related Tools

### `gnobro`: Terminal UI for Realm Browsing
`gnobro` is a terminal user interface (TUI) that allows you to browse realms within your terminal. It can automatically connect to `gnodev` for real-time development with hot reload capabilities and the ability to execute commands and interact with your realm.

`gnobro` is available as a separate tool in the `contribs/gnobro` directory. For more information, see the [gnobro README](../gnobro/README.md).

## Installation
Run `make install` to install `gnodev`.
