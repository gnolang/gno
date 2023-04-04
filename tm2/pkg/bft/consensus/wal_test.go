package consensus

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	auto "github.com/gnolang/gno/tm2/pkg/autofile"
	walm "github.com/gnolang/gno/tm2/pkg/bft/wal"
	"github.com/gnolang/gno/tm2/pkg/log"
)

const (
	walTestFlushInterval = time.Duration(100) * time.Millisecond
)

// ----------------------------------------
// copied over from wal/wal_test.go

const maxTestMsgSize int64 = 64 * 1024

func makeTempWAL(t *testing.T, walChunkSize int64) (wal walm.WAL) {
	t.Helper()

	// Create WAL file.
	walFile := filepath.Join(t.TempDir(), "wal")

	// Create WAL.
	wal, err := walm.NewWAL(walFile, maxTestMsgSize, auto.GroupHeadSizeLimit(walChunkSize))
	if err != nil {
		panic(err)
	}
	err = wal.Start()
	if err != nil {
		panic(err)
	}

	t.Cleanup(func() {
		// WAL cleanup.
		wal.Stop()
		// wait for the wal to finish shutting down so we
		// can safely remove the directory
		wal.Wait()
	})

	return wal
}

// end copy from wal/wal_test.go
// ----------------------------------------

func TestWALTruncate(t *testing.T) {
	const maxTestMsgSize = 1024 * 1024 // 1MB
	const walChunkSize = 409610        // 4KB
	wal := makeTempWAL(t, walChunkSize)

	wal.SetLogger(log.TestingLogger())

	type grouper interface {
		Group() *auto.Group
	}

	// 60 block's size nearly 70K, greater than group's wal chunk filesize (4KB).
	// When the headBuf is full, content will flush to the filesystem.
	err := WALGenerateNBlocks(t, wal.(grouper).Group(), 60)
	require.NoError(t, err)

	time.Sleep(1 * time.Millisecond) // wait groupCheckDuration, make sure RotateFile run

	wal.FlushAndSync()

	h := int64(50)
	gr, found, err := wal.SearchForHeight(h+1, &walm.WALSearchOptions{})
	assert.NoError(t, err, "expected not to err on height %d", h)
	assert.True(t, found, "expected to find end height for %d", h)
	assert.NotNil(t, gr)
	defer gr.Close()

	dec := walm.NewWALReader(gr, maxMsgSize)
	msg, meta, err := dec.ReadMessage()
	assert.NoError(t, err, "expected to decode a message")
	rs, ok := msg.Msg.(newRoundStepInfo)
	assert.Nil(t, meta, "expected no meta")
	assert.True(t, ok, "expected message of type EventRoundState")
	assert.Equal(t, rs.Height, h+1, "wrong height")
}
