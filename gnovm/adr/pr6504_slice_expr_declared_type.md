# Static type of a sliced declared string or slice

## Context

The VM types `x[a:b]` twice. Before the code runs, `doOpStaticTypeOf` gives it
a static type, which the preprocessor uses to convert an untyped constant
written beside it and to drop a conversion `T(x)` it judges redundant. When the
code runs, `TypedValue.GetSlice` builds the value and tags it with the type of
`x` whenever `x` is a string or a slice, declared or not.

For a declared string the static type was the plain `string`, and for a
declared slice the unnamed `[]T`, while the value carried the declared type.
`isEql` checks that both operands carry the same type before it reads a byte.
With `type Name string`, `n[0:1] == "_"` converted `"_"` to `string` and then
compared a `Name` with a `string`, so it answered false for every input. `!=`
answered true, a `switch` case never matched, and `string(n[1:])` was dropped as
a no-op conversion, so its value still reached `==` tagged `Name`. The builtin
`address` is a declared string too, so `addr[:2] == "g1"` answered false for
every address.

The same mismatch refused valid Go. `l[1:].Sum()` on `type L []int` found no
method on `[]int`, and `var head Name = n[:1]` failed with `cannot use string as
Name`. On a nil `L`, `l[:] == nil` answered false.

## Decision

`doOpStaticTypeOf` gives `x[a:b]` the type of `x` when `x` is a
`*DeclaredType` of string or slice kind. That is the type `GetSlice` puts on the
value, and the type Go gives the expression.

```go
} else if _, ok := xt.(*DeclaredType); ok &&
	(xt.Kind() == StringKind || xt.Kind() == SliceKind) {
	m.PushValue(asValue(xt))
}
```

A plain `string` or `[]T`, an array and a pointer to an array keep the static
type they had: `string`, `[]T` and `[]E`.

## Alternatives considered

- Fix the declared string alone. The declared-slice arm moves gas: a sliced
  declared slice that reaches a destination of its own type no longer gets a
  conversion inserted, so `Coins.Add` on two two-coin sets costs 569,329 gas
  instead of 570,490. Leaving slices out keeps that gas, but also keeps
  `l[:] == nil` false on a nil declared slice and `l[1:].Sum()` refused.
- Give every string and slice operand its own type, plain ones included. The
  branch gets shorter, but a variadic parameter's slice is then named `...int`
  rather than `[]int` in preprocess errors.
- Retag the value in `GetSlice` with the old static type. The comparison would
  then hold, but a sliced `Name` would lose its methods and match `string` in a
  type switch, where Go keeps the declared type.

## Consequences

- A node with the change and a node without it disagree on the result of a
  transaction that reaches such a comparison, and on the gas of every call that
  loses a conversion, so the change needs a coordinated upgrade. No gas golden
  moves and the app hash pin holds.
- The vm keeper re-preprocesses every stored package at boot through
  `PreprocessAllFilesAndSaveBlockNodes`, so deployed packages take the new
  types on the upgraded binary. Stored values keep the type they were written
  with. A value the old binary stored through `string(n[1:])` keeps `Name`,
  since that conversion was dropped, and a map key's bytes include its type, so
  after the upgrade a lookup by the same text misses that key. No deployed
  package slices a declared string; values stored by past `MsgRun` scripts were
  not searched.
- Code go/types refuses, such as `var s string = n[0:1]`, is now refused by the
  preprocessor too. Deploy-time type checking already refused it, so no stored
  package contains it.
- On 2026-10-28 all 689 `gno.land/` packages on gnoland-1 type-check and
  preprocess with the change. The new branch fires in them only on declared
  slices: `chain.Coins`, `bptree/v0.iterStack` and `markov/v0.Prefix`. None
  slices a declared string, so no deployed comparison changes its answer at the
  upgrade.
- `slice_declared_string.gno`, `slice_declared_string_typed.gno` and
  `slice_declared_slice.gno` in `gnovm/tests/files/` pin the comparisons, the
  typed constant and the assignment, nil equality, methods on a sliced value,
  and the array and pointer-to-array cases.
