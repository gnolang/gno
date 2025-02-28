package fuzzing_test

import (
	"math"
	"testing"
	"testing/fuzzing"
)

func Test_nrand(t *testing.T) {
	sum := 0.0
	for i := 0; i < 1000; i++ {
		result := fuzzing.Nrand()
		sum += result
	}
	avg := sum / float64(1000)

	// The value returned by nrand() should be close to 0
	// on average for large sample sizes.
	// The expectation of the standard deviation should be
	// close to 1 for large sample sizes.
	if math.Abs(avg) > 0.1 { // can be adjusted based on sample size
		t.Errorf("nrand() = %v, want in range [-0.1, 0.1]", avg)
	}
}

func Test_GenerateRandomBool(t *testing.T) {
	for _, bias := range []float64{0, 0.5, 1} {
		trueCount, falseCount := 0, 0
		sampleSize := 1000

		for i := 0; i < sampleSize; i++ {
			result := fuzzing.GenerateRandomBool(bias)
			if result {
				trueCount++
			} else {
				falseCount++
			}
		}

		if bias < 0 || bias > 1 {
			t.Errorf("Bias = %v, want in range [0, 1]", bias)
		}
	}
}

func TestRandRange(t *testing.T) {
	nums := make(map[uint64]int)
	for i := 0; i < 1000; i++ {
		res := fuzzing.RandRange(0, 10)
		if res < 0 || res > 10 {
			t.Errorf("gerandRangenerateRange() = %v, want in range [0, 9]", res)
		}
		res2 := uint64(res)

		if _, ok := nums[res2]; ok {
			nums[res2]++
		} else {
			nums[res2] = 1
		}
	}

	if len(nums) != 10 {
		t.Errorf("len(nums) = %v, want in range [0, 10]", len(nums))
	}
}
