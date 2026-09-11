# Bind the private-redeploy exemption to the deploying address

## Status

Proposed

## The problem

A package path can be deployed once. `AddPackage` refuses a second submission
with "package already exists" — unless the live package is private:

```go
pv := gnostore.GetPackage(pkgPath, false)
if pv != nil && !pv.Private {
    return ErrPkgAlreadyExists("package already exists: " + pkgPath)
}
```

The exemption exists so a private realm can be iterated on (#4877). It asks what
the live package **is** and never who it belongs to, so it waives the refusal for
every submitter rather than for the address that deployed it.

Nothing else in the path supplies the missing owner test.

- The creator-bound guard beside it reads `GetInertPackage`, so it covers bytes
  still parked awaiting an approver. `EnablePackage` ends with
  `DelInertPackage`, so after every successful activation that slot is empty and
  the guard has nothing to compare against.
- `checkGnomodConstraints` requires only that the newcomer also be private.
- `checkNamespacePermission` returns nil while `SysNamesPkgPath` is empty or
  `r/sys/names` is undeployed, and the deployed verifier opens with
  `if !isEnabled { return true }`, where disabled is that realm's compiled-in
  initial value. No transaction in `gno.land/genesis/genesis_txs.jsonl` touches
  `r/sys/names` at all, so this is the state the in-tree genesis produces.

One `MsgAddPackage` from any funded address therefore replaces a live private
realm. `stampGnomod` records the sender as `addpkg.creator`, `RunMemPackage`
re-runs `init()` with `OriginCaller` set to them — which is what `p/nt/ownable`
records as the owner — and the realm keeps its path-derived address, its coins
and its callable surface while running somebody else's code. `private` gates
imports and cross-realm references; it never gated `MsgCall`.

Under `code_submission_policy = "inert"` the same takeover runs in two messages,
and the review step does not stop it: `PackageContentHash` excludes
`gnomod.toml`, so byte-identical source parked by a different address hashes to
the digest the approver already signed. `EnablePackage`'s live-package branch
parses the live `gnomod.toml`, refuses on `!liveGm.Private`, holds
`liveGm.AddPkg.Creator`, and never reads it.

## The decision

The creator stamped into the live package's stored `gnomod.toml` is the
deployment's owner of record, and only that address may replace it.

`checkRedeployPermission` compares that stamp against the address the submission
would install as the new creator. Two call sites cover the three paths:

- `AddPackage`, at the exemption itself and above the policy branch, so the
  ordinary redeploy and the inert park inherit one check.
- `EnablePackage`, for a package parked *before* anything was live at the path.
  `AddPackage`'s check evaluated that submission against an empty path, so the
  second half of the deploy has to ask again — the same reason the gnomod rules
  are re-applied there.

The stored blob is the right place to read the owner from. It is the field
`EnablePackage` already reads back to decide who `init()` runs as and who pays
the storage deposit, and the field the parked-blob guard already treats as
deciding who may replace a submission. One record, one meaning, on all three
paths.

### Genesis delivery is exempt

Both call sites waive the binding when `auth.IsGenesisReplay(ctx)` holds, which
is every transaction `InitChain` delivers. `EnablePackage` already waives its
policy, approver and `pkg_hash` gates there on the rule that a fork reproduces a
record rather than granting it again, and this is a fourth gate of the same kind.
Live traffic is unaffected; there is no stranger at `InitChain` to refuse, since
genesis content is whatever the operator wrote.

Enforcing it during replay would break two things. A chain whose history holds a
cross-address private redeploy — the transaction this rule now refuses — would
stop replaying, so forking it either aborts at boot under the default
`PanicOnFailingTxResultHandler` or, under `-skip-failing-genesis-txs`, comes up
silently diverged from the chain it forked. And the hardfork migration path
would break for private realms: `gnogenesis fork addpkg` stamps its transaction
with `--deployer`, which equals the realm's original creator only when the source
directory already carries an `[addpkg] creator` for `LoadPackagesFromDir` to pick
up — true of a directory exported from the chain, false of one taken from
`examples/`.

### The owner of record may clear a squatted inert slot

`MsgRejectPackage` accepts the live package's stamped creator alongside the
parked blob's creator and an approver.

Without it the binding turns a self-clearing state into a permanent one. A blob
parked at a path before anything was live there can, once a private realm is
deployed, never be enabled — this is exactly what the `EnablePackage` call site
refuses. It is then dead weight that still blocks the owner, because
`AddPackage`'s parked-blob guard names its submitter and so refuses the *owner's*
own redeploys for as long as it sits there. Previously an approver's enable
consumed it, which was the takeover this ADR closes. Giving the live owner
standing to reject leaves them a move of their own; without it the only parties
who can unjam the path are the squatter and an approver.

`vm/qpkgmeta_json` reports the same fact: a live path whose pending submission
cannot be enabled carries `ReasonOwnerMismatch`, so an approver polling for work
does not pay a flat fee per attempt on a refusal the chain can predict.

### Namespace ownership and deployment ownership are separate rules

Both run, in that order. `checkNamespacePermission` answers whether an address
may deploy under a name at all; the creator binding answers whether it may
replace what is already deployed there. Neither implies the other: namespace
authority is assignable and can move to a different address after a realm is
live, while the stamped creator records who actually deployed the code that is
running. The chain applied one rule for the strictly weaker act (a first deploy)
and none for the destructive one (replacing a live realm).

### Scope

Redeploying one's own private realm stays allowed. It is what the exemption was
added for, and the existing keeper tests cover it.

Refusing re-activation over a live realm outright would close more than this:
the redeploy path builds its machine with an empty `MachineOptions.PkgPath`, so
`NewRealm` rewinds the realm's object clock, orphans the prior object graph and
zeroes the recorded `Storage` and `Deposit`. That is a separate decision with its
own migration cost, and it is not made here.

## Alternatives considered

**Put the check in `checkGnomodConstraints`, beside "a private package cannot be
overridden by a public package".** That function already runs on all three paths
and already receives the "a private package is live here" fact, so it looks like
the shared spot. It runs too late on the ordinary path: `AddPackage` deletes the
live source blobs and type-checks the submission before calling it, so the live
creator is no longer readable there, and a refused submitter has already paid
for a compile.

**Compare against the namespace owner rather than the stamped creator.** This
would make the right to replace a private realm follow the name, so a
re-registered name would carry it. Rejected as the *only* test: namespace
enforcement is off by default and off in the in-tree genesis, which is precisely
the bootstrap state the parked-blob guard was written for, so a namespace-only
test leaves the hole open on the chain as it ships. It remains a coherent
*additional* authority if namespace transfer is ever meant to carry deployment
rights.

## Consequences

**Consensus: hardfork.** A validator on the old binary accepts a submission the
new one refuses. On the old binary that transaction writes the `pkg:` and
`pkg:#allbutprod` blobs, a fresh package value and realm, the objects `init()`
creates, and a storage deposit moved into the escrow address derived from the
path; on the new binary it writes none of them and reports a `PkgExistError`,
whose type URL is part of the amino-encoded `ABCIResult.Error` that is merkleized
into `LastResultsHash`. Old and new nodes therefore disagree on both the app hash
and the results hash for such a block.

Gas moves on the private-redeploy path: `AddPackage` adds one `GetMemPackage`
read and its amino decode, both metered. `EnablePackage` adds nothing — it reuses
both the blob and the parsed `gnomod` the liveness probe above it already had.
Gas is not part of `ABCIResult` (`Error`, `Data`, `Events`), but the block gas
meter sees the difference, and a submission sized near its `GasWanted` can flip
to out-of-gas.

**Chain state.** No repair is needed, and no fork needs patching. A chain whose
history holds a cross-address private redeploy replays it unchanged, because
genesis delivery is exempt; the fork comes up in the state its source chain was
in, with the rule in force from its first live block. Where a takeover has
already landed, the victim's orphaned objects and the deposit stranded in the
path's escrow are not repaired by this change — healing that is a `--patch-txs`
decision for the operator, not something the binary should make for them.

**A live private package with no stamped creator was replaceable by anyone and
is now replaceable by nobody.** Every deploy through `AddPackage` stamps
`addpkg.creator`, so a creator-less live blob can only come from a binary older
than the field (it was `upload_metadata.uploader` until #4475) or from a
hand-written genesis file. A creator written into such a file in a non-canonical
bech32 spelling locks the path the same way: `bech32.Decode` accepts an
all-uppercase address, and the comparison is against the stored string, as the
parked-blob guard's is. Both cases fail closed — the path keeps its code and
refuses every replacement — which is the safe direction for a rule about who
owns a deployment. Such a path is not beyond rescue: genesis delivery is exempt,
so a hardfork migration transaction can still replace it.

**An owner who has rotated keys can no longer redeploy their own private realm.**
The stamp records the address that deployed, and the chain holds no key-rotation
record, so a new key is a stranger here. This is accepted: the alternative is to
admit some second authority as equivalent to the deployer, and the only candidate
on offer — namespace ownership — is off by default and re-assignable, which is
how the hole arose. An owner in that position deploys at a fresh path, or asks
governance for the namespace-based authority to be added deliberately.

The same reading applies to in-realm ownership: a realm that has transferred its
`p/nt/ownable` owner still answers only to its original deployer for a redeploy,
because the two are different notions of owner. In-realm ownership governs what
the running code permits; the stamped creator governs who may replace that code.
A realm that wants the first to carry the second has to express it in its own
API, not by replacing its bytes.

**The refusal names the live creator.** It is a `PkgExistError`, the same type
the parked-blob guard returns, with the address on the wrapped trace that reaches
`Result.Log`. A submitter who typed the wrong path sees who holds it rather than
a bare "already exists".

## Validation

- `TestVMKeeperPrivateRedeployIsCreatorBound` — ordinary path, default
  permissionless policy, no approver: a stranger's `MsgAddPackage` over a live
  private realm is refused naming the owner, the realm keeps its source, its
  recorded owner and the state written after the deploy, and the owner's own
  redeploy still runs `init()` again.
- `TestVMKeeperInertPrivateRedeployIsCreatorBound` — inert path: a stranger's
  park over a live private realm is refused and nothing is left queued for an
  approver, while the owner may still park and activate a replacement.
- `TestVMKeeperEnableCannotRedeployAnotherAddressesPrivateRealm` — a package
  parked before anything was live, activated after another address deployed
  there: refused at enable, with both blobs private so the private-override rule
  cannot account for the refusal.
- `TestInertPrivateRealmSurvivesAStrangersSubmission` — the same refusal through
  real signed transactions and the ante handler, then `Origin()` and `vm/qfile`
  showing the realm is still the owner's.
- The vm keeper suite covers same-creator private redeploy on both paths
  (`TestVMKeeperAddPackage_UpdatePrivatePackage`,
  `TestVMKeeperEnableCanRedeployALivePrivatePackage`) and both pass unchanged,
  as do the `addpkg_private` integration scripts.

## AI assistance

Written with AI assistance (Claude Code). The fix was driven test-first, and the
diff and this record were revised through several review rounds. The human
author reviewed and owns the change.
