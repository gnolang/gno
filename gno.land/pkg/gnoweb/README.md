# gnoweb

`gnoweb` is a universal web frontend for the gno.land blockchain.

This README provides instructions on how to set up and run `gnoweb` for development purposes.

## Prerequisites

Before you begin, ensure you have the following software installed on your machine:

- **Node.js**: Required for running JavaScript and CSS build tools.
- **Go**: Required for building `gnoweb`

## Development

To start the development environment, which runs multiple development tools in parallel,
use the following command:

```sh
make dev
```

This will:

- Start a Go server in development mode and watch for any Go files change (targeting [localhost](http://localhost:8888)).
- Enable Tailwind CSS in watch mode to automatically compile CSS changes.
- Use esbuild in watch mode to automatically transpile and bundle TypeScript changes.

You can customize the behavior of the Go server using the `DEV_REMOTE` and
`CHAIN_ID` environment variables. For example, to use `staging` as the
target, run:

```sh
CHAIN_ID=staging DEV_REMOTE=https://rpc.gno.land make dev
```

## Generate

To generate the public assets for the project, including static assets (fonts, CSS and JavaScript...
files), run the following command. This should be used while editing CSS, JS, or
any asset files:

```sh
make generate
```
