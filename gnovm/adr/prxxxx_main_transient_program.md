# Classify `main` as a transient program, not stdlib

## Status

Deferred: design note only, nothing implemented. Follow-up to #6218 (`pr6218_origincall_realm_boundaries.md`).

## Context

A `package main` (a filetest without `// PKGPATH:`, or `gno run`) gets its
`PkgID` from `PkgIDFromPkgPath("main")`. `IsStdlib` matches any dot-free
path, so `main` receives the stdlib bit (0x80), and since #6218 derives the
immutable bit (0x40) as "not `IsRealmPath || IsEphemeralPath`" it is immutable
too, while `nodes.go` still gives it a throwaway `Realm`. `isRecognizedPkgPath`
lists it by name so the classification is recorded, but it remains a
misclassification: `main` is a program with its own realm, the local twin
of a `gno.land/e/<addr>/run` package, and `/e/` is realm-class.

Everything below is local-only. `main` cannot be deployed (the keeper requires
`gno.land/` and `r` or `p`), so chain state and consensus are untouched.

## The change

In `PkgIDFromPkgPath`, skip the stdlib bit for `main` and add it to the
storage allowlist beside `IsRealmPath || IsEphemeralPath`, so its ID is
realm-class like a run path. `IsStdlib(path)` itself stays as is: it has about
thirty consumers (type-check import rules, mempackage typing, native
declaration handling, store keys, tooling) and none of them is the problem.

## What flips

Each consumer of the two bits starts treating `main` as a realm:

1. `ownsItsStorage` (AssertOriginCall): `main` frames count as a realm
   boundary. Filetests could assert from `package main` again; `std13`-`std17`
   could drop their `PKGPATH` headers.
2. `checkConstructionTime` / `stampPkgID` / `Copy` (alloc.go, values.go):
   types declared in `main` become realm-declared. Their values are stamped
   with `main`'s ID and may only be allocated while `main`'s realm is current,
   so a `main` struct constructed inside an `/r/` function panics with
   "cannot allocate ... in realm".
3. Borrow rules #2 and #3 (machine.go, `IsStdlibPkg` skips): a method on a
   `main`-owned receiver, or a `main` closure, called from an `/r/` realm now
   shifts `m.Realm` back to `main` instead of running as the caller.
4. `maybeFinalize` and the immutable/stdlib special cases in realm.go: `main`'s
   throwaway realm finalizes at every realm boundary, so its objects get
   refcount, dirty and escape bookkeeping; objects shared between `main` and an
   `/r/` realm change owner the way two realms' objects do.
5. The stdlib write-gate exemptions keyed on `m.Package.PkgID.IsStdlibPkg()`
   (machine.go) stop applying to `main` code, so a filetest that mutates a
   realm's object from `main` is gated like any foreign realm.

Items 2 to 5 are the cost: many `zrealm_*` filetests and some `gno run`
programs rely on `main` being transparent, and their golden outputs will move.
Item 1 is the only gain visible today.

## Validation plan

Run the full `gnovm/pkg/gnolang` filetests without `-short`, the examples
suite and `gnovm/cmd/gno` tests; classify every golden diff as either the
intended realm semantics or a test that should declare a `PKGPATH`. Land
`filetest.go`'s realm-mode decision (`IsRealmPath`) unchanged, so no new
`// Realm:` sections appear.

## Alternatives

- Keep the immutable classification from #6218 (current state).
  Zero cost, and the only observable oddity is that `package main` cannot
  assert an origin call in a filetest.
- Make `main` a run path outright (`gno.land/e/main/run` or similar) in the
  test and run loaders instead of special-casing the name in `PkgIDFromPkgPath`.
  Cleaner, but it changes every filetest's package path and error messages.
