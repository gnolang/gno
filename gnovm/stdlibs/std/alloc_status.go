package std

//Related issue: https://github.com/gnolang/gno/issues/2070

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func GetCurrAllocatedMem(m *gno.Machine) int64 {
	_, currentAlloc := m.Alloc.Status()
	return currentAlloc
}

func GetAllocMaxSize(m *gno.Machine) int64 {
	maxMemSize, _ := m.Alloc.Status()
	return maxMemSize
}

func GetAllocStatus(m *gno.Machine) (int64, int64) {
	return m.Alloc.Status()
}

func GetMemGasUsage(m *gno.Machine) (gasUsed int64, memUsage float64) {
	gasUsed = m.GasMeter.GasConsumedToLimit()
	totalMem, allocatedMem := GetAllocStatus(m)
	memUsage = float64(allocatedMem) / float64(totalMem) * 100
	return gasUsed, memUsage
}

