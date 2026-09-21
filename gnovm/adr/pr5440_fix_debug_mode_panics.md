# Fix debug mode panics during uverse initialization

## Context

Running GnoVM tests with `-tags debug` panics during `init()` before any test
executes. Three separate issues in the debug code paths cause this:

1. **`GetLocalIndex` nil dereference**: Debug logging calls
   `reflect.TypeOf(sb.Source).String()`, but `sb.Source` is nil for the empty
   `PackageNode{}` stub returned by `UverseNode()` during re-entrant
   initialization (line 260 of `uverse.go`).

2. **`Define2` calls `TypeID()` on generic types**: `DefineNative` calls
   `Preprocess` on native uverse functions like `cap(x <X>{})`. During
   preprocessing, `initStaticBlocks` defines parameter names, then the
   TRANS_BLOCK case redefines them. The re-definition path in `Define2` calls
   `TypeID()` for a safety check, but `InterfaceType.TypeID()` intentionally
   panics for generic types (`Generic != ""`). Generic types are uverse-only
   placeholders resolved at call sites; they have no meaningful TypeID.

3. **`DelAttribute` on nodes without attributes**: `Preprocess`'s deferred
   cleanup calls `DelAttribute(ATTR_PREPROCESS_SKIPPED)` and
   `DelAttribute(ATTR_PREPROCESS_INCOMPLETE)` on every node via `Transcribe`.
   Many nodes never had these attributes set, so `attr.data` is nil, triggering
   the debug assertion.

## Decision

- **`GetLocalIndex`**: Check for nil `Source` and verify via
  `sb.Location.PkgPath` that it is the uverse package. Panic if nil Source
  appears outside uverse.

- **`Define2`**: Guard `TypeID()` call with `isGeneric()`. Add a debug
  assertion that verifies via location (own or parent) that generic types only
  appear in uverse.

- **`Preprocess`**: Guard `DelAttribute` calls with `HasAttribute` checks so
  the existing assertion in `DelAttribute` is preserved.

- **`UverseNode()` stub**: Set location on the empty `PackageNode{}` stub so
  debug code can identify it as uverse via `PkgPath`.

## Alternatives considered

1. **Weaken the assertions** (e.g., remove the `DelAttribute` panic, turn the
   `TypeID()` panic into `debug.Errorf`): This would hide real bugs. The
   assertions are correct -- the callers were wrong.

2. **Skip `TypeID()` without verifying uverse**: Using just `isGeneric()` is
   sufficient since generics can only be created via `GenT()` in `uverse.go`,
   but adding the location check makes the uverse-only invariant explicit.

## Consequences

- Tests can now run with `-tags debug`, enabling the full debug logging and
  pprof infrastructure for GnoVM development.
- All existing debug assertions are preserved; new assertions added to catch
  nil Source or generic types appearing outside uverse.
