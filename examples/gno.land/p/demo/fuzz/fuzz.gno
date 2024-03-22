package fuzz

import (
	"math"
)

type Fuzzer interface {
	Mutate() Fuzzer
	String() string
}

type StringFuzzer struct {
	Value string
}

// Mutate changes a StringFuzzer's value by replacing a random character
// with a random ASCII character.
func (sf *StringFuzzer) Mutate() Fuzzer {
	runes := []rune(sf.Value)
	if len(runes) == 0 {
		return sf
	}
	index := randRange(0, len(runes)-1)
	runes[index] = rune(randomASCIIChar())
	return &StringFuzzer{Value: string(runes)}
}

func (sf *StringFuzzer) String() string {
	return sf.Value
}

func randomASCIIChar() int {
	return int(randRange(32, 126))
}

// Individual represents a single individual in the population.
type Individual struct {
	Fuzzer  Fuzzer
	Fitness int
}

func (ind *Individual) calculateFitness() {
	ind.Fitness = len(ind.Fuzzer.String())
}

// Selection selects individuals from the population based on their fitness.
//
// Use roulette wheel selection to select individuals from the population.
// ref: https://en.wikipedia.org/wiki/Fitness_proportionate_selection
// TODO: resolve scaling issues with large fitness values.
// TODO: what if fitness is negative?
func Selection(population []*Individual) []*Individual {
	totalFitness := 0
	for _, ind := range population {
		totalFitness += ind.Fitness
	}

	selected := make([]*Individual, len(population))
	for i := range selected {
		pick := randRange(0, totalFitness-1)
		sum := 0
		for _, ind := range population {
			sum += ind.Fitness
			if uint64(sum) > uint64(pick) {
				selected[i] = ind
				break
			}
		}
	}

	return selected
}

// Crossover takes two parents and creates two children by combining their genetic material.
//
// The pivot point is chosen randomly from the length of the shortest parent. after the pivot point selected,
// the genetic material of the two parents is swapped to create the two children.
func Crossover(parent1, parent2 *Individual) (*Individual, *Individual) {
	p1Runes := []rune(parent1.Fuzzer.String())
	p2Runes := []rune(parent2.Fuzzer.String())

	crossoverPoint := int(randRange(0, len(p1Runes)-1))
	child1 := append(append([]rune{}, p1Runes[:crossoverPoint]...), p2Runes[crossoverPoint:]...)
	child2 := append(append([]rune{}, p2Runes[:crossoverPoint]...), p1Runes[crossoverPoint:]...)

	return &Individual{Fuzzer: &StringFuzzer{Value: string(child1)}}, &Individual{Fuzzer: &StringFuzzer{Value: string(child2)}}
}

func (ind *Individual) Mutate() {
	ind.Fuzzer = ind.Fuzzer.Mutate()
}
