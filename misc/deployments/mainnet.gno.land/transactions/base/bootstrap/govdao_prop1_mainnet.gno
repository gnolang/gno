// govdao_prop1_mainnet.gno runs as a MsgRun at genesis to bootstrap the
// mainnet governance state. Mainnet is a fresh chain (no hardfork, no
// historical replay), so this bootstrap is minimal compared to
// gnoland1's (see misc/deployments/gnoland1/govdao_prop1.gno):
//
//   - NO validator-set seed via a realm: the consensus valset comes from
//     GenesisDoc.Validators (the build script sets it via `gnogenesis
//     validator add`), and InitChainer seeds r/sys/params' valset:current
//     from it, so r/sys/validators/v0 proposals manage the set
//     post-genesis.
//   - NO temporary-deployer T1 dance: with no proposals to execute at
//     genesis, the members are written directly. At this point the
//     loader has initialised the DAO and tier structure but has NOT set
//     AllowedDAOs — the empty-list corner case in proxy.gno:InAllowedDAOs
//     lets the deployer's calls through.
//   - names.Enable is NOT called here: on master it is gated on the
//     admin address hardcoded in r/sys/names/verifier.gno, which the
//     deployer is not. It runs as a separate genesis MsgCall with its
//     caller field patched to that admin (see
//     transactions/migration/names-enable/).
//
// Mainnet launches with a SOLE GovDAO T1 member — aeddi, under his
// operational key (confirmed 2026-09) — who then adds the remaining
// members post-genesis through regular GovDAO proposals. The other six
// gnolang/multisigs [govdao] keys were confirmed with their owners and
// are recorded in the gen-genesis.sh history (commit 97987fa06) for when
// those proposals are made. He is seeded with NINE invitation points
// rather than the T1 default of three, because each of those six adds
// burns one of HIS — see the comment on the SetMember call.
//
// Steps:
//  1. Add the sole GovDAO T1 member.
//  2. Lock down memberstore access by setting AllowedDAOs to the
//     production member-mutation realm only.
package main

import (
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/impl/v0"
	"gno.land/r/gov/dao/memberstore/v0"
)

func main(cur realm) {
	ms := memberstore.Get(0, cur)

	// ---- 1. Add the sole GovDAO T1 member ----
	//
	// NINE invitation points, not the T1 tier default of three.
	//
	// Each NewT1MemberRequest burns one point of the PROPOSER's balance at
	// execution time (r/gov/dao/impl/v0/prop_requests.gno), and the T1 tier
	// default only governs what the INVITEE is seeded with — every member
	// added this way starts at three regardless of what the inviter holds.
	// So seating the six remaining [govdao] members costs aeddi six points,
	// and three is what the seven-member bootstrap gave every founder.
	// 6 + 3 = 9 reproduces that end state exactly: seven T1 members holding
	// three points each, which is what this chain would have launched with
	// had all seven been seeded at genesis.
	//
	// Three would not merely be tight, it would be short: aeddi would run
	// out after the third proposal (the fourth NewT1MemberRequest panics at
	// creation with "proposer does not have enough invitation points"), and
	// the remaining three members could only be seated by spending the
	// points of members he had just invited — leaving the founding set with
	// an arbitrary, lopsided point distribution that no proposal can
	// rebalance, since nothing in the memberstore grants points after the
	// fact.
	must(ms.SetMember(memberstore.T1, address("g1aeddlftlfk27ret5rf750d7w5dume3kcsm8r8m"), memberstore.NewMember(9))) // aeddi

	// ---- 2. Lock down memberstore access ----
	// gno.land/r/gov/dao/impl/v0 is the production member-mutation realm;
	// after this call it is the only realm allowed to mutate membership.
	dao.UpdateImpl(cross(cur), dao.NewUpdateRequest(impl.NewGovDAO(), []string{"gno.land/r/gov/dao/impl/v0"}))
}

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}
