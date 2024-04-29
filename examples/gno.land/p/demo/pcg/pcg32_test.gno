package pcg

import (
	"math"
	"testing"
)

func TestPCG32_Uintn32(t *testing.T) {
	pcg := NewPCG32().Seed(12345, 67890)

	testCases := []struct {
		bound uint32
	}{
		{0},
		{1},
		{10},
		{100},
		{1000},
		{10000},
	}

	for _, tc := range testCases {
		result := pcg.Uintn32(tc.bound)
		if tc.bound != 0 && result >= tc.bound {
			t.Errorf("Bounded(%d) = %d; expected a value between 0 and %d", tc.bound, result, tc.bound)
		}
		if tc.bound == 0 && result != 0 {
			t.Errorf("Bounded(%d) = %d; expected 0", tc.bound, result)
		}
	}
}

func TestUint63(t *testing.T) {
	pcg := NewPCG64(42, 54)
	pcg.Seed(42, 54, 18, 27)
	for i := 0; i < 100; i++ {
		val := pcg.Uint63()
		if val < 0 || val > math.MaxInt64 {
			t.Errorf("Value out of bounds: %d", val)
		}
	}
}

func TestPCG32_UniformDistribution(t *testing.T) {
	pcg := NewPCG32().Seed(12345, 67890)
	numBins := 10
	numSamples := 1000000
	toleranceRatio := 10 // 10% tolerance
	bins := make([]int, numBins)

	for i := 0; i < numSamples; i++ {
		r := pcg.Uint32()
		binIndex := int(uint64(r) * uint64(numBins) >> 32)
		bins[binIndex]++
	}

	expected := numSamples / numBins
	tolerance := expected / toleranceRatio

	for _, count := range bins {
		if abs(count-expected) > tolerance {
			t.Errorf("bin count %d is outside the expected range [%d, %d]", count, expected-tolerance, expected+tolerance)
		}
	}
}

func TestPCG32_lcg64(t *testing.T) {
	pcg := NewPCG32()

	testCases := []struct {
		state    uint64
		delta    uint64
		expected uint64
	}{
		{1, 1, 3643462645497912072},
		{1, 10, 15256603694110904427},
		{1, 100, 5234694153321213237},
		{1, 1000, 2323235076269450313},
		{1, 10000, 6143568259046921169},
	}

	for _, tc := range testCases {
		result := pcg.advancedLCG64(tc.state, tc.delta, multiplier, incrementStep)
		if result != tc.expected {
			t.Errorf("lcg64(%d, %d) = %d; expected %d", tc.state, tc.delta, result, tc.expected)
		}
	}
}

func TestPCG32_Advance(t *testing.T) {
	pcg := NewPCG32()

	testCases := []struct {
		initialState  uint64
		delta         uint64
		expectedState uint64
	}{
		{1, 1, 3643462645497912072},
		{1, 10, 15256603694110904427},
		{1, 100, 5234694153321213237},
		{1, 1000, 2323235076269450313},
		{1, 10000, 6143568259046921169},
	}

	for _, tc := range testCases {
		pcg.state = tc.initialState
		pcg.Advance(tc.delta)
		if pcg.state != tc.expectedState {
			t.Errorf("Advance(%d) = %d; expected %d", tc.delta, pcg.state, tc.expectedState)
		}
	}
}

func TestPCG32_Retreat(t *testing.T) {
	pcg := NewPCG32()

	testCases := []struct {
		initialState  uint64
		delta         uint64
		expectedState uint64
	}{
		{3643462645497912072, 1, 1},
		{15256603694110904427, 10, 1},
		{5234694153321213237, 100, 1},
		{2323235076269450313, 1000, 1},
		{6143568259046921169, 10000, 1},
	}

	for _, tc := range testCases {
		pcg.state = tc.initialState
		pcg.Retreat(tc.delta)
		if pcg.state != tc.expectedState {
			t.Errorf("Retreat(%d) = %d; expected %d", tc.delta, pcg.state, tc.expectedState)
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func TestPCG32_Shuffle(t *testing.T) {
	pcg := NewPCG32().Seed(12345, 67890)

	testCases := []struct {
		n        int
		expected []int
	}{
		{0, []int{}},
		{1, []int{0}},
		{2, []int{0, 1}},
		{3, []int{0, 1, 2}},
		{4, []int{0, 1, 2, 3}},
		{5, []int{0, 1, 2, 3, 4}},
	}

	for _, tc := range testCases {
		arr := make([]int, tc.n)
		for i := range arr {
			arr[i] = i
		}

		pcg.Shuffle(tc.n, func(i, j int) {
			arr[i], arr[j] = arr[j], arr[i]
		})

		if tc.n > 1 && isArrayEqual(arr, tc.expected) {
			t.Errorf("Shuffle(%d) = %v; expected a shuffled version of %v", tc.n, arr, tc.expected)
		}
	}
}

func isArrayEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestPCG32_Perm(t *testing.T) {
	pcg := NewPCG32().Seed(12345, 67890)

	tests := []struct {
		n        int
		expected []int
	}{
		{0, []int{}},
		{1, []int{0}},
		{2, []int{0, 1}},
		{3, []int{0, 1, 2}},
		{4, []int{0, 1, 2, 3}},
		{5, []int{0, 1, 2, 3, 4}},
	}

	for _, tt := range tests {
		result := pcg.Perm(tt.n)
		if !isPermutation(result, tt.expected) {
			t.Errorf("Perm(%d) = %v; expected a permutation of %v", tt.n, result, tt.expected)
		}
	}
}

func isPermutation(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	counts := make(map[int]int)
	for _, num := range a {
		counts[num]++
	}
	for _, num := range b {
		if counts[num] == 0 {
			return false
		}
		counts[num]--
	}
	return true
}
