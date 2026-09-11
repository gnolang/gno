# ADR: Bind author-declared gnomod fields to the package approval hash

## Status

Proposed (PR pending). Supersedes the `PkgHash` section of
`pr6088_msgrun_allowlist_and_inert_charging.md`, which recorded the opposite
decision (exclude `gnomod.toml` from the hash).

## Context

Under the `inert` code submission policy a package is parked at submit and only
runs when an approver sends `MsgEnablePackage` naming the source they reviewed,
as `PackageContentHash` computes it. The hash exists because approval otherwise
names only a path, and the creator may replace parked bytes at any time — so an
approver who read GOOD can be made to activate EVIL.

`PackageContentHash` excluded `gnomod.toml` entirely. The justification was that
`stampGnomod` rewrites that file at submit with the creator, height and declared
deposit, so the stored copy is not the submitted one and the two could never
agree on a hash that included it.

Excluding the whole file also excluded the fields the **author** declares:
`private`, `draft`, `ignore`, `replace` and the gno version.

`checkGnomodConstraints` re-checks `draft`, `replace` and the private override at
enable, but the private guard is `priorPrivate && !gm.Private`, and
`priorPrivate` is false when nothing is live at the path yet. So a **first**
deployment could be approved as `private = true` and activated as
`private = false`: same `.gno` files, same hash, a realm that goes live public
after being approved private. Private realms cannot be imported by others and
their objects cannot be referenced from outside, so flipping the flag exposes
state persisted under an invariant the approver signed off on.

## Decision

### 1. Normalize `gnomod.toml` instead of excluding it

`PackageContentHash` parses the file, resets the fields the keeper owns, and
hashes the canonical `WriteString()` output together with the `.gno` files. The
submitter's copy and the stored copy canonicalize to the same bytes, while every
field the author declares is bound to the approval.

### 2. One list of keeper-owned fields, not two

The reset lives in `keeperOwnedGnomod`, and **`stampGnomod` calls it** rather
than repeating the list.

Two lists drift, and the consequence is not cosmetic. `stampGnomod` also sets
`gm.Module = pkgPath`, which the first cut of this change did not reset — and
nothing on the submit path requires the author's `module` line to equal the
deploy `-pkgpath` (`keeper.go` carries an `XXX` acknowledging exactly that). A
package submitted from a copy-pasted template whose `module` was never updated
would park fine and then be **permanently unenablable**: the approver's `-pkgdir`
hash covers the author's value, the keeper's covers the deploy path, and the
refusal reads `it changed after review` for bytes nobody changed. `gpao` would
retry it to `maxEnableAttempts`, paying a fee each time, and file it under
`statusGaveUp`.

`TestPackageContentHashSurvivesTheRealStamp` runs the real `stampGnomod` and
fails on the next field that is stamped but not reset. The pre-existing test
named for this property imitated the stamp by appending a hand-written
`[addpkg]` block, which is why the `module` case shipped green.

### 3. `MsgEnablePackage.PkgHeight` pins the submission

The hash covers what the author's directory declares. It structurally **cannot**
cover `[addpkg]`, because the approver's local copy has no such section — the
keeper writes it. So a creator can re-park byte-identical sources and keep the
hash while changing `creator`, `height` and `max_deposit`.

`max_deposit` is the sharp end: lowered under a standing approval, the enable
passes the hash gate, runs `init()`, and only then aborts in
`processStorageDeposit`. The approver pays gas for a transaction that could never
have succeeded, and the creator can repeat it for the price of a submission.

Pinning individual fields would need the approver to know values they cannot
read locally. Pinning the **height** closes the whole class at once — every
re-park lands at a new height — including a creator swap after
`MsgRejectPackage`.

Zero means unpinned, which is what every approval predating the field decodes as
and what genesis replay carries, so it cannot be required. `ValidateBasic` is
deliberately left alone for the same reason: it runs before the replay
exemptions in `EnablePackage`.

### 4. Errors instead of an empty hash

`PackageContentHash` returns `(string, error)`. The empty string is also what a
message carrying no approval decodes to, and both CLI callers assigned it
unchecked — so a `-pkgdir` pointing at a directory with no `gnomod.toml`, a
malformed one, or one over `gnomod.maxFileSize` produced a signed approval naming
no source, refused on chain after the fee with a message about the transaction
rather than the directory.

`EnablePackage` also now parses `gnomod.toml` **once**, before the hash check
rather than after it. An unparseable stored blob is reported as what it is
instead of as a hash mismatch with an empty right-hand side, and the enable path
no longer decodes the same blob twice — it charges no `chargePreprocessGas`, so
that decode was unmetered.

The deprecated `gno.mod` is refused here rather than accepted via
`gnomod.ParseMemPackage`'s fallback. `checkGnomodConstraints` refuses it at
submit, so the fallback could only ever produce a confident hash over a file the
chain will never hold.

### 5. Unsupported gno versions are refused at submit

`ParseCheckGnoMod` **panics** on a version the toolchain cannot compile, and it
is reached from the type checker — which the inert submit branch never runs. So
`gno = "0.8"` parked cleanly and detonated inside `EnablePackage`, on the
approver's transaction and gas, for a mistake the submitter made.
`checkGnomodConstraints` now refuses it, on all three paths, with the same
accepted set `ParseCheckGnoMod` uses (empty means "default to latest").

### 6. The canonical encoding is now consensus-critical, so it is pinned

Two defects in `tm2/pkg/toml`'s control-character escape made
`Marshal(Unmarshal(x))` disagree with `Unmarshal(x)`:
`intRr := uint16(rr); if intRr < 0x001F` excluded U+001F, which was then written
raw into a basic string the lexer refuses, and the conversion truncated, so a
rune whose low 16 bits fall under `0x1F` (U+1000A) was emitted as the escape for
those bits alone.

Since `gnomod.toml` is *stored* re-encoded and hashed after a further round trip,
an encoding that is not a fixpoint is an approval no approver can match. Both are
fixed in the fork (`TestBoundsMarshalRoundTripsEveryRune`).

`TestPackageContentHashIsPinned` freezes the digest for a fully-populated
`gnomod.toml`. Its job is not to make the hash immutable but to make moving it
deliberate: the hash shifts whenever `gnomod.File` gains or reorders a field, or
the encoder changes, and none of those look like consensus changes at the call
site that makes them.

## Consequences

**This is a state-machine-breaking change.** Every parked package's expected hash
differs from the previous binary's. Approvals prepared offline with
`enablepkg -broadcast=false`, and any hash recorded for `-pkg-hash`, are refused
after the upgrade with `it changed after review` although nothing changed. They
must be recomputed. Genesis replay is unaffected: the hash check is skipped
there.

Packages already parked whose `gnomod.toml` declares an unsupported gno version,
or whose stored file cannot be parsed, are now refused at enable with an accurate
error rather than a panic or a misattributed hash mismatch. They were not
enablable before either.

## Alternatives considered

**Fix `checkGnomodConstraints` instead.** The failing guard is the private
override, so the narrow fix is to make it fire on a first deployment. Rejected:
it addresses `private` and leaves `draft`, `ignore` and the gno version outside
the approval, which is the same shape of bug waiting for the next field.

**Record the approved flags on `MsgEnablePackage` rather than in the hash.**
Self-describing, versioned by amino field numbers, and it does not couple
consensus to a byte-stable TOML encoder. Rejected as the primary mechanism
because it binds only the fields somebody remembered to enumerate — a new
`gnomod.File` field is silently unbound, which is the bug this ADR exists to
close. The re-encode picks up new fields automatically. `PkgHeight` is that
mechanism applied where the hash genuinely cannot reach.

**Hash the parsed fields explicitly instead of re-encoding.** Stable across
encoder changes and free of the idempotence requirement, but it has the same
"new field silently unbound" failure as the option above. Rejected for the same
reason; the encoder is fixed and pinned instead.

**Validate `module == pkgpath` at submit.** Would make the `module` divergence
unrepresentable rather than harmless. Rejected as out of scope: it refuses
submissions that work today, which is a policy change rather than a fix, and the
`XXX` at `keeper.go` suggests it was already considered and deferred.
