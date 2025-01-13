package isaac

import (
	"math/rand"
	"testing"
)

type OpenISAAC struct {
	Randrsl    [256]uint32
	Randcnt    uint32
	Mm         [256]uint32
	Aa, Bb, Cc uint32
	Seed       [256]uint32
}

func TestISAACSeeding(t *testing.T) {
	isaac := New()
}

func TestISAACRand(t *testing.T) {
	source := New(987654321)
	rng := rand.New(source)

	// Expected outputs for the first 5 random floats with the given seed
	expected := []float64{
		0.17828173023837635,
		0.7327795780287832,
		0.4850369074875177,
		0.9474842397428482,
		0.6747135561813891,
		0.7522507082868403,
		0.041115261836534356,
		0.7405243709084567,
		0.672863376128768,
		0.11866211399980553,
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
		5986068031949215749,
		10437354066128700566,
		13478007513323023970,
		8969511410255984224,
		3869229557962857982,
		1762449743873204415,
		5292356290662282456,
		7893982194485405616,
		4296136494566588699,
		12414349056998262772,
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
		5986068031949215749,
		10437354066128700566,
		13478007513323023970,
		8969511410255984224,
		3869229557962857982,
	}

	expected2 := []uint64{
		1762449743873204415,
		5292356290662282456,
		7893982194485405616,
		4296136494566588699,
		12414349056998262772,
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
