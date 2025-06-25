package grc20

import (
	"std"
	"gno.land/p/moul/authz"
	"gno.land/p/demo/avl"
)

var (
	GovernanceAddr = std.DerivePkgAddr("gno.land/r/volosgov")
	Token *Token
	ledger *PrivateLedger

	Auth = authz.NewWithAuthority(
		authz.NewContractAuthority(
			"gno.land/r/volosgov",
			func(title string, action authz.PrivilegedAction) error {
				// only allow the governance contract to execute privileged actions
				if std.CurrentRealm().Address() != GovernanceAddr {
					return std.Errorf("only governance contract can execute: %s", title)
				}
				return action()
			},
		),
	)
)

func init() {
	Token, ledger = NewToken("Volos Governance Token", "VLS", 6)
}

// we can do both, use inital supply and then mint later
func Mint(volosDAO std.Address, to std.Address, amount int64) {
	must(Auth.Authority().Authorize(volosDAO, "mint", func() error {
		return ledger.Mint(to, amount)
	}))
}

// Airdrop using an avl.Tree (address string -> amount int64)
func Airdrop(volosDAO std.Address, airdropTree *avl.Tree) {
	must(Auth.Authority().Authorize(volosDAO, "airdrop", func() error {
		airdropped := false
		airdropTree.Iterate("", "", func(key string, value any) bool {
			addr := std.Address(key)
			amount := value.(int64)
			must(ledger.Mint(addr, amount))
			airdropped = true
			return false
		})
		if !airdropped {
			panic("airdrop tree is empty")
		}
		return nil
	}))
}

// helper for panics on error
func must(err error) {
	if err != nil {
		panic(err.Error())
	}
} 
