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
// those proposals are made.
//
// Steps:
//  1. Add the sole GovDAO T1 member.
//  2. Lock down memberstore access by setting AllowedDAOs to the
//     production member-mutation realm only.
package main

import (
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v0/impl"
	"gno.land/r/gov/dao/v0/memberstore"
)

func main(cur realm) {
	ms := memberstore.Get(0, cur)

	// ---- 1. Add the sole GovDAO T1 member ----
	must(ms.SetMember(memberstore.T1, address("g1aeddlftlfk27ret5rf750d7w5dume3kcsm8r8m"), memberstore.NewMember(3))) // aeddi

	// ---- 2. Lock down memberstore access ----
	// gno.land/r/gov/dao/v0/impl is the production member-mutation realm;
	// after this call it is the only realm allowed to mutate membership.
	dao.UpdateImpl(cross(cur), dao.NewUpdateRequest(impl.NewGovDAO(), []string{"gno.land/r/gov/dao/v0/impl"}))
}

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}
