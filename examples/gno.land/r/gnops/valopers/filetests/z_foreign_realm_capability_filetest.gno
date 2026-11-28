// PKGPATH: gno.land/r/zzattacker
//
// Consumer-level regression pin for the governance capability leak.
//
// The package-level pins live in p/moul/authz. This is the consumer-level
// one: an unrelated realm must not be able to drive a valopers privileged
// action.
//
// Before the fix this file could not exist as a negative test — the
// realm exported NewInstructionsProposalCallback, a crossing closure
// declared in valopers. The VM mints a crossing frame's `cur` from the
// CALLEE's declaring package, so invoking it from anywhere ran
// updateInstructions with cur.Address() == the ContractAuthority's
// contractAddr. The contract-identity gate accepted, the init.gno
// handler ran the action, and `instructions` was rewritten with no
// proposal and no vote:
//
//	err: undefined
//	PWNED-BY-ATTACKER
//
// The capability is now sealed inside a dao.ProposalRequest, whose
// executor field is unexported with no accessor. This file asserts the
// two halves of that: the leak is gone from the type system, and the
// request an attacker can still construct is inert in their hands.
//
// NOTE for whoever edits valopers next: the property under test is
// "no exported function returns anything carrying this realm's frame
// identity". Re-exporting a func(realm) error, or returning a
// dao.Executor (whose Execute method is exported and directly
// invocable), reopens the hole even though this file would still
// compile.
package zzattacker

import (
	"strings"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/r/gnops/valopers"
	"gno.land/r/gov/dao"
	daov3init "gno.land/r/gov/dao/init/v0"
)

// A real GovDAO exists and the attacker is not in it.
var govdaoMember address = testutils.TestAddress("govdao-member")

func init(cur realm) {
	daov3init.InitWithUsers(cross(cur), govdaoMember)
}

func main(cur realm) {
	// 1. The bare capability is gone. Uncommenting this must not compile:
	//
	//	cb := valopers.NewInstructionsProposalCallback("PWNED-BY-ATTACKER")
	//	cb(cross(cur))
	//
	// What remains is the sealed request. An unrelated realm may still
	// build one — construction is harmless — but dao.ProposalRequest
	// exposes only Title/Description/Filter, so there is no way to reach
	// the executor and run it. Only GovDAO executing the proposal can.
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/zzattacker"))

	req := valopers.NewInstructionsProposalRequest(cross(cur), "PWNED-BY-ATTACKER")
	println("attacker built a request:", req.Title())

	// 2. The privileged write did not happen: the instructions are still
	//    the genesis text, not the attacker's. Under the exported callback
	//    this line already printed "defaced: true".
	rendered := valopers.Render("")
	println("instructions defaced:", strings.Contains(rendered, "PWNED-BY-ATTACKER"))
	println("instructions intact:", strings.Contains(rendered, "Welcome to the **Valopers** realm"))

	// 3. Runtime pin, not merely a symbol-absence one. dao.ProposalRequest
	//    exposes Title/Description/Filter and nothing that yields the
	//    executor, so the one remaining route for a holder is to ask
	//    GovDAO to adopt the request. A hostile realm cannot: this aborts.
	//    (A GovDAO *member* still can, and then it needs the votes — that
	//    is the legitimate path, covered by
	//    r/gnops/valopers/proposal/filetests/z_governed_instructions_filetest.gno.)
	dao.MustCreateProposal(cross(cur), req)
	println("UNREACHABLE: outsider got the proposal adopted")
}

// Output:
// attacker built a request: /r/gnops/valopers: Update instructions
// instructions defaced: false
// instructions intact: true

// Error:
// proposal creation must be done directly by a user or through the r/gov/dao proxy. caller realm: realm{gno.land/r/gov/dao:g1p84dvfh4wrplyxx4zsmy77a8rxcnjup2j7zv5r}; caller's previous: realm{gno.land/r/zzattacker:g1q577nx395ehscyn9ucutvfpj03a5tuk7x4v3ad}
