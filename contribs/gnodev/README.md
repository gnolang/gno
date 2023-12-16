## `gnodev`: Your Gno Companion Tool

`gnodev` is designed to be a robust and user-friendly tool in your realm package development journey, streamlining your workflow and enhancing productivity.

### Synopsis
**gnodev** [**-minimal**] [**-no-watch**] [**PKGS_PATH ...**]

### Features
- **In-Memory Node**: Automatically loads the **example** folder and any user-specified paths.
- **Web Interface Server**: Starts a gno.land web server on `:8888`.
- **Hot Reload**: Monitors the example packages folder and additional directories for file changes, reloading the package and restarting the node as needed.
- **State Maintenance**: Ensures the current state is maintained by reapplying all previous blocks.

### Commands
- **H**: Display help information.
- **R**: Reload the node.
- **Ctrl+R**: Reset the current node state.
- **Ctrl+C**: Exit the command.

### Example Folder Loading
The **example** package folder is loaded automatically. If working within this folder, you don't have to specify any additional paths to `gnodev`. Use `--minimal` to prevent this.

### Installation
Run `make install` to install `gnodev`.
