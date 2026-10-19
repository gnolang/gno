# Bound pebble and goleveldb iterators to their range

## Status

Implemented.

## Context

Short iterations on pebble and goleveldb cost time in proportion to the
deleted keys after their range, not the keys they return.

Both stores are LSM trees: a delete writes a tombstone that stays until a
compaction drops it. An iterator opened without bounds steps over every
tombstone past its range until it reaches a live key, and a reverse iterator
walks them a second time on its way back. Pruning leaves such runs after live
ranges. Take 3 live keys followed by 100,000 deleted ones: a pebble reverse
iteration takes 4.4 ms, a goleveldb forward one 4.9 ms and a goleveldb reverse
one 11.4 ms.

Only pebble's forward iterators passed the range to the store, as `LowerBound`
and `UpperBound`. pebble's `ReverseIterator`, on the DB and on snapshots, sought
to the end of an unbounded range. goleveldb's `Iterator` and `ReverseIterator`
were unbounded in both directions.

Reverse iterators sit on the store's read paths, such as
`ReversePrefixIterator` and IAVL's legacy latest-version lookup, and every
goleveldb read path pays in both directions.

## Decision

Every pebble and goleveldb iterator passes `[start, end)` to the store as
bounds. pebble's reverse iterators, on the DB and on snapshots, set
`LowerBound` and `UpperBound` as its forward ones do. goleveldb's iterators
open with a `util.Range` of `[start, end)`.

Each iterator starts at `First()` or `Last()` inside its bounds, and the store
stops at them, so `Valid()` compares no key with the range. Two store
behaviours shape this:

- goleveldb can panic on an inverted `util.Range` once any level past 0 holds a
  table. It selects the level's tables for the range by slicing its file list
  from the start index to the end index, and an inverted range can put the
  first past the second. The `db.Iterator` contract makes an inverted range an
  invalid iterator, so `iteratorRange` clamps it to the empty `[start, start)`.
- pebble copies its bounds into a buffer of the iterator's pooled allocation.
  An empty `UpperBound` copies to nil while that buffer is still nil, and
  pebble reads nil as no upper bound. No key sorts before an empty end, so a
  pebble iterator given one starts invalid.

Each of the four pebble iterators calls `NewIter` on the concrete `*pebble.DB`
or `*pebble.Snapshot`, so its `IterOptions` stay on the stack. Through the
`pebble.Reader` interface the compiler cannot prove that, and each iterator
would heap-allocate its 192 bytes. `iteratorRange` returns its `util.Range` by
value for the same reason: goleveldb copies the range.

## Alternatives considered

- One constructor taking a `pebble.Reader` for the four pebble iterators: a
  192-byte heap allocation on every pebble iterator, forward ones included.
- Leaving an inverted range to goleveldb: a panic where the contract asks for an
  invalid iterator.
- Checking the end bound in pebble's `Valid()`: a key comparison on every step
  for a case known when the iterator is built.

## Consequences

Iteration returns the same keys on every backend; only the work behind it
changes. Gas is charged per iterator and per returned item at the store layer
(`WillIterator`, `WillIterNext`), so gas and consensus are unchanged.

A pebble reverse iteration and a goleveldb iteration in either direction cost
what a pebble forward one does, whatever follows their range. boltdb, lmdb and
mdbx are B+trees and memdb holds only live keys, so an unbounded seek on them
steps over no deleted keys; they are unchanged.

Bounds cost goleveldb a fixed amount per iterator: it copies the range into
internal keys, a net two allocations. With no deleted keys after the range, a
goleveldb iteration takes about 90 ns more forward and 20 ns more reverse.
pebble's time and allocations per iterator are unchanged when no deleted keys
follow the range: 72 B in 2 allocations forward and 80 B in 3 reverse, before
and after.

## Validation

- `TestPebbleIteratorsStayInRange`: over a range followed by 10,000 deleted
  keys, the four pebble iterators yield the in-range keys and read at most 100
  internal points. Unbounded, the reverse ones read 20,004; bounded, each reads 3.
- `TestGoLevelDBIteratorsStayInRange`: forward and reverse iterators over the
  same layout make at most 1,000 key comparisons. Unbounded they make about
  30,000 and 60,000; bounded, 35 each.
- `TestGoLevelDBInvertedRangeIsEmpty`, `TestPebbleInvertedRangeIsEmpty`: with
  keys spread over several tables of one level, an inverted range yields
  nothing, on the DB and on a snapshot for pebble. Without the clamp goleveldb
  panics with `slice bounds out of range [3:1]`.
- `TestPebbleEmptyEndYieldsNothing`: an iterator with an empty end starts
  invalid.
- `TestDBIteratorEmptyBounds`: empty start and end bounds on every backend
  except lmdbdb, which seeks an empty bound and gets `MDB_BAD_VALSIZE` from
  LMDB, with or without this change.
- `TestPebbleIteratorAllocatesOnlyItsWrapper`: each pebble iterator allocates
  at most one object more than pebble's own iterator over the same range. The
  test is excluded under `-race`, which drops `sync.Pool` items at random.

`BenchmarkPebbleDBPrunedRange` and `BenchmarkGoLevelDBPrunedRange` drain an
iterator over 3 live keys followed by a run of deleted keys, held as
tombstones in a table. Median of 5 runs on darwin/arm64, before and after:

| Iterator          | 0 deleted       | 1,000 deleted    | 10,000 deleted   | 100,000 deleted  |
|-------------------|-----------------|------------------|------------------|------------------|
| pebble forward    | 822 ns → 832 ns | 849 ns → 838 ns  | 847 ns → 859 ns  | 950 ns → 956 ns  |
| pebble reverse    | 874 ns → 863 ns | 44.4 µs → 879 ns | 425 µs → 898 ns  | 4.36 ms → 997 ns |
| goleveldb forward | 637 ns → 727 ns | 49.3 µs → 831 ns | 488 µs → 902 ns  | 4.93 ms → 940 ns |
| goleveldb reverse | 784 ns → 806 ns | 115 µs → 923 ns  | 1.13 ms → 965 ns | 11.4 ms → 1.0 µs |
