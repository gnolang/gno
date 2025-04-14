package std

import (
	"fmt"
	"testing"
)

func BenchmarkCoinsAdditionIntersect(b *testing.B) {
	benchmarkingFunc := func(numCoinsA int, numCoinsB int) func(b *testing.B) {
		return func(b *testing.B) {
			b.Helper()

			coinsA := Coins(make([]Coin, numCoinsA))
			coinsB := Coins(make([]Coin, numCoinsB))

			maxCoins := max(numCoinsA, numCoinsB)
			denomLength := len(fmt.Sprint(maxCoins))

			for i := range numCoinsA {
				denom := fmt.Sprintf("coinz_%0*d", denomLength, i)
				coinsA[i] = NewCoin(denom, int64(i+1))
			}
			for i := range numCoinsB {
				denom := fmt.Sprintf("coinz_%0*d", denomLength, i)
				coinsB[i] = NewCoin(denom, int64(i+1))
			}

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				coinsA.Add(coinsB)
			}
		}
	}

	benchmarkSizes := [][]int{{1, 1}, {5, 5}, {5, 20}, {1, 1000}, {2, 1000}}
	for i := range benchmarkSizes {
		sizeA := benchmarkSizes[i][0]
		sizeB := benchmarkSizes[i][1]
		b.Run(fmt.Sprintf("sizes: A_%d, B_%d", sizeA, sizeB), benchmarkingFunc(sizeA, sizeB))
	}
}

func BenchmarkCoinsAdditionNoIntersect(b *testing.B) {
	benchmarkingFunc := func(numCoinsA int, numCoinsB int) func(b *testing.B) {
		return func(b *testing.B) {
			b.Helper()

			coinsA := Coins(make([]Coin, numCoinsA))
			coinsB := Coins(make([]Coin, numCoinsB))

			maxCoins := max(numCoinsA, numCoinsB)
			denomLength := len(fmt.Sprint(maxCoins))

			for i := range numCoinsA {
				denom := fmt.Sprintf("coinz_%0*d", denomLength, i)
				coinsA[i] = NewCoin(denom, int64(i+1))
			}
			for i := range numCoinsB {
				denom := fmt.Sprintf("coinz_%0*d", denomLength, i)
				coinsB[i] = NewCoin(denom, int64(i+1))
			}

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				coinsA.Add(coinsB)
			}
		}
	}

	benchmarkSizes := [][]int{{1, 1}, {5, 5}, {5, 20}, {1, 1000}, {2, 1000}, {1000, 2}}
	for i := range benchmarkSizes {
		sizeA := benchmarkSizes[i][0]
		sizeB := benchmarkSizes[i][1]
		b.Run(fmt.Sprintf("sizes: A_%d, B_%d", sizeA, sizeB), benchmarkingFunc(sizeA, sizeB))
	}
}
