# Move the realm parameter last in non-crossing helpers

## Context

A function is a crossing function when its first parameter is of type
`realm`. [`FuncType.IsCrossing`][iscrossing] reads position 0 and nothing
else, and [`preprocess.go`][preprocess] rejects a leading realm parameter
that is not named `cur`.

A non-crossing helper that still wants a realm value therefore has to keep
that realm away from position 0. The migration took the shortest route
available and prepended a discarded `int`:

```go
func inviteMembers(_ int, rlm realm, boardID boards.ID, invites ...Invite)
```

with every call site paying for it:

```go
inviteMembers(0, cur, boardID, invites...)
```

The `0` carries no information. It exists only to occupy the position that
would otherwise make the function crossing. It appears in 126 signatures
and 424 call sites, and readers have to know the convention before the
signature parses as anything but a mistake.

[#5786][issue] reports the same mechanism from the other end. A `/p/` package
cannot declare `func Sync(rlm realm)` at all: the leading realm makes it
crossing, the `cur` name check rejects `rlm` there, and [`crossingAllowed`][ca]
rejects crossing functions outside `/r/` anyway. The issue asks whether the junk
leading parameter is the intended workaround.

## Decision

Put the realm last instead, wherever there is another parameter for it to
follow:

```go
func inviteMembers(boardID boards.ID, rlm realm, invites ...Invite)
inviteMembers(boardID, cur, invites...)
```

The realm is off position 0, so the function stays non-crossing by exactly
the same rule as before. Nothing in the VM changes: `IsCrossing` still reads
position 0, the `cur`-name check still applies there, and crossing functions
keep the signature they already have. Parameter position is part of
`FuncType.TypeID`, so the property is carried by the type itself and
survives assignment through interfaces and func values.

Two shapes have nowhere to put a trailing realm and keep the sentinel:

```go
func checkCurrent(_ int, rlm realm) bool               // nothing to trail
func (a *MemberAuthority) AddMembers(_ int, rlm realm, addrs ...address) error
```

A realm cannot follow a variadic parameter, and with no other parameter at
all "last" is "first". 186 signatures fall in this group.

So this decision narrows [#5786][issue] rather than answering it. The
sole-realm helper the issue opens with is still unwritable, and the junk
parameter is still its workaround. The issue stays open.

## Alternatives considered

**Make crossing-ness depend on the parameter name.** Treat `func F(rlm realm)`
as non-crossing and only `func F(cur realm)` as crossing. This would remove
the sentinel everywhere, including the 186 signatures above, and would answer
[#5786][issue] outright.

Rejected: [`FuncType.TypeID`][typeid] is built from `UnnamedTypeID()`, so
parameter names are not part of type identity. `func(cur realm)` and
`func(rlm realm)` are the same type and are mutually assignable. Crossing-ness
is baked into `FuncValue.Crossing` from the declaring `FuncType`, while the
static call-site checks in `preprocess.go` read the *static* type at the call.
A func value could then be non-crossing while the call site type-checks as
crossing. Position does not have this problem because position is in the
TypeID.

**Trailing sentinel**, `func F(rlm realm, _ int)` called as `F(cur, 0)`, for a
uniform rule across all signatures. Rejected: it keeps the meaningless
argument and only moves it.

## Consequences

- Breaking for any caller of a changed signature. Removing the sentinel drops
  the parameter count by one, so a stale call site is rejected rather than
  silently passing a realm into the wrong slot. Where that rejection lands
  depends on where the call site lives: a `.gno` file under `examples/` fails
  the build, while a realm embedded in a `.txtar` archive is only compiled when
  the integration test deploys it, and fails there.
- Two conventions coexist rather than one. The trailing form is the default;
  the sentinel is what remains where the trailing form is not expressible.
- `gnovm/adr/interrealm_v2.md`, `pr_cross_explicit.md` and `pr5890_realm_sub.md`
  are left as written. They record decisions taken at the time, and the
  sentinel they describe was the convention then. `migration_guide.md` is
  instructional rather than historical and is updated.

## Verification

The rewrite was applied by a `go/ast` codemod rather than by hand. It resolves
call sites per package, via the `gnomod.toml` import path for a qualified call
and the declaring directory for a bare one, because bare-name matching would
rewrite `ulist.Set(0, v)`, which has nothing to do with realms. A call is only
rewritten when its second argument is a realm by syntax: a realm-typed
parameter in scope, a `cross()` of one, or `.Previous()`/`.Sub()` on one.

The codemod walks `.gno` files, so it never saw the five call sites in realms
embedded in `gno.land/pkg/integration/testdata/*.txtar` or the three snippets
in `docs/resources/`. Nothing in `main / lint` or `main / build` covers them
either. They surfaced only when the integration suite deployed the realm, as a
panic rather than a build error:

```
panic: msg:0,success:false,log:--= Error =--
	Data: vm.TypeCheckError{abciError:vm.abciError{}}
	    0  gno/gno.land/pkg/sdk/vm/errors.go:90 - gno.land/r/testing/resource/resource.gno:21:27:
	       cannot use 0 (untyped int constant) as string value in argument to a.DoByPrevious
```

The panic aborts the test binary, so every later case in that package goes
unreported: the first red run showed one failure where there were five. Sweep
the archives directly before trusting a green build:

```bash
grep -rn '<FuncName>(0, ' --include=*.txtar --include=*.md .
```

Two things the CI jobs do not say on their own. `go test ./...` in `gnovm/`
fails `TestFiles/types/*` and `TestTranspile/*` here, and fails the same set
on `ddb752cac` with no diff applied, so that red predates this change. The
gno2go job was not reproduced locally; it needs a full Go build of the
transpiled tree.

`gno lint` is worth one warning for whoever revisits this: it exits 0 while
printing type errors, so its exit status cannot gate anything. Both signature
desynchronisations this change caused mid-flight, an interface method left
behind by its implementations and a realm pushed to index 0 by a variadic,
surfaced only in that output.

[issue]: https://github.com/gnolang/gno/issues/5786
[ca]: https://github.com/gnolang/gno/blob/ddb752cac/gnovm/pkg/gnolang/preprocess.go#L4555-L4570
[iscrossing]: https://github.com/gnolang/gno/blob/ddb752cac/gnovm/pkg/gnolang/types.go#L1340-L1347
[preprocess]: https://github.com/gnolang/gno/blob/ddb752cac/gnovm/pkg/gnolang/preprocess.go#L818-L830
[typeid]: https://github.com/gnolang/gno/blob/ddb752cac/gnovm/pkg/gnolang/types.go#L1259-L1273
