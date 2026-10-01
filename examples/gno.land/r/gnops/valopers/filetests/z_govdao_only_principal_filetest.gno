// PKGPATH: gno.land/r/zzprincipal
//
// The repointed authority names GovDAO, so a foreign realm crossing into
// valopers is NOT the principal the gate accepts.
//
// This is the runtime counterpart to
// ./z_foreign_realm_capability_filetest.gno. That file pins that the
// privileged CAPABILITY cannot be obtained; this one pins that the
// PRINCIPAL cannot be minted — which is the other half of what the gate
// depends on, and the half that was actually broken.
//
// An earlier version of this file asserted only the rendered description
// plus "instructions were not defaced" without ever attempting a write, so
// it went red on a repoint revert purely because the golden path string
// changed, and stayed GREEN with the authority swapped for a fully
// permissive one. It now drives the two routes an attacker actually has.
//
// Why it matters: under the previous ContractAuthority(ownPath) +
// DoByCurrent pairing this property did not hold at all. rlm.Address()
// inside any valopers crossing frame is unconditionally valopers' own
// address, so the gate accepted every caller and its only real defence was
// that no exported entrypoint happened to reach updateInstructions.
//
// And pointing the gate at r/gov/dao only helps while that path is
// unmintable. dao.NewSimpleExecutor + SimpleExecutor.Execute are both
// exported, and Execute is declared in r/gov/dao, so before Execute gated
// on its invoker ANY realm could mint a frame whose Previous() was
// r/gov/dao and satisfy this gate — directly for a closure it obtained,
// and for any exported valopers function whose signature was assignable to
// `func(realm) error`. Section 2 below is the pin for that.
package zzprincipal

import (
	"strings"

	"gno.land/p/nt/testutils/v0"
	"gno.land/r/gnops/valopers"
	"gno.land/r/gov/dao"
	daov3init "gno.land/r/gov/dao/init/v0"
)

// A real GovDAO exists and this realm is not a member of it, so the final
// step fails on the proposal-creation gate rather than on "DAO not
// initialized" -- which would pass for the wrong reason.
var govdaoMember address = testutils.TestAddress("govdao-member")

func init(cur realm) {
	daov3init.InitWithUsers(cross(cur), govdaoMember)
}

// reachInstructions is the shape an attacker needs: something declared
// where it will run with a useful frame. Declared HERE, so it carries the
// attacker's identity, not valopers' — the point being that even with the
// governance principal minted for it, it cannot reach the write.
func reachInstructions(cur realm) error {
	// Nothing exported by valopers reaches updateInstructions, so the most
	// an attacker can do from a forged governance frame is build the sealed
	// request again. Build it, to prove construction is not the hole.
	_ = valopers.NewInstructionsProposalRequest(cross(cur), "PWNED-VIA-FORGED-PRINCIPAL")
	return nil
}

func main(cur realm) {
	// 1. The authority names GovDAO, and says so in a way a reader can act
	//    on. A non-canonical proposer cannot reproduce these bytes.
	println("authority:", valopers.Auth())

	// 2. The governance principal is not mintable. Wrapping any callback in
	//    dao's own exported executor used to hand out a frame whose
	//    Previous() is r/gov/dao; Execute now refuses an invoker from
	//    outside the proxy.
	e := dao.NewSimpleExecutor(0, cur, reachInstructions, "")
	println("forged-principal Execute:", e.Execute(cross(cur)))

	// 3. The only remaining route is to ask GovDAO to adopt a request. A
	//    hostile realm cannot: this aborts (a GovDAO member still can, and
	//    then needs the votes — covered by
	//    r/gnops/valopers/proposal/filetests/z_governed_instructions_filetest.gno).
	rendered := valopers.Render("")
	println("instructions defaced:", strings.Contains(rendered, "PWNED"))
	println("authority unchanged:", valopers.Auth())

	dao.MustCreateProposal(cross(cur),
		valopers.NewAuthorityRotationProposalRequest(cross(cur), "gno.land/r/attacker/dao"))
	println("UNREACHABLE: outsider rotated the authority")
}

// Output:
// authority: contract_authority[contract=gno.land/r/gov/dao,proposer=contract-identity]
// forged-principal Execute: execution denied: executors are only invocable by gno.land/r/gov/dao
// instructions defaced: false
// authority unchanged: contract_authority[contract=gno.land/r/gov/dao,proposer=contract-identity]

// Error:
// proposal creation must be done directly by a user or through the r/gov/dao proxy. caller realm: realm{gno.land/r/gov/dao:g1p84dvfh4wrplyxx4zsmy77a8rxcnjup2j7zv5r}; caller's previous: realm{gno.land/r/zzprincipal:g1036vewy55meppdzpngzsufahr389mldycavect}
