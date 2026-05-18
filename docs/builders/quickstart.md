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

## Deploy to staging

```sh
# 5. Create a key, then fund it at https://faucet.gno.land
#    (faucet is rate-limited per address; one request is enough)
gnokey add dev
gnokey list   # copy the g1... address

# 6. Confirm the faucet landed
gnokey query bank/balances/<your-g1-addr> \
  -remote https://rpc.staging.gno.land:443

# 7. Deploy
gnokey maketx addpkg \
  -pkgpath "gno.land/r/<your-g1-addr>/counter" -pkgdir . \
  -gas-fee 1000000ugnot -gas-wanted 20000000 \
  -chainid staging -remote https://rpc.staging.gno.land:443 dev

# 8. Call a realm function
gnokey maketx call \
  -pkgpath "gno.land/r/<your-g1-addr>/counter" \
  -func "Increment" \
  -gas-fee 1000000ugnot -gas-wanted 2000000 \
  -chainid staging -remote https://rpc.staging.gno.land:443 dev
```

Live at **`https://staging.gno.land/r/<your-g1-addr>/counter`**.

## Next

[r/docs](https://gno.land/r/docs) — on-chain tour of Gno.land.
