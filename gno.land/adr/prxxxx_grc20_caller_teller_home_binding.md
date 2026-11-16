# ADR: confine `grc20`'s frame-relative teller to the token's own realm

## Status

Proposed. Third and — on the evidence below — final shape. Supersedes both the
strict construction pin and the origin guard; see *Alternatives* for why each
leaves a hole.

## Context

`CallerTeller` resolves its debited account **late**, at each write, as
`rlm.Previous()` — the frame that crossed into whichever realm invokes the
teller. Inside the token's own realm that is correct: `wugnot.Transfer` debits
the user who knowingly called wugnot. Anywhere else it is a confused deputy,
because the account debited is chosen by whoever the teller holder can get to
call in.

There are two distinct ways a foreign realm can end up holding one:

1. **Minting.** The `*grc20.Token` pointer is published on four public surfaces
   (`wugnot.Token`, `foo20.Token`, `grc20factory.Bank`, `grc20reg.Get`), so any
   realm can call the accessor on it.
2. **Travel.** A realm that legitimately owns a token can build a teller and
   then export the resulting *value*, as `bar20` used to do with `UserTeller`
   (now unexported, behind narrow wrappers).

Any fix that addresses only one of these leaves the other open. That is not
hypothetical: it is what the two earlier attempts each got wrong, in opposite
directions.

## Decision

Close both doors.

- **Construction.** The accessor moves from `*Token` to `*PrivateLedger`:
  `func (ledger *PrivateLedger) CallerTeller() Teller`. `NewToken` hands the
  ledger to the creating realm and nowhere else, and no in-tree realm exports
  it, so a foreign realm cannot mint a frame-relative teller at all.
- **Travel.** Every write checks the invoking realm:

  ```go
  host, _, _ := chain.SplitPkgSubPath(rlm.PkgPath())
  if host != ft.Token.origRealm { return ErrForeignCallerTeller }
  ```

  A teller that escapes as a value is inert everywhere but home. The host is
  compared after stripping any `":subpath"` synthesized by `realm.Sub`, so the
  token's own sub-realms are not falsely rejected.

`Token.origRealm` is captured in `NewToken` behind the existing
`rlm.IsCurrent()` assertion, so it is unforgeable.

**The check is on the invoking realm's path alone, deliberately not on whether
the resolved actor is an end user.** That distinction is the whole point of
this ADR and is argued in *Consequences*.

## Consequences

**Nothing is left open.** Concretely, against real reproductions:

| Attack | Actor resolves to | Closed by |
|---|---|---|
| Foreign realm mints a teller off the published `*Token` and drains its direct caller | signing user | construction — does not compile |
| Same, via one signed `maketx run` | `/e/<addr>/run` frame, whose `Address()` is the signer's | construction |
| Realm uses an *exported* teller value against its direct caller | signing user | travel check |
| Realm reached from an honest hub uses an exported teller's `TransferFrom` | **a realm** — the hub | travel check |

The last row is why an actor-shaped predicate cannot be the answer. `Transfer`
forces the source to the frame's actor, but `TransferFrom` resolves the frame
to the **spender** while the debited `owner` is a free parameter. A realm
reached from an honest hub therefore inherits *that hub's* allowance against
any owner who granted one — so approving a router exports your allowance to
anything the router can be induced to call. Reproduced: alice approves `hubx`
and only `hubx`, one ordinary `maketx call` into `hubx` crosses into a plugin,
and the funds land at the plugin, whose own allowance is zero throughout.

That is the same defect as the original, one level up: frame-relative
resolution means whoever you call can act as you. It cannot be patched by
asking *who* is being debited; it is closed only by ensuring the frame-relative
teller never runs outside the realm that owns the ledger.

**Cost: the generic-hub pattern is withdrawn.** A non-owner realm can no longer
charge its caller through an arbitrary token. This is the real price, but it is
smaller than the figure this decision was weighed against. A "263 call sites"
estimate was inherited from an earlier assessment and repeated here and in the
sibling proposals without verification. Measured file by file against the
largest downstream consumer:

| | count | nature |
|---|---|---|
| `.CallerTeller()` call sites | 17 | 16 on realms that own their token — the call moves to the ledger and keeps working; 1 on a generic `common` helper |
| `common.*` token movement, product code | **31**, in 18 files | the real migration: per-module `RealmTeller` for module-owned funds, `Approve` + `RealmTeller().TransferFrom` for user funds |
| same, `contract/r/scenario/**` filetests | 139 | test surface despite not matching `*_test.gno` |
| same, `_test.gno` | 75 | test surface |

The hub cannot be patched in place: substituting `RealmTeller(0, cur)` inside
the helper freezes the actor to the helper realm's own address, so it would move
its own funds. There is no safe generic pass-through, because "act as my caller"
is precisely the vulnerability.

**In-tree, the same shape had to be removed rather than migrated.**
`grc20reg` had grown `Transfer`/`Approve`/`TransferFrom` wrappers over
`MustGet(key).CallerTeller()`. A registry holds only `*Token` pointers, so it
cannot reach the ledger, and `RealmTeller` would debit the registry itself —
the wrappers are unimplementable by construction, and were deleted. Callers use
the token realm's own entry points instead.

**The capability is not lost, only renamed.** The supported route is `Approve`
+ `RealmTeller().TransferFrom` — eagerly bound to the spending realm's own
address, allowance-gated, and immune to the confused deputy because nothing is
resolved from the frame. Demonstrated in the regression test (a hub spending an
allowance granted to it) and applied to the two in-tree call sites that needed
it (`grc20_registry_emit`'s hub fixture, and `eventix.BuyTicket`).

**Migration is mechanical but not silent.** `Token.CallerTeller()` no longer
exists, so every foreign call site is a compile error rather than a runtime
panic. For a security-relevant change that is the safer failure mode: no
consumer can ship a build that fails only in production.

## Alternatives

**Strict construction pin.** `CallerTeller(_ int, rlm realm)` on `*Token`,
asserting `rlm.IsCurrent() && rlm.PkgPath() == origRealm` at construction.
Blocks minting; does **not** block travel — verified, a teller exported by a
realm that built it legally still drains a caller. Also a signature change on
the accessor itself, which this ADR avoids.

**Origin guard.** Keep the accessor on `*Token`, and refuse only when a
non-owner realm would debit an actor that `IsUser()`. Blocks minting-for-drain
and travel-for-drain against a *user*; leaves realm-to-realm debits and the
`TransferFrom` allowance hop above. Its appeal was compatibility — no signature
change, hubs keep working — which is exactly the property that keeps the hole
open, since a working hub and a hostile plugin are indistinguishable at the
frame.

An earlier iteration of that guard used `IsUserCall()`, which is `pkgPath == ""`
and therefore false for `maketx run` even though the run frame's address is the
signer's; one signed run drained the signer in full. `IsUser()` fixes that
particular hole but not the two structural ones.

## Tests

`gno.land/pkg/integration/testdata/grc20_callerteller_home.txtar` covers, in
one scenario: the exported-teller drain over `maketx call`, the same over
`maketx run`, the `TransferFrom` allowance hop through an honest hub, that the
token's own realm still moves its users' funds, and that a foreign realm can
still spend an allowance granted to it via `RealmTeller`. Each attacker
surfaces the error, so the assertions are on the specific refusal rather than
on balances happening not to move.

`filetests/caller_teller_sub_realm_filetest.gno` covers both directions of the
`realm.Sub` case, which a raw path comparison gets wrong: a teller invoked from
the token realm's own sub identity, and a token *created* from a sub frame then
used from its host. Both sides resolve the host with `chain.SplitPkgSubPath` —
`guardHome` on the invoking path, `NewToken` on the stored `origRealm` — since
stripping on only one side strands the token in whichever identity it was born
under.

**Negative controls.** Each guard was disabled in turn and the corresponding
test confirmed to go red: `guardHome` returning early makes
`grc20_callerteller_home.txtar` fail at its first exported-teller drain
(line 45, the refusal no longer raised); dropping the strip on either the
invoking path or the stored `origRealm` fails the sub-realm filetest, on the
sub-invocation and sub-creation case respectively.

`gno test` + `gno lint` green on `grc20`, `wugnot`, `foo20`, `grc20factory`,
`grc20reg`, `bar20`, `eventix`; full `TestTestdata` green.
