package fuzz

import (
	"crypto/chacha20/rand"
	"encoding/binary"
	"math"
	"time"
)

type Fuzzer interface {
	Mutate() Fuzzer
	String() string
}

type StringFuzzer struct {
	Value string
	rng   *rand.RNG
}

// NewStringFuzzer creates a new StringFuzzer with the given seed and number of rounds.
func NewStringFuzzer(seed []byte, rounds int) *StringFuzzer {
	rng := rand.NewFromSeed(seed)
	return &StringFuzzer{
		Value: "",
		rng:   rng,
	}
}

// Mutate changes a StringFuzzer's value by replacing a random character with a random ASCII character.
func (sf *StringFuzzer) Mutate() Fuzzer {
	runes := []rune(sf.Value)
	if len(runes) == 0 {
		return sf
	}

	index := sf.rng.Intn(len(runes))
	runes[index] = randomASCIIChar(sf.rng)
	return &StringFuzzer{
		Value: string(runes),
		rng:   sf.rng,
	}
}

func randomASCIIChar(r *rand.RNG) rune {
	return rune(r.Intn(127-32) + 32)
}

func (sf *StringFuzzer) String() string {
	return sf.Value
}

// Individual represents a single individual in the population.
type Individual struct {
	Fuzzer  Fuzzer
	Fitness int
}

func (idv *Individual) calculateFitness() {
	idv.Fitness = len(idv.Fuzzer.String())
}

// Selection selects individuals from the population based on their fitness.
//
// Use roulette wheel selection to select individuals from the population.
// ref: https://en.wikipedia.org/wiki/Fitness_proportionate_selection
// TODO: resolve scaling issues with large fitness values.
// TODO: what if fitness is negative?
func Selection(population []*Individual, rng *rand.RNG) []*Individual {
	fitness := 0
	for _, idv := range population {
		fitness += idv.Fitness
	}

	selected := make([]*Individual, 0, len(population))
	for i := 0; i < cap(selected); i++ {
		pick := rng.Intn(fitness)
		sum := 0
		for _, p := range population {
			sum += p.Fitness
			if sum >= pick {
				selected = append(selected, p)
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
func Crossover(parent1, parent2 *Individual, rng *rand.RNG) (*Individual, *Individual) {
	p1 := []rune(parent1.Fuzzer.String())
	p2 := []rune(parent2.Fuzzer.String())

	point := rng.Intn(len(p1))

	child1 := append(append([]rune{}, p1[:point]...), p2[point:]...)
	child2 := append(append([]rune{}, p2[:point]...), p1[point:]...)

	ch1 := &Individual{Fuzzer: &StringFuzzer{Value: string(child1)}}
	ch2 := &Individual{Fuzzer: &StringFuzzer{Value: string(child2)}}

	return ch1, ch2
}

func (idv *Individual) Mutate() {
	idv.Fuzzer = idv.Fuzzer.Mutate()
}
