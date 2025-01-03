package payrolls

import (
	"errors"
	"std"
	"time"

	"gno.land/p/demo/seqid"
)

type DistribFn func(elapsed time.Duration) std.Coins
type BreakupFn func(elapsed time.Duration, pauseDuration time.Duration, source BreakupSource) std.Coins

type BreakupSource uint

const (
	BreakupSourceCreator BreakupSource = iota
	BreakupSourceBeneficiary
)

func DistribStep(coins std.Coins, per time.Duration) DistribFn {
	return func(elapsed time.Duration) std.Coins {
		blocksCount := elapsed / per
		out := make(std.Coins, 0, len(coins))
		for _, coin := range coins {
			out = append(out, std.NewCoin(coin.Denom, coin.Amount*int64(blocksCount)))
		}
		return out
	}
}

func DistribLinear(coins std.Coins, per time.Duration) DistribFn {
	return func(elapsed time.Duration) std.Coins {
		out := make(std.Coins, 0, len(coins))
		for _, coin := range coins {
			newAmount := (int64(elapsed.Seconds()) * coin.Amount) / int64(per.Seconds())
			out = append(out, std.NewCoin(coin.Denom, newAmount))
		}
		return out
	}
}

func DistribMonthlyStep(amountPerMonth std.Coins) DistribFn {
	return DistribStep(amountPerMonth, time.Hour*24*30)
}

func DistribMonthlyContinuous(amountPerMonth std.Coins) DistribFn {
	return DistribLinear(amountPerMonth, time.Hour*24*30)
}

// CreateCDI creates a payroll that tries to match the French CDI contract
//
// See: https://travail-emploi.gouv.fr/le-contrat-de-travail-duree-indeterminee-cdi
func CreateCDI(namespace string, label string, beneficiary std.Address, amountPerMonth int64, denom string) seqid.ID {
	if amountPerMonth <= 0 || denom == "" {
		panic(errors.New("invalid input"))
	}
	coinsPerMonth := std.NewCoins(std.NewCoin(denom, amountPerMonth))
	return Create(namespace, label, beneficiary, -1, DistribMonthlyContinuous(coinsPerMonth), BreakupCDI(coinsPerMonth))
}

func BreakupCDI(coinsPerMonth std.Coins) BreakupFn {
	const hoursPerYear = float64(8_670)
	return func(elapsed time.Duration, pauseDuration time.Duration, source BreakupSource) std.Coins {
		if source != BreakupSourceCreator {
			return nil
		}

		// conventional breakup
		out := make(std.Coins, 0, len(coinsPerMonth))
		worked := elapsed - pauseDuration
		numYears := worked.Hours() / hoursPerYear
		for _, coin := range coinsPerMonth {
			newAmount := int64(0.25 * float64(coin.Amount) * numYears)
			out = append(out, std.NewCoin(coin.Denom, newAmount))
		}
		return out
	}
}
