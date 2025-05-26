# `gnodev`: Your Gno Development Companion

`gnodev` is a robust tool designed to streamline your Gno package development process, enhancing productivity
by providing immediate feedback on code changes.

Please note that this is a quick overview. For a more detailed guide, refer to the official documentation at
[docs/builders/local-dev-with-gnodev.md](../../docs/builders/local-dev-with-gnodev.md).

## Synopsis

**gnodev** [**options**] [**PKG_PATH ...**]

## Features
-  **In-Memory Node**: Gnodev starts an in-memory node, automatically loading the **examples** folder and any
   user-specified paths.
-  **Web Interface Server**: Gnodev starts a `gnoweb` server on [`localhost:8888`](https://localhost:8888).
-  **Balances and Keybase Customization**: Set account balances, load them from a file, or add new accounts via a flag.
-  **Hot Reload**: Monitors the **examples** folder and specified directories for file changes, reloading the
   package and automatically restarting the node as needed.
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
Run `gnodev` followed by any specific options and/or package paths. The **examples** directory is loaded
automatically. Use `--minimal` to prevent this.

Example:
```
gnodev --add-account <bech32/name1>[:<amount1>] ./myrealm
```

### `gnodev -h`
[embedmd]:# (.tmp/gnodev-usage.txt)
```txt
USAGE
  gnodev <cmd> [flags] 

The gnodev command starts an in-memory node and a gno.land web interface, primarily for realm package development.

Currently gnodev comes with two mode <local> and <staging>, those command mostly differ by there default values, while gnodev local as default for working locally, satging default are oriented to be use on server.

gnodev uses its own package loader and resolver system to support multiple scenarios and use cases. It currently supports three types of resolvers, each taking a location as an argument.
- root: This resolver takes a <dir> as its location. It attempts to resolve packages based on your file system structure and the package path. For example, if 'root=/user/gnome/myproject' and you try to resolve 'gno.land/r/bar/buzz' as a package, the <root> resolver will attempt to resolve it to /user/gnome/myproject/gno.land/r/bar/buzz.
- local: This resolver also takes a <dir> as its location. It is designed to load a single package, using the module name from 'gno.mod' within this package to resolve the package.
- remote: This resolver takes a <remote> RPC address as its location. It is meant to use a remote node as a resolver, primarily for testing a local package against a remote node.

Resolvers can be chained, and gnodev will attempt to use them in the order they are declared.

For example:
    gnodev -resolver root=/user/gnome/myproject -resolver remote=https://rpc.gno.lands

If no resolvers can resolve a given package path, the loader will return a "package not found" error.

If no command is provided, gnodev will automatically start in <local> mode.

For more information and flags usage description, use 'gnodev local -h'.

SUBCOMMANDS
  local    Start gnodev in local development mode (default)
  staging  Start gnodev in staging mode

```

### `gnodev local -h`
[embedmd]:# (.tmp/gnodev-local-usage.txt)
```txt
USAGE
  gnodev local [flags] [package_dir...]

LOCAL: Local mode configures the node for local development usage.
This mode is optimized for realm development, providing an interactive and flexible environment.
It enables features such as interactive mode, unsafe API access for testing, and lazy loading to improve performance.
The log format is set to console for easier readability, and the web interface is accessible locally, making it ideal for iterative development and testing.

By default, the current directory and the "example" folder from "gnoroot" will be used as the root resolver.


FLAGS
  -C ...                                                change directory context before running gnodev
  -add-account ...                                      add (or set) a premine account in the form `<bech32|name>[=<amount>]`, can be used multiple time
  -balance-file ...                                     load the provided balance file (refer to the documentation for format)
  -chain-domain gno.land                                set node ChainDomain
  -chain-id dev                                         set node ChainID
  -deploy-key g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5  default key name or Bech32 address for deploying packages
  -genesis ...                                          load the given genesis file
  -interactive=false                                    enable gnodev interactive mode
  -lazy-loader=true                                     enable lazy loader
  -log-format console                                   log output format, can be `json` or `console`
  -max-gas 10000000000                                  set the maximum gas per block
  -no-replay=false                                      do not replay previous transactions upon reload
  -no-watch=false                                       do not watch for file changes
  -no-web=false                                         disable gnoweb
  -node-rpc-listener 127.0.0.1:26657                    listening address for GnoLand RPC node
  -paths ...                                            additional paths to preload in the form of "gno.land/r/my/realm", separated by commas; glob is supported
  -resolver ...                                         list of additional resolvers (`root`, `local`, or `remote`) in the form of <resolver>=<location> will be executed in the given order
  -txs-file ...                                         load the provided transactions file (refer to the documentation for format)
  -unsafe-api=true                                      enable /reset and /reload endpoints which are not safe to expose publicly
  -v=false                                              enable verbose output for development
  -web-help-remote ...                                  gnoweb: web server help page's remote addr (default to <node-rpc-listener>)
  -web-home ...                                         gnoweb: set default home page, use `/` or `:none:` to use default web home redirect
  -web-html=false                                       gnoweb: enable unsafe HTML parsing in markdown rendering
  -web-listener 127.0.0.1:8888                          gnoweb: web server listener address
  -web-with-html=false                                  gnoweb: enable HTML parsing in markdown rendering

```

### `gnodev staging -h`
[embedmd]:# (.tmp/gnodev-staging-usage.txt)
```txt
USAGE
  gnodev staging [flags] [package_dir...]

STAGING: Staging mode configures the node for server usage.
This mode is designed for stability and security, suitable for pre-deployment testing.
Interactive mode and unsafe API access are disabled to ensure a secure environment.
The log format is set to JSON, facilitating integration with logging systems.
Since lazy-load is disabled in this mode, the entire example folder from "gnoroot" is loaded by default.

Additionally, you can specify an additional package directory to load.


FLAGS
  -add-account ...                                      add (or set) a premine account in the form `<bech32|name>[=<amount>]`, can be used multiple time
  -balance-file ...                                     load the provided balance file (refer to the documentation for format)
  -chain-domain gno.land                                set node ChainDomain
  -chain-id dev                                         set node ChainID
  -deploy-key g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5  default key name or Bech32 address for deploying packages
  -genesis ...                                          load the given genesis file
  -interactive=false                                    enable gnodev interactive mode
  -lazy-loader=false                                    enable lazy loader
  -log-format json                                      log output format, can be `json` or `console`
  -max-gas 10000000000                                  set the maximum gas per block
  -no-replay=false                                      do not replay previous transactions upon reload
  -no-watch=false                                       do not watch for file changes
  -no-web=false                                         disable gnoweb
  -node-rpc-listener 127.0.0.1:26657                    listening address for GnoLand RPC node
  -paths gno.land/**                                    additional paths to preload in the form of "gno.land/r/my/realm", separated by commas; glob is supported
  -resolver ...                                         list of additional resolvers (`root`, `local`, or `remote`) in the form of <resolver>=<location> will be executed in the given order
  -txs-file ...                                         load the provided transactions file (refer to the documentation for format)
  -unsafe-api=false                                     enable /reset and /reload endpoints which are not safe to expose publicly
  -v=false                                              enable verbose output for development
  -web-help-remote ...                                  gnoweb: web server help page's remote addr (default to <node-rpc-listener>)
  -web-home :none:                                      gnoweb: set default home page, use `/` or `:none:` to use default web home redirect
  -web-html=false                                       gnoweb: enable unsafe HTML parsing in markdown rendering
  -web-listener 127.0.0.1:8888                          gnoweb: web server listener address
  -web-with-html=false                                  gnoweb: enable HTML parsing in markdown rendering

```


## `gnobro`: realm interface
`gnobro` is a terminal user interface (TUI) that allows you to browse realms within your terminal. It
automatically connects to `gnodev` for real-time development. In addition to hot reload, it also has the
ability to execute commands and interact with your realm.


## Usage
**gnobro** [**options**] [**PKG_PATH **]

Run gnobro followed by any specific options and/or a target pacakge path.

Use `gnobro -h` for a detailed list of options.

Example:
```
gnobro gno.land/r/demo/home
```

### `gnobro -h`
[embedmd]:# (.tmp/gnobro-usage.txt)
```txt
USAGE
  gnobro [flags] [pkg_path]

Gnobro is a terminal user interface (TUI) that allows you to browse realms within your
terminal. It automatically connects to Gnodev for real-time development. In
addition to hot reload, it also has the ability to execute commands and interact
with your realm.


FLAGS
  -account ...                            default local account to use
  -banner=false                           if enabled, display a banner
  -chainid dev                            chainid
  -default-realm gno.land/r/gnoland/home  default realm to display when gnobro starts and no argument is provided
  -dev=true                               enable dev mode and connect to gnodev for realtime update
  -dev-remote ...                         dev endpoint, if empty will default to `ws://<target>:8888`
  -jsonlog=false                          display server log as json format
  -readonly=false                         readonly mode, no commands allowed
  -remote 127.0.0.1:26657                 remote gno.land URL
  -ssh ...                                ssh server listener address
  -ssh-key .ssh/id_ed25519                ssh host key path

```

## Installation
Run `make install` to install `gnodev`.

Run `make install.gnobro` to install `gnobro`.

