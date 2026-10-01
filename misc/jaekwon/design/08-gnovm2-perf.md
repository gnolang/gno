# GnoVM2 Performance Optimizations

Behavior-preserving performance improvements to the GnoVM interpreter.
These are separate from the spec — the reference Gno1 implementation
stays simple; these are optimizations a fork or post-launch mainline
can adopt.

## Plan

### 1. Pre-allocate all machine stacks (TRIVIAL)

`startingValuesCap = 25` is far too small — 25 TypedValues before
reallocation. Exprs, Stmts, Blocks, Frames stacks start at zero
capacity. Bump Values to 512, and pre-allocate the others in the
pool so Release() can preserve them.

**Files:** `machine.go` (pool, Release, consts)

### 2. Batch gas metering (SMALL)

`incrCPU()` runs on every opcode: nil check, overflow.Mulp(×1),
interface call. Accumulate cycles locally and flush to GasMeter
every N ops or at call/return boundaries. Also eliminate the
`overflow.Mulp(cycles, 1)` since GasFactorCPU == 1.

**Files:** `machine.go` (incrCPU, Run loop, call/return paths)

### 3. Inline NameExpr depth lookup (SMALL)

`doOpEval` NameExpr fast-path still calls `GetPointerTo` which
loops through parent blocks. Unroll depths 1-3 inline since most
variable accesses are shallow.

**Files:** `op_eval.go` (doOpEval NameExpr path)

### 4. Inline baseOf() in binary ops (SMALL)

Every arithmetic op does `switch baseOf(lv.T)` — a function call
plus DeclaredType type assertion. Inline it or cache the base kind.

**Files:** `op_binary.go`, `types.go`

### 5. Remove PkgIDFromPkgPath mutex (TRIVIAL)

Global sync.Mutex on a cache that's only accessed single-threaded.
Use sync.Map or pass through Store context.

**Files:** `realm.go`

### 6. Superinstructions for common op sequences (MEDIUM)

Profile top op pairs/triples; fuse into single opcodes.
Candidates: Eval+Eval+BinaryOp, Eval+Compare+IfCond.

**Files:** `machine.go`, `op_eval.go`, `op_binary.go`, `preprocess.go`

### 7. Pool big.Int temporaries (SMALL)

`big.NewInt(0)` allocated on every int literal eval. Keep a small
sync.Pool of *big.Int and reset+reuse.

**Files:** `op_eval.go`, `op_binary.go`

### 8. Generational GC (LARGE)

Current GC is stop-the-world full-graph walk. Track generations
via lastGCCycle; skip old-gen objects unless dirtied.

**Files:** `garbage_collector.go`, `alloc.go`, `ownership.go`

## Status

| # | Item | Status |
|---|------|--------|
| 1 | Pre-allocate stacks | DONE |
| 2 | Batch gas metering | SKIP (done elsewhere) |
| 3 | Inline NameExpr depth | DONE |
| 4 | Inline baseOf() | SKIP (already inlined by compiler) |
| 5 | Remove PkgID mutex | DONE |
| 6 | Superinstructions | TODO |
| 7 | Pool big.Int | TODO |
| 8 | Generational GC | TODO |
