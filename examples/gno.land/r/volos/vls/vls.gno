package vls

import (
	"std"
	"gno.land/p/moul/authz"
	"gno.land/p/demo/avl"
	"gno.land/p/demo/grc/grc20" 
	"gno.land/p/demo/ufmt"
)

var (
	volosDAOAddress = std.DerivePkgAddr("gno.land/r/volosgov")
	Token, ledger = grc20.NewToken("Volos Governance Token", "VLS", 6)

	Auth = authz.NewWithAuthority(
		authz.NewContractAuthority(
			"gno.land/r/volosgov",
			func(title string, action authz.PrivilegedAction) error {
				// only allow the governance contract to execute privileged actions
				if std.CurrentRealm().Address() != volosDAOAddress {
					return ufmt.Errorf("only governance contract can execute: %s", title)
				}
				return action()
			},
		),
	)
)

var (
	delegations = avl.NewTree() 	// delegations: delegator address (string) -> delegatee address (string)
	votingPower = avl.NewTree() 	// votingPower: delegatee address (string) -> voting power (int64)
)

// Delegate assigns the caller's voting power to a delegatee address. The voting power
// is equal to the caller's token balance. If the caller had previously delegated to
// someone else, that delegation is revoked first.
func Delegate(cur realm, delegatee std.Address) {
	delegator := std.PreviousRealm().Address()
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
	teller := Token.CallerTeller()
	must(teller.Transfer(to, amount))
	updateVotingPowerOnBalanceChange(std.PreviousRealm().Address(), -amount)
	updateVotingPowerOnBalanceChange(to, amount)
}

// Approve a spender to spend tokens on behalf of the caller.
func Approve(cur realm, spender std.Address, amount int64) {
	teller := Token.CallerTeller()
	must(teller.Approve(spender, amount))
}

// Transfer tokens from an owner to a recipient, using the caller's allowance. The voting power
// of the owner and the recipient is updated accordingly.
func TransferFrom(cur realm, owner, to std.Address, amount int64) {
	teller := Token.CallerTeller()
	must(teller.TransferFrom(owner, to, amount))
	updateVotingPowerOnBalanceChange(owner, -amount)
	updateVotingPowerOnBalanceChange(to, amount)
}

// Mint tokens. The voting power of the recipient is updated accordingly.
func Mint(cur realm, volosDAO std.Address, to std.Address, amount int64) {
	must(Auth.Authority().Authorize(volosDAO, "mint", func() error {
		err := ledger.Mint(to, amount)
		if err == nil {
			updateVotingPowerOnBalanceChange(to, amount)
		}
		return err
	}))
}

// Burn tokens. The voting power of the caller is updated accordingly.
func Burn(cur realm, volosDAO std.Address, from std.Address, amount int64) {
	must(Auth.Authority().Authorize(volosDAO, "burn", func() error {
		err := ledger.Burn(from, amount)
		if err == nil {
			updateVotingPowerOnBalanceChange(from, -amount)
		}
		return err
	}))
}

// Approve a spender to spend tokens on behalf of the caller.
func Approve(cur realm, volosDAO std.Address, owner, spender std.Address, amount int64) {
	must(Auth.Authority().Authorize(volosDAO, "approve", func() error {
		return ledger.Approve(owner, spender, amount)
	}))
}

// Update the governance contract address.
func UpdateGovernance(cur realm, newAddr std.Address) {
	if std.CurrentRealm().Address() != volosDAOAddress {
		panic("only current governance contract can update governance")
	}
	volosDAOAddress = newAddr
	Auth = authz.NewWithAuthority(
		authz.NewContractAuthority(
			newAddr.String(),
			func(title string, action authz.PrivilegedAction) error {
				if std.CurrentRealm().Address() != newAddr {
					return ufmt.Errorf("only governance contract can execute: %s", title)
				}
				emitGovernanceUpdated(volosDAOAddress.String(), newAddr.String())
				return action()
			},
		),
	)
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

// helper for panics on error
func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}
