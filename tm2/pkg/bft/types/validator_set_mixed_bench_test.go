package types

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
)

func BenchmarkMixedScheme_VerifyCommit(b *testing.B) {
	const (
		chainID = "mixed-scheme-benchmark"
		height  = int64(42)
		round   = 0
	)
	blockID := BlockID{Hash: []byte("benchmark-block-hash")}
	rowByName := make(map[string]verifyCommitBenchRow)
	var rowOrder []string

	sizes := []int{10, 50, 100, 180}
	for _, size := range sizes {
		mixes := []struct {
			name      string
			secpCount int
		}{
			{name: "all_ed25519", secpCount: 0},
			{name: "one_secp256k1", secpCount: 1},
			{name: "ten_percent_secp256k1", secpCount: max(1, size/10)},
			{name: "one_third_secp256k1", secpCount: size / 3},
			{name: "all_secp256k1", secpCount: size},
		}

		for _, mix := range mixes {
			name := fmt.Sprintf("validators_%03d/%s", size, mix.name)
			b.Run(name, func(b *testing.B) {
				valSet, commit := benchmarkCommit(b, size, mix.secpCount, chainID, blockID, height, round)

				b.ReportAllocs()

				if err := valSet.VerifyCommit(chainID, blockID, height, commit); err != nil {
					b.Fatalf("warmup VerifyCommit failed: %v", err)
				}

				b.ResetTimer()
				b.ReportMetric(float64(size), "validators")
				b.ReportMetric(float64(mix.secpCount), "secp_validators")
				start := time.Now()
				for i := 0; i < b.N; i++ {
					if err := valSet.VerifyCommit(chainID, blockID, height, commit); err != nil {
						b.Fatal(err)
					}
				}
				elapsed := time.Since(start)
				b.StopTimer()

				if _, ok := rowByName[name]; !ok {
					rowOrder = append(rowOrder, name)
				}
				rowByName[name] = verifyCommitBenchRow{
					valsetSize: size,
					mix:        benchmarkMixLabel(size, mix.secpCount),
					avg:        elapsed / time.Duration(b.N),
				}
			})
		}
	}

	printVerifyCommitBenchSummary(rowOrder, rowByName)
}

type verifyCommitBenchRow struct {
	valsetSize int
	mix        string
	avg        time.Duration
}

func benchmarkMixLabel(size int, secpCount int) string {
	switch secpCount {
	case 0:
		return "all ed25519"
	case 1:
		return "1 secp"
	case size:
		return "all secp"
	default:
		return fmt.Sprintf("%d secp", secpCount)
	}
}

func printVerifyCommitBenchSummary(rowOrder []string, rowByName map[string]verifyCommitBenchRow) {
	if len(rowOrder) == 0 {
		return
	}

	fmt.Println()
	fmt.Println("| Valset | Mix | VerifyCommit |")
	fmt.Println("|---:|---|---:|")
	for _, name := range rowOrder {
		row := rowByName[name]
		fmt.Printf("| %d | %s | ~%.2f ms |\n",
			row.valsetSize,
			row.mix,
			float64(row.avg.Nanoseconds())/float64(time.Millisecond),
		)
	}
	fmt.Println()
}

func benchmarkCommit(
	b *testing.B,
	size int,
	secpCount int,
	chainID string,
	blockID BlockID,
	height int64,
	round int,
) (*ValidatorSet, *Commit) {
	b.Helper()

	valSet, privVals := benchmarkValSet(b, size, secpCount)
	voteSet := NewVoteSet(chainID, height, round, PrecommitType, valSet)
	commit, err := MakeCommit(blockID, height, round, voteSet, privVals)
	if err != nil {
		b.Fatalf("make commit: %v", err)
	}
	return valSet, commit
}

func benchmarkValSet(b *testing.B, size int, secpCount int) (*ValidatorSet, []PrivValidator) {
	b.Helper()
	if size <= 0 {
		b.Fatalf("size must be positive")
	}
	if secpCount < 0 || secpCount > size {
		b.Fatalf("invalid secpCount %d for size %d", secpCount, size)
	}

	valz := make([]*Validator, size)
	privVals := make([]PrivValidator, size)
	for i := 0; i < size; i++ {
		privKey := benchmarkPrivKey(i < secpCount)
		valz[i] = NewValidator(privKey.PubKey(), 1)
		privVals[i] = NewMockPVWithPrivKey(privKey)
	}

	sort.Sort(PrivValidatorsByAddress(privVals))
	return NewValidatorSet(valz), privVals
}

func benchmarkPrivKey(secp bool) crypto.PrivKey {
	if secp {
		return secp256k1.GenPrivKey()
	}
	return ed25519.GenPrivKey()
}
