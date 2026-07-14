# ADR: Devirtualize DidUpdate and word-compare PkgID on per-write hot paths

## Context

`Realm.DidUpdate` is the ownership hook called after every mutation of
realm-owned state (all `op_assign` paths, inc/dec, map/slice/append
writes, pointer writes). CPU profiling of a `DidUpdate` microbenchmark
showed two dominant costs:

1. **`runtime.memequal` (~18% of samples)** — `PkgID` holds a
   `Hashlet [20]byte`; the `poPkgID != rlm.ID` comparison lowers to a
   `runtime.memequal` call (Go only unrolls small array compares), paid
   on every real-object write and again for `co`/`xo` handling.
2. **Interface dispatch** — `DidUpdate` and the `Mark*` helpers called
   `po.GetIsReal()`, `po.GetObjectID()` (28-byte struct copy per call),
   `co.IncRefCount()`, `oo.GetIsDirty()`, etc. through the `Object`
   interface: each is a dynamic call the compiler cannot inline.
   `MarkDirty` alone made 3–4 interface calls per invocation.

The same 20-byte comparisons sit in the pre-write guards that run at
every mutation site: `TypedValue.IsReadonlyBy` (cross-realm write
authority), `Machine.isExternalRealm` (NameExpr write gate), the borrow
rule receiver check in `PushFrameCall`, and `Hashlet.IsZero`
(`ObjectID.IsZero` in the same guards).

## Decision

1. **`PkgID.eq`** (realm.go): hand-unrolled equality — two
   `binary.NativeEndian.Uint64` loads plus one `Uint32` load, compared
   word-wise. Inlines to three MOV/CMP pairs; endianness is irrelevant
   for equality. Used in `DidUpdate`, `IsReadonlyBy`,
   `isExternalRealm`, and the `PushFrameCall` borrow-rule check.
2. **`Hashlet.IsZero`** (hash_image.go): same treatment — OR the three
   words, compare against zero.
3. **Devirtualize `DidUpdate`** (realm.go): fetch `GetObjectInfo()`
   once per object (`po`, `co`, `xo`); all subsequent flag/refcount
   accesses are concrete `*ObjectInfo` method calls, which inline to
   direct field accesses. No `Object` implementation overrides these
   methods (all inherit the embedded `*ObjectInfo`), so behavior is
   identical.
4. **Split `Mark*` helpers** (realm.go): each exported `MarkX(oo)` is
   now a thin wrapper over an unexported `markX(oo, oi)` body.
   `DidUpdate` calls the `markX` forms with the `*ObjectInfo` it
   already holds; external callers (machine.go) keep the old API.

Microbenchmarks (`realm_didupdate_bench_test.go`, Apple M1 Pro,
steady-state paths):

| scenario                       | before  | after   |
|--------------------------------|---------|---------|
| po unreal (early return)       | 4.1 ns  | 3.5 ns  |
| po real, primitive field write | 10.5 ns | 6.0 ns  |
| po real, attach real co        | 22.3 ns | 9.0 ns  |
| po real, swap xo→co            | 27.1 ns | 11.8 ns |

After the change the profile shows no `runtime.memequal` and no
non-inlined callees in `DidUpdate`.

## Alternatives considered

- **Change `Hashlet` to a struct of two uint64 + uint32**: would make
  `==` cheap everywhere, but `Hashlet` is serialized (Amino) and
  sliced as bytes throughout store/hashing code; too invasive.
- **`unsafe`-based comparison**: same codegen as `encoding/binary`
  loads (which are compiler-intrinsified) without the portability
  guarantees; rejected.
- **Caching the "is own realm" verdict on ObjectInfo**: stateful,
  invalidation-prone; the word compare is cheap enough.
- **Skipping the `poPkgID` check when `m.Stage != StageRun`**: changes
  invariant coverage, rejected — the check is the security net for
  missing readonly pre-checks.
- **`co == xo` fast-path** (the dormant `XXX if co == xo` at the top of
  `DidUpdate`): skip the ref-count/dirty bookkeeping when a slot is
  assigned the exact object it already holds (e.g. AVL rebalance with
  no rotation). **Rejected — this is a consensus change, not a perf
  optimization.** `MarkDirty(co)` stamps `co.ModTime = rlm.Time`, and
  `ModTime` is amino-serialized into the object's stored image
  (`pb3_gen.go`) which is hashed into the realm merkle root. Skipping
  the mark leaves `ModTime` un-advanced, so an optimized node and a
  stock node would persist different bytes for the same object and
  diverge. The refcount `+1/-1` ops do cancel and any spurious escape
  mark is demoted in `processNewEscapedMarks`, but the `ModTime`
  divergence is dispositive.
- **Skip re-saving unchanged objects on a reference shuffle**
  (`s[i],s[j] = s[j],s[i]` nets zero refcount change): same blocker —
  `ModTime` advances on the dirty mark and is committed state. Both
  optimizations would require a coordinated protocol change redefining
  when `ModTime` advances; out of scope here. Filed for discussion
  rather than implemented.

## Consequences

- Gas is unaffected: no gas-metered constants or alloc sizes change;
  the three verification suites (`-run Gas`, `-run TestTestdata`,
  `-run Files -test.short`) pass unchanged, as do the package tests
  under `-tags debugAssert`.
- `PkgID.eq`/`IsZero` must stay in sync with `HashSize` (20); both are
  covered by segment-by-segment unit tests
  (`realm_pkgid_test.go`) that cross-check against the generic `==`.
- The `markX(oo, oi)` forms require `oi == oo.GetObjectInfo()`; they
  are unexported and only called from `DidUpdate`/wrappers.
