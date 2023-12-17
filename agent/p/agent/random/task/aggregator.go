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
	intValue, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		panic("value needs to be type uint64: " + err.Error())
	}

	// Account for overflow.
	if diff := math.MaxUint64 - a.sum; diff < intValue {
		a.sum = intValue - diff
	} else {
		a.sum += intValue
	}
}
