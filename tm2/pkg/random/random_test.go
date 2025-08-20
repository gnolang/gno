// From AllInBits, Inc, TendermintClassic: github.com/tendermint/classic.
// License: Apache2.0.
package random

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRandStr(t *testing.T) {
	t.Parallel()

	l := 243
	s := RandStr(l)
	assert.Equal(t, l, len(s))
}

func TestRandBytes(t *testing.T) {
	t.Parallel()

	l := 243
	b := RandBytes(l)
	assert.Equal(t, l, len(b))
}

func TestRandIntn(t *testing.T) {
	t.Parallel()

	n := 243
	for range 100 {
		x := RandIntn(n)
		assert.True(t, x < n)
	}
}

// Test to make sure that we never call math.rand().
// We do this by ensuring that outputs are deterministic.
func TestDeterminism(t *testing.T) {
	var firstOutput string

	for i := range 100 {
		output := testThemAll()
		if i == 0 {
			firstOutput = output
		} else if firstOutput != output {
			t.Errorf("Run #%d's output was different from first run.\nfirst: %v\nlast: %v",
				i, firstOutput, output)
		}
	}
}

func testThemAll() string {
	// Such determinism.
	grand.reset(1)

	// Use it.
	out := new(bytes.Buffer)
	perm := RandPerm(10)
	blob, _ := json.Marshal(perm)
	fmt.Fprintf(out, "perm: %s\n", blob)
	fmt.Fprintf(out, "randInt: %d\n", RandInt())
	fmt.Fprintf(out, "randUint: %d\n", RandUint())
	fmt.Fprintf(out, "randIntn: %d\n", RandIntn(97))
	fmt.Fprintf(out, "randInt31: %d\n", RandInt31())
	fmt.Fprintf(out, "randInt32: %d\n", RandInt32())
	fmt.Fprintf(out, "randInt63: %d\n", RandInt63())
	fmt.Fprintf(out, "randInt64: %d\n", RandInt64())
	fmt.Fprintf(out, "randUint32: %d\n", RandUint32())
	fmt.Fprintf(out, "randUint64: %d\n", RandUint64())
	return out.String()
}

func TestRngConcurrencySafety(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()

			_ = RandUint64()
			<-time.After(time.Millisecond * time.Duration(RandIntn(100)))
			_ = RandPerm(3)
		}()
	}
	wg.Wait()
}

func BenchmarkRandBytes10B(b *testing.B) {
	benchmarkRandBytes(b, 10)
}

func BenchmarkRandBytes100B(b *testing.B) {
	benchmarkRandBytes(b, 100)
}

func BenchmarkRandBytes1KiB(b *testing.B) {
	benchmarkRandBytes(b, 1024)
}

func BenchmarkRandBytes10KiB(b *testing.B) {
	benchmarkRandBytes(b, 10*1024)
}

func BenchmarkRandBytes100KiB(b *testing.B) {
	benchmarkRandBytes(b, 100*1024)
}

func BenchmarkRandBytes1MiB(b *testing.B) {
	benchmarkRandBytes(b, 1024*1024)
}

func benchmarkRandBytes(b *testing.B, n int) {
	b.Helper()

	for i := 0; i < b.N; i++ {
		_ = RandBytes(n)
	}
	b.ReportAllocs()
}
