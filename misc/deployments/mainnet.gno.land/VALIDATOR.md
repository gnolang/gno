# Joining mainnet as a validator

> Launch: **2026-09-12T15:00:00Z**, chain-id `gnoland-1`. All values below are final.

How to run a full node on **mainnet** and put yourself forward as a validator candidate. This assumes you're comfortable with Go, Docker, and `gnokey` — it only covers what's specific to mainnet.

The flow:

1. Get the binaries.
2. Download and verify the genesis.
3. Configure your node.
4. Start it and let it sync.
5. Register as a valoper candidate — GovDAO then votes you into the set.

## 1. Binaries

Run the version in the last row of [`UPGRADES.md`](./UPGRADES.md), pinned. Never a floating tag (`latest`, `chain-mainnet`): the binary changes at every coordinated upgrade, and a node refuses to start a newer version before its halt height.

Docker (replace `v1.5.0` with the current version):

```shell
docker pull ghcr.io/gnolang/gno/gnoland:v1.5.0
docker run --rm ghcr.io/gnolang/gno/gnoland:v1.5.0 version   # must print v1.5.0
```

Binaries: attached to that version's [release page](https://github.com/gnolang/gno/releases/tag/v1.5.0), with `CHECKSUMS.txt`. To run a node from a bare binary, point `GNOROOT` at a checkout of the same tag (the node reads `gnovm/stdlibs` from it).

From source, from the tag — never from `master` or from the branch tip, which reach no consensus with the network and satisfy no upgrade gate:

```shell
git clone https://github.com/gnolang/gno.git
cd gno && git checkout v1.5.0
make -C gno.land install.gnoland install.gnokey   # installs to $GOPATH/bin
gnoland version                                     # must print v1.5.0
```

## 2. Genesis

Download `genesis.json` from the [release page](https://github.com/gnolang/gno/releases/tag/chain%2Fmainnet):

```shell
wget -O genesis.json https://github.com/gnolang/gno/releases/download/chain/mainnet/genesis.json
```

Or the compressed variant (the raw file is ~324 MB):

```shell
wget -O genesis.json.gz https://github.com/gnolang/gno/releases/download/chain/mainnet/genesis.json.gz
gunzip genesis.json.gz
```

Verify its SHA256 — it must match:

```shell
shasum -a 256 genesis.json
# ea22691003130eae3ba975b7d16460706b5d75ce6c04ae82c0c4faeab7de91f0  genesis.json
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
| `p2p.persistent_peers` | `g15rcv5yqef3kvnmueqvkyw8y05sd40jz9p3n5su@seed-1.gno.land:26656,g1ck2yeyvvnpl92237gcea0z68jx07a4nnyvuaan@seed-2.gno.land:26656` |
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

**Syncing from genesis stops at every past upgrade.** Each coordinated halt in [`UPGRADES.md`](./UPGRADES.md) fires again during replay: the node stops after committing that halt height. Restart it — with the same binary if it satisfies that row's `halt_min_version`, otherwise with that row's version — and it continues. If a restart lands between a halt proposal's execution and its halt height while your binary already satisfies that halt's version, the node refuses to start; set `skip_upgrade_height` to that height in `config.toml` for that one restart, or use the previous version until the height.

## 5. Register as a validator candidate

Get your node's consensus public key:

```shell
gnoland secrets get validator_key   # note the validator public key (gpub1...)
```

The registration transaction costs a gas fee, so your operator account needs GNOT. There is **no faucet on mainnet**, and transfers are **locked at launch** per Constitution §126 — you cannot move GNOT from one of your addresses to another unless the sender is whitelisted.

**Founding validators are funded at genesis**: both your signing address and your operator address hold 1,000 GNOT from the §122 Validator Services Treasury, which covers registration and ordinary management txs. Gas fees are payable from a locked balance — fee collection bypasses both the transfer restriction and the vesting schedule — so the §132 schedule on those rows does not get in your way.

If you join *after* genesis, your operator address will not be in the allocation. Use an address that already holds GNOT in the genesis allocation as your operator address. <!-- TODO(mainnet): confirm the recommended funding path for post-genesis operators with no allocation (whitelisted-treasury transfer?). -->

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

Registering only lists you as a **candidate**. A GovDAO member must then create and pass a proposal to add you to the active validator set (via `r/sys/validators/v0`). Once that proposal executes, your node joins the valset.

You can review registered valopers and the current set at <https://gno.land/r/gnops/valopers> and <https://gno.land/r/sys/validators/v0>.
