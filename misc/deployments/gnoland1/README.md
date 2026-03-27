# gnoland1 Validator Setup

Basic, tested instructions for joining the `gnoland1` network as a validator. Advanced operators may adapt these steps to their own infrastructure (Docker, systemd, etc.) at their discretion.

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

## 4. Start the Node

```shell
gnoland start \
  --skip-genesis-sig-verification \
  --genesis genesis.json \
  --data-dir gnoland-data
```

The `--skip-genesis-sig-verification` flag is required (known incompatibility between genesis signatures and custom package metadata).

## 5. Verify & Join

1. After block 1, verify the AppHash at block 2 matches (confirms deterministic genesis execution):

   ```shell
   curl -s http://localhost:26657/block?height=2 | jq -r '.result.block.header.app_hash'
   # expected: TBD
   ```

2. Confirm your node is syncing — `latest_block_height` should be increasing and eventually match [the network RPC](https://rpc.betanet.gno.land/status).
3. Make sure your RPC endpoint is publicly reachable: `http(s)://<your-host>:26657/status`
4. Ping the team on the validators Signal group so we can add you via a GovDAO proposal.
