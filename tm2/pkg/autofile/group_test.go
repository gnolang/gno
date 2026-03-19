package autofile

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/random"
)

func createTestGroupWithOptions(t *testing.T, opts ...func(*Group)) *Group {
	t.Helper()

	testDir := t.TempDir()
	headPath := testDir + "/myfile"
	g, err := OpenGroup(headPath, opts...)
	require.NoError(t, err, "Error opening Group")
	require.NotNil(t, g, "Failed to create Group")

	return g
}

func assertGroupInfo(t *testing.T, gInfo GroupInfo, minIndex, maxIndex int, totalSize, headSize int64) {
	t.Helper()

	assert.Equal(t, minIndex, gInfo.MinIndex)
	assert.Equal(t, maxIndex, gInfo.MaxIndex)
	assert.Equal(t, totalSize, gInfo.TotalSize)
	assert.Equal(t, headSize, gInfo.HeadSize)
}

func TestCheckHeadSizeLimit(t *testing.T) {
	t.Parallel()

	g := createTestGroupWithOptions(t, GroupHeadSizeLimit(1000*1000))
	defer g.Close()

	// At first, there are no files.
	assertGroupInfo(t, g.ReadGroupInfo(), 0, 0, 0, 0)

	// Write 1000 bytes 999 times.
	for range 999 {
		err := g.WriteLine(random.RandStr(999))
		require.NoError(t, err, "Error appending to head")
	}
	g.FlushAndSync()
	assertGroupInfo(t, g.ReadGroupInfo(), 0, 0, 999000, 999000)

	// Write 1000 more bytes.
	err := g.WriteLine(random.RandStr(999))
	require.NoError(t, err, "Error appending to head")
	g.FlushAndSync()
	assertGroupInfo(t, g.ReadGroupInfo(), 0, 1, 1000000, 0)

	// Write 1000 more bytes.
	err = g.WriteLine(random.RandStr(999))
	require.NoError(t, err, "Error appending to head")
	g.FlushAndSync()

	// Should not have rotated
	assertGroupInfo(t, g.ReadGroupInfo(), 0, 1, 1001000, 1000)

	// Write 1000 bytes 999 times.
	for range 999 {
		err = g.WriteLine(random.RandStr(999))
		require.NoError(t, err, "Error appending to head")
	}
	g.FlushAndSync()
	assertGroupInfo(t, g.ReadGroupInfo(), 0, 2, 2000000, 0)

	// Write 1000 more bytes.
	_, err = g.Head.Write([]byte(random.RandStr(999) + "\n"))
	require.NoError(t, err, "Error appending to head")
	g.FlushAndSync()
	assertGroupInfo(t, g.ReadGroupInfo(), 0, 2, 2001000, 1000)
}

func TestRotateFile(t *testing.T) {
	t.Parallel()

	g := createTestGroupWithOptions(t, GroupHeadSizeLimit(0))
	defer g.Close()

	g.WriteLine("Line 1")
	g.WriteLine("Line 2")
	g.WriteLine("Line 3")
	g.FlushAndSync()
	g.RotateFile()
	g.WriteLine("Line 4")
	g.WriteLine("Line 5")
	g.WriteLine("Line 6")
	g.FlushAndSync()

	// Read g.Head.Path+"000"
	body1, err := os.ReadFile(g.Head.Path + ".000")
	assert.NoError(t, err, "Failed to read first rolled file")
	if string(body1) != "Line 1\nLine 2\nLine 3\n" {
		t.Errorf("Got unexpected contents: [%v]", string(body1))
	}

	// Read g.Head.Path
	body2, err := os.ReadFile(g.Head.Path)
	assert.NoError(t, err, "Failed to read first rolled file")
	if string(body2) != "Line 4\nLine 5\nLine 6\n" {
		t.Errorf("Got unexpected contents: [%v]", string(body2))
	}
}

func TestWrite(t *testing.T) {
	t.Parallel()

	g := createTestGroupWithOptions(t, GroupHeadSizeLimit(0))
	defer g.Close()

	written := []byte("Medusa")
	g.Write(written)
	g.FlushAndSync()

	read := make([]byte, len(written))
	gr, err := g.NewReader(0, 0)
	require.NoError(t, err, "failed to create reader")

	_, err = gr.Read(read)
	assert.NoError(t, err, "failed to read data")
	assert.Equal(t, written, read)
}

// test that Read reads the required amount of bytes from all the files in the
// group and returns no error if n == size of the given slice.
func TestGroupReaderRead(t *testing.T) {
	t.Parallel()

	g := createTestGroupWithOptions(t, GroupHeadSizeLimit(0))
	defer g.Close()

	professor := []byte("Professor Monster")
	g.Write(professor)
	g.FlushAndSync()
	g.RotateFile()
	frankenstein := []byte("Frankenstein's Monster")
	g.Write(frankenstein)
	g.FlushAndSync()

	totalWrittenLength := len(professor) + len(frankenstein)
	read := make([]byte, totalWrittenLength)
	gr, err := g.NewReader(0, 0)
	require.NoError(t, err, "failed to create reader")

	n, err := gr.Read(read)
	assert.NoError(t, err, "failed to read data")
	assert.Equal(t, totalWrittenLength, n, "not enough bytes read")
	professorPlusFrankenstein := professor
	professorPlusFrankenstein = append(professorPlusFrankenstein, frankenstein...)
	assert.Equal(t, professorPlusFrankenstein, read)
}

// test that Read returns an error if number of bytes read < size of
// the given slice. Subsequent call should return 0, io.EOF.
func TestGroupReaderRead2(t *testing.T) {
	t.Parallel()

	g := createTestGroupWithOptions(t, GroupHeadSizeLimit(0))
	defer g.Close()

	professor := []byte("Professor Monster")
	g.Write(professor)
	g.FlushAndSync()
	g.RotateFile()
	frankenstein := []byte("Frankenstein's Monster")
	frankensteinPart := []byte("Frankenstein")
	g.Write(frankensteinPart) // note writing only a part
	g.FlushAndSync()

	totalLength := len(professor) + len(frankenstein)
	read := make([]byte, totalLength)
	gr, err := g.NewReader(0, 0)
	require.NoError(t, err, "failed to create reader")

	// 1) n < (size of the given slice), io.EOF
	n, err := gr.Read(read)
	assert.Equal(t, io.EOF, err)
	assert.Equal(t, len(professor)+len(frankensteinPart), n, "Read more/less bytes than it is in the group")

	// 2) 0, io.EOF
	n, err = gr.Read([]byte("0"))
	assert.Equal(t, io.EOF, err)
	assert.Equal(t, 0, n)
}

func TestMinIndex(t *testing.T) {
	t.Parallel()

	g := createTestGroupWithOptions(t, GroupHeadSizeLimit(0))
	defer g.Close()

	assert.Zero(t, g.MinIndex(), "MinIndex should be zero at the beginning")
}

func TestMaxIndex(t *testing.T) {
	t.Parallel()

	g := createTestGroupWithOptions(t, GroupHeadSizeLimit(0))
	defer g.Close()

	assert.Zero(t, g.MaxIndex(), "MaxIndex should be zero at the beginning")

	g.WriteLine("Line 1")
	g.FlushAndSync()
	g.RotateFile()

	assert.Equal(t, 1, g.MaxIndex(), "MaxIndex should point to the last file")
}

func TestHaltedGroupRejectsWrite(t *testing.T) {
	t.Parallel()

	// Use a very large limit so the group stays halted (the real disk has
	// less space than maxUint64).
	g := createTestGroupWithOptions(t,
		GroupMinDiskSpaceLimit(1<<62),
	)
	defer g.Close()

	// Manually set halted to true (simulating disk space exhaustion).
	g.mtx.Lock()
	g.halted = true
	g.mtx.Unlock()

	// Write should fail with ErrDiskSpaceUnavailable.
	_, err := g.Write([]byte("hello"))
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDiskSpaceUnavailable)

	// WriteLine should also fail.
	err = g.WriteLine("hello")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDiskSpaceUnavailable)
}

func TestHaltedReturnsTrueAfterHalt(t *testing.T) {
	t.Parallel()

	g := createTestGroupWithOptions(t, GroupMinDiskSpaceLimit(0))
	defer g.Close()

	assert.False(t, g.Halted(), "group should not be halted initially")

	// Manually halt.
	g.mtx.Lock()
	g.halted = true
	g.mtx.Unlock()

	assert.True(t, g.Halted(), "group should report as halted")
}

func TestDiskSpaceCheckDisabledWhenLimitZero(t *testing.T) {
	t.Parallel()

	g := createTestGroupWithOptions(t, GroupMinDiskSpaceLimit(0))
	defer g.Close()

	// With limit 0, disk space checks are disabled; writes should succeed.
	_, err := g.Write([]byte("data"))
	require.NoError(t, err)

	err = g.WriteLine("more data")
	require.NoError(t, err)

	assert.False(t, g.Halted())
}

func TestWriteSucceedsWithSufficientDiskSpace(t *testing.T) {
	t.Parallel()

	// Use a very small minimum disk space limit so the test always passes
	// on machines with any reasonable amount of free space.
	g := createTestGroupWithOptions(t,
		GroupHeadSizeLimit(0),
		GroupMinDiskSpaceLimit(1), // 1 byte
	)
	defer g.Close()

	_, err := g.Write([]byte("hello"))
	require.NoError(t, err)

	err = g.WriteLine("world")
	require.NoError(t, err)

	assert.False(t, g.Halted())
}

func TestGroupMinDiskSpaceLimitOption(t *testing.T) {
	t.Parallel()

	g := createTestGroupWithOptions(t, GroupMinDiskSpaceLimit(42))
	defer g.Close()

	g.mtx.Lock()
	limit := g.minDiskSpaceLimit
	g.mtx.Unlock()

	assert.Equal(t, int64(42), limit)
}

func TestDefaultMinDiskSpaceLimit(t *testing.T) {
	t.Parallel()

	g := createTestGroupWithOptions(t)
	defer g.Close()

	g.mtx.Lock()
	limit := g.minDiskSpaceLimit
	g.mtx.Unlock()

	assert.Equal(t, int64(defaultMinDiskSpaceLimit), limit,
		"default min disk space limit should be set")
}

func TestResumeUnhaltsGroup(t *testing.T) {
	t.Parallel()

	g := createTestGroupWithOptions(t,
		GroupMinDiskSpaceLimit(1), // 1 byte — will pass on any real FS
	)
	defer g.Close()

	// Manually halt.
	g.mtx.Lock()
	g.halted = true
	g.mtx.Unlock()
	assert.True(t, g.Halted())

	// Resume should un-halt.
	g.Resume()
	assert.False(t, g.Halted())

	// Writes should succeed after resume (there's plenty of space with 1 byte limit).
	_, err := g.Write([]byte("after resume"))
	require.NoError(t, err)
}

func TestAutoRecoveryWhenSpaceFreed(t *testing.T) {
	t.Parallel()

	// Set a tiny limit (1 byte) so the real disk always has enough space.
	// Manually halt and then attempt a write — the periodic re-check
	// should see sufficient space and auto-resume.
	g := createTestGroupWithOptions(t,
		GroupMinDiskSpaceLimit(1),
	)
	defer g.Close()

	// Manually halt.
	g.mtx.Lock()
	g.halted = true
	g.mtx.Unlock()
	assert.True(t, g.Halted())

	// Next write should trigger a re-check, see sufficient space, and auto-resume.
	_, err := g.Write([]byte("recovered"))
	require.NoError(t, err, "write should succeed after auto-recovery")
	assert.False(t, g.Halted(), "group should have auto-resumed")
}

func TestWriteCountThrottling(t *testing.T) {
	t.Parallel()

	// The check interval is diskSpaceCheckInterval (a constant).
	// With a tiny limit the check always passes.
	g := createTestGroupWithOptions(t,
		GroupMinDiskSpaceLimit(1),
	)
	defer g.Close()

	// Write diskSpaceCheckInterval+1 times — all should succeed.
	for i := range diskSpaceCheckInterval + 1 {
		_, err := g.Write([]byte("data"))
		require.NoError(t, err, "write %d should succeed", i)
	}

	g.mtx.Lock()
	counter := g.writesSinceLastCheck
	g.mtx.Unlock()

	// After diskSpaceCheckInterval+1 writes: the first N writes
	// increment the counter to N, then the (N+1)th write triggers a check
	// (resetting to 0) and increments to 1.
	assert.Equal(t, 1, counter,
		"counter should reflect writes since last check")
}
