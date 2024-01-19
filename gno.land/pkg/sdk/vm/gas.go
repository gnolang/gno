package vm

import (
	"fmt"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/overflow"
)

const (

	// We will use gasFactorMem and gasFactorCpu to multiply the vm memory allocation and cpu cycles to get the gas number.
	// We can use these two factors to keep the gas for storage access, CPU and Mem in reasonable proportion.
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

	// we simplify the log here, the storage gas log included tx size and signature verification gas in  CheckTx()
	storeLog := fmt.Sprintf("%s.txsize_sig_storage, %s %s, %d", prefix, pkgPath, expr, ctx.GasMeter().GasConsumed())
	ctx.Logger().Info(storeLog)

	memLog := fmt.Sprintf("%s.memalloc, %s %s, %d", prefix, pkgPath, expr, gasMem)
	ctx.Logger().Info(memLog)

	cpuLog := fmt.Sprintf("%s.cpucycles, %s %s, %d", prefix, pkgPath, expr, gasCpu)
	ctx.Logger().Info(cpuLog)

	defer func() {
		if r := recover(); r != nil {
			m.Release()
			switch r.(type) {
			case store.OutOfGasException: // panic in consumeGas()
				panic(r)
			default:
				panic("should not happen")
			}
		}
	}()

	ctx.GasMeter().ConsumeGas(gasMem, prefix+".MemAlloc")
	ctx.GasMeter().ConsumeGas(gasCpu, prefix+".CpuCycles")

	gasTotal := fmt.Sprintf("%s.total, %s %s, %d", prefix, pkgPath, expr, ctx.GasMeter().GasConsumed())
	ctx.Logger().Info(gasTotal)
}
