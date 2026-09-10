# Joining mainnet as a validator

> **WORK IN PROGRESS** — the chain-id (`gnoland-1`) and the endpoints (`gno.land`, `rpc.gno.land`, `seed-1.gno.land`/`seed-2.gno.land`) are final; the genesis sha and the seed node IDs are placeholders until launch. Grep `TODO(mainnet)` across this folder.

How to run a full node on **mainnet** and put yourself forward as a validator candidate. This assumes you're comfortable with Go, Docker, and `gnokey` — it only covers what's specific to mainnet.

The flow:

1. Get the binaries.
2. Download and verify the genesis.
3. Configure your node.
4. Start it and let it sync.
5. Register as a valoper candidate — GovDAO then votes you into the set.

## 1. Binaries

Everything is built from the **`chain/mainnet`** branch (<https://github.com/gnolang/gno/tree/chain/mainnet>).

Build from source:

```shell
git clone https://github.com/gnolang/gno.git
cd gno && git checkout chain/mainnet
make -C gno.land install.gnoland install.gnokey   # installs to $GOPATH/bin
```

Or build a Docker image:

```shell
docker build --target gnoland -t gnoland:mainnet .
```

Prebuilt `gnoland`/`gnokey` binaries are on the release page (below). Prebuilt container images are on the GitHub Container Registry, at `ghcr.io/gnolang/gno/gnoland`.

## 2. Genesis

Download `genesis.json` from the [release page](https://github.com/gnolang/gno/releases/tag/chain%2Fmainnet):

```shell
wget -O genesis.json https://github.com/gnolang/gno/releases/download/chain/mainnet/genesis.json
```

Verify its SHA256 — it must match:

```shell
shasum -a 256 genesis.json
# GENESIS_SHA256_PLACEHOLDER — TODO(mainnet): filled once every launch value is final  genesis.json
```

To regenerate the genesis yourself instead of downloading it, see [`README.md`](./README.md).

## 3. Configure your node

Generate a default config and your keys:

```shell
gnoland config init       # writes a default config.toml in your node directory
gnoland secrets init      # generates your validator + node keys
```

Then set the following (edit `config.toml`, or use `gnoland config set <key> <value>`).

**Required — chain-wide, must match exactly:**

| Key | Value |
| --- | --- |
| `p2p.persistent_peers` | `<node-id>@seed-1.gno.land:26656,<node-id>@seed-2.gno.land:26656` — TODO(mainnet): the two seed node IDs land here with the infra handoff |
| `application.prune_strategy` | `syncable` |
| `consensus.timeout_commit` | `3s` |
| `consensus.peer_gossip_sleep_duration` | `10ms` |
| `p2p.flush_throttle_timeout` | `10ms` |

**Per node — set to your own values:**

| Key | Value |
| --- | --- |
| `moniker` | a recognizable name for your node |
| `p2p.external_address` | your public `host:26656`, so peers can dial you back |
| `p2p.pex` | `true` for a standalone node |

Running a sentry-node setup instead of a standalone node? See the [Sentry-node architecture](https://github.com/gnolang/gno/blob/master/gno.land/cmd/gnoland/README.md#sentry-node-architecture) section of the `gnoland` README.

**Advised:**

| Key | Value |
| --- | --- |
| `mempool.size` | `10000` |
| `p2p.max_num_outbound_peers` | `40` |

## 4. Start the node

```shell
gnoland start \
  --chainid gnoland-1 \
  --genesis genesis.json \
  --skip-genesis-sig-verification
```

`--skip-genesis-sig-verification` is **required**: some genesis transactions carry placeholder or intentionally-invalidated signatures (e.g. the `names.Enable` call runs with a patched caller), so the node panics on startup without it.

Let the node sync, and wait until it has caught up to the chain tip before the next step.

## 5. Register as a validator candidate

Get your node's consensus public key:

```shell
gnoland secrets get validator_key   # note the validator public key (gpub1...)
```

The registration transaction costs a gas fee, so your operator account needs GNOT. There is **no faucet on mainnet**, and transfers are **locked at launch** per Constitution §126 — you cannot move GNOT from one of your addresses to another unless the sender is whitelisted. So use an address that already holds GNOT in the genesis allocation as your operator address: gas fees are payable from a locked balance (fee collection bypasses the transfer restriction). <!-- TODO(mainnet): confirm the recommended funding path for operators with no allocation (whitelisted-treasury transfer?). -->

Register your profile on the valoper realm, **signed by your operator key** (the `gnokey` account whose `g1...` address you pass as the operator address — the realm rejects the call if the signer doesn't control that address):

```shell
gnokey maketx call \
  --pkgpath gno.land/r/gnops/valopers \
  --func Register \
  --args "<moniker>" \
  --args "<description>" \
  --args "<cloud|on-prem|data-center>" \
  --args "<your operator g1... address>" \
  --args "<your gpub1... consensus pubkey>" \
  --gas-fee 1000000ugnot --gas-wanted 50000000 \
  --chainid gnoland-1 \
  --remote https://rpc.gno.land \
  --broadcast \
  <your-key-name>
```

Registering only lists you as a **candidate**. A GovDAO member must then create and pass a proposal to add you to the active validator set (via `r/sys/validators/v3`). Once that proposal executes, your node joins the valset.

You can review registered valopers and the current set at <https://gno.land/r/gnops/valopers> and <https://gno.land/r/sys/validators/v3>.
