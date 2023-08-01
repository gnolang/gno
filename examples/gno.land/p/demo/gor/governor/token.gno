package governor

import (
	"std"
)
import "gno.land/p/demo/grc/grc20"
import "gno.land/r/demo/groups"

// The interface supports various types of tokens that can be utilized for governance purposes, including GRC20, GRC721, and others.
// Tokens are converted into voting power through a conversion factor that takes into account the number
// of shares per token, while delegates are responsible for casting votes on behalf of token holders.

// The governance system can be seamlessly integrated with various underlying systems, such as PoS, PoC, and others,
// while maintaining transparency throughout the process
// TODO: move these to distribution package?

type VotingPower interface {
	VP(std.Address) uint64
}

type NoCoinAdapter struct {
	denom string // determines kind of token being used
	//committee   // interface get members, get size of committee, etc
	group groups.Group
}

var _ VotingPower = (*NoCoinAdapter)(nil)

func (nca *NoCoinAdapter) VP(account std.Address) uint64 {
	// TODO: return certain number from committee
	// GetWeight from r/demo/group
	return 100
}

var _ VotingPower = (*NativeCoinAdapter)(nil)

type NativeCoinAdapter struct {
	denom  string // determines kind of token being used
	banker std.Banker
}

func NewNativeCoinAdapter(denom string) *NativeCoinAdapter {
	return &NativeCoinAdapter{
		denom:  denom,
		banker: std.GetBanker(std.BankerTypeRealmSend),
	}
}

func (nca *NativeCoinAdapter) VP(account std.Address) uint64 {
	return nca.balance(account)
}

func (nca *NativeCoinAdapter) balance(owner std.Address) uint64 {
	coins := nca.banker.GetCoins(owner)
	for _, c := range coins {
		if c.Denom == nca.denom {
			return uint64(coins[0].Amount)
		}
	}
	return 0
}

// -----------------------------------------------
type Grc20Adapter struct {
	adminToken *grc20.AdminToken
}

func NewGrc20Adapter(name string, symbol string, decimals uint64) *GrcTokenAdapter {
	return &Grc20Adapter{
		adminToken: grc20.NewAdminToken(name, symbol, 4),
	}
}

var _ VotingPower = (*Grc20Adapter)(nil)

func (ga *Grc20Adapter) VP(account std.Address) uint64 {
	return ga.balance(account)
}

func (ga *Grc20Adapter) balance(addr std.Address) uint64 {
	balance, err := ga.adminToken.BalanceOf(addr)
	if err != nil {
		return 0
	}
	return balance
}
