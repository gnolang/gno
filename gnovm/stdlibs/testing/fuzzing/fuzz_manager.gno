package fuzzing

import (
	"errors"
	"fmt"
	"os"
)

type SupportedType string

// supportedType is a list of argument types that the target function of f.Fuzz can accept.
// In other words, the function passed to f.Fuzz cannot take complex structs as arguments.
// For example, the target function can receive ([]byte, bool, int32) as args, but not ([][]byte, []int) as a factor.
// The types are selected based on the supported types of Go's fuzzer (the same as Go's fuzzer).
const (
	ByteArray SupportedType = "[]byte"
	String    SupportedType = "string"
	Bool      SupportedType = "bool"
	// Removed due to redundancy. The Byte type is equivalent to uint8, and the Rune type is equivalent to int32.
	// Byte      SupportedType = "byte"
	// Rune      SupportedType = "rune"
	Float32 SupportedType = "float32"
	Float64 SupportedType = "float64"
	Int     SupportedType = "int"
	Int8    SupportedType = "int8"
	Int16   SupportedType = "int16"
	Int32   SupportedType = "int32"
	Int64   SupportedType = "int64"
	Uint    SupportedType = "uint"
	Uint8   SupportedType = "uint8"
	Uint16  SupportedType = "uint16"
	Uint32  SupportedType = "uint32"
	Uint64  SupportedType = "uint64"
)

// DeepCopyInterfaceList performs a deep copy of an array of Gno primitive types.
func DeepCopyInterfaceList(i []interface{}) []interface{} {
	newList := make([]interface{}, len(i))

	for idx, v := range i {
		switch v := v.(type) {
		case []byte:
			newBytes := make([]byte, len(v))
			copy(newBytes, v)
			newList[idx] = newBytes

		case string:
			newList[idx] = v

		case bool, float32, float64,
			int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64:
			newList[idx] = v

		default:
			panic(fmt.Sprintf("Unsupported type in deepCopyInterfaceList: %T", v))
		}
	}

	return newList
}

const (

	// It represents the length of arrays internally managed by the fuzz manager.
	// The maximum length of the array is limited to handle sorting between seeds, out, and in within O(1) time.
	// (However, input can be received infinitely.)
	MaxCapacity uint64 = 1000_000
	// PartialCapacity is MaxCapacity divided by 5.
	PartialCapacity = MaxCapacity / 5
)

// TODO: The mutator should handle the storage of seeds.
// It may be sufficient for the manager to only handle metadata, excluding the seeds themselves.
// Consider revising the design to improve space efficiency, taking this into account.

// FuzzManager is a manager that structurally organizes input seeds to quickly find the 'next input value to use'.
// It sorts them according to priority using internal arrays whenever a new input is added,
// and provides inputs with higher priority when the next input is needed.
type FuzzManager struct {
	// Trials is the maximum number of inputs that can be attempted until the fuzz fails.
	// It is set via CLI.
	Trials uint

	// InputCount is the number of inputs attempted so far.
	InputCount uint

	// SeedType holds the types of seeds.
	SeedType []SupportedType

	// Inithashnumber contains numbers used to search for "real seeds" added via f.Add.
	Inithashnumber []HashNumber

	// Arrays for placing and searching seeds
	PriorityCache         uint                        // priorityCache is the most recent priority value closest to 0.
	HashNumber2Seeds      *[MaxCapacity]*SeedQueue    // hashNumber2Seeds groups seeds with the same hash number.
	HashNumber2Priority   *[MaxCapacity]uint          // hashNumber2Priority records the priority of each hash number.
	Priority2HashNumSet   *[MaxCapacity]hashNumberSet // priority2HashNumSet is the set of hash numbers that share the same priority.
	Verbose               bool                        // verbose determines the frequency of logging based on the CLI input.
	InspectingHashNumbers uint                        // inspectingHashNumbers is the number of hash numbers currently managed by the manager.
}

// seedQueue represents a queue of seeds.
type SeedQueue struct {
	Data []Seed
}

// Seed represents a set of input data for the target program and its associated metadata.
type Seed struct {
	Pid            uint                 // Pid is the ID of the parent seed that this seed was generated from.
	Id             uint                 // Id is the unique identifier of the seed.
	Gen            uint                 // Gen is the generation number of the seed.
	IsCoordinated  bool                 // IsCoordinated indicates whether the seed has been coordinated with the manager.
	HashNumber     HashNumber           // HashNumber is the hash number representing the seed's coverage.
	Content        []interface{}        // Content represents the input values of the seed.
	MutationsInfo  map[int]MutationInfo // MutationsInfo holds detailed information about how the seed has mutated from its original form.
	ArgPowerSetIdx int                  // Refers to the index of argsPowerSet, indicating which subset has been selected.
	Result         interface{}          // Result is the output produced by the program when the seed is used as input.
}

// MutationInfo holds information about a mutation operation performed on an entity.
type MutationInfo struct {
	IsMutated            bool                         // IsMutated indicates whether the entity has been mutated.
	IsStringOrByteArray  bool                         // IsStringOrByteArray is true if the entity is a string or a byte array.
	EntId                int                          // EntId is the id of the entity being mutated.
	UsedStrategyMap      map[MutationCode]int         // UsedStrategyMap records the count of each mutation strategy used on the entity.
	UsedCodesForNotSbMap map[MutationCodeForNotSB]int // UsedCodesForNotSbMap records the mutation codes used for non-string or non-byte-array arguments.
}

// seed represents a single seed with its metadata.

// coordinateSeed places a seed into the fuzzManager's internal arrays.
// The placed seed can be selected as the 'next input'.
// It assigns priority based on how many similar seeds have been input so far.
// Seeds with lower priority (i.e., larger priority value) have a reduced probability of being selected next.
func (fm *FuzzManager) CoordinateSeed(seed Seed) EndInfo {
	if seed.IsCoordinated {
		fm.HashNumber2Seeds[seed.HashNumber].enqueue(seed)
		return EndInfo{
			CompleteTrials: false,
			MaxedCAPACITY:  false,
		}
	}
	hn := seed.HashNumber
	seed.IsCoordinated = true
	fm.InputCount++
	// Similar to the aging technique used in OS. It takes the hash number of the seed and lowers its priority.
	// By doing so, sets of values that have been input many times are selected less frequently next time.
	oldPriority := fm.HashNumber2Priority[hn]
	updatedPriority := oldPriority + 1
	fm.HashNumber2Priority[hn] = updatedPriority

	// The zero value for all hashNumbers' priority is 0.
	// This identifies newly encountered hashNumbers.
	// Thus, the seedQueue for the hash number is reliably created (it is not called when nil).
	if oldPriority == 0 {
		fm.InspectingHashNumbers++
		fm.PriorityCache = updatedPriority
		fm.HashNumber2Seeds[hn] = NewSeedQueue()
	} else {
		// Since hn has already appeared,
		// the hashNumberSet for this priority must exist (it is not nil).
		fm.Priority2HashNumSet[oldPriority].remove(hn)
	}
	fm.HashNumber2Seeds[hn].enqueue(seed)
	// Check if the hashNumberSet for updatedPriority is nil.
	if fm.Priority2HashNumSet[updatedPriority] == nil {
		fm.Priority2HashNumSet[updatedPriority] = newHashNumberSet()
	}
	fm.Priority2HashNumSet[updatedPriority].add(hn)
	fm.ConditionallyLogProgress(1)
	return fm.CheckEndCondition(updatedPriority, hn)
}

// conditionallyLogProgress logs the progress of the fuzzing process based on the number of inputs.
// It prints the progress at certain intervals depending on the verbosity level set via CLI.
func (fm *FuzzManager) ConditionallyLogProgress(progressed uint) {
	if fm.Verbose {
		if (fm.InputCount-progressed)%5000 > fm.InputCount%5000 {
			fmt.Fprintf(os.Stderr, "%d times runned: inspecting %d coverages\n", fm.InputCount, fm.InspectingHashNumbers)
		}
	} else {
		if (fm.InputCount-progressed)%(fm.Trials/5) > fm.InputCount%(fm.Trials/5) {
			fmt.Fprintf(os.Stderr, "%d times runned: inspecting %d coverages\n", fm.InputCount, fm.InspectingHashNumbers)
		}
	}
}

type EndInfo struct {
	CompleteTrials bool
	MaxedCAPACITY  bool
}

// checkEndCondition checks whether the fuzzing process should end based on the number of inputs and capacity of fuzz manager.
// It checks if
// 1) the required number of trials has been completed, or
// 2) the fuzzManager's arrays have reached their maximum capacity.
func (fm *FuzzManager) CheckEndCondition(updatedPriority uint, hn HashNumber) EndInfo {
	if fm.InputCount >= fm.Trials {
		return EndInfo{
			CompleteTrials: true,
			MaxedCAPACITY:  false,
		}
	} else if uint64(updatedPriority) >= (MaxCapacity - 1) {
		return EndInfo{
			CompleteTrials: false,
			MaxedCAPACITY:  true,
		}
	} else if uint64(hn) >= (MaxCapacity - 1) {
		fmt.Fprintf(os.Stderr, "The fuzzer has explored all %v unique coverage paths. Further coverage path exploration is not supported, so fuzzing will be terminated.\n", MaxCapacity)
		return EndInfo{
			CompleteTrials: true,
			MaxedCAPACITY:  true,
		}
	} else {
		return EndInfo{
			CompleteTrials: false,
			MaxedCAPACITY:  false,
		}
	}
}

// coordinateAbstraction processes nested coordinating based on the shared 'hash number' of seed bundles.
// abstractNumber represents how many seeds share the same hash number (i.e., coverage).
// By placing only abstract values without storing seeds, it eliminates phenomena like delay and stagnation.
// This function lowers the priority of the hash number based on the number of shared seeds.
// Thus, hash numbers that appear frequently are selected less often next time.
func (fm *FuzzManager) CoordinateAbstraction(hn HashNumber, abstractNumber uint) EndInfo {
	fm.InputCount = fm.InputCount + abstractNumber
	// Abstracts the behavior of coordinateSeed. It is a more generalized operation.
	// However, it does not handle concrete seeds.
	oldPriority := fm.HashNumber2Priority[hn]
	updatedPriority := oldPriority + abstractNumber
	fm.HashNumber2Priority[hn] = updatedPriority
	// checks precondition
	endInfo := fm.CheckEndCondition(updatedPriority, hn)
	if endInfo.MaxedCAPACITY || endInfo.CompleteTrials {
		return endInfo
	}

	// Identifies if the hashNumber is appearing for the first time.
	if oldPriority == 0 {
		fm.InspectingHashNumbers++
		// Updates the priority cache based on the updated priority.
		if fm.PriorityCache > updatedPriority {
			fm.PriorityCache = updatedPriority
		}
		// Adds the new hn if a new hashNumber is discovered.
		fm.HashNumber2Seeds[hn] = NewSeedQueue()
	} else {
		fm.Priority2HashNumSet[oldPriority].remove(hn)
	}
	if fm.Priority2HashNumSet[updatedPriority] == nil {
		fm.Priority2HashNumSet[updatedPriority] = newHashNumberSet()
	}
	fm.Priority2HashNumSet[updatedPriority].add(hn)
	fm.ConditionallyLogProgress(abstractNumber)
	return endInfo
}

func (fm *FuzzManager) IsCoordinatedHashNumber(hn HashNumber) bool {
	priority := fm.HashNumber2Priority[hn]
	if priority == 0 {
		return false
	} else {
		return true
	}
}

// seeds and hashNumbers are already  sorted by priority within o(1) time automatically by the coordinating function.
// You can also record what your most recent priorities were in the prioriCache to complete your selection in o(1) hours (average case).
func (fm *FuzzManager) PopSeed() Seed {
	var highestPriorityHashNumber uint
	// Finds the hash number with the highest priority, i.e., closest to o~o(k)
	for {
		hashNumberSet := fm.Priority2HashNumSet[fm.PriorityCache]
		if hashNumberSet == nil {
			fm.PriorityCache++
			continue
		}
		hn, isExist := fm.Priority2HashNumSet[fm.PriorityCache].anyHashNumber()
		if !isExist {
			fm.PriorityCache++
			continue
		}
		highestPriorityHashNumber = uint(hn)
		break
	}
	// Dequeues a seed with the corresponding hash number (mapped to coverage).
	popedSeed, isOnce := fm.HashNumber2Seeds[highestPriorityHashNumber].dequeue()
	if !isOnce {
		panic("logical internal error(PopSeed): it must have more than one seed")
	}
	return popedSeed
}

// popInitSeedByHN retrieves a seed based on its hash number.
// It is used during the initial formatting process of the fuzzer to obtain "initial seeds" based on their hash numbers.
func (fm *FuzzManager) PopInitSeedByHN(hn HashNumber) Seed {
	popedSeed, isOnce := fm.HashNumber2Seeds[hn].dequeue()
	if !isOnce {
		panic("logical internal error(PopInitSeedByHN): it must have more than one seed")
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

// SummarizeSeed simply collects and returns all the Seeds stored inside the FuzzManager.
func (fm *FuzzManager) SummarizeSeeds() []Seed {
	var seeds []Seed
	// fm.HashNumber2Seeds is of type *[MaxCapacity]*SeedQueue, so we iterate through each index,
	// and if the SeedQueue is not nil, we append the Seeds inside it.
	for i := 0; i < int(MaxCapacity); i++ {
		queue := fm.HashNumber2Seeds[i]
		if queue != nil {
			for _, s := range queue.Data {
				s.IsCoordinated = false
				seeds = append(seeds, s)
			}
		}
	}
	return seeds
}

// NewFuzzManager creates a new instance of fuzzManager with the specified trials and inputCount.
func NewFuzzManager(trials uint, inputCount uint) *FuzzManager {
	return &FuzzManager{
		InputCount: inputCount,
		Trials:     trials,

		PriorityCache: uint(MaxCapacity),

		HashNumber2Priority: &[MaxCapacity]uint{},
		HashNumber2Seeds:    &[MaxCapacity]*SeedQueue{},
		Priority2HashNumSet: &[MaxCapacity]hashNumberSet{},

		InspectingHashNumbers: 0,
	}
}

// NewSeedQueue initializes and returns a new seedQueue.
func NewSeedQueue() *SeedQueue {
	return &SeedQueue{
		Data: make([]Seed, 0),
	}
}

// enqueue adds a seed to the end of the seedQueue.
func (q *SeedQueue) enqueue(seed Seed) {
	q.Data = append(q.Data, seed)
}

// dequeue removes and returns the front seed from the seedQueue.
// It returns a boolean indicating whether the operation was successful.
func (q *SeedQueue) dequeue() (Seed, bool) {
	if len(q.Data) == 0 {
		return Seed{}, false
	}

	front := q.Data[0]

	q.Data = q.Data[1:]
	// Reduce memory usage
	if len(q.Data) > 0 && len(q.Data) <= cap(q.Data)/2 {
		newData := make([]Seed, len(q.Data))
		copy(newData, q.Data)
		q.Data = newData
	}
	return front, true
}

// peek returns the front seed of the seedQueue without removing it.
// It returns an error if the queue is empty.
func (q *SeedQueue) peek() (Seed, error) {
	if len(q.Data) == 0 {
		return Seed{}, errors.New("queue is empty")
	}
	return q.Data[0], nil
}

// hashNumberSet represents a set of hashNumbers.
type hashNumberSet map[HashNumber]struct{}

// newHashNumberSet creates and returns a new hashNumberSet.
func newHashNumberSet() hashNumberSet {
	return make(hashNumberSet)
}

// add inserts a hashNumber into the hashNumberSet.
func (s hashNumberSet) add(hn HashNumber) {
	s[hn] = struct{}{}
}

// remove deletes a hashNumber from the hashNumberSet.
func (s hashNumberSet) remove(hn HashNumber) {
	delete(s, hn)
}

// contains checks if a hashNumber exists in the hashNumberSet.
func (s hashNumberSet) contains(hn HashNumber) bool {
	_, exists := s[hn]
	return exists
}

// anyHashNumber retrieves any hashNumber from the hashNumberSet.
// It returns the hashNumber and a boolean indicating if it was found.
func (s hashNumberSet) anyHashNumber() (HashNumber, bool) {
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
func (s hashNumberSet) toSlice() []HashNumber {
	keys := make([]HashNumber, 0, len(s))
	for key := range s {
		keys = append(keys, key)
	}
	return keys
}
