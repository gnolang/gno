package payrolls

import (
	"chain"
	"errors"
	"time"

	"gno.land/p/nt/seqid"
)

// XXX: should pass a read-only payroll object to those funcs

// DistribFn is a distribution function used to compute and amount of released coins at a particular point in time.
//
// It must return payrolls [Coins]
type DistribFn func(cur realm, worked time.Duration) chain.Coins

// BreakupFn is a function used to compute payroll termination bonuses.
//
// It must return payrolls [Coins]
type BreakupFn func(cur realm, elapsed time.Duration, pauseDuration time.Duration, source CallSource) chain.Coins

// DistribStep returns a distribution function that will release coins every `per` duration.
func DistribStep(cur realm, coins chain.Coins, per time.Duration) DistribFn {
	return func(cur realm, elapsed time.Duration) chain.Coins {
		blocksCount := elapsed / per
		out := make(chain.Coins, 0, len(coins))
		for _, coin := range coins {
			out = append(out, chain.NewCoin(coin.Denom, coin.Amount*int64(blocksCount)))
		}
		return out
	}
}

// DistribLinear returns a distribution function that will continuously release coins every seconds and amount to a rate of `coins` every `per` duration.
func DistribLinear(cur realm, coins chain.Coins, per time.Duration) DistribFn {
	return func(cur realm, elapsed time.Duration) chain.Coins {
		out := make(chain.Coins, 0, len(coins))
		for _, coin := range coins {
			newAmount := (int64(elapsed.Seconds()) * coin.Amount) / int64(per.Seconds())
			out = append(out, chain.NewCoin(coin.Denom, newAmount))
		}
		return out
	}
}

// DistribMonthlyStep returns a distribution function that will release `coins` every months
func DistribMonthlyStep(cur realm, amountPerMonth chain.Coins) DistribFn {
	return DistribStep(cur, amountPerMonth, time.Hour*24*30)
}

// DistribMonthlyStep returns a distribution function that will continuously release coins every seconds and amount to a rate of `coins` every months
func DistribMonthlyContinuous(cur realm, amountPerMonth chain.Coins) DistribFn {
	return DistribLinear(cur, amountPerMonth, time.Hour*24*30)
}

// CreateMonthlyContinuous creates a payroll with a continuous release at a fixed rate per months of a token identified by denom.
//
// denom must be a payrolls [Coins] denom
func CreateMonthlyContinuous(cur realm, namespace string, label string, beneficiary address, amountPerMonth int64, denom string) seqid.ID {
	if amountPerMonth <= 0 || denom == "" {
		panic(errors.New("invalid input"))
	}
	coinsPerMonth := chain.NewCoins(chain.NewCoin(denom, amountPerMonth))
	return Create(cur, namespace, label, beneficiary, DistribMonthlyContinuous(cur, coinsPerMonth), nil)
}

// CreateCDI creates a payroll that tries to match the French CDI contract
//
// Denom must be a payrolls [Coins] denom
//
// See: https://travail-emploi.gouv.fr/le-contrat-de-travail-duree-indeterminee-cdi
func CreateCDI(cur realm, namespace string, label string, beneficiary address, amountPerMonth int64, denom string) seqid.ID {
	if amountPerMonth <= 0 || denom == "" {
		panic(errors.New("invalid input"))
	}
	coinsPerMonth := chain.NewCoins(chain.NewCoin(denom, amountPerMonth))
	return Create(cur, namespace, label, beneficiary, DistribMonthlyStep(cur, coinsPerMonth), BreakupCDI(cur, coinsPerMonth))
}

// BreakupCDI is a breakup function that will release an amount of coins
// according to a simplified "Rupture Conventionelle" (conventional breakup) model based on the French CDI employment contract
func BreakupCDI(cur realm, coinsPerMonth chain.Coins) BreakupFn {
	const hoursPerYear = float64(8_670)
	return func(cur realm, elapsed time.Duration, pauseDuration time.Duration, source CallSource) chain.Coins {
		if source != CallSourceCreator {
			return nil
		}

		// conventional breakup
		out := make(chain.Coins, 0, len(coinsPerMonth))
		worked := elapsed - pauseDuration
		numYears := worked.Hours() / hoursPerYear
		for _, coin := range coinsPerMonth {
			newAmount := int64(0.25 * float64(coin.Amount) * numYears)
			out = append(out, chain.NewCoin(coin.Denom, newAmount))
		}
		return out
	}
}
