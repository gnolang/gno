# chain/reflect: objects become addressable

## Context

A realm that creates something and needs a unique name for it has, otherwise,
only names it chooses itself: a package path, a symbol, a counter it keeps, or
an ID its caller passes in. Anything a realm chooses, a realm can choose twice.

`grc20.NewToken` is the case that surfaced it. It built `Token.id` from
`<pkg-path>.<symbol>.<caller-supplied-seqid>`, with the sequence number coming
from the calling realm, so one realm could issue two independent tokens under
one identifier and their `Transfer`/`Approval` events became indistinguishable.
That is #6026.

The VM already maintains the one name a realm cannot choose. `ObjectID` is
`{PkgID, NewTime}`: the realm an object belongs to, and a tick of that realm's
storage clock, minted by the VM while persisting the object.

Two earlier attempts:

- #6101 added `runtime.NewRealmID()`, returning `<realm-path>:<realm-time>`.
  Rejected in review: a second identity scheme no other VM concept consumes.
- #6139 derives a bech32 address from the ObjectID and exposes it as
  `chain/runtime.ObjectAddress(v interface{}) string`, together with a breaking
  `grc20` and `grc20reg` redesign, across 50 files.

## Decision

Add `chain/reflect`, an injected stdlib whose subject is object identity, and
give every object an address.

```gno
type Addressable interface{ Address() address }

type Object struct { /* unexported */ }

func Of(v interface{}) (Object, bool)
func HasIdentity(v interface{}) bool

func (o Object) Address() address // g1…, derived from the identity
func (o Object) ID() string       // "<realm-id>:<clock-tick>"
func (o Object) PkgPath() string  // the realm that created it
func (o Object) Type() string     // "<pkgpath>.<Name>"
func (o Object) String() string   // "gno.land/r/demo/foo.Token#7 (g1…)"
func (o Object) IsZero() bool
```

### 1. The address is the deliverable

An object address is an ordinary gno.land address, derived by hashing
`objectid:<pkgid>:<newtime>`, the same shape as a package address's
`pkgPath:<path>` and in a disjoint preimage space.

Being an address, not a bespoke identifier, is what makes an object usable by
code that has never heard of objects. `grc20`'s `Transfer(_ int, rlm realm, to
address, amount int64)` already takes a plain `address`, so an object can hold
token balances the day this lands, with no change to grc20. The banker's
recipient is likewise an unconstrained bech32, so an object can receive ugnot.
That is the difference between "a unique string a realm can key a map by" and
"a place on chain where value sits".

An earlier draft of this work returned the raw ObjectID and deliberately
withheld address derivation, on the grounds that the preimage freezes the first
time an address reaches an event and should not be fixed before the accounts
capability exists. That argument was rejected, and correctly: the capability is
the point, not a later bonus. Withholding the address gives realms a unique
string and leaves every consumer needing a translation layer, which is the
problem being solved.

What does not exist yet is **spending**. A banker's sender is pinned to the
calling realm's own package address, so a realm cannot move funds held by an
object it owns. Until that lands, an object address receives and does not
release, and both the package doc and the docs page say so in those words.

### 2. Object is a type, not a string

`Object` carries unexported fields, so `Of` is the only thing that returns a
non-zero one. The whole point is an identifier the realm does not choose, so
making it unforgeable by construction rather than by convention is the property,
not a detail. It stays comparable and map-key safe.

#6139 returns a bare `string`. A realm can therefore fabricate an "object
address", and a consumer taking a `string` cannot tell a VM-issued identity from
one the realm typed out, which is the same class of problem as #6026 itself.

`Addressable` is the deliberate exception, and its doc says so: any type can
implement `Address()` and return whatever it likes. It exists to let an API
accept a user, a realm or an object without caring which. It is safe for
deciding where value goes and unsafe for deciding who may move value; authority
comes from the VM, never from satisfying an interface.

### 3. Provenance travels with the address

`PkgPath()` answers "which realm created this", resolved through
`Store.GetRealmByID(PkgID)`. `Type()` answers "what is it", from the value's
`*DeclaredType`. Together with the address they are enough to render a
deep link to an object or its funds, which is the case that motivated them: a
governance proposal whose fund address and executable body are both separately
addressable and separately linkable.

The two can differ, and the difference is information: a realm that
instantiates a type from a `p/` package creates the object under its own path,
so `PkgPath()` is the realm and `Type()` names the package.

**Source position is not available and is not added.** `DeclaredType.ParentLoc`
is blank for every package-level type (`declareWith` in `types.go` keeps it
blank for `*PackageNode`/`*FileNode`), and `ObjectInfo` carries no allocation
site. Recording one would add a field to every persisted object in every realm:
a storage cost on all state, permanently, for a debugging nicety. Package path
plus type name plus address is enough to link to a thing.

### 4. One outcome per case, no panics

| v | `HasIdentity` | `Of` ok |
|---|---|---|
| pointer to a standalone object, persisted | true | true |
| pointer to a standalone object, created by the running call | true | false |
| func value or bound method, persisted | true | true |
| pointer into a struct field or array element | false | false |
| slice | false | false |
| scalar, nil pointer, nil interface | false | false |

`Of` never panics and never invents an address. `HasIdentity` splits its two
false cases into "ask again after the realm boundary" and "nothing to wait for".

#6139 has one function with three behaviours: `""` for an unpersisted object,
`""` for a pointer into a container or a slice, and a panic for a scalar. The
two that need different handling are the two that are indistinguishable.

Func values and bound methods are included because a func value is a reference:
two variables holding one closure hold the same object, so every holder agrees
on the address. That is what makes a stored callback addressable separately
from the struct that holds it, which is the "link to this exact lambda" case.

A pointer into a struct field or an array element resolves to the container,
so lending it the container's address would make every sibling answer the same
one. A slice resolves to its backing array. A struct, map or array arriving by
value may be the persisted object or a copy of it, depending on how it got
there. The set is meant to grow and only to grow: widening is compatible,
withdrawing an answer is not, because by then that address is in a balance.

### 5. The stamping window is reported, not papered over

`PkgID` is stamped by the allocator at creation; `NewTime` only by
`Realm.assignNewObjectID`, reached from `FinalizeRealmTransaction`, which runs
when a realm frame returns. So an object created by the running call has no
complete identity and no address, and no primitive can conjure one: attaching
to the ownership tree *is* what mints it. Stamping early is a lifecycle change,
since `GetIsReal()` *is* `ID.IsFinalized()` and `incRefCreatedDescendants`
guards on the same predicate, so an early-stamped object looks real while
unsaved and gets skipped during finalization. #6139's analysis of this is
correct and is reused here.

What is a choice is how the window is reported. `Of` returns `ok == false`
rather than an empty address. With an address the stakes are higher than with an
opaque ID: an empty address written into an event is permanent, and a caller who
treats it as real directs funds at a string that is not an account.

## Consequences

- No change to `misc/genstd`. The native's Go parameter is `gno.TypedValue`,
  which already matches any Gno parameter type and links through with no
  Go2Gno conversion, while the Gno side declares `interface{}` and
  `GnoTypeExpression` maps it to `gno.AnyT()`. Pinned by the existing
  `misc/genstd/testdata/linkFunctions_TypedValue`. #6139 instead widens
  `mapping.isTypedValue` so a *Go* `any` parameter is treated as a raw
  TypedValue, which is unnecessary and makes every future native declaring a Go
  `any` silently receive a TypedValue rather than a converted Go value.

- The preimage layout is consensus state from the first address that reaches an
  event, a registry or a balance. The prefix, the separators, the `PkgID`
  encoding (whose flag nibble encodes IsStdlib/IsImmutable/IsInternal) and the
  `NewTime` semantics all become fixed, and changing any of them moves every
  object address in existence. The preimage is taken from #6139 unchanged so
  the two approaches stay interchangeable.

- The native resolves pointer bases itself and never calls
  `TypedValue.GetFirstObject`, which panics outright on a package `RefValue`
  and on a bare `*HeapItemValue`. Every shape is handled or reported as no
  identity.

- `PkgPath()` costs a realm lookup. It is a cache hit in the normal case: the
  ID is finalized, so the object is persisted, so its realm exists and is
  loaded for the object to have been reached at all.

- Stdlib `.gno` source is genesis state, so adding the package moves the
  genesis app hash (`expectedCrossrealm38Hash`) whether or not anything imports
  it. Storage-deposit txtars are unaffected.

- Gas: one flat row at 1200, dominated by the sha256-and-bech32 derivation.
  Measured end to end through the dispatcher, 1196ns stamped, 106ns unstamped,
  90ns with no identity. Draft pending a re-fit on the reference machine; see
  the row's own comment for the reference-row ratios it was placed against.

- `grc20` is untouched. With this in, the #6026 fix is a small follow-up: drop
  `NewToken`'s caller-supplied `seqid.ID`, keep `rlmPath.symbol` as the token's
  readable name, and let `Token.Address()` be its identity.
  `zrealm_reflect_duplicate_symbol.gno` demonstrates that end to end on a local
  token type. GRC721 has the same exposure and the same fix.

## Alternatives considered

- **`chain/runtime.ObjectAddress`, as in #6139.** `chain/runtime` is `ChainID`,
  `ChainDomain`, `ChainHeight`, `AssertOriginCall`: information about the
  running chain. Which object a value is, and where it came from, is not that,
  and putting it there leaves no home for the next question of the same kind.
- **`vm/reflect` or `gno/reflect`.** Neither `vm/` nor `gno/` exists as a
  stdlib namespace. Checked while evaluating it: a new top-level namespace
  needs no core change, since `Re_gnoStdPkgPath` is just `name(/name)*` with no
  whitelist. It is a taste call, and `chain/` is where the non-Go-portable
  stdlibs already live.
- **A top-level `reflect`.** Collides with Go's `reflect` for anyone reading or
  transpiling the code, and promises an API this will not have.
- **Returning the ObjectID and withholding the address.** Covered above:
  rejected, because a unique string leaves every consumer needing a translation
  layer, and the address is what makes objects usable by existing code.
- **Recording an allocation site on every object.** Covered above: permanent
  storage cost on all state for a debugging nicety.
- **Go's reflect API, or any part of it.** Out of scope on purpose. Type and
  field introspection over values a realm was handed is a capability question of
  its own; this package answers "which object is this, and where does it live".
