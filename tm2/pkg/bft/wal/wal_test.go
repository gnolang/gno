package wal

import (
	"bytes"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	auto "github.com/gnolang/gno/tm2/pkg/autofile"
	tmtime "github.com/gnolang/gno/tm2/pkg/bft/types/time"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/random"
)

const (
	walTestFlushInterval = time.Duration(100) * time.Millisecond
)

type TestMessage struct {
	Duration time.Duration
	Height   int64
	Round    int64
	Data     []byte
}

func (TestMessage) AssertWALMessage() {}

var _ = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/tm2/pkg/bft/wal",
	"wal",
	amino.GetCallersDirname(),
).
	WithTypes(
		TestMessage{},
	))

func TestWALWriterReader(t *testing.T) {
	t.Parallel()

	now := tmtime.Now()
	msgs := []TimedWALMessage{
		{Time: now, Msg: TestMessage{Duration: time.Second, Height: 1, Round: 1}},
		{Time: now, Msg: TestMessage{Duration: time.Second, Height: 1, Round: 2}},
	}

	b := new(bytes.Buffer)

	for _, msg := range msgs {
		msg := msg

		b.Reset()

		enc := NewWALWriter(b, maxTestMsgSize)
		err := enc.Write(msg)
		require.NoError(t, err)

		dec := NewWALReader(b, maxTestMsgSize)
		decoded, meta, err := dec.ReadMessage()
		require.NoError(t, err)
		require.Nil(t, meta)

		assert.Equal(t, msg.Time.UTC(), decoded.Time)
		assert.Equal(t, msg.Msg, decoded.Msg)
	}
}

const maxTestMsgSize int64 = 64 * 1024

func makeTempWAL(t *testing.T, maxMsgSize int64, walChunkSize int64) (wal *baseWAL) {
	t.Helper()

	// Create WAL file.
	walFile := filepath.Join(t.TempDir(), "wal")

	// Create WAL.
	wal, err := NewWAL(walFile, maxTestMsgSize, auto.GroupHeadSizeLimit(walChunkSize))
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

func TestWALWrite(t *testing.T) {
	t.Parallel()

	// Create WAL
	const walChunkSize = 100000
	wal := makeTempWAL(t, maxTestMsgSize, walChunkSize)

	// 1) Write returns an error if msg is too big
	msg := TestMessage{
		Data: random.RandBytes(int(maxTestMsgSize)),
	}
	err := wal.Write(msg)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "msg is too big")
	}

	// 2) Write returns no error if msg is not too big.
	overhead := 1024 // sufficiently large.
	msg = TestMessage{
		Data: random.RandBytes(int(maxTestMsgSize) - overhead),
	}
	err = wal.Write(msg)
	assert.NoError(t, err)
}

func TestWALSearchForHeight(t *testing.T) {
	t.Parallel()

	// Create WAL
	const numHeight, numRounds, dataSize = 100, 10000, 10
	const walChunkSize = 100000
	if numHeight*numRounds*dataSize < walChunkSize*3 {
		panic("invalid walChunkSize, it should be an order of magnitude or more smaller than the product")
	}
	wal := makeTempWAL(t, maxTestMsgSize, walChunkSize)

	// Generate WAL messages.
	for h := 1; h < numHeight; h++ {
		err := wal.WriteMetaSync(MetaMessage{Height: int64(h)})
		assert.NoError(t, err)
		for r := 1; r < numRounds; r++ {
			err := wal.Write(TestMessage{Height: int64(h), Round: int64(r), Data: random.RandBytes(dataSize)})
			assert.NoError(t, err)
		}
	}
	wal.FlushAndSync()

	// Search for height.
	for h := 1; h < numHeight; h++ {
		// Search for h.
		gr, found, err := wal.SearchForHeight(int64(h), nil)
		assert.NoError(t, err, "expected not to err on height %d", h)
		assert.True(t, found, "expected to find end height for %d", h)
		assert.NotNil(t, gr)

		// Read next message.
		dec := NewWALReader(gr, maxTestMsgSize)
		msg, meta, err := dec.ReadMessage()
		assert.NoError(t, err, "expected to decode a message")
		assert.Nil(t, meta, "expected no meta")
		rs, ok := msg.Msg.(TestMessage)
		assert.True(t, ok, "expected message of type TestMessage")
		assert.Equal(t, rs.Height, int64(h), "wrong height")

		// Cleanup
		dec.Close()
	}
}

func TestWALPeriodicSync(t *testing.T) {
	t.Parallel()

	// Create WAL
	const numHeight, numRounds, dataSize = 100, 10000, 10
	const walChunkSize = 100000
	const sleepInterval = 5 * time.Second // NOTE: no longer needed.
	if numHeight*numRounds*dataSize < walChunkSize {
		panic("invalid walChunkSize, it should be an order of magnitude or more smaller than the product")
	}
	wal := makeTempWAL(t, maxTestMsgSize, walChunkSize)

	// Is this needed?
	wal.SetFlushInterval(walTestFlushInterval)
	wal.SetLogger(log.NewNoopLogger())

	// Take snapshot of starting state.
	startInfo := wal.Group().ReadGroupInfo()
	assert.True(t, startInfo.TotalSize < 1024, "WAL should start short (w/ initial Height 1 meta)")

	// Generate WAL messages.
	for h := 1; h < numHeight; h++ {
		err := wal.WriteMetaSync(MetaMessage{Height: int64(h)})
		assert.NoError(t, err)
		for r := 1; r < numRounds; r++ {
			err := wal.Write(TestMessage{Height: int64(h), Round: int64(r), Data: random.RandBytes(dataSize)})
			assert.NoError(t, err)
		}
	}
	// NOTE: but this isn't guaranteed so don't test like below:
	// assert.NotZero(t, wal.Group().Buffered())
	wal.FlushAndSync()

	// Sleep for a while, while WAL files are being created.
	time.Sleep(sleepInterval)

	// Take snapshot of ending state.
	endInfo := wal.Group().ReadGroupInfo()
	assert.NotEqual(t, 0, endInfo.TotalSize, "WAL should end not empty")

	// The data should have been flushed by the periodic sync
	assert.Zero(t, wal.Group().Buffered())

	// Try searching for the last height.
	h := int64(numHeight - 1)
	gr, found, err := wal.SearchForHeight(h, nil)
	assert.NoError(t, err, "expected not to err on height %d", h)
	assert.True(t, found, "expected to find end height for %d", h)
	assert.NotNil(t, gr)
	if gr != nil {
		gr.Close()
	}
}

/*
var initOnce sync.Once

func registerInterfacesOnce() {
	initOnce.Do(func() {
		var _ = wire.RegisterInterface(
			struct{ WALMessage }{},
			wire.ConcreteType{[]byte{}, 0x10},
		)
	})
}
*/

func benchmarkWalRead(b *testing.B, n int) {
	// registerInterfacesOnce()
	b.Helper()

	buf := new(bytes.Buffer)
	enc := NewWALWriter(buf, int64(n)+64) // n + overhead.

	msg := TestMessage{
		Height: 1,
		Round:  1,
		Data:   random.RandBytes(n),
	}
	enc.Write(TimedWALMessage{Msg: msg, Time: time.Now().Round(time.Second).UTC()})

	encoded := buf.Bytes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		buf.Write(encoded)
		dec := NewWALReader(buf, maxTestMsgSize)
		if _, _, err := dec.ReadMessage(); err != nil {
			b.Fatal(err)
		}
	}
	b.ReportAllocs()
}

func BenchmarkWalRead512B(b *testing.B) {
	benchmarkWalRead(b, 512)
}

func BenchmarkWalRead10KB(b *testing.B) {
	benchmarkWalRead(b, 10*1024)
}

func BenchmarkWalRead100KB(b *testing.B) {
	b.Skip("TODO: benchmark failing")
	benchmarkWalRead(b, 100*1024)
}

func BenchmarkWalRead1MB(b *testing.B) {
	b.Skip("TODO: benchmark failing")
	benchmarkWalRead(b, 1024*1024)
}

func BenchmarkWalRead10MB(b *testing.B) {
	b.Skip("TODO: benchmark failing")
	benchmarkWalRead(b, 10*1024*1024)
}

func BenchmarkWalRead100MB(b *testing.B) {
	b.Skip("TODO: benchmark failing")
	benchmarkWalRead(b, 100*1024*1024)
}

func BenchmarkWalRead1GB(b *testing.B) {
	b.Skip("TODO: benchmark failing")
	benchmarkWalRead(b, 1024*1024*1024)
}
