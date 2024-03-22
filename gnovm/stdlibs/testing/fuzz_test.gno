package testing

import (
	"encoding/binary"
	"testing"
	"time"

	"gno.land/p/demo/ufmt"
)

func TestMutate(t *testing.T) {
	originalValue := "Hello"
	fuzzer := StringFuzzer{Value: originalValue}

	newFuzzer := fuzzer.Mutate().(*StringFuzzer)

	if newFuzzer.Value == originalValue {
		t.Errorf("Mutate did not change the string: got %v, want different from %v", newFuzzer.Value, originalValue)
	}

	if len(newFuzzer.Value) != len(originalValue) {
		t.Errorf("Mutated string has different length: got %s (len=%v), want %s (len=%v)", newFuzzer.Value, len(newFuzzer.Value), originalValue, len(originalValue))
	}
}

func TestSelection(t *testing.T) {
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
			selected := Selection(tc.population)
			if len(selected) != len(tc.population) {
				t.Errorf("Expected selected length to be %d, got %d", len(tc.population), len(selected))
			}
		})
	}
}

func helperNewStringFuzzer(value string) *StringFuzzer {
	return &StringFuzzer{Value: value}
}

func TestCrossover(t *testing.T) {
	parent1 := &Individual{Fuzzer: helperNewStringFuzzer("foobar"), Fitness: 0}
	parent2 := &Individual{Fuzzer: helperNewStringFuzzer("bazbiz"), Fitness: 0}

	var child1, child2 *Individual
	for i := 0; i < 100; i++ {
		child1, child2 = Crossover(parent1, parent2)
	}

	if child1.Fuzzer.String() == "foobar" || child2.Fuzzer.String() == "bazbiz" {
		t.Errorf("Crossover did not modify children correctly, got %s and %s", child1.Fuzzer.String(), child2.Fuzzer.String())
	}
}

// NOT TESTING PURPOSE (just main function)
func Test_main(t *testing.T) {
	population := []*Individual{
		{Fuzzer: &StringFuzzer{Value: "hello"}},
		{Fuzzer: &StringFuzzer{Value: "fuzz"}},
		{Fuzzer: &StringFuzzer{Value: "worldwide"}},
		{Fuzzer: &StringFuzzer{Value: "barfoo"}},
	}

	for _, ind := range population {
		ind.calculateFitness()
	}

	for generation := 0; generation < 250; generation++ {
		population = Selection(population)

		newPopulation := make([]*Individual, 0, len(population))
		for i := 0; i < len(population); i += 2 {
			if i+1 < len(population) {
				child1, child2 := Crossover(population[i], population[i+1])
				newPopulation = append(newPopulation, child1, child2)
			} else {
				newPopulation = append(newPopulation, population[i])
			}
		}
		// Apply mutation operator
		for _, ind := range newPopulation {
			if GenerateRandomBool(0.1) { // 10% mutation rate
				ind.Mutate()
			}
		}

		for _, ind := range newPopulation {
            if GenerateRandomBool(0.1) {
                ind.Fuzzer = ind.Fuzzer.InsertDeleteMutate(0.3)
            }
        }

		// Calculate fitness for the new population
		for _, ind := range newPopulation {
			ind.calculateFitness()
		}

		// Print the best individual from the current generation
		var (
			bestFitness    int
			bestIndividual string
		)

		for _, ind := range newPopulation {
			if ind.Fitness > bestFitness {
				bestFitness = ind.Fitness
				bestIndividual = ind.Fuzzer.String()
			}
		}

		// Update the population for the next generation
		population = newPopulation
		println("Generation", generation, ":", bestIndividual, "Fitness:", bestFitness)
	}
}
