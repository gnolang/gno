---
id: gno-tooling-gnodev
---

# Gnodev

Gnodev allows for quick and efficient development of Gno code.

By watching your development directory, gnodev detects changes in your Gno
code, reflecting them in the state of the node immediately. Gnodev also runs a
local instance of `gnoweb`, allowing you to see the rendering of your Gno code instantly. 

## Features
- **In-Memory Node**: Gnodev starts an in-memory node, and automatically loads
the **examples** folder and any user-specified paths.
- **Web Interface Server**: Starts a `gnoweb` server on `localhost:8888`.
- **Hot Reload**: Monitors the example packages folder and specified directories for file changes,
reloading the package and automatically restarting the node as needed.
- **State Maintenance**: Ensures the current state is preserved by replay all transactions.

## Installation
Gnodev can be found in the `contribs` folder in the monorepo.
To install `gnodev`, run `make install`.

## Usage
Gnodev can be anywhere within the monorepo, and it will automatically load the examples
folder, providing all the packages and realms found in it for use.

![gnodev_usage](../../assets/reference/gnodev/gnodev.gif)

For hot reloading, `gnodev` watches the examples folder, as well as any specified folder:
```
gnodev ./myrealm
```

While `gnodev` is running, the following shortcuts are available:
- To reload manually, press `R`.
- To reset the state of the node, press `CMD+R`
- To see help, press `H`
- To stop `gnodev`, press `CMD+C`

### Options

| Flag       | Effect                                              |
|------------|-----------------------------------------------------|
| --mininal  | Start `gnodev` without loading the examples folder. |
| --no-watch | Disable hot reload                                  |





