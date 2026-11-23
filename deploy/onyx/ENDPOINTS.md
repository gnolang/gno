# The public onyx endpoints, the rate limit, and what else is out there

Measured on lyra, 2026-11-23. Every figure marked *measured* came from one
request per endpoint, deliberately: the per-IP budget on the onyx endpoints is
the thing this whole exercise exists to stop spending. Everything marked
*source* is read from `gnolang/infrastructure`, the private OpenTofu repository
that defines the edge.

## onyx-1 identity

```
chain_id          onyx-1
genesis_time      2026-09-28T00:00:00Z
genesis sha256    4b006fd7ccdec052865accc84dd29b2b76f8b57b2560789a15eedaa88f0e26c5
block 1 hash      PGlpc0OZKeAItjmtZzNVJSzsmY40K9X7XAMCuTa6wj4=
```

`chain_id + ":" + block 1 hash` is exactly the fingerprint gnoscope's endpoint
pool uses. The local node built by [INSTALL.md](./INSTALL.md) reproduced the
block 1 hash above from genesis, which is the proof that it is on the same
chain and not merely answering plausibly.

## Where the 429 comes from

**Source.** It is a **Cloudflare rate-limiting rule**, defined in
`gnolang/infrastructure` at `tofu/edge/main.tf` with its budgets in
`tofu/edge/variables.tf`, documented in `docs/edge.md`.

| | |
|---|---|
| `indexer.` budget | **3,100 requests per 300 s**, about 10.3 req/s sustained |
| `rpc.` budget | **3,900 requests per 600 s**, about 6.5 req/s sustained |
| counted by | `[cf.colo.id, ip.src]`, and only `requests_to_origin` |
| over the limit | **block for the rule's own period**: 300 s on `indexer.`, 600 s on `rpc.` |
| the body | a JSON-RPC error, `{"code":-32000,"message":"Per-IP rate limit exceeded; retry after 300 seconds."}` |

The two facts that matter most:

1. **One counter across every network's names.** Cloudflare's Business plan
   cannot count per hostname, so `indexer.gno.land`,
   `indexer.onyx.testnets.gno.land` and `indexer.latest.testnets.gno.land` all
   decrement **the same per-IP budget**. Likewise for the three `rpc.` names.
2. **One request over costs a 300 second blackout.** Cloudflare refuses a block
   shorter than the rule's own period, so the penalty is not a slowdown, it is
   five minutes of nothing. That is why 277 refusals produced 38 breaker trips.

These are not stock defaults tuned by nobody. The budgets were raised
deliberately on 2026-10-08/09, from `rpc.` 3,000 and `indexer.` 2,400, to admit
the heaviest observed client with about 10 % headroom. The variable's own
description records the measurement: the heaviest client across both networks
made 2,784 indexer requests in its busiest 5 minutes, 1,431 on mainnet and
1,499 on onyx.

So gnoscope tripping this limit means gnoscope sent **more than 3,100 indexer
requests in 5 minutes**, more than 10 per second sustained. That is already
above a budget sized for the heaviest client on the network.

## Who serves onyx publicly

| Endpoint | Chain | Measured | Separate rate-limit budget? |
|---|---|---|---|
| `https://rpc.onyx.testnets.gno.land` | onyx-1, tip 1,349,724 | moniker `gno-core-rpc-public-i-0093c339287ae2ba8` | no |
| `https://indexer.onyx.testnets.gno.land/graphql/query` | onyx-1 | block 1 hash matches | no |
| `https://rpc.latest.testnets.gno.land` | onyx-1, tip 1,349,936 | moniker `gno-core-rpc-public-i-0525e287d7d4e058d` | **no** |
| `https://indexer.latest.testnets.gno.land/graphql/query` | onyx-1, tip 1,349,935 | block 1 hash matches | **no** |

`*.latest.testnets.gno.land` is documented in
[`docs/resources/gnoland-networks.md`](../../docs/resources/gnoland-networks.md)
as "always points at the current testnet", which today is onyx. It is **not** a
CNAME onto the same box: the two RPCs report different AWS instance IDs, so
these are genuinely separate origin nodes.

But the infra source is explicit that it buys nothing against the limit: *"The
two API aliases, `rpc.` and `indexer.`, ride the edge with their onyx twins:
one service, one rate-limit budget whichever name a client uses."*

**So adding the `latest` aliases as gnoscope fallbacks is worth doing and is
not a fix.** It buys exactly two things: an origin that is down is skipped, and
an origin that is merely *behind* is overtaken. It buys zero headroom against
429s, because the counter is shared.

It is also a **moving** alias: when onyx is retired it will serve the next
gemstone. That is safe rather than dangerous only because gnoscope excludes a
pool member whose fingerprint disagrees with the pool's, so it drops out rather
than splicing two chains together. Still a reason to keep it a fallback and
never the primary.

## Who does not serve onyx

`indexer.onbloc.xyz` and `rpc.onbloc.xyz` are the **mainnet** alternates.
Probed 2026-11-23: both answer `gnoland-1`, block 1 hash
`ojbtIrq6a6zSMjUnvBY0fCI+tIFLerysma5P3YNZtR0=`, height 1,761,99x. **Neither
serves onyx.** There is no onbloc equivalent for the testnet.

They are also on their own infrastructure, off the gno.land Cloudflare edge, so
they do not touch the shared counter at all. That is very likely the whole
explanation for the asymmetry in the original measurement: gnoscope's mainnet
reads land on onbloc and cost nothing against the budget, while its onyx reads
have nowhere to go and spend the entire shared budget alone. *(Inference, not
confirmed: it follows from the configuration but was not observed directly.)*

## The one public thing that is NOT rate limited

<https://snapshots.onyx.testnets.gno.land> (and
`snapshots.latest.testnets.gno.land`). The snapshot sites point straight at
each network's CloudFront distribution with **neither Cloudflare nor AWS WAF in
front**, because Cloudflare caches nothing that large on the gno.land plan.

Measured 2026-11-23: `index.json` lists 40 snapshots, newest
`onyx-1_1349553_20261123T103758Z.tar.zst` at height 1,349,553, 14.8 GB, with
its sha256 beside it. A new one every 30 minutes, kept 24 hours. Downloaded
here at 33 MB/s.

This is what makes the local node cheap to stand up: a restore, not a replay.
