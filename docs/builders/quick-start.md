# Quick Start

> [Getting started](./getting-started.md) — full walkthrough

**Local development:**

```bash
# 1. Install the toolchain (gno, gnokey, gnodev)
curl -fsSL https://raw.githubusercontent.com/gnolang/gno/master/misc/install.sh | sh

# 2. Create a realm
mkdir counter && cd counter
gno mod init gno.land/r/myname/counter

# 3. Fetch example code
curl -fsSL https://raw.githubusercontent.com/gnolang/gno/master/examples/gno.land/r/demo/counter/counter.gno -o counter.gno
curl -fsSL https://raw.githubusercontent.com/gnolang/gno/master/examples/gno.land/r/demo/counter/counter_test.gno -o counter_test.gno

# 4. Run a local chain (hot reload)
gnodev .
# → open http://localhost:8888
```

**Deploy to staging:**

```bash
# 5. Create a key, then fund it at https://faucet.gno.land
gnokey add dev
gnokey list   # copy the g1... address

# 6. Deploy your package
# (first deploy may require signing the CLA — see Getting started)
gnokey maketx addpkg \
  -pkgpath "gno.land/r/<your-g1-addr>/counter" -pkgdir . \
  -gas-fee 1000000ugnot -gas-wanted 20000000 -broadcast \
  -chainid staging -remote https://rpc.staging.gno.land:443 dev

# 7. Call a realm function
gnokey maketx call \
  -pkgpath "gno.land/r/<your-g1-addr>/counter" \
  -func "Increment" -args "5" \
  -gas-fee 1000000ugnot -gas-wanted 2000000 -broadcast \
  -chainid staging -remote https://rpc.staging.gno.land:443 dev
```