package init

import (
	"chain/runtime"

	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/impl"
	"gno.land/r/gov/dao/v3/memberstore"
)

func Init() {
	assertIsDevChain()

	// This is needed because state is saved between unit tests,
	// and we want to avoid having real members used on tests
	memberstore.Get().DeleteAll()
	dao.UpdateImpl(cross, dao.UpdateRequest{
		DAO:         impl.NewGovDAO(),
		AllowedDAOs: []string{"gno.land/r/gov/dao/v3/impl"},
	})
}

func InitWithUsers(addrs ...address) {
	assertIsDevChain()

	// This is needed because state is saved between unit tests,
	// and we want to avoid having real members used on tests
	memberstore.Get().DeleteAll()
	memberstore.Get().SetTier(memberstore.T1)
	for _, a := range addrs {
		if !a.IsValid() {
			panic("invalid address: " + a.String())
		}
		memberstore.Get().SetMember(memberstore.T1, a, memberstore.NewMember(3))
	}

	dao.UpdateImpl(cross, dao.UpdateRequest{
		DAO:         impl.NewGovDAO(),
		AllowedDAOs: []string{"gno.land/r/gov/dao/v3/impl"},
	})
}

func assertIsDevChain() {
	chainID := runtime.ChainID()
	if chainID != "dev" && chainID != "tendermint_test" {
		panic("unauthorized")
	}
}
