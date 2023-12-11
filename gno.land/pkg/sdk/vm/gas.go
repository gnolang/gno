package vm

import (
	"fmt"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"

	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/overflow"
)

const (

	// We will use gasMemFactor and gasCpuFactor to multiply the vm memory allocation and cpu cycles to get the gas number.
	// We can use these two factoctors to keep the gas for storage access, CPU and Mem in reasonable proportion.
	gasFactorMem int64 = 1 // change this value based on gas profiling
	gasFactorCpu int64 = 1 // change this value based on gas profiling

	logPrefixAddPkg   = "gas.vm.addpkg"
	logPrefixCall     = "gas.vm.call"
	logPrefixRun      = "gas.vm.run"
	logPrefixQeval    = "gas.vm.qeval"
	logPrefixQevalStr = "gas.vm.qevalstr"
)

// consume gas and log vm gas usage
func consumeGas(ctx sdk.Context, m *gno.Machine, prefix string, pkgPath string, expr string) {
	_, mem := m.Alloc.Status()

	gasCpu := overflow.Mul64p(m.Cycles, gasFactorCpu)
	gasMem := overflow.Mul64p(mem, gasFactorMem)

	// we simplify the log here, the storage gas log included tx size and sigature verification gas.
	storeLog := fmt.Sprintf("%s.storage, %s %s, %d", prefix, pkgPath, expr, ctx.GasMeter().GasConsumed())
	ctx.Logger().Info(storeLog)

	memLog := fmt.Sprintf("%s.memalloc, %s %s, %d", prefix, pkgPath, expr, gasMem)
	ctx.Logger().Info(memLog)

	cpuLog := fmt.Sprintf("%s.cpucycles, %s %s, %d", prefix, pkgPath, expr, gasCpu)
	ctx.Logger().Info(cpuLog)

	ctx.GasMeter().ConsumeGas(gasMem, prefix+".MemAlloc")
	ctx.GasMeter().ConsumeGas(gasCpu, prefix+".CpuCycles")

	gasTotal := fmt.Sprintf("%s.total, %s %s, %d", prefix, pkgPath, expr, ctx.GasMeter().GasConsumed())
	ctx.Logger().Info(gasTotal)
}
