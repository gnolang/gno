package std

/*
	ref: https://github.com/gnolang/gno/issues/1998
	this file contains std functions for query gas cost, gas used etc...
*/
import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

var (
	gasUsedInvoked = "GasUsedCalled"
	/*
		Consider where to save this config
		gasCostDefault = store.DefaultGasConfig()
	*/
	defaultInvokerGasUsedCost = int64(3)
)

// DefaultCost will be consumed whenever GasUsed is called, now set it ReadCostPerByte
func GasUsed(m *gno.Machine) int64 {
	m.GasMeter.ConsumeGas(int64(defaultInvokerGasUsedCost), gasUsedInvoked)
	return m.GasMeter.GasConsumedToLimit()
}
