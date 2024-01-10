package task

import (
	"math"
	"strconv"
)

type Aggregator struct {
	sum            uint64
	taskDefinition *Definition
}

func (a *Aggregator) Aggregate() string {
	randomValue := a.taskDefinition.RangeStart + a.sum%a.taskDefinition.RangeEnd
	a.sum = 0
	return strconv.FormatUint(randomValue, 10)
}

func (a *Aggregator) AddValue(value string) {
	intValue, err := strconv.Atoi(value)
	if err != nil {
		panic("value needs to be type uint64: " + err.Error())
	}

	a.sum += uint64(intValue)

	// No need to check for overflow currently because ParseUint is not supported.

	// Account for overflow.
	// if diff := math.MaxUint64 - a.sum; diff < intValue {
	// 	a.sum = intValue - diff
	// } else {
	// 	a.sum += intValue
	// }
}
