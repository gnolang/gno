# gnopie

**gnopie** — like [httpie](https://httpie.io), but for gno.land.

## Install

```bash
cd contribs/gnopie && make install
gnopie config set key=moul
```

## Examples

```bash
gnopie gno.land/r/gnoland/blog                          # Browse a realm (Render)
gnopie 'gno.land/r/gnoland/blog.Render("")'             # Evaluate a function
gnopie gno.land                                          # Network info
gnopie g1manfred47kzduec920z88wfr64ylksmdcedlf5         # Inspect address
gnopie https://gno.land/r/gnoland/blog:p/beta-mainnet   # Paste gnoweb URL
gnopie INSPECT gno.land/r/gov/dao                        # List files + functions
gnopie READ gno.land/r/gnoland/blog.ModAddPost           # Function source code
gnopie READ gno.land/r/gnoland/blog/admin.gno            # Read a specific file
gnopie CALL 'gno.land/r/demo/wugnot.Deposit()'          # Sign & broadcast tx
gnopie RUN 'gno.land/r/demo/counter.Increment()'        # Execute via maketx run
gnopie CALL --print-gnokey-command 'gno.land/r/demo/wugnot.Deposit()'
gnopie --json gno.land/r/gnoland/blog | jq .result       # JSON for piping
gnopie --debug gno.land/r/gnoland/blog                   # Show internals
```

## Verbs

| Verb | Description |
|------|-------------|
| **GET** (default) | Render for realms, EVAL for calls, READ for symbols |
| **EVAL** | Evaluate a read-only function call |
| **READ** | Read source code, file, or variable value |
| **INSPECT** | Inspect network, realm, or symbol in detail |
| **CALL** | Sign and broadcast a MsgCall transaction |
| **RUN** | Generate code and broadcast a MsgRun transaction |

## How it works

- **Auto-discovery**: gnopie fetches the domain's homepage and reads `<meta name="gnoconnect:rpc">` tags. No manual config.
- **Auto-gas**: CALL and RUN simulate first, then broadcast with estimated gas + 20% buffer (configurable).
- **Crossing functions**: `cross` is auto-injected for realm functions that require it.
- **Caching**: Network discovery (24h) and query results (1h) are cached in `$GNOHOME/gnopie/cache/`.

## Configuration

```bash
gnopie config set key=moul    # default signing key
gnopie config get key
gnopie config list
```

## Shell completion

```bash
eval "$(gnopie completion bash)"
gnopie completion zsh > ~/.zsh/completions/_gnopie
```
