package std

/*
	ref: https://github.com/gnolang/gno/issues/1998
	this file contains std functions for query gas cost, gas used etc...
*/
import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/store"
)

var (
	gasUsedInvoked = "GasUsedCalled"
	/*
		Consider where to save this config
		defaultGasConfig = store.DefaultGasConfig()
	*/
	// defaultGasConfig = int64(1000)
	defaultInvokeCost = store.DefaultGasConfig().ReadCostFlat
)

// DefaultCost will be consumed whenever GasUsed is called, now set it ReadCostPerByte
func GasUsed(m *gno.Machine) int64 {
	m.GasMeter.ConsumeGas(defaultInvokeCost, gasUsedInvoked)
	return m.GasMeter.GasConsumedToLimit()
}
