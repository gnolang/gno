package grc20

import (
	"std"
	"gno.land/p/moul/authz"
)

var (
	Auth  = authz.NewWithMembers(std.DerivePkgAddr("gno.land/r/volosgov"))
	Token, ledger = NewToken("Volos Governance Token", "VLS", 6)
)

// do we want initial supply? or do we want to mint it later?
func Init(initialSupply int64) {
	if Token != nil {
		panic("already initialized")
	}
	if initialSupply > 0 {
		must(LEDGER.Mint(std.DerivePkgAddr("gno.land/r/volosgov"), initialSupply))
	}
}

// we can do both, use inital supply and then mint later
func Mint(caller std.Address, to std.Address, amount int64) {
	must(AUTH.Authority().Authorize(caller, "mint", func() error {
		return LEDGER.Mint(to, amount)
	}))
}

func Airdrop(caller std.Address, recipients []std.Address, amounts []int64) {
	must(AUTH.Authority().Authorize(caller, "airdrop", func() error {
		if len(recipients) != len(amounts) {
			panic("recipients and amounts length mismatch")
		}
		for i := 0; i < len(recipients); i++ {
			must(LEDGER.Mint(recipients[i], amounts[i]))
		}
		return nil
	}))
}

// Helper for panics on error
func must(err error) {
	if err != nil {
		panic(err.Error())
	}
} 
