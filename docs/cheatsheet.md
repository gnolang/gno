# Gno Cheatsheet

Minimal copy-paste commands. Each section links to the full guide for the details.

## Table of Contents

### User

- [Install](#install)
- [Generate a Key](#generate-a-key)
- [Manage Keys](#manage-keys)
- [Query](#query)
- [Call a Function](#call-a-function)
- [Send Coins](#send-coins)
- [Deploy a Package](#deploy-a-package)
- [Multisig](#multisig)
- [Airgap Transaction](#airgap-transaction)
- [Verify a Signature](#verify-a-signature)

### Developer

- [Create a Realm](#create-a-realm)
- [Run Locally](#run-locally)
- [Test](#test)
- [Format & Lint](#format--lint)
- [Create a Run Script](#create-a-run-script)
- [Deploy to Staging](#deploy-to-staging)

### Valoper

- [Init Validator Secrets](#init-validator-secrets)
- [Register Valoper Profile](#register-valoper-profile)
- [Update Valoper Profile](#update-valoper-profile)
- [Query Valopers](#query-valopers)

### Contributor

- [Build & Test Go](#build--test-go)
- [Start a Local Chain](#start-a-local-chain)
- [Update Golden Files](#update-golden-files)
- [Lint & Format Go](#lint--format-go)

---

## User

### Install

> [Installation](builders/install.md)

```bash
# install pre-built binaries
curl -fsSL https://raw.githubusercontent.com/gnolang/gno/master/misc/install.sh | sh

# uninstall binaries in $GOPATH/bin
curl -fsSL https://raw.githubusercontent.com/gnolang/gno/master/misc/uninstall.sh | sh
```

### Generate a Key

> [Generating a key pair](users/interact-with-gnokey.md#generating-a-key-pair)

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

### Manage Keys

> [Managing key pairs](users/interact-with-gnokey.md#managing-key-pairs)

```bash
# delete a key
gnokey delete MyKey

# export an encrypted armored key to a file
gnokey export -key MyKey -output-path mykey.asc

# import an encrypted armored key from a file
gnokey import -armor-path mykey.asc -name MyKey

# rotate the password protecting a key
gnokey rotate MyKey

# recover a key from an existing BIP39 mnemonic (prompts for the phrase)
gnokey add -recover MyKey

# recover at a custom HD account / index
gnokey add -recover -account 0 -index 1 MyKey

# generate a fresh BIP39 mnemonic (does not save a key)
gnokey generate
```

### Query

```bash
# render the realm output
gnokey query vm/qrender -data "gno.land/r/dev/counter:"

# evaluate an expression (read-only, no gas)
gnokey query vm/qeval -data "gno.land/r/dev/counter.Render(\"\")"

# check account balance
gnokey query bank/balances/$ADDRESS

# get account info (number + sequence for signing)
gnokey query auth/accounts/$ADDRESS
```

### Call a Function

> [`Call`](users/interact-with-gnokey.md#call)

```bash
gnokey maketx call \
  -pkgpath "gno.land/r/dev/counter" \
  -func "Increment" \
  -args "42" \
  -gas-fee 1000000ugnot \
  -gas-wanted 20000000 \
  MyKey
```

### Send Coins

> [`Send`](users/interact-with-gnokey.md#send)

```bash
gnokey maketx send \
  -send "1000000ugnot" \
  -to "$RECIPIENT_ADDR" \
  -gas-fee 1000000ugnot \
  -gas-wanted 2000000 \
  -chainid "staging" \
  -remote "https://rpc.staging.gno.land:443" \
  MyKey
```

### Deploy a Package

> [`AddPackage`](users/interact-with-gnokey.md#addpackage)

```bash
# upload package files at ./counter to the chain
gnokey maketx addpkg \
  -pkgpath "gno.land/r/<your_g1_address>/counter" \
  -pkgdir "./counter" \
  -gas-fee 1000000ugnot \
  -gas-wanted 20000000 \
  -chainid "staging" \
  -remote "https://rpc.staging.gno.land:443" \
  MyKey

# run an arbitrary script (Run() entrypoint) without publishing
gnokey maketx run \
  -gas-fee 1000000ugnot \
  -gas-wanted 20000000 \
  MyKey ./script.gno
```

### Multisig

> Full walkthrough: [using a k-of-n multisig](users/interact-with-gnokey.md#using-a-k-of-n-multisig)

Members must be listed in the same fixed order in every signer's keybase.

```bash
# add a k-of-n multisig key (order matters, identical for every signer)
gnokey add multisig \
  -multisig alice -multisig bob -multisig charlie \
  -threshold 2 \
  multisig-abc

# each member signs the shared unsigned tx, then combine and broadcast
gnokey multisign -tx-path tx.json -signature alice.sig -signature bob.sig multisig-abc
gnokey broadcast tx.json
```

### Airgap Transaction

> Full walkthrough: [making an airgapped transaction](users/interact-with-gnokey.md#making-an-airgapped-transaction)

Build and sign offline, broadcast online. Fetch the account number and sequence online first.

```bash
# offline: create the unsigned tx (-broadcast=false), then sign it
gnokey maketx call \
  -pkgpath "gno.land/r/dev/counter" \
  -func "Increment" \
  -gas-fee 1000000ugnot \
  -gas-wanted 2000000 \
  -broadcast=false \
  MyKey > counter.tx

gnokey sign \
  -tx-path counter.tx \
  -chainid "staging" \
  -account-number <n> -account-sequence <s> \
  MyKey

# online: broadcast the signed tx
gnokey broadcast -remote "https://rpc.staging.gno.land:443" counter.tx
```

### Verify a Signature

> Full walkthrough: [verifying a signature](users/interact-with-gnokey.md#verifying-a-transactions-signature)

```bash
# verify the signature embedded in a signed tx
gnokey verify \
  -tx-path tx.json \
  -chainid "staging" \
  -account-number <n> -account-sequence <s> \
  MyKey

# verify a detached signature file instead
gnokey verify \
  -tx-path tx.json \
  -sig-path alice.sig \
  -chainid "staging" \
  -account-number <n> -account-sequence <s> \
  MyKey
```

---

## Developer

### Create a Realm

> [Gno packages](resources/gno-packages.md)

```bash
mkdir counter && cd counter

# create gnomod.toml
gno mod init gno.land/r/example/counter
```

### Run Locally

> [Local development with `gnodev`](resources/gnodev.md)

```bash
# starts a local node + gnoweb on http://localhost:8888
gnodev

# with remote resolver (for missing dependencies)
gnodev -resolver remote=https://rpc.staging.gno.land:443

# without hot reload
gnodev -no-watch
```

### Test

> [Running & testing Gno code](resources/gno-testing.md)

```bash
# run tests for current package
gno test -v .

# run only filetests
gno test -run "_filetest.gno" .
```

### Format & Lint

```bash
gno fmt .
gno lint .
```

### Create a Run Script

> [`Run`](users/interact-with-gnokey.md#run)

```bash
# write run/create_proposal.gno, then run:
gnokey maketx run \
  -gas-fee 1000000ugnot \
  -gas-wanted 20000000 \
  MyKey ./run/create_proposal.gno
```

### Deploy to Staging

> [Deploy to a shared network](builders/getting-started.md#deploy-to-a-shared-network) | [Networks](resources/gnoland-networks.md)

```bash
# get testnet GNOT from https://faucet.gno.land

# deploy the realm to the staging network
gnokey maketx addpkg \
  -pkgpath "gno.land/r/<your_g1_address>/counter" \
  -pkgdir "." \
  -gas-fee 10000000ugnot \
  -gas-wanted 8000000 \
  -chainid staging \
  -remote "https://rpc.staging.gno.land:443" \
  MyKey
```

---

## Valoper

### Init Validator Secrets

```bash
# initialize validator key, node key, and signing state in a directory
gnoland secrets init -data-dir gnoland-data

# verify secrets are valid
gnoland secrets verify -data-dir gnoland-data

# show validator address + pubkey
gnoland secrets get -data-dir gnoland-data validator_key
```

### Register Valoper Profile

> Realm: `gno.land/r/gnops/valopers`

```bash
gnokey maketx call \
  -pkgpath "gno.land/r/gnops/valopers" \
  -func "Register" \
  -args "$MONIKER" \
  -args "$DESCRIPTION" \
  -args "$SERVER_TYPE" \
  -args "$ADDRESS" \
  -args "$PUBKEY" \
  -gas-fee 1000000ugnot \
  -gas-wanted 20000000 \
  MyKey
```

### Update Valoper Profile

```bash
# update moniker
gnokey maketx call \
  -pkgpath "gno.land/r/gnops/valopers" \
  -func "UpdateMoniker" \
  -args "$ADDRESS" \
  -args "$NEW_MONIKER" \
  -gas-fee 1000000ugnot -gas-wanted 20000000 \
  MyKey

# update description / server type / keep-running flag: same pattern with
# UpdateDescription, UpdateServerType, UpdateKeepRunning
```

### Query Valopers

```bash
# render the full valoper list
gnokey query vm/qrender -data "gno.land/r/gnops/valopers:"

# fetch a single valoper by address
gnokey query vm/qeval -data "gno.land/r/gnops/valopers.GetByAddr(\"$ADDRESS\")"
```

---

## Contributor

> [Contributing guide](https://github.com/gnolang/gno/blob/master/CONTRIBUTING.md)

### Build & Test Go

```bash
# install all binaries
make install

# run all Go tests
make test

# run tests for a specific component
make -C gnovm test
make -C gno.land test
```

### Start a Local Chain

> [Local development with `gnodev`](resources/gnodev.md)

```bash
# lightweight in-memory node (recommended for dev)
gnodev

# full persistent node with genesis
gnoland start

# with custom genesis and data dir
gnoland start -genesis genesis.json -data-dir gnoland-data
```

### Update Golden Files

> [Running & testing Gno code](resources/gno-testing.md)

```bash
# update golden outputs for *_filetest.gno files in current package
# (only applies to filetests; regular _test.gno files are unaffected)
gno test --update-golden-tests .

# update gnovm file tests
go test ./gnovm/pkg/gnolang/files_test.go -test.short --update-golden-tests

# update examples golden files
make -C examples test GOLDEN=1
```

### Lint & Format Go

```bash
# format all Go code
make fmt

# run linter
make lint

# tidy go.mod files
make tidy
```
