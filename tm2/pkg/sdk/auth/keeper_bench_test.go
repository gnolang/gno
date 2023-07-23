package auth

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func BenchmarkAccountMapperGetAccountFound(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping testing in short mode")
	}

	env := setupTestEnv()

	// assumes b.N < 2**24
	for i := 0; i < b.N; i++ {
		addr := make([]byte, crypto.AddressSize)
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		copy(addr[:len(arr)], arr[:])
		caddr := crypto.AddressFromBytes(addr)
		acc := env.acck.NewAccountWithAddress(env.ctx, caddr)
		env.acck.SetAccount(env.ctx, acc)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		addr := make([]byte, crypto.AddressSize)
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		copy(addr[:len(arr)], arr[:])
		caddr := crypto.AddressFromBytes(addr)
		env.acck.GetAccount(env.ctx, caddr)
	}
}

func BenchmarkAccountMapperGetAccountFoundWithCoins(b *testing.B) {
	env := setupTestEnv()
	coins := std.Coins{
		std.NewCoin("ltc", int64(1000)),
		std.NewCoin("btc", int64(1000)),
		std.NewCoin("eth", int64(1000)),
		std.NewCoin("xrp", int64(1000)),
		std.NewCoin("bch", int64(1000)),
		std.NewCoin("eos", int64(1000)),
	}

	// assumes b.N < 2**24
	for i := 0; i < b.N; i++ {
		addr := make([]byte, crypto.AddressSize)
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		copy(addr[:len(arr)], arr[:])
		caddr := crypto.AddressFromBytes(addr)
		acc := env.acck.NewAccountWithAddress(env.ctx, caddr)
		acc.SetCoins(coins)
		env.acck.SetAccount(env.ctx, acc)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		addr := make([]byte, crypto.AddressSize)
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		copy(addr[:len(arr)], arr[:])
		caddr := crypto.AddressFromBytes(addr)
		env.acck.GetAccount(env.ctx, caddr)
	}
}

func BenchmarkAccountMapperSetAccount(b *testing.B) {
	env := setupTestEnv()

	b.ResetTimer()

	// assumes b.N < 2**24
	for i := 0; i < b.N; i++ {
		addr := make([]byte, crypto.AddressSize)
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		copy(addr[:len(arr)], arr[:])
		caddr := crypto.AddressFromBytes(addr)
		acc := env.acck.NewAccountWithAddress(env.ctx, caddr)
		env.acck.SetAccount(env.ctx, acc)
	}
}

func BenchmarkAccountMapperSetAccountWithCoins(b *testing.B) {
	env := setupTestEnv()
	coins := std.Coins{
		std.NewCoin("ltc", int64(1000)),
		std.NewCoin("btc", int64(1000)),
		std.NewCoin("eth", int64(1000)),
		std.NewCoin("xrp", int64(1000)),
		std.NewCoin("bch", int64(1000)),
		std.NewCoin("eos", int64(1000)),
	}

	b.ResetTimer()

	// assumes b.N < 2**24
	for i := 0; i < b.N; i++ {
		addr := make([]byte, crypto.AddressSize)
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		copy(addr[:len(arr)], arr[:])
		caddr := crypto.AddressFromBytes(addr)
		acc := env.acck.NewAccountWithAddress(env.ctx, caddr)
		acc.SetCoins(coins)
		env.acck.SetAccount(env.ctx, acc)
	}
}
