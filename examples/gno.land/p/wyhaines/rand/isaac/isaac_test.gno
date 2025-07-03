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
	_ = New()
}

func TestISAACRand(t *testing.T) {
	rnd := New(987654321)
	rng := rand.New(rnd)

	expected := []float64{
		0.3590173976876423,
		0.7045500585814575,
		0.3307624938209778,
		0.9174646414250772,
		0.11232269485263391,
		0.9276658847827113,
		0.9561549853128902,
		0.3921638978394879,
		0.9824881209760224,
		0.8213784955963486,
	}

	for i, exp := range expected {
		val := rng.Float64()
		if exp != val {
			t.Errorf("Rand.Float64() at iteration %d: got %g, expected %g", i, val, exp)
		}
	}
}

func TestISAACUint64(t *testing.T) {
	rnd := New(1000)

	expected := []uint64{
		13706738165129397958,
		11158865851759859683,
		7282911433372880595,
		10191257834247701829,
		6756510588635422211,
		9188469355127567259,
		3870407692778398450,
		7510499000403643056,
		11921506945015596058,
		5594436529078496461,
	}

	for i, exp := range expected {
		val := rnd.Uint64()
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
	rnd := New(1001)

	expected1 := []uint64{
		15520355889829550420,
		14048062122838424762,
		17386984608929258959,
		6631892771034928162,
		10939419587267807737,
	}

	expected2 := []uint64{
		3835981378089717525,
		13955658256492919532,
		10077846634918708781,
		5773204650675356768,
		4971527406542890875,
	}

	for i, exp := range expected1 {
		val := rnd.Uint64()
		if exp != val {
			t.Errorf("ISAAC.Uint64() at iteration %d: got %d, expected %d", i, val, exp)
		}
	}

	marshalled, err := rnd.MarshalBinary()

	// t.Logf("State: [%v]\n", dupState(rnd))
	// t.Logf("Marshalled State: [%x] -- %v\n", marshalled, err)
	state_before := dupState(rnd)

	if err != nil {
		t.Errorf("ISAAC.MarshalBinary() error: %v", err)
	}

	// Advance state by one number; then check the next 5. The expectation is that they _will_ fail.
	rnd.Uint64()

	for i, exp := range expected2 {
		val := rnd.Uint64()
		if exp == val {
			t.Errorf("  Iteration %d matched %d; which is from iteration %d; something strange is happening.", (i + 6), val, (i + 5))
		}
	}

	// t.Logf("State before unmarshall: [%v]\n", dupState(rnd))

	// Now restore the state of the PRNG
	err = rnd.UnmarshalBinary(marshalled)

	// t.Logf("State after unmarshall: [%v]\n", dupState(rnd))

	if state_before.Seed != dupState(rnd).Seed {
		t.Errorf("Seed mismatch")
	}
	if state_before.Randrsl != dupState(rnd).Randrsl {
		t.Errorf("Randrsl mismatch")
	}
	if state_before.Mm != dupState(rnd).Mm {
		t.Errorf("Mm mismatch")
	}
	if state_before.Aa != dupState(rnd).Aa {
		t.Errorf("Aa mismatch")
	}
	if state_before.Bb != dupState(rnd).Bb {
		t.Errorf("Bb mismatch")
	}
	if state_before.Cc != dupState(rnd).Cc {
		t.Errorf("Cc mismatch")
	}
	if state_before.Randcnt != dupState(rnd).Randcnt {
		t.Errorf("Randcnt mismatch")
	}

	// Now we should be back on track for the last 5 numbers
	for i, exp := range expected2 {
		val := rnd.Uint64()
		if exp != val {
			t.Errorf("ISAAC.Uint64() at iteration %d: got %d, expected %d", (i + 5), val, exp)
		}
	}
}
