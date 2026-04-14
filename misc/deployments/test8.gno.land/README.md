# Overview

This deployment folder contains minimal information needed to launch a full test8.gno.land validator node.

Below is the information relating to the `genesis.json`, how to download it or generate it independently.

## Important Notice

Due to an issue on the `master` branch at the moment, custom package metadata deployers and genesis signatures are
incompatible. This means that if signature verification is turned on when starting the node, it will cause a panic.

Make sure to run the node using:

```shell
gnoland start --skip-genesis-sig-verification --genesis genesis.json
```

## `genesis.json`

The initial `genesis.json` validator set is consisted of 2 entities (4 validators in total):

- Gno Core - the gno core team (**2 validators**)
- Onbloc (**2 validators**)

Subsequent validators will be added through the governance mechanism in govdao.

**The premined genesis balances are for testing purposes only, and in no way reflect the actual premine balances
in future network deployments.**

## `config.toml`

The `config.toml` located in this directory is a **_guideline_**, and not a definitive configuration on how
all nodes should be configured in the network.

### Important params

Some configuration params are required, while others are advised to be set.

- `moniker` - the recognizable identifier of the node.
- `application.prune_strategy` - the prune strategy. ⚠️ **Required to be `syncable`
- `consensus.timeout_commit` - the timeout value after the consensus commit phase. ⚠️ **Required to be `3s`** ⚠️.
- `consensus.peer_gossip_sleep_duration` - the timeout for peer gossip. ⚠️ **Required to be `10ms`** ⚠️.
- `mempool.size` - the maximum number of txs in the mempool. **Advised to be `10000`**.
- `p2p.laddr` - the listen address for P2P traffic, **specific to every node deployment**. It is advised to use a
  reverse-proxy, and keep this value at `tcp://0.0.0.0:<port>`.
- `p2p.max_num_outbound_peers` - the max number of outbound peer connections. **Advised to be `40`**.
- `p2p.persistent_peers` - the persistent peers. ⚠️ **Required to be
  `g195079rf97efj5enrkfh5pjeg23uzqcex64jnes@gno-core-sen-01.test8.testnets.gno.land:26656`
  ** ⚠️.
- `p2p.seeds` - the bootnode peers. ⚠️ **Required to be
  `g195079rf97efj5enrkfh5pjeg23uzqcex64jnes@gno-core-sen-01.test8.testnets.gno.land:26656`
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

## test8 `genesis.json` download

You can download the full `genesis.json` using the following steps:

```shell
wget -O genesis.json https://gno-testnets-genesis.s3.eu-central-1.amazonaws.com/test8/genesis.json
```

The `shasum` hash of the `genesis.json` should be `75b23ea9069a528a31f34fecf7d1a027969c076f945f0cfb3189349e4070044e`.
Verify it by running:

```sh
shasum -a 256 genesis.json
75b23ea9069a528a31f34fecf7d1a027969c076f945f0cfb3189349e4070044e  genesis.json
```

---

## Generating the test8 `genesis.json`

To generate the `genesis.json`, you will need the `gnogenesis` tool from this branch.

### Step 1: Install `gnogenesis` globally

From the repo root:

```shell
cd contribs/gnogenesis
make install
```

### Step 2: Run the generation script

The required `genesis.json` generation logic is packaged up in `generate.sh`.
To run it, make sure to adjust correct permissions:

```shell
chmod +x ./generate.sh
```

Run the script, and it should generate a `genesis.json` locally:

```shell
./generate.sh
```

---

### genesis.json Artifacts

The `genesis_balances.txt` can be fetched locally by:

```shell
wget -O genesis_balances.txt https://gno-testnets-genesis.s3.eu-central-1.amazonaws.com/test8/genesis_balances.txt
```

To verify the checksum of the genesis balances sheet:

```shell
shasum -a 256 genesis_balances.txt
7350080e52deccef31dc8b576b10fc183adab17f7f4ab618bacb2c6ffb044fc3  genesis_balances.txt
```

The `genesis_txs.jsonl` can be fetched locally by:

```shell
wget -O genesis_txs.jsonl https://gno-testnets-genesis.s3.eu-central-1.amazonaws.com/test8/genesis_txs.jsonl
```

To verify the checksum of the genesis transaction sheet:

```shell
shasum -a 256 genesis_txs.jsonl
81e82ad1ec850974b3a29bffbc92b67b9e8c2c3d135f88a55310288814f2cb86  genesis_txs.jsonl
```

### Reconstructing the genesis transactions

Test8 genesis transactions are generated from the `chain/test8` branch, and they are exclusively the `examples` deploy
transactions, with the state on that branch.

The deployer account for each test8 genesis transaction is derived from the mnemonic (`index 0`, `account 0`):

```shell
anchor hurt name seed oak spread anchor filter lesson shaft wasp home improve text behind toe segment lamp turn marriage female royal twice wealth
```

You can run the following steps to regenerate the `genesis_txs.jsonl`, from the root of the `chain/test8` branch

```shell
mkdir -p tmp-gnokey

gnokey add --recover Test8Deployer --home tmp-gnokey
gnogenesis generate -chain-id test8 -genesis-time 1757055600
gnogenesis txs add packages ./examples -gno-home tmp-gnokey -key-name Test8Deployer
gnogenesis txs export genesis_txs.jsonl

rm -rf tmp-gnokey
```
