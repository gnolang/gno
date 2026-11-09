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

# 6. Confirm the faucet landed
gnokey query bank/balances/<your-g1-addr> \
  -remote https://rpc.latest.testnets.gno.land:443

# 7. Deploy; <chain-id> is the testnet's, listed at
#    https://docs.gno.land/resources/gnoland-networks/
gnokey maketx addpkg \
  -pkgpath "gno.land/r/<your-g1-addr>/counter" -pkgdir . \
  -gas-fee 5000ugnot -gas-wanted 5000000 \
  -chainid <chain-id> -remote https://rpc.latest.testnets.gno.land:443 dev

# 8. Call a realm function. The testnet enables a new package a few seconds
#    after addpkg: if the call fails, wait and retry. A refused one's
#    reason is at https://gpao.latest.testnets.gno.land/status/<pkgpath>,
#    which only knows packages submitted since its service last restarted;
#    if it still reads "unknown" after a minute, ask on Discord at
#    https://discord.com/invite/gnoland
gnokey maketx call \
  -pkgpath "gno.land/r/<your-g1-addr>/counter" \
  -func "Increment" \
  -gas-fee 2000ugnot -gas-wanted 2000000 \
  -chainid <chain-id> -remote https://rpc.latest.testnets.gno.land:443 dev
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
