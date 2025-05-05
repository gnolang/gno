package xorshiftr128plus

import (
	"math/rand"
	"testing"
)

func TestXorshift64StarSeeding(t *testing.T) {
	rnd := New()
	value1 := rnd.Uint64()

	rnd = New(987654321)
	value2 := rnd.Uint64()

	rnd = New(987654321, 9876543210)
	value3 := rnd.Uint64()

	if value1 != 4368859828809982745 ||
		value2 != 6152356058823566752 ||
		value3 != 8285073084540510 ||
		value1 == value2 ||
		value2 == value3 ||
		value1 == value3 {
		t.Errorf("Expected three different values\n  got: %d, %d, %d", value1, value2, value3)
	}
}

func TestXorshiftr128PlusRand(t *testing.T) {
	rnd := New(987654321)
	rng := rand.New(rnd)

	// Expected outputs for the first 5 random floats with the given seed
	expected := []float64{
		0.048735219800779106,
		0.0372152171449619,
		0.667254760531175,
		0.16615979111253953,
		0.27578895545492665,
		0.48342823127830337,
		0.7825693830495895,
		0.14643955390763952,
		0.29003469381875835,
		0.726334398545258,
	}

	for i, exp := range expected {
		val := rng.Float64()
		if exp != val {
			t.Errorf("Rand.Float64() at iteration %d: got %g, expected %g", i, val, exp)
		}
	}
}

func TestXorshiftr128PlusUint64(t *testing.T) {
	rnd := New(987654321, 9876543210)

	expected := []uint64{
		8285073084540510,
		97010855169053386,
		11353359435625603792,
		10289232744262291728,
		14019961444418950453,
		15829492476941720545,
		2764732928842099222,
		6871047144273883379,
		16142204260470661970,
		11803223757041229095,
	}

	for i, exp := range expected {
		val := rnd.Uint64()
		if exp != val {
			t.Errorf("Xorshiftr128Plus.Uint64() at iteration %d: got %d, expected %d", i, val, exp)
		}
	}
}

func TestXorshiftr128PlusMarshalUnmarshal(t *testing.T) {
	rnd := New(987654321, 9876543210)

	expected1 := []uint64{
		8285073084540510,
		97010855169053386,
		11353359435625603792,
		10289232744262291728,
		14019961444418950453,
	}

	expected2 := []uint64{
		15829492476941720545,
		2764732928842099222,
		6871047144273883379,
		16142204260470661970,
		11803223757041229095,
	}

	for i, exp := range expected1 {
		val := rnd.Uint64()
		if exp != val {
			t.Errorf("Xorshiftr128Plus.Uint64() at iteration %d: got %d, expected %d", i, val, exp)
		}
	}

	marshalled, err := rnd.MarshalBinary()

	// t.Logf("Original State: [%x]\n", rnd.seed)
	// t.Logf("Marshalled State: [%x] -- %v\n", marshalled, err)
	state_before := rnd.seed

	if err != nil {
		t.Errorf("Xorshiftr128Plus.MarshalBinary() error: %v", err)
	}

	// Advance state by one number; then check the next 5. The expectation is that they _will_ fail.
	rnd.Uint64()

	for i, exp := range expected2 {
		val := rnd.Uint64()
		if exp == val {
			t.Errorf("  Iteration %d matched %d; which is from iteration %d; something strange is happening.", (i + 6), val, (i + 5))
		}
	}

	// t.Logf("State before unmarshall: [%x]\n", rnd.seed)

	// Now restore the state of the PRNG
	err = rnd.UnmarshalBinary(marshalled)

	// t.Logf("State after unmarshall: [%x]\n", rnd.seed)

	if state_before != rnd.seed {
		t.Errorf("States before and after marshal/unmarshal are not equal; go %x and %x", state_before, rnd.seed)
	}

	// Now we should be back on track for the last 5 numbers
	for i, exp := range expected2 {
		val := rnd.Uint64()
		if exp != val {
			t.Errorf("Xorshiftr128Plus.Uint64() at iteration %d: got %d, expected %d", (i + 5), val, exp)
		}
	}
}
