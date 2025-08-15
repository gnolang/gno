# gnoweb

`gnoweb` is a universal web frontend for the gno.land blockchain.

This README provides instructions on how to set up and run `gnoweb` for development purposes.

## Prerequisites

Before you begin, ensure you have the following software installed on your machine:

- **Node.js**: Required for running JavaScript and CSS build tools.
- **Go**: Required for building `gnoweb`

## Development

In order to run gnoweb in developement mode ensure that a local gnoland node is up and running.
For development purposes, it's recommended to use [gnodev](../../../contribs/gnodev) as the development node.

You can launch a fresh node using:
```sh
make node-start # or use `gnodev -no-web` directly
```

Then use the following command on another terminal to start the development environment, which runs multiple tools in parallel:

```sh
make dev
```

This will:

- Start a `gnoweb` server in development mode and watch for any Go files change (listening on [localhost](http://localhost:8888)).
- Enable Tailwind CSS in watch mode to automatically compile CSS changes.
- Use esbuild in watch mode to automatically transpile and bundle TypeScript changes.

### Custom remote
By default, `make dev` uses a local node. However, you can specify a custom target using the `DEV_REMOTE` environment variable (and optionally set `CHAIN_ID` variable).

For example, to use `gno.land` as the target, run:
```sh
DEV_REMOTE=https://rpc.gno.land make dev
```

### Static Assets in Development

When running in development mode (with `make dev`), static assets are **not embedded** in the binary. Instead,
they are served from a directory specified by the `GNOWEB_ASSETDIR` environment variable or the `AssetDir`
preprocessor variable (set via `-ldflags`).

### Editor Integration

We use Biome for frontend linting and formatting.

You can either install the appropriate Biome extension for your editor by following the official guide. Or simply run `make lint` or `make fmt` (that will automatically run `biome` under the hood).

## Generate

To generate the public assets for the project, including static assets (fonts, CSS and JavaScript... files),
run the following command. This should be used while editing CSS, JS, or any asset files:

```sh
make generate
```
