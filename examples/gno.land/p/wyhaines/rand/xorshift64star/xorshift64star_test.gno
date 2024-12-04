package xorshift64star

import (
	"math/rand"
	"testing"
)

func TestXorshift64StarSeeding(t *testing.T) {
	xs64s := New()
	value1 := xs64s.Uint64()

	xs64s = New(987654321)
	value2 := xs64s.Uint64()

	if value1 != 5083824587905981259 || value2 != 18211065302896784785 || value1 == value2 {
		t.Errorf("Expected 5083824587905981259 to be != to 18211065302896784785; got: %d == %d", value1, value2)
	}
}

func TestXorshift64StarRand(t *testing.T) {
	source := New(987654321)
	rng := rand.New(source)

	// Expected outputs for the first 5 random floats with the given seed
	expected := []float64{
		.8344002228310946,
		0.01777174153236205,
		0.23521769507865276,
		0.5387610198576143,
		0.631539862225968,
		0.9369068148346704,
		0.6387002315083188,
		0.5047507613688854,
		0.5208486273732391,
		0.25023746271541747,
	}

	for i, exp := range expected {
		val := rng.Float64()
		if exp != val {
			t.Errorf("Rand.Float64() at iteration %d: got %g, expected %g", i, val, exp)
		}
	}
}

func TestXorshift64StarUint64(t *testing.T) {
	xs64s := New()

	expected := []uint64{
		5083824587905981259,
		4607286371009545754,
		2070557085263023674,
		14094662988579565368,
		2910745910478213381,
		18037409026311016155,
		17169624916429864153,
		10459214929523155306,
		11840179828060641081,
		1198750959721587199,
	}

	for i, exp := range expected {
		val := xs64s.Uint64()
		if exp != val {
			t.Errorf("Xorshift64Star.Uint64() at iteration %d: got %d, expected %d", i, val, exp)
		}
	}
}

func TestXorshift64StarMarshalUnmarshal(t *testing.T) {
	xs64s := New()

	expected1 := []uint64{
		5083824587905981259,
		4607286371009545754,
		2070557085263023674,
		14094662988579565368,
		2910745910478213381,
	}

	expected2 := []uint64{
		18037409026311016155,
		17169624916429864153,
		10459214929523155306,
		11840179828060641081,
		1198750959721587199,
	}

	for i, exp := range expected1 {
		val := xs64s.Uint64()
		if exp != val {
			t.Errorf("Xorshift64Star.Uint64() at iteration %d: got %d, expected %d", i, val, exp)
		}
	}

	marshalled, err := xs64s.MarshalBinary()

	t.Logf("Original State: [%x]\n", xs64s.seed)
	t.Logf("Marshalled State: [%x] -- %v\n", marshalled, err)
	state_before := xs64s.seed

	if err != nil {
		t.Errorf("Xorshift64Star.MarshalBinary() error: %v", err)
	}

	// Advance state by one number; then check the next 5. The expectation is that they _will_ fail.
	xs64s.Uint64()

	for i, exp := range expected2 {
		val := xs64s.Uint64()
		if exp == val {
			t.Errorf("  Iteration %d matched %d; which is from iteration %d; something strange is happening.", (i + 6), val, (i + 5))
		}
	}

	t.Logf("State before unmarshall: [%x]\n", xs64s.seed)

	// Now restore the state of the PRNG
	err = xs64s.UnmarshalBinary(marshalled)

	t.Logf("State after unmarshall: [%x]\n", xs64s.seed)

	if state_before != xs64s.seed {
		t.Errorf("States before and after marshal/unmarshal are not equal; go %x and %x", state_before, xs64s.seed)
	}

	// Now we should be back on track for the last 5 numbers
	for i, exp := range expected2 {
		val := xs64s.Uint64()
		if exp != val {
			t.Errorf("Xorshift64Star.Uint64() at iteration %d: got %d, expected %d", (i + 5), val, exp)
		}
	}
}
