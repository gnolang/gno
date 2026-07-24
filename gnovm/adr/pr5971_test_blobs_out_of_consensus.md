# Exclude package test/filetest blobs from consensus state

## Status

Accepted (implemented in PR #5971). Consensus-breaking; activation rides a
coordinated node upgrade / fresh testnet.

## Context

When a Gno package is deployed, the chain stores its files in the VM's package
store. PR #5891 ("split mempackage storage into prod and test blobs") split each
stored `MP*All` package into two keys so that importers no longer read/typecheck
a dependency's test files:

- **prod blob** at `pkg:<path>` — non-test `.gno` files plus non-`.gno` files;
- **`#allbutprod` sibling** at `pkg:<path>#allbutprod` — `_test.gno` /
  `_filetest.gno` files (and, for a prod-less package, its non-`.gno` files).

Both blobs were written to the **merkleized** `iavlStore`. The VM store has two
backends (see `gnovm/pkg/gnolang/store.go`):

- `iavlStore` — merkleized; its root feeds the multistore commit `AppHash`.
- `baseStore` — a `dbadapter` store; persisted and queryable but **not**
  merkleized. `dbadapter.Commit()` returns a nil hash, and rootmulti's
  `commitStores` folds each store's `CommitID.Hash` into a simple Merkle map,
  so a nil-hash store contributes a constant regardless of its contents.

Realm object *bytes* already live in `baseStore`, with only their *hashes*
escaped into `iavlStore`. Mempackage blobs were the exception: written whole,
directly into `iavlStore`.

### Problem

Because the `#allbutprod` sibling lived in `iavlStore`, changing a package's
**test file only** — which has zero effect on execution — shifted the committed
multistore Merkle root. That makes an edit to a `_test.gno` file a
consensus-breaking change.

This was surfaced by PR #4008: adding one `bc1…` row to
`gnovm/stdlibs/chain/address_test.gno` moved the pinned `TestAppHashCrossrealm38`
hash (`058910b2…` → `d68f032b…`), even though `Address.IsValid()` behavior was
unchanged. Test/filetest files are not execution state and must not participate
in consensus.

## Decision

Store the `#allbutprod` (test/filetest) sibling in the non-merkleized
`baseStore` instead of `iavlStore`. The prod blob stays in `iavlStore` (it is
importable code and legitimately part of consensus state). Test files remain
persisted and fully queryable for tooling (`GetMemPackageAll`, `vm/qfile`,
explorers) but no longer enter the `AppHash`.

Concretely (`gnovm/pkg/gnolang/store.go`):

- `AddMemPackage` — prod → `iavlStore`, `#allbutprod` → `baseStore`.
- `setMemPackageBlob(dst, key, mpkg)` — destination store is now a parameter.
- `DeleteMemPackage` — deletes the sibling from `baseStore`.
- `getMemPackageAllButProd` — reads the sibling from `baseStore`.
- `FindPathsByPrefix` — sorted **two-store merge** of iavl prod keys and base
  sibling keys, de-duped by suffix-trimmed path, so a package is enumerated
  exactly once and a prod-less (test-only) package — whose blob exists only as a
  sibling — is still listed. This is correct both when the two stores are
  distinct (on-chain: `baseKey` = `dbadapter`, `iavlKey` = bptree) and when they
  are the same underlying store (tooling: `NewStore(_, base, base)`, where the
  two iterators stay in lockstep and every key compares equal).
- `debugger.go` `isMemPackage` — checks the sibling in `baseStore`.

Only `pkg:` keys fall in the `FindPathsByPrefix` scan range; object/type/node/
index keys use disjoint prefixes (`oid:` / `tid:` / `node:` / `pkgidx:`), all of
which sort outside `[pkg:\x00, pkg:\xFF]`, so the baseStore scan sees siblings
only.

## Consequences

- **Consensus-breaking (intended).** Dropping every existing package's test blob
  out of the Merkle root changes the `AppHash`. It must ship in a coordinated
  node upgrade / fresh testnet, never hot-applied to a live chain. The pinned
  `expectedCrossrealm38Hash` in `apphash_crossrealm38_test.go` was re-derived
  (`058910b2…` → `4ffebf22…`) deliberately for this reason.
- **Test-file edits are now consensus-neutral.** Verified: with this change,
  re-adding the PR #4008 test line leaves `TestAppHashCrossrealm38` unchanged.
- **Storage deposits are unaffected.** Deposits are computed from
  `RealmStorageDiffs()` (realm object byte deltas), not from mempackage blobs, so
  moving the test blob between stores does not change any deposit. Confirmed by
  `TestTestdata` passing unchanged.
- **Durability / state-sync.** `baseStore` is persisted to the same node DB and
  written deterministically by every block-executing node. A node that restores
  from a state snapshot (rather than replaying blocks) may lack historical test
  blobs; since they are read only on query/tooling paths and never during
  execution, this cannot cause consensus divergence.
- **Encode gas is still charged** for writing the sibling (deterministic in the
  blob length), so per-tx gas accounting is unchanged.

## Alternatives considered

1. **Bump the pinned hash and keep test blobs in `iavlStore`.** Rejected: it
   accepts test-file edits as permanently consensus-breaking — every future
   stdlib/example test change would require a coordinated upgrade.
2. **Stop storing test blobs on-chain entirely.** Rejected: #5891 deliberately
   keeps them so tooling/explorers can serve the whole package. `baseStore`
   preserves that while removing them from consensus.
3. **Store only a hash of the test blob in `iavlStore`** (mirroring object
   escaping). Rejected as unnecessary: the goal is to remove test files from
   consensus, and a hash would still make test-file edits shift the root.
