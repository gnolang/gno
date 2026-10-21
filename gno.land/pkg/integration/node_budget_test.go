package integration

import (
	"runtime"
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

// memAt returns a readMem stub reporting whatever *available holds, for tests
// that change it as they go.
func memAt(available *uint64) func() (testutils.MemInfo, bool) {
	return func() (testutils.MemInfo, bool) {
		return testutils.MemInfo{Total: 32 << 30, Available: *available}, true
	}
}

// grown is a readHeld stub for nodes that have all grown to full size, so that
// the reading is the whole story.
func grown() uint64 { return 1 << 40 }

// elapse back-dates the resize clock so the next resize is allowed, rather than
// sleeping out the read interval.
func elapse(b *nodeBudget) {
	b.lastCheck = time.Now().Add(-nodeReadInterval)
}

// settleCuts back-dates every cut so its memory counts as back, rather than
// sleeping out nodeCutSettle.
func settleCuts(b *nodeBudget) {
	for i := range b.cuts {
		b.cuts[i] = b.cuts[i].Add(-nodeCutSettle)
	}
}

func TestNodeBudgetMinIgnoresMemory(t *testing.T) {
	t.Parallel()

	// No memory free at all: the minimum must still be admitted, or a loaded
	// machine would never finish the suite.
	b := &nodeBudget{min: 2, max: 8, limit: 2, reserve: 8 << 30, readMem: memOf(0), readHeld: grown}
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
		readHeld:  grown,
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
		reserve:  8 << 30,
		readMem:  memOf(8<<30 + nodeMemCost - 1),
		readHeld: grown,
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
		reserve:  8 << 30,
		readMem:  memAt(&avail),
		readHeld: grown,
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

func TestNodeBudgetGrowthCountsOwedMemory(t *testing.T) {
	t.Parallel()

	// The start of a run: room for three more nodes by the reading, but the
	// two running have not grown yet and will take two of them.
	held := uint64(0)
	b := &nodeBudget{
		min: 2, max: 8, limit: 2,
		reserve:  8 << 30,
		readMem:  memOf(8<<30 + 3*nodeMemCost),
		readHeld: func() uint64 { return held },
	}
	for range 2 {
		require.True(t, b.tryAcquire())
	}
	elapse(b)
	require.True(t, b.tryAcquire(), "refused the one node there is room for")
	assert.Equal(t, 3, b.limit)

	// The reading has not moved, but three nodes are now owed the room it
	// shows: growing again would overshoot.
	elapse(b)
	assert.False(t, b.tryAcquire(), "grew on memory the running nodes still owe")
	assert.Equal(t, 3, b.limit)

	// Once they have grown and the reading still shows room, it does.
	held = 3 * nodeMemCost
	elapse(b)
	assert.True(t, b.tryAcquire(), "refused room the grown nodes left")
	assert.Equal(t, 4, b.limit)
}

func TestNodeBudgetCutsCountMemoryOnItsWayBack(t *testing.T) {
	t.Parallel()

	// Half a node short with six running: one cut covers it.
	avail := uint64(8<<30 - nodeMemCost/2)
	b := &nodeBudget{
		min: 2, max: 8, limit: 6, running: 6,
		reserve:  8 << 30,
		readMem:  memAt(&avail),
		readHeld: grown,
	}
	elapse(b)
	b.tryAcquire()
	assert.Equal(t, 5, b.limit)

	// The cut node's memory has yet to come back, so the reading has not
	// moved: cutting again would only overshoot.
	for range 3 {
		elapse(b)
		b.tryAcquire()
	}
	assert.Equal(t, 5, b.limit, "cut again while the last cut's memory was on its way back")

	// Nor does memory on its way back count towards growing again, even
	// with the room it will make.
	b.running = 5
	avail = 8<<30 + nodeMemCost/2
	elapse(b)
	b.tryAcquire()
	assert.Equal(t, 5, b.limit, "grew on memory not yet back")

	// Still short once it should have come back: cut again.
	avail = 8<<30 - nodeMemCost/2
	settleCuts(b)
	elapse(b)
	b.tryAcquire()
	assert.Equal(t, 4, b.limit)
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
	// Clear any developer override: it would pin the budget rather than let it ramp.
	t.Setenv(testutils.MaxParallelEnv, "")
	if _, ok := testutils.ReadMemInfo(); !ok {
		t.Skip("no memory reading on this platform")
	}
	// The budget sizes itself off GOMAXPROCS, so set it rather than depend on
	// the machine running the test. Safe: as a sequential test, which
	// t.Setenv requires, this one never runs alongside a parallel one.
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(0))

	t.Run("room to ramp", func(t *testing.T) {
		runtime.GOMAXPROCS(16)
		b := newNodeBudget()
		assert.Equal(t, 16, b.max)
		assert.NotZero(t, b.reserve, "should ramp where memory can be read")
		assert.NotNil(t, b.readHeld)
		// Starts at the static cap, and never shrinks below it: a busy machine
		// is never worse off than without the ramp.
		assert.Equal(t, testutils.MaxParallel(), b.limit)
		assert.Equal(t, testutils.MaxParallel(), b.min)
	})

	t.Run("no room to ramp", func(t *testing.T) {
		// No more cores than the static cap: nowhere to ramp to, so nothing
		// worth reading memory every second for.
		runtime.GOMAXPROCS(2)
		b := newNodeBudget()
		assert.Equal(t, 2, b.min)
		assert.Equal(t, 2, b.max)
		assert.Zero(t, b.reserve, "reads memory for a limit that cannot move")
	})
}

func TestNodeBudgetPacesReadings(t *testing.T) {
	t.Parallel()

	// Sitting inside the hysteresis band must not re-read on every poll; see
	// resize.
	var reads int
	b := &nodeBudget{
		min: 1, max: 8, limit: 1,
		reserve: 8 << 30,
		readMem: func() (testutils.MemInfo, bool) {
			reads++
			// In the band: above the reserve, no room for a whole node.
			return testutils.MemInfo{Total: 32 << 30, Available: 8<<30 + nodeMemCost - 1}, true
		},
		readHeld: grown,
	}
	for range 20 {
		b.tryAcquire()
	}
	assert.Equal(t, 1, reads, "re-read inside the read interval")
	assert.Equal(t, 1, b.limit, "limit moved while inside the band")
}
