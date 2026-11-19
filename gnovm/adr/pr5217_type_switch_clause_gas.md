# PR #5217: charge type-switch clauses by what the handler scans

## Context

`doOpTypeSwitch` charged for a type switch up front:

```go
m.incrCPU(OpCPUSlopeTypeSwitchCase * int64(len(ss.Clauses)))
```

`OpCPUSlopeTypeSwitchCase` (254) was fitted against `BenchmarkOpTypeSwitch_{1,10,100,1000}`, which measure a switch whose matching clause is last, so every clause is scanned. The constant therefore prices a *scanned* clause, but the handler multiplied it by the number of *declared* clauses.

The handler does not scan every clause. It breaks out of `matchLoop` on the first match, and it skips the default clause during the first pass entirely. So the charge diverged from the work in both directions:

- A switch matching an early clause paid for every clause behind it. A 100-clause switch matching first charged 25,400 for one `TypeID` comparison.
- The charge was blind to how many *cases* a clause carries. `case A, B, C:` costs three comparisons and was charged as one clause.

## Decision

Charge inside the match loop, per clause and per case actually reached, reusing the constants the value switch already uses for the same work:

```go
m.incrCPU(OpCPUSwitchClause)      // per clause reached
for _, cx := range cs.Cases {
    m.incrCPU(OpCPUSwitchClauseCase)  // per case reached
}
```

`OpCPUSwitchClause` (87) and `OpCPUSwitchClauseCase` (109) are measured constants for exactly this work — `doOpSwitchClause` and `doOpSwitchClauseCase` perform the same per-clause dispatch and per-case comparison in the value switch. A single-case clause now costs 196 rather than 254.

`OpCPUSlopeTypeSwitchCase` has no remaining user and is removed.

### Why this is not an under-charge

196 < 254 per single-case clause, so the worst case (all clauses scanned) charges 23% less than before. That is still a substantial over-charge against measured cost: a 100-clause switch scanning every clause was measured at ~13 ns/clause on a development machine, roughly 20 ns/clause scaled to the gas-table reference hardware, against 208 gas/clause charged — about 5x over.

The 254 ns/clause figure in `cmd/calibrate/op_bench_analysis.txt` is far above what a scanned clause costs today and looks stale independently of this change. Both it and the superseded `TypeSwitch (concrete) = 280.5 + 253.92*clauses` fit are flagged with a `TODO(calibration)` at the constants, to be re-derived when the reference-hardware numbers are next refreshed.

## Scope

Interface satisfaction is deliberately **not** touched here.

A type-switch case whose type is an interface calls `IsImplementedBy`, which walks the concrete type's embedding graph once per interface method. That walk is unmetered by size on this path and on several others (type assertions, selector dispatch, lazy method binding, `Stringer`/`error` formatting, and the preprocessor's assignability checks). Metering it correctly means threading a meter through all of those paths and pricing the walk by the graph it traverses rather than by any per-op constant — a change of a different shape and much wider blast radius, handled separately.

This PR leaves every one of those paths exactly as it is on master, so the two changes compose without conflicting over the same constants.

## Tests

Eleven filetests under `gnovm/tests/files/gas/`. Before this PR no `// Gas:` golden exercised a type switch or an interface type assertion at all, so these are new coverage for both:

- `typeswitch_clauses_{small,large}.gno` pin the change itself. The large one puts the matching clause last so every clause is scanned; a leading match would short-circuit and measure only source size.
- The nine `typeswitch_iface_*` / `typeassert_iface_*` fixtures pin current interface-satisfaction gas across flat, embedded-struct, embedded-interface and pointer shapes. They are unchanged in behaviour by this PR; their value is that any future repricing of the interface walk has to update them, making it visible in review.
