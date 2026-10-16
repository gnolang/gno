# Quick Start

> [Getting started](./getting-started.md) — full walkthrough

## Run locally

```sh
# 1. Install the toolchain (gno, gnokey, gnodev).
#    Docker, source build, or pinned version: see install.md.
curl -fsSL https://raw.githubusercontent.com/gnolang/gno/master/misc/install.sh | sh

# 2. Create a realm
mkdir counter && cd counter
gno mod init gno.land/r/myname/counter

# 3. Fetch example code (or paste counter.gno below by hand)
curl -fsSL https://raw.githubusercontent.com/gnolang/gno/master/examples/gno.land/r/demo/counter/counter.gno -o counter.gno
curl -fsSL https://raw.githubusercontent.com/gnolang/gno/master/examples/gno.land/r/demo/counter/counter_test.gno -o counter_test.gno

# 4. Run a local chain (hot reload)
gnodev .
# → open http://localhost:8888
```

`counter.gno`:

```gno
package counter

import "strconv"

var count int

func Increment(_ realm) int {
	count++
	return count
}

func Render(path string) string {
	return "Count: " + strconv.Itoa(count)
}
```

## Deploy to the testnet

```sh
# 5. Create a key, fund it at https://faucet.gno.land on the current testnet
#    (faucet is rate-limited per address; one request is enough)
gnokey add dev
gnokey list   # copy the g1... address

# 6. Point at the current testnet; its chain ID changes with each testnet
REMOTE=https://rpc.latest.testnets.gno.land:443
CHAINID=$(curl -fsS $REMOTE/status | grep -o '"network": *"[^"]*"' | cut -d'"' -f4)

# 7. Confirm the faucet landed
gnokey query bank/balances/<your-g1-addr> -remote $REMOTE

# 8. Deploy
gnokey maketx addpkg \
  -pkgpath "gno.land/r/<your-g1-addr>/counter" -pkgdir . \
  -gas-fee 1000000ugnot -gas-wanted 20000000 \
  -chainid "$CHAINID" -remote $REMOTE dev

# 9. Wait for "status":"live": the testnet holds a new package until an
#    approver enables it; https://gpao.latest.testnets.gno.land/status/<pkgpath>
#    gives the reason it refused one
gnokey query vm/qpkgmeta_json \
  -data "gno.land/r/<your-g1-addr>/counter" -remote $REMOTE

# 10. Call a realm function
gnokey maketx call \
  -pkgpath "gno.land/r/<your-g1-addr>/counter" \
  -func "Increment" \
  -gas-fee 1000000ugnot -gas-wanted 2000000 \
  -chainid "$CHAINID" -remote $REMOTE dev
```

Live at **`https://latest.testnets.gno.land/r/<your-g1-addr>/counter`**.

## Check security patterns

Before publishing a non-trivial realm, read the
[Gno Security Guide](../resources/gno-security-guide.md). The
[`misc/audit-pattern-harness`](../../misc/audit-pattern-harness/README.md)
fixtures provide a quick sanity check for recurring audit patterns and show how
to run the executable checks.

## Next

[r/docs](https://github.com/gnolang/gno/tree/master/examples/quarantined/gno.land/r/docs): example realms, one per Gno.land concept.
