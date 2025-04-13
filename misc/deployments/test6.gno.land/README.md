# Overview

This deployment folder contains minimal information needed to launch a full test6.gno.land validator node.

Below is the information relating to the `genesis.json`, how to download it or generate it independently.

## `genesis.json`

The initial `genesis.json` validator set is consisted of 4 entities (8 validators in total):

- Gno Core - the gno core team (**2 validators**)
- Gno DevX - the gno devX team (**2 validators**)
- AiB - the AiB DevOps team (**2 validators**)
- Onbloc - the [Onbloc](https://onbloc.xyz/) team (**2 validators**)

Subsequent validators will be added through the governance mechanism in govdao.

**The premined genesis balances are for testing purposes only, and in no way reflect the actual premine balances
in future network deployments.**

## `config.toml`

The `config.toml` located in this directory is a **_guideline_**, and not a definitive configuration on how
all nodes should be configured in the network.

### Important params

Some configuration params are required, while others are advised to be set.

- `moniker` - the recognizable identifier of the node.
- `consensus.timeout_commit` - the timeout value after the consensus commit phase. ⚠️ **Required to be `3s`** ⚠️.
- `consensus.peer_gossip_sleep_duration` - the timeout for peer gossip. ⚠️ **Required to be `10ms`** ⚠️.
- `mempool.size` - the maximum number of txs in the mempool. **Advised to be `10000`**.
- `p2p.laddr` - the listen address for P2P traffic, **specific to every node deployment**. It is advised to use a
  reverse-proxy, and keep this value at `tcp://0.0.0.0:<port>`.
- `p2p.max_num_outbound_peers` - the max number of outbound peer connections. **Advised to be `40`**.
- `p2p.persistent_peers` - the persistent peers. ⚠️ **Required to be
  `g16384atcuf6ew3ufpwtvhymwfyl2aw390aq8jtt@gno-core-sen-01.test6.testnets.gno.land:26656,g16384atcuf6ew3ufpwtvhymwfyl2aw390aq8jtt@gno-core-sen-02.test6.testnets.gno.land:26656`
  ** ⚠️.
- `p2p.seeds` - the bootnode peers. ⚠️ **Required to be
  `g16384atcuf6ew3ufpwtvhymwfyl2aw390aq8jtt@gno-core-sen-01.test6.testnets.gno.land:26656,g16384atcuf6ew3ufpwtvhymwfyl2aw390aq8jtt@gno-core-sen-02.test6.testnets.gno.land:26656`
  ** ⚠️.
- `p2p.pex` - if using a sentry node architecture, should be `false`. **If not, please set to `true`**.
- `p2p.external_address` - the advertised peer dial address. If empty, will use the same port as the `p2p.laddr`. This
  value should be **changed to `{{ your_ip_address }}:26656`**
- `p2p.flush_throttle_timeout` - the timeout for flushing multiplex data. ⚠️ **Required to be `10ms`** ⚠️.
- `rpc.laddr` - the JSON-RPC listen address, **specific to every node deployment**.
- `telemetry.enabled` - flag indicating if telemetry should be turned on. **Advised to be `true`**.
- `telemetry.exporter_endpoint` - endpoint for the otel exported. ⚠️ **Required if `telemetry.enabled=true`** ⚠️.
- `telemetry.service_instance_id` - unique ID of the node telemetry instance, **specific to every node deployment**.

---

## test6 `genesis.json` download

You can download the full `genesis.json` using the following steps:

```shell
wget TODO
```

The `shasum` hash of the `genesis.json` should be `TODO`.
Verify it by running:

```sh
shasum -a 256 genesis.json
TODO genesis.json
```

The `genesis_balances.txt` can be fetched locally by:

```shell
wget -O genesis_balances.txt https://gno-testnets-genesis.s3.eu-central-1.amazonaws.com/test6/genesis_balances.txt
```

To verify the checksum of the genesis balances sheet:

```shell
shasum -a 256 genesis_balances.txt
3c265896283fa144f6cc5c7db7fd0583d762a075d9ea06ccb98084f73888429d  genesis_balances.txt
```

The `genesis_txs.jsonl` can be fetched locally by:

```shell
wget -O genesis_txs.jsonl https://gno-testnets-genesis.s3.eu-central-1.amazonaws.com/test6/genesis_txs.jsonl
```

To verify the checksum of the genesis transaction sheet:

```shell
shasum -a 256 genesis_txs.jsonl
85341afa5f273ae7837ef8217db320da8869debf8399675c66b246c6ca807c22  genesis_txs.jsonl
```

### Reconstructing the genesis transactions

Test6 genesis transactions are generated from the `chain/test6` branch, and they are exclusively the `examples` deploy
transactions, with the state on that branch.

The deployer account for each test6 genesis transaction is derived from the mnemonic (`index 0`, `account 0`):

```shell
anchor hurt name seed oak spread anchor filter lesson shaft wasp home improve text behind toe segment lamp turn marriage female royal twice wealth
```

You can run the following steps to regenerate the `genesis_txs.jsonl`, from the root of the `chain/test6` branch

```shell
# Use a temporary directory for intermediary states
TMP_DIR=./tmp-gnokey

printf "\nAdding txs (this may take a while). Ignore the prompts that come up...\n"

# Add the deployer (genesis txs) key into gnokey
DEPLOYER_MNEMONIC="anchor hurt name seed oak spread anchor filter lesson shaft wasp home improve text behind toe segment lamp turn marriage female royal twice wealth"
DEPLOYER_NAME=Test6Deployer

echo "$DEPLOYER_MNEMONIC" | gnokey add --recover "$DEPLOYER_NAME" --home "$TMP_DIR" -insecure-password-stdin -quiet

# Add the transactions (all examples).
# Test6Deployer is the deployer key for all genesis transactions, and
# it has an adequate premine amount in the balances already
gnogenesis generate -chain-id test6
echo "" | gnogenesis txs add packages ./examples -gno-home "$TMP_DIR" -insecure-password-stdin -key-name "$DEPLOYER_NAME"

# Cleanup
rm -rf $TMP_DIR
```

---

## Generating the test6 `genesis.json`

To generate the `genesis.json`, you will need the `gnogenesis` tool from this branch.

### Step 1: Install `gnogenesis` globally

From the repo root:

```shell
cd contribs/gnogenesis
make install
```

### Step 2: Install `gnokey` globally

From the repo root:

```shell
make install.gnokey
```

### Step 3: Run the generation script

The required `genesis.json` generation logic is packaged up in `generate.sh`.
To run it, make sure to adjust correct permissions:

```shell
chmod +x ./generate.sh
```

Run the script, and it should generate a `genesis.json` locally. We also opt to verify the genesis sum:

```shell
CHECKSUM=TODO ./generate.sh

TODO output
```
