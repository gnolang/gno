# PR5609: Enforce Go's addressability rules at preprocess

## Context

The VM accepted expressions Go rejects. `&(a + b)`, `&m["k"]`, and
`getArr()[:]` either panicked at runtime with an internal message or silently
did the wrong thing. Issue #5586.

The original PR added an `isAddressable` helper and called it from two places:
the operand of `&` (the `*RefExpr` case) and the slice-of-array case. Review
found the helper still let several forms through, and that one address the VM
takes is never checked at all.

### What the helper got wrong

`isAddressable` answered "yes" for whole node kinds rather than for what the
node denotes.

- `*NameExpr` was addressable unconditionally, but a name can denote a declared
  func, a type, or a package. `&someFunc` produced a working `*func()`, and
  writing through it replaced the function.
- `*CompositeLitExpr` was addressable unconditionally. That is true only for
  the *direct* operand of `&` (the spec's stated exception), not for a literal
  used as the base of something else, so `&T{}.f`, `&[3]int{1,2,3}[0]`, and
  `[3]int{1,2,3}[:]` were all accepted.
- The `*IndexExpr` arm had no `StringKind` case, so `&s[0]` on a string
  recursed into the string variable and came back addressable.
- The `*SelectorExpr` arm did not separate a field from a method value, so
  `&t.M` was addressable.

### The address nobody checked

`x.M()` where `M` takes a pointer receiver is shorthand for `(&x).M()`, and the
preprocessor builds that `&x` itself. That synthesized `RefExpr` is marked
preprocessed and never re-enters the `*RefExpr` case, so no addressability
check ever saw it — even though the spec sentence the code quotes at that site
begins "If x is addressable". The consequences were worse than a missing
diagnostic:

| written | before |
|---|---|
| `m["k"].Inc()` | accepted, and **mutated the stored map element** |
| `T{}.Inc()` | accepted, write silently discarded |
| `mk().Inc()` | `illegal assignment X expression type *gnolang.CallExpr` at runtime |
| `m["k"].Inc` (method value) | accepted |

Go rejects all four with `cannot call pointer method Inc on T`.

## Decision

Make `isAddressable` answer the spec's question — *is this a variable, a
pointer indirection, a slice index, a field of an addressable struct, or an
element of an addressable array* — and call it everywhere an address is taken,
including the synthesized one.

1. **A name is addressable iff it denotes a variable.** Delegated to
   `StaticBlock.IsAssignable`, which already records the fact declaratively:
   a package-level func decl is appended to its block's `UnassignableNames`,
   and uverse names are refused. It resolves in the innermost block the name
   appears in, so a local shadowing a func keeps its own answer, and
   `UnassignableNames` is serialized with the package node, so an imported
   package loaded from the store answers the same as one preprocessed in
   process. Types and constants cannot reach the check — both fold to
   `constTypeExpr`/`ConstExpr` first.
2. **The composite-literal exception moved to the `&` call site.** It is an
   exception for that operand only, and `isAddressable` now also serves
   slicing and the pointer-method receiver, where the exception must not
   apply.
3. **`*StarExpr` is an indirection only when its operand is not a type.**
   `*T` in the method expression `(*T).M` is a pointer *type* expression that
   survives preprocess as a `StarExpr`; it is not an indirection of anything.
4. **Index and selector arms follow the spec**: string and map indexes are
   never addressable; a pointer-to-array index is (`p[i]` is `(*p)[i]`); an
   array element is addressable iff the array is; method values are not fields.
5. **The synthesized receiver address is gated**, reported as
   `cannot call pointer method M on T` against the type written at the call
   site rather than the embedded type a promoted method resolved to.
6. **Untyped `nil` is reported instead of crashing.** `&nil` has no static
   type, and the existing error path formatted that type — a nil dereference
   that took the preprocessor down with a Go runtime panic.

## Alternatives considered

**Reuse `assertValidAssignLhs`** (suggested in review). Rejected: addressable
and assignable are different sets and neither contains the other. A map index
is assignable but not addressable — `m[k] = v` is legal, `&m[k]` is not — so
reusing that predicate would have re-admitted the exact cases this change
rejects. It accepts *any* `*SelectorExpr` and any non-string `*IndexExpr`.

The useful direction is the converse: Go's assignability rule is "addressable,
or a map index expression", so `assertValidAssignLhs` could be expressed in
terms of `isAddressable`. That would fix a live family of bugs on the
assignment path (`m["k"].n = 1` and `m["k"].n++` are accepted today and the
write lands on the stored element; `mk().n = 1` and `T{}.n = 1` are accepted
and discarded). It is a separate change to a different code path, and is left
out here deliberately — see below.

**Keep the check purely syntactic.** Rejected: a name's addressability depends
on what it denotes, which is not recoverable from the node kind. This is why
the original helper could not distinguish `func bar()` from `var bar func()`.

**Detect a declared func by sniffing its static slot for a `*FuncValue`.**
This works — a declaration's slot carries its value at preprocess time while a
variable's does not — and was the first implementation. Replaced by
`IsAssignable` because the slot shape is a proxy for a fact the block already
records, and because a store-loaded package node can hold a `RefValue` where
the sniff expects a `*FuncValue`, which would silently classify a func as a
variable.

## Consequences

Preprocess now rejects code it used to accept. Gno code that relied on any of
these forms stops compiling — but `AddPackage` and `Run` have always run
`TypeCheckMemPackage`, which rejects all of them via `go/types`, so no package
could have reached the chain through the normal path. The remaining exposure is
a path that skips type-checking (genesis seeding), where a stored package using
one of these forms would now fail to re-preprocess on load. `examples/` (222
packages) and the stdlibs are clean.

Gas is unaffected. Preprocess gas is `PreprocessGasPerByte * srcBytes`, charged
before preprocessing runs, so it is a function of source length and not of the
work the preprocessor does. The added checks are bounded per node:
`isAddressable` runs once per `&`, once per array slice, and once per
pointer-method rewrite, recursing only down `.X`, and `evalStaticTypeOf` is
served from the `ATTR_TYPEOF_VALUE` cache because children are visited first.
The package-qualified branch reads the `*PackageValue` out of the import slot
rather than calling `evalConst`, which would spin up a sub-Machine that
inherits the transaction's gas meter and bill for it per occurrence.

### Known gaps left open

- **Identity conversions.** `&T(v)` and `&T(T{1})` are accepted; Go rejects
  both. An optimization elides a conversion whose source and target `TypeID`
  match, so by the time the `&` is checked the conversion is gone and the
  operand looks like `&v` or `&T{1}`. Fixing it means changing that
  optimization, not the addressability rule, so it is out of scope here.
  `go/types` rejects both on the on-chain path.
- **The assignment path**, as described under Alternatives.
- Both are pre-existing and unchanged by this PR.

## Verification

- A differential harness ran 70 snippets through real `gc` and through the VM
  and compared accept/reject. All 70 agree now; 16 disagreed before this
  change. Run under go1.25.9, the version `go.mod` pins: `go/types` reworded
  several of these diagnostics after 1.23, so a comparison built on an older
  toolchain reports wording mismatches that are not real. The harness checks its own expectations against `gc` rather than
  trusting them, which caught one case labelled wrong by hand and one that is
  an intentional Gno/Go divergence unrelated to addressability (Gno forbids
  shadowing builtins).
- 15 `_err` filetests, each confirmed to fail against the unfixed
  preprocessor, so each one pins a real behavior change. Two positive
  filetests (`addressable_12.gno`, `addressable_13.gno`) cover the forms that
  must keep working — variables holding funcs, and pointer methods on every
  addressable receiver shape including one reached through an embedded
  pointer.
- `go test ./gnovm/...`, the filetest suite, `gno test ./...` over `examples/`
  (222 packages), `sdk/vm -run Gas`, and `integration -run TestTestdata`: all
  pass.
- CPU: a 124 KB file of 8000 statements built only from affected forms
  preprocesses in the same time before and after (0.28–0.30 s, indistinguishable
  across runs).

The `TypeCheckError` lines in `addressable_6e_err.gno` and
`addressable_9d_err.gno` were updated for `go/types` wording that changed in
the toolchain (`invalid operation: cannot slice x ... (value not addressable)`
became `cannot slice unaddressable value x ...`). Those two edits are toolchain
drift, not a consequence of this change.
