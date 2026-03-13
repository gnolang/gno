# gnoland1 Validator Setup

This guide walks you through setting up a validator node for the `gnoland1` network. The steps below cover genesis generation, node build, secrets initialization, and connecting to the network.

## 1. Generate the Genesis

Every validator must generate the same genesis file to ensure they are on the same chain. You can generate the genesis file using the provided `make generate` target and verify the sha256 hash of the generated `genesis.json` file matches the expected value: `ac530875f51afaae015cbf54b298bfd7254ac537b0c5f1fd99bfa30bad8d398e`

```shell
make generate
echo 'ac530875f51afaae015cbf54b298bfd7254ac537b0c5f1fd99bfa30bad8d398e  genesis.json' | shasum -a 256 -c
# genesis.json: OK  <- this should be the output
```

## 2. Build the Node

**Binary** — from the `gno.land/` directory at the repository root:

```shell
make build.gnoland
```

**Docker image** — from the repository root:

```shell
docker build -t gnoland .
```

## 3. Initialize Secrets and Configure

Initialize your node secrets for **local signing**:

```shell
gnoland secrets init
```

For **remote signing** via `gnokms`, refer to the [gnokms documentation](../../../contribs/gnokms/README.md) for setup instructions.

Then use the provided `config.toml` as your base configuration. For most setups, editing only the fields marked with `# TODO` comments is sufficient.

## 4. Run Your Node

1. Start the node with the `--skip-genesis-sig-verification` flag (required due to a known incompatibility between genesis signatures and custom package metadata):

```shell
gnoland start --skip-genesis-sig-verification --genesis genesis.json
```

2. Ensure that **your RPC status page is reachable** so the team can verify your validator node's sync status. The following endpoint must be publicly accessible: `http(s)://<hostname-or-public-ip>:26657/status`

3. Once your node is fully synced, ping **Manfred on Signal** and we'll work on adding you as a validator through a GovDAO proposal.
