```markdown
# Gnoweb

This README provides instructions on how to set up and run Gnoweb for development purposes.

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

- Start a Go server in development mode and watch for any Go files change.
- Enable Tailwind CSS in watch mode to automatically compile CSS changes.
- Use esbuild in watch mode to automatically bundle JavaScript changes.

You can customize the behavior of the Go server using the `DEV_REMOTE` and
`CHAIN_ID` environment variables. For example, to use `portal-loop` as the
target, run:

```sh
CHAIN_ID=portal-loop DEV_REMOTE=https://rpc.gno.land make dev
```

## Generate

To generate the public assets for the project, including CSS and JavaScript
files, run the following command. This should be used while editing CSS, JS, or
any asset files:

```sh
make generate
```

## Fmt

To format all supported files in the project, use the following command:

```sh
make fmt
```

This ensures that your code follows standard formatting conventions.

## Cleanup

To clean up build artifacts, you can use the following commands:

- To remove the `public` and `tmp` directories:

```sh
make clean
```

For a full clean, which also removes `.cache`, use:

```sh
make fclean
```
