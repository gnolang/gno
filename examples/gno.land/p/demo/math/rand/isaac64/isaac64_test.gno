package isaac64

import (
	"math/rand"
	"testing"
)

type OpenISAAC struct {
	Randrsl    [256]uint64
	Randcnt    uint64
	Mm         [256]uint64
	Aa, Bb, Cc uint64
	Seed       [256]uint64
}

func TestISAACSeeding(t *testing.T) {
	isaac := New()
}

func TestISAACRand(t *testing.T) {
	source := New(987654321)
	rng := rand.New(source)

	// Expected outputs for the first 5 random floats with the given seed
	expected := []float64{
		0.9273376778618531,
		0.327620245173309,
		0.49315436150113456,
		0.9222536383598948,
		0.2999297342641162,
		0.4050531597269049,
		0.5321357451089953,
		0.19478000239059667,
		0.5156043950865713,
		0.9233494881511063,
	}

	for i, exp := range expected {
		val := rng.Float64()
		if exp != val {
			t.Errorf("Rand.Float64() at iteration %d: got %g, expected %g", i, val, exp)
		}
	}
}

func TestISAACUint64(t *testing.T) {
	isaac := New()

	expected := []uint64{
		6781932227698873623,
		14800945299485332986,
		4114322996297394168,
		5328012296808356526,
		12789214124608876433,
		17611101631239575547,
		6877490613942924608,
		15954522518901325556,
		14180160756719376887,
		4977949063252893357,
	}

	for i, exp := range expected {
		val := isaac.Uint64()
		if exp != val {
			t.Errorf("ISAAC.Uint64() at iteration %d: got %d, expected %d", i, val, exp)
		}
	}
}

func dupState(i *ISAAC) *OpenISAAC {
	state := &OpenISAAC{}
	state.Seed = i.seed
	state.Randrsl = i.randrsl
	state.Mm = i.mm
	state.Aa = i.aa
	state.Bb = i.bb
	state.Cc = i.cc
	state.Randcnt = i.randcnt

	return state
}

func TestISAACMarshalUnmarshal(t *testing.T) {
	isaac := New()

	expected1 := []uint64{
		6781932227698873623,
		14800945299485332986,
		4114322996297394168,
		5328012296808356526,
		12789214124608876433,
	}

	expected2 := []uint64{
		17611101631239575547,
		6877490613942924608,
		15954522518901325556,
		14180160756719376887,
		4977949063252893357,
	}

	for i, exp := range expected1 {
		val := isaac.Uint64()
		if exp != val {
			t.Errorf("ISAAC.Uint64() at iteration %d: got %d, expected %d", i, val, exp)
		}
	}

	marshalled, err := isaac.MarshalBinary()

	t.Logf("State: [%v]\n", dupState(isaac))
	t.Logf("Marshalled State: [%x] -- %v\n", marshalled, err)
	state_before := dupState(isaac)

	if err != nil {
		t.Errorf("ISAAC.MarshalBinary() error: %v", err)
	}

	// Advance state by one number; then check the next 5. The expectation is that they _will_ fail.
	isaac.Uint64()

	for i, exp := range expected2 {
		val := isaac.Uint64()
		if exp == val {
			t.Errorf("  Iteration %d matched %d; which is from iteration %d; something strange is happening.", (i + 6), val, (i + 5))
		}
	}

	t.Logf("State before unmarshall: [%v]\n", dupState(isaac))

	// Now restore the state of the PRNG
	err = isaac.UnmarshalBinary(marshalled)

	t.Logf("State after unmarshall: [%v]\n", dupState(isaac))

	if state_before.Seed != dupState(isaac).Seed {
		t.Errorf("Seed mismatch")
	}
	if state_before.Randrsl != dupState(isaac).Randrsl {
		t.Errorf("Randrsl mismatch")
	}
	if state_before.Mm != dupState(isaac).Mm {
		t.Errorf("Mm mismatch")
	}
	if state_before.Aa != dupState(isaac).Aa {
		t.Errorf("Aa mismatch")
	}
	if state_before.Bb != dupState(isaac).Bb {
		t.Errorf("Bb mismatch")
	}
	if state_before.Cc != dupState(isaac).Cc {
		t.Errorf("Cc mismatch")
	}
	if state_before.Randcnt != dupState(isaac).Randcnt {
		t.Errorf("Randcnt mismatch")
	}

	// Now we should be back on track for the last 5 numbers
	for i, exp := range expected2 {
		val := isaac.Uint64()
		if exp != val {
			t.Errorf("ISAAC.Uint64() at iteration %d: got %d, expected %d", (i + 5), val, exp)
		}
	}
}
