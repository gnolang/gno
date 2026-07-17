# Joining topaz as a validator

How to run a full node on **topaz** and put yourself forward as a validator candidate. This assumes you're comfortable with Go, Docker, and `gnokey` — it only covers what's specific to topaz.

The flow:

1. Get the binaries.
2. Download and verify the genesis.
3. Configure your node.
4. Start it and let it sync.
5. Register as a valoper candidate — GovDAO then votes you into the set.

## 1. Binaries

Everything is built from the **`chain/topaz`** branch (<https://github.com/gnolang/gno/tree/chain/topaz>).

Build from source:

```shell
git clone https://github.com/gnolang/gno.git
cd gno && git checkout chain/topaz
make -C gno.land install.gnoland install.gnokey   # installs to $GOPATH/bin
```

Or build a Docker image:

```shell
docker build --target gnoland -t gnoland:topaz .
```

Prebuilt `gnoland`/`gnokey` binaries are on the release page (below). Prebuilt container images are on the GitHub Container Registry, at `ghcr.io/gnolang/gno/gnoland`.

## 2. Genesis

Download `genesis.json` from the [release page](https://github.com/gnolang/gno/releases/tag/chain%2Ftopaz):

```shell
wget -O genesis.json https://github.com/gnolang/gno/releases/download/chain/topaz/genesis.json
```

Verify its SHA256 — it must match:

```shell
shasum -a 256 genesis.json
# 2dd049f973b82858727440df9aff5722cb0b322fd00890f40f2b0688276898ff  genesis.json
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
| `p2p.seeds` | `g19q07ssuafhmg6r7ys7wp7rpc4jxc85cpvdy426@seed-1.topaz.testnets.gno.land:26656,g15k98e65gm8h7fdr3yr4tqn82lvch4a97a3sg3j@seed-2.topaz.testnets.gno.land:26656` |
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

Running a sentry-node setup instead of a standalone node? Follow the sentry architecture guide on the gnops.io blog (<https://gnops.io>).

**Advised:**

| Key | Value |
| --- | --- |
| `mempool.size` | `10000` |
| `p2p.max_num_outbound_peers` | `40` |

## 4. Start the node

```shell
gnoland start \
  --chainid topaz-1 \
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

The registration transaction costs a gas fee, so your operator account needs GNOT. If it's empty, request a drip for your `g1...` address from the topaz faucet at <https://topaz.testnets.gno.land/faucet>.

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
  --chainid topaz-1 \
  --remote https://rpc.topaz.testnets.gno.land \
  --broadcast \
  <your-key-name>
```

Registering only lists you as a **candidate**. A GovDAO member must then create and pass a proposal to add you to the active validator set (via `r/sys/validators/v3`). Once that proposal executes, your node joins the valset.

You can review registered valopers and the current set at <https://topaz.testnets.gno.land/r/gnops/valopers> and <https://topaz.testnets.gno.land/r/sys/validators/v3>.
