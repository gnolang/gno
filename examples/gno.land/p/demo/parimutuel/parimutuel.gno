package parimutuel

// CalculatePayout returns the payout owed to a single winning bet,
// proportional to its share of the winning outcome's pool, drawn from
// the full pool (losers' stakes included).
//
// userBet must be part of winningPool (i.e. userBet <= winningPool).
// winningPool must be <= totalPool and > 0.
func CalculatePayout(userBet, totalPool, winningPool int64) int64 {
	assertNonNegative(userBet, "userBet")
	assertNonNegative(totalPool, "totalPool")
	if winningPool <= 0 {
		panic("parimutuel: winningPool must be positive")
	}
	if winningPool > totalPool {
		panic("parimutuel: winningPool cannot exceed totalPool")
	}
	if userBet > winningPool {
		panic("parimutuel: userBet cannot exceed winningPool")
	}
	return mulDiv(userBet, totalPool, winningPool)
}

// CalculatePayoutWithRake is CalculatePayout after deducting a commission
// (rake) from the total pool before distribution. rakeBps is in basis
// points: 100 = 1%, 10000 = 100%.
//
// If the rake consumes the entire pool (rakeBps == 10000, or rounding
// leaves netPool at 0), there is nothing left to distribute and this
// returns 0 rather than delegating to CalculatePayout, whose invariants
// assume a non-degenerate pool.
func CalculatePayoutWithRake(userBet, totalPool, winningPool, rakeBps int64) int64 {
	if rakeBps < 0 || rakeBps > 10000 {
		panic("parimutuel: rakeBps must be between 0 and 10000")
	}
	netPool := totalPool - mulDiv(totalPool, rakeBps, 10000)
	if netPool <= 0 {
		return 0
	}
	return CalculatePayout(userBet, netPool, winningPool)
}

// ImpliedProbabilityBps returns the market-implied probability of an
// outcome, in basis points (10000 = 100%), based on its share of the
// total pool.
//
// Basis points (not float64) are used deliberately: floating-point
// arithmetic is not guaranteed deterministic across platforms, which
// matters for consensus-critical code that must produce identical
// results on every validator.
func ImpliedProbabilityBps(outcomePool, totalPool int64) int64 {
	if totalPool <= 0 {
		panic("parimutuel: totalPool must be positive")
	}
	assertNonNegative(outcomePool, "outcomePool")
	if outcomePool > totalPool {
		panic("parimutuel: outcomePool cannot exceed totalPool")
	}
	return mulDiv(outcomePool, 10000, totalPool)
}

func assertNonNegative(v int64, name string) {
	if v < 0 {
		panic("parimutuel: " + name + " must be non-negative")
	}
}

// mulDiv computes (a * b) / c. Amounts are assumed to fit comfortably
// within int64 for realistic ugnot-denominated pools; for pools that
// could approach 2^63, callers should pre-scale inputs or use a
// bignum-backed alternative, which this package does not provide.
func mulDiv(a, b, c int64) int64 {
	if c == 0 {
		panic("parimutuel: division by zero")
	}
	return (a * b) / c
}
