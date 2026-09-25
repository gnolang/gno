# Joining onyx as a validator

> Launch: **2026-09-28T00:00:00Z**, chain-id `onyx-1`. Onyx runs mainnet's exact binaries and is upgraded whenever mainnet is.

How to run a full node on **onyx** and put yourself forward as a validator candidate. This assumes you're comfortable with Go, Docker, and `gnokey` — it only covers what's specific to onyx.

The flow:

1. Get the binaries.
2. Download and verify the genesis.
3. Configure your node.
4. Start it and let it sync.
5. Register as a valoper candidate — GovDAO then votes you into the set.

## 1. Binaries

Run the version in the last row of [`UPGRADES.md`](./UPGRADES.md) — `v1.5.0` in the examples below — and pin it: the binary changes at every coordinated upgrade, and a node refuses to start a newer version before its halt height.

**Native binary (recommended for validators).** Download `gnoland` for your platform from the version's [release page](https://github.com/gnolang/gno/releases/tag/v1.5.0), check it against `CHECKSUMS.txt` there (the same checksums are in `upgrades.json`), and point `GNOROOT` at a checkout of the same tag — the node reads `gnovm/stdlibs` from it:

```shell
curl -fsSLO https://github.com/gnolang/gno/releases/download/v1.5.0/gnoland_linux_amd64
curl -fsSLO https://github.com/gnolang/gno/releases/download/v1.5.0/CHECKSUMS.txt
shasum -a 256 gnoland_linux_amd64                   # must match the line in CHECKSUMS.txt
git clone --branch v1.5.0 --depth 1 https://github.com/gnolang/gno.git ~/gno
export GNOROOT=~/gno
chmod +x gnoland_linux_amd64 && ./gnoland_linux_amd64 version   # must print v1.5.0
```

Or build it from the tag (`git checkout v1.5.0 && make -C gno.land install.gnoland install.gnokey`; `gnoland version` must print `v1.5.0`). Never from `master` or from a branch tip: such a binary reaches no consensus with the network and satisfies no upgrade gate.

**Container image.** `ghcr.io/gnolang/gno/gnoland:v1.5.0` is the same binary with `GNOROOT` preset. Weigh what it adds before running a validator on it: a base image you also have to trust, a registry that must be reachable when you restart, a root daemon in the path, and one more way to end up with two instances of your validator signing at once — a restart policy, a `compose up` on a second host, a forgotten container. Double-signing is not punished today; it will be. If you use it: never `latest`, never a restart policy, and mount `secrets/` read-only.

```shell
docker pull ghcr.io/gnolang/gno/gnoland:v1.5.0
docker run --rm ghcr.io/gnolang/gno/gnoland:v1.5.0 version   # must print v1.5.0
```

## 2. Genesis

Download `genesis.json` from the [release page](https://github.com/gnolang/gno/releases/tag/chain%2Fonyx):

```shell
wget -O genesis.json https://github.com/gnolang/gno/releases/download/chain/onyx/genesis.json
```

Verify its SHA256 — it must match the value on the release page and in [`upgrades.json`](./upgrades.json) (`genesis_sha256`):

```shell
shasum -a 256 genesis.json
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
| `p2p.persistent_peers` | `g1x5mlj5ava0dw9vkf4j6admjlzswm6f06p44krn@seed-1.onyx.testnets.gno.land:26656,g1grq5zswt0dlwwe7clr4359w70k2ewgse0gcwck@seed-2.onyx.testnets.gno.land:26656` |
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
  --chainid onyx-1 \
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

The registration transaction costs a gas fee, so your operator account needs GNOT. If it's empty, request a drip for your `g1...` address from the onyx faucet at <https://onyx.testnets.gno.land/faucet>.

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
  --chainid onyx-1 \
  --remote https://rpc.onyx.testnets.gno.land \
  --broadcast \
  <your-key-name>
```

Registering only lists you as a **candidate**. A GovDAO member must then create and pass a proposal to add you to the active validator set (via `r/sys/validators/v0`). Once that proposal executes, your node joins the valset.

You can review registered valopers and the current set at <https://onyx.testnets.gno.land/r/gnops/valopers> and <https://onyx.testnets.gno.land/r/sys/validators/v0>.
