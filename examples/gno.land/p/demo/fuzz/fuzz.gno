package fuzz

import (
	"crypto/chacha20/rand"
	"encoding/binary"
)

type Fuzzer interface {
	Mutate() Fuzzer
	String() string
}

type StringFuzzer struct {
	Value string
	rng *rand.RNG
}

// NewStringFuzzer creates a new StringFuzzer with the given seed and number of rounds.
func NewStringFuzzer(seed []byte, rounds int) *StringFuzzer {
	rng := rand.NewFromSeed(seed)
	return &StringFuzzer{
		Value: "",
		rng: rng,
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
		rng: sf.rng,
	}
}

func randomASCIIChar(r *rand.RNG) rune {
	return rune(32 + r.Intn(95))
}

func (sf *StringFuzzer) String() string {
	return sf.Value
}

type Individual struct {
	Fuzzer Fuzzer
	Fitness int
}

func (idv *Individual) calculateFitness() {
	idv.Fitness = len(idv.Fuzzer.String())
}

// Selection selects individuals from the population based on their fitness.
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

func Crossover(parent1, parent2 *Individual, rng *rand.RNG) (*Individual, *Individual) {
	str1 := parent1.Fuzzer.String()
	str2 := parent2.Fuzzer.String()

	minLength := len(str1)
	if len(str2) < minLength {
		minLength = len(str2)
	}
	if minLength == 0 {
		return parent1, parent2
	}
	crossOverPoint := rng.Intn(minLength)

	childStr1 := str1[:crossOverPoint] + str2[crossOverPoint:]
	childStr2 := str2[:crossOverPoint] + str1[crossOverPoint:]

	child1 := &Individual{Fuzzer: &StringFuzzer{Value: childStr1}, Fitness: 0}
	child2 := &Individual{Fuzzer: &StringFuzzer{Value: childStr2}, Fitness: 0}

	return child1, child2
}
