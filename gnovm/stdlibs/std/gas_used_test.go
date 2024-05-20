package std

import (
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/stretchr/testify/assert"
)

// Focus on test gasToConsume()
func TestGasUsed(t *testing.T) {
	t.Parallel()
	m := gno.NewMachine("gasToConsume", nil)
	testTable := []struct {
		tcName          string
		gasToConsume    int64
		gasLimit        int64
		expectPastLimit bool
		invokeCost      int64
	}{
		{"Test GasUsed Get", 10, 100, false, defaultInvokerGasUsedCost},
		{"Test GasUsed Invoke Cost", 4, 4 + defaultInvokerGasUsedCost, false, defaultInvokerGasUsedCost},
		{"Test GasUsed Past Limit", 4, 4, true, defaultInvokerGasUsedCost},
		// this case is OutOfGas's behavior
		// {"Test GasUsed Get When Out Of Gas", 40, 4, true, cf.ReadCostPerByte},
	}
	for _, tc := range testTable {
		t.Run(tc.tcName, func(t *testing.T) {
			m.GasMeter = store.NewGasMeter(tc.gasLimit)
			if tc.expectPastLimit {
				m.GasMeter.ConsumeGas(tc.gasToConsume, tc.tcName)
				// After consume gasToConsume, the remaining gas is lower than invokeCost
				// then GasUsed() should panic(OutOfGas)
				beforeInvoke := m.GasMeter.Remaining()
				assert.Panics(t, func() {
					GasUsed(m)
				})
				afterInvoke := m.GasMeter.Remaining()
				// Check if GasMeter() acts as expected
				assert.Equal(t, tc.expectPastLimit, m.GasMeter.IsOutOfGas())
				// Check if GasUsed() will not consume invokeCost
				assert.Equal(t, beforeInvoke, afterInvoke)
			} else {
				m.GasMeter.ConsumeGas(tc.gasToConsume, tc.tcName)
				beforeInvoke := m.GasMeter.Remaining()
				result := GasUsed(m)
				afterInvoke := m.GasMeter.Remaining()
				// Check if GasUsed() invoked the invokeCost
				assert.Equal(t, beforeInvoke-afterInvoke, tc.invokeCost)
				// Check if the process consumes ecxactly amount of gas
				assert.Equal(t, tc.gasToConsume+tc.invokeCost, result)
			}
		})
	}
}
