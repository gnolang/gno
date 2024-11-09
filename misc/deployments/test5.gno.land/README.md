# Overview

This deployment folder contains minimal information needed to launch a full test5.gno.land validator node.

## `genesis.json`

The initial `genesis.json` validator set is consisted of 6 entities (17 validators in total):

- Gno Core - the gno core team (**6 validators**)
- Gno DevX - the gno devX team (**4 validators**)
- AiB - the AiB DevOps team (**3 validators**)
- Onbloc - the [Onbloc](https://onbloc.xyz/) team (**2 validator**)
- Teritori - the [Teritori](https://teritori.com/) team (**1 validator**)
- Berty - the [Berty](https://berty.io/) team (**1 validator**)

Subsequent validators will be added through the governance mechanism in govdao, employing a preliminary simplified
version Proof of Contribution.

The addresses premined belong to different faucet accounts, validator entities and implementation partners.

## `config.toml`

The `config.toml` located in this directory is a **_guideline_**, and not a definitive configuration on how
all nodes should be configured in the network.

### Important params

Some configuration params are required, while others are advised to be set.

- `moniker` - the recognizable identifier of the node.
- `consensus.timeout_commit` - the timeout value after the consensus commit phase. ⚠️ **Required to be `1s`** ⚠️.
- `conseuns.peer_gossip_sleep_duration` - the timeout for peer gossip. ⚠️ **Required to be `10ms`** ⚠️.
- `mempool.size` - the maximum number of txs in the mempool. **Advised to be `10000`**.
- `p2p.laddr` - the listen address for P2P traffic, **specific to every node deployment**. It is advised to use a
  reverse-proxy, and keep this value at `tcp://0.0.0.0:<port>`.
- `p2p.max_num_outbound_peers` - the max number of outbound peer connections. **Advised to be `40`**.
- `p2p.persistent_peers` - the persistent peers. ⚠️ **Required to be
  `TODO`** ⚠️.
- `p2p.pex` - if using a sentry node architecture, should be `false`. **If not, please set to `true`**.
- `p2p.external_address` - the advertised peer dial address. If empty, will use the same port as the `p2p.laddr`. This
  value should be **changed to `{{ your_ip_address }}:26656`**
- `p2p.flush_throttle_timeout` - the timeout for flushing multiplex data. ⚠️ **Required to be `10ms`** ⚠️.
- `rpc.laddr` - the JSON-RPC listen address, **specific to every node deployment**.
- `telemetry.enabled` - flag indicating if telemetry should be turned on. **Advised to be `true`**.
- `telemetry.exporter_endpoint` - endpoint for the otel exported. ⚠️ **Required if `telemetry.enabled=true`** ⚠️.
- `telemetry.service_instance_id` - unique ID of the node telemetry instance, **specific to every node deployment**.

## Gno Core DNS entries

- CNAME test5.gno.land a78095e89431e41f2b506544efd9f79e-dd80971e7e1cf6bc.elb.eu-central-1.amazonaws.com
- CNAME rpc.test5.gno.land test5.gno.land
- CNAME gno-core-sen-01.test5.gno.land test5.gno.land
- CNAME gno-core-sen-02.test5.gno.land test5.gno.land
- CNAME gno-core-sen-03.test5.gno.land test5.gno.land
- CNAME indexer.test5.gno.land test5.gno.land
- CNAME gnofaucet.test5.gno.land test5.gno.land
- CNAME o11y.test5.gno.land test5.gno.land (Grafana)
