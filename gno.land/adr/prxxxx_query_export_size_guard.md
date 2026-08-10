# ADR: Size guard on the value-returning query endpoints (export-walk DoS)

## Status

Proposed (PR pending). Fixes the memory-exhaustion DoS reported as dora
finding `97180e82-ab56-4b69-a2ca-a249cfbfce8d`.

## Context

Five query endpoints return a view of GnoVM values built outside the eval
meters. Four are Amino-JSON — `vm/qeval_json`, `vm/qobject_json`,
`vm/qobject_binary`, `vm/qpkg_json` — and run the same two-step tail:

1. `gno.ExportValues` / `gno.ExportObject` (`gnovm/pkg/gnolang/values_export.go`)
   walk the value tree and build a defensive copy (persisted objects become
   `RefValue`, ephemeral cycles are broken with `ExportRefValue{":N"}`).
2. `amino.MarshalJSON` serializes that copy; the qeval path then wraps it with
   a second `json.Marshal`.

The fifth, `vm/qeval`, skips amino and renders with `TypedValue.String()`
instead — a different serializer with the same problem: no meter, no cap.
Measured, `String()` on a 64 MB string result peaks at +319 MB of heap, ~5× the
rendered size. It is the more widely used endpoint of the two eval forms, so
bounding only the JSON one would have left the reported attack usable at
roughly half its original impact.

The eval-time limits (`maxGasQuery = 3e9`, `maxAllocQuery = 1.5e9`) only meter
VM op execution (`incrCPU`) and store I/O. The export walk and the JSON marshal
run **outside** that envelope — plain Go `make()`/`cp()`/reflection with no
accounting and no output-size cap. There is no response-size limit at the
handler, ABCI, or RPC layer either (the RPC server's `MaxBodyBytes = 5MB` caps
the *request*, not the response).

ABCI queries are unauthenticated and carry no fee. So an attacker deploys a
realm with a function returning a large ephemeral value (a one-time, cheap tx),
then issues `qeval_json` queries for free. Measured on this checkout, a single
query drives real transient heap far past the intended 1.5 GB eval ceiling:

| Query (single, free) | Metered eval | Response | Peak heap |
|---|---|---|---|
| `Big(536870912)` (512 MB string) | ~0.5 GB | 512 MB | **5.67 GB** |
| `Structs(1000000)` (1 M structs) | ~1.2 GB | 592 MB | **4.77 GB** |

The dominant amplifier is `amino.MarshalJSON`, not the export copy: for the
string vector `exportCopyValue` returns the `StringValue` by value and copies
nothing, yet the marshal churns ~740 MB for a 64 MB string. Go's OOM is a fatal
runtime error, not a recoverable panic, so `doRecoverQuery` cannot catch it —
the process dies. (App-level queries are serialized through a single
`queryMtx`, so the effect does not stack N-way across concurrent queries; the
risk is single-query spikes and sustained back-to-back pressure, not
parallelism.)

## Decision

Add a **pre-marshal size guard**: bound the estimated serialized size of the
value tree *during the export walk*, aborting before the copy completes and
before `amino.MarshalJSON` is ever called.

1. **`exportLimiter`** (`values_export.go`) — a size accumulator threaded
   through the whole export walk alongside the existing `seen` map. Every
   visited node charges an estimate via `lim.add(n)`:
   - `exportNodeEst = 32` bytes per value/type node (amino's per-node JSON
     overhead: the `T`/`V`/`N` envelope, braces, commas);
   - **every variable-length string the walk emits, at its own length**: value
     content (`StringValue`), `[]byte`-backed array `Data` at `len × 4/3`
     (base64), and the node-attached names copied out of types — field names,
     struct tags, `PkgPath`s, type IDs, `FuncValue` identifiers, and a
     `BoundMethodValue`'s `Method`/`MethodPkg` selector (emitted per node, and
     for a lazy interface bind — `Func == nil` — not otherwise charged by the
     `FuncValue` identifier branch). `BigintValue`
     is charged `BitLen/3` for its decimal text; `BigdecValue` likewise —
     `(Num().BitLen()+Denom().BitLen())/3` for the rat form's `RatString`, and
     `Prec()/3` for the float form's hex text. Its two components run up to
     `ratOverflowBits` (rat) / `BigdecFloatPrec` (float), so the flat
     `exportNodeEst` under-charges the worst case ~20× (a rat component near
     4096 bits emits ~1.2 KB); the estimate over-charges as elsewhere.

   When the running total exceeds the budget, `add` panics
   `ErrExportSizeExceeded`. Because the charge happens *before* recursing into
   or copying the large payload, the walk aborts early — bounding both the
   export copy and (by never reaching it) the marshal. The array-`List` and
   field-list branches pre-charge their element counts before allocating the
   backing slice, since those lengths are attacker-controllable.

   Charging emitted names is load-bearing, not defensive: a deployed package may
   declare a struct tag or identifier of arbitrary length, and a *non-declared*
   struct type is re-emitted once per element of a slice of that type. An
   earlier revision charged only `exportNodeEst` there, which left the effective
   bound at `nodeCount × maxTagLen` — measured, a 10 KB tag on an anonymous
   struct passed a 10 MB budget and marshaled to 198 MB. Regression test:
   `TestExportValuesLimit_FieldNamesAndTagsCharged`, plus the `TaggedSlice()`
   case in the txtar. The same class applies to a slice of lazy interface-bound
   method values whose selector is a long identifier — charged since
   `TestExportValuesLimit_BoundMethodNameCharged` /
   `TestQueryEvalJSON_BoundMethodNameGuarded`.

   One realm-controlled string reaches a query response *outside* this walk: the
   `@error` field on `vm/qeval_json`, populated from the returned value's own
   `Error()` method (`stringifyJSONResults` → `tryGetError`). Only `jres.Results`
   goes through `ExportValues`; the error text does not, so it is bounded
   separately with `truncate` (the module's `maxBoundedBytes` diagnostic cap,
   used for every panic/error string), keeping a large `Error()` from
   re-introducing the amplification the guard removes. Regression test:
   `TestQueryEvalJSON_AtErrorBounded`.

2. **`ExportValues(tvs, maxBytes)` / `ExportObject(obj, maxBytes)`** — the size
   budget is part of the only entry points, which install the limiter and
   recover the panic into a clean `ErrExportSizeExceeded` via one shared
   `withExportLimit` helper. `maxBytes <= 0` disables the bound, for trusted
   callers and tests. There is deliberately no unbounded exported variant: an
   `ExportValues(tvs)` that silently skips the guard is what a future query
   handler would reach for by default.

3. **Wire the handlers** (`keeper.go`, `convert.go`) to pass
   `maxQueryExportBytes = 10_000_000` (~10 MB estimated). `stringifyJSONResults`
   takes the budget and returns `(string, error)`; its ~40 test callers go
   through a test-only `mustStringifyJSONResults` wrapper in `convert_test.go`,
   so no unbounded form exists in production code.

4. **Bound `vm/qeval` with the same budget**, by running the export walk as a
   size oracle before rendering: `QueryEval` calls `ExportValues` purely for its
   verdict, discards the copy (itself bounded by the budget), and renders with
   `TypedValue.String()` exactly as before for everything that fits. Reusing the
   walk keeps one estimate for all five endpoints instead of a second,
   divergent, render-shaped one. Its blind spot is persisted children, charged
   as a `RefValue` (~32 B) while `String()` renders them inline — that path is
   storage-deposit- and load-gas-gated, unlike the free ephemeral vector this
   bounds.

5. **Surface it as an ABCI error type.** `gno.ErrExportSizeExceeded` is mapped
   at the keeper boundary onto `vm.ExportSizeExceededError` (registered in
   `package.go`, like every other error in the module), so clients get a stable
   code for "response too large" rather than an untyped internal error. Without
   the amino registration the error cannot cross the wire at all — the RPC layer
   fails with `cannot encode unregistered concrete type`. `gnoweb`'s query
   client remaps that type onto its own `ErrClientResponseTooLarge` sentinel, so
   the state explorer renders a 502 ("upstream response too large") — the same
   page its own 8 MiB response cap produces — instead of a generic 500.

6. **Bound `vm/qtype_json` (`marshalTypeJSON`) with the same budget.** This
   endpoint does not go through the export walk: `QueryType` hand-rolls JSON from
   a `gno.Type` under a throwaway store with no allocator, so nothing metered it.
   `maxTypeDepth = 8` caps recursion depth but *not* fanout, and unlike
   `realm.go`'s `fillType` the walk re-expands each `DeclaredType.Base` per
   reference with no seal — so a struct DAG whose fields all reference the next
   named type expands into a `fanout^levels` tree. Measured: ~850 B of source
   (24 fields × 6 levels) emits ~36 MB, and it scales past that under 1 KB of
   source. `marshalTypeJSONBounded` checks the buffer length at entry to every
   recursive call and aborts with `ErrExportSizeExceeded` once it passes the
   budget, mapped to the same ABCI error as the walk endpoints. (This corrects
   an earlier assessment that filed `qtype` with the ~1× string endpoints below;
   it is the one real amplifier among them.)

This caps the node count *and* every string the export emits,
which is what removes the export+marshal amplification: the response is bounded
by a constant instead of by the size of whatever the expression produced.

## Alternatives considered

- **Post-marshal size check** (dora's proposed fix). `amino.MarshalJSON`
  materializes the full `[]byte` before any length check can run, so the
  multi-GB spike still happens — the check only rejects the already-built
  response. Rejected as insufficient for the headline vector.
- **Thread an `*Allocator` through export and charge every copy** (also in
  dora's fix). Bounds the export copy, but does nothing for the marshal (the
  real amplifier) and nothing for the string vector (export copies ~0). Would
  need the size guard anyway.
- **Size-limited `io.Writer` into a new `amino.MarshalJSONTo`.** Caps the
  streamed output, but a single huge scalar (e.g. a 1 GB string) is escaped
  into a full temporary buffer in one `Write` before the limit writer can
  reject it — so it does not prevent the scalar-string spike, which the
  pre-marshal estimate does. More invasive (new public amino API) for less
  coverage.
- **Exact size estimate.** Not attempted; a precise amino-output predictor is
  brittle. The estimate only needs to bound amplification, which a per-node
  lower bound plus content length achieves.

## Consequences

- A query whose estimated export size exceeds ~10 MB now returns
  `export size limit exceeded` instead of being serviced. These are
  explorer/debug endpoints; legitimate responses are far smaller (persisted
  objects export as `RefValue`, so only ephemeral values expand inline), so
  the practical blast radius on real usage is nil. The constant is a single
  named tunable if a legitimate consumer ever needs more.
- The *returned* response is bounded to within a small constant factor of the
  budget rather than exactly `maxQueryExportBytes`; the goal is DoS prevention,
  not an exact response cap. Callers must not assume a hard byte ceiling on the
  response.
- **A 10 MB response still costs far more than 10 MB of memory.** Measured at
  the production budget, the largest accepted results marshal to 11.6–18.5 MB
  of JSON (1.2–1.9× the budget, so the estimate tracks output well) but peak at
  **130–255 MB of heap**, because amino allocates roughly 20× the JSON it emits.
  So the guard converts a result-proportional amplifier into a constant one; it
  does not make these queries cheap. 10 MB is deliberately permissive — if a
  node shows memory thrashing under query load, the constant can be cut 10× to
  1 MB with no realistic consumer impact.

  | Shape (largest accepted at 10 MB) | JSON output | Peak heap |
  |---|---|---|
  | `[]Node` × 23,255 | 11.6 MB | — |
  | `[]*Node` × 18,552 | 14.7 MB | +131 MB |
  | `[][]int` × 44,642 | 17.3 MB | +140 MB |
  | one ~10 MB string | 9.5 MB | +57 MB |
  | `[]*Node` × 27,114 † | 18.1 MB | +240 MB |
  | `[]*Node` × 67,038, all → one persisted node † | 18.5 MB | +254 MB |

  † These two rows are sampled differently from the ones above them: max
  `HeapAlloc` polled every 0.5 ms during the query, against a `runtime.GC()`
  quiesced baseline read immediately before it. They land roughly 2× higher, so
  **~250 MB, not ~140 MB, is the residual to plan around** — the earlier figures
  understate the true peak rather than describing a different workload.

  The last row is the shape to watch: N ephemeral references to a *single*
  persisted object. Each element emits a `RefValue{ObjectID, Hash}` whose two
  strings the limiter never charges (only the enclosing node's 32 B), so it
  admits ~2.5× more elements than the same slice with distinct inline nodes for
  the same estimated size. It is simultaneously the cheapest such shape for an
  attacker to build — one persisted object, referenced N times from an ephemeral
  slice — and the most common legitimate one, an explorer's
  `func All() []*Post`. Charging those strings would tighten the estimate, at
  the cost of rejecting exactly the listing queries the JSON endpoints exist to
  serve; left uncharged deliberately, and recorded here because it sets the
  residual.

- `qobject_json` / `qobject_binary` expand the *queried* object inline (only its
  children become `RefValue`s), so a single large persisted object — an 8 MB
  byte array, charged 10.7 MB — is now rejected outright, with no partial or
  paginated view. Such an object had to be paid for in storage, so this is a
  real (if narrow) loss of introspection, not just an attack surface removed.
- The estimate is JSON-shaped and shared, so `qobject_binary` — whose output is
  raw bytes, not base64 — is rejected about 25% earlier than its actual response
  size warrants. Accepted deliberately: one estimate keeps the two object
  endpoints from drifting apart, and erring small is the safe direction.
- **This does not bring peak query memory down to "tens of MB" in general — it
  removes the export+marshal *amplifier* only.** The eval that produces the
  value still runs under `maxAllocQuery = 1.5 GB`, and that allocation happens
  before the export walk can reject anything: `Big(1_400_000_000)` still makes
  the node allocate ~1.4 GB for one free, unauthenticated query, it just no
  longer multiplies that by the copy and the JSON encoder. Bringing the floor
  down further means lowering `maxAllocQuery` for queries or metering export and
  marshal into the same allocator — deliberately out of scope here, but the
  reason the follow-up below matters more than "defense in depth".
- The export walk gains one addition and one comparison per node, plus a `len()`
  per emitted string; `add` inlines at every call site, so this is ~1 ns/node —
  negligible against the copy the four JSON endpoints already perform. For
  `qeval` the walk itself is new work, and it is not free on node-heavy results:
  measured, a 1,000-element struct slice adds 751 µs / 622 KB to a 2.85 ms query
  (~26%), and the largest shape the 10 MB budget accepts would add roughly 17 ms
  and 14 MB of discarded garbage. It stays free where the DoS actually lives: a
  large scalar string charges O(1) and copies nothing (117 ns / 80 B for a 1 MB
  string, against 5.9 ms / 4.1 MB to render it).
- Charges are conservative by design and in places doubled: an array pre-charges
  its element count and then charges each element again as it is walked, so a
  bare `int` element costs ~96 B of budget. At 10 MB that admits ~104K elements,
  not the ~312K a single 32 B charge per node would suggest.
- `ExportValues` / `ExportObject` are now `(…, maxBytes)` and return an error;
  every caller had to be touched. There is no unbounded exported variant by
  design (see Decision 2). Trusted callers and tests pass `maxBytes <= 0`, which
  installs a nil limiter and never errors, so their semantics are unchanged.
- Still unbounded, and left for a follow-up: the endpoints that return strings
  the VM itself produced — `qrender`, `qfile`, `qdoc`, `qfuncs`. Their
  amplification factor is ~1× (the string was already built under
  `maxAllocQuery`), so they are a response-size concern rather than a
  multiplier, and a generic ABCI response cap is the right shape for them. Note
  the two limits must be calibrated against each other: this guard's estimate is
  of the *input* to serialization and admits responses 1.2-1.8x the budget, so a
  handler-level cap below ~2x `maxQueryExportBytes` would reject responses this
  guard deliberately accepted. (`qtype_json` was originally in this list, on the
  assumption it was ~1× too; it is not — it re-expands a type DAG into a tree, so
  it is bounded here directly, see Decision 6.)

## Verification

- `gnovm/pkg/gnolang/values_export_limit_test.go` — limiter unit tests: single
  large string, many small nodes, the field-name/tag bypass regression, the
  `[]byte` base64 charge (rejected at a budget above the raw length but below
  `len × 4/3`), the `BigintValue` and `BigdecValue` charges (each rejected at a
  budget that covers every emitted digit, proving the estimate over-charges),
  `ExportObject` bounded on its own, and `maxBytes <= 0` still unbounded. The
  bypass test was confirmed to fail with the `exportCopyFieldsWithRefs` charge
  removed.
- `gno.land/pkg/sdk/vm/keeper_dos_test.go` — rejection through every wired
  endpoint: `qeval_json`, `qeval`, `qpkg_json`, `qobject_json` +
  `qobject_binary` via a persisted 100 KB byte array (resolved by ObjectID, so
  the shared `exportObject` path is actually exercised, not just assumed), and
  `qtype_json` via the wide type DAG (`TestQueryType_SizeGuard`: the amplifying
  root refused, an ordinary leaf type still served). Plus a measurement showing
  the bounded path allocates 80 B where the unbounded path churns ~21 MB on a
  2 MB-string result.
- `gno.land/pkg/integration/testdata/query_export_size_guard.txtar` — end-to-end
  through `gnokey` against the real 10 MB budget: small query OK on both eval
  endpoints; oversized string, slice-of-structs and long-struct-tag queries
  rejected on `qeval_json`, oversized string rejected on `qeval`.
- Test cost: the unit tests lower `maxQueryExportBytes` (hence it is a `var`,
  and hence they must not be `t.Parallel()`) and run on KB-sized inputs; only the
  txtar exercises the real constant. The earlier revision allocated 64 MB
  strings, 1 M-element slices and a 12 MB package var per case, which is enough
  to OOM a small runner.
- Existing `qeval_json` / `qobject_json` / `qobject_binary` txtars, the
  `gno.land/pkg/sdk/vm` package, and the gnoweb state suite still pass.
