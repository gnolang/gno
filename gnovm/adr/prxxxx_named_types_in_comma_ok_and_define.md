# Named types in comma-ok and `:=` assignment

## Context

Go accepts both of the following programs; the Gno preprocessor rejected them.

**Comma-ok into a declared bool.** `parseMultipleAssignFromOneExpr` gave the
second value of a comma-ok a static type of `BoolType`:

```go
tuple = &tupleType{Elts: []Type{tt, BoolType}}        // x.(T)
tuple = &tupleType{Elts: []Type{mt.Value, BoolType}}  // m[k]
```

The runtime value, from `doOpIndex2` and `doOpTypeAssert2`, is an untyped bool.
`checkAssignableTo` rejects `BoolType` against a declared `type B bool`, since
two named types need an explicit conversion, so `var v, ok B = m["k"]` panicked
at preprocess with `cannot use ... (value of type bool) as main.B value`. Go
treats the result as an untyped boolean, which is assignable to any boolean
type.

**`:=` that names an existing variable.** When a `:=` target is already declared
in the same block, Go assigns to it: the variable keeps its type, the value must
be assignable to it, and the implicit conversion `a = x` would apply. Gno instead
handed the value's type to `Define2`, whose redefinition branch panics with
`StaticBlock.Define2(a) cannot change .T; was main.NamedMap, new map[string]int`.
`defineOrDecl` recorded nothing for the `a = x` conversion in that case either,
because it took the `isDefine && ok` branch unconditionally.

Both defects were reported in #6493.

## Decision

Give the comma-ok result its Go type, and treat a `:=` target that already
exists in the block as an assignment.

- `parseMultipleAssignFromOneExpr` builds the tuple with `UntypedBoolType` for
  the ok result of both the type-assertion and the map-index form.
- `defineOrDecl` collects, for a `:=`, the declared static type of every target
  that already has a slot with a type. It checks the value against that type
  with `checkAssignableTo`, records it in `ATTR_ASSIGN_CONVERT`, and keeps the
  declared type in the target's slot. `doOpDefine` converts the values with
  `convertTopValues` before popping, the same way `doOpAssign` does.
- A slot that exists but has no type yet is only reserved, so `Define2` still
  fills it in: such a name is declared here, not assigned to.
- A newly declared variable takes the default type of an untyped value, so
  `v, ok := m["k"]` declares `ok` as `bool`, not untyped bool. Because the
  placeholder in the slot has to agree with the slot type — a repeat preprocess
  of the same block compares them — the placeholder carries the default type
  too. A `const` keeps its untyped type and its value.
- `doOpValueDecl` converts an untyped result to the declared type rather than to
  its default type. `var v, ok B = m["k"]` then binds `ok` to a `B`, as
  `var ok B = true` does; before this change only a default `bool` was produced,
  which is not what the declaration asks for.

## Alternatives considered

1. **Keep `BoolType` in the tuple and special-case a declared bool in
   `checkAssignableTo`.** This hides the defect in the assignability check
   rather than modelling what Go's type checker does: the result is an untyped
   constant. It would also leave the type wrong for every other consumer of the
   tuple.

2. **Convert the ok value by wrapping the expression in `B(...)` at
   preprocess.** A comma-ok expression yields two values, so it cannot be
   wrapped in a conversion; the conversion has to happen on the value stack,
   which is what `ATTR_ASSIGN_CONVERT` is for.

3. **Turn the `:=` into an `AssignStmt` at preprocess.** The statement is the
   parsed form of the source and carries the paths of the names it declares;
   rewriting its kind would change what the runtime executes and what the
   source mapping reports. Recording the conversions keeps the statement's shape.

4. **Declare the `:=` target with the value's type anyway and let later uses
   coerce.** This is the current behaviour and is exactly the panic being fixed;
   the declared type is observable through `reflect`-style type reporting and
   through method sets, so it must not change.

## Consequences

- `var v, ok B = m["k"]`, `v, ok := x.(B)`, and `a, b := f()` where `a` has a
  declared type now run and produce the values Go produces.
- A `:=` that only ever names variables already declared in the block is still
  rejected with `no new variables on left side of :=`, as in Go; this change
  does not affect that check.
- An untyped result declared with an explicit type now binds that type. Code
  that relied on a default `bool` being produced for a declared bool target
  would see the declared type instead, which is what the declaration says.
- `ATTR_ASSIGN_CONVERT` on a `:=` statement is a new producer of the attribute;
  `seal.go` already walks it for every node that may carry it, so no new
  sealing path is needed.
- The regression tests are `gnovm/tests/files/assign_named_comma_ok.gno`,
  `assign_named_type_assert_comma_ok.gno`, and `assign_redefine_named_type.gno`.
