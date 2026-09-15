# JSON-RPC endpoints

A gno.land node serves its JSON-RPC interface on port 26657 by default. This
page lists the endpoints it exposes, what each one takes, and what each one
returns.

The ABCI query paths reached through `abci_query` are documented separately in
[Querying On-Chain State](../builders/query-state-api.md). Node addresses and
chain IDs are in [Gno networks](./gnoland-networks.md).

## Calling an endpoint

Three transports serve the same endpoints.

| Transport | Form |
|---|---|
| URI over HTTP | `GET /<method>?<arg>=<value>` |
| JSON-RPC over HTTP | `POST /` with a request object |
| WebSocket | `/websocket` |

```bash
curl -s 'https://rpc.gno.land:443/block?height=51942'

curl -s -X POST https://rpc.gno.land:443/ \
  -d '{"jsonrpc":"2.0","id":1,"method":"block","params":{"height":"51942"}}'
```

Both JSON-RPC transports also accept an array of request objects and answer
with an array. `params` takes either the object form above or a positional
array, and the array form requires every parameter the endpoint declares. Send
`"params":{}` rather than omitting the key for an endpoint whose arguments are
all optional.

A request to `/` with an empty body returns an HTML index of the endpoints
that node serves.

## Node and network

### `health`

Takes nothing. Returns an empty object. A response at all means the node is
answering.

### `status`

Takes an optional `heightGte`. Returns:

| Field | Contents |
|---|---|
| `node_info` | Moniker, network, software version, P2P address, channels |
| `sync_info` | `latest_block_hash`, `latest_app_hash`, `latest_block_height`, `latest_block_time`, `catching_up` |
| `validator_info` | This node's validator `address`, `pub_key` and `voting_power` |
| `build_version` | The binary's build string |

`heightGte` turns the call into a readiness probe: the node answers 409 instead
of 200 when it has not reached the height asked for. That status appears only
on the URI transport; over JSON-RPC the same case returns 200 with an error
object.

### `net_info`

Takes nothing. Returns `listening`, `listeners`, `n_peers` and `peers`, each
peer carrying its node info, whether the connection is outbound, its connection
status and its remote IP.

### `genesis`

Takes nothing. Returns the genesis document under `genesis`.

The response is streamed rather than buffered, and on a genesis the size of
gno.land's it does not complete: the write deadline is 30 seconds over HTTP and 10 over
WebSocket, set once for the whole body. Over HTTP the transfer ends at the
deadline with a 200 status and truncated JSON; over WebSocket the frames flow
but the message is never finalised. A batch rejects it outright. There is no
chunked variant.

## Blocks

### `blockchain`

Takes optional `minHeight` and `maxHeight`. Returns `last_height` and
`block_metas`, highest first.

At most 20 entries come back. A wider range is cut at the low end, with no
error and nothing in the response marking it as partial.

### `block`

Takes an optional `height`, defaulting to the latest. Returns `block_meta` and
`block`. The block id lives under `block_meta`; the header appears under both.

### `block_results`

Takes an optional `height`, defaulting to the latest. Returns `height` and
`results`, which holds `deliver_tx` — one entry per transaction, positionally
matching `block.data.txs` at the same height — alongside `begin_block` and
`end_block`.

A `height` of `0` is valid here and addresses the genesis transactions, which
an indexer starting at height 1 never sees.

### `commit`

Takes an optional `height`, defaulting to the latest. Returns `signed_header`
and `canonical`, the latter reporting whether the commit is the canonical one
for that block.

### `validators`

Takes an optional `height`. Returns `block_height` and `validators`.

Omitting `height` returns the set for the height *after* the last committed
one, so omitting it and passing the chain tip explicitly give different answers
across a validator change.

### `consensus_params`

Takes an optional `height`, with the same default as `validators`. Returns
`block_height` and `consensus_params`.

### `consensus_state`, `dump_consensus_state`

Take nothing. The first returns a summary of the current round state, the
second the full round state with the node's consensus config and per-peer
state. Both are marked unstable in the source and are aimed at operators rather
than integrators.

## Transactions

### `tx`

Takes `hash`. Returns `hash`, `height`, `index`, `tx_result` and the raw `tx`.

A transaction that is not found may be in the mempool, may have been rejected,
or may never have been sent. The response does not distinguish them.

### `broadcast_tx_sync`

Takes `tx`. Returns `error`, `data`, `log` and `hash` once the mempool has run
the application's `CheckTx`.

`CheckTx` validates a transaction for the mempool; it does not run its
messages. An accepted transaction can still fail when it executes, and that
outcome appears only in `tx_result`, once the transaction is in a block.

### `broadcast_tx_async`

Takes `tx`. Returns the same shape, but without the `CheckTx` result: the node
runs it and the mempool acts on the outcome, while the caller is told only that
the request was accepted. A transaction the application rejects is reported
here as a success.

### `broadcast_tx_commit`

Takes `tx`. Returns `check_tx`, `deliver_tx`, `hash` and `height` once the
transaction is in a committed block.

The source marks it for testing and development rather than production. On
timeout it returns an error while the transaction may still commit later.

### `unconfirmed_txs`, `num_unconfirmed_txs`

The first takes `limit` and returns `n_txs`, `total`, `total_bytes` and `txs`.
`limit` defaults to 30 and is capped at 100. The second takes nothing and
returns the same shape with `txs` left null.

## Application

### `abci_query`

Takes `path`, `data`, `height` and `prove`. Returns `response`, holding `Key`,
`Value`, `Proof`, `Height` and a `ResponseBase` of `Error`, `Data`, `Log` and
`Info`. The module paths — `auth/`, `bank/`, `vm/`, `params/` — return their
result in `ResponseBase.Data`, not at the top level.

A failed query still comes back as HTTP 200 with no top-level `error`; the
failure sits at `response.ResponseBase.Error`. A client that checks only the
top-level key reads it as a success carrying no data.

`prove` returns real proof operations only on the store paths, which take their
key in `data`: `path=".store/main/key"&data="<base64 key>"`. The module paths —
`auth/`, `bank/`, `vm/`, `params/` — return `"Proof": null`. Asking for a proof
at height 1 or below is an error, and a `height` of `0` is rewritten to the
chain tip before that check.

### `abci_info`

Takes nothing. Returns `response`, holding the application name in `Data` and
the last block height.

## Reading a response

Every result is encoded with Amino JSON. Three of its conventions have to be
implemented deliberately.

| Value | On the wire |
|---|---|
| Byte arrays, every hash included | Base64, standard alphabet, padded. A nil one is `null`, an empty one `""`. |
| `int64`, `uint64`, `int`, `uint` | Quoted strings, because JavaScript cannot hold them |
| `int32` and narrower | Bare JSON numbers |

A single `tx` response therefore carries `"height": "51942"` next to
`"index": 0`. The same rule governs arguments: over JSON-RPC,
`{"height":"51942"}` is accepted where `{"height":51942}` is not.

Addresses are bech32 strings such as
`g1manfred47kzduec920z88wfr64ylksmdcedlf5`, not byte arrays. A transaction hash
is the untruncated `sha256` of the raw transaction bytes.

An HTTP 200 carries no meaning of its own, since successful results and
JSON-RPC errors both use it. The exceptions are `status`'s 409 on the URI
transport, an unregistered path giving a plain-text 404, and a handler panic
giving a 500 whose body is still a JSON-RPC error.

### Passing byte arguments

`hash`, `abci_query`'s `data` and the `tx` of the broadcast endpoints are byte
arrays. Each accepts base64, quoted or not, and on the URI transport a
`0x`-prefixed hex string.

Two rules follow. A `+` in a base64 value has to be percent-encoded as `%2B`,
since a literal `+` means a space in a query string, and roughly half of all
hashes contain one. And the `0x` prefix is case-sensitive and URI-only: over
JSON-RPC the argument goes through Amino, which takes base64 alone.

Hex without the prefix is the trap worth naming. Hexadecimal characters are a
subset of the base64 alphabet, so 64 hex characters decode cleanly into 48
bytes of noise and the node answers "not found" rather than rejecting the
argument.

## Not available

| | |
|---|---|
| `tx_search`, `block_search` | Transactions are looked up by exact hash only. Run [tx-indexer](https://github.com/gnolang/tx-indexer) for anything else. |
| `subscribe`, `unsubscribe` | No event stream. `/websocket` serves the same request-response endpoints as HTTP. |
| `?page` / `?per_page` | No endpoint paginates; both are accepted and ignored. |
| A gRPC server | `grpc_laddr` appears in generated config files and nothing reads it. |
| `dial_seeds`, `dial_persistent_peers` | Removed. |
| Hex output | Every byte value is base64. |

Four `unsafe_*` endpoints, a mempool flush and the pprof profilers, are
registered only when a node runs with `rpc.unsafe = true`, which public nodes
do not. They are not safe to expose: two of them write to a file path the
caller chooses. An unregistered endpoint answers 404 on the URI transport and
`-32601` over the other two.

## See also

- [Querying On-Chain State](../builders/query-state-api.md) — the ABCI paths
  carried over `abci_query`
- [RPC clients](../builders/rpc-clients.md) — client libraries
- [Gno networks](./gnoland-networks.md) — node addresses and chain IDs
