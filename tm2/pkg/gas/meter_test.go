package gas

import (
	"math"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/overflow"
	"github.com/stretchr/testify/require"
)

func TestGasMeter(t *testing.T) {
	t.Parallel()

	cases := []struct {
		limit Gas
		usage []Gas
	}{
		{10, []Gas{1, 2, 3, 4}},
		{1000, []Gas{40, 30, 20, 10, 900}},
		{100000, []Gas{99999, 1}},
		{100000000, []Gas{50000000, 40000000, 10000000}},
		{65535, []Gas{32768, 32767}},
		{65536, []Gas{32768, 32767, 1}},
	}

	for tcnum, tc := range cases {
		meter := NewMeter(tc.limit, DefaultConfig())
		used := int64(0)

		for unum, usage := range tc.usage {
			used += usage
			require.NotPanics(t, func() { meter.ConsumeGas(OpTesting, float64(usage)) }, "Not exceeded limit but panicked. tc #%d, usage #%d", tcnum, unum)
			require.Equal(t, used, meter.GasConsumed(), "Gas consumption not match. tc #%d, usage #%d", tcnum, unum)
			require.Equal(t, used, meter.GasConsumedToLimit(), "Gas consumption (to limit) not match. tc #%d, usage #%d", tcnum, unum)
			require.False(t, meter.IsPastLimit(), "Not exceeded limit but got IsPastLimit() true")
			if unum < len(tc.usage)-1 {
				require.False(t, meter.IsOutOfGas(), "Not yet at limit but got IsOutOfGas() true")
			} else {
				require.True(t, meter.IsOutOfGas(), "At limit but got IsOutOfGas() false")
			}
		}

		require.Panics(t, func() { meter.ConsumeGas(OpTesting, 1) }, "Exceeded but not panicked. tc #%d", tcnum)
		require.Equal(t, meter.GasConsumedToLimit(), meter.Limit(), "Gas consumption (to limit) not match limit")
		require.Equal(t, meter.GasConsumed(), meter.Limit()+1, "Gas consumption not match limit+1")
	}
}

func TestAddUint64Overflow(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		a, b     int64
		result   int64
		overflow bool
	}{
		{0, 0, 0, false},
		{100, 100, 200, false},
		{math.MaxInt64 / 2, math.MaxInt64/2 + 1, math.MaxInt64, false},
		{math.MaxInt64 / 2, math.MaxInt64/2 + 2, math.MinInt64, true},
	}

	for i, tc := range testCases {
		res, ok := overflow.Add(tc.a, tc.b)
		overflow := !ok
		require.Equal(
			t, tc.overflow, overflow,
			"invalid overflow result; tc: #%d, a: %d, b: %d", i, tc.a, tc.b,
		)
		require.Equal(
			t, tc.result, res,
			"invalid int64 result; tc: #%d, a: %d, b: %d", i, tc.a, tc.b,
		)
	}
}

func TestGasDetailTracking(t *testing.T) {
	t.Parallel()

	config := DefaultConfig()
	meter := NewMeter(1000000, config)

	// Consume gas for different operations
	// Store operations
	meter.ConsumeGas(OpStoreReadFlat, 1)
	meter.ConsumeGas(OpStoreReadFlat, 1)
	meter.ConsumeGas(OpStoreWriteFlat, 1)

	// CPU operations
	meter.ConsumeGas(OpCPUAdd, 1)
	meter.ConsumeGas(OpCPUAdd, 1)
	meter.ConsumeGas(OpCPUMul, 1)

	// Memory operations
	meter.ConsumeGas(OpMemoryAllocPerByte, 10)

	// Get gas detail
	detail := meter.GasDetail()

	// Verify total consumed
	require.Equal(t, meter.GasConsumed(), detail.Total.GasConsumed, "Total consumed should match gas consumed")

	// Verify operation counts
	require.Equal(t, int64(2), detail.Operations[OpStoreReadFlat].OperationCount, "OpStoreReadFlat should be called 2 times")
	require.Equal(t, int64(1), detail.Operations[OpStoreWriteFlat].OperationCount, "OpStoreWriteFlat should be called 1 time")
	require.Equal(t, int64(2), detail.Operations[OpCPUAdd].OperationCount, "OpCPUAdd should be called 2 times")
	require.Equal(t, int64(1), detail.Operations[OpCPUMul].OperationCount, "OpCPUMul should be called 1 time")
	require.Equal(t, int64(1), detail.Operations[OpMemoryAllocPerByte].OperationCount, "OpMemoryAllocPerByte should be called 1 time")

	// Verify operation gas totals
	expectedStoreReadFlatGas := config.Costs[OpStoreReadFlat] * 2
	require.Equal(t, Gas(expectedStoreReadFlatGas), detail.Operations[OpStoreReadFlat].GasConsumed, "OpStoreReadFlat total gas should match")

	expectedCPUAddGas := config.Costs[OpCPUAdd] * 2
	require.Equal(t, Gas(expectedCPUAddGas), detail.Operations[OpCPUAdd].GasConsumed, "OpCPUAdd total gas should match")

	// Get category details
	categoryDetails := detail.CategoryDetails()

	// Verify category totals
	require.Greater(t, categoryDetails["Store"].Total.GasConsumed, Gas(0), "Store category should have gas consumed")
	require.Greater(t, categoryDetails["CPU"].Total.GasConsumed, Gas(0), "CPU category should have gas consumed")
	require.Greater(t, categoryDetails["Memory"].Total.GasConsumed, Gas(0), "Memory category should have gas consumed")

	// Verify category gas equals sum of operations in that category
	expectedStoreGas := detail.Operations[OpStoreReadFlat].GasConsumed + detail.Operations[OpStoreWriteFlat].GasConsumed
	require.Equal(t, expectedStoreGas, categoryDetails["Store"].Total.GasConsumed, "Store category gas should match sum of store operations")

	expectedCPUGas := detail.Operations[OpCPUAdd].GasConsumed + detail.Operations[OpCPUMul].GasConsumed
	require.Equal(t, expectedCPUGas, categoryDetails["CPU"].Total.GasConsumed, "CPU category gas should match sum of CPU operations")

	expectedMemoryGas := detail.Operations[OpMemoryAllocPerByte].GasConsumed
	require.Equal(t, expectedMemoryGas, categoryDetails["Memory"].Total.GasConsumed, "Memory category gas should match memory operation")

	// Verify total equals sum of all categories
	totalFromCategories := Gas(0)
	for _, category := range categoryDetails {
		totalFromCategories += category.Total.GasConsumed
	}
	require.Equal(t, detail.Total.GasConsumed, totalFromCategories, "Total consumed should equal sum of all categories")
}

func TestGasDetailInfiniteMeter(t *testing.T) {
	t.Parallel()

	config := DefaultConfig()
	meter := NewInfiniteMeter(config)

	// Consume gas for different operations
	meter.ConsumeGas(OpStoreReadFlat, 1)
	meter.ConsumeGas(OpCPUAdd, 1)

	// Get gas detail
	detail := meter.GasDetail()

	// Verify total consumed
	require.Equal(t, meter.GasConsumed(), detail.Total.GasConsumed, "Total consumed should match gas consumed")

	// Verify operation counts
	require.Equal(t, int64(1), detail.Operations[OpStoreReadFlat].OperationCount, "OpStoreReadFlat should be called 1 time")
	require.Equal(t, int64(1), detail.Operations[OpCPUAdd].OperationCount, "OpCPUAdd should be called 1 time")

	// Get category details
	categoryDetails := detail.CategoryDetails()

	// Verify category totals
	require.Greater(t, categoryDetails["Store"].Total.GasConsumed, Gas(0), "Store category should have gas consumed")
	require.Greater(t, categoryDetails["CPU"].Total.GasConsumed, Gas(0), "CPU category should have gas consumed")
}

func TestGasDetailPassthroughMeter(t *testing.T) {
	t.Parallel()

	config := DefaultConfig()
	baseMeter := NewMeter(1000000, config)
	passthroughMeter := NewPassthroughMeter(baseMeter, 500000, config)

	// Consume gas through passthrough meter
	passthroughMeter.ConsumeGas(OpStoreReadFlat, 1)
	passthroughMeter.ConsumeGas(OpCPUAdd, 1)

	// Get gas detail from passthrough meter (should return Head's detail)
	detail := passthroughMeter.GasDetail()

	// Verify total consumed
	require.Equal(t, passthroughMeter.GasConsumed(), detail.Total.GasConsumed, "Total consumed should match gas consumed")

	// Verify operation counts
	require.Equal(t, int64(1), detail.Operations[OpStoreReadFlat].OperationCount, "OpStoreReadFlat should be called 1 time")
	require.Equal(t, int64(1), detail.Operations[OpCPUAdd].OperationCount, "OpCPUAdd should be called 1 time")

	// Get category details
	categoryDetails := detail.CategoryDetails()

	// Verify category totals
	require.Greater(t, categoryDetails["Store"].Total.GasConsumed, Gas(0), "Store category should have gas consumed")
	require.Greater(t, categoryDetails["CPU"].Total.GasConsumed, Gas(0), "CPU category should have gas consumed")
}

func TestMeterPanics(t *testing.T) {
	t.Parallel()

	const maxSafeFloat64 float64 = (1 << 53) - 1.0
	config := DefaultConfig()

	t.Run("negative gas limit", func(t *testing.T) {
		t.Parallel()

		require.Panics(t, func() {
			NewMeter(-1, config)
		}, "Should panic with negative gas limit")
	})

	t.Run("zero multiplier", func(t *testing.T) {
		t.Parallel()

		invalidConfig := config
		invalidConfig.GlobalMultiplier = 0
		require.Panics(t, func() {
			NewMeter(1000, invalidConfig)
		}, "Should panic with zero multiplier")
	})

	t.Run("negative multiplier", func(t *testing.T) {
		t.Parallel()

		invalidConfig := config
		invalidConfig.GlobalMultiplier = -1
		require.Panics(t, func() {
			NewMeter(1000, invalidConfig)
		}, "Should panic with negative multiplier")
	})

	t.Run("infinite meter zero multiplier", func(t *testing.T) {
		t.Parallel()

		invalidConfig := config
		invalidConfig.GlobalMultiplier = 0
		require.Panics(t, func() {
			NewInfiniteMeter(invalidConfig)
		}, "Should panic with zero multiplier")
	})

	t.Run("infinite meter negative multiplier", func(t *testing.T) {
		t.Parallel()

		invalidConfig := config
		invalidConfig.GlobalMultiplier = -1
		require.Panics(t, func() {
			NewInfiniteMeter(invalidConfig)
		}, "Should panic with negative multiplier")
	})

	t.Run("passthrough negative limit", func(t *testing.T) {
		t.Parallel()

		baseMeter := NewMeter(1000, config)
		require.Panics(t, func() {
			NewPassthroughMeter(baseMeter, -1, config)
		}, "Should panic with negative limit")
	})

	t.Run("overflow in gas consumption", func(t *testing.T) {
		t.Parallel()

		meter := NewMeter(math.MaxInt64, config)

		// Consume gas multiple times to get close to MaxInt64
		iterations := math.MaxInt64 / int64(maxSafeFloat64)
		for range iterations {
			meter.ConsumeGas(OpTesting, maxSafeFloat64)
		}

		// Now we're close enough to MaxInt64, consuming another time will overflow.
		require.Panics(t, func() {
			meter.ConsumeGas(OpTesting, maxSafeFloat64)
		}, "Should panic on overflow")
	})

	t.Run("overflow in gas consumption (infinite meter)", func(t *testing.T) {
		t.Parallel()

		meter := NewInfiniteMeter(config)

		// Consume gas multiple times to get close to MaxInt64
		iterations := math.MaxInt64 / int64(maxSafeFloat64)
		for range iterations {
			meter.ConsumeGas(OpTesting, maxSafeFloat64)
		}

		// Now we're close enough to MaxInt64, consuming another time will overflow.
		require.Panics(t, func() {
			meter.ConsumeGas(OpTesting, maxSafeFloat64)
		}, "Should panic on overflow")
	})

	t.Run("overflow in calculateGasCost", func(t *testing.T) {
		t.Parallel()

		hugeMultiplierConfig := config
		hugeMultiplierConfig.GlobalMultiplier = math.MaxFloat64
		hugeMultiplierConfig.Costs[OpTesting] = math.MaxFloat64
		meter := NewMeter(math.MaxInt64, hugeMultiplierConfig)
		require.Panics(t, func() {
			meter.ConsumeGas(OpTesting, math.MaxFloat64)
		}, "Should panic on gas calculation overflow")
	})

	t.Run("precision error in calculateGasCost", func(t *testing.T) {
		t.Parallel()

		meter := NewMeter(math.MaxInt64, config)

		require.Panics(t, func() {
			meter.ConsumeGas(OpTesting, maxSafeFloat64+1.0)
		}, "Should panic on gas calculation precision loss")

		require.NotPanics(t, func() {
			meter.ConsumeGas(OpTesting, maxSafeFloat64)
		}, "Should not panic on gas calculation without precision loss")
	})
}

func TestErrorMessages(t *testing.T) {
	t.Parallel()

	t.Run("OutOfGasError", func(t *testing.T) {
		t.Parallel()

		err := OutOfGasError{"test-operation"}
		expected := "out of gas in location: test-operation"
		require.Equal(t, expected, err.Error(), "OutOfGasError message should match")
	})

	t.Run("OverflowError", func(t *testing.T) {
		t.Parallel()

		err := OverflowError{"test-overflow"}
		expected := "gas overflow in location: test-overflow"
		require.Equal(t, expected, err.Error(), "OverflowError message should match")
	})

	t.Run("PrecisionError", func(t *testing.T) {
		t.Parallel()

		err := PrecisionError{"test-precision"}
		expected := "gas precision loss in location: test-precision"
		require.Equal(t, expected, err.Error(), "PrecisionError message should match")
	})
}

func TestUtilityMethods(t *testing.T) {
	t.Parallel()

	config := DefaultConfig()

	t.Run("Config getter", func(t *testing.T) {
		t.Parallel()

		meter := NewMeter(1000, config)
		retrievedConfig := meter.Config()
		require.Equal(t, config.GlobalMultiplier, retrievedConfig.GlobalMultiplier, "Config should match")
	})

	t.Run("Remaining gas", func(t *testing.T) {
		t.Parallel()

		meter := NewMeter(1000, config)
		require.Equal(t, Gas(1000), meter.Remaining(), "Initial remaining should equal limit")

		meter.ConsumeGas(OpTesting, 100)
		require.Equal(t, Gas(900), meter.Remaining(), "Remaining should decrease after consumption")
	})

	t.Run("GetCostForOperation", func(t *testing.T) {
		t.Parallel()

		cost := config.GetCostForOperation(OpStoreReadFlat)
		require.Equal(t, config.Costs[OpStoreReadFlat], cost, "GetCostForOperation should return correct cost")
	})

	t.Run("Detail String", func(t *testing.T) {
		t.Parallel()

		detail := Detail{OperationCount: 5, GasConsumed: 100}
		str := detail.String()
		require.Contains(t, str, "5", "String should contain operation count")
		require.Contains(t, str, "100", "String should contain gas consumed")
	})

	t.Run("Operation String", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, "StoreReadFlat", OpStoreReadFlat.String(), "Operation name should match")
		require.Equal(t, "CPUAdd", OpCPUAdd.String(), "Operation name should match")
	})

	t.Run("Unknown operation String", func(t *testing.T) {
		t.Parallel()

		// Create an operation with a value that has no name
		unknownOp := Operation(100) // Assuming 100 has no name defined
		// First verify it has no name in the array
		if operationNames[unknownOp] == "" {
			require.Equal(t, "UnknownOperation", unknownOp.String(), "Unknown operation should return 'UnknownOperation'")
		}
	})
}

func TestInfiniteMeterBehavior(t *testing.T) {
	t.Parallel()

	config := DefaultConfig()
	meter := NewInfiniteMeter(config)

	t.Run("never runs out of gas", func(t *testing.T) {
		t.Parallel()

		// Consume a huge amount of gas
		for i := 0; i < 1000; i++ {
			require.NotPanics(t, func() {
				meter.ConsumeGas(OpTesting, 1000000)
			}, "Infinite meter should never panic")
		}
	})

	t.Run("IsPastLimit always false", func(t *testing.T) {
		t.Parallel()

		meter.ConsumeGas(OpTesting, 999999999)
		require.False(t, meter.IsPastLimit(), "Infinite meter should never be past limit")
	})

	t.Run("IsOutOfGas always false", func(t *testing.T) {
		t.Parallel()

		require.False(t, meter.IsOutOfGas(), "Infinite meter should never be out of gas")
	})

	t.Run("Limit is zero", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, Gas(0), meter.Limit(), "Infinite meter limit should be 0")
	})

	t.Run("Remaining is MaxInt64", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, Gas(math.MaxInt64), meter.Remaining(), "Infinite meter remaining should be MaxInt64")
	})

	t.Run("GasConsumedToLimit equals GasConsumed", func(t *testing.T) {
		t.Parallel()

		consumed := meter.GasConsumed()
		require.Equal(t, consumed, meter.GasConsumedToLimit(), "GasConsumedToLimit should equal GasConsumed")
	})

	t.Run("Config getter", func(t *testing.T) {
		t.Parallel()

		retrievedConfig := meter.Config()
		require.Equal(t, config.GlobalMultiplier, retrievedConfig.GlobalMultiplier, "Config should match")
	})
}

func TestPassthroughMeterLimits(t *testing.T) {
	t.Parallel()

	config := DefaultConfig()

	t.Run("respects head limit", func(t *testing.T) {
		t.Parallel()

		baseMeter := NewMeter(10000, config)
		passthroughMeter := NewPassthroughMeter(baseMeter, 5000, config)

		// Consume up to head limit
		passthroughMeter.ConsumeGas(OpTesting, 5000)
		require.True(t, passthroughMeter.IsOutOfGas(), "Should be out of gas at head limit")

		// Should panic on next consume
		require.Panics(t, func() {
			passthroughMeter.ConsumeGas(OpTesting, 1)
		}, "Should panic when exceeding head limit")
	})

	t.Run("delegates to base meter", func(t *testing.T) {
		t.Parallel()

		baseMeter := NewMeter(10000, config)
		passthroughMeter := NewPassthroughMeter(baseMeter, 5000, config)

		passthroughMeter.ConsumeGas(OpTesting, 1000)

		// Both meters should have consumed the same gas
		require.Equal(t, Gas(1000), passthroughMeter.GasConsumed(), "Passthrough should track consumption")
		require.Equal(t, Gas(1000), baseMeter.GasConsumed(), "Base should track consumption")
	})

	t.Run("GasConsumedToLimit", func(t *testing.T) {
		t.Parallel()

		baseMeter := NewMeter(10000, config)
		passthroughMeter := NewPassthroughMeter(baseMeter, 5000, config)

		passthroughMeter.ConsumeGas(OpTesting, 3000)
		require.Equal(t, Gas(3000), passthroughMeter.GasConsumedToLimit(), "GasConsumedToLimit should match")
	})

	t.Run("Limit returns head limit", func(t *testing.T) {
		t.Parallel()

		baseMeter := NewMeter(10000, config)
		passthroughMeter := NewPassthroughMeter(baseMeter, 5000, config)

		require.Equal(t, Gas(5000), passthroughMeter.Limit(), "Should return head limit")
	})

	t.Run("Remaining", func(t *testing.T) {
		t.Parallel()

		baseMeter := NewMeter(10000, config)
		passthroughMeter := NewPassthroughMeter(baseMeter, 5000, config)

		require.Equal(t, Gas(5000), passthroughMeter.Remaining(), "Initial remaining should equal head limit")
		passthroughMeter.ConsumeGas(OpTesting, 2000)
		require.Equal(t, Gas(3000), passthroughMeter.Remaining(), "Remaining should decrease")
	})

	t.Run("CalculateGasCost", func(t *testing.T) {
		t.Parallel()

		baseMeter := NewMeter(10000, config)
		passthroughMeter := NewPassthroughMeter(baseMeter, 5000, config)

		cost := passthroughMeter.CalculateGasCost(OpTesting, 10)
		expectedCost := Gas(10) // OpTesting has cost 1 in default config
		require.Equal(t, expectedCost, cost, "CalculateGasCost should work correctly")
	})

	t.Run("IsPastLimit", func(t *testing.T) {
		t.Parallel()

		baseMeter := NewMeter(10000, config)
		passthroughMeter := NewPassthroughMeter(baseMeter, 5000, config)

		require.False(t, passthroughMeter.IsPastLimit(), "Should not be past limit initially")
		passthroughMeter.ConsumeGas(OpTesting, 5000)
		require.False(t, passthroughMeter.IsPastLimit(), "Should not be past limit at exact limit")

		require.Panics(t, func() {
			passthroughMeter.ConsumeGas(OpTesting, 1)
		})
		require.True(t, passthroughMeter.IsPastLimit(), "Should be past limit after exceeding")
	})

	t.Run("IsOutOfGas", func(t *testing.T) {
		t.Parallel()

		baseMeter := NewMeter(10000, config)
		passthroughMeter := NewPassthroughMeter(baseMeter, 5000, config)

		require.False(t, passthroughMeter.IsOutOfGas(), "Should not be out of gas initially")
		passthroughMeter.ConsumeGas(OpTesting, 5000)
		require.True(t, passthroughMeter.IsOutOfGas(), "Should be out of gas at limit")
	})

	t.Run("Config getter", func(t *testing.T) {
		t.Parallel()

		baseMeter := NewMeter(10000, config)
		passthroughMeter := NewPassthroughMeter(baseMeter, 5000, config)

		retrievedConfig := passthroughMeter.Config()
		require.Equal(t, config.GlobalMultiplier, retrievedConfig.GlobalMultiplier, "Config should match")
	})
}

func TestGlobalMultiplier(t *testing.T) {
	t.Parallel()

	t.Run("multiplier 1.0", func(t *testing.T) {
		t.Parallel()

		config := DefaultConfig()
		config.GlobalMultiplier = 1.0
		config.Costs[OpTesting] = 100
		meter := NewMeter(10000, config)

		meter.ConsumeGas(OpTesting, 1)
		require.Equal(t, Gas(100), meter.GasConsumed(), "Should consume exact cost with multiplier 1.0")
	})

	t.Run("multiplier 0.5", func(t *testing.T) {
		t.Parallel()

		config := DefaultConfig()
		config.GlobalMultiplier = 0.5
		config.Costs[OpTesting] = 100
		meter := NewMeter(10000, config)

		meter.ConsumeGas(OpTesting, 1)
		require.Equal(t, Gas(50), meter.GasConsumed(), "Should halve gas with multiplier 0.5")
	})

	t.Run("multiplier 2.0", func(t *testing.T) {
		t.Parallel()

		config := DefaultConfig()
		config.GlobalMultiplier = 2.0
		config.Costs[OpTesting] = 100
		meter := NewMeter(10000, config)

		meter.ConsumeGas(OpTesting, 1)
		require.Equal(t, Gas(200), meter.GasConsumed(), "Should double gas with multiplier 2.0")
	})

	t.Run("multiplier with operation multiplier", func(t *testing.T) {
		t.Parallel()

		config := DefaultConfig()
		config.GlobalMultiplier = 2.0
		config.Costs[OpTesting] = 100
		meter := NewMeter(10000, config)

		// Operation multiplier 3, global multiplier 2
		// Expected: 100 * 3 * 2 = 600
		meter.ConsumeGas(OpTesting, 3)
		require.Equal(t, Gas(600), meter.GasConsumed(), "Should apply both multipliers")
	})

	t.Run("fractional result rounding", func(t *testing.T) {
		t.Parallel()

		config := DefaultConfig()
		config.GlobalMultiplier = 1.5
		config.Costs[OpTesting] = 10
		meter := NewMeter(10000, config)

		// 10 * 1 * 1.5 = 15.0 (rounds to 15)
		meter.ConsumeGas(OpTesting, 1)
		require.Equal(t, Gas(15), meter.GasConsumed(), "Should round correctly")

		// 10 * 1 * 1.5 = 15.0, so total is 30
		meter.ConsumeGas(OpTesting, 1)
		require.Equal(t, Gas(30), meter.GasConsumed(), "Should round correctly on second call")
	})
}
