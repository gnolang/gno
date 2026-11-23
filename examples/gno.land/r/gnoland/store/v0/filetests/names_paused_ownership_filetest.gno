// PKGPATH: gno.land/r/test/selfstar
package selfstar

import (
	"chain/runtime/unsafe"
	"testing"

	"gno.land/p/nt/testutils/v0"
	store "gno.land/r/gnoland/store/v0"
	"gno.land/r/gov/dao"
	daoinit "gno.land/r/gov/dao/init/v0"
	"gno.land/r/sys/names"
	"gno.land/r/sys/users"
)

var c address = unsafe.OriginCaller()

func init(cur realm) {
	daoinit.InitWithUsers(cross(cur), c)
}

// A GovDAO pause of r/sys/names halts deploys; it must not make a namespace's
// owner a stranger to the store: their star on their own app still does not
// rank, and they can still claim under their name.
func main(cur realm) {
	owner := testutils.TestAddress("selfowner")
	pathA := "gno.land/r/selfowner/a"
	pathB := "gno.land/r/selfowner/b"

	testing.SetRealm(testing.NewUserRealm("g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx")) // r/sys/names admin
	names.Enable(cross(cur))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/sys/users/init"))
	if err := users.RegisterUserIgnoreCanonical(cross(cur), "selfowner", owner); err != nil {
		panic(err)
	}
	testing.SetRealm(testing.NewCodeRealm(pathA))
	store.Register(cross(cur), "own-a", "Own A", "A tagline", "utilities", 0)
	testing.SetRealm(testing.NewCodeRealm(pathB))
	store.Register(cross(cur), "own-b", "Own B", "A tagline", "utilities", 0)

	testing.SetRealm(testing.NewUserRealm(owner))
	store.Star(cross(cur), "blog") // starts the name's clock
	testing.SkipHeights(3 * 17280) // 3 days at 5 s a block

	testing.SetRealm(testing.NewUserRealm(owner))
	store.Star(cross(cur), "own-a")
	println("live badge:", store.Badge(pathA))

	testing.SetOriginCaller(c)
	testing.SetRealm(testing.NewUserRealm(c))
	pid := dao.MustCreateProposal(cross(cur), names.ProposeSetPaused(cross(cur), true))
	dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.YesVote, pid))
	dao.ExecuteProposal(cross(cur), pid)

	testing.SetOriginCaller(owner)
	testing.SetRealm(testing.NewUserRealm(owner))
	store.Star(cross(cur), "own-b")
	println("paused badge:", store.Badge(pathB))
	store.Claim(cross(cur), "gno.land/p/selfowner/lib", "own-lib", "Own Lib", "A tagline", "dev-tools", 0, "package")
	println("paused claim: ok")
}

// Output:
// live badge: ★ 0 on Explore · [Star](/r/gnoland/store/v0$help&func=Star&slug=own-a)
// paused badge: ★ 0 on Explore · [Star](/r/gnoland/store/v0$help&func=Star&slug=own-b)
// paused claim: ok
