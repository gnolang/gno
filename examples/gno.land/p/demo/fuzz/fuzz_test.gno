package fuzz

import (
	"crypto/chacha20/chacha"
	"crypto/chacha20/rand"
	"testing"
)

func TestMutate(t *testing.T) {
	seed := make([]byte, chacha.KeySize)
	rng := rand.NewFromSeed(seed)

	originalValue := "Hello"
	fuzzer := StringFuzzer{Value: originalValue, rng: rng}

	newFuzzer := fuzzer.Mutate().(*StringFuzzer)

	if newFuzzer.Value == originalValue {
		t.Errorf("Mutate did not change the string: got %v, want different from %v", newFuzzer.Value, originalValue)
	}

	if len(newFuzzer.Value) != len(originalValue) {
		t.Errorf("Mutated string has different length: got %s (len=%v), want %s (len=%v)", newFuzzer.Value, len(newFuzzer.Value), originalValue, len(originalValue))
	}

	if newFuzzer.rng != fuzzer.rng {
		t.Errorf("RNG instance in StringFuzzer has changed after mutation")
	}
}

func TestSelection(t *testing.T) {
	seed := make([]byte, 32)
	rng := rand.NewFromSeed(seed)

	tests := []struct {
		name       string
		population []*Individual
	}{
		{
			name:       "Empty population",
			population: []*Individual{},
		},
		{
			name: "Uniform fitness",
			population: []*Individual{
				{Fitness: 10},
				{Fitness: 10},
				{Fitness: 10},
			},
		},
		{
			name: "Different fitness",
			population: []*Individual{
				{Fitness: 5},
				{Fitness: 15},
				{Fitness: 10},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			selected := Selection(tc.population, rng)
			if len(selected) != len(tc.population) {
				t.Errorf("Expected selected length to be %d, got %d", len(tc.population), len(selected))
			}
		})
	}
}

func helperNewStringFuzzer(value string) *StringFuzzer {
	seed := make([]byte, 32)
	rng := rand.NewFromSeed(seed)
	return &StringFuzzer{Value: value, rng: rng}
}

func TestCrossover(t *testing.T) {
	parent1 := &Individual{Fuzzer: helperNewStringFuzzer("ABCDEFG"), Fitness: 0}
	parent2 := &Individual{Fuzzer: helperNewStringFuzzer("WXYZGSW"), Fitness: 0}

	seed := make([]byte, 32)
	rng := rand.NewFromSeed(seed)

	var child1, child2 *Individual
	for i := 0; i < 100; i++ {
		child1, child2 = Crossover(parent1, parent2, rng)
		// println("ch1: ", child1.Fuzzer.String())
		// println("ch2: ", child2.Fuzzer.String())
	}

	if child1.Fuzzer.String() == "ABCDEFG" || child2.Fuzzer.String() == "WXYZGSW" {
		t.Errorf("Crossover did not modify children correctly, got %s and %s", child1.Fuzzer.String(), child2.Fuzzer.String())
	}
}
