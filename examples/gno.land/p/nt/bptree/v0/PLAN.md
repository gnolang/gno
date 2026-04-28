# Mutable B+ Tree for Gno

## Goal

A mutable (in-place) B+ tree with the same API as `gno.land/p/nt/avl/v0`.
No merkle hashing, no versioning, no persistence — just a simple, efficient
ordered map with configurable fanout.

## API

```go
// Constructors
func NewBPTreeN(fanout int) *BPTree   // arbitrary fanout (minimum 4)
func NewBPTree32() *BPTree            // convenience: fanout 32

// ITree interface (same as avl.ITree)
type ITree interface {
    Size() int
    Has(key string) bool
    Get(key string) (value any, exists bool)
    GetByIndex(index int) (key string, value any)
    Iterate(start, end string, cb IterCbFn) bool
    ReverseIterate(start, end string, cb IterCbFn) bool
    IterateByOffset(offset int, count int, cb IterCbFn) bool
    ReverseIterateByOffset(offset int, count int, cb IterCbFn) bool
    Set(key string, value any) (updated bool)
    Remove(key string) (value any, removed bool)
}

type IterCbFn func(key string, value any) bool
```

`var _ ITree = (*BPTree)(nil)` enforces the interface at compile time.

## Semantics (matching avl exactly)

- `Set`: insert or update. Returns true if key already existed.
- `Remove`: delete key. Returns (old value, true) if found.
- `GetByIndex`: 0-based index into sorted keys. Panics on invalid index.
- `Iterate(start, end, cb)`: ascending, start inclusive, end exclusive.
  Empty string means no bound.
- `ReverseIterate(start, end, cb)`: descending, start inclusive, end inclusive.
  Empty string means no bound.
- `IterateByOffset(offset, count, cb)`: ascending from the offset-th leaf entry
  (0-indexed from the smallest key). Returns false if offset >= size or count <= 0.
- `ReverseIterateByOffset(offset, count, cb)`: descending, where offset is
  0-indexed from the largest key. offset=0 starts at the largest key,
  offset=1 skips the largest and starts at the second-largest, etc.
  Equivalent to: visit entries in descending order, skip `offset`, take `count`.
- All iteration callbacks return true to stop early.

## Node types

### leafNode

```go
type leafNode struct {
    keys   []string  // sorted, len <= fanout
    values []*any    // parallel to keys
}
```

Leaf nodes store all key-value data. No sibling pointers — iteration
uses stack-based traversal to avoid ref-count >= 2 (see
[Ref-Count Safety](#ref-count-safety)).

Capacity: up to `fanout` entries. Splits when full after insert.
Minimum occupancy enforced during deletion: `fanout/2` (except the root leaf).
Note: the 90/10 split optimization intentionally creates a right leaf with
only 2 entries, which may be below `fanout/2` for large fanouts. This is
standard B+ tree practice — if a subsequent Remove causes underflow, the
normal rebalance logic (redistribute or merge) handles it.

### innerNode

```go
type innerNode struct {
    keys     []string // separator keys, len = len(children)-1
    children []node   // child pointers, len <= fanout
    sizes    []int    // sizes[i] = total leaf count in children[i] subtree
}
```

`keys[i]` = minimum key of `children[i+1]` (standard B+ tree convention).
An inner node with k separator keys has k+1 children.

Capacity: up to `fanout` children (= `fanout-1` keys). Splits when full.
Minimum occupancy: `fanout/2` children (except root, which may have as few as 2).

### node interface

```go
type node interface {
    isLeaf() bool
    nodeSize() int    // total leaf entries in subtree
    minKey() string   // leftmost key in subtree
}
```

`nodeSize()`: leafNode returns `len(keys)` (O(1)), innerNode sums `sizes[]` (O(fanout)).

Used internally so tree methods can handle both node types polymorphically.
Type assertions to `*leafNode` or `*innerNode` are used when accessing
type-specific fields (children, sizes).

## BPTree struct

```go
type BPTree struct {
    root   node
    size   int       // total number of key-value pairs
    fanout int       // max children per inner node / max entries per leaf
}
```

No `first`/`last` pointers — they would create ref-count >= 2 on leaves.
Full-range iteration descends from the root (O(height) to find the first
or last leaf, amortized O(1) per entry thereafter).

## Key algorithms

### Search (Get, Has)

Descend from root. At each inner node, binary search `keys` to find the
child index. At the leaf, binary search `keys` for exact match.

### Insert (Set)

1. Descend root-to-leaf, recording the path (stack of `(innerNode, childIdx)` pairs).
2. Binary search in the leaf for the key.
3. If found: update value in place, return `updated=true`. No structural change.
4. If not found: insert key-value at the sorted position.
   - Increment `size` on the tree and all ancestor `sizes[childIdx]` entries.
   - If the leaf now has `fanout+1` entries, **split**:
     - **90/10 split** (append pattern): if the new key was inserted at
       position `fanout` in the overflowed leaf (i.e., it is greater than
       all pre-existing keys), split asymmetrically: left gets `fanout-1` entries, right
       gets 2 entries (the last existing key + the new key). This keeps left
       leaves ~97% full for sequential inserts.
     - **50/50 split** (random pattern): otherwise, left gets `(fanout+1)/2`
       (floor), right gets the rest.
     - Create new right leaf node.
     - Promote `right.keys[0]` as separator to the parent.
     - Update parent's `sizes` for both children.
     - If the parent now has `fanout+1` children, split the parent recursively.
   - If the root splits, create a new inner root with 2 children.

### Remove

1. Descend root-to-leaf, recording the path.
2. Binary search in the leaf for the key.
3. If not found: return `(nil, false)`.
4. If found: remove entry, decrement `size` and ancestor `sizes`.
   - If the leaf is the root and now empty: set root to nil.
   - If the leaf is the root and non-empty: done (root has no minimum).
   - Otherwise, if leaf has fewer than `fanout/2` entries, **rebalance**:
     - Try to **redistribute** from left sibling (if it has more than `fanout/2`).
     - Try to **redistribute** from right sibling (if it has more than `fanout/2`).
     - Otherwise **merge** with a sibling:
       - Concatenate into the left node, remove the right child and its
         separator from parent.
       - Update parent's `sizes` and `size`.
       - If parent is the root and drops to 1 child, replace root with that child.
       - If parent is not the root and has fewer than `fanout/2` children,
         rebalance recursively (the root is exempt — it may have as few as 2 children).
   - If the minimum key was removed (pos == 0), update ancestor separator keys
     before rebalancing. Rebalance operations also fix separators as needed.

### Redistribute detail

**Redistribute from left sibling to deficient child (both leaves):**
1. Move left sibling's last key-value to the front of the deficient child.
2. Update parent `keys[childIdx-1]` = deficient child's new `keys[0]` (its min key changed).
3. Adjust parent `sizes`: decrement left's size, increment child's size.

**Redistribute from right sibling to deficient child (both leaves):**
1. Move right sibling's first key-value to the end of the deficient child.
2. Update parent `keys[childIdx]` = right sibling's new `keys[0]` (its min key changed).
3. Adjust parent `sizes`: decrement right's size, increment child's size.

**Redistribute from left sibling to deficient child (both inner nodes):**
1. Pull down parent's separator `keys[childIdx-1]` — **prepend** it to deficient child's keys
   (insert at position 0, since the moved child goes to the front).
2. Move left sibling's last child to the **front** of deficient child's children.
   Move left sibling's last `sizes` entry to the front of deficient's sizes too.
3. Push up left sibling's last key to replace parent's `keys[childIdx-1]`.
4. Remove left sibling's last key, last child, and last size entry.
5. Update parent's `sizes[childIdx-1]` and `sizes[childIdx]` to reflect
   the new child sizes. (Parent's total size is unchanged since we just
   moved entries between siblings.)

**Redistribute from right sibling to deficient child (both inner nodes):**
1. Pull down parent's separator `keys[childIdx]` — **append** it to deficient child's keys
   (insert at the end, since the moved child goes to the end).
2. Move right sibling's first child to the **end** of deficient child's children.
   Move right sibling's first `sizes` entry to the end of deficient's sizes too.
3. Push up right sibling's first key to replace parent's `keys[childIdx]`.
4. Remove right sibling's first key, first child, and first size entry.
5. Update parent's `sizes[childIdx]` and `sizes[childIdx+1]` to reflect
   the new child sizes. (Parent's total size is unchanged.)

**Merge two inner nodes:**
1. Pull down the parent's separator between them into the left node's keys.
2. Append all of right node's keys, children, and sizes to left node.
3. Remove right child and its separator from parent.
4. Update parent's `sizes` for the merged left child.

### GetByIndex

Use `sizes[]` at each inner node to find which child contains the i-th
leaf entry, then descend. At the leaf, index directly into `keys[i]`/`values[i]`.
Panics if index is out of range (matching avl behavior).

### Stack-based iteration

All iteration uses a stack of `(innerNode, childIndex)` pairs representing
the path from root to the current leaf. When a leaf is exhausted, pop the
stack, advance (or retreat) the child index, and descend to the next leaf.
Amortized O(1) per entry — each node is pushed/popped at most once across
the full traversal.

The stack is a local slice built during iteration and discarded after —
it creates no persistent references to nodes and does not affect ref-counts.

### Iterate / ReverseIterate (key-range)

**Ascending (Iterate):**
1. Descend from root using separator keys to find the leaf containing `start`
   (or descend to the leftmost leaf if start is ""), recording the path.
2. Within the leaf, find the first key >= start.
3. Visit entries from that position forward.
4. When the leaf is exhausted, advance to the next leaf via the stack:
   pop the stack, increment childIdx. If childIdx is now past the last
   child, pop again (repeat until a valid childIdx is found or the stack
   is empty — if empty, iteration is done). Then descend to the leftmost
   leaf of that child, pushing each inner node onto the stack.
5. Stop when key >= end (if end != "") or tree is exhausted.
6. Call `cb(key, value)` for each entry; stop if cb returns true.

**Descending (ReverseIterate):**
1. Descend from root to find the leaf containing `end` (or the rightmost
   leaf if end is ""), recording the path.
2. Within the leaf, find the last key <= end.
3. Visit entries from that position backward.
4. When the leaf is exhausted going backward, retreat to the previous leaf
   via the stack: pop the stack, decrement childIdx. If childIdx < 0, pop
   again (repeat until a valid childIdx is found or the stack is empty —
   if empty, iteration is done). Then descend to the rightmost leaf of
   that child, pushing each inner node onto the stack.
5. Stop when key < start (if start != "") or tree is exhausted.
6. Call `cb(key, value)` for each entry; stop if cb returns true.

### IterateByOffset / ReverseIterateByOffset

**Ascending:**
1. Use `sizes[]` to descend to the leaf containing the offset-th entry,
   recording the path. Maintain a running offset counter: at each inner node,
   subtract `sizes[i]` for each skipped child. When `offset < sizes[i]`,
   descend into that child. Upon reaching a leaf, the remaining offset is
   the position within the leaf.
2. Visit entries from that position forward, advancing through leaves via
   the stack (same as Iterate), counting up to `count`.

**Descending:**
The descending view is: entries in reverse sorted order, 0-indexed from
the largest key. offset=0 is the largest, offset=1 is the second-largest, etc.

1. Compute the ascending index of the starting entry:
   `ascIdx = size - 1 - offset` (the entry at position `offset` in descending order).
2. Use `sizes[]` to descend to the leaf containing entry `ascIdx`,
   recording the path.
3. Visit entries from that position backward, retreating through leaves via
   the stack (same as ReverseIterate), counting up to `count`.
4. If `ascIdx < 0` or `offset >= size` or `count <= 0`, return false.

## Split details

### Leaf split

Given a leaf with `fanout+1` entries (one over capacity):

**Detection:** after inserting the new key, if it ended up at position
`fanout` in the overflowed `fanout+1`-entry leaf, it is an append-pattern
insert (the new key is greater than all pre-existing keys).

**90/10 split (append pattern):**
- `mid = fanout - 1`
- Left leaf keeps entries `[0, fanout-1)` = `fanout-1` entries.
- New right leaf gets entries `[fanout-1, fanout+1)` = 2 entries.
- Left has `fanout-1` entries (~97% full), right has 2.
- For large fanouts, the right leaf may be below the `fanout/2` deletion
  threshold. This is intentional — the 90/10 split prioritizes fill factor
  for append-heavy workloads. If a subsequent Remove causes the right leaf
  to underflow, the standard rebalance logic handles it.

**50/50 split (random pattern):**
- `mid = (fanout + 1) / 2`
- Left leaf keeps entries `[0, mid)`.
- New right leaf gets entries `[mid, fanout+1)`.

In both cases:
- Separator promoted to parent = `right.keys[0]`.
- No linked list updates needed (no sibling pointers).

### Inner split

Given an inner node with `fanout+1` children:
- `mid = (fanout + 1) / 2`
- Left keeps children `[0, mid)` with keys `[0, mid-1)` and sizes `[0, mid)`.
- Right gets children `[mid, fanout+1)` with keys `[mid, fanout)` and sizes `[mid, fanout+1)`.
- The separator at `keys[mid-1]` is **promoted** to the parent (not kept in either child).
- Parent's sizes entry for the original child is replaced by the sum of left's sizes,
  and a new entry is inserted for the right child with the sum of right's sizes.

## Minimum fanout

Fanout must be >= 4. With fanout 3, a leaf splits into (2, 2) and
the minimum occupancy is 1, which makes merge logic degenerate. Fanout 4
gives minimum occupancy 2 and clean split/merge behavior.

`NewBPTreeN` panics if fanout < 4.

## File structure

```
examples/gno.land/p/nt/bptree/v0/
  gnomod.toml
  doc.gno          — package doc
  node.gno         — leafNode, innerNode, node interface, binary search
  tree.gno         — BPTree struct, constructors, ITree methods,
                     insert/split, remove/merge/redistribute, iteration
  tree_test.gno    — comprehensive tests (mirroring avl tests + B+ tree specifics)
```

Two source files (`node.gno` + `tree.gno`) plus one test file. The node
types and tree logic are tightly coupled, so fewer files is better than
spreading thin.

## Ref-count safety

In Gno's persistence model, objects with ref-count >= 2 "escape" — they are
persisted separately in an iavl tree rather than inlined in their parent's
serialized form. Once escaped, they are forever escaped. This is expensive
and should be avoided.

**Design constraint: every node must have exactly one persistent reference.**

This means:
- **No sibling pointers** on leaf nodes (a leaf would be referenced by both
  its parent and its neighbor → ref-count >= 2).
- **No `first`/`last` pointers** on BPTree (a leaf would be referenced by
  both BPTree and its parent inner node → ref-count >= 2).
- **No shared subtrees** (each child is owned by exactly one parent).

The tree structure is a pure tree (not a graph) — every node has exactly
one parent reference. The `BPTree.root` is the sole reference to the root
node. Each `innerNode.children[i]` is the sole reference to child `i`.

Iteration uses an ephemeral stack (local slice) that is built and discarded
within a single method call. It does not create any persistent references.

## Edge cases (must match avl behavior exactly)

### Values and keys
- `nil` is a valid value. `Set("foo", nil)` stores it; `Get("foo")` returns
  `(nil, true)`. `Remove("foo")` returns `(nil, true)`.
- `""` is a valid key. Stored, retrieved, removed like any other key.
- `Get` on missing key returns `(nil, false)`.
- `Remove` on missing key returns `(nil, false)`.
- `Set` same key twice replaces value, returns `updated=true`.

### Zero-value and structural
- `var t BPTree` must work — zero-value tree is usable without a constructor.
  All methods work on it immediately. This means `root == nil` must be handled
  gracefully everywhere, and the default `fanout` (0) must be promoted to 32
  on first use. `Set` promotes on first call: `if t.fanout == 0 { t.fanout = 32 }`.
  All other methods that read `t.fanout` guard with `if t.root == nil` first, so
  fanout is always initialized before it is read. Once set, fanout never resets
  (even if tree becomes empty again).
- Remove last key → tree returns to empty state (root = nil).
- Insert after removing everything works normally.
- `Size()` on empty tree returns 0.
- `GetByIndex` on empty tree panics. Negative index or index >= size also panics.
- Single entry: root is a leafNode with 1 entry.
- Root is a leaf: no inner nodes until first split.
- Root inner node collapses: after merge leaves root with 1 child,
  replace root with that child.

### Separator key maintenance
- When the minimum key of a subtree changes (deletion of leftmost key,
  or redistribution), the parent's separator key must be updated.
  After modifying a child, check if `keys[childIdx-1]` still equals
  `children[childIdx].minKey()`.

### Iteration — key range
- All iteration on an empty tree returns `false` without calling cb.
- `Iterate("", "", cb)` visits ALL entries ascending (canonical pattern).
- `ReverseIterate("", "", cb)` visits ALL entries descending.
- `Iterate("a", "a", cb)` → empty. [a,a) = nothing (start inclusive, end exclusive).
- `ReverseIterate("a", "a", cb)` → visits "a". [a,a] = one entry (both inclusive).
- `Iterate("z", "a", cb)` → empty (no validation, logic just excludes everything).
- `ReverseIterate("z", "a", cb)` → empty (bounds don't swap).
- `Iterate("", "a", cb)` → visits all keys < "a".
- `ReverseIterate("a", "", cb)` → visits all keys >= "a", in descending order.
- Return value = true if callback stopped iteration early, false otherwise.

### Iteration — offset
- `IterateByOffset(0, 0, cb)` → nothing (count <= 0).
- `IterateByOffset(size, 1, cb)` → nothing (offset >= size).
- `ReverseIterateByOffset(0, N, cb)` → starts at largest key, takes N descending.
- `ReverseIterateByOffset(1, 2, cb)` on [a,b,c,d,e] → [d, c].
- Negative count → treated as count <= 0 (no iteration).
- Negative offset → clamped to 0 (avl silently treats negative as 0; we do the same explicitly).
