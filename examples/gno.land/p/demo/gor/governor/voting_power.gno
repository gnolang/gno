package governor

import (
	"gno.land/r/demo/groups"
	"std"
)

// The interface supports various types of tokens that can be utilized for governance purposes, including GRC20, GRC721, and others.
// Tokens are converted into voting power through a conversion factor that takes into account the number
// of shares per token, while delegates are responsible for casting votes on behalf of token holders.

// The governance system can be seamlessly integrated with various underlying systems, such as PoS, PoC, and others,
// while maintaining transparency throughout the process.
type VotingPower interface {
	Total() int32
	GetVotingPower(std.Address) int32
}

type NoCoinAdapter struct {
	groupID *groups.GroupID // where voting power comes from
}

func NewNoCoinAdapter(gid *groups.GroupID) *NoCoinAdapter {
	return &NoCoinAdapter{
		groupID: gid,
	}
}

var _ VotingPower = (*NoCoinAdapter)(nil)

func (nca *NoCoinAdapter) Total() int32 {
	return groups.GetMemberCount(nca.groupID)
}

func (nca *NoCoinAdapter) GetVotingPower(account std.Address) int32 {
	// TODO: return certain number from committee
	// GetMemberWeight from r/demo/group
	// TODO: mid and addr to be consistent
	return 100
}

var _ VotingPower = (*NoCoinAdapter)(nil)

// -----------------------------------------------------
//type NativeCoinAdapter struct {
//	denom  string // determines kind of token being used
//	banker std.Banker
//}
//
//func NewNativeCoinAdapter(denom string) *NativeCoinAdapter {
//	return &NativeCoinAdapter{
//		denom:  denom,
//		banker: std.GetBanker(std.BankerTypeRealmSend),
//	}
//}
//
//func (nca *NativeCoinAdapter) Total() int32 {
//	return nca.banker.TotalCoin(nca.denom)
//}
//
//func (nca *NativeCoinAdapter) GetVotingPower(account std.Address) int32 {
//	return nca.balance(account)
//}
//
//func (nca *NativeCoinAdapter) balance(owner std.Address) uint64 {
//	coins := nca.banker.GetCoins(owner)
//	for _, c := range coins {
//		if c.Denom == nca.denom {
//			return uint64(coins[0].Amount)
//		}
//	}
//	return 0
//}
//
//// -----------------------------------------------
//type Grc20Adapter struct {
//	adminToken *grc20.AdminToken
//}
//
//func NewGrc20Adapter(name string, symbol string, decimals uint64) *Grc20Adapter {
//	return &Grc20Adapter{
//		adminToken: grc20.NewAdminToken(name, symbol, 4),
//	}
//}
//
//var _ VotingPower = (*Grc20Adapter)(nil)
//
//func (nca *Grc20Adapter) Total() uint64 {
//	return nca.adminToken.TotalSupply()
//}
//
//func (ga *Grc20Adapter) GetVotingPower(account std.Address) uint64 {
//	return ga.balance(account)
//}
//
//func (ga *Grc20Adapter) balance(addr std.Address) uint64 {
//	balance, err := ga.adminToken.BalanceOf(addr)
//	if err != nil {
//		return 0
//	}
//	return balance
//}
