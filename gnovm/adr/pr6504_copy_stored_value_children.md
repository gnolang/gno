# A value copy loads the arrays and structs it holds by reference

## Context

A stored array or struct keeps each array or struct nested in it as an object
of its own. Loaded from the store, the parent holds each such child as a
`RefValue`, and the child loads when something first reads it.

`TypedValue.Copy` duplicated a big integer, an `*ArrayValue` or a
`*StructValue` and copied every other value as it was, a `RefValue` included.
So a copy of a stored struct whose nested struct nobody had read yet held the
same `RefValue` as the original. A write through the copy loaded that child
and changed it in place, and the realm transaction committed the change to the
original:

```go
tmp := G       // G was stored by an earlier transaction
tmp.I.X = 99   // G.I.X is now 99 too
```

Whether a copy aliased the original depended on whether an earlier statement
of the same transaction had read the child. These by-value paths ran through
`Copy`: assignment and definition, the `copy` builtin, call arguments,
results, composite literals, `range`, conversion, map keys, and the receiver a
value method binds.
The receiver is the one another realm reaches: a value method of a realm's
type, called on that realm's stored variable, could change the variable.
Only the `append` builtin used `unrefCopy`, which loaded the outermost
`RefValue` and then copied its children through `Copy`, so it aliased one
level further down.

`doOpReturnToBlock`, which puts the results of `return x` into the result slots
before the deferred calls run, stored `x` without copying it at all. A deferred
closure writing a named result wrote into `x`.

## Decision

`Copy` takes the store, as do `ArrayValue.Copy`, `StructValue.Copy` and
`TypedValue.Assign`. A `RefValue` whose type is an array or a struct is loaded
and copied, so a copy owns every array and struct in it, however deep.
`unrefCopy` is gone: `append` calls `Copy`. `doOpReturnToBlock` copies each
result, as `doOpReturnAfterCopy` already did, except `return r` of the named
result `r` itself, which keeps `r`, so a pointer into `r` still sees what the
deferred calls write.

State stored before this change can hold a struct that contains itself:
`tmp := G` then `G.I.V[0] = tmp`, with `G.I` unread, stores one. So a copy
carries the IDs of the stored arrays and structs it is inside, whether it met
each as a ref or already loaded, and one met again on that path is copied as
a ref.

`len` and `cap` take their argument without copying it, since they read the
length alone.

A `range` over an array copies it only when the loop reads its elements, so a
key-only `range` over a stored array loads none of them. A `range` reading
values copies the whole array, as Go does, even when it breaks early.

A `RefValue` of any other type, a map, a function or a package, is still
copied as it is, since those are references in Go as well.

## Alternatives considered

- Loading every nested array and struct when its parent loads. It closes the
  same alias, but every read of a stored struct would load every child,
  whether or not a copy ever happens.
- Keeping `unrefCopy` as a wrapper around `Copy`. It would add one name for
  one behaviour.
- Panicking on a deleted object, as `unrefCopy` did through `GetObject`. That
  refuses `commondao_dao_member.txtar`, a transaction the base completes.
- Recording one copy per object and reusing it. A shared child is then copied
  once, but the copy keeps the sharing between its parents, and every copy of
  an array or struct holding another needs a map.

## Consequences

- A transaction that copies a stored value with a nested array or struct not
  yet read now pays to load it and to allocate its copy, about 63,500 gas for
  each one: a value-receiver method on a stored struct holding
  `[100]struct{ X int }` went from 1634982 to 7976922 gas. A copy that finds
  every nested array and struct already loaded costs the same as before.
- A value the base stored by boxing a stored struct into an interface field of
  a new one shares the old struct by ref, so a list pushed with
  `Head = Node{V: v, Next: Head}` cost the same at any length. A copy of it
  now copies the whole chain, as Go does: such a push measured 66654 gas on
  the base at 10 and at 40 nodes, and 86244 and 162534 here, so a chain the
  base let grow long enough no longer fits in a transaction.
- A child the base stored under two parents of one value is copied once per
  path to it, so the cost doubles with each level of such sharing: a value
  the base built with `n.V[0] = N; n.W[0] = N; N = n` copies for 14155 gas on
  the base at any depth, and here for 1215241 gas at 10 levels, 18439972 at
  14, and runs out of a 100000000 gas limit at 18.
- Storing back a copy of a struct the base left containing itself, `G = G`,
  keeps the original under the copy's ref, so each such transaction adds a
  level: over six of them `G = G` cost 40857, 47479, 54104, 60729, 67357 and
  74084 gas.
- A realm's copy of another package's value used to share that package's
  unread children, so writing them failed as a write to a readonly object.
  `X = pp.Get(); X.I.D.X = 5` and a value method writing `o.I.D.X` on
  `pp.D` now succeed, and the package's value stays as it was.
- A transaction that wrote through a copy now leaves the original as it was,
  and stops paying to store the changed child.
- A method value a realm stored before this change, bound to a value receiver
  of a stored struct, shared the receiver's unread children by ref; a call to
  it now copies a real receiver first, so a value method writing its receiver
  leaves the stored original as it was. A receiver bound after this change is
  already a private copy, so the copy is a no-op for it.
- A function with deferred calls that returns an array or a struct now
  allocates a copy of it: returning a two-field struct went from 1096370 to
  1096480 gas.
- `copy_stored_value_children.txtar` writes through a copy of a stored struct
  to a nested struct, an array and a struct two levels down, through the
  elements each `append` path and `copy` take out of a stored slice, through a
  value method another realm calls, a call argument and the result a deferred
  call writes, each in a transaction of its own, and runs a key-only `range`,
  `len` and `cap` over a stored array within a gas limit a copy would exceed.
  `gnovm/tests/files/defer_result_copy.gno` pins the deferred write to a named
  result, `defer_named_result_ptr.gno` a pointer into a named result `return`
  keeps, and `TestCopyStoredCycle` copies a stored struct that contains itself,
  as refs and as loaded objects. `zrealm_cross_copymethod.gno`,
  `zrealm_cross_copywrite.gno` and `zrealm_cross_copydefer.gno` run a value
  method, a written copy and a deferred write to a named result on a stored
  struct while another realm is current.
