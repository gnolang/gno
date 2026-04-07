# gnopie

**gnopie** — like [httpie](https://httpie.io), but for gno.land.

An opinionated CLI for interacting with gno.land chains. Uses `gnokey` as a Go library under the hood. Network configuration is auto-discovered via [gnoconnect](https://docs.gno.land) meta tags — no manual config needed.

## Install

```bash
cd contribs/gnopie && make install
gnopie config set key=moul  # set your default signing key
```

## Examples

### Browse a realm (calls Render by default)

```bash
$ gnopie gno.land/r/gnoland/blog
# Gno.land's blog

<gno-columns>
### [Gno.land Beta Mainnet is Live](/r/gnoland/blog:p/beta-mainnet)
25 Mar 2026
...
```

### Paste gnoweb URLs directly

```bash
$ gnopie https://gno.land/r/gnoland/blog:p/beta-mainnet
# Gno.land Beta Mainnet is Live
...
```

### Evaluate a function (read-only)

```bash
$ gnopie 'gno.land/r/gnoland/blog.Render("")'
("# Gno.land's blog\n..." string)

$ gnopie 'gno.land/r/gnoland/wugnot.BalanceOf("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")'
(0 int64)
```

### Inspect a network

```bash
$ gnopie gno.land
Network: gno.land
  RPC:          https://rpc.betanet.testnets.gno.land
  Chain ID:     gnoland1
  Block height: 461788
  App version:  dev
```

### Inspect a realm

```bash
$ gnopie INSPECT gno.land/r/gnoland/blog
Realm: gno.land/r/gnoland/blog
Storage: storage: 1208835, deposit: 120883500

Files:
  admin.gno
  admin_test.gno
  gnoblog.gno
  gnoblog_test.gno
  gnomod.toml
  util.gno

Functions:
  func AdminSetAdminAddr(...)
  func ModAddPost(...)
  func Render(.arg_0 string) .res.0 string
  ...
```

### Read a function's source code

```bash
$ gnopie READ gno.land/r/gnoland/blog.ModAddPost
// gno.land/r/gnoland/blog/admin.gno
func ModAddPost(_ realm, slug, title, body, publicationDate, authors, tags string) {
	assertIsModerator()
	caller := runtime.OriginCaller()
	addPost(caller, slug, title, body, publicationDate, authors, tags)
}
```

### Read a specific file

```bash
$ gnopie READ gno.land/r/gnoland/blog/admin.gno
package gnoblog

import (
	"chain/runtime"
	...
```

### Read source via gnoweb URL

```bash
$ gnopie 'https://gno.land/r/gnoland/blog$source&file=admin.gno'
package gnoblog
...
```

### Inspect an address

```bash
$ gnopie g1manfred47kzduec920z88wfr64ylksmdcedlf5
Address: g1manfred47kzduec920z88wfr64ylksmdcedlf5
{
  "BaseAccount": {
    "address": "g1manfred47kzduec920z88wfr64ylksmdcedlf5",
    "coins": "754954090ugnot",
    ...
```

### Execute a transaction (CALL)

```bash
$ gnopie CALL 'gno.land/r/demo/wugnot.Deposit()' --send 1000000ugnot
Enter password (moul):
Estimating gas...
Estimated gas: 150000
TX committed — height: 461800, hash: ABC123...
```

### Execute via maketx run (RUN)

```bash
$ gnopie RUN 'gno.land/r/demo/boards.CreateThread(1,"Hello","World")'
Enter password (moul):
Estimating gas...
TX committed — height: 461801, hash: DEF456...
```

### Print equivalent gnokey command

```bash
$ gnopie CALL --print-gnokey-command 'gno.land/r/demo/wugnot.Deposit()' --send 1000000ugnot
gnokey \
  maketx \
  call \
  -broadcast \
  -chainid=gnoland1 \
  -remote=https://rpc.betanet.testnets.gno.land \
  -gas-wanted=10000000 \
  -gas-fee=1000000ugnot \
  -send=1000000ugnot \
  -pkgpath=gno.land/r/demo/wugnot \
  -func=Deposit \
  moul
```

### JSON output for piping

```bash
$ gnopie --json gno.land | jq .block_height
461788
```

### Debug mode

```bash
$ gnopie --debug gno.land/r/gnoland/blog
[debug] args: [gno.land/r/gnoland/blog]
[debug] verb=GET expr=gno.land/r/gnoland/blog
[debug] path parsed: kind=2 domain=gno.land pkgpath=gno.land/r/gnoland/blog symbol= args=[]
[debug] GET dispatch → Render (package)
[debug] cache hit for gno.land (rpc=https://rpc.betanet.testnets.gno.land, chainid=gnoland1)
[debug] qrender: gno.land/r/gnoland/blog:
# Gno.land's blog
...
```

## Verbs

| Verb | Description |
|------|-------------|
| **GET** (default) | Smart dispatch: Render for realms, EVAL for calls, READ for symbols |
| **EVAL** | Evaluate a read-only function call via qeval |
| **READ** | Read source code, file, or variable value |
| **INSPECT** | Inspect network, namespace, realm, or symbol in detail |
| **CALL** | Sign and broadcast a MsgCall transaction |
| **RUN** | Generate code and broadcast a MsgRun transaction |

**CALL** sends a `MsgCall` directly. **RUN** wraps the expression in a `main()` package with imports and sends a `MsgRun`.

## Configuration

```bash
gnopie config set key=moul    # set default signing key
gnopie config get key          # get a config value
gnopie config list             # show all settings
```

## Network Discovery

gnopie auto-discovers network configuration by fetching the domain's homepage and reading `<meta name="gnoconnect:rpc">` and `<meta name="gnoconnect:chainid">` tags. Results are cached in `$GNOHOME/gnopie/cache/` for 24h.

## Shell Completion

```bash
eval "$(gnopie completion bash)"
gnopie completion zsh > ~/.zsh/completions/_gnopie
gnopie completion fish > ~/.config/fish/completions/gnopie.fish
```

## Roadmap

- Name resolution (`@moul` → address via `r/sys/users`)
- Transaction history via indexer
- localhost/gnodev integration
- Interactive/REPL mode
