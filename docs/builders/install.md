# Installation

This page covers how to install the Gno toolchain: `gnokey` (key & transaction CLI),
`gno` (language tooling), `gnodev` (local development node with hot reload),
`gnobro` (package browser), and `gnoweb` (realm explorer).

## One-line installer

Download precompiled binaries (Linux/macOS, amd64/arm64) with a single command:

```sh
curl -fsSL https://raw.githubusercontent.com/gnolang/gno/master/misc/install.sh | sh
```

This installs `gno`, `gnokey`, `gnodev`, `gnobro`, and `gnoweb` to
`$HOME/.gno/bin` by default. You can override the version and directory:

```sh
curl -fsSL https://raw.githubusercontent.com/gnolang/gno/master/misc/install.sh | sh -s -- --version <tag> --dir /usr/local/bin
```

To also install the validator node (`gnoland`), pass `--full`:

```sh
curl -fsSL https://raw.githubusercontent.com/gnolang/gno/master/misc/install.sh | sh -s -- --full
```

To uninstall:

```sh
curl -fsSL https://raw.githubusercontent.com/gnolang/gno/master/misc/uninstall.sh | sh
```

Scripts used by the one-line installer:

- [misc/install.sh](https://github.com/gnolang/gno/blob/master/misc/install.sh)
- [misc/uninstall.sh](https://github.com/gnolang/gno/blob/master/misc/uninstall.sh)

### Environment variables

| Variable | Equivalent flag | Description |
|---|---|---|
| `GNO_VERSION` | `--version` | Release tag to install (default: latest) |
| `GNO_INSTALL_DIR` | `--dir` | Installation directory (default: `$HOME/.gno/bin`) |
| `GITHUB_TOKEN` | — | Authenticates GitHub API requests; raises the 60 requests/hour anonymous rate limit |

## Install from source

Building from source requires:

- **Go** — version **1.24+** (see [`go.mod`](https://github.com/gnolang/gno/blob/master/go.mod)). Install from [go.dev/dl](https://go.dev/dl/).
- **Git**
- **Make**

```sh
# Clone the repository
git clone https://github.com/gnolang/gno.git
cd gno

# Install all tools (gnokey, gno, gnodev)
make install
```

You can also install individual tools:

```sh
make install.gnokey   # Only gnokey
make install.gno      # Only gno
```

:::tip
Make sure `$GOPATH/bin` is in your `PATH`.
You can check with `go env GOPATH` and add it: `export PATH=$PATH:$(go env GOPATH)/bin`.
:::

## Docker

Official Docker images are available at `ghcr.io/gnolang/gno`. You can use them
to run individual tools without installing from source:

```sh
# Run gnokey
docker run -it ghcr.io/gnolang/gno/gnokey --help

# Run gnoland node
docker run -it ghcr.io/gnolang/gno/gnoland start
```

You can also build locally from the repository root:

```sh
docker build -t gno .
```

## Verify installation

After installing, verify that the tools are available:

```sh
gno version
gnokey version
gnodev version # TODO: https://github.com/gnolang/gno/issues/5550
```

## Next steps

- [Running a local dev node](./local-dev-with-gnodev.md) — spin up a local environment with `gnodev`
- [Interacting with gnokey](../users/interact-with-gnokey.md) — manage keys and send transactions
- [Anatomy of a Gno package](./anatomy-of-a-gno-package.md) — learn how to write Gno packages
