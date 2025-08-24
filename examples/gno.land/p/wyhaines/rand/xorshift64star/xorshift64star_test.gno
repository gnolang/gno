package xorshift64star

import (
	"math/rand"
	"testing"
)

func TestXorshift64StarSeeding(t *testing.T) {
	rnd := New()
	value1 := rnd.Uint64()

	rnd = New(987654321)
	value2 := rnd.Uint64()

	if value1 != 3506338452768534464 || value2 != 18211065302896784785 || value1 == value2 {
		t.Errorf("Expected 2 different values; got: %d and %d", value1, value2)
	}
}

func TestXorshift64StarRand(t *testing.T) {
	rnd := New(987654321)
	rng := rand.New(rnd)

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
	rnd := New(1000)

	expected := []uint64{
		12487186097140327178,
		14661465266242046735,
		14694269887751719025,
		15763651124252725051,
		6358063137690011177,
		16123467710013993794,
		8086526499083208127,
		5907916440057635441,
		7074828965897564835,
		219959441764368518,
	}

	for i, exp := range expected {
		val := rnd.Uint64()
		if exp != val {
			t.Errorf("Xorshift64Star.Uint64() at iteration %d: got %d, expected %d", i, val, exp)
		}
	}
}

func TestXorshift64StarMarshalUnmarshal(t *testing.T) {
	rnd := New(1001)

	expected1 := []uint64{
		17667678392346722343,
		8998941448230025236,
		6038719778092581458,
		9916057400810746083,
		3240504040424884895,
	}

	expected2 := []uint64{
		3977048231667561376,
		18438555156602529247,
		2172795924893074637,
		12043507788481457357,
		8032279100325099159,
	}

	for i, exp := range expected1 {
		val := rnd.Uint64()
		if exp != val {
			t.Errorf("Xorshift64Star.Uint64() at iteration %d: got %d, expected %d", i, val, exp)
		}
	}

	marshalled, err := rnd.MarshalBinary()

	// t.Logf("Original State: [%x]\n", rnd.seed)
	// t.Logf("Marshalled State: [%x] -- %v\n", marshalled, err)
	state_before := rnd.seed

	if err != nil {
		t.Errorf("Xorshift64Star.MarshalBinary() error: %v", err)
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
			t.Errorf("Xorshift64Star.Uint64() at iteration %d: got %d, expected %d", (i + 5), val, exp)
		}
	}
}
