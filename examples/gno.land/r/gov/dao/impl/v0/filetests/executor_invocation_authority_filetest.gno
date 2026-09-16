// PKGPATH: gno.land/r/test/execauthority
package execauthority

// Regression pin for who may invoke a proposal executor.
//
// dao.Executor.Execute is a public method on a public type and
// dao.NewSimpleExecutor is a public constructor, so any realm can build an
// executor around any func(realm) error it can obtain. Execute is a CROSSING
// method declared in r/gov/dao, so invoking it mints a gno.land/r/gov/dao
// frame for the callback — and that value is a capability: realms gate on it
// (authz.NewContractAuthority("gno.land/r/gov/dao") + DoByPrevious, ownable,
// anything reading Previous().Address()).
//
// Until SimpleExecutor.Execute gated on its invoker, the forged frame was
// byte-identical to an approved one: same PkgPath, same Previous().PkgPath(),
// same Previous().Address(). A callback could not tell a supermajority-backed
// execution from one no member ever voted on, and the forgery minted no
// proposal id, so there was nothing to render, audit or deny afterwards.
//
// The documented containment was never the one holding the line. types.gno
// said invocation was "contained by the executor being unexported inside
// ProposalRequest/Proposal with no accessor" — that protects the executor
// OBJECT, but an attacker never needs someone else's executor: a reachable
// CALLBACK is enough, and an ordinary exported entrypoint of the victim whose
// signature is assignable to func(realm) error is such a callback.
// r/gnops/valopers had exactly that shape.
//
// Sections 2, 3 and 6 are the inversion: the forgery is now refused, the two
// frames are no longer comparable, and the proxy identity is not mintable.
// Section 5 keeps SafeExecutor characterized — it remains dead code that
// cannot be wired as written, since its allowlist holds the impl path while
// an executor's caller is the proxy.

import (
	"chain"
	"strings"
	"testing"

	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/impl/v0"
	"gno.land/r/gov/dao/memberstore/v0"
)

const (
	member  address = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
	outside address = "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn"
)

// observed records the frame the callback was handed, so the approved and the
// forged invocation can be compared.
type frame struct {
	cur      string
	prev     string
	prevAddr address
	ran      bool
}

var seen frame

func probe(cur realm) error {
	seen = frame{
		cur:      cur.PkgPath(),
		prev:     cur.Previous().PkgPath(),
		prevAddr: cur.Previous().Address(),
		ran:      true,
	}
	return nil
}

func init(cur realm) {
	memberstore.Get(0, cur).DeleteAll()
	memberstore.Get(0, cur).SetTier(memberstore.T1)
	memberstore.Get(0, cur).SetMember(memberstore.T1, member, memberstore.NewMember(3))
	dao.UpdateImpl(cross(cur), dao.NewUpdateRequest(
		impl.NewGovDAO(), []string{"gno.land/r/gov/dao/impl/v0"}))
}

func main(cur realm) {
	// 1. The approved path still works: a member proposes, votes it to
	// supermajority, and executes. The gate must not break governance.
	testing.SetOriginCaller(member)
	testing.SetRealm(testing.NewUserRealm(member))

	pid := dao.MustCreateProposal(cross(cur), dao.NewProposalRequest(
		"approved", "carried by a supermajority vote", dao.NewSimpleExecutor(0, cur, probe, "")))
	dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.YesVote, pid))
	println("approved path executes:", dao.ExecuteProposal(cross(cur), pid))
	approved := seen
	println("callback ran:", approved.ran)
	println("approved frame cur :", approved.cur)
	println("approved frame prev:", approved.prev)

	// 2. The forged path: no proposal, no vote, and a caller who is not a
	// member of any tier. Execute is called directly on a locally built
	// executor, and is refused.
	seen = frame{}
	testing.SetOriginCaller(outside)
	testing.SetRealm(testing.NewUserRealm(outside))

	// The same caller cannot open a proposal -- PreCreateProposal rejects
	// non-members. That is the authority level the DAO grants them, and
	// Execute must not grant more.
	_, err := dao.CreateProposal(cross(cur), dao.NewProposalRequest(
		"rejected", "a non-member cannot propose", dao.NewSimpleExecutor(0, cur, probe, "")))
	println("forger cannot even open a proposal:", err != nil)
	println("callback ran:", seen.ran)

	forged := dao.NewSimpleExecutor(0, cur, probe, "no proposal backs this")
	println("forged Execute is refused:", forged.Execute(cross(cur)))
	println("callback ran:", seen.ran)

	// 3. And because it never ran, there is no forged frame to compare: the
	// governance identity was not minted for an unauthorized invoker.
	println("no frame was handed out:", seen.cur == "" && seen.prev == "")
	println("approved prev address was the gov/dao package address:",
		approved.prevAddr == chain.PackageAddress("gno.land/r/gov/dao"))

	// 4. The forgery leaves no trace on the DAO either: no proposal id was
	// minted, so there is nothing to render, audit, or deny after the fact.
	_, missing := dao.GetProposal(pid + 1)
	println("no proposal was created for the forged execution:", missing != nil)

	// 5. SafeExecutor still cannot be wired as written: allowedDAOs holds the
	// impl path, while an executor's caller is the proxy path, so its check
	// rejects the APPROVED path. It remains dead code; the gate that works
	// lives in SimpleExecutor.Execute and keys on the proxy, not on
	// InAllowedDAOs.
	println("InAllowedDAOs(proxy path):", dao.InAllowedDAOs("gno.land/r/gov/dao"))
	println("allowedDAOs holds instead:", dao.AllowedDAOs()[0])

	seen = frame{}
	testing.SetOriginCaller(member)
	testing.SetRealm(testing.NewUserRealm(member))
	spid := dao.MustCreateProposal(cross(cur), dao.NewProposalRequest(
		"safe", "same approved path, wrapped in SafeExecutor",
		dao.NewSafeExecutor(dao.NewSimpleExecutor(0, cur, probe, ""))))
	dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.YesVote, spid))
	println("SafeExecutor rejects the approved path:", !dao.ExecuteOrRejectProposal(cross(cur), spid))
	println("and its callback never ran:", !seen.ran)
	println("denial reason names the allowlist:",
		strings.Contains(dao.Render(cross(cur), spid.String()), "execution only allowed by validated govDAOs"))

	// 6. The reachability precondition is now closed at the primitive. A
	// realm-authored callback is stuck with its own current realm (the one
	// value a forger never controlled), and it can no longer borrow the
	// proxy as Previous() either -- so gates keying on Previous(), which
	// were the forgeable ones, hold.
	testing.SetOriginCaller(outside)
	testing.SetRealm(testing.NewUserRealm(outside))
	seen = frame{}
	println("forged Execute refused again:",
		dao.NewSimpleExecutor(0, cur, probe, "").Execute(cross(cur)) != nil)
	println("no proxy identity was borrowed:", seen.prev != "gno.land/r/gov/dao")

	// The member store was never reachable this way -- it gates on PkgPath(),
	// which the forgery does not control -- but the attempt is now refused one
	// layer earlier, at Execute, so the abort below never happens.
	e := dao.NewSimpleExecutor(0, cur, func(cur realm) error {
		memberstore.Get(0, cur).SetMember(memberstore.T1, outside, memberstore.NewMember(3))
		return nil
	}, "")
	println("member-store escalation refused at Execute:", e.Execute(cross(cur)) != nil)
}

// Output:
// approved path executes: true
// callback ran: true
// approved frame cur : gno.land/r/test/execauthority
// approved frame prev: gno.land/r/gov/dao
// forger cannot even open a proposal: true
// callback ran: false
// forged Execute is refused: execution denied: executors are only invocable by gno.land/r/gov/dao
// callback ran: false
// no frame was handed out: true
// approved prev address was the gov/dao package address: true
// no proposal was created for the forged execution: true
// InAllowedDAOs(proxy path): false
// allowedDAOs holds instead: gno.land/r/gov/dao/impl/v0
// SafeExecutor rejects the approved path: true
// and its callback never ran: true
// denial reason names the allowlist: true
// forged Execute refused again: true
// no proxy identity was borrowed: true
// member-store escalation refused at Execute: true
