package vm

import (
	"encoding/json"
	"fmt"
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/stdlibs"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
)

// Inert-package lifecycle: the second half of a deploy under the "inert" code
// submission policy, and the reads that make a parked package visible.
//
// Split from keeper.go because it is a self-contained stage with its own rules
// -- the tests were already split this way (keeper_inert_test.go,
// keeper_inert_guards_test.go) while the code was not. AddPackage's inert
// branch stays in keeper.go: it is a branch inside the submit path, not a
// stage of its own.

// approverGateSatisfied reports whether addr may exercise pkg-approver
// authority for an inert-lifecycle message.
//
// Genesis replay passes unconditionally. A fork reproduces a record rather than
// granting it again, so it must not refuse its own history because
// pkg_approvers moved since -- the same reason EnablePackage exempts replay
// from the policy gate beside this one.
//
// One function rather than the test repeated per message: every inert
// lifecycle message shares this gate, and a copy that forgets the exemption
// does not fail in tests, it fails the next time somebody forks the chain.
func approverGateSatisfied(ctx sdk.Context, params Params, addr crypto.Address) bool {
	return auth.IsGenesisReplay(ctx) || isApprover(params.PkgApprovers, addr)
}

// EnablePackage activates an inert package: runs the typechecker and
// initializes the package so it becomes importable and callable on-chain.
// Only addresses listed in Params.PkgApprovers may call this.
func (vm *VMKeeper) EnablePackage(ctx sdk.Context, msg MsgEnablePackage) (err error) {
	params := vm.GetParams(ctx)
	// Genesis replay reproduces a record rather than granting it again, so the
	// two authorization gates below are exempt: the fork may have moved off
	// "inert" or rotated pkg_approvers since, and must not refuse its own
	// history. Everything after them runs, so a package the source chain parked
	// is activated here as it was there -- which is the point, since AddPackage
	// parks during replay whenever the replayed policy says to.
	replay := auth.IsGenesisReplay(ctx)

	// Enable exists only to complete a submission the "inert" policy split in
	// two, so it is valid only while that policy is in force.
	//
	// Checked before the approver check so the refusal names the actual reason.
	// Without it, parked packages stay activatable forever under any later
	// policy: governance moves to "permissioned" precisely to stop strangers
	// getting code onto the chain, and every package parked during the "inert"
	// era would remain a stranger's pending deploy that one approver could still
	// land. PkgApprovers is not a substitute — it is not cleared when the policy
	// changes, and an approver's mandate was to activate what the policy of the
	// day accepted, not to carry it across a governance decision.
	//
	// This makes parked packages unactivatable once the policy moves, which is
	// the intended outcome; returning to "inert" makes them activatable again.
	// Note that nothing evicts them in the meantime — see DisablePackage.
	if !replay && params.CodeSubmissionPolicy != CodeSubmissionPolicyInert {
		return std.ErrUnauthorized(fmt.Sprintf(
			"code_submission_policy is %q, not %q: packages cannot be enabled",
			params.CodeSubmissionPolicy, CodeSubmissionPolicyInert))
	}
	if !approverGateSatisfied(ctx, params, msg.Approver) {
		return std.ErrUnauthorized(fmt.Sprintf(
			"address %s is not a pkg approver", msg.Approver))
	}
	gnostore := vm.getGnoTransactionStore(ctx)
	memPkg := gnostore.GetInertPackage(msg.PkgPath)
	if memPkg == nil {
		// Nothing parked AND the package is live: the replayed submission took
		// the ordinary path, because the policy at that point in the replayed
		// history was not "inert". The enable is a genuine no-op, and every
		// genesis exported before this branch existed looks exactly like this.
		//
		// Nothing parked and nothing live means replay has already gone wrong --
		// the submission failed, so there was never anything to enable. Returning
		// nil there reports success for a package that is not on the chain, and
		// the fork comes up silently missing a realm with the enable recorded as
		// the last word on it. Refuse, so the replay report names it.
		//
		// Blob presence is the liveness test, for the reason spelled out at the
		// liveBlob probe below: GetPackage would populate the object cache and
		// make RunMemPackage panic later.
		if replay && gnostore.GetMemPackage(msg.PkgPath) != nil {
			return nil
		}
		return ErrInvalidPkgPath("no inert package at path: " + msg.PkgPath)
	}
	// The approver names the source it reviewed, and this is where that is
	// checked. Approval otherwise names a path, and a path's contents can
	// change: the same creator may replace parked bytes at any time, and that
	// replacement is also the legitimate retry after a failed enable. So the
	// creator who parked GOOD and had it reviewed can park EVIL before the
	// enable lands, and nothing above would notice -- the creator-bound guard
	// at submit stops a stranger doing it, not the submitter themselves.
	//
	// Skipped on replay, like the two gates above: history predating the field
	// carries no hash, and refusing it would fail every replayed enable. The
	// bytes are already fixed by then in any case -- replay is not racing a
	// submitter.
	if !replay {
		if msg.PkgHash == "" {
			return ErrInvalidPackage(
				"missing pkg_hash: an approval has to name the source it approves")
		}
		if got := PackageContentHash(memPkg); got != msg.PkgHash {
			return ErrInvalidPackage(fmt.Sprintf(
				"the parked source at %s is not what was approved "+
					"(approved %s, parked %s); it changed after review",
				msg.PkgPath, msg.PkgHash, got))
		}
	}
	// Refuse to activate over a package that is already live, applying exactly
	// the rule AddPackage applies: a public package may not be replaced, a
	// private one may.
	//
	// Enable is the deferred second half of a deploy, so it has to enforce the
	// deploy's preconditions. "The sender is an approver" and "something is
	// parked at this path" are not enough on their own, because they leave a
	// package takeover open. A path can be parked and
	// live at once — the two live in different key spaces, and nothing clears a
	// parked blob when governance moves the policy off "inert" — so: A parks at
	// P and is never approved; the policy flips to permissionless; B deploys at
	// P through the now-open normal path; any approver then enables P and A's
	// bytes replace B's live package, running with OriginCaller = A, which is
	// what p/nt/ownable records as the owner. It does not even panic, because
	// runMemPackage takes its fresh-package branch when MachineOptions.PkgPath
	// is empty, so it silently rebuilds the node and package value over B's and
	// orphans B's realm objects.
	//
	// Deleting the prior blobs for the private case is the same requirement
	// AddPackage documents: AddMemPackage's writes are conditional across the
	// prod and #allbutprod keys and are not a full replace, so a stale sibling
	// would otherwise survive and be served by qfile.
	// Probed through the stored blob, NOT GetPackage. Loading the live
	// PackageValue populates the object cache, and RunMemPackage below then
	// panics in SetCachePackage because the package is already cached — which
	// made the private-replacement branch dead on arrival. AddPackage does read
	// the PackageValue, and escapes only incidentally: checkNamespacePermission
	// re-enters getGnoTransactionStore, whose ClearObjectCache evicts the entry
	// between the read and the run. EnablePackage calls neither, so it must not
	// create the entry in the first place.
	//
	// Blob presence is an exact liveness test here: a parked package is stored
	// under a different key prefix and is invisible to GetMemPackage, while a
	// live one always has a production blob (hasProdGnoFile guarantees it at
	// deploy). `private` comes from the same gnomod.toml the deploy stored.
	gm, err := gnomod.ParseMemPackage(memPkg)
	if err != nil {
		return ErrInvalidPackage(err.Error())
	}
	liveBlob := gnostore.GetMemPackage(msg.PkgPath)
	priorPrivate := false
	if liveBlob != nil {
		liveGm, perr := gnomod.ParseMemPackage(liveBlob)
		if perr != nil || !liveGm.Private {
			return ErrPkgAlreadyExists("package already exists: " + msg.PkgPath)
		}
		priorPrivate = true
	}

	// The full gnomod rule set, re-applied at enable.
	//
	// It ran at SUBMIT, but against the world as it was THEN, and every rule
	// whose answer can change between the two messages has to be asked again.
	// The private-override rule is the one that actually moves: for a package
	// parked before anything existed at this path, submit evaluated it against
	// nothing at all, so without this a public package parked early can be
	// activated over a private realm deployed later and flip it public,
	// retroactively exposing objects persisted under the invariant that nothing
	// outside the realm could reference them.
	//
	// The rest are stable across the split — the stored bytes cannot change, so
	// replaces, draft and gno.mod give the same answer — and are re-checked
	// anyway rather than hand-picked. Enable is the second half of a deploy;
	// enumerating which of a deploy's preconditions it may skip is how the
	// override rule went missing in the first place.
	if err := checkGnomodConstraints(gm, memPkg, msg.PkgPath, priorPrivate, ctx.BlockHeight()); err != nil {
		return err
	}
	// AddPackage wrote the creator into gnomod.toml before storing (see the
	// inert branch), so it round-trips; genesis.go reads it back the same way.
	// gm was parsed above, before the liveness probe that needs it. Parsed here
	// rather than lower down because the namespace check below needs it.
	creator, err := crypto.AddressFromBech32(gm.AddPkg.Creator)
	if err != nil {
		return ErrInvalidPackage(fmt.Sprintf(
			"invalid creator %q in stored gnomod.toml: %v", gm.AddPkg.Creator, err))
	}

	// Re-check namespace and CLA, which ran at SUBMIT against whatever was true
	// then.
	//
	// Same reasoning as the private-override rule above: a package must not reach
	// execution having cleared a weaker rule set than the deploy path enforces.
	// It matters most at bootstrap, because checkNamespacePermission returns nil
	// while r/sys/names is undeployed -- under "inert", the state a chain boots
	// in. Without this an attacker parks under a namespace nobody owns yet, and
	// an approver (typically an oracle checking only that the code type-checks)
	// activates it later under a namespace that by then belongs to someone else,
	// with OriginCaller set to the attacker.
	//
	// Placed HERE, before the type check and RunMemPackage, and not next to the
	// deposit: both of these evaluate a realm, which re-enters
	// getGnoTransactionStore and clears realmStorageDiffs. Running them after
	// RunMemPackage wipes the storage the deposit is computed from, so the
	// creator is charged nothing -- caught by the deposit tests.

	// The chain domain, the last of AddPackage's path rules. chain_domain is a
	// governance param, so a change between submit and enable would otherwise
	// let a package go live under a domain AddPackage would refuse.
	//
	// Not redundant with checkNamespacePermission, which applies the same rule
	// but returns nil early when sys_names_pkgpath is empty. Placed first, as
	// AddPackage orders it: the two checks below each evaluate a realm, so a
	// mismatch would otherwise pay for both before being refused.
	//
	// Same source as AddPackage: params.ChainDomain, defaulted once in
	// applyLegacyDefaults. Two halves of one deploy applying one rule must not
	// be able to read it from two places.
	if !strings.HasPrefix(msg.PkgPath, params.ChainDomain+"/") {
		return ErrInvalidPkgPath("invalid domain: " + msg.PkgPath)
	}
	if err := vm.checkNamespacePermission(ctx, params, creator, msg.PkgPath); err != nil {
		return err
	}
	if err := vm.checkCLASignature(ctx, params, creator); err != nil {
		return err
	}
	// Typecheck the stored package.
	opts := gno.TypeCheckOptions{
		Getter: gnostore,
		// No TestGetter, and ProdOnly: mirrors AddPackage. GetMemPackage
		// returns the production blob only, and resolving test-stdlib imports
		// would make this consensus path depend on node-local state. #5888
		// predates that change on master and passed a test getter here.
		ProdOnly: true,
		Mode:     gno.TCLatestStrict,
		Cache:    vm.getTypeCheckCache(ctx),
	}
	if _, err = gno.TypeCheckMemPackage(memPkg, opts); err != nil {
		return ErrTypeCheck(err)
	}
	// The origin caller is the package's creator, not the approver.
	//
	// The approver authorises activation; they did not write the code. init()
	// runs here, and it commonly records chain.OriginCaller() as the owner
	// (p/nt/ownable's default). Passing the approver would hand every inert
	// package's ownership to whichever approver happened to sign the enable,
	// and would make ownership depend on the order approvers act in. It also
	// diverges from the non-inert path, where init() sees the deployer — so
	// the same source would initialize differently under a different policy.
	//
	// Execute and persist the package.
	ctx = ContextWithParamsAccum(ctx)
	msgCtx := stdlibs.ExecContext{
		ChainID:     ctx.ChainID(),
		ChainDomain: params.ChainDomain,
		Height:      ctx.BlockHeight(),
		Timestamp:   ctx.BlockTime().Unix(),
		// Height/Timestamp are enable-time, not submit-time: init() observes
		// when it actually ran. Only the caller identity is inherited.
		OriginCaller:    creator.Bech32(),
		OriginSend:      std.Coins{},
		OriginSendSpent: new(std.Coins),
		Banker:          NewSDKBanker(vm, ctx),
		Params:          NewSDKParams(vm.prmk, ctx),
		EventLogger:     ctx.EventLogger(),
		// Keyed on the creator to stay coherent with OriginCaller. The creator
		// is normally not a signer of this tx, so this is normally nil — the
		// correct answer, since no session of theirs authorised the enable.
		SessionAccount: getSessionAccount(ctx, creator),
		// Successful inert initialization commits realm state. The package
		// object gets ID 1 before init, so token realms can issue later IDs.
		RealmIDEnabled: true,
	}
	m2 := gno.NewMachineWithOptions(gno.MachineOptions{
		PkgPath:            "",
		Output:             vm.Output,
		Store:              gnostore,
		Alloc:              gnostore.GetAllocator(),
		Context:            msgCtx,
		GasMeter:           ctx.GasMeter(),
		BoundedPanicRender: true,
	})
	defer m2.Release()
	defer doRecover(m2, &err)
	preAlloc := gno.NewAllocator(maxAllocTx)
	preAlloc.SetGasMeter(ctx.GasMeter())
	gnostore.SetPreprocessAllocator(preAlloc)
	defer gnostore.SetPreprocessAllocator(nil)
	if liveBlob != nil {
		// Private redeploy: clear the prior blobs, as the normal path does.
		gnostore.DeleteMemPackage(msg.PkgPath)
	}
	m2.RunMemPackage(memPkg, true)

	// Take the storage deposit for the realm objects this enable just created.
	//
	// It cannot be taken at submit: processStorageDeposit is driven entirely by
	// RealmStorageDiffs(), and nothing has executed yet at that point, so there
	// are no diffs to price. Enable is the first moment the realm state exists.
	//
	// Charged to the creator, not the approver. The creator caused the storage
	// and only their own submission can lock their own funds, so this is
	// consented by the act of submitting. Charging the approver would be worse
	// than merely unfair: an approver is typically an automated oracle with a
	// hot key, so an attacker could submit large packages to bleed its balance
	// and stall approvals for everyone.
	//
	// Capped by the ceiling recorded at submit, read back from the same stamped
	// gnomod.toml the creator address came from.
	//
	// The inert path always records one -- what the submitter declared, or the
	// chain default as it stood at submit time. So the empty case below is not
	// the normal path; it covers a blob written by something other than that
	// branch, such as a genesis file, and falls back to the current default
	// exactly as an ordinary deploy would.
	//
	// This is the only path where the message that DECLARES the ceiling is not
	// the message that SPENDS against it, so the declaration has to be carried
	// or it is silently discarded. Reading params.DefaultDeposit here instead
	// would both ignore a creator who asked for a lower limit and let a
	// governance raise between submit and enable widen their exposure.
	//
	// So the split follows what each stage can know. Submit knows the source
	// length and the creator's declared limit; only execution reveals the realm
	// state, so enable charges the deposit against that limit.
	//
	// Escrowing the ceiling at submit was considered and rejected. The ceiling
	// is not a quote — the sample package in
	// TestVMKeeperEnableTakesStorageDepositFromCreator needs 210_200ugnot
	// against a 100_000_000ugnot default cap — and there is no source-bytes to
	// realm-bytes estimator to size it better (one sample is ~24x, and another
	// in examples/ is ~74x, so it is a data point, not a model). More to the
	// point, escrow would buy little: a failed enable leaves the package parked
	// and retryable once the creator is funded, so the loss it prevents is one
	// approver's gas on a transaction that can be simulated first, while the
	// cost it adds is funds locked on every submission that is never approved.
	//
	// Before DelInertPackage so the ordering reads pay-then-activate. An error
	// aborts the whole message either way, leaving the package inert and
	// retryable once the creator is funded.
	var declaredDeposit std.Coins
	if gm.AddPkg.MaxDeposit != "" {
		declaredDeposit, err = std.ParseCoins(gm.AddPkg.MaxDeposit)
		if err != nil {
			return ErrInvalidPackage(fmt.Sprintf(
				"invalid max_deposit %q in stored gnomod.toml: %v", gm.AddPkg.MaxDeposit, err))
		}
	}
	if err := vm.processStorageDeposit(ctx, creator, declaredDeposit, gnostore, params); err != nil {
		return err
	}

	// Remove from inert store now that it is active.
	gnostore.DelInertPackage(msg.PkgPath)
	return nil
}

// DisablePackage moves an active package back to inert state.
// NOTE: full disable requires evicting executed objects from the base store,
// which is not yet implemented. This stub is provided for interface completeness.
// RejectPackage deletes a package that is parked awaiting approval.
//
// Nothing else could remove one. DelInertPackage ran only after a successful
// enable and DisablePackage is unimplemented, so a submission an approver
// declined occupied the store forever -- and so did one parked under a policy
// that has since moved off "inert", which no enable can ever activate.
//
// Either the creator or an approver may send it. Both have standing: the bytes
// are the creator's, and declining them is the approver's job. Anyone else is
// refused, or a stranger could clear a queue they have no part in.
//
// Not gated on the policy still being "inert". Cleanup is most needed exactly
// when it is not: those packages are unactivatable and would otherwise be
// unremovable too.
//
// The submission charge is not refunded. See MsgRejectPackage.
func (vm *VMKeeper) RejectPackage(ctx sdk.Context, msg MsgRejectPackage) error {
	gnostore := vm.getGnoTransactionStore(ctx)
	memPkg := gnostore.GetInertPackage(msg.PkgPath)
	if memPkg == nil {
		return ErrInvalidPkgPath("no inert package at path: " + msg.PkgPath)
	}

	gm, err := gnomod.ParseMemPackage(memPkg)
	if err != nil {
		return ErrInvalidPackage(fmt.Sprintf(
			"cannot read the parked package at %s: %v", msg.PkgPath, err))
	}

	params := vm.GetParams(ctx)
	if !approverGateSatisfied(ctx, params, msg.Sender) && gm.AddPkg.Creator != msg.Sender.String() {
		return std.ErrUnauthorized(fmt.Sprintf(
			"address %s is neither a pkg approver nor the creator of %s",
			msg.Sender, msg.PkgPath))
	}

	gnostore.DelInertPackage(msg.PkgPath)
	return nil
}

func (vm *VMKeeper) DisablePackage(ctx sdk.Context, msg MsgDisablePackage) error {
	params := vm.GetParams(ctx)
	if !approverGateSatisfied(ctx, params, msg.Approver) {
		return std.ErrUnauthorized(fmt.Sprintf(
			"address %s is not a pkg approver", msg.Approver))
	}
	// TODO: evict executed package objects from baseStore and move source back
	// to inert_pkg key. Tracked in a follow-up PR.
	return std.ErrUnknownRequest("disable_package is not yet implemented")
}

// QueryPaths returns public facing function signatures.
// XXX: Implement pagination
// QueryInertPaths lists packages parked awaiting an approver, under an optional
// path prefix.
//
// QueryPaths cannot answer this: it ranges the live key space only, so a parked
// package is missing from every listing. An operator asking "what is waiting on
// me" had no way to find out, and an oracle restarting had no way to catch up
// on what it missed while down -- it learns about packages by watching blocks.
//
// Deliberately a plain prefix match, without QueryPaths' @user handling. This
// answers an operational question about a queue, not a browsing one.
func (vm *VMKeeper) QueryInertPaths(ctx sdk.Context, prefix string, limit int) (paths []string, err error) {
	// Named returns so the recover has somewhere to write. The iteration below
	// is lazy and metered, and maxGasQuery is a real ceiling, so exhausting it
	// panics with OutOfGasError -- and handleQueryCustom does not recover, so
	// without this the panic leaves the ABCI Query rather than being reported
	// as an error. Under "inert" anyone may park a package, so the size of the
	// key space being walked is not the operator's to bound.
	defer doRecoverQueryNoMachine(&err)

	if limit < 0 {
		return nil, errors.New("cannot have negative limit value")
	}

	ctx = ctx.WithGasMeter(store.NewGasMeter(maxGasQuery))
	gnostore := vm.newGnoTransactionStore(ctx) // throwaway (never committed)
	return collectWithLimit(gnostore.FindInertPathsByPrefix(prefix), limit), nil
}

// Reasons a parked package is not live, reported in PackageMeta.Reason.
//
// Each mirrors a gate in EnablePackage, in the order EnablePackage applies
// them, so the reason is the one a MsgEnablePackage would actually hit.
const (
	// ReasonAwaitingApprover is the ordinary case: nothing is wrong, no
	// approver has enabled it yet.
	ReasonAwaitingApprover = "waiting for a package approver to enable it"
	// ReasonNoApprovers means nothing on this chain can be enabled at all.
	// isApprover is a membership test, so an empty list admits nobody.
	ReasonNoApprovers = "no package approvers are configured on this chain"
	// ReasonPolicyMoved means the chain has left the policy the package was
	// submitted under. Reversible: returning to "inert" makes it enablable
	// again, and nothing evicts the package meanwhile.
	ReasonPolicyMoved = `the chain no longer runs the "inert" code submission policy`
)

// Package statuses reported by QueryPackageMeta.
const (
	PackageStatusLive   = "live"   // deployed and callable
	PackageStatusInert  = "inert"  // submitted, stored, awaiting an approver
	PackageStatusAbsent = "absent" // the chain holds nothing at this path
)

// PackageMeta is the vm/qpkgmeta_json response.
//
// Fields beyond Path and Status are whatever the chain stamped at submit time,
// and are omitted when the package is absent.
type PackageMeta struct {
	Path   string `json:"path"`
	Status string `json:"status"`
	// Creator submitted the package, and pays its storage deposit whichever
	// message activates it.
	Creator string `json:"creator,omitempty"`
	// Height is the block the package was submitted in, not the one that
	// activated it. Under "inert" those differ.
	Height     int    `json:"height,omitempty"`
	MaxDeposit string `json:"max_deposit,omitempty"`
	// Reason says why a parked package is not live yet, in terms its submitter
	// can act on. Empty unless Status is "inert".
	Reason string `json:"reason,omitempty"`
	// Pending reports a parked submission awaiting an approver. It is always
	// true for status "inert", and also true for a live PRIVATE realm with a
	// redeploy parked over it -- AddPackage refuses to park over a live public
	// package but allows it for a private one, so the two can coexist.
	//
	// The fields above describe what is callable. When a live package has a
	// submission pending, they are the live one's, and the parked submission's
	// own creator and height are not reported here.
	Pending bool `json:"pending,omitempty"`
}

// QueryPackageMeta reports what the chain holds at a package path: live, parked
// awaiting an approver, or nothing at all.
//
// Absent is a successful response carrying Status, not an error. Callers have
// to tell "this path was never submitted" from "the node could not answer",
// and an error collapses the two.
//
// It exists because every other query reads the live key space only. Under the
// "inert" policy a submitted package is stored without being activated, so its
// creator could see nothing between paying to submit and somebody approving --
// vm/qpkg_json and vm/qfile both answer "package not found", the same as for a
// path nobody ever used.
func (vm *VMKeeper) QueryPackageMeta(ctx sdk.Context, pkgPath string) (res string, err error) {
	defer doRecoverQueryNoMachine(&err)
	ctx = ctx.WithGasMeter(store.NewGasMeter(maxGasQuery))
	gnostore := vm.newGnoTransactionStore(ctx) // throwaway (never committed)

	info := PackageMeta{Path: pkgPath, Status: PackageStatusAbsent}
	mpkg := gnostore.GetMemPackage(pkgPath)
	if mpkg != nil {
		info.Status = PackageStatusLive
		// Both key spaces are read even when the live one answers, or a
		// redeploy parked over a live private realm would be invisible -- the
		// case this query exists to make visible.
		info.Pending = gnostore.GetInertPackage(pkgPath) != nil
	} else if mpkg = gnostore.GetInertPackage(pkgPath); mpkg != nil {
		info.Status = PackageStatusInert
		info.Pending = true
		info.Reason = enableBlockedReason(vm.GetParams(ctx))
	}

	// A stored package always carries a gnomod.toml the keeper stamped, so a
	// parse failure here is a corrupt store rather than bad input. Report the
	// status anyway: "it is parked" is the answer the caller came for, and is
	// still true.
	if mpkg != nil {
		if gm, perr := gnomod.ParseMemPackage(mpkg); perr == nil {
			info.Creator = gm.AddPkg.Creator
			info.Height = gm.AddPkg.Height
			info.MaxDeposit = gm.AddPkg.MaxDeposit
		}
	}

	bz, err := json.Marshal(info)
	if err != nil {
		return "", err
	}
	return string(bz), nil
}

// enableBlockedReason reports why a parked package cannot be enabled right now,
// checking the same gates as EnablePackage in the same order.
//
// It reads the chain's own configuration only. Whether an approver will choose
// to enable a given package, or whether the creator can still cover the deposit
// at that point, is not knowable here.
func enableBlockedReason(params Params) string {
	switch {
	case params.CodeSubmissionPolicy != CodeSubmissionPolicyInert:
		return ReasonPolicyMoved
	case len(params.PkgApprovers) == 0:
		return ReasonNoApprovers
	default:
		return ReasonAwaitingApprover
	}
}
