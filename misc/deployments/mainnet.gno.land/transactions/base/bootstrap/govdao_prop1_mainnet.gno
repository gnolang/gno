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
// The T1 members are the seven people of the gnolang/multisigs [govdao]
// section. Every key below was confirmed with its owner (2026-09); aeddi's
// is his operational key rather than the accounts.csv-derived one.
//
// Steps:
//  1. Add the seven GovDAO T1 members.
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

	// ---- 1. Add the GovDAO T1 members ----
	must(ms.SetMember(memberstore.T1, address("g1aeddlftlfk27ret5rf750d7w5dume3kcsm8r8m"), memberstore.NewMember(3))) // aeddi
	must(ms.SetMember(memberstore.T1, address("g1gzhj234kpajz963z5vf42j4ylddscnkez2wvly"), memberstore.NewMember(3))) // dongwon
	must(ms.SetMember(memberstore.T1, address("g12y822ervrzlghe06v5g888nranl5z5f56985ya"), memberstore.NewMember(3))) // howl
	must(ms.SetMember(memberstore.T1, address("g1ecsuj0q572jr0dhu29q9njtnmw03hyu7tyyvv6"), memberstore.NewMember(3))) // jaekwon
	must(ms.SetMember(memberstore.T1, address("g1xnv893735y5ef0we0e8w3empa55lnvkfcmn3qs"), memberstore.NewMember(3))) // maxwell
	must(ms.SetMember(memberstore.T1, address("g10rsn5amklf2gmzy5eej2f6upg5eju47zvzv4rr"), memberstore.NewMember(3))) // milos
	must(ms.SetMember(memberstore.T1, address("g1m7nvpk9trgj5sqhwdu2r0p9v5e2pzemedglvzt"), memberstore.NewMember(3))) // moul

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
