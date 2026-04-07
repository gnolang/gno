# gnopie

**gnopie** — like [httpie](https://httpie.io), but for gno.land.

An opinionated CLI for interacting with gno.land chains. Uses `gnokey` as a Go library under the hood.

## Install

```bash
cd contribs/gnopie && make install
```

## Quick Start

```bash
# Set your default key once
gnopie config set key=moul
```

## Usage

```bash
# Default verb is GET — smart dispatch
gnopie 'gno.land/r/gnoland/blog.Render("")'           # EVAL: call function (read-only)
gnopie gno.land/r/demo/boards                          # INSPECT: list files + functions
gnopie gno.land                                        # INSPECT: network info

# Explicit verbs
gnopie EVAL 'gno.land/r/demo/boards.GetBoardIDFromName("testboard")'
gnopie READ gno.land/r/demo/boards.CreateThread        # source code
gnopie READ gno.land/r/gnoland/blog/admin.gno          # specific file
gnopie INSPECT gno.land/r/gnoland/blog                 # detailed realm info

# Transactions (uses default key from config)
gnopie CALL 'gno.land/r/demo/wugnot.Deposit()' --send 1000000ugnot
gnopie CALL 'gno.land/r/demo/boards.CreateThread(1,"Hello","World")'
gnopie RUN 'gno.land/r/demo/boards.CreateThread(1,"Hello","World")'

# Override key for a specific call
gnopie CALL 'gno.land/r/demo/wugnot.Deposit()' --key other-account --send 1000000ugnot

# Dry-run and gnokey generation
gnopie CALL --dry-run 'gno.land/r/demo/wugnot.Deposit()' --send 1000000ugnot
gnopie CALL --generate-gnokey 'gno.land/r/demo/wugnot.Deposit()' --send 1000000ugnot

# JSON output for piping
gnopie --json 'gno.land/r/gnoland/blog.Render("")' | jq .result

# Debug mode
gnopie --debug gno.land/r/gnoland/blog

# Shell completion
eval "$(gnopie completion bash)"
```

## Verbs

| Verb | Description |
|------|-------------|
| **GET** (default) | Smart dispatch: EVAL for calls, READ for symbols, INSPECT for the rest |
| **EVAL** | Evaluate a read-only function call via qeval |
| **READ** | Read variable value, source code, or specific file |
| **INSPECT** | Inspect network, namespace, realm, or symbol |
| **CALL** | Sign and broadcast a transaction |
| **RUN** | Generate and execute Gno code via maketx run |

CALL sends a `MsgCall` directly. RUN wraps the expression in a `main()` package and sends a `MsgRun`, which allows importing the realm and calling with full Go-like syntax.

## Network Discovery

Network configuration is auto-discovered via [gnoconnect](https://github.com/gnolang/gno) meta tags. When you use a domain like `gno.land`, gnopie fetches `https://gno.land/` and reads `<meta name="gnoconnect:rpc">` and `<meta name="gnoconnect:chainid">` tags. Results are cached locally for 24h.

## Configuration

```bash
gnopie config set key=moul    # set default signing key
gnopie config get key          # get a config value
gnopie config list             # show all settings
```

## Roadmap

- Name resolution (`@moul` → address via `r/sys/users`)
- Transaction history via indexer
- localhost/gnodev integration
- Interactive/REPL mode
