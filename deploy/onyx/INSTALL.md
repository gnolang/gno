# Installing the onyx-1 node and indexer

What you end up with: a `gnoland` full node and a `tx-indexer` on the
deployment host, both bound to loopback, serving gnoscope's onyx pages without
touching `indexer.onyx.testnets.gno.land` or `rpc.onyx.testnets.gno.land`.

Read [RESET.md](./RESET.md) before you start. Onyx is a testnet. It will be
retired, and when it is, every step below becomes wrong in a way that does not
announce itself.

## 0. Two ways in, and which one to use

| | Snapshot restore (§A) | Replay from genesis (§B) |
|---|---|---|
| Node at the tip | **25 minutes**, measured | about 6 hours, extrapolated from 63 blocks/s |
| Indexer backfilled | **18 minutes** more, either way | 18 minutes more, either way |
| Binaries needed | one, `v1.8.0-rc.1` | two, `v1.7.1-rc.1` then `v1.8.0-rc.1` |
| Hands-on moments | one | four |
| Trust | you trust gno.land's published snapshot | you trust only the genesis and the chain |

**Use the snapshot.** gno.land publishes a pruned node database for onyx every
30 minutes at <https://snapshots.onyx.testnets.gno.land>, with a sha256 beside
it. That site is on CloudFront, **off the Cloudflare edge**, so downloading it
costs nothing against the rate limit this node exists to stop paying.

§B is kept because it is the only path that trusts nothing but genesis, and
because it is what you fall back to if the snapshot site is down or if you ever
need to verify the published snapshot against an independent replay.

## 1. Budget

| | |
|---|---|
| Disk | **31.5 GB measured** for node (23 GB) plus indexer (8.5 GB), plus 14 GB transient for the snapshot archive |
| RAM | **4.8 GB resident for the node once it is at the tip**, measured. The indexer peaks near 0.7 GB while backfilling and settles at ~0.25 GB. A node still replaying history sits at 1.2 GB; do not size from that figure |
| CPU | one core while replaying, far less once caught up |
| Ports | 26756 (p2p, loopback), 26757 (node RPC, loopback), 8567 (indexer, loopback), 8568 (indexer metrics, loopback) |

None of those four ports is in the host's in-use list. 8546, the indexer's own
default, **is** taken by the existing mainnet tx-indexer, which is why 8567 is
used here.

## 2. User and directories

```shell
sudo useradd --system --home-dir /var/lib/gnoland-onyx --create-home --shell /usr/sbin/nologin gnoland || true
sudo install -d -o gnoland -g gnoland /data/onyx /data/onyx/node /data/onyx/indexer-db
sudo install -d -o root -g root /opt/gnoland
```

If a `gnoland` user already exists for the mainnet units, reuse it and skip the
`useradd`. Do not reuse `/var/lib/gnoland`: a shared `HOME` between a mainnet
node and a testnet node is how you end up with one chain's keys in the other's
directory.

## 3. Binaries and GNOROOTs

Build per [BUILD.md](./BUILD.md). For §A you need only the `v1.8.0-rc.1` pair;
build both if you intend §B.

```shell
sudo install -m 0755 gnoland-v1.8.0-rc.1 tx-indexer-v1.5.1 /usr/local/bin/
sudo git clone --branch v1.8.0-rc.1 --depth 1 https://github.com/gnolang/gno.git /opt/gnoland/gnoroot-v1.8.0-rc.1
/usr/local/bin/gnoland-v1.8.0-rc.1 version   # must print v1.8.0-rc.1
```

Each GNOROOT must be the checkout of the same tag as the binary that uses it.
The node reads `gnovm/stdlibs` from GNOROOT and the tags' stdlibs differ: a
mismatched pair reaches a different app hash than the network and stops.

## 4. Genesis

A snapshot is a database and nothing else: no key, no configuration, no
genesis. The node still needs `genesis.json`.

```shell
sudo install -o gnoland -g gnoland -m 0644 genesis.json /data/onyx/genesis.json
sha256sum /data/onyx/genesis.json
# must be 4b006fd7ccdec052865accc84dd29b2b76f8b57b2560789a15eedaa88f0e26c5
```

That value is `genesis_sha256` in
[`misc/deployments/onyx.gno.land/upgrades.json`](../../misc/deployments/onyx.gno.land/upgrades.json),
on the [`chain/onyx` release page](https://github.com/gnolang/gno/releases/tag/chain%2Fonyx),
and at `https://snapshots.onyx.testnets.gno.land/genesis.json.sha256`. Three
independent publications of the same number. A copy of the file is committed
next to this one as [`genesis.json`](./genesis.json).

## 5. Config and secrets

```shell
sudo -u gnoland /usr/local/bin/gnoland-v1.8.0-rc.1 config init \
    -config-path /data/onyx/node/config/config.toml
sudo -u gnoland /usr/local/bin/gnoland-v1.8.0-rc.1 secrets init \
    -data-dir /data/onyx/node/secrets

C=/data/onyx/node/config/config.toml
cfg() { sudo -u gnoland /usr/local/bin/gnoland-v1.8.0-rc.1 config set -config-path "$C" "$1" "$2"; }
cfg moniker                              "onyx-gnoscope"
cfg p2p.persistent_peers                 "g1x5mlj5ava0dw9vkf4j6admjlzswm6f06p44krn@seed-1.onyx.testnets.gno.land:26656,g1grq5zswt0dlwwe7clr4359w70k2ewgse0gcwck@seed-2.onyx.testnets.gno.land:26656"
cfg application.prune_strategy           "syncable"
cfg consensus.timeout_commit             "3s"
cfg consensus.peer_gossip_sleep_duration "10ms"
cfg p2p.flush_throttle_timeout           "10ms"
cfg mempool.size                         "10000"
cfg p2p.max_num_outbound_peers           "40"
cfg p2p.laddr                            "tcp://127.0.0.1:26756"
cfg rpc.laddr                            "tcp://127.0.0.1:26757"
```

Full annotated list in [config.toml.fragment](./config.toml.fragment). The
validator key `secrets init` writes is never used: this node is not in the
valset and never will be. It is generated because the node refuses to start
without one.

## 6. Units

```shell
sudo install -m 0644 gnoland-onyx.service tx-indexer-onyx.service /etc/systemd/system/
sudo systemctl daemon-reload
```

Do **not** start either yet.

---

# §A. Restore from the published snapshot

## A1. Pick and fetch

```shell
curl -fsS https://snapshots.onyx.testnets.gno.land/index.json | jq '.snapshots[0]'
```

Take the newest. Each entry carries `file`, `height`, `time`, `size` and
`sha256`. Snapshots older than 24 hours are deleted, so there is always a
recent one and never an old one.

```shell
cd /var/tmp
F=$(curl -fsS https://snapshots.onyx.testnets.gno.land/index.json | jq -r '.snapshots[0].file')
curl -fL -O "https://snapshots.onyx.testnets.gno.land/$F"
curl -fL -O "https://snapshots.onyx.testnets.gno.land/$F.sha256"
sha256sum -c "$(basename "$F").sha256"
```

**Check the sha256 before unpacking.** It is a tar archive from the network, so
unpack it only after the checksum passes, and only into the directory named
below.

## A2. Unpack

The archive's single top-level entry is `db/`. It unpacks into the directory
that will **contain** `db/`, which for `--data-dir /data/onyx/node` is
`/data/onyx/node` itself, beside `config/` and `secrets/`:

```shell
sudo systemctl stop gnoland-onyx 2>/dev/null
zstd -d -c "$(basename "$F")" | sudo -u gnoland tar -x -C /data/onyx/node
sudo chown -R gnoland:gnoland /data/onyx/node/db
ls /data/onyx/node/db   # blockstore.db  gnolang.db  state.db  genesis-cache
```

The infra documentation says "the node's data directory, `/var/lib/gnoland/data`
on our hosts". That is **their** `--data-dir`, not a `data/` subdirectory of
ours. Unpacking one level too deep leaves the node starting from genesis with
23 GB of unreachable database beside it, which looks like the snapshot having
silently failed.

If a `db/` is already there from an earlier attempt, move it aside first rather
than unpacking over it.

`db/genesis-cache/` is named after the genesis sha256. Confirm it reads
`4b006fd7ccdec052865accc84dd29b2b76f8b57b2560789a15eedaa88f0e26c5`: that is a
third, independent check that the snapshot belongs to the genesis you verified
in step 4.

## A3. Start

```shell
sudo systemctl start gnoland-onyx
journalctl -u gnoland-onyx -f
```

The node replays from the snapshot's height to the tip. A snapshot cut within
the last 30 minutes is at most ~600 blocks behind, which it closes in seconds.

Then go to §8.

---

# §B. Replay from genesis

Only if §A is unavailable, or if you specifically want a node that trusts
nothing but the genesis file.

## B1. Start on v1.7.1-rc.1

The shipped `gnoland-onyx.service` names `v1.8.0-rc.1`. For a genesis replay it
must name `v1.7.1-rc.1` instead, in all three places (the `GNOROOT`
environment line, `ExecStart`'s binary, `--gnoroot-dir`):

```shell
sudo sed -i 's/v1\.8\.0-rc\.1/v1.7.1-rc.1/g' /etc/systemd/system/gnoland-onyx.service
sudo systemctl daemon-reload
sudo systemctl start gnoland-onyx
```

`v1.8.0-rc.1` cannot do this. Started at genesis it stops within a few blocks
with `wrong Block.Header.AppHash`, because it refuses the non canonical
transaction encoding that earlier onyx blocks contain.

## B2. The three stops

Replay re-runs every past governance halt. The node commits the halt height,
panics in the next block's `BeginBlock` and exits. `Restart=on-failure` brings
it straight back, and for the first two that is all that is needed.

| Stops after height | What to do |
|---|---|
| 445000 | nothing. `v1.7.1-rc.1` satisfies `halt_min_version` `v1.6.0-rc.1`, so the automatic restart continues |
| 754000 | nothing. Same, for `v1.7.0-rc.1` |
| 1203400 | **switch the unit to `v1.8.0-rc.1`** |

At the third the node will restart-loop, because `v1.7.1-rc.1` is below
`v1.8.0-rc.1`'s `halt_min_version`. Watch for it:

```shell
journalctl -u gnoland-onyx | grep -i halt | tail
```

Then:

```shell
sudo systemctl stop gnoland-onyx
sudo sed -i 's/v1\.7\.1-rc\.1/v1.8.0-rc.1/g' /etc/systemd/system/gnoland-onyx.service
sudo systemctl daemon-reload
sudo systemctl start gnoland-onyx
```

If a restart ever lands between a halt proposal executing and its halt height
while the running binary already satisfies that halt, the node refuses to
start. Set `skip_upgrade_height` to that height for that one restart
and set it back to `0` immediately after. Left in place it disables both
upgrade gates.

> **Unverified here.** The whole of §B2 is taken from onyx's own
> [`VALIDATOR.md`](../../misc/deployments/onyx.gno.land/VALIDATOR.md) and
> [`UPGRADES.md`](../../misc/deployments/onyx.gno.land/UPGRADES.md), which
> state it was verified upstream on 2026-11-17. The genesis replay run on lyra
> on 2026-11-23 reached height 174,063 and was stopped there, so it never met
> the first halt at 445,000. Expect to debug this section the first time you
> use it.

---

## 8. Telling it is caught up

```shell
curl -s http://127.0.0.1:26757/status | jq '.result.sync_info'
```

Three things must all hold:

1. `catching_up` is `false`.
2. `latest_block_height` is within a handful of the public tip. Compare against
   `curl -s https://rpc.onyx.testnets.gno.land/status | jq -r .result.sync_info.latest_block_height`,
   **once**, not in a loop: that request counts against the per-IP budget this
   whole exercise exists to stop spending.
3. `latest_block_time` is within a few seconds of now. Onyx commits every 3s,
   so a node more than a minute behind in block time is stuck, whatever
   `catching_up` says.

`catching_up: false` on its own is not enough. It means the node has left
block-sync fast catch-up, not that it is at the tip.

## 9. Start the indexer

Only once the node is caught up. Earlier works, but the indexer competes with
the node for CPU: measured on lyra, running it during a genesis replay cut the
node's block rate from 66/s to 44/s, a third.

```shell
sudo systemctl enable --now tx-indexer-onyx
journalctl -u tx-indexer-onyx -f
```

The indexer reindexes from height 0 over loopback. It only reads blocks and
does not execute them, so it is far faster than the node's replay: measured at
about 780 blocks/s against a node on the same host, roughly twelve times the
node's own rate. It is never the bottleneck.

Caught up when

```shell
curl -s -X POST http://127.0.0.1:8567/graphql/query \
  -H 'Content-Type: application/json' \
  --data '{"query":"{ latestBlockHeight }"}'
```

returns a height matching the node's.

## 10. Verify it is the right chain, and that gnoscope's queries answer

This is the query gnoscope's endpoint pool actually sends to fingerprint and
rank a candidate endpoint
([`pkg/indexer/indexer.go`](https://github.com/gnoverse/gnoscope/blob/master/pkg/indexer/indexer.go),
`func (c *Client) probe`):

```shell
curl -s -X POST http://127.0.0.1:8567/graphql/query \
  -H 'Content-Type: application/json' \
  --data '{"query":"{ latestBlockHeight getBlocks(where: { height: { eq: 1 } }) { chain_id hash } }"}'
```

`chain_id` must be `onyx-1` and `hash` must be
`PGlpc0OZKeAItjmtZzNVJSzsmY40K9X7XAMCuTa6wj4=`. That pair is the fingerprint.
It is what distinguishes onyx-1 from a reset chain that kept the same ID, and
RESET.md turns on it.

Worked transcripts of this and of a full gnoscope block and transaction query,
run against this stack, are in [PROOF.md](./PROOF.md).

## 11. Point gnoscope at it

Merge [gnoscope-networks.onyx.json](./gnoscope-networks.onyx.json) into
`/etc/gnoscope/networks.json`, keeping the mainnet entry as it is, then restart
gnoscope and check the line it logs at startup:

```
networks [gnoland1 onyx] (from config file)
```

and `curl -s http://127.0.0.1:8888/api/networks`.

A wrongly configured gnoscope starts cleanly, syncs real data and looks
healthy. It is just the wrong chain. Check the line.

## 12. Arm the staleness check

Do this on the same day, not later. A node that silently stops is worse than no
node: see the opening section of [RESET.md](./RESET.md) for why. The script is
there.
