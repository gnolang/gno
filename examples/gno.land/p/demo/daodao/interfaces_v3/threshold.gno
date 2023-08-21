package dao_interfaces

import (
	"strconv"
)

type Percent uint16 // 4 decimals fixed point

type PercentageThreshold struct {
	Percent *Percent
}

func (p *PercentageThreshold) String() string {
	if p == nil || p.Percent == nil {
		return "nil"
	}
	return p.Percent.String()
}

type ThresholdQuorum struct {
	Threshold PercentageThreshold
	Quorum    PercentageThreshold
}

type Threshold struct {
	ThresholdQuorum *ThresholdQuorum
}

func (p Percent) String() string {
	s := strconv.FormatUint(uint64(p)/100, 10)
	decPart := uint64(p) % 100
	if decPart != 0 {
		s += "." + strconv.FormatUint(decPart, 10)
	}
	return s + "%"
}
