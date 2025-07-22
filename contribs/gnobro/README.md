# gnobro

`gnobro` is a terminal UI to explore realms in your browser. It provides an
alternative to [gnoweb](../../gno.land/cmd/gnoweb) for developers who prefer
working in the terminal.

## Installation

### Minimum Go Version

The minimum Go version required to build `gnobro` is **1.23**.

### Installation instructions

To install `gnobro`, run `make install.gnobro` from the root of the Gno
repository.

## Usage

Start the `gnobro` with the development server:

```sh
gnobro -remote localhost:8888
```

Or connect to a `gnodev` instance by specifying a ws endpoint:

```sh
gnobro -remote localhost:8888 -dev-ws localhost:8889
```

### Command Line Help

[embedmd]:# (.tmp/gnobro-usage.txt)
```txt
USAGE
  gnobro [flags] [pkg_path]

Gnobro is a terminal user interface (TUI) that allows you to browse realms within your
terminal. It automatically connects to Gnodev for real-time development. In
addition to hot reload, it also has the ability to execute commands and interact
with your realm.


FLAGS
  -account ...	default local account to use
  -banner=false 	if enabled, display a banner
  -chainid dev	chainid
  -default-realm gno.land/r/gnoland/home	default realm to display when gnobro starts and no argument is provided
  -dev=true 	enable dev mode and connect to gnodev for realtime update
  -dev-remote ...	dev endpoint, if empty will default to `ws://<target>:8888`
  -jsonlog=false 	display server log as json format
  -readonly=false 	readonly mode, no commands allowed
  -remote 127.0.0.1:26657	remote gno.land URL
  -ssh ...	ssh server listener address
  -ssh-key .ssh/id_ed25519	ssh host key path

```

## Features

- **Terminal UI**: Browse realms directly in your terminal
- **WebSocket Support**: Real-time updates when connected to gnodev
- **Interactive Navigation**: Explore realms with keyboard navigation
- **SSH Access**: Can be configured to run over SSH for remote access

## Development

### Building from source

To build `gnobro` from source:

```sh
cd contribs/gnobro
make build
```

### Running tests

```sh
make test
```

## Architecture

`gnobro` is composed of several packages:

- `pkg/browser`: Core browser UI implementation using BubbleTea
- `pkg/events`: Event types for real-time updates
- `pkg/emitter`: WebSocket server for browser integration

## Related Tools

- [gnodev](../gnodev): The Gno development tool that gnobro can connect to for
  real-time updates
