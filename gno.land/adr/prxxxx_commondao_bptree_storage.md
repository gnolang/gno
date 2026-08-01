# ADR: commondao/bylaws persisted sets move from AVL to B+ tree

## Context

Persisted-storage measurements (filetests storing N distinct 40-char
addresses; storage deltas from the test runner) showed the AVL backing
costs ~1.9× the bytes of a fanout-32 B+ tree at every size, because AVL
persists one node object per entry while the B+ tree packs ~32 entries
per node (fewer object headers and refs):

| items | addrset (avl) | bptree32 | avl B/item | bptree B/item | ratio |
|------:|--------------:|---------:|-----------:|--------------:|------:|
| 10 | 18,328 B | 11,497 B | 1,833 | 1,150 | 1.6× |
| 100 | 193,095 B | 104,185 B | 1,931 | 1,042 | 1.9× |
| 1,000 | 1,953,508 B | 1,021,525 B | 1,954 | 1,022 | 1.9× |
| 10,000 | 19,683,514 B | 10,305,748 B | 1,968 | 1,031 | 1.9× |

Insert gas follows the same shape (10k adds: 7.9B vs 2.9B, 2.7×). The
branch's growth-prone sets — councils, per-proposal electorate
snapshots (one per proposal), voting records, bylaws documents — were
split between backings: commondao's kinds registry and proposal storage
already used `p/nt/bptree/v0`, but its address sets rode
`p/moul/addrset` (AVL-backed) and the bylaws store used `p/nt/avl/v0`.

## Decisions

- **New `gno.land/p/nt/addrset/v0`**: same API as `p/moul/addrset`
  (`Set` + `ReadonlySet`, zero-value usable — the B+ tree's zero value
  is an empty fanout-32 tree) backed by `p/nt/bptree/v0`. The moul
  package is untouched (shared library; other consumers keep their
  behavior). The `Tree()` escape hatch is deliberately omitted: no
  branch caller used it, and not exporting the backing store keeps the
  readonly view actually read-only.
- **commondao `/p/` and the realm switch to `p/nt/addrset/v0`** for
  council, electorate snapshots and the tally/readonly/render surfaces;
  every commondao container is now B+-tree-backed.
- **`p/nt/bylaws/v0` swaps `avl.Tree` → `bptree.BPTree`** for its
  document store. Iteration semantics are identical (sorted keys,
  half-open `[start, end)`, empty bound = unbounded), so the prefix
  "folder" scans and all goldens are unchanged.

## Alternatives considered

- Change `p/moul/addrset` in place — rejected: it is a shared moul
  library with its own consumers and an exported `Tree() avl.ITree`
  contract that a B+ tree cannot honor.
- Leave AVL and wait for object-persistence overhead work — the
  measured floor (~1 KB persisted per 40-byte key even on bptree, a
  ~25× object-model overhead) dwarfs the backing choice and is worth
  chain-level attention, but the 2× win is free today and the API swap
  is mechanical.

## Consequences

- Same sorted iteration order everywhere → no golden or behavior
  changes; all suites and the six commondao txtars pass unmodified.
- Roughly half the storage bytes and ~2.7× less insert gas for every
  council/electorate/vote/document entry going forward.
- `ReadonlySet` is now a realm-boundary-safe view with no backing-store
  escape hatch.
