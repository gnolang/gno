package payrolls

import (
	"errors"
	"std"
	"time"

	"gno.land/p/demo/seqid"
)

type DistribFn func(elapsed time.Duration) uint64
type BreakupFn func(elapsed time.Duration, pauseDuration time.Duration, source BreakupSource) int64

type BreakupSource uint

const (
	BreakupSourceCreator BreakupSource = iota
	BreakupSourceBeneficiary
)

func DistribStep(amount uint64, per time.Duration) DistribFn {
	return func(elapsed time.Duration) uint64 {
		blocksCount := elapsed / per
		return amount * uint64(blocksCount)
	}
}

func DistribLinear(amount uint64, per time.Duration) DistribFn {
	return func(elapsed time.Duration) uint64 {
		return (uint64(elapsed.Seconds()) * amount) / uint64(per.Seconds())
	}
}

func DistribMonthlyStep(amountPerMonth uint64) DistribFn {
	return DistribStep(amountPerMonth, time.Hour*24*30)
}

func DistribMonthlyContinuous(amountPerMonth uint64) DistribFn {
	return DistribLinear(amountPerMonth, time.Hour*24*30)
}

// CreateCDI creates a payroll that tries to match the French CDI contract
//
// See: https://travail-emploi.gouv.fr/le-contrat-de-travail-duree-indeterminee-cdi
func CreateCDI(namespace string, label string, beneficiary std.Address, amountPerMonth uint64) seqid.ID {
	if amountPerMonth <= 0 {
		panic(errors.New("invalid amount"))
	}
	return Create(namespace, label, beneficiary, -1, DistribMonthlyContinuous(amountPerMonth), BreakupCDI(amountPerMonth))
}

func BreakupCDI(amountPerMonth uint64) BreakupFn {
	const hoursPerYear = float64(8_670)
	return func(elapsed time.Duration, pauseDuration time.Duration, source BreakupSource) int64 {
		if source != BreakupSourceCreator {
			return 0
		}

		// conventional breakup
		worked := elapsed - pauseDuration
		numYears := worked.Hours() / hoursPerYear
		return int64(0.25 * float64(amountPerMonth) * numYears)
	}
}
