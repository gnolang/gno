# Validate type-declaration cycles before predefinition

## Context

Issue #6036: gno rejected legal cyclic type declarations that pass through
aliases, and its verdict on some others depended on declaration order. The
Go rule (golang/go#25838): a cycle among type declarations is invalid when
every edge on it is a direct containment (infinite size) or every
declaration on it is an alias (nothing to resolve to); every other cycle is
legal.

Cycle detection lived in the dependency walk (`findUndefinedAny`), which
resolves one declaration at a time, stops at the first undefined name and
forgets a name once it has a slot, so it only ever sees one path through
the graph. A member can be reached before it has a slot (`type A B` with B
pointing back: `should not happen`, `name not defined`), and a direct cycle
whose first reference on the path is a pointer is missed, built, and
overflows the Go stack on `TypeID`. PR #6048 kept the walk's verdict and
added a second walk over the built types, as go/types does because it
resolves lazily; gno holds the whole declaration set in `PredefineFileSet`.

## Decision

Judge and reserve the type group first, then build (`typedecl_group.go`,
`predefineTypeDeclGroup` at the three predefinition sites: `PredefineFileSet`,
its single-file twin in `Preprocess`, each function-body type declaration).

1. **Pre-scan.** Collect each declaration's references with a direct flag
   (struct field, array element, embedded interface, bare name, and a
   selector whose base is a member are direct). Two cycle searches apply
   the rule, one over direct edges and one over alias-to-alias edges; a
   single DFS judging the cycle it closes first is wrong when a legal and an
   invalid cycle share nodes (`TestTypeDeclInvalidCycle`). A cycle is
   reported with its earliest-declared member first.
2. **Reservation.** Every member gets its slot before any is predefined: a
   shell of its kind filled in place at LEAVE, or for an alias of a member
   the slot its chain ends at. A legal cycle closes on any member whichever
   is predefined first. `tryPredefine` builds nothing; it resolves only an
   alias of a name outside the group, which cannot be bound before imports.
3. **The seam with values.** Type expressions name every dependency
   syntactically, so types are judged statically; values stay on demand with
   the real evaluator, as in go/types. Two rules join them: a directly
   contained type is built before its container, since sealing looks
   through it; and outside a type declaration's own expression a reserved
   type counts as undefined until built, so a constant, a variable's type
   or a signature that names it builds it first (`decltype_constlen*`). A
   type cycle through an array-length constant or a variable is reported by
   the walk's `defining` check (`recursive16`–`18`).
4. **Settled-type checks** run once the group is built: embed depth and
   map-key comparability. `Seal` met a pointer-embedded member before its
   base was set (`embed_depth1`), and a map key can be a shell when its map
   is constructed, which would also poison the `comparable` memo
   (`maptype2*`).
5. **Gas.** The walks bill the preprocess meter through `chargeCPUGas`, as
   `embedWalk` does: per declaration, per type-expression node (dependency
   scan, map-key walk), per graph step (cycle edge, `embedDepth` visit).
   Slopes are dev-box fits from `BenchmarkTypeDeclGroup` converted with a
   machine factor measured as in #6164 (gas table over measured pure ns of
   seven flat ops, 1.9 to 2.2; 2.1 used), rounded up to floors; recorded in
   `cmd/calibrate/typedeclgroup_bench_m5_arm64.txt`, re-fit on reference
   hardware pending. Runtime inline struct/interface construction bills the
   same depth walk, which #5642 added unmetered; eleven `gas/` goldens move.

## Alternatives considered

- **Two verdicts as in #6048.** Matches go/types, but the walk keeps its
  `direct` plumbing, map keys need a settledness walk, and the slot crash
  family stays (24 of its 25 harness disagreements). Here nothing is built
  before the verdict; the alias-only pass is what keeps reservation safe,
  since `type A = []B; type B = A` would otherwise fill A's shell with itself.
- **A static graph over values too**, making predefinition one topological
  fill. Rejected: a composite-literal key (`S{a: 1}` vs `map[K]int{a: 1}`)
  is a field name or a reference depending on the literal's type, unknown
  before anything is built, so the graph would need a second type resolver
  kept in step with `evalStaticType`.
- **Extend the walk to see the whole graph.** It would record every edge of
  every declaration, which is this pre-scan under another name.

## Consequences

- Invalid cycles are reported before any type is built, with the full path;
  `recursive9i`/`9j` goldens changed from `refers to itself` to that form.
- `findUndefinedV/T/Any` and `tryPredefine` lose `direct`, `stack` and
  `defining`; `predefineRecursively2` keeps them for value cycles and the
  cycles through values above. `declareWith` is called only at reservation.
- Bare-name references into a cycle and alias chains out of the group run
  (`recursive14`, `15`, `decltype_mutual`, `alias_chain_uverse`); #5763 is
  superseded. Function-local forward references still fail as before.
- Differential harness, 373 groups judged by go/types: master 91
  disagreements and 3 stack overflows, #6048 25 and none, this branch 2,
  both channel types, which gno rejects outright.
- Each gas slope is pinned by a test that fails when it is zeroed
  (`typedecl_group_gas_test.go`).
