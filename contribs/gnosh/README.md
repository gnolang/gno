# gnosh

**gnosh** (gno shell) is an opinionated CLI for interacting with gno.land chains.

It wraps `gnokey` under the hood (as a Go library, not a shell command) and provides:

- **Better UX** — sensible defaults, auto-gas estimation, human-friendly output
- **Chainable** — `--json` flag on all commands for piping to `jq` and scripts
- **`--dry-run`** — simulate transactions without broadcasting
- **`--generate-gnokey`** — print the equivalent `gnokey` command for any operation
- **Smart defaults** — gno.land mainnet by default, auto account/sequence lookup

## Install

```bash
cd contribs/gnosh && make install
```

## Usage

```bash
# Query a realm
gnosh query gno.land/r/demo/boards 'GetBoardIDFromName("testboard")'

# Render a realm page
gnosh query --render gno.land/r/demo/boards "testboard"

# Call a realm function (auto-gas)
gnosh call --key mykey gno.land/r/demo/wugnot Deposit --send 1000000ugnot

# Dry-run a call
gnosh call --dry-run --key mykey gno.land/r/demo/boards CreateThread 1 "Hello" "World"

# Generate equivalent gnokey command
gnosh call --generate-gnokey gno.land/r/demo/wugnot Deposit --send 1000000ugnot
```

## Roadmap

- `gnosh send` — bank send
- `gnosh run` — execute Gno code
- `gnosh inspect` — inspect realm/package state
- `gnosh history` — transaction history via indexer
- `gnosh addpkg` — deploy packages
- Name resolution (`@moul` → address via `r/sys/users`)
- Network discovery and domain management
- localhost/gnodev integration
- Interactive/REPL mode
