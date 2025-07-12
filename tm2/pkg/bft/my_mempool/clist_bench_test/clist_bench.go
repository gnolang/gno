// This benchmark test was written to match the original CList mempool behavior
// as closely as possible, with recheck and cache explicitly disabled,
// to allow a fair and consistent performance comparison with the custom mempool.
//
// To run the comparison, simply copy this file into the official CList mempool folder
// and execute the benchmark. This allows you to measure the performance difference
// under identical conditions.
//
// The existing benchmark test in this folder already measures the performance
// of the custom mempool. This test provides the complementary side — measuring CList.

package mempool

import (
	"encoding/binary"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/bft/abci/example/kvstore"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/proxy"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
)

func benchmarkCheckTxCommon(b *testing.B, batchSize int) {
	app := kvstore.NewKVStoreApplication()
	cc := proxy.NewLocalClientCreator(app)
	mempool, cleanup := newMempoolWithApp(cc)
	defer cleanup()

	txs := make([]types.Tx, batchSize)
	for i := 0; i < batchSize; i++ {
		tx := make([]byte, 8)
		binary.BigEndian.PutUint64(tx, uint64(i))
		txs[i] = types.Tx(tx)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mempool.Flush()
		for j := 0; j < batchSize; j++ {
			if err := mempool.CheckTx(txs[j], nil); err != nil {
				b.Fatalf("CheckTx failed: %v", err)
			}
		}
	}
}

func BenchmarkCheckTx_10(b *testing.B)    { benchmarkCheckTxCommon(b, 10) }
func BenchmarkCheckTx_100(b *testing.B)   { benchmarkCheckTxCommon(b, 100) }
func BenchmarkCheckTx_1000(b *testing.B)  { benchmarkCheckTxCommon(b, 1000) }
func BenchmarkCheckTx_10000(b *testing.B) { benchmarkCheckTxCommon(b, 10000) }

func benchmarkReapCommon(b *testing.B, batchSize int) {
	app := kvstore.NewKVStoreApplication()
	cc := proxy.NewLocalClientCreator(app)
	mempool, cleanup := newMempoolWithApp(cc)
	defer cleanup()

	// Dodaj batchSize transakcija
	for i := 0; i < batchSize; i++ {
		tx := make([]byte, 8)
		binary.BigEndian.PutUint64(tx, uint64(i))
		if err := mempool.CheckTx(tx, nil); err != nil {
			b.Fatalf("CheckTx failed: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mempool.ReapMaxBytesMaxGas(100000000, 10000000)
	}
}

func BenchmarkReap_10(b *testing.B)    { benchmarkReapCommon(b, 10) }
func BenchmarkReap_100(b *testing.B)   { benchmarkReapCommon(b, 100) }
func BenchmarkReap_1000(b *testing.B)  { benchmarkReapCommon(b, 1000) }
func BenchmarkReap_10000(b *testing.B) { benchmarkReapCommon(b, 10000) }

func benchmarkUpdateCommon(b *testing.B, batchSize int) {
	app := kvstore.NewKVStoreApplication()
	cc := proxy.NewLocalClientCreator(app)
	mempool, cleanup := newMempoolWithApp(cc)
	defer cleanup()

	mempool.config.Recheck = false
	mempool.config.CacheSize = 0

	txs := make([]types.Tx, batchSize)
	for i := 0; i < batchSize; i++ {
		tx := make([]byte, 8)
		binary.BigEndian.PutUint64(tx, uint64(i))
		if err := mempool.CheckTx(tx, nil); err != nil {
			b.Fatalf("CheckTx failed: %v", err)
		}
		txs[i] = tx
	}

	deliverTxResponses := make([]abci.ResponseDeliverTx, batchSize)
	for i := 0; i < batchSize; i++ {
		deliverTxResponses[i] = abci.ResponseDeliverTx{}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := mempool.Update(1, txs, deliverTxResponses, nil, 0)
		if err != nil {
			b.Fatalf("Update failed: %v", err)
		}
	}
}

func BenchmarkUpdate_10(b *testing.B)    { benchmarkUpdateCommon(b, 10) }
func BenchmarkUpdate_100(b *testing.B)   { benchmarkUpdateCommon(b, 100) }
func BenchmarkUpdate_1000(b *testing.B)  { benchmarkUpdateCommon(b, 1000) }
func BenchmarkUpdate_10000(b *testing.B) { benchmarkUpdateCommon(b, 10000) }
