## `gnodev`: Your Gno Development Companion

`gnodev` is a robust tool designed to streamline your Gno package development process, enhancing productivity
by providing immediate feedback on code changes.

Please note that this is a quick overview. For a more detailed guide, refer to the official documentation at
[docs/builders/local-dev-with-gnodev.md](../../docs/builders/local-dev-with-gnodev.md).

### Synopsis
**gnodev** [**options**] [**PKG_PATH ...**]

### Features
-  **In-Memory Node**: Gnodev starts an in-memory node, automatically loading the **examples** folder and any
   user-specified paths.
-  **Web Interface Server**: Gnodev starts a `gnoweb` server on [`localhost:8888`](https://localhost:8888).
-  **Balances and Keybase Customization**: Set account balances, load them from a file, or add new accounts via a flag.
-  **Hot Reload**: Monitors the **examples** folder and specified directories for file changes, reloading the
   package and automatically restarting the node as needed.
-  **State Maintenance**: Ensures the previous node state is preserved by replaying all transactions.
-  **Transaction Manipulation**: Allows for interactive cancellation and redoing of transactions.
-  **State Export**: Export the current state at any time in a genesis doc format.

### Commands
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

### Usage
Run `gnodev` followed by any specific options and/or package paths. The **examples** directory is loaded
automatically. Use `--minimal` to prevent this.

Example:
```
gnodev --add-account <bech32/name1>[:<amount1>] ./myrealm
```

### `gnobro`: realm interface
`gnobro` is a terminal user interface (TUI) that allows you to browse realms within your terminal. It
automatically connects to `gnodev` for real-time development. In addition to hot reload, it also has the
ability to execute commands and interact with your realm.


#### Usage
**gnobro** [**options**] [**PKG_PATH **]

Run gnobro followed by any specific options and/or a target pacakge path.

Use `gnobro -h` for a detailed list of options.

Example:
```
gnobro gno.land/r/demo/home
```


### Installation
Run `make install` to install `gnodev`.

Run `make install.gnobro` to install `gnobro`.
