# ADR: `p/nt/hashmap` — O(1)-object keyed storage, and an opt-in grc20 ledger backend

## Status

Proposed.

## Context

gnolang/gno#5906 reports test13 saturating its 3B block-gas limit during the
GnoSwap trading competition. Investigating where transaction gas actually goes
showed that a plain GRC20 transfer's cost is dominated by **loading persisted
objects from the store**, and that this cost **grows with the size of the
token's ledger**:

- In the GnoVM, every struct, array, and map value is a separate persisted
  object; the base store charges a flat 59,000 gas per object read
  (calibrated to ~59µs of random-read latency, `tm2/pkg/store/types/gas.go`).
- `avl.Tree` materializes its search path as one object per node, so a lookup
  on an N-entry tree loads O(log N) objects. grc20's `Transfer` performs ~4-5
  tree traversals (two balance reads, two writes re-walking the path).
- On a long-running node, old tree nodes are evicted from the object cache, so
  those loads are cold and each pays the flat read cost.

Measured with a two-keeper cold-cache harness (build state, commit, then
measure one `Transfer` from a fresh keeper — the state a validator is in after
cache GC), on the `chain/test13` tag and on master (identical results):

| ledger size (holders) | objects read (cold) | est. production gas |
|---|---|---|
| 1 | 82 | ~5.9M |
| 2,000 | 166 | ~13.5M |
| 20,000 | 182 | ~14.8M |
| real `test_atone` on test13 (observed on-chain) | ~310 | 18.3M |

At ~72M gas per swap and 15-18M per transfer, the 3B block fits ~10 swaps/sec —
the ceiling the issue is about. The per-operation cost is a property of the
storage data structure, not of the contract code: the same transfer over the
same grc20 code is ~5.9M when the ledger is small.

## Decision

1. **Add `p/nt/hashmap` (v0)**: a persistent string-keyed map storing entries
   in a fixed, power-of-two-sized array of native Gno maps ("buckets"). A
   native map persists as a *single* object regardless of entry count
   (`MapListItem` carries no `ObjectInfo`), so any Get/Set/Remove loads a
   constant number of objects — Map struct, bucket array, one bucket —
   **independent of N**. Bucket index is FNV-1a(key) masked to the bucket
   count; iteration is deterministic (bucket order, insertion order within a
   bucket) but not key-sorted.

   API is a minimal KV: `Get / Set / Remove / Has / Size / Iterate`, mirroring
   the avl.Tree signatures for those methods. Bucket count is fixed at
   construction (`New()` = 1024, `NewWithBuckets(n)`), with sizing guidance in
   the package doc. No ordered iteration, by design — avl remains the tool for
   sorted keys and range queries.

2. **Add an opt-in storage backend to grc20**: a `KV` interface (satisfied by
   both `*avl.Tree` and `*hashmap.Map`), a `WithStorage(func() KV)` option on
   `NewToken`, and interface-typed `balances`/`allowances` fields on
   `PrivateLedger`. The default remains `avl.Tree` — no behavior change for
   existing consumers. grc20 does **not** import hashmap (the constructor is
   caller-supplied), so the default token's package closure is unchanged.

## Measured results (validation)

Same cold-cache harness, master, the same grc20 token realm with avl vs
hashmap backends (objects = cold store loads; gas = estimated production
DeliverTx incl. 59k/read + 24k/write):

| ledger size | avl: objects / gas | hashmap: objects / gas | reduction |
|---|---|---|---|
| 1 | 86 / 6.18M | 76 / 6.59M | −7% (slightly worse) |
| 2,000 | 170 / 13.75M | **77 / 7.21M** | **−48%** |
| 20,000 | 186 / 15.06M | **77 / 7.33M** | **−51%** |

- The O(1) claim holds exactly: hashmap reads a constant **77 objects** at
  every ledger size, vs avl's growth (86 → 186). Cost is flat within **1.7%**
  from 2k → 20k holders.
- At a tiny ledger the hashmap is marginally worse (bucket-array decode
  overhead ≈ +0.4M); it wins from a few hundred entries up. This matches the
  intended positioning: opt in for ledgers expected to grow.
- Note: introducing the `KV` interface costs the **default avl path ~4 extra
  object loads** (~0.25M, +1.7% at 20k) because the trees are now held by
  pointer behind the interface rather than inline in `PrivateLedger`. Called
  out for reviewers; judged acceptable against the opt-in's −51%.

### Scaling to 1,000,000 entries (KV backends isolated)

To confirm the O(1) claim at realistic scale and compare against the in-tree
B+ tree, the same cold-cache harness was run on the KV backends directly
(arbitrary string keys, batched in-gno seeding — no grc20 package graph, so
subtract ~5.5M for a full token transfer; the *slope* is the data structure's).
Op is transfer-shaped (2 Get + 2 Set):

| backend | 20k | 100k | 1,000,000 | vs avl @1M | ordered? |
|---|---|---|---|---|---|
| avl | 13.8M | 15.5M | **18.6M** | — | yes |
| **hashmap (1024 buckets)** | 4.3M | 4.5M | **5.7M** | **−69%** | no |
| hashmap (4096 buckets) | 7.8M | 7.9M | 8.2M | −56% | no |
| bptree (fanout 128) | 7.8M | 10.7M | 11.7M | −37% | yes |

Findings:

- **End-to-end validation:** avl at 1M entries costs **18.6M**, matching the
  real on-chain `test_atone` transfer (18.3M) — independent confirmation that
  avl depth on a ~1M-entry ledger is what produces the number this ADR set out
  to explain.
- **hashmap stays flat to 1M:** 4.3M → 5.7M across a 50× state increase
  (−69% vs avl). The bucket-bloat concern (≈977 entries/bucket at 1M with 1024
  buckets) is real but mild (~+1.4M).
- **Do NOT over-size buckets** (corrects an earlier assumption): 4096 buckets
  is *worse* than 1024 at every size, because every op decodes the whole
  bucket-pointer array once — a 4× larger array is a fixed tax that outweighs
  the smaller per-bucket decodes. 1024 is a good default from ~1k to ~1M
  entries; only raise it past ~1M when per-bucket decode finally dominates.
  The package sizing guidance was updated to match.
- **bptree keeps ordering but costs ~2× hashmap at scale** (−37% vs avl at 1M);
  the gap widens with N as a transfer's 4 traversals plus node-split writes add
  up. It is the right avl replacement where sorted access is required, not
  where flat cost is the goal.

### Storage footprint (persisted bytes / deposit)

The backend choice affects not only access **gas** but also persisted
**storage** — and therefore the storage *deposit* (100 ugnot/byte) a realm must
lock. Measured with the same harness via `rlm.Storage` after commit (grc20
token realm; deploy base = empty ledger, then N holders minted):

| holders | avl (bytes) | hashmap (bytes) | bptree (bytes) |
|---|---|---|---|
| 0 (deploy) | 9,321 | 169,967 | 9,508 |
| 20,000 | 41,315,124 | **3,343,815** | 11,805,414 |
| 100,000 | 208,629,782 | **14,870,242** | 58,578,909 |
| 1,000,000 | ~2.09B* | **144,473,034** | panic† |

`*` avl 1M inferred from its flat slope (the 1M avl seed is impractically slow
to build). `†` bptree 1M hit the same unrelated cold-reboot VM panic seen at
grc20 scale; not a storage property.

Marginal cost **per holder** — the number that matters at scale:

| backend | bytes/holder | why |
|---|---|---|
| avl | **~2,090** | each entry is a full persisted tree node: balance + height + two child pointers + object metadata |
| bptree | ~585 | entries packed into shared leaf nodes, plus node objects + ordering metadata |
| **hashmap** | **~144** | just a native-map entry (key→balance); the bucket map is one object, entries add raw bytes only |

Findings:

- **The gas-heaviest structure is also the storage-heaviest.** avl materializes
  one persisted object per entry, so it pays both the cold-load gas *and* ~14×
  hashmap's bytes per holder — same root cause, both axes.
- **hashmap is −92% storage at 20k and −93% at 100k** vs avl (3.3M vs 41M bytes;
  14.9M vs 209M). Its 170 KB fixed deploy overhead (1024 pre-allocated bucket
  maps) is paid back after **~85 holders** — negligible for any real ledger.
- **Deposit follows directly** (bytes × 100 ugnot): a 100k-holder avl ledger
  locks **~20,863 GNOT** of storage deposit vs **~1,487 GNOT** for hashmap.
- **bptree is the ordered middle ground:** ~4× hashmap's bytes but still −72% vs
  avl at 100k, while keeping sorted iteration.

## Alternatives considered

- **Higher-fanout B-tree** — the in-tree `p/nt/bptree/v0` already satisfies the
  `KV` interface (`NewBPTreeN(fanout)`), so it drops into the same
  `WithStorage` seam. Measured (table above): at fanout 128 it is −37% vs avl
  at 1M and, crucially, **preserves ordered iteration** — making it the
  recommended avl replacement for state that needs sorted access (registries,
  DAOs, gnoswap ticks/positions), while hashmap is reserved for pure-lookup
  ledgers. Still O(log N) objects, so ~2× hashmap's cost at scale.
- **Combined get+set traversal in avl.Tree** — halves the constant factor,
  worth doing independently, but does not remove the growth with N.
- **One giant native map** — a single object, but every op decodes and every
  write re-serializes the entire ledger (~14M/write at 20k entries). Rejected.
- **Native VM keyed storage (EVM-slot style)** — the long-term answer but
  requires VM/persistence changes and a hard fork. This package is the
  no-fork solution available to contracts today.
- **Block-level warm/cold gas pricing** (first object access in a block cold,
  subsequent accesses cheap — deterministic since all validators execute the
  same block) — orthogonal protocol-level improvement that would additionally
  help under congestion; left as a separate proposal.

## Consequences

- Token ledgers (and other write-heavy KV state) can opt into flat, ~O(1)
  storage-access gas, removing the "gas grows with holder count" failure mode
  observed on test13.
- The `KV` seam is backend-agnostic: the same `WithStorage` option accepts
  `*avl.Tree`, `*hashmap.Map`, and `*bptree.BPTree` unchanged. The recommended
  choice is by access pattern — hashmap for pure-lookup ledgers, high-fanout
  bptree where ordered iteration is needed, avl for small/bounded collections.
- Hashmap-backed state loses key-ordered iteration; consumers that render
  sorted listings must keep avl or maintain a separate index.
- Bucket count is fixed at construction. Measured, the default 1024 stays flat
  from ~1k to ~1M entries and over-sizing is counter-productive (the array
  decode tax dominates); guidance is in the package doc. Progressive splitting
  (linear hashing) is possible as a v2 without API changes if maps far exceed
  ~1M entries.
- The bucket hash (FNV-1a) has a fixed seed, so keys are grindable: a party
  that can create many funded entries could concentrate them in one bucket to
  raise the per-op cost of keys hashing there. The effect is bounded by that
  bucket's size (not global) and is inherent to any fixed hash in a
  deterministic VM. Mitigated by adequate bucket counts; documented in the
  package. Realms where adversarial key concentration is a real threat should
  stay on avl.Tree (O(log N) regardless of distribution).
- `PrivateLedger.balances/allowances` change from concrete `avl.Tree` to the
  `KV` interface; `NewToken` gains variadic options. Both are source-compatible
  for all existing callers (fields are unexported; the only construction site
  is `NewToken`).
- `KV` is deliberately defined in grc20 rather than a shared package: a shared
  `p/nt/kv` import would enter every token's package closure (a per-call gas
  cost for all tokens, including default avl ones), and Gno interfaces are
  structural, so other consumers (registries, DAOs) can declare an identical
  interface and interoperate with the same backends without a shared named
  type. If several consumers converge on the pattern, lifting the interface
  into a neutral package later is additive and non-breaking.

## Notes

AI-assisted: investigation, design, implementation and measurements done with
Claude Code; reviewed and owned by the human author.
