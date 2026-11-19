# ADR: Bring the stored stdlibs in line with the binary's at VM init

## Status

Proposed. Adding a stdlib is implemented; extending one with new files is
designed below and not yet implemented. Builds on #6508 (stdlibs embedded in
the node).

## Context

A stdlib has two halves. Its Go natives are compiled into the binary and
resolved by name at run time (`stdlibs.NativeResolver`), so a new binary
brings them. Its `.gno` sources are state: `InitChainer` writes every stdlib
once, at genesis, and nothing writes them again. Every node re-runs the stored
sources at boot (`PreprocessAllFilesAndSaveBlockNodes`), never its own.

So a release that adds or changes a stdlib does not reach a chain upgraded in
place. Adding one is worse than a no-op: `VMKeeper.Initialize` type-checks
every stdlib the binary lists, finds no package for the new one, and panics.
`crypto/bls12381` (#6375) is the first stdlib added since `gnoland-1`'s
genesis, and hit exactly that.

#6508 embeds the stdlib sources in the node, so the binary carries the exact
sources its natives were built from, and every node running a release holds
the same bytes. That is what makes a migration of the stored stdlibs
deterministic, if it runs at the same point on every node.

## Decision

`VMKeeper.Initialize` brings the stored stdlibs in line with the embedded ones,
after rebuilding the stored packages and before type-checking them
(`migrateStdlibs`). For each stdlib of `stdlibs.InitOrder()`, it compares the
stored production files with the embedded ones:

- **the same:** nothing to do. This is every boot but one.
- **missing:** the stdlib is run into the store from the embedded sources,
  exactly as genesis runs it.
- **extended with new files:** see *Extending a stdlib* (designed, not yet
  implemented; refused until then).
- **any other change:** refused, naming the stdlib.

### Only at a versioned halt

The writes go to the multistore `Initialize` is given, which the app writes
back at boot, so they land in the app hash of the next block. That is
deterministic only if every node makes them at the same height. The keeper
therefore migrates only when the app says the node boots right after a
governance halt that names a version (`atVersionedHalt`):
`halt_height > 0`, `LastBlockHeight == halt_height` and
`halt_min_version != ""`. At that boot every node runs the new binary on the
same state; `halt_min_version` keeps an older binary from resuming, and the
startup checks keep the new one from starting early.

Outside that boot, a stdlib that differs from the binary's stops the node,
naming it. A halt with an empty `halt_min_version`, `skip_upgrade_height`, or a
release rolled out without a halt would otherwise write the change at whatever
height each node happens to boot, and split the network; this way only the
node that makes the mistake stops.

A node that crashes after the migration but before committing the next block
boots at the halt height again and migrates again from the same state. A node
joining from genesis replays each version over its own range (the release
ledger, `gno.land/pkg/upgrades`), so it boots the new binary at the same halt
height and makes the same writes.

### Why at init and not in a migration handler

#6507 runs state migrations in the block after a halt. Stdlibs do not go
there: the new binary's `Initialize` runs before that block and type-checks
the stdlibs it lists, so it must already find them; and a change derived from
the binary's own sources needs no code per release. Handlers at `H+1` then see
the new stdlibs.

### Extending a stdlib

What a release can change in an existing stdlib is narrow, because realms
hold references into stdlib state:

- A stored function value does not hold its body. It holds a source location
  (file, line, column), resolved against the block nodes rebuilt from the
  stored source at every boot (`FuncValue.GetSource`). Any edit to an existing
  file moves the declarations after it, and the stored function values,
  including those realms hold (`gnovm/tests/files/zrealm_stdlib_ref.gno`),
  resolve to the wrong node or to none.
- Realms reference stdlib objects by object ID. Re-running a stdlib from its
  new sources renumbers its objects: on #6387's branch, `strconv.Itoa` moves
  from `:869` to `:881` because `int256.gno` sorts before `itoa.gno`.

Adding new files is the one change that moves no existing location. Extending
a stdlib with them keeps every existing object, with its ID and its block
index, and appends the new declarations:

1. **Run the new files into the existing package.** `runFileDecls` already
   runs extra files into a package node that has a file set, predefining only
   them, so their names take block indices after the existing ones;
   `PrepareNewValues` appends their values to the package block. New declared
   types are saved, and the new files' `init` functions run.
2. **Number new objects after the existing ones.** A stdlib's objects are
   numbered by a throwaway realm counting from zero, and no counter is
   persisted. The extension starts it from the highest object ID the stdlib
   holds, found by scanning its `oid:` keys at boot.
3. **Re-save the existing objects** that change: the package value (its file
   blocks) and its block (its values), marked dirty under that realm.
4. **Keep the predefinition order across boots.** Boot predefines a package's
   files in stored order, and stored mempackages must be sorted
   (`MemPackage.ValidateBasic`). A new file that sorts before an existing one
   would be predefined first at the next boot, shifting every existing index.
   The extension persists the names of the files it added, in order, and
   `PreprocessAllFilesAndSaveBlockNodes` predefines those files last.
5. **Replace the mempackage without re-indexing it.** `AddMemPackage` adds a
   package index entry, and boot order keeps a path at its last entry
   (`indexedPackagePaths`), so a re-added stdlib would be rebuilt after the
   packages that import it. The extension replaces the stored blobs in place.

Points 4 and 5 change behavior shared with every package (boot preprocessing
and the store's mempackage API), which is why the extension is a separate
change.

## Alternatives considered

- **A migration handler per release** (#6507). Rejected for stdlibs: the new
  binary's boot needs them before any block runs, and the change is derived
  from sources the binary already carries.
- **Migrate at boot without the guard** (4f9cb6627). Correct only if every node
  boots the new binary at the same height, which nothing checks; a halt
  without a version, `skip_upgrade_height` or a release rolled out without a
  halt split the network.
- **Re-run a changed stdlib from scratch.** Renumbers its objects and moves its
  source locations, so the references realms hold point at the wrong objects
  or bodies.
- **Genesis replay** (`gnogenesis fork`, pr5511). Works for any change, at the
  cost of rebuilding the chain's state; kept for what an in-place extension
  cannot do.

## Consequences

- A release that only adds stdlibs, or adds files to existing ones once the
  extension lands, upgrades in place with a versioned halt and no code beyond
  the stdlib itself.
- A release that edits an existing stdlib file stops every node at the halt
  with an error naming the stdlib. It needs a genesis replay, or a new stdlib
  package instead of an edit.
- Every boot reads the embedded stdlibs and compares them with the stored
  production files: about 2 MB, once per start.
- The halt that brings in such a release must name its version.

## Tests

- `TestInitializeAddsMissingStdlib`: a store lacking a stdlib, booted at a
  versioned halt, gains it; a realm imports it; a later boot changes nothing.
- `TestInitializeRefusesMigrationOutsideHalt` and
  `TestInitializeRefusesChangedStdlib`: the two refusals.
- `TestAtVersionedHalt`: the guard.
- To come with the extension: a realm's stored references into a stdlib still
  resolve after an extension that adds a file sorting first, and after a
  further reboot.
