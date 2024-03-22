package testing

import (
	"math"
	"strings"
	"time"
)

type Fuzzer interface {
	InsertDeleteMutate(p float64) Fuzzer
	Mutate() Fuzzer
	String() string
}

type StringFuzzer struct {
	Value string
	f     *F
}

// Mutate changes a StringFuzzer's value by replacing a random character
// with a random ASCII character.
func (sf *StringFuzzer) Mutate() Fuzzer {
	runes := []rune(sf.Value)
	if len(runes) == 0 {
		return sf
	}
	index := randRange(0, len(runes)-1)
	runes[index] = randomASCIIChar()
	return &StringFuzzer{Value: string(runes)}
}

func (sf *StringFuzzer) InsertDeleteMutate(p float64) Fuzzer {
	return &StringFuzzer{Value: InsertDelete(sf.Value, p)}
}

func (sf *StringFuzzer) Fuzz() string {
	if GenerateRandomBool(0.2) {
		return InsertDelete(sf.Value, 0.1)
	} else {
		rs := []rune(sf.Value)
		if len(rs) > 0 {
			index := randRange(0, len(rs)-1)
			rs[index] = randomASCIIChar()
		}

		return string(rs)
	}
}

func (sf *StringFuzzer) String() string {
	return sf.Value
}

func randomASCIIChar() rune {
	r := int(randRange(32, 126))
	return rune(r)
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

	p1Len := len(p1Runes)
	p2Len := len(p2Runes)

	point := 0
	if p1Len >= p2Len {
		point = int(randRange(0, p2Len-1))
	} else {
		point = int(randRange(0, p1Len-1))
	}
	child1 := append(append([]rune{}, p1Runes[:point]...), p2Runes[point:]...)
	child2 := append(append([]rune{}, p2Runes[:point]...), p1Runes[point:]...)

	return &Individual{Fuzzer: &StringFuzzer{Value: string(child1)}}, &Individual{Fuzzer: &StringFuzzer{Value: string(child2)}}
}

func (ind *Individual) Mutate() {
	ind.Fuzzer = ind.Fuzzer.Mutate()
}

// InsertDelete randomly inserts or deletes a character from a string.
func InsertDelete(s string, p float64) string {
	rr := []rune(s)
	l := len(rr)

	// Insert
	if GenerateRandomBool(p) {
		pos := randRange(0, l-1)
		char := randomASCIIChar()
		rr = append(rr, 0)
		copy(rr[pos+1:], rr[pos:])
		rr[pos] = char
	} else {
		// delete
		if l > 0 {
			pos := randRange(0, l-1)
			rr = append(rr[:pos], rr[pos+1:]...)
		}
	}

	return string(rr)
}

type F struct {
	corpus []string
}

func (f *F) Add(values ...interface{}) []Fuzzer {
	fuzzers := make([]Fuzzer, len(values))

	for i, v := range values {
		f.corpus = append(f.corpus, v.(string))
		fuzzers[i] = &StringFuzzer{Value: v.(string)}
	}

	return fuzzers
}
