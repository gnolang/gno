package xorshiftr128plus

import (
	"math/rand"
	"testing"
)

func TestXorshift64StarSeeding(t *testing.T) {
	xs128p := New()
	value1 := xs128p.Uint64()

	xs128p = New(987654321)
	value2 := xs128p.Uint64()

	xs128p = New(987654321, 9876543210)
	value3 := xs128p.Uint64()

	if value1 != 13970141264473760763 ||
		value2 != 17031892808144362974 ||
		value3 != 8285073084540510 ||
		value1 == value2 ||
		value2 == value3 ||
		value1 == value3 {
		t.Errorf("Expected three different values: 13970141264473760763, 17031892808144362974, and 8285073084540510\n  got: %d, %d, %d", value1, value2, value3)
	}
}

func TestXorshiftr128PlusRand(t *testing.T) {
	source := New(987654321)
	rng := rand.New(source)

	// Expected outputs for the first 5 random floats with the given seed
	expected := []float64{
		0.9199548549485674,
		0.0027491282372705816,
		0.31493362274701164,
		0.3531250819119609,
		0.09957852858060356,
		0.731941362705936,
		0.3476937688876708,
		0.1444018086140385,
		0.9106467321832331,
		0.8024870151488901,
	}

	for i, exp := range expected {
		val := rng.Float64()
		if exp != val {
			t.Errorf("Rand.Float64() at iteration %d: got %g, expected %g", i, val, exp)
		}
	}
}

func TestXorshiftr128PlusUint64(t *testing.T) {
	xs128p := New(987654321, 9876543210)

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
		val := xs128p.Uint64()
		if exp != val {
			t.Errorf("Xorshiftr128Plus.Uint64() at iteration %d: got %d, expected %d", i, val, exp)
		}
	}
}

func TestXorshiftr128PlusMarshalUnmarshal(t *testing.T) {
	xs128p := New(987654321, 9876543210)

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
		val := xs128p.Uint64()
		if exp != val {
			t.Errorf("Xorshiftr128Plus.Uint64() at iteration %d: got %d, expected %d", i, val, exp)
		}
	}

	marshalled, err := xs128p.MarshalBinary()

	t.Logf("Original State: [%x]\n", xs128p.seed)
	t.Logf("Marshalled State: [%x] -- %v\n", marshalled, err)
	state_before := xs128p.seed

	if err != nil {
		t.Errorf("Xorshiftr128Plus.MarshalBinary() error: %v", err)
	}

	// Advance state by one number; then check the next 5. The expectation is that they _will_ fail.
	xs128p.Uint64()

	for i, exp := range expected2 {
		val := xs128p.Uint64()
		if exp == val {
			t.Errorf("  Iteration %d matched %d; which is from iteration %d; something strange is happening.", (i + 6), val, (i + 5))
		}
	}

	t.Logf("State before unmarshall: [%x]\n", xs128p.seed)

	// Now restore the state of the PRNG
	err = xs128p.UnmarshalBinary(marshalled)

	t.Logf("State after unmarshall: [%x]\n", xs128p.seed)

	if state_before != xs128p.seed {
		t.Errorf("States before and after marshal/unmarshal are not equal; go %x and %x", state_before, xs128p.seed)
	}

	// Now we should be back on track for the last 5 numbers
	for i, exp := range expected2 {
		val := xs128p.Uint64()
		if exp != val {
			t.Errorf("Xorshiftr128Plus.Uint64() at iteration %d: got %d, expected %d", (i + 5), val, exp)
		}
	}
}
