# chain/reflect: objects become addressable, and the address is an account

## Context

A realm that creates something and needs a unique name for it has, otherwise,
only names it chooses itself: a package path, a symbol, a counter it keeps, or
an ID its caller passes in. Anything a realm chooses, a realm can choose twice.

`grc20.NewToken` is the case that surfaced it. It builds `Token.id` from
`<pkg-path>.<symbol>.<caller-supplied-seqid>`, with the sequence number coming
from the calling realm, so one realm can issue two independent tokens under one
identifier and their `Transfer`/`Approval` events become indistinguishable.
That is #6026.

The VM already maintains the one name a realm cannot choose. `ObjectID` is
`{PkgID, NewTime}`: the realm an object belongs to, and a tick of that realm's
storage clock, minted by the VM while persisting the object.

Three earlier attempts:

- #6101 added `runtime.NewRealmID()`, returning `<realm-path>:<realm-time>`.
  Rejected in review: a second identity scheme no other VM concept consumes.
- #6139 derives a bech32 address from the ObjectID and exposes it as
  `chain/runtime.ObjectAddress(v any) string`, together with a breaking `grc20`
  and `grc20reg` redesign, across 50 files.
- #6214 adds `chain/reflect` and gives every object an address, in 17 files. It
  is the direct ancestor of this document, which keeps its API unchanged and
  answers the four questions it left open.

### What #6214 left open

Its own package doc says an object address "receives and does not release",
because a banker's sender is pinned to the calling realm's package address. It
says nothing about what the address means once the object is gone. It says the
set of values that get an address "is meant to grow, and to only ever grow",
without saying what admits a new member. And it leaves `grc20` untouched,
calling the #6026 fix "a small follow-up".

Morgan Bazalgette, reviewing the proposal, asked for "either a simpler solution
or a more complete idea of what it means to give addresses to objects". This is
the second. Four sections below are the four answers; everything else is #6214,
restated because it is still the decision.

## Decision

Add `chain/reflect`, an injected stdlib whose subject is object identity, give
every object an address, and make that address an account the owning realm can
spend from.

```gno
// chain/reflect
type Addressable interface{ Address() address }

type Object struct { /* unexported */ }

func Of(v any) (Object, bool)
func HasIdentity(v any) bool

func (o Object) Address() address // g1…, derived from the identity
func (o Object) ID() string       // "<realm-id>:<clock-tick>"
func (o Object) PkgPath() string  // the realm whose clock minted the ID
func (o Object) String() string   // "gno.land/r/demo/foo#7 (g1…)"
func (o Object) IsZero() bool

// chain/banker
const BankerTypeObjectSend BankerType = …

func NewObjectBanker(obj reflect.Object, rlm realm) Banker
```

### 1. The address is the deliverable

An object address is an ordinary gno.land address, derived by hashing
`objectid:<pkgid>:<newtime>`, the same shape as a package address's
`pkgPath:<path>` and in a disjoint preimage space.

`<pkgid>` is spelled by `PkgID.String()`, which is `RID` followed by uppercase
hex, and `<newtime>` in decimal. That is deliberately not the spelling of
`Object.ID()` (`ObjectID.MarshalAmino`: lowercase hex, no prefix), so hashing
`"objectid:" + obj.ID()` gives a different address. Both spellings are fixed by
#6139, whose preimage this one copies, and `DeriveObjectCryptoAddr`'s godoc
spells the preimage out byte for byte.

Being an address, not a bespoke identifier, is what makes an object usable by
code that has never heard of objects. `grc20`'s `Transfer(_ int, rlm realm, to
address, amount int64)` already takes a plain `address`, so an object can hold
token balances the day this lands, with no change to grc20. The banker's
recipient is likewise an unconstrained bech32, so an object can receive ugnot.

An earlier draft of this work returned the raw ObjectID and deliberately
withheld address derivation, on the grounds that the preimage freezes the first
time an address reaches an event and should not be fixed before the accounts
capability exists. That argument was rejected, and correctly: the capability is
the point, not a later bonus. Section 2 is that capability.

### 2. Spending: who may move an object's money

> *This is the question #6214 names and defers. It is answered here.*

`banker.NewObjectBanker(obj, cur)` returns a `Banker` whose `SendCoins` moves
value out of `obj`'s address instead of the realm's own.

**The authority rule is: an object address is spendable by the realm whose
storage clock minted the object's identity, and by nothing else.**

That rule is not a policy this design picked out of several. It is the only one
the VM can state without inventing new state:

- An ObjectID is `{PkgID, NewTime}`, and `PkgID` is
  `PkgIDFromPkgPath(<realm path>)`. The owner is therefore *inside* the
  identity, not recorded beside it.
- A live `realm` value proves which realm is executing, and `IsCurrent()` is
  already what `NewBanker` trusts for the realm's own address.
- So "is this object mine" is `oid.PkgID == PkgIDFromPkgPath(rlm.PkgPath())`:
  an equality between two values the VM computed, which no realm supplies and
  which cannot drift.

Every alternative is realm state. An owner field, a transferable capability
token, a registry of delegates: each is a thing a realm writes down, which a
realm can write down wrongly, and which an address created before it and
outliving it would then contradict. An address that exists for as long as a
realm's storage does needs an authority rule with the same lifetime.

**What the banker can do, and what it deliberately cannot.**

| operation | object banker | why |
|---|---|---|
| `SendCoins(objAddr, to, amt)` | yes | the capability this exists for |
| `SendCoins(anythingElse, …)` | panics | `pkgAddr == from` pin, unchanged from `NewBanker` |
| `GetCoins` / `GetCoin` | yes | reads are unrestricted on every banker |
| `IssueCoin` / `RemoveCoin` | panics | a denom is namespaced by a realm path; an object has none, and minting under the realm's own name is already `NewBanker(BankerTypeRealmIssue, cur)` |
| an origin-send variant | does not exist | the send-envelope of a message is credited to the realm's package address, never to an object's, so there would be nothing to spend |
| a sub-realm (`cur.Sub`) constructing one | panics | a sub identity is a different address synthesised from `host#subpath`; its PkgID is not the host's and it has no claim on the host's objects |

**Implementation.** `BankerTypeObjectSend` is a new `BankerType`, appended after
`BankerTypeRealmIssue` so no existing value moves, and `NewBanker` refuses it:
the authority check lives in exactly one place. The `banker` struct is
unchanged, so no persisted banker changes shape. `X_bankerSendCoins` handles
the new type on the same arm as `btRealmSend`, which is correct because the
trust chain is identical: `from` is pinned by `banker.gno` to the address bound
at construction, and that address came from the VM in both cases.

**What the Gno side is trusted with: nothing.** `NewObjectBanker` takes a
`reflect.Object`, but the only thing it reads out of it is `obj.ID()`, a string,
which the native `objectBankerAddress` then re-parses into an `ObjectID`,
re-renders and compares (so a short or malformed hashlet, which `hex.Decode`
would silently zero-extend, is rejected), checks `IsFinalized` on, compares the
PkgID of against the caller's realm path, and re-derives the address from. A
realm that fabricates an `Object` gains nothing: an ID naming another realm's
object fails the comparison, and an ID naming an unminted tick of the realm's
*own* namespace yields an address in a space that realm already controls in
full and could mint on demand. There is no address reachable by forgery that
honest use does not also reach.

**Consequence for the security model.** Holding an object banker is an
irrevocable grant over that one address, exactly as holding a realm banker is
over the realm address. The SECURITY note at the top of `chain/banker` applies
unchanged, and the blast radius is strictly smaller: one object instead of
every coin the realm holds. A realm that wants to delegate spending over a fund
now has a way to delegate over *that fund alone*.

Pinned by `gnovm/tests/files/zrealm_reflect_object_spend.gno` (spends, scope,
wind-down) and `zrealm_reflect_object_spend_foreign.gno` (a realm holding
another realm's object can read its address and nothing else).

### 3. Lifecycle: what the address means when the object is gone

> *#6214 says nothing about this. An address that can hold a balance cannot
> simply stop existing, so here is what happens in each case.*

**An address is never reassigned.** `NewTime` comes from `Realm.Time`, which
`assignNewObjectID` only ever increments (`targetRlm.Time++`) and which no
deletion path decrements. A tick, once spent, is spent forever, so no second
object is ever minted onto a first object's address. This is the load-bearing
invariant of the whole scheme and it is why "the balance survives" is safe
rather than alarming.

**Deletion does not delete the address.** The VM deletes the object; the bank
keeper knows nothing about objects and holds the balance at the address
regardless. Nor could the VM intervene: it cannot see balances, so it cannot
refuse to delete a funded object, and making it able to would couple the
GnoVM's ownership walk to the bank module.

That makes the balance reachable only if the realm can still name the address,
which is why `NewObjectBanker` takes a `reflect.Object` and not the object. An
`Object` is an immutable snapshot of three strings; it persists like any value,
it keeps nothing alive, and it is everything the authority check needs. So:

> **A realm that sends value to an object's address and keeps no
> `reflect.Object` for it has stranded that value permanently.** Persisting the
> `Object` alongside, or instead of, the object is the realm's obligation.
> There is no recovery path and there should not be one: a recovery path is a
> way for someone other than the minting realm to reach the address.

**Unreachable is the same as deleted.** The VM draws no distinction, and
neither does this. An object that is no longer referenced is collected at
finalization; what decides whether its balance is recoverable is whether the
realm kept the snapshot, not how the object died.

**In-place upgrade keeps everything.** A redeploy to the same path takes over
the realm persisted there: `VMKeeper.AddPackage` reads `GetPackageRealm(pkgPath)`
and hands it to `RunMemPackageOverRealm`, precisely so "a redeployment's
ObjectIDs [stay] clear of the objects the path already holds". Same path, same
`PkgIDFromPkgPath`, same clock, so every address the old code minted is still
unique, still funded, and still spendable by the new code. Upgrading a realm
costs its objects nothing.

**A versioned upgrade is a migration.** `gno.land/r/foo/v2` is a different path,
so a different PkgID, so a disjoint address space. Addresses minted by `v1` stay
`v1`'s, and only `v1`'s code can ever spend them. A realm that versions itself
must drain its object addresses from `v1`, or keep a `v1` entry point that `v2`
calls. Stated here because it is the one lifecycle case that needs the author to
do something, and nothing in the VM will remind them.

Pinned across real transactions by
`gno.land/pkg/integration/testdata/object_address_lifecycle.txtar`: 700000ugnot
sent to an object's address, the object dropped from realm state, the balance
still there, a new object minted onto a *different* tick (`:18` → `:22`), and
the dead object's balance drained to a user through the stored snapshot.

### 4. The set that has an address, and the rule for growing it

> *#6214 says the set "is meant to grow, and to only ever grow", which says
> when a change is allowed, not which change. Here is the admission test.*

Today the set is: a pointer to a standalone object, a func value, a bound
method value, and a map, in each case persisted by a realm.

**A kind of value joins the set when all four of these hold.** They are
mechanical; none of them is a taste call.

1. **One object, one answer.** Every route to the value resolves to the same VM
   object. If two spellings of the same thing can produce two Objects, or two
   different things can produce one, the kind fails.
2. **Reference, not copy.** `TypedValue.Copy` shares the underlying object
   rather than cloning it, so every holder agrees on the identity.
3. **Minted by a realm clock.** The value's `ObjectID` is finalized with
   `PkgID.IsRealmPkg()`. No clock, no identity, and no address to derive.
4. **Monotone.** Admitting the kind changes no address that is already
   derivable, and no address it would hand out is one that any other kind
   already does.

Rule 4 is why this only grows: an address in a balance cannot be withdrawn, so a
kind is admitted only when the answer for it has been settled, and never
"provisionally". A value whose behaviour under copying and aliasing is unsettled
fails rule 1 or 2 and stays out, which costs nothing, because admitting it later
is a compatible change.

| kind | 1 one object | 2 reference | 3 realm clock | 4 monotone | in |
|---|---|---|---|---|---|
| `*T` to a standalone heap item | yes | yes | yes | yes | **yes** |
| func value | yes | yes | yes | yes | **yes** |
| bound method value | yes | yes | yes | yes | **yes** |
| map | yes | yes | yes | yes | **yes** |
| `&x.Field`, `&arr[i]`, `&pkgVar` | **no**, resolves to the container, so every sibling answers one address | yes | yes | yes | no |
| slice | **no**, resolves to its backing array, shared with every other view | yes | yes | yes | no |
| struct or array by value | n/a | **no**, `TypedValue.Copy` clones it, so a value may be the object or a copy depending on how it arrived | yes | yes | no |
| anything in a stdlib or `p/` package | yes | yes | **no**, `!PkgID.IsRealmPkg()`; `GetRealmByID` would name `strings` as the creating realm of `strings.ToUpper` | yes | no |
| scalar, nil pointer, nil map, nil interface | n/a | n/a | **no**, there is no object | yes | no |

Two of these have a visible route in: a *container* pointer would pass rule 1 if
the VM ever gave struct fields and array elements their own heap items, and
immutable-package objects would pass rule 3 if a decision were taken that they
should be addressable. Both are then ordinary admissions under rules 1 to 4, not
judgement calls, which is the point of writing the rule down.

### 5. Object is a type, not a string

`Object` carries unexported fields, so `Of` is the only thing that returns a
non-zero one. The whole point is an identifier the realm does not choose, so
making it unforgeable by construction rather than by convention is the property,
not a detail. It stays comparable and map-key safe.

#6139 returns a bare `string`. A realm can therefore fabricate an "object
address", and a consumer taking a `string` cannot tell a VM-issued identity from
one the realm typed out, which is the same class of problem as #6026 itself.

The unforgeability is a property of the *type*, and section 2 does not lean on
it: the native re-derives everything from the ID, so an authority decision never
rests on "this struct came from `Of`". That is deliberate. A type-level
guarantee is the right default for the ninety per cent of uses that only pass an
address around; it is the wrong thing to make a spending check depend on.

`Addressable` is the deliberate exception, and its doc says so: any type can
implement `Address()` and return whatever it likes. It exists to let an API
accept a user, a realm or an object without caring which. It is safe for
deciding where value goes and unsafe for deciding who may move value; authority
comes from the VM, never from satisfying an interface.

### 6. Provenance travels with the address

`PkgPath()` answers "whose object is this", resolved through
`Store.GetRealmByID(PkgID)`. Together with the address it is enough to render a
deep link to an object or its funds, which is the case that motivated it: a
governance proposal whose fund address and executable body are both separately
addressable and separately linkable.

It is the realm whose clock minted the ID, which is usually but not always the
realm whose code allocated the object: `Realm.assignNewObjectID` re-stamps an
object allocated by a `maketx run` script, or inside a stdlib function, to the
realm that persists it. It is also the realm section 2 checks against, so
`PkgPath()` and "who may spend this" are the same question with one answer.

**No `Type()`.** An earlier revision carried the pointee's declared type name in
`Object`. Review showed that broke the one promise `Object` makes, that two
Objects are equal exactly when they name one object: `Of(p)` and
`Of((*Alias)(p))` share an ID and an address but differed in type, so a map
keyed by `Object` counted one object once per pointer type, and a realm's `type
Mine strings.Builder` made a `strings.Builder` report as the realm's own `Mine`.
There is no canonical type to report instead. The type at hand is the caller's
route to the object, and the heap item's stored `Value.T` is not stable either:
an assignment through a converted pointer replaces it wholesale
(`TypedValue.Assign` copies `tv2`, `T` included). So the type is out, which is
also what keeps the native's output fixed-size for gas. Adding a canonical type
later, if the VM ever records one, is compatible under rule 4; shipping a wrong
one is not.

**Source position is not available and is not added.** `DeclaredType.ParentLoc`
is blank for every package-level type (`declareWith` in `types.go` keeps it blank
for `*PackageNode`/`*FileNode`), and `ObjectInfo` carries no allocation site.
Recording one would add a field to every persisted object in every realm: a
storage cost on all state, permanently, for a debugging nicety. Package path plus
address is enough to link to a thing.

### 7. One outcome per case, no panics

| v | `HasIdentity` | `Of` ok |
|---|---|---|
| pointer to a standalone object, persisted | true | true |
| pointer to a standalone object, created by the running call | true | false |
| func value, bound method or map, persisted | true | true |
| any of the above, persisted by a stdlib or `p/` package | false | false |
| pointer into a struct field or array element | false | false |
| slice, struct or array by value | false | false |
| scalar, nil pointer, nil interface | false | false |

`Of` never panics and never invents an address. `HasIdentity` splits its two
false cases into "ask again after the realm boundary" and "nothing to wait for".

`NewObjectBanker` is the opposite and panics on every refusal, because its
caller is about to move money and a `false` there is a branch somebody forgets
to take.

#6139 has one function with three behaviours: `""` for an unpersisted object,
`""` for a pointer into a container or a slice, and a panic for a scalar. The
two that need different handling are the two that are indistinguishable.

### 8. The stamping window is reported, not papered over

`PkgID` is stamped by the allocator at creation; `NewTime` only by
`Realm.assignNewObjectID`, reached from `FinalizeRealmTransaction`, which runs
when a realm frame returns. So an object created by the running call has no
complete identity and no address, and no primitive can conjure one: attaching to
the ownership tree *is* what mints it. Stamping early is a lifecycle change,
since `GetIsReal()` *is* `ID.IsFinalized()` and `incRefCreatedDescendants` guards
on the same predicate, so an early-stamped object looks real while unsaved and
gets skipped during finalization. #6139's analysis of this is correct and is
reused here.

What is a choice is how the window is reported. `Of` returns `ok == false`
rather than an empty address. With an address the stakes are higher than with an
opaque ID: an empty address written into an event is permanent, and a caller who
treats it as real directs funds at a string that is not an account.

Section 9 is what that window costs.

### 9. #6026 and grc20: why the fix is still not in this PR

> *#6214 calls it "a small follow-up". It is not, and the reason is section 8.*

The fix everyone has in mind is: drop `NewToken`'s caller-supplied `seqid.ID`,
and let the token's identity be its object address. Concretely
`Token.ID()` would resolve through `reflect.Of(tok)`.

**It cannot resolve in the transaction that creates the token.** `NewToken`
returns a `*Token` allocated by the running call, so it is inside the stamping
window and `reflect.Of` returns `ok == false` until the realm frame returns.

**Every GRC20 in the tree creates and uses its token in one transaction.**
`r/demo/defi/foo20` is the shape:

```gno
func init(cur realm) {
	Token, privateLedger = grc20.NewToken("Foo", "FOO", 4, 0, cur)
	userTeller = privateLedger.CallerTeller()
	privateLedger.Mint(Ownable.Owner(), 1_000_000*10_000)
	grc20reg.Register(cross(cur), Token, "")
}
```

`Mint` emits an event carrying `led.token.ID()`, and `grc20reg.Register` keys on
it. Both run before the frame returns. An address-backed `ID()` panics there, at
deploy time, for `foo20`, `bar20`, `wugnot`, `grc20factory`, `test20`, the
`treasury` filetests and every realm that follows the pattern. Not a migration:
a redeploy of every token on the chain, with a changed authoring rule ("you may
not mint in the transaction that creates the token") that nothing in the API
expresses.

The alternatives, all worse in this PR:

- **Lazily resolve and cache in `ID()`.** The resolution is a write, and `ID()`
  is reached from `Render` and from read-only tellers. A cache that can only be
  filled from a mutating path is a second window, and the token's id differs
  between its creation transaction and every later one, which breaks indexers
  more thoroughly than #6026 does.
- **A two-phase `NewToken` + `Bind(cur)`.** Correct, and a real API burden on
  every token author, imposed by a PR whose subject is the VM.
- **Emit no id at creation.** Gives up `NewToken`'s documented guarantee that
  token creation is fully observable, which is the property that lets an indexer
  notice a realm issuing two ledgers under one name.

**What actually unblocks it** is separating "has an identity" from "is
persisted" in the VM, so an object can be stamped at allocation. Today
`GetIsReal()` *is* `ID.IsFinalized()` and `incRefCreatedDescendants` guards on
the same predicate, so that means a third `ObjectInfo` state and an audit of
every `IsFinalized()` call site. That is its own change, with its own risk, and
it is the prerequisite the #6026 follow-up should be written against.

So `grc20` is untouched here, and this is the replacement for #6214's one-line
"small follow-up": the follow-up is specified above, its blocker is named, and
`zrealm_reflect_duplicate_symbol.gno` demonstrates the end state on a local
token type so that the target is not in doubt. GRC721 has the same exposure and
the same blocker.

## Consequences

- **`chain/banker` imports `chain/reflect`.** genstd recomputes `initOrder`
  from the import graph, which moves `chain/reflect` ahead of `chain/banker`;
  no manual ordering and no core change. This is the second stdlib-to-stdlib
  import after `chain/runtime/unsafe` → `chain/runtime`.

- **`BankerTypeObjectSend` is appended**, after `BankerTypeRealmIssue` and
  before `maxBanker`, so every existing `BankerType` value is unchanged and no
  persisted banker is reinterpreted. The `banker` struct gains no field for the
  same reason.

- No change to `misc/genstd`. The native's Go parameter is `gno.TypedValue`,
  which already matches any Gno parameter type and links through with no Go2Gno
  conversion, while the Gno side declares `any`, which the generated binding
  spells `gno.X("any")` and resolves through uverse (`GnoTypeExpression` emits
  `gno.AnyT()` only for a type spelled `interface{}`). Pinned by the existing
  `misc/genstd/testdata/linkFunctions_TypedValue`. #6139 instead widens
  `mapping.isTypedValue` so a *Go* `any` parameter is treated as a raw
  TypedValue, which is unnecessary and makes every future native declaring a Go
  `any` silently receive a TypedValue rather than a converted Go value.

- **`NewObjectBanker`'s realm argument comes last.** A crossing function's
  first `realm` parameter must be named `cur`; `NewObjectBanker(obj, rlm)`
  matches `NewBanker(bt, rlm)` and keeps the function non-crossing.

- The preimage layout is consensus state from the first address that reaches an
  event, a registry or a balance. The prefix, the separators, the `PkgID`
  encoding (whose flag nibble encodes IsStdlib/IsImmutable/IsInternal) and the
  `NewTime` semantics all become fixed, and changing any of them moves every
  object address in existence. The preimage is taken from #6139 unchanged so the
  two approaches stay interchangeable.

- The native resolves pointer bases itself and never calls
  `TypedValue.GetFirstObject`, which panics outright on a package `RefValue` and
  on a bare `*HeapItemValue`. Every shape is handled or reported as no identity.

- `PkgPath()` costs a realm lookup. It is a cache hit in the normal case: the ID
  is finalized, so the object is persisted, so its realm exists and is loaded for
  the object to have been reached at all. `objectBankerAddress` does not pay it:
  it compares PkgIDs, and `PkgIDFromPkgPath` is memoised.

- Stdlib `.gno` source is genesis state, so adding the package moves the genesis
  app hash (`expectedCrossrealm38Hash`) whether or not anything imports it.
  Storage-deposit txtars are unaffected.

- Gas: two flat rows, `chain/reflect.objectInfo` at 1200 and
  `chain/banker.objectBankerAddress` at 1100, both dominated by the
  sha256-and-bech32 derivation. Measured end to end through the dispatcher;
  drafts pending a re-fit on the reference machine, see each row's own comment.

- **Storage rent is unchanged and unaware of object addresses.** An object's
  balance lives in the bank keeper, its storage deposit in the realm. Funding an
  object does not pay for its storage, and deleting it refunds the storage
  without touching the balance. Nothing here changes that; it is noted because
  "the object has money" invites the opposite assumption.

## Alternatives considered

- **Leaving spending out, as #6214 does.** The package doc then has to tell
  realms that an object address "receives and does not release", which is a
  standing invitation to send value somewhere it cannot come back from, in a
  release that ships the address. Deferring the capability does not defer the
  risk, it just ships the risk first.
- **An owner field, or a transferable spending capability.** Covered in section
  2: realm state outlived by the address it governs.
- **Blocking deletion of a funded object.** The GnoVM cannot see balances, and
  giving its ownership walk a dependency on the bank module to gain a check
  that would fire at the worst possible moment is a poor trade. Section 3 puts
  the obligation on the realm instead, and says so in one sentence a realm
  author can act on.
- **Landing the #6026 grc20 fix here.** Section 9.
- **`chain/runtime.ObjectAddress`, as in #6139.** `chain/runtime` is `ChainID`,
  `ChainDomain`, `ChainHeight`, `AssertOriginCall`: information about the running
  chain. Which object a value is, and where it came from, is not that, and
  putting it there leaves no home for the next question of the same kind.
- **`vm/reflect` or `gno/reflect`.** Neither `vm/` nor `gno/` exists as a stdlib
  namespace. Checked while evaluating it: a new top-level namespace needs no core
  change, since `Re_gnoStdPkgPath` is just `name(/name)*` with no whitelist. It
  is a taste call, and `chain/` is where the non-Go-portable stdlibs already live.
- **A top-level `reflect`.** Collides with Go's `reflect` for anyone reading or
  transpiling the code, and promises an API this will not have.
- **Returning the ObjectID and withholding the address.** Covered above:
  rejected, because a unique string leaves every consumer needing a translation
  layer, and the address is what makes objects usable by existing code.
- **Recording an allocation site on every object.** Covered above: permanent
  storage cost on all state for a debugging nicety.
- **Go's reflect API, or any part of it.** Out of scope on purpose. Type and
  field introspection over values a realm was handed is a capability question of
  its own; this package answers "which object is this, where does it live, and
  who may spend what it holds".
