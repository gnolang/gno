package ga

import (
	"math"

	"gno.land/p/demo/entropy"
)

// Individual represents a single solution in the population
type Individual struct {
	Genome  string
	Fitness float64
}

// Target is the string we want to evolve
const Target = "HEY"

// MutationRate controls how often mutations occur
const MutationRate = 0.01

// PopulationSize defines the number of individuals in the population
const PopulationSize = 100

// Charset defines the possible characters in the genome
var Charset = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ, !")

func Start(ei *entropy.Instance) (int, string) {
	var generation = 0
	// Step 1: Initialize Population
	population := initializePopulation(PopulationSize, ei)

	for generation < 200 {
		// Step 2: Calculate Fitness
		for i := range population {
			population[i].Fitness = calculateFitness(population[i].Genome)
		}

		// Step 3: Check for Solution
		best := findBestIndividual(population)
		//fmt.Printf("Generation %d: %s (Fitness: %f)\n", generation, best.Genome, best.Fitness)
		if best.Genome == Target {
			println("Target reached!")
			return generation, string(best.Genome)
		}

		// Step 4: Generate Next Generation
		population = generateNextGeneration(population, ei)

		generation++
	}
	return generation, ""
}

// initializePopulation creates a random population of individuals
func initializePopulation(size int, ei *entropy.Instance) []Individual {
	population := make([]Individual, size)
	for i := 0; i < size; i++ {
		population[i] = Individual{
			Genome: randomGenome(len(Target), ei), // new random each loop
		}
	}
	return population
}

// randomGenome generates a random string of the same length as the target
func randomGenome(length int, ei *entropy.Instance) string {
	genome := make([]rune, length)
	for i := 0; i < length; i++ {
		index := ei.Value() % uint32(len(Charset))
		genome[i] = Charset[index]
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
func generateNextGeneration(population []Individual, ei *entropy.Instance) []Individual {
	//println("---generate next generation")
	nextGeneration := make([]Individual, len(population))

	for i := 0; i < len(population); i++ {
		// Selection: Pick two parents
		parent1 := selectIndividual(population, ei)
		parent2 := selectIndividual(population, ei)

		// Crossover: Create a child
		childGenome := crossover(parent1.Genome, parent2.Genome, ei)

		// Mutation: Introduce random changes
		childGenome = mutate(childGenome, ei)

		// Add child to the next generation
		nextGeneration[i] = Individual{
			Genome: childGenome,
		}
	}

	return nextGeneration
}

// selectIndividual selects an individual from the population based on fitness (roulette wheel selection)
func selectIndividual(population []Individual, ei *entropy.Instance) Individual {
	totalFitness := 0.0
	for _, individual := range population {
		totalFitness += individual.Fitness
	}
	r := (float64(ei.Value()) / float64(math.MaxUint32)) * totalFitness

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
func crossover(parent1, parent2 string, ei *entropy.Instance) string {
	child := make([]rune, len(parent1))

	cutPoint := int(ei.Value()) % len(parent1)

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
func mutate(genome string, ei *entropy.Instance) string {
	mutated := []rune(genome)
	for i := 0; i < len(mutated); i++ {
		rv := ei.Value()
		if float64(rv) < MutationRate {
			index := rv % uint32(len(Charset))
			mutated[i] = Charset[index]
		}
	}
	return string(mutated)
}
