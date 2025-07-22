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
	"gno.land/r/volos/vls"
	"gno.land/r/volos/xvls"
)

type UnstakeInfo struct {
	Amount    int64      // Amount of VLS to be withdrawn after cooldown
	UnlockAt  int64      // Unix timestamp when withdrawal is allowed
	Delegatee std.Address // Address from which xVLS was burned
}

var (
	UnstakeLockPeriod = int64(7 * 24 * 60 * 60) // 7 days cooldown for unstaking
	pendingUnstakes = avl.NewTree()  // staker address (string) -> []UnstakeInfo
)

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
}

// BeginUnstake burns xVLS from the delegatee and starts a cooldown period for the original staker.
// The staker can have multiple pending unstakes, each with its own amount and unlock time.
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

	xvls.Burn(cross, delegatee, amount)
	key := caller.String()
	var list []UnstakeInfo
	if existing, ok := pendingUnstakes.Get(key); ok {
		list = existing.([]UnstakeInfo)
	}

	list = append(list, UnstakeInfo{
		Amount:    amount,
		UnlockAt:  time.Now().Unix() + UnstakeLockPeriod,
		Delegatee: delegatee,
	})

	pendingUnstakes.Set(key, list)
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
		remaining []UnstakeInfo
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
} 
