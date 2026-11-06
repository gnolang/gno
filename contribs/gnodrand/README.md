# gnodrand

Relayer for [`gno.land/r/drand/v0`](../../examples/gno.land/r/drand/v0): it brings [drand](https://drand.love) evmnet beacons on-chain for every round that has pending requests.

The realm verifies every beacon's BLS signature on-chain with the BN254 natives, so a relayer is trusted for liveness only: it can deliver the real beacon, or nothing. Anyone can run one, and any number can run side by side.

## Quick start

```sh
make install

# key from the gnokey keybase
GNODRAND_PASSWORD=... gnodrand -key relayer

# or from a mnemonic, against a local gnodev
GNODRAND_MNEMONIC="..." gnodrand -remote http://127.0.0.1:26657 -chain-id dev
```

## How it works

Every `-interval`:

1. Read `PendingRounds` from the realm (a free query).
2. Skip rounds drand has not published yet.
3. Fetch the others from the drand mirrors, in order, and verify each one locally with drand/kyber, so a broken mirror never costs gas.
4. Submit them in one transaction, one `Submit` call per beacon.

With no pending requests, it sends nothing.

## Configuration

| Flag | Default | Meaning |
|---|---|---|
| `-remote` | `https://rpc.gno.land:443` | gno.land RPC |
| `-chain-id` | `gnoland-1` | Chain id |
| `-pkgpath` | `gno.land/r/drand/v0` | Realm to serve |
| `-key` | | gnokey key name or address, password in `GNODRAND_PASSWORD` |
| `-home` | gnokey default | Keybase directory |
| `-fee-per-beacon` | `22000ugnot` | Fee per beacon, multiplied by the batch size |
| `-gas-per-beacon` | `22000000` | Gas wanted per beacon, multiplied by the batch size |
| `-max-batch` | `10` | Beacons per transaction |
| `-interval` | `3s` | Poll interval (the evmnet period) |
| `-drand` | api.drand.sh, api2, api3, drand.cloudflare.com | Mirrors, tried in order |

The chain deducts the whole fee whatever the transaction uses. Keep `-fee-per-beacon` at `-gas-per-beacon` times the network gas price (`gnokey query auth/gasprice`); the defaults match 1 ugnot per 1000 gas.

## Cost

Measured on gnodev, priced at 1 ugnot per 1000 gas and 100 ugnot per stored byte:

| Per delivered round | |
|---|---|
| Gas (`Submit`: verify and store) | 17.9M gas, about 0.018 GNOT |
| Storage deposit | 351 bytes, 35,100 ugnot |

A round is paid once, however many requests share it.

## Development

```sh
make test       # unit tests, including the SVDW reference checked against kyber
make lint
make test.e2e   # gnodev + live drand: request, relay, on-chain verify, settle
```

`internal/evmnet` is the off-chain reference for `gno.land/p/drand/v0`: the gno test vectors (`vectors_test.gno`) were produced with it, and `svdw_test.go` pins the hash-to-curve constants against kyber.

`make test.e2e` needs `gnodev` on `PATH` (or `$GNODEV`) built from the same checkout as `GNOROOT`, and network access to drand.
