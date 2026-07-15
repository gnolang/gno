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
   in native Gno maps ("buckets"). A native map persists as a *single* object
   regardless of entry count (`MapListItem` carries no `ObjectInfo`), so any
   Get/Set/Remove loads a constant number of objects **independent of N**.

   Two design choices matter for cold gas, both settled here because they fix
   bucket placement and so cannot change after a map is persisted:

   - **Two-level buckets.** Buckets are not one flat array — a flat B-slot
     array is a single object every op must decode in full (~190KB at B=1024,
     fixed at any N). Instead they live in a directory of √B pages of √B
     buckets; an op decodes the directory, one page, and one bucket (√B-slot
     arrays), and pages are lazily allocated. This keeps the bucket count high
     (small leaves) without a large array. `New()` = 4096 buckets (64×64);
     `NewWithBuckets(n)` splits n evenly across the two levels.
   - **Native SHA-256 hashing.** Bucket index derives from `crypto/sha256`
     (low 64 bits), a calibrated native binding, rather than an interpreted
     FNV-1a loop metered per byte on every op.

   API is a minimal KV: `Get / Set / Remove / Has / Size / Iterate`, mirroring
   the avl.Tree signatures. Iteration is deterministic (page, bucket, then
   insertion order) but not key-sorted — avl remains the tool for sorted keys
   and range queries.

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

### Bucket hashing: native SHA-256, not interpreted FNV-1a

`bucketIndex` uses `crypto/sha256.Sum256` (low 64 bits) rather than an FNV-1a
loop written in Gno. FNV-1a is interpreted — a per-byte loop metered several VM
ops per byte, run ~4× per transfer over ~40-byte address keys. `Sum256` is a
calibrated **native** binding (a flat native charge), so it moves the hash off
the interpreter.

Measured end-to-end (cold grc20 transfer, production gas, this ADR's harness):

| holders | FNV-1a (reads/writes · gas) | SHA-256 (reads/writes · gas) | Δ |
|---|---|---|---|
| 20k | 78 / 8 · 7.40M | 82 / 8 · 6.89M | −0.51M |
| 100k | 78 / 8 · 7.53M | 82 / 8 · 7.06M | −0.48M |
| 1,000,000 | 78 / 8 · 9.13M | 82 / 8 · **8.58M** | **−0.55M** |

The saving is real but **smaller than a compute-only measurement suggests**.
Decomposed: the interpreted hash removed saves **−0.75M in the VM term**, but
calling native `crypto/sha256` pulls its package objects into the store cold on
first use — a fixed **+4 reads (+0.24M)** that FNV-1a (pure in-package Gno, zero
imports) never paid. Net is **−0.5M**, flat across scale. A "−0.9M" figure
counts only the VM term and omits the package cold-load — a warm-vs-cold trap.

Two consequences worth recording:

- **This must ship before any hashmap is persisted.** Placement is a function of
  the digest; changing the hash after state exists relocates every key. Hence it
  is in this PR, not a follow-up.
- The +4-read package tax is itself a data point: **fixed package/stdlib
  cold-loads are a measurable slice of the O(1) read floor** (now ~82 objects).
  Shrinking that floor — packing immutable package code, or realm-native KV — is
  a larger lever than bucket tuning, and the next thing to profile.

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
- **With a *flat* array, do NOT over-size buckets:** 4096-flat was *worse* than
  1024-flat at every size, because every op decodes the whole bucket-pointer
  array once — a 4× larger array is a fixed tax that outweighs the smaller
  per-bucket decodes. **This is exactly what the two-level directory fixes**
  (see below): by splitting the array, more buckets no longer means a larger
  array to decode, so 4096 becomes the better default. This finding is what
  motivated the two-level design.
- **bptree keeps ordering but costs ~2× hashmap at scale** (−37% vs avl at 1M);
  the gap widens with N as a transfer's 4 traversals plus node-split writes add
  up. It is the right avl replacement where sorted access is required, not
  where flat cost is the goal.

### Two-level directory: the shipped design

The findings above (flat array tax + bucket bloat) were resolved by making v0's
buckets two-level. Profiling one cold 1M-holder grc20 transfer by object type
(temporary `ReadProfileHook` in the store) shows *where* the flat design spent
its decode gas, and that the entire 20k→1M growth is the buckets swelling:

| REALM object (1M transfer) | flat 1024 | two-level 64×64 | drop |
|---|---|---|---|
| `ArrayValue` (bucket directory) | 190,595 B → 571,785 gas | 39,454 B → 118,362 gas | −0.45M |
| `MapValue` (leaf buckets) | 281,712 B → 845,136 gas | 70,758 B → 212,274 gas | −0.63M |
| ledger decode total | **1.42M** | **0.33M** | **−1.09M** |

The flat 1024-slot directory decodes ~190KB on *every* op regardless of N; the
two-level split (dir + one page, √-sized) cuts that, and the higher usable bucket
count (4096) shrinks each leaf from ~977 to ~244 entries. Full grc20 transfer
cost (cold, production gas, incl. the flat 59k/read and sha256 hashing):

| holders | flat hashmap 1024 | **two-level v0 (4096)** | vs flat | vs avl |
|---|---|---|---|---|
| 20,000 | 6.89M | **6.17M** | −0.72M | — |
| 100,000 | 7.53M | **6.22M** | −1.31M | — |
| 1,000,000 | 8.58M | **6.59M** | **−1.99M** | **≈−65%** |

The −1.99M is ~−1.09M decode (above) plus ~−0.9M alloc gas on the same shrunken
objects, for +3 cold reads (the extra directory/page loads). Cost is nearly flat
in N (6.17M → 6.59M across 50×), versus flat-hashmap's 6.89M → 8.58M. Beyond the
ledger, a fixed ~200-object *package/code floor* remains (stdlib+`/p/` blocks,
identical across backends and N) — the next lever, but VM-level (package packing
/ realm-native KV), out of scope here.

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
