# Cheatsheet

## Developer

- [Install](#install)
- [Create a Realm](#create-a-realm)
- [Create a Run Script](#create-a-run-script)
- [Generate a Key](#generate-a-key)
- [Test](#test)
- [Format & Lint](#format--lint)
- [Run Locally](#run-locally)
- [Query](#query)
- [Call a Function](#call-a-function)
- [Airgap Transaction](#airgap-transaction)
- [Deploy to Staging](#deploy-to-staging)

## Contributor

- [Build & Test Go](#build--test-go)
- [Start a Local Chain](#start-a-local-chain)
- [Update Golden Files](#update-golden-files)
- [Lint & Format Go](#lint--format-go)

---

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

# interactive wizard — picks kind, template, generates starter code (gnolang/gno#5557)
gno init

# or non-interactive
gno init gno.land/r/example/counter

# or bare (gnomod.toml only)
gno mod init gno.land/r/example/counter
```

## Create a Run Script

> [Using `gnokey`](../users/interact-with-gnokey.md#run)

```bash
# .gno shorthand creates the run directory + starter script (gnolang/gno#5557)
gno init run/create_proposal.gno

# then run it
gnokey maketx run \
  -gas-fee 1000000ugnot \
  -gas-wanted 20000000 \
  MyKey ./run/create_proposal.gno
```

## Generate a Key

> [Managing key pairs](../users/interact-with-gnokey.md#managing-key-pairs)

```bash
# create a new keypair
gnokey add MyKey

# list existing keys
gnokey list
```

Default `gnodev` test account (`devtest`, `g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5`):
```
source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast
```

## Test

> [Testing Gno](../resources/gno-testing.md)

```bash
# run tests for current package
gno test -v .

# run only filetests
gno test -run "_filetest.gno" .
```

## Format & Lint

> [Effective Gno](../resources/effective-gno.md)

```bash
gno fmt .
gno lint .
```

## Run Locally

> [Local development with `gnodev`](local-dev-with-gnodev.md)

```bash
# starts a local node + gnoweb on http://localhost:8888
gnodev

# with remote resolver (for missing dependencies)
gnodev -resolver remote=https://rpc.staging.gno.land:443

# without hot reload
gnodev -no-watch
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

# get account info (number + sequence for signing)
gnokey query auth/accounts/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
```

## Call a Function

> [Using `gnokey`](../users/interact-with-gnokey.md#call)

```bash
# interactive wizard
gnokey maketx

# or manually
gnokey maketx call \
  -pkgpath "gno.land/r/dev/counter" \
  -func "Increment" \
  -args "42" \
  -gas-fee 1000000ugnot \
  -gas-wanted 20000000 \
  MyKey
```

## Airgap Transaction

> [Making an airgapped transaction](../users/interact-with-gnokey.md#making-an-airgapped-transaction)

```bash
# 1. online machine: fetch account info
gnokey query auth/accounts/<address> -remote "https://rpc.staging.gno.land:443"

# 2. offline machine: create unsigned tx
gnokey maketx call \
  -pkgpath "gno.land/r/demo/counter" \
  -func "Increment" \
  -gas-fee 1000000ugnot \
  -gas-wanted 2000000 \
  mykey > counter.tx

# 3. offline machine: sign the tx
gnokey sign \
  -tx-path counter.tx \
  -chainid "staging" \
  -account-number 468 \
  -account-sequence 0 \
  mykey

# 4. online machine: broadcast the signed tx
gnokey broadcast -remote "https://rpc.staging.gno.land:443" counter.tx
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

---

## Build & Test Go

> [Contributing guide](https://github.com/gnolang/gno/blob/master/CONTRIBUTING.md)

```bash
# install all binaries
make install

# run all Go tests
make test

# run tests for a specific component
make -C gnovm test
make -C gno.land test
```

## Start a Local Chain

> [Local development with `gnodev`](local-dev-with-gnodev.md)

```bash
# lightweight in-memory node (recommended for dev)
gnodev

# full persistent node with genesis
gnoland start

# with custom genesis and data dir
gnoland start -genesis genesis.json -data-dir gnoland-data
```

## Update Golden Files

> [Testing Gno](../resources/gno-testing.md)

```bash
# update golden filetest outputs for current package
gno test --update-golden-tests .

# update gnovm file tests
go test ./gnovm/pkg/gnolang/files_test.go -test.short --update-golden-tests

# update examples golden files
make -C examples test GOLDEN=1
```

## Lint & Format Go

```bash
# format all Go code
make fmt

# run linter
make lint

# tidy go.mod files
make tidy
```

---

## Next Steps

- [Writing Gno code](anatomy-of-a-gno-package.md) - Language basics and package structure
- [Local development with `gnodev`](local-dev-with-gnodev.md) - Hot reload, premining, auto-deploy
- [Deploying packages](deploy-packages.md) - Gas fees, namespaces, deployment details
- [Effective Gno](../resources/effective-gno.md) - Best practices for writing Gno
- [Using `gnokey`](../users/interact-with-gnokey.md) - Full key management and transaction reference
