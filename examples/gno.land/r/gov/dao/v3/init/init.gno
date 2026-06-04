package init

import (
	"chain/runtime"

	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/impl"
	"gno.land/r/gov/dao/v3/memberstore"
)

func Init(cur realm) {
	assertIsDevChain()

	// This is needed because state is saved between unit tests,
	// and we want to avoid having real members used on tests
	memberstore.Get(0, cur).DeleteAll()
	dao.UpdateImpl(cross(cur), dao.NewUpdateRequest(impl.NewGovDAO(), []string{"gno.land/r/gov/dao/v3/impl"}))
}

func InitWithUsers(cur realm, addrs ...address) {
	assertIsDevChain()

	// This is needed because state is saved between unit tests,
	// and we want to avoid having real members used on tests
	memberstore.Get(0, cur).DeleteAll()
	memberstore.Get(0, cur).SetTier(memberstore.T1)
	for _, a := range addrs {
		if !a.IsValid() {
			panic("invalid address: " + a.String())
		}
		memberstore.Get(0, cur).SetMember(memberstore.T1, a, memberstore.NewMember(3))
	}

	dao.UpdateImpl(cross(cur), dao.NewUpdateRequest(impl.NewGovDAO(), []string{"gno.land/r/gov/dao/v3/impl"}))
}

func assertIsDevChain() {
	chainID := runtime.ChainID()
	if chainID != "dev" && chainID != "tendermint_test" {
		panic("unauthorized")
	}
}
