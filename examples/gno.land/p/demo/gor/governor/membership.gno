package governor

import (
	"gno.land/p/demo/grc/grc20"
	"gno.land/r/demo/groups"
)

type Membership interface {
	AddMember(addr string, weight int, meta string)
	UpdateMemberWeight(mid uint64, weight int)
	TotalVotingPower() int32                   // use int32 corresponding to maths.Rat for now
	GetVotingPowerFromMember(mid uint64) int32 // this get weight, not direct tokens held
	Transfer(from std.Address, to std.Address, amount uint64) error
	RenderMembership() string
}

type NoCoinAdapter struct {
	groupID groups.GroupID // where voting power comes from
}

func NewNoCoinAdapter(gid groups.GroupID) *NoCoinAdapter {
	return &NoCoinAdapter{
		groupID: gid,
	}
}

func (nca *NoCoinAdapter) AddMember(addr string, weight int, meta string) {
	groups.AddMember(nca.groupID, addr, weight, meta)
}

func (nca *NoCoinAdapter) UpdateMemberWeight(mid uint64, weight int) {
	groups.UpdateMemberWeight(nca.groupID, mid, weight)
}

func (nca *NoCoinAdapter) TotalVotingPower() int32 {
	// here differs between different DAOs, some like multi-sig style, while some with token hold/stake style.
	return int32(groups.GetMemberCount(nca.groupID))
}

func (nca *NoCoinAdapter) GetVotingPowerFromMember(mid uint64) int32 {
	return int32(groups.GetMemberWeight(nca.groupID, mid))
}

func (nca *NoCoinAdapter) RenderMembership() string {
	return groups.RenderGroup(nca.groupID)
}

// no tokens and voting units to transfer
func (nca *NoCoinAdapter) Transfer(from std.Address, to std.Address, amount uint64) error {
	return nil
}

var _ Membership = (*NoCoinAdapter)(nil)

// ----------------------------------------------------------------
// ----------------------------------------------------------------
// ----------------------------------------------------------------

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

// TODO: also transfer voting uints, hook
func (nca *NativeCoinAdapter) Transfer(from std.Address, to std.Address, amount uint64) error {
	coins := std.Coins{std.Coin{Denom: nca.denom, Amount: int64(amount)}}
	nca.banker.SendCoins(from, to, coins)

	//

	return nil
}

func (nca *NativeCoinAdapter) TotalVotingPower() uint64 {
	return uint64(nca.banker.TotalCoin(nca.denom))
}

func (nca *NativeCoinAdapter) GetVotingPowerFromMember(owner std.Address) uint64 {
	coins := nca.banker.GetCoins(owner)
	for _, c := range coins {
		if c.Denom == nca.denom {
			return uint64(coins[0].Amount)
		}
	}
	return 0
}

func (nca *NativeCoinAdapter) GetDenom() string {
	return nca.denom
}

// ----------------------------------------------------------------
// ----------------------------------------------------------------
// ----------------------------------------------------------------
type Grc20Adapter struct {
	adminToken *grc20.AdminToken
}

func NewGrc20Adapter(name string, symbol string, decimals uint64) *GrcTokenAdapter {
	return &Grc20Adapter{
		adminToken: grc20.NewAdminToken(name, symbol, 4),
	}
}

func (ga *Grc20Adapter) Mint(to std.Address, amount uint64) error {
	return ga.adminToken.Mint(to, amount)
}

func (ga *Grc20Adapter) Transfer(from std.Address, to std.Address, amount uint64) error {
	return ga.adminToken.Transfer(from, to, amount)
}

func (ga *Grc20Adapter) TotalSupply() uint64 {
	return ga.adminToken.TotalSupply()
}

func (ga *Grc20Adapter) Balance(addr std.Address) uint64 {
	balance, err := ga.adminToken.BalanceOf(addr)
	if err != nil {
		return 0
	}
	return balance
}

func (ga *Grc20Adapter) GetDenom() string {
	return ga.adminToken.GetSymbol()
}
