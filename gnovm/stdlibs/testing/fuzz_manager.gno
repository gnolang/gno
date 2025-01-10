package testing

import (
	"errors"
	"fmt"
	"os"
	"sort"
)

const (

	// It represents the length of arrays internally managed by the fuzz manager.
	// The maximum length of the array is limited to handle sorting between seeds, out, and in within O(1) time.
	// (However, input can be received infinitely.)
	MaxCapacity uint64 = 100_000

	// PartialCapacity is MaxCapacity divided by 5.
	PartialCapacity = MaxCapacity / 5
)

// fuzzManager is a manager that structurally organizes input seeds to quickly find the 'next input value to use'.
// It sorts them according to priority using internal arrays whenever a new input is added,
// and provides inputs with higher priority when the next input is needed.
type fuzzManager struct {
	// trials is the maximum number of inputs that can be attempted until the fuzz fails.
	// It is set via CLI.
	trials uint

	// inputCount is the number of inputs attempted so far.
	inputCount uint

	// seedType holds the types of seeds.
	seedType []supportedType

	// inithashnumber contains numbers used to search for "real seeds" added via f.Add.
	inithashnumber []hashNumber

	// Arrays for placing and searching seeds
	priorityCache         uint                        // priorityCache is the most recent priority value closest to 0.
	hashNumber2Seeds      *[MaxCapacity]*seedQueue    // hashNumber2Seeds groups seeds with the same hash number.
	hashNumber2Priority   *[MaxCapacity]uint          // hashNumber2Priority records the priority of each hash number.
	priority2HashNumSet   *[MaxCapacity]hashNumberSet // priority2HashNumSet is the set of hash numbers that share the same priority.
	verbose               bool                        // verbose determines the frequency of logging based on the CLI input.
	inspectingHashNumbers uint                        // inspectingHashNumbers is the number of hash numbers currently managed by the manager.
}

// coordinateSeed places a seed into the fuzzManager's internal arrays.
// The placed seed can be selected as the 'next input'.
// It assigns priority based on how many similar seeds have been input so far.
// Seeds with lower priority (i.e., larger priority value) have a reduced probability of being selected next.
func (fm *fuzzManager) coordinateSeed(seed seed) endInfo {
	if seed.isCoordinated {
		fm.hashNumber2Seeds[seed.hashNumber].enqueue(seed)
		return endInfo{
			completeTrials: false,
			maxedCAPACITY:  false,
		}
	}
	hn := seed.hashNumber
	seed.isCoordinated = true
	fm.inputCount++
	// Similar to the aging technique used in OS. It takes the hash number of the seed and lowers its priority.
	// By doing so, sets of values that have been input many times are selected less frequently next time.
	oldPriority := fm.hashNumber2Priority[hn]
	updatedPriority := oldPriority + 1
	fm.hashNumber2Priority[hn] = updatedPriority

	// The zero value for all hashNumbers' priority is 0.
	// This identifies newly encountered hashNumbers.
	// Thus, the seedQueue for the hash number is reliably created (it is not called when nil).
	if oldPriority == 0 {
		fm.inspectingHashNumbers++
		fm.priorityCache = updatedPriority
		fm.hashNumber2Seeds[hn] = newSeedQueue()
	} else {
		// Since hn has already appeared,
		// the hashNumberSet for this priority must exist (it is not nil).
		fm.priority2HashNumSet[oldPriority].remove(hn)
	}
	fm.hashNumber2Seeds[hn].enqueue(seed)
	// Check if the hashNumberSet for updatedPriority is nil.
	if fm.priority2HashNumSet[updatedPriority] == nil {
		fm.priority2HashNumSet[updatedPriority] = newHashNumberSet()
	}
	fm.priority2HashNumSet[updatedPriority].add(hn)
	fm.conditionallyLogProgress(1)
	return fm.evaluateEndState(updatedPriority, hn)
}

// conditionallyLogProgress logs the progress of the fuzzing process based on the number of inputs.
// It prints the progress at certain intervals depending on the verbosity level set via CLI.
func (fm *fuzzManager) conditionallyLogProgress(progressed uint) {
	if fm.verbose {
		if (fm.inputCount-progressed)%2000 > fm.inputCount%2000 {
			fmt.Fprintf(os.Stderr, "%d times runned: inspecting %d coverages\n", fm.inputCount, fm.inspectingHashNumbers)
		}
	} else {
		if (fm.inputCount-progressed)%(fm.trials/5) > fm.inputCount%(fm.trials/5) {
			fmt.Fprintf(os.Stderr, "%d times runned: inspecting %d coverages\n", fm.inputCount, fm.inspectingHashNumbers)
		}
	}
}

type endInfo struct {
	completeTrials bool
	maxedCAPACITY  bool
}

// evaluateEndState checks whether the fuzzing process should end based on the number of inputs.
// It evaluates if 1) the required number of trials has been completed, or
// 2) the fuzzManager's arrays have reached their maximum capacity.
func (fm *fuzzManager) evaluateEndState(updatedPriority uint, hn hashNumber) endInfo {
	if fm.inputCount >= fm.trials {
		return endInfo{
			completeTrials: true,
			maxedCAPACITY:  false,
		}
	} else if uint64(updatedPriority) >= (MaxCapacity-1) || uint64(hn) >= (MaxCapacity-1) {
		// Case 1: Priority reaches MAX. Priority increases discontinuously
		// Case 2: HashNumber reaches Max.
		return endInfo{
			completeTrials: false,
			maxedCAPACITY:  true,
		}
	} else {
		return endInfo{
			completeTrials: false,
			maxedCAPACITY:  false,
		}
	}
}

// coordinateAbstraction processes nested coordinating based on the shared 'hash number' of seed bundles.
// abstractNumber represents how many seeds share the same hash number (i.e., coverage).
// By placing only abstract values without storing seeds, it eliminates phenomena like delay and stagnation.
// This function lowers the priority of the hash number based on the number of shared seeds.
// Thus, hash numbers that appear frequently are selected less often next time.
func (fm *fuzzManager) coordinateAbstraction(hn hashNumber, abstractNumber uint) endInfo {
	fm.inputCount = fm.inputCount + abstractNumber
	// Abstracts the behavior of coordinateSeed. It is a more generalized operation.
	// However, it does not handle concrete seeds.
	oldPriority := fm.hashNumber2Priority[hn]
	updatedPriority := oldPriority + abstractNumber
	fm.hashNumber2Priority[hn] = updatedPriority

	// Identifies if the hashNumber is appearing for the first time.
	if oldPriority == 0 {
		fm.inspectingHashNumbers++
		// Updates the priority cache based on the updated priority.
		if fm.priorityCache > updatedPriority {
			fm.priorityCache = updatedPriority
		}
		// Adds the new hn if a new hashNumber is discovered.
		fm.hashNumber2Seeds[hn] = newSeedQueue()
	} else {
		fm.priority2HashNumSet[oldPriority].remove(hn)
	}
	if fm.priority2HashNumSet[updatedPriority] == nil {
		fm.priority2HashNumSet[updatedPriority] = newHashNumberSet()
	}
	fm.priority2HashNumSet[updatedPriority].add(hn)
	fm.conditionallyLogProgress(abstractNumber)
	return fm.evaluateEndState(updatedPriority, hn)
}

// seeds and hashNumbers are already  sorted by priority within o(1) time automatically by the coordinating function.
// You can also record what your most recent priorities were in the prioriCache to complete your selection in o(1) hours (average case).
func (fm *fuzzManager) popSeeds() []seed {
	var highestProrityHashNumber uint
	// Finds the hash number with the highest priority, i.e., closest to o~o(k)
	for {
		hashNumberSet := fm.priority2HashNumSet[fm.priorityCache]
		if hashNumberSet == nil {
			fm.priorityCache++
			continue
		}
		hn, isExist := fm.priority2HashNumSet[fm.priorityCache].getAny()
		if !isExist {
			fm.priorityCache++
			continue
		}
		highestProrityHashNumber = uint(hn)
		break
	}
	// Dequeues a seed with the corresponding hash number (mapped to coverage).
	popedSeed1, isOnce := fm.hashNumber2Seeds[highestProrityHashNumber].dequeue()
	if !isOnce {
		panic("logical internal error: it must have more than one seed")
	}
	peekedSeed2, err := fm.hashNumber2Seeds[highestProrityHashNumber].peek()
	if err != nil {
		return []seed{popedSeed1}
	}
	// Checks who the 'parent' of each seed is.
	pid1 := popedSeed1.pid
	pid2 := peekedSeed2.pid
	if pid1 == pid2 {
		// Seeds from the same lineage should be popped together for crossover.
		popedSeed2, _ := fm.hashNumber2Seeds[highestProrityHashNumber].dequeue()
		return []seed{popedSeed1, popedSeed2}
	} else {
		// Ensures that seeds from different lineages are not grouped together.
		// This prevents stagnation.
		return []seed{popedSeed1}
	}
}

// popInitSeedByHN retrieves a seed based on its hash number.
// It is used during the initial formatting process of the fuzzer to obtain "initial seeds" based on their hash numbers.
func (fm *fuzzManager) popInitSeedByHN(hn hashNumber) seed {
	popedSeed, isOnce := fm.hashNumber2Seeds[hn].dequeue()
	if !isOnce {
		panic("logical internal error: it must have more than one seed")
	}
	return popedSeed
}

type (
	// priorytyHashNumberList is a list of priority and hash number pairs used for sorting.
	priorytyHashNumberList []priorytyAndHashNumber

	// priorytyAndHashNumber holds a priority and its corresponding hash number.
	priorytyAndHashNumber struct {
		priority   uint
		hashNumber int
	}
)

// Len returns the length of the priorytyHashNumberList.
func (phl priorytyHashNumberList) Len() int { return len(phl) }

// Less compares two elements in the priorytyHashNumberList based on priority.
func (phl priorytyHashNumberList) Less(i, j int) bool { return phl[i].priority < phl[j].priority }

// Swap swaps two elements in the priorytyHashNumberList.
func (phl priorytyHashNumberList) Swap(i, j int) { phl[i], phl[j] = phl[j], phl[i] }

// seedAndPrioryty holds a seed and its associated priority.
type seedAndPrioryty struct {
	seed     seed
	priority uint
}

// summarize compresses the information of high-priority seeds when the fuzzManager's arrays are full.
//  1. It compresses seeds that share the same hash number (i.e., coverage) into one.
//  2. It transforms the compressed seed information based on the degree of compression.
//     The more compressed, the lower the priority.
//
// This compresses the fuzzManager's information into the seed array.
func (fm *fuzzManager) summarize() []seed {
	// Collects all (priority, hashNumber) pairs currently held by the fuzzManager.
	// Each hashNumber has its corresponding priority.
	collectedPriorytyAndHN := priorytyHashNumberList{}
	for i, e := range fm.hashNumber2Priority {
		if e == 0 {
			break
		}
		collectedPriorytyAndHN = append(collectedPriorytyAndHN, priorytyAndHashNumber{
			priority:   e,
			hashNumber: i,
		})
		// Exit when slice length exceeds PartialCapacity
		if len(collectedPriorytyAndHN) >= int(PartialCapacity) {
			break
		}
	}
	sort.Sort(collectedPriorytyAndHN)

	// Creates an array of (seed, priority).
	// At this time, the relationship is seed: priority = n:1,
	// but one seed is selected to form a 1:1 pair, effectively randomizing.
	sampledSeedAndPrioryty := []seedAndPrioryty{}

	for _, pair := range collectedPriorytyAndHN {
		hn := pair.hashNumber
		// Samples the first seed. This is effectively random.
		seed := fm.hashNumber2Seeds[hn].data[0]
		priority := pair.priority
		sampledSeedAndPrioryty = append(sampledSeedAndPrioryty, seedAndPrioryty{seed, priority})
	}

	total := len(sampledSeedAndPrioryty)

	summarizedSeeds := []seed{}
	// Compresses the fuzzManager's information into the seed array.
	// Adjusts seed.gen based on how many similar seeds have been input so far.
	// Lowering seed.gen causes the evolve function to produce more values.
	// Additionally, it causes the seed to be popped earlier.
	for i, seedAndPriority := range sampledSeedAndPrioryty {
		rank := i + 1

		seed := seedAndPriority.seed
		seed.pid = uint(i)
		seed.id = uint(i)
		seed.gen = allocateGen(seedAndPriority.priority, rank, total)
		seed.isCoordinated = false

		summarizedSeeds = append(summarizedSeeds, seed)
	}
	return summarizedSeeds
}

// allocateGen returns the gen value based on priority, rank, and the total number of priorities.
// It is similar to rarity.
func allocateGen(priority uint, rank int, total int) uint {
	var bigger uint

	// The closer the rank is to 1, the lower the gen becomes.
	switch rank {
	case 1:
		bigger = 1
	case 2:
		bigger = 2
	case 3:
		bigger = 3
	case 4:
		bigger = 4
	case 5:
		bigger = 5
	default:
		bigger = 15
	}
	c1CutLine := uint(float64(MaxCapacity) * 0.01)
	c2CutLine := uint(float64(MaxCapacity) * 0.02)
	c3CutLine := uint(float64(MaxCapacity) * 0.03)
	c4CutLine := uint(float64(MaxCapacity) * 0.05)
	c5CutLine := uint(float64(MaxCapacity) * 0.08)
	c6CutLine := uint(float64(MaxCapacity) * 0.1)
	c7CutLine := uint(float64(MaxCapacity) * 0.15)
	c8CutLine := uint(float64(MaxCapacity) * 0.30)
	// Higher rarity results in lower gen
	switch {
	case priority <= c1CutLine:
		if bigger > 1 {
			bigger = 1
		}
	case priority <= c2CutLine:
		if bigger > 2 {
			bigger = 2
		}
	case priority <= c3CutLine:
		if bigger > 3 {
			bigger = 3
		}
	case priority <= c4CutLine:
		if bigger > 4 {
			bigger = 4
		}
	case priority <= c5CutLine:
		if bigger > 5 {
			bigger = 5
		}
	case priority <= c6CutLine:
		if bigger > 6 {
			bigger = 6
		}
	case priority <= c7CutLine:
		if bigger > 7 {
			bigger = 7
		}
	case priority <= c8CutLine:
		if bigger > 10 {
			bigger = 10
		}
	default:

		if bigger > 15 {
			bigger = 15
		}
	}
	return bigger
}

// newFuzzManager creates a new instance of fuzzManager with the specified trials and inputCount.
func newFuzzManager(trials uint, inputCount uint) *fuzzManager {
	return &fuzzManager{
		inputCount: inputCount,
		trials:     trials,

		priorityCache: uint(MaxCapacity),

		hashNumber2Priority: &[MaxCapacity]uint{},
		hashNumber2Seeds:    &[MaxCapacity]*seedQueue{},
		priority2HashNumSet: &[MaxCapacity]hashNumberSet{},

		inspectingHashNumbers: 0,
	}
}

// seedQueue represents a queue of seeds.
type (
	// seedQueue holds a slice of seeds.
	seedQueue struct {
		data []seed
	}

	// seed represents a single seed with its metadata.
	seed struct {
		pid           uint
		id            uint
		gen           uint
		isCoordinated bool
		hashNumber    hashNumber
		content       []interface{}
		result        interface{}
	}
)

// newSeedQueue initializes and returns a new seedQueue.
func newSeedQueue() *seedQueue {
	return &seedQueue{
		data: make([]seed, 0),
	}
}

// enqueue adds a seed to the end of the seedQueue.
func (q *seedQueue) enqueue(seed seed) {
	q.data = append(q.data, seed)
}

// dequeue removes and returns the front seed from the seedQueue.
// It returns a boolean indicating whether the operation was successful.
func (q *seedQueue) dequeue() (seed, bool) {
	if len(q.data) == 0 {
		return seed{}, false
	}

	front := q.data[0]

	q.data = q.data[1:]
	// Reduce memory usage
	if len(q.data) > 0 && len(q.data) <= cap(q.data)/2 {
		newData := make([]seed, len(q.data))
		copy(newData, q.data)
		q.data = newData
	}
	return front, true
}

// peek returns the front seed of the seedQueue without removing it.
// It returns an error if the queue is empty.
func (q *seedQueue) peek() (seed, error) {
	if len(q.data) == 0 {
		return seed{}, errors.New("queue is empty")
	}
	return q.data[0], nil
}

// hashNumberSet represents a set of hashNumbers.
type hashNumberSet map[hashNumber]struct{}

// newHashNumberSet creates and returns a new hashNumberSet.
func newHashNumberSet() hashNumberSet {
	return make(hashNumberSet)
}

// add inserts a hashNumber into the hashNumberSet.
func (s hashNumberSet) add(hn hashNumber) {
	s[hn] = struct{}{}
}

// remove deletes a hashNumber from the hashNumberSet.
func (s hashNumberSet) remove(hn hashNumber) {
	delete(s, hn)
}

// contains checks if a hashNumber exists in the hashNumberSet.
func (s hashNumberSet) contains(hn hashNumber) bool {
	_, exists := s[hn]
	return exists
}

// getAny retrieves any hashNumber from the hashNumberSet.
// It returns the hashNumber and a boolean indicating if it was found.
func (s hashNumberSet) getAny() (hashNumber, bool) {
	for k := range s {
		return k, true
	}
	return 0, false
}

// size returns the number of elements in the hashNumberSet.
func (s hashNumberSet) size() int {
	return len(s)
}

// clear removes all elements from the hashNumberSet.
func (s hashNumberSet) clear() {
	for key := range s {
		delete(s, key)
	}
}

// toSlice converts the hashNumberSet to a slice of hashNumbers.
func (s hashNumberSet) toSlice() []hashNumber {
	keys := make([]hashNumber, 0, len(s))
	for key := range s {
		keys = append(keys, key)
	}
	return keys
}
