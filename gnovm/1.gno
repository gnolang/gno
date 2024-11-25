package main

import (
	"fmt"
	"gno.land/p/demo/entropy"
	"math/rand"
)

// Individual represents a single solution in the population
type Individual struct {
	Genome  string
	Fitness float64
}

// Target is the string we want to evolve
const Target = "HELLO, WORLD!"

// MutationRate controls how often mutations occur
const MutationRate = 0.01

// PopulationSize defines the number of individuals in the population
const PopulationSize = 100

// Charset defines the possible characters in the genome
var Charset = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ, !")

func main() {
	//rand.Seed(time.Now().UnixNano())
	//seed := uint64(entropy.New().Value())
	//rand.Seed(int64(entropy.New().Value()))

	// Step 1: Initialize Population
	population := initializePopulation(PopulationSize)

	generation := 0
	for {
		// Step 2: Calculate Fitness
		for i := range population {
			population[i].Fitness = calculateFitness(population[i].Genome)
		}

		// Step 3: Check for Solution
		best := findBestIndividual(population)
		fmt.Printf("Generation %d: %s (Fitness: %f)\n", generation, best.Genome, best.Fitness)
		if best.Genome == Target {
			fmt.Println("Target reached!")
			break
		}

		// Step 4: Generate Next Generation
		population = generateNextGeneration(population)

		generation++
	}
}

// initializePopulation creates a random population of individuals
func initializePopulation(size int) []Individual {
	population := make([]Individual, size)
	for i := 0; i < size; i++ {
		population[i] = Individual{
			Genome: randomGenome(len(Target)),
		}
	}
	return population
}

// randomGenome generates a random string of the same length as the target
func randomGenome(length int) string {
	genome := make([]rune, length)
	for i := 0; i < length; i++ {
		//genome[i] = Charset[rand.Intn(len(Charset))]
		genome[i] = Charset[entropy.New().Value()]
	}
	return string(genome)
}

// calculateFitness computes how close an individual is to the target string
func calculateFitness(genome string) float64 {
	matches := 0
	for i, char := range genome {
		if char == rune(Target[i]) {
			matches++
		}
	}
	return float64(matches) / float64(len(Target))
}

// findBestIndividual finds the individual with the highest fitness in the population
func findBestIndividual(population []Individual) Individual {
	best := population[0]
	for _, individual := range population {
		if individual.Fitness > best.Fitness {
			best = individual
		}
	}
	return best
}

// generateNextGeneration produces the next generation using selection, crossover, and mutation
func generateNextGeneration(population []Individual) []Individual {
	nextGeneration := make([]Individual, len(population))

	for i := 0; i < len(population); i++ {
		// Selection: Pick two parents
		parent1 := selectIndividual(population)
		parent2 := selectIndividual(population)

		// Crossover: Create a child
		childGenome := crossover(parent1.Genome, parent2.Genome)

		// Mutation: Introduce random changes
		childGenome = mutate(childGenome)

		// Add child to the next generation
		nextGeneration[i] = Individual{
			Genome: childGenome,
		}
	}

	return nextGeneration
}

// selectIndividual selects an individual from the population based on fitness (roulette wheel selection)
func selectIndividual(population []Individual) Individual {
	totalFitness := 0.0
	for _, individual := range population {
		totalFitness += individual.Fitness
	}

	r := rand.Float64() * totalFitness
	cumulative := 0.0
	for _, individual := range population {
		cumulative += individual.Fitness
		if cumulative >= r {
			return individual
		}
	}

	return population[len(population)-1]
}

// crossover combines two parent genomes to produce a child genome
func crossover(parent1, parent2 string) string {
	child := make([]rune, len(parent1))
	cutPoint := rand.Intn(len(parent1))
	for i := 0; i < len(parent1); i++ {
		if i < cutPoint {
			child[i] = rune(parent1[i])
		} else {
			child[i] = rune(parent2[i])
		}
	}
	return string(child)
}

// mutate introduces random changes to a genome based on the mutation rate
func mutate(genome string) string {
	mutated := []rune(genome)
	for i := 0; i < len(mutated); i++ {
		if rand.Float64() < MutationRate {
			mutated[i] = Charset[rand.Intn(len(Charset))]
		}
	}
	return string(mutated)
}
