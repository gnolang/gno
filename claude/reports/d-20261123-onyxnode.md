# A local onyx-1 node for gnoscope: recipe, measurements, recommendation

Date: **2026-11-23**. Host: **lyra**. Nothing was deployed anywhere, and no
host but lyra was touched.

Artefacts: [`deploy/onyx/`](../../deploy/onyx/). Working data, binaries and
chain databases: `~/ws/d-20261123-onyxnode/.work/` (see
[Disk to reclaim](#disk-to-reclaim)).

---

## Recommendation

**Run the node.** Not because the alternatives are bad engineering, but because
measuring them showed one of them buys literally nothing and the other has
already been tried and already regressed.

The decisive number is this: the public onyx endpoints sit behind a Cloudflare
rate-limiting rule whose budget is **3,100 indexer requests per 300 seconds per
IP, counted across every network's `indexer.` hostnames together**. Not per
hostname. Not per network. One bucket. Going one request over costs a **300
second blackout**, because Cloudflare refuses a block shorter than the rule's
own period.

Everything else follows from that.

### Why not the second public endpoint

There **is** a second public onyx endpoint, and I found it:
`rpc.latest.testnets.gno.land` and `indexer.latest.testnets.gno.land`. They are
real, separate origin nodes (different AWS instance IDs in the RPC moniker),
both serving onyx-1 with the matching block 1 hash, both within one block of
the primary.

And they share the same rate-limit budget. The infra source is explicit: *"The
two API aliases, `rpc.` and `indexer.`, ride the edge with their onyx twins:
one service, one rate-limit budget whichever name a client uses."*

So adding them buys resilience against an origin being **down** or **behind**,
which is worth having and costs one line, and **zero headroom against 429s**.
It is in the networks.json fragment for the first reason. It is not an answer
to this problem.

There is no onbloc equivalent for onyx. `indexer.onbloc.xyz` and
`rpc.onbloc.xyz` both answer `gnoland-1` (probed 2026-11-23, block 1 hash
`ojbtIrq6a6zSMjUnvBY0fCI+tIFLerysma5P3YNZtR0=`). That absence is the whole
reason mainnet does not 429 and onyx does: mainnet's reads land on onbloc,
which is off the gno.land edge entirely and costs nothing against the shared
budget, while onyx's reads have nowhere to go and spend the whole budget alone.
*(Inference from the configuration, not directly observed.)*

### Why not ask for a raised limit

**Who to ask:** whoever administers `gnolang/infrastructure`. The decisions in
that repository's rate-limit variables are attributed to **Antoine**
(2026-10-08, 2026-10-09, 2026-11-11). It is the same organisation, so this is a
pull request, not a support ticket.

**What to ask for:** not a raise. An **exemption**. Concretely, add
`and ip.src ne <gnoscope.com's egress IP>` to the indexer rate-limit rule's
counting expression in `tofu/edge/main.tf`, rather than raising
`var.api_indexer_rate_limit_requests` in `tofu/edge/variables.tf`.

**Why a raise is the wrong ask, with the history:** it has already been done
once. The budgets were raised on 2026-10-08/09 from `rpc.` 3,000 and `indexer.`
2,400 to 3,900 and 3,100, specifically sized to admit the heaviest observed
client with about 10 % headroom. Six weeks later gnoscope took 277 refusals in
a day. A raise is a treadmill, it lifts the ceiling for crawlers too, and the
Business plan caps the counting period at 600 seconds so there is not much room
left to move.

An exemption is better and is still worth asking for **even if the node is
deployed**, for a reason that has nothing to do with capacity: it is what makes
the *degraded* state tolerable. When the local node is down, mid-upgrade, or
mid-reset, gnoscope falls back to the public endpoints, and that fallback
should not immediately start 429ing. Ask for it as insurance on the fallback
path, not as the fix.

### Why the node

1. **It is the only option that removes the dependency rather than
   renegotiating it.** A budget someone else sets can be re-tightened; a local
   node cannot be rate limited.
2. **It protects mainnet too.** The counter is shared. Every onyx request
   gnoscope stops making is budget returned to its mainnet reads.
3. **It is far cheaper to stand up than expected.** I assumed a multi-hour
   replay from genesis. gno.land publishes a pruned node database for onyx
   every 30 minutes at <https://snapshots.onyx.testnets.gno.land>, on
   CloudFront, **off the Cloudflare edge and not rate limited**. Measured
   end to end on lyra: **25 minutes from nothing to a node at the live tip**,
   43 minutes to a fully indexed stack.

### The cost, honestly

| | |
|---|---|
| Stand-up | **43 minutes measured**: 25 min to a node at the tip, 18 min more for the indexer to backfill |
| Disk | **31.5 GB** steady state, plus 14 GB transient during the restore |
| RAM | node **4.8 GB** resident at the tip, indexer 234 MB once backfilled |
| Recurring | a coordinated upgrade roughly monthly: rebuild a binary, edit a unit, restart. ~30 min, and it is a **blocking** failure that restart-loops until a human acts |
| Occasional | a testnet reset or retirement every few months: a few hours |
| The real risk | a stale local node can make gnoscope serve a dead chain and look healthy doing it |

That last one is not a footnote. gnoscope's pool adopts its chain identity from
the earliest configured endpoint that can prove one and then **excludes anything
that disagrees**. Local first means a local node left running on a reset onyx
defines the pool, and the real public endpoints are locked out as impostors.
No errors, no 429s, no alerts, just a frozen chain reported as a healthy one.

**Arm the staleness cron on day one.** It is fifteen lines, it is in
[RESET.md](../../deploy/onyx/RESET.md), and without it this recommendation is
worse than doing nothing.

### Do all three

1. **Now, free:** add the `latest` aliases to the onyx fallbacks. One line.
2. **Now, one PR:** ask gno infra for the IP exemption, as insurance on the
   fallback path.
3. **Then:** deploy the node per [INSTALL.md](../../deploy/onyx/INSTALL.md),
   with the staleness cron armed the same day.

A fourth lever exists that is outside this dispatch's scope: gnoscope #526
landed `Retry-After` handling and singleflight but **not** the short-TTL
response cache that #525 part 2 also proposed. If gnoscope is really sending
more than 10 requests per second sustained to one indexer, a response cache is
worth more than any of the above and costs no operations at all.

---

## What I measured

### Versions

| Component | Version | Commit |
|---|---|---|
| gnoland (genesis replay) | `v1.7.1-rc.1` | `83431a7f0261882541a8a8dcf1643e6f246bbbe5` |
| gnoland (tip, and snapshot restore) | `v1.8.0-rc.1` | `fc79041f4cd1a7b3ed4b316dc5ac14162eb78730` |
| tx-indexer | `v1.5.1` | `085228830cd5feda125e924720e3c0dfaeaf1663` |
| Go toolchain | go1.25.9 linux/amd64 | |

Both gnoland commits match `misc/deployments/onyx.gno.land/upgrades.json`
exactly. tx-indexer `v1.5.1` is the tip of its master and pins
`github.com/gnolang/gno v1.8.0-rc.1` in `go.mod`, the exact version onyx runs;
its newest commit is *"fix(deps): build against gno v1.8.0-rc.1 so MsgRun's
max_send decodes"*, so an older indexer mis-decodes onyx's recent `MsgRun`
transactions.

All three built `CGO_ENABLED=0`, `-trimpath`, static ELF linux/amd64. The
gnoland build is **byte reproducible**: building `v1.8.0-rc.1` twice gave
identical sha256. Commands and checksums in
[BUILD.md](../../deploy/onyx/BUILD.md).

### Genesis, verified three ways

```
sha256  4b006fd7ccdec052865accc84dd29b2b76f8b57b2560789a15eedaa88f0e26c5
```

matching `upgrades.json`'s `genesis_sha256`, the `chain/onyx` release page, and
the name of the `db/genesis-cache/` directory inside the published snapshot.

### Path A: restore from the published snapshot (recommended)

| Step | Measured |
|---|---|
| Download 14.83 GB `.tar.zst` | **532 s** (8 m 52 s) at 27.9 MB/s |
| sha256 verify | matched `efdc885d…a5397` |
| Extract | **90 s**, 23 GB on disk |
| Node start to RPC answering | **40 s** |
| Catch up from snapshot height 1,349,553 to the live tip | under a minute of productive work, see note |
| **Total, nothing to live tip** | **about 25 minutes** |

Final state, verified against the public RPC with a single request:

```
local  1350473
public 1350473      delta 0
```

`catching_up: false`, 7 peers, block time within seconds of wall clock.

*Note on the catch-up figure:* wall clock from node2's start to
`catching_up: false` was 698 s, but roughly 560 s of that was a self-inflicted
stall in my test setup, not a property of the procedure. I had the genesis
replay node still running on the same machine, and **onyx's seeds accept only
one connection per source IP**, so the second node sat at 0 peers until I
stopped the first. The moment the first node stopped, the second attached 7
peers and closed its ~920 block gap almost immediately. On the deployment host
there is only one node and this does not arise. It is worth knowing anyway: two
onyx nodes behind one NAT will not both find peers through these seeds.

### Path B: replay from genesis (fallback)

Run to measure the path that trusts nothing but genesis, then stopped
deliberately to free the seed slot for path A.

| | |
|---|---|
| Reached | height **174,063** of 1,349,724 in **2,360 s** (39 m 20 s) |
| Average | 73.8 blocks/s, inflated by the near-empty first 40k blocks (322 b/s) |
| Steady state, clean | **60 to 66 blocks/s**, measured over 5 minutes with nothing else running |
| Steady state, with the indexer running alongside | **44 blocks/s**, a third slower |
| RSS | 1.17 GB |
| Disk | 2.2 GB at 174k blocks; marginal cost converging to ~20 KB/block |

**Extrapolation, stated as such:** at the measured clean steady state of 63
blocks/s, 1,349,724 blocks is **about 6 hours**, plus three mandatory hands-on
stops. Disk extrapolated naively from the marginal rate would be ~27 GB, but
the published snapshot of the same pruned database is 23 GB extracted, so
~23 to 27 GB is the honest range. **I did not run path B to completion** and
I did not reach the first halt height.

Path B also needs two binaries: `v1.8.0-rc.1` cannot replay onyx from genesis
at all, because it refuses the non-canonical transaction encoding earlier
blocks hold. Started at genesis it stops within a few blocks with `wrong
Block.Header.AppHash`. *(Documented in onyx's own `UPGRADES.md`; I did not
reproduce that failure myself.)*

### The indexer

Measured twice, against two different nodes.

| | |
|---|---|
| Backfill rate, against a replaying node | ~780 blocks/s |
| Backfill rate, against a caught-up node | ~1,200 blocks/s |
| Cost to the node during a genesis replay | node fell from 66 to 44 blocks/s |
| Full backfill, 0 to 1,350,786 | **1,066 s** (17 m 46 s) |
| Indexer DB when done | **8.5 GB**, about 6.3 KB/block |
| Indexer RSS | peaked near 700 MB during backfill, 234 MB after |

The indexer is **never the bottleneck**: it only reads blocks and does not
execute them, so it runs an order of magnitude faster than the node's own
replay and tracks a caught-up node live.

### Where the 429 comes from

Not tx-indexer. The string `Per-IP rate limit exceeded; retry after 300
seconds.` appears nowhere in gnolang/tx-indexer at any commit in its history,
and the indexer's own limiters emit different text entirely
(`--http-rate-limit` answers `{"error": "too many requests"}`).

It is a **Cloudflare rate-limiting rule** defined in `gnolang/infrastructure`
(`tofu/edge/main.tf`, budgets in `tofu/edge/variables.tf`, rationale in
`docs/edge.md`):

| | |
|---|---|
| `indexer.` | 3,100 requests / 300 s, about 10.3 req/s |
| `rpc.` | 3,900 requests / 600 s, about 6.5 req/s |
| counted by | `[cf.colo.id, ip.src]`, `requests_to_origin` only |
| scope | **every network's names together**: Business plan cannot count per hostname |
| penalty | block for the rule's own period, 300 s / 600 s |

Full detail, with the quotes, in
[ENDPOINTS.md](../../deploy/onyx/ENDPOINTS.md).

---

## Proof that gnoscope's own queries answer

Full transcripts in [PROOF.md](../../deploy/onyx/PROOF.md). The short version:

**The fingerprint matches.** gnoscope's pool identifies a chain by
`chain_id + ":" + hash of block 1`. The local stack answers exactly what the
public indexer answers:

```
onyx-1:PGlpc0OZKeAItjmtZzNVJSzsmY40K9X7XAMCuTa6wj4=
```

The local node derived that hash by replaying the genesis file, so this is
evidence of the same chain, not just a plausible answer.

**gnoscope's real queries answer**, taken verbatim from
`gnoverse/gnoscope:pkg/indexer/indexer.go`: the pool probe, the `getBlocks`
range with `blockFields`, and the `getTransactions` selection with the full
message and event fragment set. A real onyx `MsgCall` to
`gno.land/r/gnops/valopers` came back with its arguments decoded and both
`StorageDepositEvent`s intact.

**One query needed gnoscope's own gating**, which is worth recording because it
looks like a failure and is not. Sent verbatim, gnoscope's `txFieldsLight`
selection 422s against tx-indexer v1.5.1, which defines no `MsgCreateSession`,
`MsgRevokeSession`, `MsgRevokeAllSessions` or `Signature` type. gnoscope
already handles this: `supportsTypeAt` probes one representative type per
fragment group and strips the groups the endpoint lacks. I reproduced the four
probes and re-sent the trimmed query, which answered. The public onyx indexer
would behave identically, being the same version.

**The RPC half answers too.** gnoscope reads balances and session grants from
RPC, not the indexer. `abci_query path="auth/accounts/<addr>"` against the local
node returned a correct `BaseAccount` with coins, pubkey, account number and
sequence.

---

## Configuration decisions worth a reviewer's attention

**Ports.** 26756 node p2p, 26757 node RPC, 8567 indexer, 8568 indexer metrics.
All loopback, none in the deployment host's in-use list. Note that **8546, the
tx-indexer default, is taken** on that host by the existing mainnet indexer, so
the `--listen-address` flag is not optional.

**p2p bound to 127.0.0.1 and it still syncs.** The listener only accepts
inbound connections; outbound dialing is unaffected. Verified: the node
discovered and gossiped with a dozen peers through PEX while bound to loopback.
The cost is that it never serves blocks back, which is acceptable for one
explorer backend and would not be for a fleet. `p2p.external_address` is
deliberately unset.

**Two additions to the sentry unit's hardening**, since the instruction was to
reproduce it rather than invent:

- `ReadWritePaths=`. `ProtectSystem=strict` makes the entire filesystem
  read-only except `/dev`, `/proc` and `/sys`, so the data directory must be
  named back in explicitly. **The mainnet sentry unit as given does not carry
  this line.** Either something not visible in that excerpt grants it write
  access, or that unit does not write where it appears to. Worth a look.
- `IPAddressDeny=any` / `IPAddressAllow=localhost` on the indexer unit only.
  The indexer's only peer is the node on loopback and its only client is
  gnoscope on loopback, so it has no reason to reach the network at all. This
  is an addition, not a reproduction; drop both lines if you would rather stay
  exactly with the reference.

**No rate limit on the local indexer.** Deliberately no `--http-rate-limit` and
no `--client-time-budget`: the only client is gnoscope over loopback, and a
limit there would reintroduce precisely the failure this node exists to remove.
Revisit if anything else is ever pointed at 8567.

---

## What I could not determine or could not test

- **Path B was not run to completion.** It reached height 174,063 of 1,349,724
  and never reached the first halt at 445,000, so **the halt-and-restart
  procedure in INSTALL.md §B2 is documented but unverified by me.** It comes
  from onyx's own `VALIDATOR.md` and `UPGRADES.md`, which state it was verified
  on 2026-11-17. The claim that `v1.8.0-rc.1` cannot replay from genesis is
  likewise taken from that documentation, not reproduced here.
- **The gnoscope side was never run.** I verified the queries gnoscope sends
  and that the local stack answers them, by reading its source and replaying
  its exact query text. I did not start a gnoscope process against this stack,
  so the pool-selection behaviour described in RESET.md, including the
  stale-node-poisons-the-pool trap, is read from
  `pkg/indexer/indexer.go` and **not** demonstrated.
- **Whether gnoscope is in fact one of the heavy clients** the rate-limit
  budgets were sized around is inference from the numbers in the infra
  variable descriptions, not confirmed.
- **The deployment host was never touched**, by design. Everything about it,
  including whether a `gnoland` user exists, whether `/data` is a suitable
  location, and whether the services start cleanly there, is untested. The
  units do **parse**: `systemd-analyze verify` on lyra accepted both, its only
  complaints being that `/usr/local/bin/gnoland-v1.8.0-rc.1` and
  `/usr/local/bin/tx-indexer-v1.5.1` are not installed here. They were never
  loaded or started by systemd anywhere. The node and indexer were run on lyra
  by invoking the same binaries with the same flags directly, so the command
  lines are exercised; the unit plumbing and the sandboxing directives around
  them are not.
- **Whether a Cloudflare rate-limit exemption is actually feasible** on the
  gno.land plan. The custom-rules ruleset has a `skip` precedent, but only for
  Browser Integrity Check, not for rate limiting. An IP exclusion in the
  counting expression should work; I did not confirm it with the plan's
  documentation.
- **Testnet rotation cadence.** I could not find a stated lifetime for a
  gemstone testnet. Onyx launched 2026-09-28, replacing pearl. The "every few
  months" in the cost table is a guess from that single data point.
- **The published snapshot's independence.** I verified its sha256 against the
  publisher's own `.sha256` file, which proves transfer integrity and nothing
  else. The strong check, that its state matches an independent replay from
  genesis, is exactly what path B would have provided and did not.

---

## Disk to reclaim

Everything lives under `~/ws/d-20261123-onyxnode/.work/`. Nothing was written
to `/tmp`.

| Path | Size | What |
|---|---|---|
| `.work/snap/` | 14 GB | the downloaded snapshot archive |
| `.work/onyx-node2/` | 23 GB | the restored node database, at the live tip |
| `.work/onyx-node/` | 2.2 GB | the genesis replay node, stopped at height 174,063 |
| `.work/indexer-db2/` | see below | the indexer database |
| `.work/src/` | ~2 GB | two git worktrees at the rc tags |
| `.work/bin/` | 170 MB | the three binaries |

```shell
rm -rf ~/ws/d-20261123-onyxnode/.work
cd ~/ws/d-20261123-onyxnode && git worktree prune
rm -rf ~/ws/d-20261123-onyxnode-txindexer
```

The `git worktree prune` matters: the two worktrees are registered in the main
checkout's `.git`, so deleting `.work` without it leaves stale entries.

---

## Final numbers

Everything below was observed, not estimated.

### The stack, caught up

| | |
|---|---|
| Local node height | **1,350,786** |
| Local indexer height | **1,350,786** |
| `rpc.onyx.testnets.gno.land` | **1,350,786** |
| Node `catching_up` | `false`, 7 peers, block time seconds behind wall clock |
| Chain fingerprint | `onyx-1:PGlpc0OZKeAItjmtZzNVJSzsmY40K9X7XAMCuTa6wj4=`, identical to the public indexer's |

All three heights equal, with the public RPC asked exactly once.

### Time, from nothing to a fully indexed stack

| Step | Measured |
|---|---|
| Build both gnoland binaries + tx-indexer | under 5 minutes, warm Go cache |
| Download 14.83 GB snapshot | **532 s** at 27.9 MB/s |
| sha256 verify | matched the publisher's `.sha256` |
| Extract 23 GB | **90 s** |
| Node start to RPC answering | **40 s** |
| Node to the live tip | seconds, once peers attached |
| **Node usable** | **about 25 minutes** |
| Indexer backfill, 0 to 1,350,786 | **1,066 s** (17 m 46 s) at a sustained ~1,280 blocks/s |
| **Everything usable** | **about 43 minutes** |

### Resources, at the tip

| | |
|---|---|
| Node data | **23 GB** |
| Indexer data | **8.5 GB** |
| **Steady-state total** | **31.5 GB** |
| Transient, the snapshot archive | 14 GB, deletable after extraction |
| Node RSS | **4.8 GB** |
| Indexer RSS | **234 MB** once backfilled; peaked near 700 MB during it |

The node's 4.8 GB is the number to size from. A node still replaying history
sat at 1.2 GB, which would badly under-provision. The deployment host has 54 GB
available and 703 GB free, so neither figure is close to a constraint.

### Disk used on lyra right now

```
23 G   .work/onyx-node2      restored node, at the tip
14 G   .work/snap            the snapshot archive, now redundant
8.5G   .work/indexer-db2     the indexer
2.2G   .work/onyx-node       genesis replay node, stopped at 174,063
299 M  .work/src             two git worktrees
158 M  .work/bin             three binaries
-----
~48 G total
```

Removal is in [Disk to reclaim](#disk-to-reclaim). Nothing was written outside
`~/ws/d-20261123-onyxnode` and `~/ws/d-20261123-onyxnode-txindexer`.

### Still running on lyra as of this report

The restored node (`gnoland-v1.8.0-rc.1`, ports 26758/26759) and the indexer
(port 8567/8568) were **left running** so the stack can be inspected. They are
plain foreground processes under the shell, not services, so they do not
survive a reboot and nothing will restart them. Kill them with:

```shell
pkill -f 'gnoland-v1.8.0-rc.1 start'
pkill -f 'tx-indexer-v1.5.1 start'
pkill -f sample2.sh
```

The genesis replay node was already stopped during the run.
