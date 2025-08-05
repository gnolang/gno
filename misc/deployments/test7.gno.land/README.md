# Overview

This deployment folder contains minimal information needed to launch a full test7.gno.land validator node.

Below is the information relating to the `genesis.json`, how to download it or generate it independently.

## Important Notice

Due to an issue on the `master` branch at the moment, custom package metadata deployers and genesis signatures are
incompatible. This means that if signature verification is turned on when starting the node, it will cause a panic.

Make sure to run the node using:

```shell
gnoland start --skip-genesis-sig-verification --genesis genesis.json
```

Another point is that the schema of the file `gnoland-data/secrets/node_key.json` has changed. If you are reusing one from a previous version, you can apply the following modification.

**Before:**

```json
{
    "priv_key": {
        "@type": "/tm.PrivKeyEd25519",
        "value": "RmQJFRawC+HnnzlZs8bvEFMPnnz9l8Fw8GpXonsnHEDipEA4N0Zekt/H8XSkEcb6FjWd5Ic13ZjefYJnsUCazg=="
    }
}
```

**After:**

```json
{
    "priv_key": "RmQJFRawC+HnnzlZs8bvEFMPnnz9l8Fw8GpXonsnHEDipEA4N0Zekt/H8XSkEcb6FjWd5Ic13ZjefYJnsUCazg=="
}
```

## `genesis.json`

The initial `genesis.json` validator set is consisted of 1 entity (4 validators in total):

- Gno Core - the gno core team (**4 validators**)

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
  `g137jz3hjhz6psrxxjtj5h7h4s6llfyrv2zxtfq3@gno-core-sen-01.test7.testnets.gno.land:26656,g1kpxll39mgzfhsepazzs0vne2l42mmkylxkt6un@gno-core-sen-02.test7.testnets.gno.land:26656`
  ** ⚠️.
- `p2p.seeds` - the bootnode peers. ⚠️ **Required to be
  `g137jz3hjhz6psrxxjtj5h7h4s6llfyrv2zxtfq3@gno-core-sen-01.test7.testnets.gno.land:26656,g1kpxll39mgzfhsepazzs0vne2l42mmkylxkt6un@gno-core-sen-02.test7.testnets.gno.land:26656`
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

## test7 `genesis.json` download

You can download the full `genesis.json` using the following steps:

```shell
wget -O genesis.json https://gno-testnets-genesis.s3.eu-central-1.amazonaws.com/test7/genesis.json
```

The `shasum` hash of the `genesis.json` should be `6ee51ee0e484d1deef777ef11de5b23144d7fd13d4cb1a8e7d9553d37a5cc34c`.
Verify it by running:

```sh
shasum -a 256 genesis.json
6ee51ee0e484d1deef777ef11de5b23144d7fd13d4cb1a8e7d9553d37a5cc34c  genesis.json
```

**NOTE**: Keep in mind that the generated genesis.json checksum will differ from the downloaded one,
because of a bug in `gnogenesis balances` that doesn't deterministically generate the genesis balance list:
https://github.com/gnolang/gno/issues/4122

---

## Generating the test7 `genesis.json`

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
wget -O genesis_balances.txt https://gno-testnets-genesis.s3.eu-central-1.amazonaws.com/test7/genesis_balances.txt
```

To verify the checksum of the genesis balances sheet:

```shell
shasum -a 256 genesis_balances.txt
c1efb09b3263e1a56a118fab6134a26892a4100b6b702191350371a9d06d03d5  genesis_balances.txt
```

The `genesis_txs.jsonl` can be fetched locally by:

```shell
wget -O genesis_txs.jsonl https://gno-testnets-genesis.s3.eu-central-1.amazonaws.com/test7/genesis_txs.jsonl
```

To verify the checksum of the genesis transaction sheet:

```shell
shasum -a 256 genesis_txs.jsonl
58953b590e5a837911ac13d13c332077fef3e379b5754a3063ea42b749a59052  genesis_txs.jsonl
```

### Reconstructing the genesis transactions

Test7 genesis transactions are generated from the `chain/test7` branch, and they are exclusively the `examples` deploy
transactions, with the state on that branch.

The deployer account for each test7 genesis transaction is derived from the mnemonic (`index 0`, `account 0`):

```shell
anchor hurt name seed oak spread anchor filter lesson shaft wasp home improve text behind toe segment lamp turn marriage female royal twice wealth
```

You can run the following steps to regenerate the `genesis_txs.jsonl`, from the root of the `chain/test7` branch

```shell
mkdir -p tmp-gnokey

gnokey add --recover Test7Deployer --home tmp-gnokey
gnogenesis generate -chain-id test7.2 -genesis-time 1753862400
gnogenesis txs add packages ./examples -gno-home tmp-gnokey -key-name Test7Deployer
gnogenesis txs export genesis_txs.jsonl

rm -rf tmp-gnokey
```
