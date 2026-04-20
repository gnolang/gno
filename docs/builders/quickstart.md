# Quick Start

- [Install](#install)
- [Create a Realm](#create-a-realm)
- [Generate a Key](#generate-a-key)
- [Test](#test)
- [Run Locally](#run-locally)
- [Query](#query)
- [Call a Function](#call-a-function)
- [Scripting](#scripting)
- [Deploy to Staging](#deploy-to-staging)

## Install

> [Full installation guide](TODO)

<!-- TODO: replace with one-line installer (curl | sh) once available (gnolang/gno#5492) -->

```bash
git clone git@github.com:gnolang/gno.git
cd gno && make install
```

## Create a Realm

> [Writing Gno code](anatomy-of-a-gno-package.md)

```bash
mkdir counter && cd counter
gno mod init gno.land/r/example/counter
```

<!-- TODO: replace with `gno init` once it generates a starter realm -->

Write your realm code in `counter.gno`. See [Writing Gno code](anatomy-of-a-gno-package.md)
for a full Counter example.

## Generate a Key

> [Managing key pairs](../users/interact-with-gnokey.md#managing-key-pairs)

```bash
# create a new keypair
gnokey add MyKey
```

## Test

> [Testing Gno](../resources/gno-testing.md)

```bash
gno test -v .
```

## Run Locally

> [Local development with `gnodev`](local-dev-with-gnodev.md)

```bash
# starts a local node + gnoweb on http://localhost:8888
gnodev
```

## Query

> [Using `gnokey`](../users/interact-with-gnokey.md#querying-a-gnoland-network)

```bash
# render the realm output
gnokey query vm/qrender -data "gno.land/r/dev/counter:"

# evaluate an expression (read-only, no gas)
gnokey query vm/qeval -data "gno.land/r/dev/counter.Render(\"\")"

# check account balance
gnokey query bank/balances/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
```

## Call a Function

> [Using `gnokey`](../users/interact-with-gnokey.md#call)

```bash
# call Increment(42) on the counter realm
gnokey maketx call \
  -pkgpath "gno.land/r/dev/counter" \
  -func "Increment" \
  -args "42" \
  -gas-fee 1000000ugnot \
  -gas-wanted 20000000 \
  MyKey
```

## Scripting (maketx run)

> [Using `gnokey`](../users/interact-with-gnokey.md#run)

<!-- TODO: `gno init --main` not working yet, replace once available -->

```bash
# gno init --main gno.land/r/example/counter/run
mkdir run && cd run
gno mod init gno.land/r/example/counter/run

# write your program in main.gno, then run it
gnokey maketx run \
  -gas-fee 1000000ugnot \
  -gas-wanted 20000000 \
  MyKey .
```

## Deploy to Staging

> [Deploying packages](deploy-packages.md) | [Networks](../resources/gnoland-networks.md)

```bash
# get testnet GNOT from https://faucet.gno.land

# deploy the realm to the staging network
gnokey maketx addpkg \
  -pkgpath "gno.land/r/<your_address>/counter" \
  -pkgdir "." \
  -gas-fee 10000000ugnot \
  -gas-wanted 8000000 \
  -chainid staging \
  -remote "https://rpc.staging.gno.land:443" \
  MyKey
```

## Next Steps

- [Writing Gno code](anatomy-of-a-gno-package.md) - Language basics and package structure
- [Local development with `gnodev`](local-dev-with-gnodev.md) - Hot reload, premining, auto-deploy
- [Deploying packages](deploy-packages.md) - Gas fees, namespaces, deployment details
- [Effective Gno](../resources/effective-gno.md) - Best practices for writing Gno
- [Using `gnokey`](../users/interact-with-gnokey.md) - Full key management and transaction reference
