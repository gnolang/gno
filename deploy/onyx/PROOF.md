# Proof transcript

Everything below was run on lyra on 2026-11-23 against a `gnoland` node synced
from the onyx genesis and a `tx-indexer` reading it over loopback. Nothing here
touched a public endpoint.

## 1. The node is on the real onyx-1, not a plausible lookalike

`chain_id` alone proves nothing: a reset chain keeps it. The block 1 hash is
what distinguishes them, and it is derived from genesis, so a node that
reproduces it replayed the same genesis the network did.

```
$ curl -fsS "http://127.0.0.1:26757/block?height=1" | jq -r '.result.block_meta.block_id.hash'
PGlpc0OZKeAItjmtZzNVJSzsmY40K9X7XAMCuTa6wj4=

$ curl -fsS https://indexer.onyx.testnets.gno.land/graphql/query -X POST \
    -H 'Content-Type: application/json' \
    --data '{"query":"{ getBlocks(where: { height: { eq: 1 } }) { chain_id hash } }"}'
{"data":{"getBlocks":[{"chain_id":"onyx-1","hash":"PGlpc0OZKeAItjmtZzNVJSzsmY40K9X7XAMCuTa6wj4="}]}}
```

Identical. Same for `chain_id`, `onyx-1`.

## 2. gnoscope's endpoint-pool probe, against the local indexer

This is the exact query from
[`pkg/indexer/indexer.go`](https://github.com/gnoverse/gnoscope/blob/master/pkg/indexer/indexer.go),
`func (c *Client) probe`. It is how gnoscope decides whether an endpoint may
join the onyx pool at all.

```
$ curl -fsS -X POST http://127.0.0.1:8567/graphql/query \
    -H 'Content-Type: application/json' \
    --data '{"query":"{ latestBlockHeight getBlocks(where: { height: { eq: 1 } }) { chain_id hash } }"}'
{
    "data": {
        "latestBlockHeight": 46100,
        "getBlocks": [
            {
                "chain_id": "onyx-1",
                "hash": "PGlpc0OZKeAItjmtZzNVJSzsmY40K9X7XAMCuTa6wj4="
            }
        ]
    }
}
```

Fingerprint `onyx-1:PGlpc0OZKeAItjmtZzNVJSzsmY40K9X7XAMCuTa6wj4=`, matching the
public pool members. gnoscope will accept this endpoint.

## 3. gnoscope's block list query

`blockFields` verbatim from the same file.

```
$ curl -fsS -X POST http://127.0.0.1:8567/graphql/query \
    -H 'Content-Type: application/json' \
    --data '{"query":"{ getBlocks(where: { height: { gt: 40000, lt: 40006 } } order: { height: ASC }) { hash height chain_id time num_txs total_txs proposer_address_raw } }"}'
{
    "data": {
        "getBlocks": [
            {
                "hash": "mImygtNm+aTWy1mF+ji0oasIYqmAI7qzBnocJbP05+w=",
                "height": 40001,
                "chain_id": "onyx-1",
                "time": "2026-09-29T13:30:30.621168545Z",
                "num_txs": 0,
                "total_txs": 118,
                "proposer_address_raw": "g19429enajcectdlhwulwza8x2h6mkygu0hpt6y4"
            },
            ... 40002, 40003, 40005 elided ...
            {
                "hash": "EDblzZEZiVlFGUVziw9MTbku7xlzkltBkLdPdmCRWjk=",
                "height": 40004,
                "chain_id": "onyx-1",
                "time": "2026-09-29T13:30:40.743320284Z",
                "num_txs": 1,
                "total_txs": 119,
                "proposer_address_raw": "g19429enajcectdlhwulwza8x2h6mkygu0hpt6y4"
            }
        ]
    }
}
```

## 4. gnoscope's transaction query, including the schema gating

Sent verbatim, gnoscope's `txFieldsLight` selection **422s** against this
indexer:

```
{"errors":[
 {"message":"Unknown type \"MsgCreateSession\"."},
 {"message":"Unknown type \"MsgRevokeSession\"."},
 {"message":"Unknown type \"MsgRevokeAllSessions\"."},
 {"message":"Cannot query field \"signatures\" on type \"Transaction\"."}]}
```

This is **not** a defect in the local stack. It is the case gnoscope already
handles: `supportsTypeAt` probes one representative type per fragment group and
strips the groups the endpoint's schema lacks. tx-indexer `v1.5.1` defines no
session messages and no `Signature` type, on any chain, which is also why
gnoscope's own API docs say session grants have to be read live from RPC.

The four probes gnoscope sends, against the local indexer:

```
MsgEnablePackage     {"data":{"__type":{"name":"MsgEnablePackage"}}}
MsgCreateSession     {"data":{"__type":null}}
TransferEvent        {"data":{"__type":{"name":"TransferEvent"}}}
Signature            {"data":{"__type":null}}
```

With the two unsupported groups stripped, exactly as gnoscope does it, the full
selection answers:

```
$ curl -fsS -X POST http://127.0.0.1:8567/graphql/query -H 'Content-Type: application/json' \
    --data @gnoscope-transactions.json
{
    "data": {
        "getTransactions": [
            {
                "hash": "jb9A+sEJG3DAOZrW3aenuQKp5TZy31b/xY91fX2BqIc=",
                "success": true,
                "block_height": 40004,
                "gas_wanted": 80000000,
                "gas_used": 46736672,
                "gas_fee": { "amount": 1000000, "denom": "ugnot" },
                "memo": "",
                "messages": [
                    {
                        "typeUrl": "exec",
                        "route": "vm",
                        "value": {
                            "__typename": "MsgCall",
                            "caller": "g14ymnmcklzc8zlr7q5hzc370jv433ltx0aq90np",
                            "send": "",
                            "pkg_path": "gno.land/r/gnops/valopers",
                            "func": "Register",
                            "args": [
                                "BZE Community",
                                "BeeZee Blockchain community validator. ...",
                                "cloud",
                                "g14ymnmcklzc8zlr7q5hzc370jv433ltx0aq90np",
                                "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zptlru50g4dv9j5zuy5n9vx5nw4fewrmn58anyk7rwvyw74srtygpgja0ff"
                            ]
                        }
                    }
                ],
                "response": {
                    "log": "msg:0,success:true,log:,events:[]",
                    "info": "",
                    "error": "",
                    "data": "\n\n",
                    "events": [
                        {
                            "__typename": "StorageDepositEvent",
                            "type": "StorageDepositEvent",
                            "bytes_delta": 11924,
                            "fee_delta": { "amount": 1192400, "denom": "ugnot" },
                            "pkg_path": "gno.land/r/gnops/valopers"
                        },
                        {
                            "__typename": "StorageDepositEvent",
                            "type": "StorageDepositEvent",
                            "bytes_delta": 1046,
                            "fee_delta": { "amount": 104600, "denom": "ugnot" },
                            "pkg_path": "gno.land/r/sys/validators/v0"
                        }
                    ]
                }
            }
        ]
    }
}
```

A real onyx transaction, with its messages decoded into the right concrete type
and its storage-deposit events intact.

## 5. The RPC half: account reads

gnoscope reads balances and session grants from RPC, not from the indexer
(`abci_query path="auth/accounts/<addr>"`). Against the local node:

```
$ curl -fsS -G http://127.0.0.1:26757/abci_query \
    --data-urlencode 'path="auth/accounts/g14ymnmcklzc8zlr7q5hzc370jv433ltx0aq90np"' \
    --data-urlencode 'data=""'
{
  "BaseAccount": {
    "address": "g14ymnmcklzc8zlr7q5hzc370jv433ltx0aq90np",
    "coins": "87519700ugnot",
    "public_key": { "@type": "/tm.PubKeySecp256k1", "value": "AsglTdfYN+xIEEflmWTinORO1s8GiDzmjQ6KgU2tADMp" },
    "account_number": "157",
    "sequence": "2"
  },
  "attributes": "0"
}
```

## 6. Loopback-only p2p still syncs

`p2p.laddr = tcp://127.0.0.1:26756`. The node discovered and gossiped with more
than a dozen peers through PEX and replayed from genesis without interruption.
The listener only accepts inbound connections, which this node does not want;
outbound dialing is unaffected. The cost is that it never serves blocks back,
which is acceptable for one explorer backend and would not be for a fleet.

## 7. The same queries, against a node and indexer at the live tip

Sections 1 to 5 ran against a node replaying history. This section re-ran them
after a snapshot restore, with everything caught up.

**All three heights agree**, and the public RPC was asked exactly once:

```
local node        1350786
local indexer     1350786
rpc.onyx.…        1350786
```

`catching_up: false`, block time `2026-11-23T11:48:23Z`, seconds behind wall
clock.

**The fingerprint still matches** at the tip:

```
{"data":{"latestBlockHeight":1350786,
         "getBlocks":[{"chain_id":"onyx-1",
                       "hash":"PGlpc0OZKeAItjmtZzNVJSzsmY40K9X7XAMCuTa6wj4="}]}}
```

**gnoscope's block list, at the tip:**

```
height 1350781  E2fdsCic6zABxoDJVWsfAW8Qk4TtRVzWGnbpIO9KPGM=  2026-11-23T11:48:06Z  num_txs 0  total_txs 12876
height 1350782  1Lf0ZQPe47RxE+M4wcckf8QQxCLrk97a4hs17kt+nys=  2026-11-23T11:48:09Z  num_txs 0  total_txs 12876
height 1350783  MS79BQ8f9aEdbIHVsoBaiDd6Bkq7gaaSe3FebG52uIM=  2026-11-23T11:48:13Z  num_txs 0  total_txs 12876
height 1350784  n9ZWQyH+40gxtVhG9fGrG0RlNB1G23wV743GteBDYxg=  2026-11-23T11:48:16Z  num_txs 0  total_txs 12876
```

**gnoscope's transaction query over the last ~10,700 blocks** returned 211
rows, heights 1,340,054 to 1,350,731, every message decoded into a concrete
type and **no `UnexpectedMessage` among them**:

```
MsgCall 146 | MsgAddPackage 36 | BankMsgSend 27 | MsgEnablePackage 10
```

One row in full, a realm call with its `GnoEvent` attributes intact:

```json
{
  "hash": "p8irAKQqI1HATwxkGarSAjgzPEuWxzRwBxYq4he3K6o=",
  "success": true,
  "block_height": 1350731,
  "gas_wanted": 4064556,
  "gas_used": 3126642,
  "gas_fee": { "amount": 8130, "denom": "ugnot" },
  "memo": "GnoPulse access quote",
  "messages": [{
    "typeUrl": "exec", "route": "vm",
    "value": {
      "__typename": "MsgCall",
      "caller": "g14wv9hyngyrzpxn3ztvnx2jygs50m0240qncm8j",
      "send": "",
      "pkg_path": "gno.land/r/g14wv9hyngyrzpxn3ztvnx2jygs50m0240qncm8j/access",
      "func": "SetQuote",
      "args": ["422749053"]
    }
  }],
  "response": {
    "log": "msg:0,success:true,log:,events:[]",
    "error": "",
    "events": [{
      "__typename": "GnoEvent",
      "type": "SetQuote",
      "pkg_path": "gno.land/r/g14wv9hyngyrzpxn3ztvnx2jygs50m0240qncm8j/access",
      "attrs": [
        { "key": "ugnot",     "value": "422749053" },
        { "key": "usdCents",  "value": "2500" },
        { "key": "expiresAt", "value": "1795435211" }
      ]
    }]
  }
}
```

## What is NOT proven here

- No `MsgRun` appeared in the window sampled in section 7, so the specific
  decode that tx-indexer `v1.5.1` exists to fix (`MsgRun`'s `max_send`) was
  never exercised. Every other message type onyx used in those 10,700 blocks
  decoded correctly.
- No gnoscope process was ever started against this stack. The queries are
  gnoscope's, taken verbatim from its source and replayed by hand; its pool
  selection, caching and rendering are not demonstrated.
- The halt-and-restart procedure at heights 445000, 754000 and 1203400, and the
  binary switch to `v1.8.0-rc.1`. See the report for how far this got.
- Anything about the deployment host. Nothing was deployed anywhere.
