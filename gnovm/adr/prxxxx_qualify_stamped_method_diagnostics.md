# Qualify stamped cross-package methods in every interface diagnostic

## Context

Interface method-set flattening records a method's defining package in
`FieldType.PkgPath` when that method is hoisted out of another package's
interface. The stamp makes a cross-package unexported method a distinct identity
from a same-spelled local one, so a single interface can legitimately hold two
methods named `sec`.

`FieldTypeList.string` qualifies a stamped method when it prints the interface,
and `VerifyImplementedBy` qualified it in the "missing method" error. The sibling
diagnostics did not, so a message could read:

```
main.T does not implement interface {filetests/extern/ifaceext.sec func() int; sec func() int} (method sec has pointer receiver)
```

It shows two distinct methods, then names neither. The same held for "wrong type
for method" and for the duplicate-method panic in `flattenInterfaceMethods`.

## Decision

Add `FieldType.diagName`, which qualifies a method name only when the method
carries a stamp, and route every `VerifyImplementedBy` error and the
duplicate-method panic through it.

Keying on the stamp alone is the rule `FieldTypeList.string` already applies, so
the method named in an error is spelled exactly as its entry in the interface
printed beside it.

## Alternatives considered

Reuse `idName`. Rejected. `idName` qualifies every unexported name, falling back
to the enclosing package, so a directly-declared unexported method would begin
printing as `main.sec`. That rewrites messages in unrelated tests such as
`access6` and `typeassert9a`, and it disagrees with how the interface itself is
printed.

Leave the diagnostics as they are. Rejected. The message contradicts itself, and
a reader cannot tell which of two same-spelled methods is at fault.

## Consequences

Message text only. Method resolution, interface satisfaction, type identity, and
the wire format are unchanged, and a directly-declared unexported method still
prints bare. Five filetests and two unit tests cover the qualified messages; four
of the filetests fail before this change, and the fifth pins the already-qualified
"missing method" case so the shared `diagName` path cannot regress it.
