# Fix GC allocation/recount mismatch

## Context

When Gno code triggers infinite recursion (e.g., `func f() int { return f() }`),
the VM panics with `"should not happen, allocation limit exceeded while gc."`
instead of the expected `"allocation limit exceeded"`.

The allocator tracks memory via `alloc.bytes`. When the limit is reached, GC runs:
it resets `alloc.bytes` to 0 and walks all live objects, recounting their sizes via
`GetShallowSize()`. If GC recounts more bytes than were originally tracked, it
concludes the limit was exceeded during GC itself -- a path marked as "should not happen".

The root cause is in `PrepareNewValues` (`nodes.go`): when new package-level
declarations (functions, variables) are added after initial package creation,
their block items are appended to the package block's `Values` slice via
`block.Values = append(block.Values, nvs...)` **without** calling
`AllocateBlockItems`. This means:

- **During allocation**: `alloc.bytes` does not account for the new block items
- **During GC recount**: `Block.GetShallowSize()` uses `len(b.Values)` which
  includes the appended items

The mismatch is small (e.g., 80 bytes for 2 functions), but in infinite recursion
the allocator fills to near-max with block allocations, and the untracked bytes
tip the GC recount past `maxBytes`.

## Decision

Add `alloc.AllocateBlockItems(int64(len(nvs)))` in `PrepareNewValues` before
appending to `block.Values`. This ensures the allocator tracks the same bytes
that GC will recount.

## Alternatives considered

1. **Fix GC to not stop early when recount exceeds maxBytes** -- this masks the
   mismatch rather than fixing it. GC would always succeed, but the allocator
   state would be inconsistent.

2. **Track allocations in preprocessing** (adding `AllocateFunc` calls in
   `preprocess.go` for FuncValues created during `tryPredefine`) -- this was
   explored but turned out to be unnecessary. The FuncValues are properly
   allocated via `fv.Copy(alloc)` in `PrepareNewValues`. The actual mismatch
   was only the block items.

3. **Change the panic message** from "should not happen" to "allocation limit
   exceeded" -- this was considered as a minimal fix but doesn't address the
   underlying inconsistency.

## Consequences

- Infinite recursion now correctly panics with `"allocation limit exceeded"`
- GC recount is consistent with allocator tracking for package block items
- Gas values for some tests change slightly because `AllocateBlockItems` charges
  gas during package setup
- The "should not happen" panic path remains as a safety net for any future
  mismatches -- it should now truly never trigger
