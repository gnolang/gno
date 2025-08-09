package my_mempool

import (
	"encoding/binary"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/bft/abci/example/kvstore"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/appconn"
	"github.com/gnolang/gno/tm2/pkg/bft/proxy"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
)

func benchmarkAddTxCommon(b *testing.B, batchSize int) {
	app := kvstore.NewKVStoreApplication()
	cc := proxy.NewLocalClientCreator(app)
	client, err := cc.NewABCIClient()
	if err != nil {
		b.Fatalf("failed to create ABCI client: %v", err)
	}

	if err := client.Start(); err != nil {
		b.Fatalf("failed to start ABCI client: %v", err)
	}
	defer client.Stop()

	appConnMem := appconn.NewMempool(client)

	// Set a no-op response callback for mempool ABCI requests
	appConnMem.SetResponseCallback(func(req abci.Request, res abci.Response) {})

	mempool := NewMempool(appConnMem)

	// Prepare a batch of transactions
	txs := make([]types.Tx, batchSize)
	for i := 0; i < batchSize; i++ {
		tx := make([]byte, 8)
		binary.BigEndian.PutUint64(tx, uint64(i))
		txs[i] = types.Tx(tx)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mempool.Flush() // Reset mempool between rounds
		for j := 0; j < batchSize; j++ {
			if err := mempool.AddTx(txs[j]); err != nil {
				b.Fatalf("AddTx failed: %v", err)
			}
		}
	}
}

// Benchmark functions for different batch sizes
func BenchmarkAddTx_10(b *testing.B)    { benchmarkAddTxCommon(b, 10) }
func BenchmarkAddTx_100(b *testing.B)   { benchmarkAddTxCommon(b, 100) }
func BenchmarkAddTx_1000(b *testing.B)  { benchmarkAddTxCommon(b, 1000) }
func BenchmarkAddTx_10000(b *testing.B) { benchmarkAddTxCommon(b, 10000) }

func benchmarkPendingCommon(b *testing.B, batchSize int) {
	app := kvstore.NewKVStoreApplication()
	cc := proxy.NewLocalClientCreator(app)
	client, err := cc.NewABCIClient()
	if err != nil {
		b.Fatalf("failed to create ABCI client: %v", err)
	}

	if err := client.Start(); err != nil {
		b.Fatalf("failed to start ABCI client: %v", err)
	}
	defer client.Stop()

	appConnMem := appconn.NewMempool(client)
	appConnMem.SetResponseCallback(func(req abci.Request, res abci.Response) {})

	mempool := NewMempool(appConnMem)

	// Add batchSize transactions to the mempool
	for i := 0; i < batchSize; i++ {
		tx := make([]byte, 8)
		binary.BigEndian.PutUint64(tx, uint64(i))
		if err := mempool.AddTx(types.Tx(tx)); err != nil {
			b.Fatalf("AddTx failed: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Call Pending with large limits to simulate fetching all transactions
		mempool.Pending(100000000, 10000000)
	}
}

// Benchmark functions for different batch sizes
func BenchmarkPending_10(b *testing.B)    { benchmarkPendingCommon(b, 10) }
func BenchmarkPending_100(b *testing.B)   { benchmarkPendingCommon(b, 100) }
func BenchmarkPending_1000(b *testing.B)  { benchmarkPendingCommon(b, 1000) }
func BenchmarkPending_10000(b *testing.B) { benchmarkPendingCommon(b, 10000) }

func benchmarkUpdateCommon(b *testing.B, batchSize int) {
	app := kvstore.NewKVStoreApplication()
	cc := proxy.NewLocalClientCreator(app)
	client, err := cc.NewABCIClient()
	if err != nil {
		b.Fatalf("failed to create ABCI client: %v", err)
	}
	if err := client.Start(); err != nil {
		b.Fatalf("failed to start ABCI client: %v", err)
	}
	defer client.Stop()

	appConnMem := appconn.NewMempool(client)
	appConnMem.SetResponseCallback(func(req abci.Request, res abci.Response) {})

	mempool := NewMempool(appConnMem)

	// Prepare and add batchSize transactions to the mempool
	txs := make([]types.Tx, batchSize)
	for i := 0; i < batchSize; i++ {
		tx := make([]byte, 8)
		binary.BigEndian.PutUint64(tx, uint64(i))
		txObj := types.Tx(tx)
		if err := mempool.AddTx(txObj); err != nil {
			b.Fatalf("AddTx failed: %v", err)
		}
		txs[i] = txObj
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Update mempool with the same batch of transactions
		mempool.Update(txs)
	}
}

// Benchmark functions for different batch sizes
func BenchmarkUpdate_10(b *testing.B)    { benchmarkUpdateCommon(b, 10) }
func BenchmarkUpdate_100(b *testing.B)   { benchmarkUpdateCommon(b, 100) }
func BenchmarkUpdate_1000(b *testing.B)  { benchmarkUpdateCommon(b, 1000) }
func BenchmarkUpdate_10000(b *testing.B) { benchmarkUpdateCommon(b, 10000) }
