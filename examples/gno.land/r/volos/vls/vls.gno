// Package vls implements the Volos Governance Token for the Volos DAO on Gno.land.
//
// This contract defines a GRC20-compatible governance token (VLS) designed for use by the
// Volos DAO. Minting and burning of tokens are strictly restricted to the governance
// contract, ensuring that only authorized proposals or upgrades can affect the total supply.
// The contract supports on-chain delegation, allowing users to delegate their voting power
// to any address, including themselves. Voting power is dynamically tracked and updated
// whenever tokens are delegated, transferred, minted, or burned, ensuring that governance
// accurately reflects the current distribution of influence.
//
// The governance contract address itself can be updated by the current governance,
// providing a secure upgrade path for future changes. All privileged actions are protected
// by an authority model built on the authz package, which enforces that only the correct
// contract can execute sensitive operations. Error messages are defined in a separate
// errors.gno file for consistency and maintainability. Users must explicitly delegate to
// themselves if they wish to use their own tokens for voting, as voting power is the sum
// of tokens delegated to an address.
package vls

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/authz"
)

var (
	VolosDAO        = "gno.land/r/volosgov"
	VolosDAOAddress = std.DerivePkgAddr(VolosDAO)
	Token, ledger   = grc20.NewToken("Volos Governance Token", "VLS", 6)

	Auth = authz.NewWithAuthority(
		authz.NewContractAuthority(
			VolosDAO,
			func(title string, action authz.PrivilegedAction) error {
				// only allow the governance contract to execute privileged actions
				if std.CurrentRealm().Address() != VolosDAOAddress {
					return ErrOnlyGovernanceCanExecute
				}
				return action()
			},
		),
	)
)

var (
	delegations = avl.NewTree() // delegations: delegator address (string) -> delegatee address (string)
	votingPower = avl.NewTree() // votingPower: delegatee address (string) -> voting power (int64)
)

// Delegate assigns the caller's voting power to a delegatee address. The voting power
// is equal to the caller's token balance. If the caller had previously delegated to
// someone else, that delegation is revoked first.
func Delegate(cur realm, delegatee std.Address) {
	delegator := std.CurrentRealm().Address()
	delegatorStr := delegator.String()
	oldDelegatee, _ := delegations.Get(delegatorStr)
	amount := Token.BalanceOf(delegator)

	if oldDelegatee != nil && oldDelegatee.(string) != "" {
		oldVotes, _ := votingPower.Get(oldDelegatee.(string))
		newVotes := oldVotes.(int64) - amount
		votingPower.Set(oldDelegatee.(string), newVotes)
		emitVotingPowerChanged(oldDelegatee.(string), newVotes)
	}

	delegations.Set(delegatorStr, delegatee.String())
	newVotes := amount
	if prevVotes, found := votingPower.Get(delegatee.String()); found {
		newVotes += prevVotes.(int64)
	}

	votingPower.Set(delegatee.String(), newVotes)
	var oldDel string
	if oldDelegatee == nil {
		oldDel = ""
	} else {
		oldDel = oldDelegatee.(string)
	}

	emitDelegateChanged(delegatorStr, oldDel, delegatee.String())
	emitVotingPowerChanged(delegatee.String(), newVotes)
}

func GetVotingPower(addr std.Address) int64 {
	votes, found := votingPower.Get(addr.String())
	if !found {
		return 0
	}
	return votes.(int64)
}

func Name() string {
	return Token.GetName()
}

func Symbol() string {
	return Token.GetSymbol()
}

func Decimals() int {
	return Token.GetDecimals()
}

func TotalSupply() int64 {
	return Token.TotalSupply()
}

func BalanceOf(addr std.Address) int64 {
	return Token.BalanceOf(addr)
}

func Allowance(owner, spender std.Address) int64 {
	return Token.Allowance(owner, spender)
}

// Transfer tokens from caller to recipient. The voting power of the caller and the recipient
// is updated accordingly.
func Transfer(cur realm, to std.Address, amount int64) {
	teller := Token.RealmTeller()
	if err := teller.Transfer(to, amount); err != nil {
		panic(err)
	}
	updateVotingPowerOnBalanceChange(std.CurrentRealm().Address(), -amount)
	updateVotingPowerOnBalanceChange(to, amount)
}

// Approve a spender to spend tokens on behalf of the caller.
func Approve(cur realm, spender std.Address, amount int64) {
	teller := Token.RealmTeller()
	if err := teller.Approve(spender, amount); err != nil {
		panic(err)
	}
}

// Transfer tokens from an owner to a recipient, using the caller's allowance. The voting power
// of the owner and the recipient is updated accordingly.
func TransferFrom(cur realm, owner, to std.Address, amount int64) {
	teller := Token.RealmTeller()
	if err := teller.TransferFrom(owner, to, amount); err != nil {
		panic(err)
	}
	updateVotingPowerOnBalanceChange(owner, -amount)
	updateVotingPowerOnBalanceChange(to, amount)
}

// Public wrapper functions that handle authorization
func Mint(cur realm, VolosDAO std.Address, to std.Address, amount int64) {
	if err := Auth.Authority().Authorize(VolosDAO, "mint", func() error {
		return mint(to, amount)
	}); err != nil {
		panic(err)
	}
}

func Burn(cur realm, VolosDAO std.Address, from std.Address, amount int64) {
	if err := Auth.Authority().Authorize(VolosDAO, "burn", func() error {
		return burn(from, amount)
	}); err != nil {
		panic(err)
	}
}

func UpdateGovernance(cur realm, newAddr std.Address) {
	if err := Auth.Authority().Authorize(VolosDAOAddress, "update_governance", func() error {
		return updateGovernance(newAddr)
	}); err != nil {
		panic(err)
	}
}

func mint(to std.Address, amount int64) error {
	if err := ledger.Mint(to, amount); err != nil {
		return err
	}
	updateVotingPowerOnBalanceChange(to, amount)
	return nil
}

func burn(from std.Address, amount int64) error {
	if err := ledger.Burn(from, amount); err != nil {
		return err
	}
	updateVotingPowerOnBalanceChange(from, -amount)
	return nil
}

func updateGovernance(newAddr std.Address) error {
	if std.CurrentRealm().Address() != VolosDAOAddress {
		return ErrOnlyGovernanceCanUpdate
	}
	VolosDAOAddress = newAddr
	VolosDAO = newAddr.String()
	Auth = authz.NewWithAuthority(
		authz.NewContractAuthority(
			newAddr.String(),
			func(title string, action authz.PrivilegedAction) error {
				if std.CurrentRealm().Address() != newAddr {
					return ErrOnlyGovernanceCanExecute
				}
				emitGovernanceUpdated(VolosDAOAddress.String(), newAddr.String())
				return action()
			},
		),
	)
	return nil
}

// Call this after every transfer/mint/burn to keep voting power in sync
func updateVotingPowerOnBalanceChange(addr std.Address, delta int64) {
	delegatee, found := delegations.Get(addr.String())
	if !found || delegatee.(string) == "" {
		return
	}
	oldVotes, _ := votingPower.Get(delegatee.(string))
	newVotes := oldVotes.(int64) + delta
	votingPower.Set(delegatee.(string), newVotes)
	emitVotingPowerChanged(delegatee.(string), newVotes)
}
