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

### Options

|Flag|Effect|
|---------|--------|
|`-h`|Display help information|
|`-remote`|gno.land JSON-RPC URL|
|`-dev-ws`|Dev gnodev websocket URL|
|`-browser-ws`|Enable a local websocket server for browser integration|
|`-legacy`|Use legacy TUI behavior|

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
