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
	_ = New()
}

func TestISAACRand(t *testing.T) {
	rnd := New(987654321)
	rng := rand.New(rnd)

	// Expected outputs for the first 5 random floats with the given seed
	expected := []float64{
		0.2818878834295122,
		0.8575461830821571,
		0.9878021063787968,
		0.6503544780116336,
		0.5158329690433359,
		0.7959152461588924,
		0.5432366486934906,
		0.824665978209607,
		0.8615372170680458,
		0.22954589404739578,
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
		10083220283665581455,
		10039389761195725041,
		6820016387036140989,
		6784213597523088182,
		13120722600477653778,
		3491117614651563646,
		1297676147275528930,
		15006384980354042338,
		3104467119059991036,
		4914319123654344819,
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
		4398183556077595549,
		14479654616302101831,
		15852653767232940552,
		2801765968457115882,
		8875575139772470433,
	}

	expected2 := []uint64{
		17583056722733587141,
		16906215529544723388,
		7599862885469865851,
		9623269843822592805,
		4311429062865512072,
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
