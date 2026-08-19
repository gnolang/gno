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
cache GC), on the `chain/test13` tag and on master (identical results). The
count of cold object reads is what grows with the ledger:

| ledger size (holders) | objects read (cold) |
|---|---|
| 1 | 82 |
| 2,000 | 166 |
| 20,000 | 182 |
| real `test_atone` on test13 (observed on-chain) | ~310 |

The per-operation cost is a property of the storage data structure, not of the
contract code: the same transfer over the same grc20 code loads 82 objects when
the ledger is small and ~310 on the ledger that produced the on-chain 18.3M
figure. Gas figures are in "Measured results" below — this intro deliberately
reports only read *counts*, because the gas model is subtler than a per-read
constant (see the note there).

(Absolute counts differ slightly between harnesses — e.g. avl@20k is 191 reads
in the refined validation runs, not 182 — but the growth trend is the same.)

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
     `NewWithBuckets(n)` splits n across the two levels.
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

   This holds only if no *test* file in the grc20 package imports a backend
   either: `MsgAddPackage` type-checks a package as `MPUserAll`, test files
   included, so a `_test.gno` importing hashmap/bptree would pull both into
   the on-chain closure of every realm that imports grc20 (measured: +52k gas
   on each such deploy). The multi-backend coverage therefore lives in
   `grc20/filetests/storage_option_filetest.gno` — `filetests/` has no
   `gnomod.toml`, so it is never deployed and its imports stay out of the
   closure. Do not move it back into the package.

## Measured results (validation)

**How these were measured.** The cold transfer is run through the *real*
cache-wrapped, gas-metered store — `DefaultGasConfig` plus `DefaultParams`
depth params, exactly what a gno.land node charges (`app.go` `WithGasConfig`) —
and the number is `GasConsumed()`. This matters: an earlier draft of this ADR
*simulated* production cost as `VM_gas + reads×59k + writes×24k`, which is only
2 of the store's **7** gas dimensions and omits the depth multiplier entirely,
undercounting by **~2×**. A maintainer's independent gnodev reproduction caught
it. Every figure in this section is real metering; the superseded simulated
runs are quarantined in the appendix and none of their absolutes are quoted
here.

Cold transfer, real metering (fresh keeper = validator after cache GC / under
congestion — a *warm* transfer is far cheaper; see "Cold vs warm"):

| holders | avl | two-level hashmap (shipped) | bptree | hashmap reduction |
|---|---|---|---|---|
| 20,000 | 27.9M | **16.2M** | 22.2M | **−42%** |
| 100,000 | 29.8M | **16.5M** | 26.3M | **−45%** |
| 1,000,000 | ~35M* | **18.8M** | — | **−46%** |

`*` avl 1M extrapolated from 27.9 / 29.8 / 32.0M at 20k / 100k / 200k.

**−42% to −46% is the headline result of this PR**, and the figure to quote.
The appendix reports a larger −69% for the same backend; that run isolates the
data structure with no grc20 package graph, so it excludes the ~10.7M package
floor that a real transfer always pays. It is reconciled where it appears.

- **A fixed ~10.7M package-code "depth floor" dominates every cold transfer —
  identical across all backends.** It is the cold, depth-multiplied load of the
  grc20/realm package graph; the backend only moves the *ledger* reads on top of
  it. That's why even hashmap can't drop below ~16M cold and the win is smaller
  than a data-structure-only view suggests. Shrinking that floor (packing
  immutable package code / realm-native KV) is now the biggest lever.
- hashmap stays ~flat (16.2 → 18.8M over 20k → 1M); avl climbs (27.9 → ~35M).

### Cold vs warm

The depth/flat/per-byte reads are charged **only on a cache miss**
([cache/store.go](tm2/pkg/store/cache/store.go): `DepthReadFlat` etc. inside the
`!ok` branch). So a **warm** transfer (package already cached — e.g. a `txtar`
test that deploys and transfers in the same session) skips almost all of it and
costs **~7M**; a **cold** one (validator after GC, or congestion churning the
cache) pays it in full (~16–28M). #5906's congestion is the **cold** case, which
is why the issue is real there even though casual/warm testing looks cheap. The
on-chain test13 `test_atone` = 18.3M is a cold-config number at a **modest**
ledger, not a 1M one: a fully-cold 1M avl transfer is ~35M. (Under the
simulated harness those two numbers happened to coincide, which made an
earlier draft read `test_atone` as confirmation of a 1M-holder ledger. It was
a coincidence of the ~2× undercount, and that claim has been withdrawn.)

### Bucket hashing: native SHA-256, not interpreted FNV-1a

Bucket placement (`locate`/`hash64`) uses `crypto/sha256.Sum256` (low 64 bits)
rather than an FNV-1a loop written in Gno. FNV-1a is interpreted — a per-byte
loop metered several VM ops per byte, run ~4× per transfer over ~40-byte
address keys. `Sum256` is a calibrated **native** binding (a flat native
charge), so it moves the hash off the interpreter.

Measured under real gas metering: two-level SHA-256 (16.18M) vs an
otherwise-identical two-level FNV-1a build (16.92M) at 20k holders =
**−0.74M**, with *identical* store gas — depth, flat and per-byte all equal, so
the whole difference is the interpreted FNV-1a CPU loop.

There is no offsetting package penalty. An earlier draft assumed calling
`crypto/sha256` would pull its package objects into the store cold on first use
(+4 reads / +0.24M) and netted the win down to −0.5M; that is wrong, because
`crypto/sha256` is stdlib and served from the in-memory byte cache, which pays
no depth or flat charge. **−0.74M, flat across scale, is the figure.**

Two consequences worth recording:

- **This must ship before any hashmap is persisted.** Placement is a function of
  the digest; changing the hash after state exists relocates every key. Hence it
  is in this PR, not a follow-up.
- Fixed package cold-loads dominate what is left: the ~10.7M package-code depth
  floor is far larger than any bucket-tuning win. Shrinking it — packing
  immutable package code, or realm-native KV — is the next thing to profile.

### Design rationale: comparisons on the superseded harness

> The two subsections below predate the metering correction, and their absolute
> gas figures are the simulated ones — **~2× low, and not comparable to the
> corrected table above.** They are kept because the comparisons *within* each
> table are like-for-like (one harness, one op shape) and are what drove the
> design decisions. Read the deltas, ignore the absolutes.

#### Scaling to 1,000,000 entries (KV backends isolated)

To confirm the O(1) claim at realistic scale and compare against the in-tree
B+ tree, the same cold-cache harness was run on the KV backends directly
(arbitrary string keys, batched in-gno seeding — no grc20 package graph, so a
full token transfer adds ~0.5–1M of fixed grc20/package overhead on top; the
*slope* is the data structure's).
Op is transfer-shaped (2 Get + 2 Set):

| backend | 20k | 100k | 1,000,000 | ledger-only Δ @1M | ordered? |
|---|---|---|---|---|---|
| avl | 13.8M | 15.5M | 18.6M | — | yes |
| **hashmap (1024 buckets)** | 4.3M | 4.5M | 5.7M | **−69%** | no |
| hashmap (4096 buckets) | 7.8M | 7.9M | 8.2M | −56% | no |
| bptree (fanout 128) | 7.8M | 10.7M | 11.7M | −37% | yes |

**Why −69% here and −46% in the headline table.** These runs isolate the KV
backends: no grc20 package graph, so the ~10.7M package-code depth floor that
every real transfer pays is absent. Removing a fixed cost from both sides of a
ratio inflates it. −69% is the *data structure's* reduction; **−46% is what a
real cold grc20 transfer sees**, and −46% is the number that should be quoted.

Findings:

- **hashmap stays flat to 1M:** 4.3M → 5.7M across a 50× state increase. The
  bucket-bloat concern (≈977 entries/bucket at 1M with 1024 buckets) is real
  but mild (~+1.4M).
- **With a *flat* array, do NOT over-size buckets:** 4096-flat was *worse* than
  1024-flat at every size, because every op decodes the whole bucket-pointer
  array once — a 4× larger array is a fixed tax that outweighs the smaller
  per-bucket decodes. **This is exactly what the two-level directory fixes**
  (see below): by splitting the array, more buckets no longer means a larger
  array to decode, so 4096 becomes the better default. This finding is what
  motivated the two-level design.
- **bptree keeps ordering but costs ~2× hashmap at scale**; the gap widens with
  N as a transfer's 4 traversals plus node-split writes add up. It is the right
  avl replacement where sorted access is required, not where flat cost is the
  goal. (End to end under real metering it lands at 22.2M / 26.3M vs avl's
  27.9M / 29.8M — see the headline table.)

#### Two-level directory: the shipped design

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

| holders | flat hashmap 1024 | **two-level v0 (4096)** | vs flat |
|---|---|---|---|
| 20,000 | 6.89M | **6.17M** | −0.72M |
| 100,000 | 7.53M | **6.22M** | −1.31M |
| 1,000,000 | 8.58M | **6.59M** | **−1.99M** |

The −1.99M is ~−1.09M decode (above) plus ~−0.9M alloc gas on the same shrunken
objects, for +3 cold reads (the extra directory/page loads). Cost is nearly flat
in N (6.17M → 6.59M across 50×), versus flat-hashmap's 6.89M → 8.58M. The
flat-vs-two-level delta is like-for-like — both columns are the same harness,
so the choice it justifies stands even though the absolutes are ~2× low.
Beyond the ledger, the fixed package/code floor remains, and it is the larger
lever — but a VM-level one (package packing / realm-native KV), out of scope
here.

### Storage footprint (persisted bytes / deposit)

The backend choice affects not only access **gas** but also persisted
**storage** — and therefore the storage *deposit* (100 ugnot/byte) a realm must
lock. Measured with the same harness via `rlm.Storage` after commit (grc20
token realm; deploy base = empty ledger, then N holders minted):

| holders | avl (bytes) | hashmap (bytes) | bptree (bytes) |
|---|---|---|---|
| 0 (deploy) | 9,321 | 23,119 | 9,508 |
| 20,000 | 41,315,124 | **4,419,897** | 11,805,414 |
| 100,000 | 208,629,782 | **15,976,554** | 58,578,909 |
| 1,000,000 | ~2.09B* | **145,588,842** | panic† |

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
- **hashmap is −89% storage at 20k, −92% at 100k, −93% at 1M** vs avl (4.4M vs
  41M bytes; 16.0M vs 209M; 146M vs ~2.09B). Marginal cost is ~144 bytes/holder,
  identical to a flat map; the two-level directory's larger bucket count (4096)
  adds a one-time object overhead that fully amortizes by 1M (145.6M ≈ a
  flat-map's 144.5M). Its deploy base is a small **23 KB** — pages and buckets
  are allocated lazily on first write, not pre-allocated, so an empty ledger
  costs almost nothing.
- **Deposit follows directly** (bytes × 100 ugnot): a 100k-holder avl ledger
  locks **~20,863 GNOT** of storage deposit vs **~1,598 GNOT** for hashmap.
- **bptree is the ordered middle ground:** ~4× hashmap's bytes but still −72% vs
  avl at 100k, while keeping sorted iteration.

## Alternatives considered

- **Higher-fanout B-tree** — the in-tree `p/nt/bptree/v0` already satisfies the
  `KV` interface (`NewBPTreeN(fanout)`), so it drops into the same
  `WithStorage` seam. Under real metering it is **−20% vs avl at 20k and −12%
  at 100k** (22.2M / 26.3M against 27.9M / 29.8M — headline table) and,
  crucially, **preserves ordered iteration** — making it the recommended avl
  replacement for state that needs sorted access (registries, DAOs, gnoswap
  ticks/positions), while hashmap is reserved for pure-lookup ledgers. Still
  O(log N) objects, so its advantage narrows as N grows, where hashmap's stays
  flat.
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
- **The seam is not free for tokens that never opt in.** `balances`/
  `allowances` were inlined `avl.Tree` values in `PrivateLedger`; behind an
  interface they become separately persisted objects, so every grc20 call pays
  two extra cold object loads. Measured on `wugnot.txtar` against master:

  | tx | master | this branch | Δ |
  |---|---|---|---|
  | wugnot Deposit | 6,981,080 | 7,331,636 | +350,556 (+5.0%) |
  | wugnot Transfer | 5,382,444 | 5,732,382 | +349,938 (+6.5%) |
  | wugnot Withdraw | 8,000,902 | 8,347,817 | +346,915 (+4.3%) |
  | addpkg importing grc20 | 7,914,971 | 8,000,571 | +85,600 (+1.1%) |

  Of the per-call delta, ~313k is the indirection itself: a control build with
  concrete `*avl.Tree` fields and no interface measured the same, so it is the
  extra objects rather than interface dispatch, and it cannot be tuned away
  while keeping the seam. The remaining ~37k is the constructor guards and the
  package source they add, which is charged per call because the package block
  is read cold. The whole thing is ~3% of what a cold transfer at scale saves,
  but it is charged to every default-avl token, so it is recorded here rather
  than left implicit.
- Hashmap-backed state loses key-ordered iteration; consumers that render
  sorted listings must keep avl or maintain a separate index. `Iterate` has no
  cursor either, so a hashmap-backed collection cannot be paginated: a consumer
  can stop early but not resume, making a full paged listing O(n²). Realms that
  need paging should stay on avl or bptree.
- Bucket count is fixed at construction; the default is **4096**, split across
  the two-level directory. Because the array-decode tax is now paid on small
  √-sized levels rather than one flat array, a higher bucket count is cheap and
  stays flat from ~1k to ~1M entries; guidance is in the package doc.
  Progressive splitting (linear hashing) is possible as a v2 without API changes
  if maps far exceed ~1M entries.
- The bucket hash (SHA-256) is unkeyed, so keys are grindable — though grinding
  SHA-256 preimages is far costlier than inverting a non-cryptographic hash: a party
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
- The `WithStorage` constructor contract ("a fresh, empty store per call") is
  enforced, not just documented, because both ways of breaking it fail
  silently rather than loudly:
  - one *shared* store for both maps still resolves balances and allowances
    correctly (the key spaces do not collide), but `KnownAccounts()` — and so
    `RenderHome()` — counts allowance rows as holders;
  - a *pre-populated* store yields balances that `totalSupply` never accounted
    for, i.e. a token reporting a holder balance with zero supply, straight
    out of the constructor.

  `WithStorage` rejects a nil constructor, and `NewToken` rejects a
  constructor that returns nil, returns the same instance twice, or returns a
  non-empty store (`ErrNilStorage` / `ErrSharedStorage` /
  `ErrNonEmptyStorage`). The cost is two `Size()` calls on empty stores, once
  per token creation — nothing on the per-call path.
- Persistence is covered end to end by
  `gno.land/pkg/integration/testdata/grc20_kv_backend_restart.txtar`: a
  hashmap-backed ledger is read *and written* across a `gnoland restart` by a
  realm whose package closure never includes hashmap (it resolves the token
  through `grc20reg`). This is the seam's sharpest exposure — a caller-supplied
  concrete type behind an interface field in realm state, reloaded cold and
  reached from a package that cannot name it.
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
