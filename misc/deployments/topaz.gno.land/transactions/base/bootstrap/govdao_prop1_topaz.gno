// govdao_prop1_topaz.gno runs as a MsgRun at genesis to bootstrap the
// topaz governance state. topaz is a fresh chain (no hardfork, no
// historical replay), so this bootstrap is minimal compared to
// gnoland1's (see misc/deployments/gnoland1/govdao_prop1.gno):
//
//   - NO validator-set seed via a realm: the consensus valset comes from
//     GenesisDoc.Validators (the build script sets it via `gnogenesis
//     validator add`), and InitChainer seeds r/sys/params' valset:current
//     from it, so r/sys/validators/v3 proposals manage the set
//     post-genesis.
//   - NO transfer lock and NO unrestricted-accounts proposals: topaz
//     token transfers are unrestricted.
//   - NO temporary-deployer T1 dance: with no proposals to execute at
//     genesis, the T1 member is written directly. At this point the
//     loader has initialised the DAO and tier structure but has NOT set
//     AllowedDAOs — the empty-list corner case in proxy.gno:InAllowedDAOs
//     lets the deployer's calls through.
//   - names.Enable is NOT called here: on master it is gated on the
//     admin address hardcoded in r/sys/names/verifier.gno, which the
//     deployer is not. It runs as a separate genesis MsgCall with its
//     caller field patched to that admin (see
//     transactions/migration/names-enable/).
//
// Steps:
//  1. Add aeddi as the sole GovDAO T1 member.
//  2. Lock down memberstore access by setting AllowedDAOs to the
//     production member-mutation realm only.
package main

import (
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/impl"
	"gno.land/r/gov/dao/v3/memberstore"
)

// govdaoT1Addr is aeddi's address — the sole GovDAO T1 member at launch.
// Additional members are added post-genesis via GovDAO proposals.
const govdaoT1Addr = address("g1aeddlftlfk27ret5rf750d7w5dume3kcsm8r8m")

func main(cur realm) {
	ms := memberstore.Get(0, cur)

	// ---- 1. Add aeddi as sole T1 member ----
	must(ms.SetMember(memberstore.T1, govdaoT1Addr, memberstore.NewMember(3)))

	// ---- 2. Lock down memberstore access ----
	// gno.land/r/gov/dao/v3/impl is the production member-mutation realm;
	// after this call it is the only realm allowed to mutate membership.
	dao.UpdateImpl(cross(cur), dao.NewUpdateRequest(impl.NewGovDAO(), []string{"gno.land/r/gov/dao/v3/impl"}))
}

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}
