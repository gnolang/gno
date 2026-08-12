package integration

import (
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// memOf returns a readMem stub reporting a fixed amount of available memory.
func memOf(available uint64) func() (testutils.MemInfo, bool) {
	return func() (testutils.MemInfo, bool) {
		return testutils.MemInfo{Total: 32 << 30, Available: available}, true
	}
}

// elapse back-dates the resize clock so the next resize is allowed, rather than
// sleeping out the settle window.
func elapse(b *nodeBudget) {
	b.lastCheck = time.Now().Add(-nodeStartSettle)
}

func TestNodeBudgetMinIgnoresMemory(t *testing.T) {
	t.Parallel()

	// No memory free at all: the minimum must still be admitted, or a loaded
	// machine would never finish the suite.
	b := &nodeBudget{min: 2, max: 8, limit: 2, reserve: 8 << 30, readMem: memOf(0)}
	for i := range 2 {
		require.True(t, b.tryAcquire(), "minimum node %d refused", i+1)
	}
	elapse(b)
	assert.False(t, b.tryAcquire(), "ramped past the minimum with no memory free")
	assert.Equal(t, 2, b.running)
	assert.Equal(t, 2, b.limit, "limit must not fall below min")
}

func TestNodeBudgetRampsWithHeadroom(t *testing.T) {
	t.Parallel()

	b := &nodeBudget{
		min: 2, max: 5, limit: 2,
		lastCheck: time.Now(), // start inside the settle window
		reserve:   8 << 30,
		readMem:   memOf(8<<30 + nodeMemCost), // room for exactly one more node
	}
	for range 2 {
		require.True(t, b.tryAcquire())
	}
	for want := 3; want <= 5; want++ {
		assert.False(t, b.tryAcquire(), "grew inside the settle window")
		elapse(b)
		require.True(t, b.tryAcquire(), "refused node %d with headroom to spare", want)
		assert.Equal(t, want, b.running)
	}

	// max is a hard ceiling however much memory is free.
	elapse(b)
	assert.False(t, b.tryAcquire(), "admitted past max")
	assert.Equal(t, 5, b.limit)
}

func TestNodeBudgetHoldsBetweenThresholds(t *testing.T) {
	t.Parallel()

	// Above the reserve, but without room for a whole further node: the limit
	// should sit still rather than oscillate.
	b := &nodeBudget{
		min: 1, max: 8, limit: 3,
		reserve: 8 << 30,
		readMem: memOf(8<<30 + nodeMemCost - 1),
	}
	for range 3 {
		require.True(t, b.tryAcquire())
	}
	elapse(b)
	assert.False(t, b.tryAcquire(), "ate into the reserve")
	assert.Equal(t, 3, b.limit, "limit moved while inside the hysteresis band")
}

func TestNodeBudgetShrinksUnderPressure(t *testing.T) {
	t.Parallel()

	// Ramped up, then the machine loses memory to something else.
	avail := uint64(8<<30 + nodeMemCost)
	b := &nodeBudget{
		min: 2, max: 8, limit: 6,
		reserve: 8 << 30,
		readMem: func() (testutils.MemInfo, bool) {
			return testutils.MemInfo{Total: 32 << 30, Available: avail}, true
		},
	}
	avail = 1 << 30 // now below the reserve
	for want := 5; want >= 2; want-- {
		elapse(b)
		b.tryAcquire()
		assert.Equal(t, want, b.limit)
	}
	// Never below min, however tight things get.
	elapse(b)
	b.tryAcquire()
	assert.Equal(t, 2, b.limit, "shrank past min")
}

func TestNodeBudgetStaticWhenMemoryUnknown(t *testing.T) {
	t.Parallel()

	// reserve == 0 marks a platform whose free memory we cannot read: the
	// allowance is then whatever min says, and never moves.
	b := &nodeBudget{min: 4, max: 16, limit: 4, readMem: func() (testutils.MemInfo, bool) {
		return testutils.MemInfo{}, false
	}}
	for range 4 {
		require.True(t, b.tryAcquire())
	}
	elapse(b)
	assert.False(t, b.tryAcquire(), "ramped without a memory reading")
	assert.Equal(t, 4, b.limit)
}

func TestNodeBudgetReleaseFreesRoom(t *testing.T) {
	t.Parallel()

	b := &nodeBudget{min: 1, max: 1, limit: 1, readMem: memOf(0)}
	require.True(t, b.tryAcquire())
	require.False(t, b.tryAcquire())
	b.release()
	assert.Equal(t, 0, b.running)
	assert.True(t, b.tryAcquire(), "slot not reusable after release")
}

func TestNewNodeBudgetHonoursOverride(t *testing.T) {
	t.Setenv(testutils.MaxParallelEnv, "3")

	b := newNodeBudget()
	assert.Equal(t, 3, b.min)
	assert.Equal(t, 3, b.max)
	assert.Equal(t, 3, b.limit)
	assert.Zero(t, b.reserve, "override should pin the budget, not ramp")
}

func TestNewNodeBudgetDefaults(t *testing.T) {
	// Not parallel: reads the ambient environment.
	if _, ok := testutils.ReadMemInfo(); !ok {
		t.Skip("no memory reading on this platform")
	}
	b := newNodeBudget()
	assert.Equal(t, nodeMinParallel, b.min)
	assert.GreaterOrEqual(t, b.max, b.min)
	assert.NotZero(t, b.reserve, "should ramp where memory can be read")
	// Starts at the static cap, so no platform is slower than before ramping.
	assert.Equal(t, max(testutils.MaxParallel(), nodeMinParallel), b.limit)
	assert.GreaterOrEqual(t, b.limit, b.min)
	assert.LessOrEqual(t, b.limit, b.max)
}

func TestNodeBudgetPacesReadings(t *testing.T) {
	t.Parallel()

	// Sitting inside the hysteresis band must not re-read on every poll: a
	// blocked script polls five times a second, and reading costs a subprocess
	// on darwin.
	var reads int
	b := &nodeBudget{
		min: 1, max: 8, limit: 1,
		reserve: 8 << 30,
		readMem: func() (testutils.MemInfo, bool) {
			reads++
			// In the band: above the reserve, no room for a whole node.
			return testutils.MemInfo{Total: 32 << 30, Available: 8<<30 + nodeMemCost - 1}, true
		},
	}
	for range 20 {
		b.tryAcquire()
	}
	assert.Equal(t, 1, reads, "re-read inside the settle window")
	assert.Equal(t, 1, b.limit, "limit moved while inside the band")
}
