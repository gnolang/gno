# gnoland1 Validator Setup

Basic, tested instructions for joining the `gnoland1` network as a validator. Advanced operators may adapt these steps to their own infrastructure (Docker, systemd, etc.) at their discretion.

## Prerequisites

The following tools must be installed before proceeding:

- **Go** — required to build node binaries
- **make**
- **curl**
- **jq**
- **python3**
- **gzip**

Hardware: at least **16 GB RAM** is recommended. Node startup temporarily exceeds 8 GB during genesis execution. See also the [dedicated gnops article](https://gnops.io/articles/effective-gnops/validator-specs/).

## 1. Generate the Genesis

Every validator must produce the same `genesis.json`. Run from this directory:

```shell
make generate
```

This generates the genesis and verifies the sha256 checksum automatically.

## 2. Build & Install the Node

From the **repository root** (`gno/`):

```shell
make install.gnoland install.gnokey
```

This installs `gnoland` and `gnokey` to your `$GOPATH/bin`.

## 3. Initialize Secrets and Configure

```shell
gnoland secrets init
```

For remote signing via `gnokms`, see the [gnokms documentation](../../../contribs/gnokms/README.md).

Copy the provided config and edit the `# TODO` fields:

```shell
mkdir -p gnoland-data/config
cp config.toml gnoland-data/config/config.toml
grep -n TODO gnoland-data/config/config.toml   # shows what to change
```

Fields to update:

- `moniker` — human-readable name for your node, e.g. `"myorg-val-01"`
- `external_address` — your public IP/host for P2P, e.g. `"tcp://1.2.3.4:26656"`
- `service_instance_id` — telemetry identifier, e.g. `"myorg-val-01"`

## 4. Start the Node

```shell
gnoland start \
  --skip-genesis-sig-verification
```

The `--skip-genesis-sig-verification` flag is required (known incompatibility between genesis signatures and custom package metadata).

`config.toml` already includes a persistent peer, so the node will connect to the network automatically on startup — no extra peering configuration needed.

## 5. Verify & Join

1. After block 1, verify the AppHash at block 2 matches (confirms deterministic genesis execution):

   ```shell
   curl -s http://localhost:26657/block?height=2 | jq -r '.result.block.header.app_hash'
   # expected: TBD
   ```

2. Confirm your node is syncing — `latest_block_height` should be increasing and eventually match [the network RPC](https://rpc.betanet.gno.land/status).
3. Make sure both your RPC port (`26657`) and P2P port (`26656`) are publicly reachable: `http(s)://<your-host>:26657/status`
4. Get your validator key info to share with the team:

   ```shell
   gnoland secrets get -raw validator_key.address
   gnoland secrets get -raw validator_key.pub_key
   ```

5. Ping the team on the validators Signal group so we can add you via a GovDAO proposal. Include your validator address and public key from the step above.
