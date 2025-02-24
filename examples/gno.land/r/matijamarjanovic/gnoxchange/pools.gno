package gnoxchange

import (
	"errors"
	"math"
	"std"

	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ufmt"
	"gno.land/r/matijamarjanovic/tokenhub"
)

var (
	poolPager *pager.Pager
)

// CreatePool is the public function that handles validation and security checks before creating a pool.
// Since it is possible for tokens to have different decimals, it is the lowest amount of both tokens
// is the lowest ammount of the one with less decimals.
// If the pool creation is successful, the pool key is returned.
func CreatePool(tokenA, tokenB string, initialAmountA, initialAmountB uint64) (string, error) {
	caller := std.PreviousRealm().Address()

	if tokenA == tokenB {
		return "", errors.New("identical tokens")
	}

	tokenAInstance := tokenhub.GetToken(tokenA)
	if tokenAInstance == nil {
		return "", errors.New(ufmt.Sprintf("token %s not found", tokenA))
	}

	tokenBInstance := tokenhub.GetToken(tokenB)
	if tokenBInstance == nil {
		return "", errors.New(ufmt.Sprintf("token %s not found", tokenB))
	}

	decimalsA := tokenAInstance.GetDecimals()
	decimalsB := tokenBInstance.GetDecimals()

	minDecimals := decimalsA
	if decimalsB < decimalsA {
		minDecimals = decimalsB
	}

	if decimalsA > minDecimals {
		expectedScale := uint64(math.Pow10(int(decimalsA - minDecimals)))
		if initialAmountA%expectedScale != 0 {
			return "", errors.New(ufmt.Sprintf(
				"invalid amount for %s: amount must be in units of %d (token has %d decimals vs %d decimals for the other token)",
				tokenA, expectedScale, decimalsA, minDecimals,
			))
		}
	}

	if decimalsB > minDecimals {
		expectedScale := uint64(math.Pow10(int(decimalsB - minDecimals)))
		if initialAmountB%expectedScale != 0 {
			return "", errors.New(ufmt.Sprintf(
				"invalid amount for %s: amount must be in units of %d (token has %d decimals vs %d decimals for the other token)",
				tokenB, expectedScale, decimalsB, minDecimals,
			))
		}
	}

	poolKey := createPoolKey(tokenA, tokenB)
	if _, exists := poolRegistry.pools.Get(poolKey); exists {
		return "", errors.New("pool already exists")
	}

	allowanceA := tokenAInstance.Allowance(caller, std.CurrentRealm().Address())
	if allowanceA < initialAmountA {
		return "", errors.New(ufmt.Sprintf("insufficient allowance for token %s: have %d, need %d", tokenA, allowanceA, initialAmountA))
	}

	allowanceB := tokenBInstance.Allowance(caller, std.CurrentRealm().Address())
	if allowanceB < initialAmountB {
		return "", errors.New(ufmt.Sprintf("insufficient allowance for token %s: have %d, need %d", tokenB, allowanceB, initialAmountB))
	}

	if err := createPool(caller, tokenA, tokenB, initialAmountA, initialAmountB); err != nil {
		return "", err
	}

	return poolKey, nil
}

// AddLiquidity is a public function that allows the caller to add liquidity to the pool for the given pool key.
// Poolkey is constucted by concatenating the two tokens in alphabetical order with a colon between them.
func AddLiquidity(poolKey string, amountA, amountB uint64) error {
	pool, exists := poolRegistry.pools.Get(poolKey)
	if !exists {
		return errors.New("pool not found")
	}

	return pool.(*Pool).addLiquidity(std.PreviousRealm().Address(), amountA, amountB)
}

// WithdrawLiquidity is a public function that allows the caller to withdraw liquidity from the pool.
// Poolkey is constucted by concatenating the two tokens in alphabetical order with a colon between them.
func WithdrawLiquidity(poolKey string, lpAmount uint64) (uint64, uint64, error) {
	poolInterface, exists := poolRegistry.pools.Get(poolKey)
	if !exists {
		return 0, 0, errors.New("pool not found")
	}

	pool, ok := poolInterface.(*Pool)
	if !ok {
		return 0, 0, errors.New("invalid pool type in registry")
	}

	if lpAmount == 0 {
		return 0, 0, errors.New("cannot withdraw zero liquidity")
	}

	caller := std.PreviousRealm().Address()
	lpBalance := pool.GetLPBalance(caller)
	if lpBalance < lpAmount {
		return 0, 0, errors.New("insufficient LP token balance")
	}

	return pool.withdrawLiquidity(caller, lpAmount)
}

// createPoolKey is a helper function that creates a pool key and ensures consistent ordering of tokens
func createPoolKey(tokenA, tokenB string) string {
	if tokenA < tokenB {
		return tokenA + ":" + tokenB
	}
	return tokenB + ":" + tokenA
}

// addLiquidity is a helper function that adds liquidity to the pool. After the liquidity is added, the pool state is updated.
// Liquidy cannot be added to the pool unless the caller himself gives the gnoXchange realm enough allowance to take for the amount of tokens he wants to add.
// If user gives enough allowance for one token and not the other, the first token will be returned to the caller.
// After the liquidity is added, the pool state is updated.
func (p *Pool) addLiquidity(provider std.Address, amountA, amountB uint64) error {
	tokenAInstance := tokenhub.GetToken(p.tokenA)
	tokenBInstance := tokenhub.GetToken(p.tokenB)

	decimalsA := tokenAInstance.GetDecimals()
	decimalsB := tokenBInstance.GetDecimals()

	minDecimals := decimalsA
	if decimalsB < decimalsA {
		minDecimals = decimalsB
	}

	if decimalsA > minDecimals {
		expectedScale := uint64(math.Pow10(int(decimalsA - minDecimals)))
		if amountA%expectedScale != 0 {
			return errors.New(ufmt.Sprintf(
				"invalid amount for %s: amount must be in units of %d (token has %d decimals vs %d decimals for the other token)",
				p.tokenA, expectedScale, decimalsA, minDecimals,
			))
		}
	}

	if decimalsB > minDecimals {
		expectedScale := uint64(math.Pow10(int(decimalsB - minDecimals)))
		if amountB%expectedScale != 0 {
			return errors.New(ufmt.Sprintf(
				"invalid amount for %s: amount must be in units of %d (token has %d decimals vs %d decimals for the other token)",
				p.tokenB, expectedScale, decimalsB, minDecimals,
			))
		}
	}

	allowanceA := tokenAInstance.Allowance(provider, std.CurrentRealm().Address())
	if allowanceA < amountA {
		return errors.New(ufmt.Sprintf("insufficient allowance for token %s: have %d, need %d", p.tokenA, allowanceA, amountA))
	}

	allowanceB := tokenBInstance.Allowance(provider, std.CurrentRealm().Address())
	if allowanceB < amountB {
		return errors.New(ufmt.Sprintf("insufficient allowance for token %s: have %d, need %d", p.tokenB, allowanceB, amountB))
	}

	tellerA := tokenAInstance.RealmTeller()
	if err := tellerA.TransferFrom(
		provider,
		std.CurrentRealm().Address(),
		amountA,
	); err != nil {
		return errors.New(ufmt.Sprintf("failed to transfer token %s: %v", p.tokenA, err))
	}

	tellerB := tokenBInstance.RealmTeller()
	if err := tellerB.TransferFrom(
		provider,
		std.CurrentRealm().Address(),
		amountB,
	); err != nil {
		// if second transfer fails, we need to refund the first transfer
		if refundErr := tellerA.Transfer(provider, amountA); refundErr != nil {
			// serious error - couldn't refund
			panic(ufmt.Sprintf("CRITICAL: failed to refund token %s after failed transfer of token %s: original error: %v, refund error: %v",
				p.tokenA, p.tokenB, err, refundErr))
		}
		return errors.New(ufmt.Sprintf("failed to transfer token %s: %v", p.tokenB, err))
	}

	var shares uint64
	if p.totalSupplyLp == 0 {
		if amountA == 0 || amountB == 0 {
			return errors.New("cannot add zero liquidity")
		}
		shares = uint64(math.Sqrt(float64(amountA * amountB)))
	} else {
		if (amountA * p.reserveB) != (amountB * p.reserveA) {
			return errors.New(ufmt.Sprintf("incorrect token ratio, should be %d:%d", p.reserveA, p.reserveB))
		}

		sharesA := (amountA * p.totalSupplyLp) / p.reserveA
		sharesB := (amountB * p.totalSupplyLp) / p.reserveB
		if sharesA < sharesB {
			shares = sharesA
		} else {
			shares = sharesB
		}
	}

	if shares == 0 {
		return errors.New("insufficient liquidity provided")
	}

	p.reserveA += amountA
	p.reserveB += amountB
	p.totalSupplyLp += shares

	if err := p.lpLedger.Mint(provider, shares); err != nil {
		return err
	}

	return nil
}

// withdrawLiquidity is a helper function that burns LP tokens and updates the pool state
// if the tokens to be withdrawn fail to transfer back to the caller, the pool state is still updated
// and the caller's LPs are burned. This ensures that the pool state is consistent even if the transfer fails,
// but the caller loses their LPs.
func (p *Pool) withdrawLiquidity(caller std.Address, lpAmount uint64) (uint64, uint64, error) {
	if lpAmount == 0 {
		return 0, 0, errors.New("cannot withdraw zero liquidity")
	}

	amountA := (lpAmount * p.reserveA) / p.totalSupplyLp
	amountB := (lpAmount * p.reserveB) / p.totalSupplyLp

	if err := p.lpLedger.Burn(caller, lpAmount); err != nil {
		return 0, 0, errors.New(ufmt.Sprintf("failed to burn LP tokens: %v", err))
	}

	p.reserveA = p.reserveA - amountA
	p.reserveB = p.reserveB - amountB
	p.totalSupplyLp = p.totalSupplyLp - lpAmount

	tokenAInstance := tokenhub.GetToken(p.tokenA)
	tokenBInstance := tokenhub.GetToken(p.tokenB)

	tellerA := tokenAInstance.RealmTeller()
	if err := tellerA.Transfer(caller, amountA); err != nil {
		panic(ufmt.Sprintf("CRITICAL: pool state updated but token A transfer failed: %v", err))
	}

	tellerB := tokenBInstance.RealmTeller()
	if err := tellerB.Transfer(caller, amountB); err != nil {
		panic(ufmt.Sprintf("CRITICAL: pool state updated but token B transfer failed: %v", err))
	}

	return amountA, amountB, nil
}

// createPool handles the core pool creation logic. First it takes the tokens from the caller and transfers them to the pool.
// Then it creates the LP tokens and mints them to the caller. It calculates the initial liquidity based on the sqrt of the product of the two tokens.
// The pool is then created and added to the pool registry.
func createPool(creator std.Address, tokenA, tokenB string, initialAmountA, initialAmountB uint64) error {
	tokenAInstance := tokenhub.GetToken(tokenA)
	tokenBInstance := tokenhub.GetToken(tokenB)

	tellerA := tokenAInstance.RealmTeller()
	if err := tellerA.TransferFrom(
		creator,
		std.CurrentRealm().Address(),
		initialAmountA,
	); err != nil {
		return errors.New(ufmt.Sprintf("failed to transfer token %s: %v", tokenA, err))
	}

	tellerB := tokenBInstance.RealmTeller()
	if err := tellerB.TransferFrom(
		creator,
		std.CurrentRealm().Address(),
		initialAmountB,
	); err != nil {
		if refundErr := tellerA.Transfer(creator, initialAmountA); refundErr != nil {
			panic(ufmt.Sprintf("CRITICAL: failed to refund token %s after failed transfer of token %s: original error: %v, refund error: %v",
				tokenA, tokenB, err, refundErr))
		}
		return errors.New(ufmt.Sprintf("failed to transfer token %s: %v", tokenB, err))
	}

	lpName := ufmt.Sprintf("LP %s-%s", tokenAInstance.GetName(), tokenBInstance.GetName())
	lpSymbol := ufmt.Sprintf("LP-%s-%s", tokenAInstance.GetSymbol(), tokenBInstance.GetSymbol())
	lpToken, lpLedger := grc20.NewToken(lpName, lpSymbol, 6)
	tokenhub.RegisterToken(lpToken.Getter(), lpSymbol)

	shares := uint64(math.Sqrt(float64(initialAmountA * initialAmountB)))
	if shares == 0 {
		return errors.New("insufficient initial liquidity")
	}

	pool := &Pool{
		tokenA:        tokenA,
		tokenB:        tokenB,
		reserveA:      initialAmountA,
		reserveB:      initialAmountB,
		lpToken:       lpToken,
		lpLedger:      lpLedger,
		totalSupplyLp: shares,
	}

	if err := pool.lpLedger.Mint(creator, shares); err != nil {
		return errors.New(ufmt.Sprintf("failed to mint LP tokens: %v", err))
	}

	poolRegistry.pools.Set(createPoolKey(tokenA, tokenB), pool)

	return nil
}
