package gnoxchange

import (
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
		return "", ErrIdenticalTokens
	}

	tokenAInstance := tokenhub.GetToken(tokenA)
	if tokenAInstance == nil {
		return "", ErrTokenNotFound(tokenA)
	}

	tokenBInstance := tokenhub.GetToken(tokenB)
	if tokenBInstance == nil {
		return "", ErrTokenNotFound(tokenB)
	}

	decimalsA := tokenAInstance.GetDecimals()
	decimalsB := tokenBInstance.GetDecimals()

	if decimalsA > 6 {
		return "", ErrTokenDecimalsExceedRange(tokenA, decimalsA)
	}
	if decimalsB > 6 {
		return "", ErrTokenDecimalsExceedRange(tokenB, decimalsB)
	}

	minDecimals := decimalsA
	if decimalsB < decimalsA {
		minDecimals = decimalsB
	}

	if decimalsA > minDecimals {
		expectedScale := uint64(math.Pow10(int(decimalsA - minDecimals)))
		if initialAmountA%expectedScale != 0 {
			return "", ErrInvalidInputAmount(tokenA, expectedScale, decimalsA, minDecimals)
		}
	}

	if decimalsB > minDecimals {
		expectedScale := uint64(math.Pow10(int(decimalsB - minDecimals)))
		if initialAmountB%expectedScale != 0 {
			return "", ErrInvalidInputAmount(tokenB, expectedScale, decimalsB, minDecimals)
		}
	}

	poolKey := createPoolKey(tokenA, tokenB)
	if _, exists := poolRegistry.pools.Get(poolKey); exists {
		return "", ErrPoolAlreadyExists
	}

	allowanceA := tokenAInstance.Allowance(caller, std.CurrentRealm().Address())
	if allowanceA < initialAmountA {
		return "", ErrInsufficientAllowance(tokenA, allowanceA, initialAmountA)
	}

	allowanceB := tokenBInstance.Allowance(caller, std.CurrentRealm().Address())
	if allowanceB < initialAmountB {
		return "", ErrInsufficientAllowance(tokenB, allowanceB, initialAmountB)
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
		return ErrPoolNotFound
	}

	return pool.(*Pool).addLiquidity(std.PreviousRealm().Address(), amountA, amountB)
}

// WithdrawLiquidity is a public function that allows the caller to withdraw liquidity from the pool.
// Poolkey is constucted by concatenating the two tokens in alphabetical order with a colon between them.
func WithdrawLiquidity(poolKey string, lpAmount uint64) (uint64, uint64, error) {
	poolInterface, exists := poolRegistry.pools.Get(poolKey)
	if !exists {
		return 0, 0, ErrPoolNotFound
	}

	pool, ok := poolInterface.(*Pool)
	if !ok {
		return 0, 0, ErrInvalidPoolType
	}

	if lpAmount == 0 {
		return 0, 0, ErrCannotWithdrawZero
	}

	caller := std.PreviousRealm().Address()
	lpBalance := pool.GetLPBalance(caller)
	if lpBalance < lpAmount {
		return 0, 0, ErrInsufficientLPBalance
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

	if decimalsA > 6 {
		return ErrTokenDecimalsExceedRange(p.tokenA, decimalsA)
	}
	if decimalsB > 6 {
		return ErrTokenDecimalsExceedRange(p.tokenB, decimalsB)
	}

	minDecimals := decimalsA
	if decimalsB < decimalsA {
		minDecimals = decimalsB
	}

	if decimalsA > minDecimals {
		expectedScale := uint64(math.Pow10(int(decimalsA - minDecimals)))
		if amountA%expectedScale != 0 {
			return ErrInvalidInputAmount(p.tokenA, expectedScale, decimalsA, minDecimals)
		}
	}

	if decimalsB > minDecimals {
		expectedScale := uint64(math.Pow10(int(decimalsB - minDecimals)))
		if amountB%expectedScale != 0 {
			return ErrInvalidInputAmount(p.tokenB, expectedScale, decimalsB, minDecimals)
		}
	}

	allowanceA := tokenAInstance.Allowance(provider, std.CurrentRealm().Address())
	if allowanceA < amountA {
		return ErrInsufficientAllowance(p.tokenA, allowanceA, amountA)
	}

	allowanceB := tokenBInstance.Allowance(provider, std.CurrentRealm().Address())
	if allowanceB < amountB {
		return ErrInsufficientAllowance(p.tokenB, allowanceB, amountB)
	}

	tellerA := tokenAInstance.RealmTeller()
	if err := tellerA.TransferFrom(
		provider,
		std.CurrentRealm().Address(),
		amountA,
	); err != nil {
		return ErrFailedTokenTransfer(p.tokenA, err)
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
		return ErrFailedTokenTransfer(p.tokenB, err)
	}

	var shares uint64
	if p.totalSupplyLp == 0 {
		if amountA == 0 || amountB == 0 {
			return ErrInsufficientLiquidity
		}
		shares = uint64(math.Sqrt(float64(amountA * amountB)))
	} else {
		if (amountA * p.reserveB) != (amountB * p.reserveA) {
			return ErrIncorrectTokenRatio(p.reserveA, p.reserveB)
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
		return ErrInsufficientLiquidity
	}

	p.reserveA += amountA
	p.reserveB += amountB
	p.totalSupplyLp += shares

	if err := p.lpLedger.Mint(provider, shares); err != nil {
		return ErrFailedMintLPTokens(err)
	}

	return nil
}

// withdrawLiquidity is a helper function that burns LP tokens and updates the pool state
// if the tokens to be withdrawn fail to transfer back to the caller, the pool state is still updated
// and the caller's LPs are burned. This ensures that the pool state is consistent even if the transfer fails,
// but the caller loses their LPs.
func (p *Pool) withdrawLiquidity(caller std.Address, lpAmount uint64) (uint64, uint64, error) {
	if lpAmount == 0 {
		return 0, 0, ErrCannotWithdrawZero
	}

	amountA := (lpAmount * p.reserveA) / p.totalSupplyLp
	amountB := (lpAmount * p.reserveB) / p.totalSupplyLp

	if err := p.lpLedger.Burn(caller, lpAmount); err != nil {
		return 0, 0, ErrFailedTokenTransfer(p.tokenA, err)
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
		return ErrFailedTokenTransfer(tokenA, err)
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
		return ErrFailedTokenTransfer(tokenB, err)
	}

	lpName := ufmt.Sprintf("LP %s-%s", tokenAInstance.GetName(), tokenBInstance.GetName())
	lpSymbol := ufmt.Sprintf("LP-%s-%s", tokenAInstance.GetSymbol(), tokenBInstance.GetSymbol())
	lpToken, lpLedger := grc20.NewToken(lpName, lpSymbol, 6)
	tokenhub.RegisterToken(lpToken.Getter(), lpSymbol)

	shares := uint64(math.Sqrt(float64(initialAmountA * initialAmountB)))
	if shares == 0 {
		return ErrInsufficientInitialLiquidity
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
		return ErrFailedMintLPTokens(err)
	}

	poolRegistry.pools.Set(createPoolKey(tokenA, tokenB), pool)

	return nil
}
