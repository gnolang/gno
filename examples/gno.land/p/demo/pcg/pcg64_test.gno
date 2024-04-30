package pcg

import (
	"math"
	"testing"
	"time"
)

func TestPCG_Uint63(t *testing.T) {
	pcg := NewPCG64(12345, 67890)

	for i := 0; i < 100000; i++ {
		val := pcg.Uint63()
		if val < 0 {
			t.Errorf("Uint63() = %d; want a non-negative number", val)
		}
		if val > 0x7FFFFFFFFFFFFFFF {
			t.Errorf("Uint63() = %d; want a 63-bit integer", val)
		}
	}
}

func TestPCG_Advance(t *testing.T) {
	pcg := NewPCG64(12345, 67890)

	tests := []struct {
		delta           uint64
		expectedStateHi uint64
		expectedStateLo uint64
	}{
		{1, 5288318170876267920, 8586121761925336369},
		{10, 658778831176772942, 680070833917286903},
		{100, 11328633206433140426, 2423557365805501827},
		{1000, 14055077221971198434, 7063816904067589179},
		{10000, 3283529072340023762, 12459493142276168811},
	}

	for _, tc := range tests {
		pcg.Advance(tc.delta)
		if pcg.hi.state != tc.expectedStateHi {
			t.Errorf("Advance(%d) hi state = %d; expected %d", tc.delta, pcg.hi.state, tc.expectedStateHi)
		}
		if pcg.lo.state != tc.expectedStateLo {
			t.Errorf("Advance(%d) lo state = %d; expected %d", tc.delta, pcg.lo.state, tc.expectedStateLo)
		}
	}
}

func TestPCG(t *testing.T) {
	p := NewPCG64(1, 2)
	want := []uint64{
		0x6a3b45f887fad67c,
		0x1b0a91e2d7d75723,
		0x905a1518e26e8445,
		0x8cb6b7c0ea9f200c,
		0x59afa674b44b2509,
		0x86ab5e04d104bd4c,
		0x13e180669e2d07ea,
		0x5a6a0bb349dd26ae,
		0x1cd5f134a8581b57,
		0xc807c2686fe5baff,
		0x846b5fc66eb343cb,
		0x169adcefdd042f97,
		0x9b58ba4c9ef301fa,
		0xd54fc77c467d80bf,
		0xb776b6e2508ebab3,
		0xfd002241e862137b,
		0xc6a0fa2cfa1f95e9,
		0x67469082d8377e0a,
		0x7f30faccba8029d4,
		0xf200f5696d060925,
	}

	for i, x := range want {
		u := p.Uint64nWithMCG()
		if u != x {
			t.Errorf("PCG #%d = %#x, want %#x", i, u, x)
		}
	}
}

func TestPCG_Retreat(t *testing.T) {
	pcg := NewPCG64(12345, 67890)

	tests := []struct {
		delta           uint64
		expectedStateHi uint64
		expectedStateLo uint64
	}{
		{1, 9562276038148940761, 762819840426222678},
		{10, 6073181909808593307, 4206663794773473040},
		{100, 13320423970107457037, 15716292132216561082},
		{1000, 2035794469121007023, 329538305516032724},
		{10000, 8361027235883110657, 2874612980081388766},
	}

	for _, tt := range tests {
		pcg.Retreat(tt.delta)
		if pcg.hi.state != tt.expectedStateHi {
			t.Errorf("Retreat(%d) hi state = %d; expected %d", tt.delta, pcg.hi.state, tt.expectedStateHi)
		}
		if pcg.lo.state != tt.expectedStateLo {
			t.Errorf("Retreat(%d) lo state = %d; expected %d", tt.delta, pcg.lo.state, tt.expectedStateLo)
		}
	}
}

func Test_ExamplePCG64_Shuffle(t *testing.T) {
	pcg := NewPCG64(42, 54)
	pcg.Seed(42, 54, 18, 27)

	array := []int{1, 2, 3, 4, 5}
	pcg.Shuffle(len(array), func(i, j int) {
		array[i], array[j] = array[j], array[i]
	})

	if len(array) != 5 {
		t.Errorf("Shuffle() len(array) = %d; want 5", len(array))
	}

	if isArrayEqual(array, []int{1, 2, 3, 4, 5}) {
		t.Errorf("Shuffle() array = %v; want shuffled", array)
	}
}

func TestFloat64(t *testing.T) {
	pcg := NewPCG64(42, 54)
	for i := 0; i < 1000; i++ {
		val := pcg.Float64()
		if val < 0.0 || val > 1.0 {
			t.Errorf("Float64() returned a value out of bounds: %f", val)
		}
	}
}

func TestFloat64Full(t *testing.T) {
	pcg := NewPCG64(42, 54)
	for i := 0; i < 1000; i++ {
		val := pcg.Float64Full()
		if val < 0.0 || val >= 1.0 {
			t.Errorf("Float64Full() returned a value out of bounds: %f", val)
		}
	}
}

func TestPCG64Read(t *testing.T) {
	now := uint64(time.Now().UnixNano())
	testSizes := []int{16, 32, 48, 64, 100, 1023, 2048}
	for _, size := range testSizes {
		buf := make([]byte, size)
		pcg := NewPCG64(12345, now)
		n, err := pcg.Read(buf)
		if err != nil {
			t.Errorf("Read returned an error: %v", err)
		}
		if n != size {
			t.Errorf("Read returned wrong number of bytes: got %v, want %v", n, size)
		}
		// Check if bytes are not all zero; this is a simplistic randomness check
		allZero := true
		for _, b := range buf {
			if b != 0 {
				allZero = false
				break
			}
		}
		if allZero {
			t.Errorf("Buffer of size %d is filled with zeros, which is highly improbable", size)
		}
	}
}

func TestPCG64ReadEdgeCases(t *testing.T) {
	now := uint64(time.Now().UnixNano())
	edgeSizes := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 15, 17}
	for _, size := range edgeSizes {
		buf := make([]byte, size)
		pcg := NewPCG64(12345, now)
		n, err := pcg.Read(buf)
		if err != nil {
			t.Errorf("Read returned an error: %v", err)
		}
		if n != size {
			t.Errorf("Read returned wrong number of bytes: got %v, want %v", n, size)
		}

		nonZeroFound := false
		for _, b := range buf {
			if b != 0 {
				nonZeroFound = true
				break
			}
		}
		if !nonZeroFound && size != 0 {
			t.Errorf("Buffer of size %d has no non-zero bytes, which is highly improbable", size)
		}
	}
}
