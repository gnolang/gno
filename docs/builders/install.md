# Installation

This page covers how to install the Gno toolchain: `gnokey` (key & transaction CLI),
`gno` (language tooling), `gnodev` (local development node with hot reload),
`gnobro` (package browser), and `gnoweb` (realm explorer).

## One-line installer

Install precompiled `gno`, `gnokey`, `gnodev`, `gnobro`, and `gnoweb`
(Linux/macOS, amd64/arm64) from [GitHub Releases](https://github.com/gnolang/gno/releases)
into `$GOPATH/bin`:

```sh
curl -fsSL https://raw.githubusercontent.com/gnolang/gno/master/misc/install.sh | sh -s -- --dir "$(go env GOPATH)/bin"
```


To pin a version:

```sh
curl -fsSL https://raw.githubusercontent.com/gnolang/gno/master/misc/install.sh | sh -s -- --version <tag>
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
| `GNO_INSTALL_DIR` | `--dir` | Installation directory (recommended: `$(go env GOPATH)/bin`) |
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
make install.gno      # Only gno
make install.gnokey   # Only gnokey
make install.gnodev   # Only gnodev
```

Make sure `$GOPATH/bin` is in your `PATH` if `gno`/`gnokey`/`gnodev` are not
found after install — see [Troubleshooting](#troubleshooting).

## Docker

Official Docker images are published under [`ghcr.io/gnolang/gno`](https://ghcr.io/gnolang/gno)
([full list](https://github.com/gnolang/gno/packages)). Run individual tools
without installing from source:

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
# List installed binaries
ls "$(go env GOPATH)/bin"

# Check binary versions
gno version
gnokey version
gnodev version
```

## Troubleshooting

**`command not found`** — `$GOPATH/bin` is not in your `PATH`. Add it:

```sh
export PATH="$PATH:$(go env GOPATH)/bin"
```

Append the `export` line to your shell rc file (`~/.bashrc`, `~/.zshrc`, …) to
persist it. If you installed to a custom `--dir`, add that directory instead.

**Stale binary on `PATH`** — older install shadows the new one. Check with
`command -v gno`; fix by reordering `PATH` or running `hash -r`.

**Go version too old** — `make install` fails on missing language features.
Requires Go **1.24+**: check with `go version`, upgrade from [go.dev/dl](https://go.dev/dl/).

**GitHub API rate limit during one-line install** — anonymous requests are
capped at 60/hour. Set `GITHUB_TOKEN` to authenticate:

```sh
GITHUB_TOKEN=<token> curl -fsSL https://raw.githubusercontent.com/gnolang/gno/master/misc/install.sh | sh
```

## Next steps

- [Anatomy of a Gno package](./anatomy-of-a-gno-package.md) — learn how to write Gno packages
- [Running a local dev node](./local-dev-with-gnodev.md) — spin up a local environment with `gnodev`
- [Deploy packages](./deploy-packages.md) — publish to a network
- [Interacting with gnokey](../users/interact-with-gnokey.md) — manage keys and send transactions
