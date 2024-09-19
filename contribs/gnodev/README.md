## `gnodev`: Your Gno Companion Tool

`gnodev` is designed to be a robust and user-friendly tool in your realm package development journey, streamlining your workflow and enhancing productivity.

We will only give a quick overview below. You may find the official documentation at [docs/gno-tooling/gnodev.md](../../docs/gno-tooling/cli/gnodev.md).

### Synopsis
**gnodev** [**-minimal**] [**-no-watch**] [**PKG_PATH ...**]

### Features
- **In-Memory Node**: Gnodev starts an in-memory node, and automatically loads
  the **examples** folder and any user-specified paths.
- **Web Interface Server**: Starts a `gnoweb` server on `localhost:8888`.
- **Hot Reload**: Monitors the example packages folder and specified directories for file changes,
  reloading the package and automatically restarting the node as needed.
- **State Maintenance**: Ensures the current state is preserved by replaying all transactions.

### Commands
While `gnodev` is running, the user can trigger specific actions by pressing
the following combinations:
- **H**: Display help information.
- **R**: Reload the node, without resetting the state.
- **Ctrl+R**: Reset the current node state.
- **Ctrl+C**: Exit `gnodev`.

### Loading 'examples'
The **examples** directory is loaded automatically. If working within this folder, you don't have to specify any additional paths to `gnodev`. Use `--minimal` to prevent this.

### Installation
Run `make install` to install `gnodev`.
