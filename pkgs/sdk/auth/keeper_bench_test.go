package auth

import (
	"testing"

	"github.com/gnolang/gno/pkgs/crypto"
	"github.com/gnolang/gno/pkgs/std"
)

func BenchmarkAccountMapperGetAccountFound(b *testing.B) {
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
		std.NewCoin("LTC", int64(1000)),
		std.NewCoin("BTC", int64(1000)),
		std.NewCoin("ETH", int64(1000)),
		std.NewCoin("XRP", int64(1000)),
		std.NewCoin("BCH", int64(1000)),
		std.NewCoin("EOS", int64(1000)),
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
		std.NewCoin("LTC", int64(1000)),
		std.NewCoin("BTC", int64(1000)),
		std.NewCoin("ETH", int64(1000)),
		std.NewCoin("XRP", int64(1000)),
		std.NewCoin("BCH", int64(1000)),
		std.NewCoin("EOS", int64(1000)),
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
