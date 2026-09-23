# JSON-RPC endpoints

A gno.land node serves its JSON-RPC interface on port 26657 by default. This
page lists the endpoints it exposes, what each one takes, and what each one
returns.

The `vm/` JSON state-traversal paths reached through `abci_query` are
documented separately in
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

Both JSON-RPC transports also accept an array of request objects. HTTP answers
a batch carrying at least one request with an array; WebSocket answers a one-element batch with the bare
object. `params` takes either the object form above or a positional
array, and the array form requires every parameter the endpoint declares. Send
`"params":{}` rather than omitting the key for an endpoint whose arguments are
all optional.

A request to `/` with an empty body returns an HTML index of the endpoints
that node serves.

### Passing byte arguments

`hash`, `abci_query`'s `data` and the `tx` of the broadcast endpoints are byte
arrays. Over JSON-RPC each takes a base64 string. The URI transport is looser:
it also accepts the value unquoted, and a `0x`-prefixed hex string.

Two rules follow. A `+` in a base64 value has to be percent-encoded as `%2B`,
since a literal `+` means a space in a query string, and roughly half of all
hashes contain one. And the `0x` prefix is case-sensitive and URI-only: over
JSON-RPC the argument goes through Amino, which takes base64 alone.

Hex without the prefix is the trap worth naming. Hexadecimal characters are a
subset of the base64 alphabet, so 64 hex characters decode cleanly into 48
bytes of noise, so the node returns a `Could not find tx result` error rather
than rejecting the argument.

```bash
# the + percent-encoded, as it has to be
curl -s 'https://rpc.gno.land:443/tx?hash=%22SzaFJZgg%2BQIFOQ7MFsKdbKlKXkU5lBvW3y3MF2l673o%3D%22'
# {"jsonrpc":"2.0","id":"","result":{"hash":"SzaFJZgg+QIF...","height":"206563","index":0,...}}

# the same hash with a literal +, which the query string reads as a space
curl -s 'https://rpc.gno.land:443/tx?hash=%22SzaFJZgg+QIFOQ7MFsKdbKlKXkU5lBvW3y3MF2l673o%3D%22'
# {"jsonrpc":"2.0","id":"","error":{"code":-32602,"message":"Invalid params",
#                                    "data":"illegal base64 data at input byte 8"}}
```

## Node and network

What a node reports about itself: whether it is answering, how far it has
synced, who it is connected to, and the genesis it started from.

### `health`

Takes nothing. Returns an empty object. A response at all means the node is
answering.

### `status`

Takes an optional `heightGte`. Returns:

| Field | Contents |
|---|---|
| `node_info` | Moniker, network, versions, P2P address, channels, and `other` with the transaction index flag |
| `sync_info` | `latest_block_hash`, `latest_app_hash`, `latest_block_height`, `latest_block_time`, `catching_up` |
| `validator_info` | This node's validator `address`, `pub_key` and `voting_power` |
| `build_version` | The binary's build string |

`heightGte` turns the call into a readiness probe: the node answers 409 instead
of 200 when it has not reached the height asked for. That status appears only
on the URI transport; over JSON-RPC on HTTP the same case returns 200 with an
error object, and over WebSocket there is no HTTP status at all, only the
error in the response.

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

Chain state by height, plus the two consensus_state endpoints, which take
nothing and report the live round rather than a committed block. Where
`height` is optional it defaults to the latest block, except on `validators`
and `consensus_params`, which default one block further on.

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
and `canonical`.

`canonical` is `false` at the latest height, where the node has only the
commit it saw itself, and `true` below it, where the commit comes from the
block above. Calling without `height` therefore always answers `false`, which
is normal rather than a sign of anything wrong.

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

Sending a transaction, looking one up, and reading the mempool. A transaction
is found by exact hash only, and the three broadcast endpoints differ in how
much of the outcome they report back.

### `tx`

Takes `hash`. Returns `hash`, `height`, `index`, `tx_result` and the raw `tx`.

A transaction that is not found comes back as a JSON-RPC error, `-32603` with
`Could not find tx result for hash #<hex>`, not as a result with empty fields.
It may be in the mempool, may have been rejected, or may never have been sent:
the error is the same in all three cases. Code that polls `tx` until a
transaction appears has to treat that error as "not yet", not as a transport
failure.

### `broadcast_tx_sync`

Takes `tx`. Returns `error`, `data`, `log` and `hash` once the mempool has run
the application's `CheckTx`.

A rejected transaction still arrives as a successful JSON-RPC response, with
the reason at `result.error` — `{"@type": "/std.TxDecodeError"}` for an
undecodable one. Only the mempool's own refusals, a full mempool or a
transaction already in its cache among them, come back as a JSON-RPC error, so
there are three outcome shapes rather than two.

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
transaction is in a committed block. A transaction `CheckTx` rejects comes back
immediately instead, with an empty `deliver_tx` and a `height` of `"0"`.

The source marks it for testing and development rather than production. On
timeout it returns an error while the transaction may still commit later.

### `unconfirmed_txs`, `num_unconfirmed_txs`

The first takes an optional `limit` and returns `n_txs`, `total`, `total_bytes`
and `txs`.
`limit` defaults to 30 and is capped at 100. The second takes nothing and
returns the same shape with `txs` left null.

## Application

The two endpoints that reach past consensus into the application. Only
`abci_query` carries a path, and paths come in two shapes: the module routes
`auth/`, `bank/`, `vm/` and `params/`, and the two special prefixes `.app/`
and `.store/`. The `vm/` state-traversal endpoints — `qeval_json`, `qpkg_json`,
`qobject_json` and `qtype_json` — are documented in
[Querying On-Chain State](../builders/query-state-api.md).

### `abci_query`

Takes `path`, and optional `data`, `height` and `prove`. `height` defaults to
the chain tip. Returns `response`, holding `Key`, `Value`, `Proof`, `Height`
and a `ResponseBase` of `Error`, `Data`, `Events`, `Log` and `Info`. The module
paths — `auth/`, `bank/`, `vm/`, `params/` — return their result in
`ResponseBase.Data`, not at the top level.

A failed query still comes back as HTTP 200 with no top-level `error`; the
failure sits at `response.ResponseBase.Error`. A client that checks only the
top-level key reads it as a success carrying no data.

```bash
curl -s 'https://rpc.gno.land:443/abci_query?path=%22auth/accounts/g1manfred47kzduec920z88wfr64ylksmdcedlf5%22'
# "response": { "ResponseBase": { "Error": null, "Data": "ewogICJCYXNlQWNjb3VudCI6...",
#                                 "Events": null, "Log": "", "Info": "" },
#               "Key": null, "Value": null, "Proof": null, "Height": "0" }
```

The account is in `Data`, base64, two levels down. On the module paths `Key`
and `Value` stay null and `Height` stays `"0"`, whatever height was asked for;
the `.store/` paths set all three.

```bash
curl -s 'https://rpc.gno.land:443/abci_query?path=%22auth/accounts/g1manfred47kzduec920z88wfr64ylksmdcedlf5%22' \
  | jq -r '.result.response.ResponseBase.Data' | base64 -d
# {"BaseAccount":{"address":"g1manfred47kzduec920z88wfr64ylksmdcedlf5",...}}
```

`prove` returns real proof operations only on the store paths, which take their
key in `data`: `path=".store/main/key"&data="<base64 key>"`. The module paths —
`auth/`, `bank/`, `vm/`, `params/` — return `"Proof": null`. Asking for a proof
at height 1 or below is an error, and a `height` of `0` is rewritten to the
chain tip before that check.

### `abci_info`

Takes nothing. Returns `response`, with the last block height at
`LastBlockHeight` and the application name one level down, at
`ResponseBase.Data` — base64 like every other byte array, so `Z25vbGFuZA==`
rather than `gnoland`.

## Reading a response

Every result is encoded with Amino JSON. Three of its conventions have to be
implemented deliberately.

| Value | On the wire |
|---|---|
| Byte arrays, every hash included | Base64, standard alphabet, padded. A nil one is `null`, an empty one `""`. |
| `int64`, `uint64`, `int`, `uint` | Quoted strings, because JavaScript cannot hold them |
| 32-bit and narrower | Bare JSON numbers |

A single `tx` response therefore carries `"height": "51942"` next to
`"index": 0`. The same rule governs arguments: over JSON-RPC,
`{"height":"51942"}` is accepted where `{"height":51942}` is not.

Addresses are bech32 strings such as
`g1manfred47kzduec920z88wfr64ylksmdcedlf5`, not byte arrays. A transaction hash
is the untruncated `sha256` of the raw transaction bytes.

An HTTP 200 carries no meaning of its own, since successful results and
JSON-RPC errors both use it. One endpoint departs from that, `status` with
`heightGte` answering 409 on the URI transport. Other statuses come from the
server rather than from an endpoint: an unregistered path gives a plain-text
404, `/websocket` without a valid upgrade handshake gives a 400, and a handler
panic gives a 500 whose body is still a JSON-RPC error.

## Not available

| | |
|---|---|
| `tx_search`, `block_search` | Transactions are looked up by exact hash only. Run [tx-indexer](https://github.com/gnolang/tx-indexer) for anything else. |
| `subscribe`, `unsubscribe` | No event stream. `/websocket` serves the same request-response endpoints as HTTP. |
| `?page` / `?per_page` | No endpoint paginates; both are accepted and ignored. |
| A gRPC server | `grpc_laddr` appears in generated config files, and nothing starts a server from it. |
| `dial_seeds`, `dial_persistent_peers` | Removed. |
| Hex output | Byte arrays go out as base64, never as hex. |

Four `unsafe_*` endpoints — a mempool flush and three pprof profilers — are
registered only when a node runs with `rpc.unsafe = true`, which public nodes
do not. They are not safe to expose: two of them write to a file path the
caller chooses. An unregistered endpoint answers 404 on the URI transport and
`-32601` over the other two.

## See also

- [Querying On-Chain State](../builders/query-state-api.md) — the `vm/` JSON
  state-traversal paths carried over `abci_query`
- [RPC clients](../builders/rpc-clients.md) — client libraries
- [Gno networks](./gnoland-networks.md) — node addresses and chain IDs
