package fuzz

import (
	"math"
	"strconv"
	"testing"
	"time"
)

func updateSeed() {
	seed := time.Now().UnixNano()
	_srand(seed)
}

func Test_rand(t *testing.T) {
	updateSeed()

	for i := 0; i < 32768; i++ {	// 32768 is max range of rand()
		name := strconv.Itoa(i)
		t.Run(name, func(t *testing.T) {
			result := _rand()
			if result < 0 || result >= 32768 {
				t.Errorf("rand() = %v, want in range [0, 32767]", result)
			}
		})
	}
}

func Test_nrand(t *testing.T) {
	updateSeed()

	tests := []struct {
		name string
		maxRuns int
	}{
		{"1", 10000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sum := 0.0
			for i := 0; i < tt.maxRuns; i++ {
				result := nrand()
				sum += result
			}
			avg := sum / float64(tt.maxRuns)

			// The value returned by nrand() should be close to 0 
			// on average for large sample sizes.
			// The expectation of the standard deviation should be 
			// close to 1 for large sample sizes.
			if math.Abs(avg) > 0.1 { // can be adjusted based on sample size
				t.Errorf("nrand() = %v, want in range [-0.1, 0.1]", avg)
			}
		})
	}
}

func Test_GenerateRandomBool(t *testing.T) {
	updateSeed()

	for _, bias := range []float64{0, 0.5, 1} {
		trueCount, falseCount := 0, 0
		sampleSize := 1000

		for i := 0; i < sampleSize; i++ {
			result := GenerateRandomBool(bias)
			if result {
				trueCount++
			} else {
				falseCount++
			}
		}

		if trueCount == 0 || falseCount == 0 {
			t.Errorf("Bias = %v, trueCount = %v, falseCount = %v, want both > 0", bias, trueCount, falseCount)
		}

		if bias < 0 || bias > 1 {
			t.Errorf("Bias = %v, want in range [0, 1]", bias)
		}
	}
}