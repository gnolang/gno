package auth

import (
	"testing"

	"github.com/gnolang/gno/pkgs/crypto"
	"github.com/gnolang/gno/pkgs/std"
)

func BenchmarkAccountMapperGetAccountFound(b *testing.B) {
	input := setupTestInput()

	// assumes b.N < 2**24
	for i := 0; i < b.N; i++ {
		addr := make([]byte, crypto.AddressSize)
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		copy(addr[:len(arr)], arr[:])
		caddr := crypto.AddressFromBytes(addr)
		acc := input.acck.NewAccountWithAddress(input.ctx, caddr)
		input.acck.SetAccount(input.ctx, acc)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		addr := make([]byte, crypto.AddressSize)
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		copy(addr[:len(arr)], arr[:])
		caddr := crypto.AddressFromBytes(addr)
		input.acck.GetAccount(input.ctx, caddr)
	}
}

func BenchmarkAccountMapperGetAccountFoundWithCoins(b *testing.B) {
	input := setupTestInput()
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
		acc := input.acck.NewAccountWithAddress(input.ctx, caddr)
		acc.SetCoins(coins)
		input.acck.SetAccount(input.ctx, acc)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		addr := make([]byte, crypto.AddressSize)
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		copy(addr[:len(arr)], arr[:])
		caddr := crypto.AddressFromBytes(addr)
		input.acck.GetAccount(input.ctx, caddr)
	}
}

func BenchmarkAccountMapperSetAccount(b *testing.B) {
	input := setupTestInput()

	b.ResetTimer()

	// assumes b.N < 2**24
	for i := 0; i < b.N; i++ {
		addr := make([]byte, crypto.AddressSize)
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		copy(addr[:len(arr)], arr[:])
		caddr := crypto.AddressFromBytes(addr)
		acc := input.acck.NewAccountWithAddress(input.ctx, caddr)
		input.acck.SetAccount(input.ctx, acc)
	}
}

func BenchmarkAccountMapperSetAccountWithCoins(b *testing.B) {
	input := setupTestInput()
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
		acc := input.acck.NewAccountWithAddress(input.ctx, caddr)
		acc.SetCoins(coins)
		input.acck.SetAccount(input.ctx, acc)
	}
}
