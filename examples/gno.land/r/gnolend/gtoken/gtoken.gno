package gtoken

import (
	"std"
	"strings"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/grc20reg"

	"gno.land/p/gnolend/consts"
	"gno.land/p/gnolend/errors"
	"gno.land/p/gnolend/math"
	"gno.land/p/gnolend/types"
)

var (
	token         *grc20.Token
	privateLedger *grc20.PrivateLedger
	UserTeller    grc20.Teller

	// Pool state
	underlyingAsset std.Address
	liquidityIndex uint64
	lastUpdateBlock int64
)

// Initialize a new gToken
func init() {
	
}

// InitToken initializes the gToken with the underlying asset
func InitToken(asset std.Address, name, symbol string) {
	if token != nil {
		panic(errors.ErrPoolAlreadyExists)
	}

	underlyingAsset = asset
	token, privateLedger = grc20.NewToken(name, symbol, 18) // Use 18 decimals for precise calculations
	UserTeller = token.CallerTeller()

	liquidityIndex = uint64(consts.PRECISION) // Start with 1.0
	lastUpdateBlock = std.GetHeight()

	// Register the token
	getter := func() *grc20.Token { return token }
	grc20reg.Register(getter, "")
}

// UpdateLiquidityIndex updates the liquidity index based on accrued interest
func UpdateLiquidityIndex(supplyRate uint64) {
	currentBlock := std.GetHeight()
	timeDelta := currentBlock - lastUpdateBlock
	if timeDelta > 0 {
		interest := math.CalculateCompoundedInterest(supplyRate, timeDelta)
		liquidityIndex = (liquidityIndex * interest) / uint64(consts.PRECISION)
		lastUpdateBlock = currentBlock
	}
}

// Mint creates new gTokens when assets are supplied
func Mint(to std.Address, amount uint64) error {
	if amount == 0 {
		return errors.ErrInvalidAmount
	}

	// Calculate gToken amount based on current liquidity index
	gTokenAmount := (amount * uint64(consts.PRECISION)) / liquidityIndex

	err := privateLedger.Mint(to, gTokenAmount)
	if err != nil {
		return err
	}

	std.Emit(types.SupplyEvent{
		User:      to,
		Asset:     underlyingAsset,
		Amount:    amount,
		Timestamp: std.GetTimestamp(),
	})

	return nil
}

// Burn destroys gTokens when assets are withdrawn
func Burn(from std.Address, gTokenAmount uint64) (uint64, error) {
	if gTokenAmount == 0 {
		return 0, errors.ErrInvalidAmount
	}

	// Calculate underlying amount based on current liquidity index
	underlyingAmount := (gTokenAmount * liquidityIndex) / uint64(consts.PRECISION)

	err := privateLedger.Burn(from, gTokenAmount)
	if err != nil {
		return 0, err
	}

	std.Emit(types.WithdrawEvent{
		User:      from,
		Asset:     underlyingAsset,
		Amount:    underlyingAmount,
		Timestamp: std.GetTimestamp(),
	})

	return underlyingAmount, nil
}

// CalculateUnderlyingAmount calculates how much underlying asset a user can withdraw
func CalculateUnderlyingAmount(gTokenAmount uint64) uint64 {
	return (gTokenAmount * liquidityIndex) / uint64(consts.PRECISION)
}

// GetLiquidityIndex returns the current liquidity index
func GetLiquidityIndex() uint64 {
	return liquidityIndex
}

// GetLastUpdateBlock returns the last block when index was updated
func GetLastUpdateBlock() int64 {
	return lastUpdateBlock
}

// Standard GRC20 interface implementations
func Transfer(to std.Address, amount uint64) error {
	return UserTeller.Transfer(to, amount)
}

func Approve(spender std.Address, amount uint64) error {
	return UserTeller.Approve(spender, amount)
}

func TransferFrom(from, to std.Address, amount uint64) error {
	return UserTeller.TransferFrom(from, to, amount)
}

func TotalSupply() uint64 {
	return token.TotalSupply()
}

func BalanceOf(owner std.Address) uint64 {
	return token.BalanceOf(owner)
}

func Allowance(owner, spender std.Address) uint64 {
	return token.Allowance(owner, spender)
}

// Render implements the rendering interface
func Render(path string) string {
	parts := strings.Split(path, "/")
	c := len(parts)

	switch {
	case path == "":
		return token.RenderHome()
	case c == 2 && parts[0] == "balance":
		owner := std.Address(parts[1])
		balance := token.BalanceOf(owner)
		return ufmt.Sprintf("%d\n", balance)
	default:
		return "404\n"
	}
} 