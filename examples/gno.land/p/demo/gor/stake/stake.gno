package stake

import (
	"gno.land/p/demo/avl"
	"gno.land/p/demo/grc/grc20"
	std "std"
)

// The interface supports various types of tokens that can be utilized for governance purposes, including GRC20, GRC721, and others.
// Tokens are converted into voting power through a conversion factor that takes into account the number
// of shares per token, while delegates are responsible for casting votes on behalf of token holders.

// The governance system can be seamlessly integrated with various underlying systems, such as PoS, PoC, and others,
// while maintaining transparency throughout the process
// TODO: move these to distribution package?
type GovToken interface {
	Mint(to std.Address, amount uint64) error // TODO: leave here for test
	Transfer(from std.Address, to std.Address, amount uint64) error
	TotalSupply() uint64 // native coins using int64 for now, should be converted into uint64, to be consistent with grc20
	Balance(addr std.Address) uint64
	GetDenom() string
}

var _ GovToken = (*NativeCoinAdapter)(nil)

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

func (nca *NativeCoinAdapter) Mint(to std.Address, amount uint64) error {
	nca.banker.IssueCoin(to, nca.denom, int64(amount)) // safe, native coins has upper bound of int64
	return nil
}

func (nca *NativeCoinAdapter) Transfer(from std.Address, to std.Address, amount uint64) error {
	coins := std.Coins{std.Coin{Denom: nca.denom, Amount: int64(amount)}}
	nca.banker.SendCoins(from, to, coins)
	return nil
}

func (nca *NativeCoinAdapter) TotalSupply() uint64 {
	return uint64(nca.banker.TotalCoin(nca.denom))
}

func (nca *NativeCoinAdapter) Balance(owner std.Address) uint64 {
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

type GrcTokenAdapter struct {
	adminToken *grc20.AdminToken
}

func NewGrc20Adapter(name string, symbol string, decimals uint64) *GrcTokenAdapter {
	return &GrcTokenAdapter{
		adminToken: grc20.NewAdminToken(name, symbol, 4),
	}
}

func (ga *GrcTokenAdapter) Mint(to std.Address, amount uint64) error {
	return ga.adminToken.Mint(to, amount)
}

func (ga *GrcTokenAdapter) Transfer(from std.Address, to std.Address, amount uint64) error {
	return ga.adminToken.Transfer(from, to, amount)
}

func (ga *GrcTokenAdapter) TotalSupply() uint64 {
	return ga.adminToken.TotalSupply()
}

func (ga *GrcTokenAdapter) Balance(addr std.Address) uint64 {
	balance, err := ga.adminToken.BalanceOf(addr)
	if err != nil {
		return 0
	}
	return balance
}

func (ga *GrcTokenAdapter) GetDenom() string {
	return ga.adminToken.GetSymbol()
}

// TODO : implement a Grc721 adapter

type StakeOptions struct {
	BondingLockDuration   int64 // time wait for a stake done
	UnbondingLockDuration int64 // time wait for a unbond done
}

func DefaultOptions() *StakeOptions {
	return &StakeOptions{
		UnbondingLockDuration: 21,
		BondingLockDuration:   0,
	}
}

// TODO: param getters

type Stake struct {
	GovToken                    // token staking
	delegations       *avl.Tree // delegatorAddr => DelegatePair
	delegates         *avl.Tree // addr => Delegate, to keep a map of Delegates, i.e, Validators
	unbondDelegations *avl.Tree // delegatorAddr =>  UnbondDelegatePair
}

func NewStake(token GovToken) *Stake {
	stake := &Stake{
		GovToken:          token,
		delegations:       avl.NewTree(),
		delegates:         avl.NewTree(),
		unbondDelegations: avl.NewTree(),
	}
	return stake
}

func (s *Stake) getTotalSupply() uint64 {
	return s.GovToken.TotalSupply()
}

func (s *Stake) SetDelegates(ds *avl.Tree) {
	s.delegates = ds
}
