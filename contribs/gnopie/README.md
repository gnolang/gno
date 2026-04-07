# gnopie

**gnopie** — like [httpie](https://httpie.io), but for gno.land.

An opinionated CLI for interacting with gno.land chains. Uses `gnokey` as a Go library under the hood.

## Install

```bash
cd contribs/gnopie && make install
```

## Usage

```bash
# Default verb is GET — smart dispatch
gnopie gno.land/r/gnoland/blog.Render("")             # EVAL: call function
gnopie gno.land/r/demo/boards                          # INSPECT: list files + functions
gnopie gno.land                                        # INSPECT: network info

# Explicit verbs
gnopie EVAL 'gno.land/r/demo/boards.GetBoardIDFromName("testboard")'
gnopie READ gno.land/r/demo/boards.CreateThread        # source code
gnopie INSPECT gno.land/r/gnoland/blog                 # detailed realm info

# Transactions
gnopie CALL 'gno.land/r/demo/wugnot.Deposit()' --key mykey --send 1000000ugnot
gnopie RUN 'gno.land/r/demo/boards.CreateThread(1,"Hello","World")' --key mykey

# Dry-run and gnokey generation
gnopie CALL --dry-run 'gno.land/r/demo/wugnot.Deposit()' --key mykey
gnopie CALL --generate-gnokey 'gno.land/r/demo/wugnot.Deposit()' --key mykey

# JSON output for piping
gnopie --json gno.land/r/gnoland/blog.Render("") | jq .result

# Manage remotes
gnopie remotes list
gnopie remotes add testnet --rpc https://rpc.test5.gno.land --chain-id test5

# Shell completion
eval "$(gnopie completion bash)"
```

## Verbs

| Verb | Description |
|------|-------------|
| **GET** (default) | Smart dispatch: EVAL for calls, READ for symbols, INSPECT for the rest |
| **EVAL** | Evaluate a read-only function call via qeval |
| **READ** | Read variable value or source code |
| **INSPECT** | Inspect network, namespace, realm, or symbol |
| **CALL** | Sign and broadcast a transaction (requires `--key`) |
| **RUN** | Generate and execute Gno code via maketx run (requires `--key`) |

## Remotes

Network configs are stored in `$GNOHOME/gnopie/remotes.toml`. Default remote is `gno.land`.

The domain in your expression (e.g., `gno.land/r/...`) automatically selects the right remote.

## Roadmap

- Name resolution (`@moul` → address via `r/sys/users`)
- Network discovery from chain metadata
- Transaction history via indexer
- localhost/gnodev integration
- Interactive/REPL mode
