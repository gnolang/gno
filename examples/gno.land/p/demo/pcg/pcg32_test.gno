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

func TestUint63PCG64(t *testing.T) {
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

	tests := []struct {
		state    uint64
		delta    uint64
		expected uint64
	}{
		{1, 1, 17764851043169991490},
		{1, 10, 13321747199226635079},
		{1, 100, 7812368804252231469},
		{1, 1000, 7287210203119977849},
		{1, 10000, 13461019320290411953},
	}

	for _, tt := range tests {
		result := pcg.advancedLCG64(tt.state, tt.delta, multiplier, incrementStep)
		if result != tt.expected {
			t.Errorf("lcg64(%d, %d) = %d; expected %d", tt.state, tt.delta, result, tt.expected)
		}
	}
}

func TestPCG32_Advance(t *testing.T) {
	pcg := NewPCG32()

	tests := []struct {
		state    uint64
		delta    uint64
		expected uint64
	}{
		{1, 1, 17764851043169991490},
		{1, 10, 13321747199226635079},
		{1, 100, 7812368804252231469},
		{1, 1000, 7287210203119977849},
		{1, 10000, 13461019320290411953},
	}

	for _, tt := range tests {
		pcg.state = tt.state
		pcg.Advance(tt.delta)
		if pcg.state != tt.expected {
			t.Errorf("Advance(%d) = %d; expected %d", tt.delta, pcg.state, tt.expected)
		}
	}
}

func TestPCG32_Retreat(t *testing.T) {
	pcg := NewPCG32()

	tests := []struct {
		initialState  uint64
		delta         uint64
		expectedState uint64
	}{
		{3643462645497912072, 1, 2297219049549015711},
		{15256603694110904427, 10, 9089490691196273925},
		{5234694153321213237, 100, 9614261562837832073},
		{2323235076269450313, 1000, 1018981208295873745},
		{6143568259046921169, 10000, 9666238883299984929},
	}

	for _, tt := range tests {
		pcg.state = tt.initialState
		pcg.Retreat(tt.delta)
		if pcg.state != tt.expectedState {
			t.Errorf("Retreat(%d) = %d; expected %d", tt.delta, pcg.state, tt.expectedState)
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

func TestPCG32_Read(t *testing.T) {
	tests := []struct {
		name     string
		seed     uint64
		bufSize  int
		expected []byte
	}{
		{
			name:     "Read 0 bytes",
			seed:     9876543210,
			bufSize:  0,
			expected: []byte{},
		},
		{
			name:     "Read 1 byte",
			seed:     42,
			bufSize:  1,
			expected: []byte{0xb0},
		},
		{
			name:     "Read 7 bytes",
			seed:     42,
			bufSize:  7,
			expected: []byte{0xb0, 0xfd, 0xe1, 0x6a, 0xe8, 0x9, 0xe4},
		},
		{
			name:    "Read 16 bytes",
			seed:    42,
			bufSize: 16,
			expected: []byte{
				0xb0, 0xfd, 0xe1, 0x6a,
				0xe8, 0x9, 0xe4, 0x3d,
				0x50, 0x24, 0x2c, 0x70,
				0xf0, 0xae, 0x88, 0x77,
			},
		},
		{
			name:    "Read 32 bytes",
			seed:    1234567890,
			bufSize: 32,
			expected: []byte{
				0x64, 0x44, 0xd9, 0x30,
				0x8, 0x8, 0x49, 0x6e,
				0x3f, 0xbd, 0xfe, 0xbf,
				0x30, 0x76, 0xe8, 0x56,
				0x45, 0xcd, 0xa1, 0xba,
				0xd4, 0xda, 0x3e, 0x62,
				0xa7, 0xa9, 0xf0, 0x9c,
				0x1b, 0x75, 0x2, 0xad,
			},
		},
		{
			name:    "Read 64 bytes",
			seed:    9876543210,
			bufSize: 64,
			expected: []byte{
				0xf5, 0xde, 0x18, 0xa3,
				0x19, 0x9b, 0xbb, 0xd6,
				0xc7, 0x7c, 0x69, 0xb3,
				0xaa, 0x21, 0x15, 0x69,
				0xbc, 0xea, 0x78, 0x1d,
				0xce, 0x78, 0xf8, 0x46,
				0xae, 0x98, 0xe4, 0x20,
				0xe2, 0x16, 0x6, 0x6d,
				0x3e, 0x99, 0x96, 0x66,
				0xd, 0xfc, 0x14, 0x79,
				0x4e, 0x59, 0x42, 0x25,
				0x9, 0x39, 0x41, 0x95,
				0x11, 0x3c, 0xa1, 0xa9,
				0xd2, 0x2, 0x10, 0x51,
				0x6e, 0xaa, 0xe4, 0x25,
				0xa8, 0x45, 0xe0, 0x76,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := NewPCG32().Seed(tc.seed, 0)

			buf := make([]byte, tc.bufSize)

			n, err := p.Read(buf)
			if err != nil {
				t.Fatalf("Read() error = %v; want nil", err)
			}

			// assert.Equal(t, tc.bufSize, n)
			// assert.Equal(t, tc.expected, buf)
			if tc.bufSize != n {
				t.Errorf("Read() = %d; expected %d", n, tc.bufSize)
			}

			for i := 0; i < tc.bufSize; i++ {
				if buf[i] != tc.expected[i] {
					t.Errorf("Read() = %v; expected %v", buf, tc.expected)
					break
				}
			}
		})
	}
}
