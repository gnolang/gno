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

func TestAutoRecoveryWhenSpaceFreed(t *testing.T) {
	t.Parallel()

	// The constant minDiskSpaceLimit is 16 MB, which is well below the
	// free space on any reasonable CI/dev machine. Manually halting the
	// group and then writing should trigger a re-check that finds
	// sufficient space and auto-resumes.
	g := createTestGroupWithOptions(t)
	defer g.Close()

	// Manually halt.
	g.mtx.Lock()
	g.halted = true
	g.mtx.Unlock()

	// Next write should trigger a re-check, see sufficient space, and auto-resume.
	_, err := g.Write([]byte("recovered"))
	require.NoError(t, err, "write should succeed after auto-recovery")

	g.mtx.Lock()
	halted := g.halted
	g.mtx.Unlock()
	assert.False(t, halted, "group should have auto-resumed")
}

func TestWriteCountThrottling(t *testing.T) {
	t.Parallel()

	// The check interval is diskSpaceCheckInterval (a constant).
	// With minDiskSpaceLimit at 16 MB the check always passes on real disks.
	g := createTestGroupWithOptions(t)
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
