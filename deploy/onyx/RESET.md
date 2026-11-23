# When onyx resets or is retired

Onyx is a testnet. Two different things will eventually happen to it, and they
need different responses. Neither announces itself to this host.

| | What it is | What breaks |
|---|---|---|
| **Reset in place** | onyx-1 is wiped and relaunched, possibly keeping the chain ID | the local node is on an orphan chain and keeps serving it |
| **Retirement** | the next gemstone testnet ships and onyx is switched off | `*.onyx.testnets.gno.land` stops resolving; the node loses every peer |

## The trap, stated plainly

With the local endpoints listed first, **a stale local node can make gnoscope
serve a dead chain and look completely healthy while doing it.**

gnoscope's endpoint pool
([`pkg/indexer/indexer.go`](https://github.com/gnoverse/gnoscope/blob/master/pkg/indexer/indexer.go),
`selectEndpoint`) works like this:

1. On its first probe it adopts a chain identity: `chain_id + ":" + hash of
   block 1`, taken from **the earliest endpoint in configured order that can
   prove one**. With local first, that is the local indexer.
2. That identity is then **fixed for the lifetime of the gnoscope process**.
3. Any pool member whose fingerprint disagrees is **never selected**, however
   healthy and however far along it is.

So after onyx resets: the local node, still running, still answering, still
holding the old block 1, defines the pool. The public endpoints now serve a
different block 1, fail the fingerprint check, and are excluded. gnoscope
reports a chain frozen at the reset height. Nothing errors. No 429s. No alerts.

The lever that fixes it is short, and it is the first thing to reach for:

```shell
sudo systemctl stop tx-indexer-onyx gnoland-onyx
sudo systemctl restart gnoscope
```

With the local endpoint unreachable it cannot prove a fingerprint, so identity
is adopted from the next entry in configured order, which is the public
indexer. gnoscope comes back on the real chain, with the 429s back, which is
the correct degraded state. Everything else below is cleanup done at leisure.

Note that stopping the node is not enough on its own: the identity is only
re-adopted on a gnoscope restart. Restart gnoscope too.

## How we notice

Three checks, cheapest first. None of them calls a public endpoint on a timer.

### 1. The local node's own clock (the one that actually fires)

The node knows it is stuck without asking anyone. Onyx commits every 3s, so the
block time of the tip should never be far behind wall clock.

```shell
#!/bin/sh
# /usr/local/bin/onyx-staleness-check, run from cron every 15 minutes.
# Silent when healthy; cron mails whatever it prints.
S=$(curl -fs --max-time 10 http://127.0.0.1:26757/status 2>/dev/null) || {
  echo "onyx node unreachable on 127.0.0.1:26757"; exit 1; }
BT=$(printf '%s' "$S" | jq -r .result.sync_info.latest_block_time)
LAG=$(( $(date -u +%s) - $(date -u -d "$BT" +%s) ))
[ "$LAG" -gt 1800 ] && echo "onyx node tip is ${LAG}s stale (block time $BT): reset, retirement, or a stuck node"
exit 0
```

Verified on lyra 2026-11-23 against both a healthy node (`lag=4s`, silent,
exit 0) and a stopped one (prints the unreachable line, exit 1).

A 30 minute threshold is 600 blocks. Ordinary catch-up never gets there; a
reset, a retirement and a hung node all do. This check does not tell them
apart, which is fine: all three need a human.

### 2. Peers

A retirement takes the DNS with it. Verified on lyra 2026-11-23:
`pearl.testnets.gno.land` and `rpc.pearl.testnets.gno.land`, the testnet onyx
replaced, both answer **NXDOMAIN**. gnoscope's own config notes record the same
for topaz, sapphire and pearl on 2026-09-29. Retirement is not a redirect, so
the seeds simply stop existing and the node can never find a peer again.

```shell
getent hosts seed-1.onyx.testnets.gno.land || echo "onyx seeds gone: retirement"
curl -s http://127.0.0.1:26757/net_info | jq '.result.n_peers'
```

`n_peers` at 0 with the seeds NXDOMAIN is a retirement. `n_peers` healthy with
a stale tip is a reset or a stuck node.

### 3. The gno repository, which is where it is actually announced

The authoritative record is
[`misc/deployments/`](../../misc/deployments/). A new testnet arrives as a new
`misc/deployments/<stone>.gno.land/` directory with its own `upgrades.json`
carrying a new `chain_id` and `genesis_sha256`. A reset in place shows up as a
changed `genesis_sha256` under
[`misc/deployments/onyx.gno.land/upgrades.json`](../../misc/deployments/onyx.gno.land/upgrades.json).

Worth a monthly look, and worth looking at before believing check 1:

```shell
curl -fsS https://raw.githubusercontent.com/gnolang/gno/master/misc/deployments/onyx.gno.land/upgrades.json \
  | jq -r '.chain_id, .genesis_sha256'
# expected today: onyx-1
#                 4b006fd7ccdec052865accc84dd29b2b76f8b57b2560789a15eedaa88f0e26c5
```

Record the block 1 hash from INSTALL.md step 10 next to this file when the node
first syncs. It is the other half of the fingerprint and the only thing that
distinguishes a reset onyx-1 from the original.

## Procedure: reset in place

1. Stop both units and restart gnoscope, as above. gnoscope is now correct and
   rate limited. The rest is not urgent.
2. Confirm the new genesis:
   `curl -fsSL -o genesis.json https://github.com/gnolang/gno/releases/download/chain%2Fonyx/genesis.json`
   and check its sha256 against the new `genesis_sha256` in `upgrades.json`.
   If the sha256 is unchanged, this was not a reset: look for a stuck node.
3. Delete the old chain data. It is worthless and it is the thing that would
   poison the pool again:
   ```shell
   sudo rm -rf /data/onyx/node/data /data/onyx/indexer-db
   ```
   Keep `/data/onyx/node/config` and `/data/onyx/node/secrets`.
4. Re-read [`misc/deployments/onyx.gno.land/VALIDATOR.md`](../../misc/deployments/onyx.gno.land/VALIDATOR.md)
   and `UPGRADES.md` from scratch. Do not assume the seeds, the version ladder
   or the chain ID carried over. A relaunch is free to change all three.
5. Rebuild the binaries at whatever versions the new `upgrades.json` names. The
   two in [BUILD.md](./BUILD.md) are almost certainly wrong by then.
6. Redo INSTALL.md from step 3. Re-record the new block 1 hash.
7. Only when the local node is caught up, put it back into
   `/etc/gnoscope/networks.json` and restart gnoscope. Check the startup line.

## Procedure: retirement

The onyx entry is simply deleted, local and public alike. Do not leave it in
with the public endpoints as fallbacks: those hostnames no longer resolve, so
the network degrades to a per-network circuit breaker and every onyx page in
gnoscope becomes a dead end that costs DNS lookups to discover.

1. `sudo systemctl disable --now tx-indexer-onyx gnoland-onyx`
2. Remove the onyx entry from `/etc/gnoscope/networks.json`, restart gnoscope,
   check the startup line names only the networks you expect.
3. `sudo rm -rf /data/onyx` and the two unit files, and remove the
   `gnoland-v1.*-rc.1` and `tx-indexer-*` binaries and `/opt/gnoland/gnoroot-*`
   checkouts if nothing else uses them.
4. If a new stone replaces it: everything in this directory applies to it with
   the chain ID, the genesis, the seeds and the version ladder swapped. Start
   from its own `misc/deployments/<stone>.gno.land/VALIDATOR.md`.

## Procedure: a routine upgrade, which is not a reset

Onyx takes every mainnet release first, as a release candidate. Roughly monthly
a new row appears in
[`misc/deployments/onyx.gno.land/UPGRADES.md`](../../misc/deployments/onyx.gno.land/UPGRADES.md)
with a halt height. At that height the node commits, panics in the next block's
`BeginBlock` and exits, and `Restart=on-failure` restart-loops it because the
installed binary is below the new `halt_min_version`.

This is the ordinary maintenance cost of running this node and it is the reason
check 1 exists: a missed upgrade looks exactly like a stale tip.

1. Build the new version per [BUILD.md](./BUILD.md), with `VERSION` set to the
   candidate tag, and clone a matching GNOROOT.
2. Point `gnoland-onyx.service` at the new binary and the new GNOROOT, all
   three places.
3. `sudo systemctl daemon-reload && sudo systemctl restart gnoland-onyx`.

The indexer does not need to move in lockstep, but check whether tx-indexer has
cut a release pinning the new gno version: `v1.5.1` exists precisely because a
gno bump was needed for `MsgRun` to decode. An indexer built against an older
gno silently mis-decodes new message shapes rather than failing.

**Subscribe to the signal rather than discovering it.** Watch
[gnolang/gno releases](https://github.com/gnolang/gno/releases) for `*-rc.*`
tags, or diff `upgrades.json` on a monthly cron. An upgrade is announced
roughly a week ahead through a GovDAO proposal on onyx itself.
