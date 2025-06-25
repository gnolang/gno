package vls

import (
	"std"
	"gno.land/p/moul/authz"
	"gno.land/p/demo/avl"
)

var (
	volosDAOAddress = std.DerivePkgAddr("gno.land/r/volosgov")
	Token, ledger = NewToken("Volos Governance Token", "VLS", 6)

	Auth = authz.NewWithAuthority(
		authz.NewContractAuthority(
			"gno.land/r/volosgov",
			func(title string, action authz.PrivilegedAction) error {
				// only allow the governance contract to execute privileged actions
				if std.CurrentRealm().Address() != volosDAOAddress {
					return std.Errorf("only governance contract can execute: %s", title)
				}
				return action()
			},
		),
	)
)

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

func Burn(volosDAO std.Address, from std.Address, amount int64) {
	must(Auth.Authority().Authorize(volosDAO, "burn", func() error {
		return ledger.Burn(from, amount)
	}))
}

func Approve(volosDAO std.Address, owner, spender std.Address, amount int64) {
	must(Auth.Authority().Authorize(volosDAO, "approve", func() error {
		return ledger.Approve(owner, spender, amount)
	}))
}

// is this something we want to do? if volosDao is going to have versioning, we need this probably
func UpdateGovernance(newAddr std.Address) {
	if std.CurrentRealm().Address() != volosDAOAddress {
		panic("only current governance contract can update governance")
	}
	volosDAOAddress = newAddr
	Auth = authz.NewWithAuthority(
		authz.NewContractAuthority(
			newAddr.String(),
			func(title string, action authz.PrivilegedAction) error {
				if std.CurrentRealm().Address() != newAddr {
					return std.Errorf("only governance contract can execute: %s", title)
				}
				return action()
			},
		),
	)
}

// helper for panics on error
func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}
