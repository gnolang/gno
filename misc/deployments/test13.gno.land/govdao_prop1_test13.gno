// govdao_prop1_test13.gno is the test-13 variant of gnoland1's
// govdao_prop1.gno (see misc/deployments/gnoland1/govdao_prop1.gno).
//
// It runs as a MsgRun at genesis to bootstrap the test-13 governance
// state. test-13 is a hardfork of gnoland1: the new chain replays
// gnoland1's historical txs on top of this genesis. To make the replay
// idempotent w.r.t. consensus, this variant differs from gnoland1's
// bootstrap on two points:
//
//  1. NO validator-set seed via r/sys/validators/v2.NewPropRequest.
//     gnoland1's bootstrap registered its 7 launch validators in v2.
//     On the master-based test-13 chain v2 is vestigial — EndBlocker
//     reads the consensus valset from the params keeper (PR #5485, v3)
//     and from GenesisDoc.Validators, not from v2 events. We set the
//     consensus valset directly in GenesisDoc.Validators (build script
//     does this via `gnogenesis validator add`). v2 stays empty; any
//     historical replay that touches v2 (add/remove via govDAO
//     proposals manfred made on gnoland1) succeeds against an empty
//     realm for adds and fails-then-skipped for removes — both
//     cosmetic, no consensus impact.
//
//  2. 10 additional faucet addresses added to the
//     ProposeAddUnrestrictedAcctsRequest call. These are pre-funded
//     addresses we control on test-13 so we can transact under the
//     genesis-restricted-denoms regime. Generating them off the chain
//     (rather than asking the unrestricted GovDAO multisig to send)
//     lets the script run end-to-end without manual steps.
//
// Everything else is identical to gnoland1's bootstrap: the temporary
// deployer becomes T1 → executes prop2 → adds manfred as real T1 →
// removes the deployer → locks AllowedDAOs.
package main

import (
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/memberstore"
	"gno.land/r/sys/params"
)

// genesisDeployerAddr is the address derived from the genesis deployer
// mnemonic (see gnoland1/gen-genesis.sh DEPLOYER_MNEMONIC). test-13
// reuses the same deployer so the post-replay history reproducibly
// references the same g1edq4… address that gnoland1's history references.
const genesisDeployerAddr = address("g1edq4dugw0sgat4zxcw9xardvuydqf6cgleuc8p")

func main(cur realm) {
	ms := memberstore.Get(0, cur)

	// ---- 1. Add temporary deployer as sole T1 member ----
	// Gives the deployer 100% supermajority so proposals pass.
	must(ms.SetMember(memberstore.T1, genesisDeployerAddr, memberstore.NewMember(0)))

	// ---- 2. (gnoland1 step 2 was the v2 valset seed — removed for test-13) ----

	// ---- 3. Set chain parameters via governance proposals ----

	// Unrestricted addresses (can transfer ugnot even when bank is locked).
	govExec(0, cur, params.ProposeAddUnrestrictedAcctsRequest(cross(cur),
		address("g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"), // GovDAO T1 multisig (inherited from gnoland1)
		address("g148583t5x66zs6p90ehad6l4qefeyaf54s69wql"), // gnoland1 airdrop faucet
		// ---- test-13 faucets ----
		address("g15tjaykkykxa7e8nmtagm2swkphchj4j6rnukes"), // faucet-01
		address("g18qhq2fl54lszhmxeyqlvxnwjzc3xpu4nnakclp"), // faucet-02
		address("g18kre0dtu9sz25ux67pgcjfdqhas525rls34xz9"), // faucet-03
		address("g157heusxh73m0wh6myjfd2f69uuwuu77kcc9vhs"), // faucet-04
		address("g18x40r2smn4telaps0cg9cw2znsjhnay9d353qh"), // faucet-05
		address("g16pdtpgrcwtq0hvh5lvdlrffx72pf8exqd4pnzn"), // faucet-06
		address("g18tv2p7jyk8dfqwl07v3vnyarvdz7ggprvlq8kt"), // faucet-07
		address("g1c24pc2rt3clps6tyd97rtsfevluxk5lp9k8n6e"), // faucet-08
		address("g1l58jdp5yannfd027j6yq37hprpkv32lnadshcm"), // faucet-09
		address("g1k28nhw04v54602jkdfrnu25gq07nyc2rehz9vl"), // faucet-10
	))

	// Lock ugnot transfers (restricted denom).
	govExec(0, cur, params.ProposeLockTransferRequest(cross(cur)))

	// ---- 4. (gnoland1 step 4 was names.Enable — skipped for test-13) ----
	// Master's r/sys/names.Enable checks `PreviousRealm().Address() != admin`
	// (post-PR #5244, removed the p/nt/ownable dependency). The genesis
	// deployer that runs this MsgRun is not the admin, so the call panics
	// with "caller is not admin". gnoland1 launched before #5244 and used
	// the Ownable-based Enable that the deployer could call.
	//
	// Consequence: on test-13, namespace enforcement is OFF at genesis.
	// Any address can addpkg under any namespace until someone (the admin
	// multisig) calls Enable manually. For our test cluster this is
	// acceptable; production launches must enable it via a separate
	// admin-driven tx post-genesis.

	// ---- 5. Add govDAO T1 member (manfred) ----
	// Mirrors gnoland1's bootstrap. Manfred is the sole T1 at the start
	// of historical replay so the proposals he made on gnoland1 (add
	// validator, add member, etc.) execute against the same DAO
	// configuration on test-13. Post-replay we rotate T1 from manfred to
	// aeddi via a separate migration (see apply-test-13-replay.sh).
	must(ms.SetMember(memberstore.T1, address("g1manfred47kzduec920z88wfr64ylksmdcedlf5"), memberstore.NewMember(3)))

	// ---- 6. Remove the genesis deployer ----
	ms.RemoveMember(genesisDeployerAddr)

	// ---- 7. Lock down memberstore access ----
	// gno.land/r/gov/dao/v3/impl is the production member-mutation realm.
	// gno.land/r/test13/rotate is a single-use realm shipped only for
	// test-13: a phase-2 migration MsgCall to its Rotate() function swaps
	// the sole T1 from manfred (the gnoland1 inherited T1) to the test-13
	// T1, then Rotate self-ejects from this list. See rotate-pkg/rotate.gno
	// for the gate ensuring it cannot fire post-genesis.
	dao.UpdateImpl(cross(cur), dao.NewUpdateRequest(impl.NewGovDAO(), []string{"gno.land/r/gov/dao/v3/impl", "gno.land/r/test13/rotate"}))
}

// govExec creates a proposal, votes YES, and executes it immediately.
// Works because the deployer is the sole T1 member at call time (100%
// supermajority).
func govExec(_ int, rlm realm, r dao.ProposalRequest) {
	pid := dao.MustCreateProposal(cross(rlm), r)
	dao.MustVoteOnProposal(cross(rlm), dao.NewVoteRequest(dao.YesVote, pid))
	dao.ExecuteProposal(cross(rlm), pid)
}

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}
