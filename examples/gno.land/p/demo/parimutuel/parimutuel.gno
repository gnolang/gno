package parimutuel

import (
	"math/overflow"
)

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

// CalculateRake returns the commission taken from totalPool at rakeBps
// (basis points, 100 = 1%, 10000 = 100%). Exposed so a caller that
// needs to account for or display the house cut itself doesn't have to
// re-derive totalPool * rakeBps / 10000 with its own rounding.
func CalculateRake(totalPool, rakeBps int64) int64 {
	assertNonNegative(totalPool, "totalPool")
	if rakeBps < 0 || rakeBps > 10000 {
		panic("parimutuel: rakeBps must be between 0 and 10000")
	}
	return mulDiv(totalPool, rakeBps, 10000)
}

// CalculatePayoutWithRake is CalculatePayout after deducting a
// commission (rake) from the total pool before distribution.
//
// Validation is against the pre-rake totalPool, not the post-rake net
// pool -- winningPool can legitimately exceed the net pool (a rake
// simply means winners get back slightly less than their full stake),
// so that is not an error condition and does not panic. If the rake
// consumes the entire pool (rakeBps == 10000, or rounding leaves
// netPool at 0), there is nothing left to distribute and this returns
// 0 directly rather than dividing by a winningPool that may exceed it.
func CalculatePayoutWithRake(userBet, totalPool, winningPool, rakeBps int64) int64 {
	assertNonNegative(userBet, "userBet")
	rake := CalculateRake(totalPool, rakeBps) // also validates totalPool and rakeBps
	if winningPool <= 0 {
		panic("parimutuel: winningPool must be positive")
	}
	if winningPool > totalPool {
		panic("parimutuel: winningPool cannot exceed totalPool")
	}
	if userBet > winningPool {
		panic("parimutuel: userBet cannot exceed winningPool")
	}

	netPool := totalPool - rake
	if netPool <= 0 {
		return 0
	}
	return mulDiv(userBet, netPool, winningPool)
}

// ImpliedProbabilityBps returns the market-implied probability of an
// outcome, in basis points (10000 = 100%), based on its share of the
// total pool.
//
// Basis points (not float64) are used deliberately: floating-point
// arithmetic is not guaranteed deterministic across platforms, which
// matters for consensus-critical code that must produce identical
// results on every validator. Integer division also means this rounds
// down, so summing ImpliedProbabilityBps across every outcome in a
// pool can land a few basis points under 10000 (dust) rather than
// exactly 10000 -- that is expected, not a bug.
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

// mulDiv computes (a * b) / c, panicking on multiplication overflow
// rather than silently wrapping. c == 0 is not guarded here: every
// call site already establishes c > 0 before calling this, and integer
// division by zero already panics natively.
func mulDiv(a, b, c int64) int64 {
	return overflow.Mul64p(a, b) / c
}
