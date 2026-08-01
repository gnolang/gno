# ADR: commondao/bylaws persisted sets move from AVL to B+ tree

## Context

Persisted-storage measurements (filetests storing N distinct
zero-padded 40-char addresses; storage deltas from the test runner)
showed the AVL backing costs ~2.2× the bytes of a fanout-32 B+ tree
asymptotically (1.6× at 10 entries), because AVL persists one node
object per entry while the B+ tree packs ~32 entries per node (fewer
object headers and refs):

| items | addrset (avl) | bptree32 | avl B/item | bptree B/item | ratio |
|------:|--------------:|---------:|-----------:|--------------:|------:|
| 10 | 19,031 B | 11,867 B | 1,903 | 1,187 | 1.6× |
| 100 | 200,280 B | 97,351 B | 2,003 | 974 | 2.1× |
| 1,000 | 2,023,691 B | 917,606 B | 2,024 | 918 | 2.2× |
| 10,000 | 20,365,699 B | 9,180,737 B | 2,037 | 918 | 2.2× |

Insert gas follows the same shape (measured once at 10k adds: 8.8B vs
4.2B, 2.1×; reads unmeasured). The branch's growth-prone sets —
councils, per-proposal electorate snapshots (one per proposal), voting
records, bylaws documents — were split between backings: commondao's
kinds registry and proposal storage already used `p/nt/bptree/v0`, but
its address sets rode `p/moul/addrset` (AVL-backed) and the bylaws
store used `p/nt/avl/v0`.

Reproducing the table: for each N and backing, run a throwaway filetest
realm that persists the container as a package var and inserts N keys
in main; the test runner's PASS line reports the realm's storage delta.
Shape (swap `addrset.Set` for `bptree.NewBPTree32()` + `Set(key,
struct{}{})` for the other column):

```
// PKGPATH: gno.land/r/bench/addrset_N
var s = addrset.Set{}
func main(cur realm) {
    for i := 0; i < N; i++ {
        d := strconv.Itoa(i)
        s.Add(address("g1" + strings.Repeat("0", 38-len(d)) + d))
    }
    println("n:", s.Size())
}
```

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
  measured floor (~0.9 KB persisted per 40-byte key even on bptree, a
  ~23× object-model overhead) dwarfs the backing choice and is worth
  chain-level attention, but the 2× win is free today and the API swap
  is mechanical.

## Consequences

- Same sorted iteration order everywhere → no golden or behavior
  changes; all suites and the six commondao txtars pass unmodified.
- ~2.2× fewer storage bytes and ~2.1× less insert gas for every
  council/electorate/vote/document entry going forward.
- `ReadonlySet` is now a realm-boundary-safe view with no backing-store
  escape hatch.
- Two contract narrowings vs the AVL backing, documented on the new
  package: no Add/Remove from an iteration callback (AVL's
  copy-on-write tolerated it; the B+ tree mutates in place), and no
  by-value copies of a non-zero `Set` (AVL copies were independent
  snapshots). Every existing call site was audited against both.
- Follow-up: `p/nt/groups/v0` (a master-owned package) still rides the
  AVL-backed `p/moul/addrset` for its membership sets; migrate it the
  same way if its sets become part of hot persisted state.
