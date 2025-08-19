// Package staker implements the staking and delegation logic for the Volos protocol.
//
// This contract allows users to stake VLS tokens and mint non-transferable xVLS tokens
// to any address, which represents delegated voting power. Users can choose to delegate
// to themselves or another address at the time of staking. When unstaking, xVLS is burned
// from the delegatee and the original staker enters a cooldown period. After the cooldown,
// the staker can withdraw their VLS tokens. The contract supports multiple pending unstakes
// per user, so users can initiate several unstakes with different amounts and unlock times
// without risk of losing tokens. Only this contract can mint and burn xVLS, ensuring that
// voting power is always tied to staked VLS and the chosen delegatee.
package staker

import (
	"std"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/rotree"
	"gno.land/p/moul/authz"
	"gno.land/r/volos/gov/governance"
	"gno.land/r/volos/gov/vls"
	"gno.land/r/volos/gov/xvls"
)

type UnstakeInfo struct {
	Amount    int64       // Amount of VLS to be withdrawn after cooldown
	UnlockAt  int64       // Unix timestamp when withdrawal is allowed
	Delegatee std.Address // Address from which xVLS was burned
}

var (
	unstakeLockPeriod = int64(7 * 24 * 60 * 60) // 7 days cooldown for unstaking
	pendingUnstakes   = avl.NewTree()           // staker address (string) -> []UnstakeInfo
	delegations       = avl.NewTree()           // staker address (string) -> avl.Tree(delegatee address (string) -> int64 amount)

	authorizer = authz.NewWithMembers(std.DerivePkgAddr("gno.land/r/volos/gov/governance"))
)

func UnstakeLockPeriod() int64 {
	return unstakeLockPeriod
}

// Delegations returns a read-only view of all delegations.
// The tree structure is: staker address (string) -> *rotree.ReadOnlyTree(delegatee address (string) -> int64 amount)
func Delegations() *rotree.ReadOnlyTree {
	makeEntrySafeFn := func(v any) any {
		delegationTree := v.(*avl.Tree)
		return rotree.Wrap(delegationTree, nil)
	}
	
	return rotree.Wrap(delegations, makeEntrySafeFn)
}

// Stake locks VLS tokens from the caller and mints an equal amount of xVLS to the specified delegatee.
// The delegatee can be the caller or any other address, allowing flexible delegation of voting power.
func Stake(cur realm, amount int64, delegatee std.Address) {
	caller := std.PreviousRealm().Address()
	if amount <= 0 {
		panic(ErrInvalidAmount)
	}

	if !delegatee.IsValid() {
		panic(ErrInvalidDelegatee)
	}

	vls.TransferFrom(cross, caller, std.CurrentRealm().Address(), amount)
	xvls.Mint(cross, delegatee, amount)
	governance.AddMember(cross, delegatee)

	updateDelegation(caller, delegatee, amount)

	emitStake(caller, caller, delegatee, amount)
	emitMemberAdded(caller, delegatee)
}

// BeginUnstake burns xVLS from the delegatee and starts a cooldown period for the original staker.
// The staker can have multiple pending unstakes, each with its own amount and unlock time.
// If the user has voted on active proposals, the unlock time is extended until the last proposal
// expires plus the standard cooldown period to prevent vote-and-exit scenarios.
// After the cooldown, the staker can withdraw the corresponding amount of VLS.
func BeginUnstake(cur realm, amount int64, delegatee std.Address) {
	caller := std.PreviousRealm().Address()
	if amount <= 0 {
		panic(ErrInvalidAmount)
	}

	if !delegatee.IsValid() {
		panic(ErrInvalidDelegatee)
	}

	if xvls.BalanceOf(delegatee) < amount {
		panic(ErrInsufficientBalance)
	}

	delegatedAmount := GetDelegatedAmount(caller, delegatee)
	if delegatedAmount < amount {
		panic(ErrInsufficientDelegation)
	}

	xvls.Burn(cross, delegatee, amount)
	if xvls.BalanceOf(delegatee) == 0 {
		governance.RemoveMember(cross, delegatee)
		emitMemberRemoved(caller, delegatee)
	}

	updateDelegation(caller, delegatee, -amount)

	key := caller.String()
	var list []UnstakeInfo
	if existing, ok := pendingUnstakes.Get(key); ok {
		list = existing.([]UnstakeInfo)
	}

	baseUnlockAt := time.Now().Unix() + unstakeLockPeriod
	unlockAt := calculateUnlockTime(delegatee, baseUnlockAt)

	list = append(list, UnstakeInfo{
		Amount:    amount,
		UnlockAt:  unlockAt,
		Delegatee: delegatee,
	})

	pendingUnstakes.Set(key, list)

	emitBeginUnstake(caller, caller, delegatee, amount, unlockAt)
}

// WithdrawUnstaked allows the original staker to withdraw all matured VLS unstakes.
// It checks all pending unstakes for the caller, and if the cooldown has passed, transfers
// the corresponding VLS back to the staker. Only matured unstakes are withdrawn; others remain pending.
func WithdrawUnstaked(cur realm) {
	caller := std.PreviousRealm().Address()
	key := caller.String()
	infoAny, ok := pendingUnstakes.Get(key)
	if !ok {
		panic(ErrCooldownNotFinished)
	}

	list := infoAny.([]UnstakeInfo)
	var (
		remaining       []UnstakeInfo
		totalToWithdraw int64 = 0
	)

	now := time.Now().Unix()
	for _, info := range list {
		if now >= info.UnlockAt {
			totalToWithdraw += info.Amount
		} else {
			remaining = append(remaining, info)
		}
	}

	if totalToWithdraw == 0 {
		panic(ErrCooldownNotFinished)
	}

	vls.Transfer(cross, caller, totalToWithdraw)

	if len(remaining) == 0 {
		pendingUnstakes.Remove(key)
	} else {
		pendingUnstakes.Set(key, remaining)
	}

	emitWithdraw(caller, caller, totalToWithdraw, len(remaining))
}

// SetUnstakeLockPeriod allows governance to change the unstake lock period.
func SetUnstakeLockPeriod(cur realm, newPeriod int64) {
	if err := authorizer.DoByPrevious("set_unstake_lock_period", func() error {
		unstakeLockPeriod = newPeriod
		return nil
	}); err != nil {
		panic(err)
	}
}

// GetDelegatedAmount returns how much a staker has delegated to a specific delegatee.
func GetDelegatedAmount(staker, delegatee std.Address) int64 {
	stakerKey := staker.String()
	delegateeKey := delegatee.String()

	stakerDelegationsAny, ok := delegations.Get(stakerKey)
	if !ok {
		return 0
	}

	stakerDelegations := stakerDelegationsAny.(*avl.Tree)
	amountAny, ok := stakerDelegations.Get(delegateeKey)
	if !ok {
		return 0
	}

	return amountAny.(int64)
}
