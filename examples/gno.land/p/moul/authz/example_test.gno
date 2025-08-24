package authz_test

import (
	"std"
	"time"

	"gno.land/p/moul/addrset"
	"gno.land/p/moul/authz"
)

// Membership is a one-method interface{ Has(std.Address) bool } that provides
// the simplest authorization mechanism for this package.
func ExampleMembership_basic() {
	var (
		users        authz.Membership = addrset.New("g1user1", "g1user2")
		tokenHolders authz.Membership = authz.NewTokenHolderMembership(std.Coins{{"ugnot", 100}})
	)
	if users.Has("g1user1") {
		println("g1user1 is a user")
	}
	if tokenHolders.Has(std.CurrentRealm().Address()) {
		println("current realm is a token holder")
	}

}

// Membership supports composition.
func ExampleMembership_complex() {
	var (
		users              authz.Membership = addrset.New("g1user1", "g1user2")
		tokenHolders       authz.Membership = authz.NewTokenHolderMembership(std.Coins{{"ugnot", 100}})
		evenHourConstraint authz.Membership = authz.NewCustomMembership(func(_ std.Address) bool {
			return time.Now().Hour()%2 == 0
		})
		blacklist authz.Membership = addrset.New("g1user3")
		composite authz.Membership = authz.And( // (users OR tokenHolders) AND evenHour AND NOT blacklist
			authz.Or(users, tokenHolders),
			evenHourConstraint,
			authz.Not(blacklist),
		)
	)
	if users.Has("g1user1") {
		println("g1user1 is a user")
	}
	if tokenHolders.Has(std.CurrentRealm().Address()) {
		println("current realm is a token holder")
	}
	if composite.Has(std.PreviousRealm().Address()) {
		println("previous realm meets composite criteria")
	}
}

// Authority is a one-method interface{ Can(std.Address, string) bool } that
// provides a more robust and complete authorization solution.
//
// It is intended to
func ExampleAuthority_basic() {
	// init
	holders := authz.NewTokenHolderMembership(std.Coins{{"ugnot", 100}})
	Auth, priv := authz.New() // Auth is a safe-object that can be exposed.
	priv.Add(authz.Admin, "g1admin")
	priv.Add("propose", "g1user1", "g1user2")
	priv.Add("propose", holders)

	// usage
	if Auth.Can("g1user1", "propose") {
		println("g1user1 can propose")
	}
	if Auth.CurrentCan("vote") {
		println("current can vote")
	}
	if Auth.PreviousCan("vote") {
		println("previous can vote")
	}
	Auth.AssertCurrentCan("propose")
	println("current can propose")
	Auth.AssertPreviousCan("vote")
	println("previous can vote")
	Auth.AssertCan("g1user2", "propose")
	println("g1user2 can propose")
}

func ExampleAuthority_composition() {
	// init
	var (
		martialLaw bool = false
		govdao     govdaoType     // mock instance
		policedao  policedaoType // mock instance
		worxdao    worxdaoType   // mock instance
		authz_grc20 authz_grc20Type // mock instance
		// simple memberships
		admins             authz.Membership = addrset.New("g1admin")
		govdaot1, govdaot2 authz.Membership = govdao.T1(), govdao.T2()
		govdaot1t2         authz.Membership = authz.Or(govdaot1, govdaot2)
		policedaoMembers   authz.Membership = &policedao.Members
		ugnotHolders       authz.Membership = authz.NewTokenHolderMembership(std.Coins{{"ugnot", 100}})
		wugnotHolders      authz.Membership = authz_grc20.NewGrc20HolderMembership("wugnot", 1_000)
		// custom memberships
		seniorityConstraint authz.Membership = authz.NewCustomMembership(func(addr std.Address) bool {
			joiningTime := govdao.GetJoiningTime(addr)
			return time.Now().Sub(joiningTime) > 7*24*time.Hour
		})
		activityConstraint authz.Membership = authz.NewCustomMembership(func(addr std.Address) bool {
			return worxdao.PointsForTheLast6Months(addr) > 50
		})
		isMartialLaw authz.Membership = authz.NewCustomMembership(func(addr std.Address) bool {
			return martialLaw
		})
		// compositions
		holders   authz.Membership = authz.Or(ugnotHolders, wugnotHolders)
		proposers authz.Membership = authz.Or(admins, holders)
		vetoers   authz.Membership = authz.Or(govdaot1, policedaoMembers)
		voters    authz.Membership = authz.Or(
			authz.And(holders, authz.Not(isMartialLaw)),
			authz.And(admins, isMartialLaw),
		)
		superadmins authz.Membership = authz.And(govdaot1t2, holders, activityConstraint, seniorityConstraint)
	)
	Auth, priv := authz.New()                 // returns authz.Authority, authz.PrivateConfig
	priv.Add("like", authz.Anyone)            // Add appends or create.
	priv.Set("propose", proposers)            // Set replaces or create.
	priv.Add("comment", proposers)            // first Add creates,
	priv.Add("comment", "g1user1", "g1user2") // second Add appends.
	priv.Add("veto", vetoers)                 // govdaot1 OR policedao
	priv.Set("vote", voters)                  // holders AND NOT isMartialLaw OR admins AND isMartialLaw
	priv.Add("direct_execute", admins)        // admins
	priv.Add(authz.Admin, superadmins)        // authz.Admin is a special membership that allows all actions
	priv.Del("dontcareanymore")

	// Auth usage
	ok := Auth.Can("g1user1", "propose")
	ok = Auth.PreviousCan("vote")
	ok = Auth.CurrentCan("foo")
	Auth.AssertCurrentCan("propose")
	Auth.AssertPreviousCan("vote")
	Auth.AssertCan("g1user2", "bar")
}

func ExampleActionable_dao() {
	// XXX: improve usage
	auth, priv := authz.New()
	priv.Add("propose", "g1user1", "g1user2")
	priv.Add("execute", std.DerivePkgAddr("gno.land/r/dao/executor"))

	// TODO: implement NewContractActionable
	_ = auth.MembershipFor("propose")
	_ = auth.MembershipFor("execute")
	
	// contractBasedActioner := NewContractActionable(
	//	auth.MembershipFor("propose"),
	//	auth.MembershipFor("execute"),
	//)
	
	//contractBasedActioner.ProposeByPrevious("do_this", func() error {
	//	println("execute")
	//	return nil
	//})
}

// mocks
type govdaoType struct{} // replaces "import gno.land/r/gov/dao"

func (g govdaoType) T1() *addrset.Set                           { return &addrset.Set{} } // TODO: implement properly
func (g govdaoType) T2() *addrset.Set                           { return &addrset.Set{} } // TODO: implement properly
func (g govdaoType) GetJoiningTime(addr std.Address) time.Time { return time.Now() }

type policedaoType struct {
	Members addrset.Set
}

type worxdaoType struct{}

func (w worxdaoType) PointsForTheLast6Months(addr std.Address) int { return 100 }

type authz_grc20Type struct{}

func (a authz_grc20Type) NewGrc20HolderMembership(token string, amount int) authz.Membership {
	return authz.NewCustomMembership(func(addr std.Address) bool { return true }) // TODO: implement properly
}

type readonlyMembershipFromPBAC struct {
	pbac   *authz.PBAC
	action string
}

func (romw *readonlyMembershipFromPBAC) Has(addr std.Address) bool {
	return romw.pbac.Can(addr, romw.action)
}
