# Namespace authority as a team ACL, declared by `r/<name>/home`

## Status

Proposed, with a **draft implementation** in this PR so the shape can be argued
against running code rather than prose. Companion to #6317 rather than a
competitor to it: the two cover different halves of the same problem, and the
section "What this does not solve" says which half is whose.

The code is a draft in the literal sense. The tests are adversarial, but nothing
in it has run on a chain and the open questions at the end are still open.

## The problem

A namespace's deploy authority is one address, decided once, for good.

`r/sys/names` has exactly two authorization paths:

```go
// Path 1: PA (personal-address) namespace.
if address_XXX.String() == namespace {
    return true
}

// Path 2: registered-name namespace via r/sys/users.
if lookup != nil {
    if owner, ok := lookup(namespace); ok && owner == address_XXX {
        return true
    }
}
```

`lookup` is `resolveCurrentName`, which returns the single address
`r/sys/users.ResolveName` holds for that name. One name, one address, and no
realm anywhere can change which address that is: every mutator on the user
record is controller-gated, and `r/sys/namereg/v0` exposes only `Register` to
users. Renames and deletions are GovDAO proposals.

Four things fall out of that, and they are not edge cases:

- **Key rotation is impossible.** Moving to a new key means abandoning the
  namespace, because the name cannot follow.
- **A multisig cannot change its quorum.** Changing a 2-of-3 to a 4-of-7
  produces a **new address**, so the name would have to move. This has already
  been raised by at least two teams holding namespaces this way.
- **Shared ownership means a shared key.** A team publishing under one namespace
  today shares one private key, which is the practice every other part of the
  system is designed to avoid.
- **A lost key is a dead namespace**, permanently, along with everything that
  would ever have been published under it.

#6317 answers this by making the **name** mobile: an id and a rev on the user
record, an offer and accept transfer, a recovery vote, and a resolver in the
keeper. This ADR answers the first three by making the **authority set**
mutable, and leaving the name exactly where it is.

## The decision

### A namespace's ACL is written by the one realm that is already named for it

Add `r/sys/teams`, holding a map from namespace to a member set. Authority to
write the entry for `<name>` is not granted by a list, a controller or an
argument. It is **derived from the caller's own package path**:

```go
// Sketch, not final code.
func assertHomeCaller(_ int, rlm realm) string {
    if !rlm.IsCurrent() {
        panic("unauthorized: rlm is not the caller's live cur")
    }
    caller := rlm.Previous().PkgPath()
    if caller == "" {
        panic("unauthorized: no calling realm")
    }
    ns, ok := namespaceOfHome(caller) // "gno.land/r/<ns>/home" -> ns
    if !ok {
        panic("unauthorized: caller is not a /home realm: " + caller)
    }
    return ns
}
```

Nothing identifies the namespace except the caller itself, so there is no
whitelist to maintain and no controller address to compromise. This is a
deliberate improvement on the `r/sys/users` model, which this design is
otherwise shaped after: that realm's authority flows through
`controllers.Has(...)`, a set someone must curate and an attacker can target.

The shape is not new. Two realms already run it:

- `r/sys/params`, in `assertDelegate`: `IsCurrent`, then an empty check, then an
  exact `rlm.Previous().PkgPath()` comparison.
- `r/sys/validators/v0`, in `assertValopersCaller`, with the same three steps and
  the same reasoning recorded beside it.

Both record why each step is load bearing, and both should be read before this
is implemented rather than re-derived:

1. **`IsCurrent()` first.** A caller threading a stashed or sibling-frame realm
   value otherwise forges the designation. `docs/resources/gno-security.md`
   names the class.
2. **An empty path denies, explicitly, before anything parses it.** A direct
   user `MsgCall` presents the empty package path. `r/sys/params` calls this
   check "load-bearing, not defensive" and points at `r/gov/dao`'s allowlist as
   the mirror-image bug. The general warning is that a design which *derives*
   the namespace from the caller must refuse the empty path rather than parse it
   into an empty namespace.

   **In the code here that check is not load bearing, and an earlier draft of
   this ADR claimed it was.** Deriving the prefix from the chain domain means
   the empty path already fails the prefix test, so deleting the empty check
   changes the error message and nothing else. It stays because a named refusal
   beats a generic one and because a refactor could make it load bearing again.
   The claim did not survive being tested, so it is corrected rather than
   quietly dropped.
3. **Exact matching, never a prefix.** `cur.Sub(subpath)` presents the
   synthesized identity `host#subpath`, so `gno.land/r/<ns>/home#anything` is a
   genuinely presentable caller. Whether sub-realms of `home` inherit the
   namespace is a decision to take, and the proposed default is no.

The accepted caller grammar is therefore the chain domain, `/r/`, exactly one
path segment, then `/home`, with no `#`. That grammar is a security boundary and
not a naming convention: `maketx run` realms present `<domain>/e/<addr>/run`, so
a looser "is this code" test would admit them.

The prefix is **derived from `runtime.ChainDomain()` rather than hardcoded**. A
hardcoded `gno.land/` makes the realm refuse every caller on a chain configured
with any other domain, silently, which reads as a dead realm rather than as a
misconfiguration.

`r/sys/teams` must also **never call back into `r/<name>/home`**. Membership is
pushed in as arguments. Beyond the fact that gno has no dynamic dispatch by
package path, `r/sys/validators/v0` documents the pull direction as a
confused-deputy class it deliberately eliminated.

### `r/sys/names` grows a third path

Next to PA and registered-name, a third check asks `r/sys/teams` whether the
deploying address is a member of that namespace's team. The three are OR'd, so a
registered name keeps its existing authority after adopting a team. A namespace
may additionally **retire** the registered-name path for itself, which is what
makes the original key inert once a team is in place.

A solo user is then a team of one, and the personal case stops being a separate
mechanism to maintain.

### Two ways for `home` to update its own ACL

`init()`, which runs at deploy time, and exported helper functions, which can be
called afterwards. Both reach `r/sys/teams` through the same gate.

The two are not equivalent, and the difference decides whether a namespace
survives its first key retirement:

- `init()` re-runs on **every** redeploy, over the prior realm's state
  (`RunMemPackageOverRealm(memPkg, true, priorRealm)`). An `init()` that writes
  the ACL unconditionally turns a routine redeploy into an undeclared team
  reset. It should **seed only**: write when the namespace's slot is empty, and
  leave it alone otherwise.
- Only the stamped original creator may ever redeploy a private realm, so once
  the first key is retired the `init()` route is gone for good. **The helpers
  must exist and work before that point, not after.**

### `r/sys/teams` carries no version suffix, deliberately

This is the one decision here that cannot be walked back later, so it belongs in
the record rather than in a comment.

A namespace that has retired its registered-name path can no longer redeploy its
own `r/<ns>/home`, because only a private realm's stamped original creator may
replace it. If this realm were versioned and a `v1` appeared, every such
namespace would be pinned to `v0` for good: it could not redeploy `home` to
point at `v1`, and nothing else may write its ACL. **A design whose purpose is
removing that kind of permanence must not reintroduce it one layer up.**

So `r/sys/teams` is a state holder and nothing else. Policy lives in the
`r/sys/names` successor that reads it, and that realm is swappable, because the
keeper reaches it through `vm:p:sysnames_pkgpath`. Upgrades happen there.
Anything added to `r/sys/teams` has to keep that division, or the pin returns.

### Rollout: one parameter proposal, no binary change

For packages that do not exist yet, this is deployable on the current chain:

1. Deploy `r/sys/teams` and an `r/sys/names` successor carrying the third path.
2. Pass one GovDAO proposal setting `vm:p:sysnames_pkgpath` to the successor.

`checkNamespacePermission` already hands the realm `(creator, namespace)`, which
is everything a membership check needs, and `sysnames_pkgpath` is not among the
key families reserved from the generic parameter factory (only `node:valset:*`
and `vm:p:run_submitters` are). Nothing in step 1 or 2 touches the binary.

A worked migration, with no name moving and nothing to coordinate:

```
- ORIGOWNER publishes r/<name>/home, whose ACL names the new multisig
- ORIGOWNER transfers its coins to the new multisig
- ORIGOWNER retires the registered-name path for <name> in r/sys/teams
```

`ORIGOWNER` is then inert for that namespace. That is safe precisely because of
the constraint described at the top: a retired key cannot rename or delete its
way back in, because it never could. The registered-name deploy path was its
only residual power, and step 3 removes it.

## What this does not solve, and why #6317 is still needed

Two cases are out of reach of any realm-only design, and pretending otherwise
would be the easiest way to get this wrong.

**Redeploying an existing private realm.** `checkRedeployPermission` refuses
unless the signer is byte-for-byte the address stamped as the original creator.
That gate is separate from the namespace check and no realm can reach it, so
under this design a private realm stays owned by whichever member deployed it,
and `home` itself becomes immutable once its deployer retires. It is also not an
oversight to route around: it was decided deliberately in #6171, recorded in
`gno.land/adr/pr6171_private_redeploy_creator_binding.md`, to close a takeover.
Changing it is a conscious reversal, it needs the binary, and it is exactly what
#6317's `effectiveOwner` resolver proposes. **This is the one place the two
designs must agree on a single answer.**

**A key already lost, exploited, or whose holder has died.** Every step of the
migration above requires the original key to still sign. This design is
*prospective*. The case most often raised against the current model is
*retrospective*, and only a recovery vote of the kind #6317 specifies reaches
it.

So the split, stated plainly:

| Case | This ADR | #6317 |
|---|---|---|
| Rotate a key you still hold | add a member, remove a member | offer and accept transfer |
| Change a multisig's quorum | a call | a new address, then a transfer |
| Publish as a team | the native case | one shared key, or a multisig |
| Key lost, exploited, holder dead | nothing | recovery vote |
| Redeploy a private realm under new authority | blocked by the creator stamp | the `effectiveOwner` resolver |

## Alternatives considered

**Name transfer alone (#6317).** It solves rotation by moving the record, which
is why it reaches into the keeper, the wire format and a replay upgrade, and it
accepts that a stolen key can move a name away and lock its owner out. It does
not make shared ownership any cheaper: a team still shares one key or one
multisig address. Complementary, as above.

**Hold the name with a multisig, as teams do today.** This already works and is
already deployed. It fails on the one operation teams actually perform, which is
changing who is in the group: a different quorum is a different address, so the
name has to move, which is the problem being solved.

**Per-path ACLs inside the realm** (`r/<ns>/x/*`, create versus redeploy).
Impossible today, and worth stating so nobody designs for it: the keeper passes
the namespace only, never the full package path and never a redeploy bit. That
needs a keeper hook, and it should ship separately from the membership check
rather than holding it up.

**A controller whitelist, as `r/sys/users` uses.** Rejected: it reintroduces a
curated list and a compromise target, for no capability a pkgpath-derived gate
does not already have.

**Burning a key globally, at the auth layer.** Rejected for this design. A
namespace-scoped retirement is realm-only and is all the migration needs;
refusing a public key chain-wide is a tm2 change that buys nothing here and
costs a hardfork.

## Consequences

- **`home` becomes the ACL root, and `r/sys/teams` trusts it completely.** A
  weak gate on one exported helper in one `home` is a namespace takeover, and
  nothing in `r/sys/teams` can detect it. This argues for a standard, audited
  template that most namespaces import rather than hand-roll.
- An `init()` that is not seed-only makes every redeploy an undeclared ACL
  reset.
- Retiring the registered-name path is one-way under this design. If the ACL is
  wrong at that moment, nothing recovers it except governance.
- Private realms under a shared namespace remain personally owned until the
  creator-stamp question above is settled.
- Teams and the existing paths are OR'd. That must be written in the verifier
  rather than emerge from the order of two `if` statements.
- A namespace that never publishes a `home` realm is entirely unaffected.

## Validation

The draft in this PR has tests, and each gate below was **broken on purpose and
the suite re-run**, because a passing test is not evidence until it has been
seen to fail. Two of the first four gates turned out to be uncovered, which is
the reason this section reports the method rather than a count.

Covered and break-verified in `r/sys/teams`:

- A `home` realm writes its own namespace and no other.
- A direct user `MsgCall` is refused, with the empty path asserted explicitly.
- A non-`home` realm is refused, and the package-path parser refuses a nested
  `home`, a pure package, a `run` realm, a foreign domain and the `host#subpath`
  sub-realm identity. The last one was **not** covered until breaking the guard
  showed the suite still green.
- A stashed, non-live realm value is refused. This one needs
  `testing.MakeRealm` to build a synthetic realm whose `Previous()` names an
  authorized `home`; the exported crossing functions cannot reach it, since the
  compiler already guarantees their `cur` is live.
- Seeding is idempotent, so a redeploy re-running `init()` does not reset a team
  the helpers have since changed.
- Retiring with an empty team is refused, and so is emptying or removing the
  last member of a retired namespace: the combination is a namespace nobody can
  ever deploy under again.
- The member cap is enforced on both write paths, and re-adding an existing
  member at the cap still passes, because that is not growth.

Covered and break-verified in `r/sys/names`:

- A team member deploys; a non-member does not; a team covers only its own
  namespace.
- Paths 2 and 3 are OR'd before retirement, which is what lets a migration
  publish `home` while the original key still works.
- After retirement the registered-name holder is refused and the team is not.
- Retirement does not touch personal-address namespaces, since the PA check sits
  above both.
- Pause and the pre-`Enable` bypass still outrank teams in both directions.

Not covered, and worth saying so: nothing here has run against a real keeper, so
the interaction with `checkNamespacePermission` is argued rather than
demonstrated. The `r/<ns>/home` side has no template yet, which matters because
`home` is the ACL root under this design.

## AI assistance

Written with AI assistance (Claude Code). The behaviour of `r/sys/names`,
`r/sys/users`, `r/sys/namereg/v0`, `r/sys/params`, `r/sys/validators/v0` and the
VM keeper paths cited here was read from the tree rather than recalled, and the
two limits in "What this does not solve" were found that way rather than
assumed. The human author reviewed and owns the proposal.
