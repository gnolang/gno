package fuzzing

import (
	"fmt"
	"math"
	"os"
	"sort"
)

// Mutator manages mutation strategies and clusters mutation samples based on hash numbers.
type Mutator struct {
	seedType                []SupportedType // Type of seeds managed by the mutator.
	argsPowerListExceptNull [][]int         // Power set of argument indices excluding the empty set.
	// Example for two arguments: [[0], [1], [0,1]]
	lenPowerList int                          // Length of the power set (= 2^len(args) - 1).
	data         map[HashNumber]*MutSampleSet // Clusters mutation samples by hash number.
}

// Coordinate2SampleSet adds a new mutation sample to the appropriate cluster by hash number.
func (m *Mutator) Coordinate2SampleSet(seed Seed, abstractPopularity int) {
	newSample := NewMutSample(seed)
	if m.data[seed.HashNumber] == nil {
		m.data[seed.HashNumber] = NewMutSampleSet()
	}
	m.data[seed.HashNumber].Add(newSample)
	m.data[seed.HashNumber].PlusClusterPopularity(abstractPopularity)
}

// MutateSeed2Seeds selects an optimal mutation strategy and generates new mutated seeds.
// The fuzzer selects the highest-priority seed through the manager,
// and the mutator applies an optimized mutation strategy to the seed,
// transforming it and generating offspring.
func (m *Mutator) MutateSeed2Seeds(sampleSeed Seed, seedCount *uint) []Seed {
	hashNumber := sampleSeed.HashNumber
	seedId := sampleSeed.Id
	sampleSet, foundSampleSet := m.FindMutSampleSet(hashNumber)
	if !foundSampleSet {
		panic("Mutator could not find sample set by hash number. Possible logic or runtime error.")
	}
	mutationSample, foundSample := sampleSet.FindSample(seedId)
	if !foundSample {
		panic("Mutator could not find sample set by hash number. Possible logic or runtime error.")
	}
	const NumGenerateSetting = 8
	boundarySamplePopularity := [NumGenerateSetting]uint64{2, 8, 32, 128, 512, 2048, 8096, 16192}
	mappedGeneratingNumber := [NumGenerateSetting]int{1440, 1080, 720, 600, 480, 360, 240, 120}
	genIdx := sort.Search(len(boundarySamplePopularity), func(i int) bool {
		return mutationSample.samplePopularity <= boundarySamplePopularity[i]
	})
	var generatingNumber int
	if genIdx < len(boundarySamplePopularity) {
		generatingNumber = mappedGeneratingNumber[genIdx]
	} else {
		generatingNumber = 120
	}
	const NumStrengthSetting = 4
	boundaryClusterPopularity := [NumStrengthSetting]uint64{32, 256, 4096, 259072}
	mappedMutationStrength := [NumStrengthSetting]int{5, 4, 3, 2}
	mutIdx := sort.Search(len(boundaryClusterPopularity), func(i int) bool {
		return m.data[mutationSample.seed.HashNumber].ClusterPopularity <= boundaryClusterPopularity[i]
	})
	var mutationStrength int
	if mutIdx < len(boundaryClusterPopularity) {
		mutationStrength = mappedMutationStrength[mutIdx]
	} else {
		mutationStrength = 1
	}

	mutationSample.UpdatePowerSetCache()
	cum := mutationSample.PowerSetCache()

	chunk := 120 //* Must be a divisor of the generating number.

	chunkIters := generatingNumber / chunk
	var generatedSeeds []Seed
	for i := 0; i < chunkIters; i++ {
		var selectedPowerSetIdx int
		if len(cum) == 0 || cum[len(cum)-1] == 0 {
			selectedPowerSetIdx = 0
		} else {
			total := cum[len(cum)-1]
			r := OncePCG().Uint64n(total)
			selectedPowerSetIdx = sort.Search(len(cum), func(i int) bool {
				return cum[i] > r
			})
		}
		selectedSubSetArgs := m.argsPowerListExceptNull[selectedPowerSetIdx]

		// Completed the decision of the argument set to mutate.
		// From this loop onwards, actual seed generation begins.

		for j := 0; j < chunk; j++ {
			castedContent := efficientCasting(mutationSample.seed.Content)
			mutationsInfo := make(map[int]MutationInfo)

			for _, idx := range selectedSubSetArgs {
				// Mutate the selected index element.
				var mutationInfo MutationInfo
				var mutatedArgument interface{}
				// Modify the value and receive the mutated data.
				mutatedArgument, mutationInfo = mutationSample.MutateArg(castedContent[idx], mutationStrength, idx)
				if mutatedArgument == nil {
					panic("MutateArg returned nil")
				}
				castedContent[idx] = mutatedArgument
				mutationsInfo[idx] = mutationInfo
			}
			// Synchronize seed counting.
			*seedCount++
			// Generate seed.
			newSeed := Seed{
				Content:        castedContent,
				Pid:            mutationSample.seed.Id,
				Id:             *seedCount,
				Gen:            mutationSample.seed.Gen + 1,
				ArgPowerSetIdx: selectedPowerSetIdx,
				MutationsInfo:  mutationsInfo,
			}
			generatedSeeds = append(generatedSeeds, newSeed)
		}
	}

	return generatedSeeds
}

//	FindMutSampleSet safely returns the MutSampleSet corresponding to the given hn.
//
// If it does not exist, it returns nil and false.
func (m *Mutator) FindMutSampleSet(hn HashNumber) (*MutSampleSet, bool) {
	mss, exists := m.data[hn]
	return mss, exists
}

// efficientCasting creates a shallow copy of the given content slice,
// replacing string and byte slice elements with placeholders.
// This ensures that mutation logic in the mutator can handle
// these types separately without modifying them here.
func efficientCasting(content []interface{}) []interface{} {
	copiedContent := make([]interface{}, len(content))
	for i, v := range content {
		switch v.(type) {
		case string:
			copiedContent[i] = string(" ")
		case []byte:
			copiedContent[i] = []byte(" ")
		default:
			copiedContent[i] = v
		}
	}
	return copiedContent
}

// HasHashNumber checks if the given hash number (hn) exists
// in the Mutator's data map.
func (m *Mutator) HasHashNumber(hn HashNumber) bool {
	_, exists := m.data[hn]
	return exists
}

// NewMutator initializes and returns a new Mutator instance
// with the given seedType. It precomputes the power set of
// argument indices excluding the empty set.
func NewMutator(seedType []SupportedType) *Mutator {
	return &Mutator{
		seedType:                seedType,
		argsPowerListExceptNull: buildArgsPowerListExceptNull(len(seedType)),
		lenPowerList:            (1 << len(seedType)) - 1, // equals to 2^n - 1
		data:                    make(map[HashNumber]*MutSampleSet),
	}
}

// buildArgsPowerListExceptNull generates all subsets of
// argument indices except the empty set.
//
// For example, if n == 2, the output will be:
// [[0], [1], [0,1]]
func buildArgsPowerListExceptNull(n int) [][]int {
	var result [][]int
	// The mask ranges from 1 to 2^n - 1 (0 is excluded as it represents the empty set)
	for mask := 1; mask < (1 << n); mask++ {
		var subset []int
		for j := 0; j < n; j++ {
			if mask&(1<<j) != 0 {
				subset = append(subset, j)
			}
		}
		result = append(result, subset)
	}
	return result
}

// MutSampleSet manages a collection of mutation samples
// and tracks the overall popularity of the cluster.
type MutSampleSet struct {
	Samples           map[uint]*MutSample // ID-based dictionary of mutation samples
	ClusterPopularity uint64              // Aggregate popularity score of the cluster
}

// GetAllSamplesList returns a slice containing all mutation samples in the set.
func (mss *MutSampleSet) GetAllSamplesList() []*MutSample {
	samples := make([]*MutSample, 0, len(mss.Samples))
	for _, sample := range mss.Samples {
		samples = append(samples, sample)
	}
	return samples
}

// FindSample retrieves a MutSample by its seed ID.
// It returns the sample and a boolean indicating success.
func (mss *MutSampleSet) FindSample(seedId uint) (*MutSample, bool) {
	sample, exists := mss.Samples[seedId]
	return sample, exists
}

// FindSiblingID searches for a sibling sample within the set
// that shares the given parent ID (pid).
func (mss *MutSampleSet) FindSiblingID(pid uint) (uint, bool) {
	for id, sample := range mss.Samples {
		if sample.seed.Pid == pid {
			return id, true
		}
	}
	return 0, false
}

// FindParentID locates a sample that acts as a parent
// to the given child ID (pid).
func (mss *MutSampleSet) FindParentID(pid uint) (uint, bool) {
	for id, sample := range mss.Samples {
		if sample.seed.Id == pid {
			return id, true
		}
	}
	return 0, false
}

// Add inserts a new sample into the MutSampleSet.
func (mss *MutSampleSet) Add(sample *MutSample) {
	mss.Samples[sample.seed.Id] = sample
}

// PlusClusterPopularity increases the cluster's popularity score.
func (mss *MutSampleSet) PlusClusterPopularity(delta int) {
	mss.ClusterPopularity += uint64(delta)
}

// IsEmpty returns true if there are no samples in the set.
func (mss *MutSampleSet) IsEmpty() bool {
	return len(mss.Samples) == 0
}

// Size returns the number of samples in the set.
func (mss *MutSampleSet) Size() int {
	return len(mss.Samples)
}

// NewMutSampleSet creates an empty MutSampleSet.
func NewMutSampleSet() *MutSampleSet {
	return &MutSampleSet{
		Samples:           make(map[uint]*MutSample),
		ClusterPopularity: 0,
	}
}

// MutSample represents a mutation sample containing the original seed,
// scoring information, and mappings to entity representations.
type MutSample struct {
	seed              Seed             // Original seed data
	argPowerSetScores []uint64         // Argument scores, considering aspects like [arguments, entity, strategy]
	samplePopularity  uint64           // Number of times the sample has been leveraged
	rootEntsDict      map[int]*RootEnt // Mapping of argument index to RootEnt
	notSbEntsDict     map[int]*EntForNotSB

	// Caching fields for cumulative sum calculations
	cachedPowerSetCumSum []uint64
	needsUpdateCache     bool
}

// MutateArg applies a mutation to the given input argument based on the specified strength
// and argument index. The mutation method is determined by the data type of the input.
func (ms *MutSample) MutateArg(input interface{}, strength int, argIdx int) (interface{}, MutationInfo) {
	switch v := input.(type) {
	case string:
		result, entID, usedStrategyMap := MutateRoot2String(ms.rootEntsDict[argIdx], strength)
		isMutated := true
		if usedStrategyMap == nil {
			isMutated = false
		}
		return result, MutationInfo{
			IsMutated:            isMutated,
			IsStringOrByteArray:  true,
			EntId:                entID,
			UsedStrategyMap:      usedStrategyMap,
			UsedCodesForNotSbMap: nil,
		}
	case []byte:
		result, entID, usedStrategyMap := MutateRoot2String(ms.rootEntsDict[argIdx], strength)
		isMutated := true
		if usedStrategyMap == nil {
			isMutated = false
		}
		return []byte(result), MutationInfo{
			IsMutated:            isMutated,
			IsStringOrByteArray:  true,
			EntId:                entID,
			UsedStrategyMap:      usedStrategyMap,
			UsedCodesForNotSbMap: nil,
		}
	case bool:
		result, usedStrategyMap := MutateNotSbEnt2interface(ms.notSbEntsDict[argIdx], strength)
		boolResult, ok := result.(bool)
		if !ok {
			panicWithExpectedType(Bool, result)
		}
		isMutated := true
		if usedStrategyMap == nil {
			isMutated = false
		}
		return boolResult, MutationInfo{
			IsMutated:            isMutated,
			IsStringOrByteArray:  false,
			EntId:                0,
			UsedStrategyMap:      nil,
			UsedCodesForNotSbMap: usedStrategyMap,
		}

	case float32:
		result, usedStrategyMap := MutateNotSbEnt2interface(ms.notSbEntsDict[argIdx], strength)
		float32Result, ok := result.(float32)
		if !ok {
			panicWithExpectedType(Float32, result)
		}
		isMutated := true
		if usedStrategyMap == nil {
			isMutated = false
		}
		return float32Result, MutationInfo{
			IsMutated:            isMutated,
			IsStringOrByteArray:  false,
			EntId:                0,
			UsedStrategyMap:      nil,
			UsedCodesForNotSbMap: usedStrategyMap,
		}
	case float64:
		result, usedStrategyMap := MutateNotSbEnt2interface(ms.notSbEntsDict[argIdx], strength)
		float64Result, ok := result.(float64)
		if !ok {
			panicWithExpectedType(Float64, result)
		}
		isMutated := true
		if usedStrategyMap == nil {
			isMutated = false
		}
		return float64Result, MutationInfo{
			IsMutated:            isMutated,
			IsStringOrByteArray:  false,
			EntId:                0,
			UsedStrategyMap:      nil,
			UsedCodesForNotSbMap: usedStrategyMap,
		}
	case int:
		result, usedStrategyMap := MutateNotSbEnt2interface(ms.notSbEntsDict[argIdx], strength)
		intResult, ok := result.(int)
		if !ok {
			panicWithExpectedType(Int, result)
		}
		isMutated := true
		if usedStrategyMap == nil {
			isMutated = false
		}
		return intResult, MutationInfo{
			IsMutated:            isMutated,
			IsStringOrByteArray:  false,
			EntId:                0,
			UsedStrategyMap:      nil,
			UsedCodesForNotSbMap: usedStrategyMap,
		}

	case int8:
		result, usedStrategyMap := MutateNotSbEnt2interface(ms.notSbEntsDict[argIdx], strength)
		int8Result, ok := result.(int8)
		if !ok {
			panicWithExpectedType(Int8, result)
		}
		isMutated := true
		if usedStrategyMap == nil {
			isMutated = false
		}
		return int8Result, MutationInfo{
			IsMutated:            isMutated,
			IsStringOrByteArray:  false,
			EntId:                0,
			UsedStrategyMap:      nil,
			UsedCodesForNotSbMap: usedStrategyMap,
		}

	case int16:
		result, usedStrategyMap := MutateNotSbEnt2interface(ms.notSbEntsDict[argIdx], strength)
		int16Result, ok := result.(int16)
		if !ok {
			panicWithExpectedType(Int16, result)
		}
		isMutated := true
		if usedStrategyMap == nil {
			isMutated = false
		}
		return int16Result, MutationInfo{
			IsMutated:            isMutated,
			IsStringOrByteArray:  false,
			EntId:                0,
			UsedStrategyMap:      nil,
			UsedCodesForNotSbMap: usedStrategyMap,
		}

	case int32:
		result, usedStrategyMap := MutateNotSbEnt2interface(ms.notSbEntsDict[argIdx], strength)
		int32Result, ok := result.(int32)
		if !ok {
			panicWithExpectedType(Int32, result)
		}
		isMutated := true
		if usedStrategyMap == nil {
			isMutated = false
		}
		return int32Result, MutationInfo{
			IsMutated:            isMutated,
			IsStringOrByteArray:  false,
			EntId:                0,
			UsedStrategyMap:      nil,
			UsedCodesForNotSbMap: usedStrategyMap,
		}

	case int64:
		result, usedStrategyMap := MutateNotSbEnt2interface(ms.notSbEntsDict[argIdx], strength)
		int64Result, ok := result.(int64)
		if !ok {
			panicWithExpectedType(Int64, result)
		}
		isMutated := true
		if usedStrategyMap == nil {
			isMutated = false
		}
		return int64Result, MutationInfo{
			IsMutated:            isMutated,
			IsStringOrByteArray:  false,
			EntId:                0,
			UsedStrategyMap:      nil,
			UsedCodesForNotSbMap: usedStrategyMap,
		}

	case uint:
		result, usedStrategyMap := MutateNotSbEnt2interface(ms.notSbEntsDict[argIdx], strength)
		uintResult, ok := result.(uint)
		if !ok {
			panicWithExpectedType(Uint, result)
		}
		isMutated := true
		if usedStrategyMap == nil {
			isMutated = false
		}
		return uintResult, MutationInfo{
			IsMutated:            isMutated,
			IsStringOrByteArray:  false,
			EntId:                0,
			UsedStrategyMap:      nil,
			UsedCodesForNotSbMap: usedStrategyMap,
		}

	case uint8:
		result, usedStrategyMap := MutateNotSbEnt2interface(ms.notSbEntsDict[argIdx], strength)
		uint8Result, ok := result.(uint8)
		if !ok {
			panicWithExpectedType(Uint8, result)
		}
		isMutated := true
		if usedStrategyMap == nil {
			isMutated = false
		}
		return uint8Result, MutationInfo{
			IsMutated:            isMutated,
			IsStringOrByteArray:  false,
			EntId:                0,
			UsedStrategyMap:      nil,
			UsedCodesForNotSbMap: usedStrategyMap,
		}

	case uint16:
		result, usedStrategyMap := MutateNotSbEnt2interface(ms.notSbEntsDict[argIdx], strength)
		uint16Result, ok := result.(uint16)
		if !ok {
			panicWithExpectedType(Uint16, result)
		}
		isMutated := true
		if usedStrategyMap == nil {
			isMutated = false
		}
		return uint16Result, MutationInfo{
			IsMutated:            isMutated,
			IsStringOrByteArray:  false,
			EntId:                0,
			UsedStrategyMap:      nil,
			UsedCodesForNotSbMap: usedStrategyMap,
		}

	case uint32:
		result, usedStrategyMap := MutateNotSbEnt2interface(ms.notSbEntsDict[argIdx], strength)
		uint32Result, ok := result.(uint32)
		if !ok {
			panicWithExpectedType(Uint32, result)
		}
		isMutated := true
		if usedStrategyMap == nil {
			isMutated = false
		}
		return uint32Result, MutationInfo{
			IsMutated:            isMutated,
			IsStringOrByteArray:  false,
			EntId:                0,
			UsedStrategyMap:      nil,
			UsedCodesForNotSbMap: usedStrategyMap,
		}

	case uint64:
		result, usedStrategyMap := MutateNotSbEnt2interface(ms.notSbEntsDict[argIdx], strength)
		uint64Result, ok := result.(uint64)
		if !ok {
			panicWithExpectedType(Uint64, result)
		}
		isMutated := true
		if usedStrategyMap == nil {
			isMutated = false
		}
		return uint64Result, MutationInfo{
			IsMutated:            isMutated,
			IsStringOrByteArray:  false,
			EntId:                0,
			UsedStrategyMap:      nil,
			UsedCodesForNotSbMap: usedStrategyMap,
		}
	default:
		panic(fmt.Sprintf("Type (%T) not implemented", v))
		return nil, MutationInfo{}
	}
}

// MutateRoot2String mutates a RootEnt structure and returns the resulting string,
// the ID of the mutated entity, and a map of the strategies used during mutation.
func MutateRoot2String(r *RootEnt, strength int) (string, int, map[MutationCode]int) {
	// Update score cache for selecting mutation target
	r.updateScoreCache()
	cachedScoreCums := r.ScoreCache()

	// Select an entity ID for mutation based on weighted probability
	selectedID := selectID(cachedScoreCums)
	// Extract surrounding string content while removing the selected entity
	startStr, endStr := SliceWithOutID(selectedID, r)
	targetPtr := r.Id2Ent[selectedID]
	// Perform mutation on the selected entity
	mutatedString, usedStrategy := MutateEnt2String(targetPtr, strength, r.Pattern)
	// Reconstruct the final mutated string
	finalString := startStr + mutatedString + endStr
	return finalString, selectedID, usedStrategy
}

// panicWithExpectedType generates a structured panic message based on expected and actual types.
func panicWithExpectedType(expectedType SupportedType, actual interface{}) {
	panic(fmt.Sprintf("Logical Error: Expected a %s, but got %T.", expectedType, actual))
}

// selectID selects an ID from the cumulative sum slice (cum) using a random value.
func selectID(cums []uint64) int {
	total := cums[len(cums)-1]
	r := OncePCG().Uint64n(total)
	idx := sort.Search(len(cums), func(i int) bool { return cums[i] > r })
	return idx
}

// cumAndStrategy represents a structure that stores the cumulative sum at each mutation strategy point.
type cumAndStrategy struct {
	cumSum   uint64
	strategy MutationCode
}

// MutateEnt2String mutates an entity (Ent) into a string using cached strategy weights.
// It selects mutation strategies probabilistically and applies them until the mutation
// strength limit is reached. If a mutation fails, the strategy is removed.
func MutateEnt2String(e Ent, strength int, pattern string) (string, map[MutationCode]int) {
	clone := cloneEnt(e)
	strategyCounts := make(map[MutationCode]int)

	cnt := 0
	e.updateStrategyCache()
	cachedCums := e.StrategyCache()
	localCumsAndStrategy := make([]cumAndStrategy, len(cachedCums))

	// Local copies for mutation
	for i := 0; i < len(cachedCums); i++ {
		localCumsAndStrategy[i] = cumAndStrategy{
			cumSum:   cachedCums[i],
			strategy: MutationCode(i),
		}
	}

	for {
		// If all available strategies are exhausted, exit the loop.
		if len(localCumsAndStrategy) == 0 {
			return SerializeToString(clone), nil
			break
		}
		strategy := selectStrategy(localCumsAndStrategy)
		selectedMutator := entMutators[strategy]

		isSuccess := selectedMutator(clone, pattern)
		if isSuccess {
			strategyCounts[strategy]++
			cnt++

		} else {

			// If a failure occurs, compute the contribution value (delta) of the strategy
			// and remove it from localKeys and localCum.
			delta := e.StrategyCount(strategy)
			localCumsAndStrategy = removeStrategy(localCumsAndStrategy, strategy, delta)
			continue
		}
		if cnt >= strength {
			break
		}
	}
	result := SerializeToString(clone)
	return result, strategyCounts
}

// selectStrategy selects a mutation strategy from the cumulative sum slice.
func selectStrategy(cumAndStrategy []cumAndStrategy) MutationCode {
	total := cumAndStrategy[len(cumAndStrategy)-1].cumSum
	rng := OncePCG()
	r := rng.Uint64n(total)

	idx := sort.Search(len(cumAndStrategy), func(i int) bool {
		return cumAndStrategy[i].cumSum > r
	})
	return cumAndStrategy[idx].strategy
}

// removeStrategy removes a strategy from localCumsAndStrategy and updates cumulative sums.
func removeStrategy(localCumsAndStrategy []cumAndStrategy, strategy MutationCode, delta uint64) []cumAndStrategy {
	// remove strategy
	index := -1
	for i, entry := range localCumsAndStrategy {
		if entry.strategy == strategy {
			index = i
			break
		}
	}
	if index == -1 {
		return localCumsAndStrategy
	}
	localCumsAndStrategy = append(localCumsAndStrategy[:index], localCumsAndStrategy[index+1:]...)

	// reflect removing to cumulative sum list
	for i := index; i < len(localCumsAndStrategy); i++ {
		localCumsAndStrategy[i].cumSum -= delta
	}

	return localCumsAndStrategy
}

// cumAndStrategyForNotSb represents a structure that stores the cumulative sum at each mutation strategy point.
type cumAndStrategyForNotSb struct {
	cumSum   uint64
	strategy MutationCodeForNotSB
}

// MutateNotSbEnt2interface mutates a NotSb entity using cached strategy probabilities.
// It applies strategies until the strength limit is reached, removing failed strategies.
func MutateNotSbEnt2interface(ens *EntForNotSB, strength int) (interface{}, map[MutationCodeForNotSB]int) {
	clone := cloneEntForNotSb(ens)
	strategyCounts := make(map[MutationCodeForNotSB]int)
	cnt := 0
	ens.updateStrategyCache()
	cachedCums := ens.StrategyCache()
	localCumsAndStrategy := make([]cumAndStrategyForNotSb, len(cachedCums))

	// Local copies for mutation
	for i := 0; i < len(cachedCums); i++ {
		localCumsAndStrategy[i] = cumAndStrategyForNotSb{
			cumSum:   cachedCums[i],
			strategy: MutationCodeForNotSB(i),
		}
	}

	for {
		if len(localCumsAndStrategy) == 0 {
			break
		}
		strategy := selectStrategyForNotSB(localCumsAndStrategy)
		selectedMutator := notSbEntMutators[strategy]
		isSuccess := selectedMutator(clone)
		if isSuccess {
			strategyCounts[strategy]++
			cnt++
		} else {
			delta := ens.StrategyCount(strategy)
			localCumsAndStrategy = removeStrategyForNotSB(localCumsAndStrategy, strategy, delta)
			continue
		}
		if cnt >= strength {
			break
		}
	}
	result := clone.data
	return result, strategyCounts
}

// selectStrategyForNotSB selects a NotSb mutation strategy using a random value.
func selectStrategyForNotSB(cumAndStrategy []cumAndStrategyForNotSb) MutationCodeForNotSB {
	total := cumAndStrategy[len(cumAndStrategy)-1].cumSum
	rng := OncePCG()
	r := rng.Uint64n(total)

	idx := sort.Search(len(cumAndStrategy), func(i int) bool {
		return cumAndStrategy[i].cumSum > r
	})
	return cumAndStrategy[idx].strategy
}

// removeStrategyForNotSB removes an element from a NotSb strategy slice at a given index.
func removeStrategyForNotSB(localCumsAndStrategy []cumAndStrategyForNotSb, strategy MutationCodeForNotSB, delta uint64) []cumAndStrategyForNotSb {
	// remove strategy
	index := -1
	for i, entry := range localCumsAndStrategy {
		if entry.strategy == strategy {
			index = i
			break
		}
	}
	if index == -1 {
		return localCumsAndStrategy
	}
	localCumsAndStrategy = append(localCumsAndStrategy[:index], localCumsAndStrategy[index+1:]...)

	// reflect removing to cumulative sum list
	for i := index; i < len(localCumsAndStrategy); i++ {
		localCumsAndStrategy[i].cumSum -= delta
	}

	return localCumsAndStrategy
}

// ReflectSeed updates the mutation sample using the given seed's mutation information.
func (ms *MutSample) ReflectSeed(seed Seed) {
	ms.PlusArgPowerSetScore(uint64(seed.ArgPowerSetIdx), 2)
	ms.reflectMutationsInfo(seed.MutationsInfo)
	ms.PlusSamplePopularity(1)
}

// PlusArgPowerSetScore increases the score of a given argument index in the power set.
func (ms *MutSample) PlusArgPowerSetScore(i uint64, delta uint64) {
	if int(i) >= len(ms.argPowerSetScores) {
		newSize := int(i) + 1
		newScores := make([]uint64, newSize)
		copy(newScores, ms.argPowerSetScores)
		ms.argPowerSetScores = newScores
	}
	ms.argPowerSetScores[i] += delta
	ms.needsUpdateCache = true
}

// reflectMutationsInfo updates mutation data for all mutated arguments in the sample.
func (ms *MutSample) reflectMutationsInfo(mutInfoDict map[int]MutationInfo) {
	for argIdx, mi := range mutInfoDict {
		if !mi.IsMutated {
			continue
		}
		if mi.IsStringOrByteArray {
			if root, exists := ms.rootEntsDict[argIdx]; exists {
				root.ReflectMutationInfo(mi)
			}
		} else {
			if ens, exists := ms.notSbEntsDict[argIdx]; exists {
				ens.ReflectMutationInfo(mi)
			}
		}
	}
}

// PlusSamplePopularity increases the sample's popularity count.
func (ms *MutSample) PlusSamplePopularity(delta int) {
	ms.samplePopularity += uint64(delta)
}

func (ms *MutSample) SamplePopularity() uint64 {
	return ms.samplePopularity
}

// UpdatePowerSetCache updates the cached cumulative sum array based on ArgPowerSetScores.
func (ms *MutSample) UpdatePowerSetCache() {
	if !ms.needsUpdateCache && ms.cachedPowerSetCumSum != nil {
		return
	}
	n := len(ms.argPowerSetScores)
	ms.cachedPowerSetCumSum = make([]uint64, n)
	var total uint64 = 0
	for i, score := range ms.argPowerSetScores {
		total += score
		ms.cachedPowerSetCumSum[i] = total
	}
	ms.needsUpdateCache = false
}

// PowerSetCache retrieves the cached cumulative sum array. If it is outdated, it updates it first.
func (ms *MutSample) PowerSetCache() []uint64 {
	if ms.needsUpdateCache {
		ms.UpdatePowerSetCache()
	}
	return ms.cachedPowerSetCumSum
}

// NewMutSample creates a new MutSample from the given Seed.
// If seed.Content has length L, the ArgPowerSetScores array will have a size of (1 << L) - 1, initialized to 1.
func NewMutSample(seed Seed) *MutSample {
	L := len(seed.Content)
	n := (int(1) << L) - 1 // Equivalent to 2^L - 1
	// Initialize ArgPowerSetScores with 1s
	scores := make([]uint64, n)
	for i := 0; i < n; i++ {
		scores[i] = 1
	}
	// Parse seed content into root and non-root entities
	formattedRootEntsDict, formattedNotSbEntsDict := formatToRootOrNotSbEnts(seed.Content)
	return &MutSample{
		seed:                 seed,
		argPowerSetScores:    scores,
		samplePopularity:     1,
		rootEntsDict:         formattedRootEntsDict,
		notSbEntsDict:        formattedNotSbEntsDict,
		cachedPowerSetCumSum: nil, // Cache starts empty
		needsUpdateCache:     true,
	}
}

// formatToRootOrNotSbEnts processes content by parsing string and byte elements into RootEnt entities
// and other types into EntForNotSB entities. The function returns two maps indexed by argument position.
func formatToRootOrNotSbEnts(content []interface{}) (map[int]*RootEnt, map[int]*EntForNotSB) {
	resultRoot := make(map[int]*RootEnt)
	resultNotSB := make(map[int]*EntForNotSB)
	for i, v := range content {
		switch val := v.(type) {
		case string:
			if rootPtr := ParseToRoot([]byte(val)); rootPtr != nil {
				resultRoot[i] = rootPtr
			}
		case []byte:
			if rootPtr := ParseToRoot(val); rootPtr != nil {
				resultRoot[i] = rootPtr
			}
		default:
			notSbPtr := ParseToEntForNotSB(val)
			resultNotSB[i] = notSbPtr
		}
	}
	return resultRoot, resultNotSB
}

// MutatorEntry represents an entry mapping a mutation code to its corresponding function.
type MutatorEntry struct {
	Code MutationCode // The mutation strategy code.
	Func MutateFunc   // The function implementing the mutation.
}

// MutationCode defines various mutation operations for mutateing entities.
type MutationCode int

// MutateFunc defines the type of functions that mutate Ent values.
type MutateFunc func(Ent, string) bool

// NotSbMutatorEntry represents an entry mapping a non-string/byte mutation code to its function.
type NotSbMutatorEntry struct {
	Code MutationCodeForNotSB // The mutation strategy code for non-string/byte entities.
	Func MutateFuncForNotSB   // The function implementing the mutation.
}

// MutationCodeForNotSB defines mutation strategies for non-string/byte entities.
type MutationCodeForNotSB int

// MutateFuncForNotSB defines the type of functions that mutate EntForNotSB (not string or []byte) values.
type MutateFuncForNotSB func(*EntForNotSB) bool

const (
	DS    MutationCode = iota // Delete Start
	DM                        // Delete Middle
	DE                        // Delete End
	AS                        // Append Start
	ASI                       // Append Start Interesting
	ASP                       // Append Start Pattern
	AE                        // Append End
	AEI                       // Append End Interesting
	AEP                       // Append End Pattern
	OS                        // Overwrite Start
	OM                        // Overwrite Middle
	CPS                       // Copy Start
	CPM                       // Copy Middle
	CPE                       // Copy End
	CHR                       // Change Randomly
	CHG                       // Change Gradual
	CHG16                     // Change Gradual 16
	CHG32                     // Change Gradual 32
	CHG64                     // Change Gradual 64
	CHI                       // Change Interesting
	CHI16                     // Change Interesting 16
	CHI32                     // Change Interesting 32
	CHP                       // Change Pattern
	IR                        // Insert Randomly
	IG                        // Insert Gradual
	II                        // Insert Interesting
	IP                        // Insert Pattern
	SW                        // Swap
	PKL                       // Preserve Kind Light
	PKH                       // Preserve Kind Hard
	SKL                       // Switch Kind Light
	SKH                       // Switch Kind Hard
	// MutationBoundaryMarker is an `iota` value used to track the number of registered mutation codes.
	//
	// **DO NOT USE THIS VALUE AS A MUTATION CODE!**
	// This is strictly for tracking the total number of mutation codes.
	MutationBoundaryMarker // tracking the number of mutation codes
)

// mutatorList stores a list of mutation strategy mappings.
var mutatorList = []MutatorEntry{
	{DS, deleteStart},
	{DM, deleteMiddle},
	{DE, deleteEnd},
	{AS, appendStart},
	{ASI, appendStartInteresting},
	{ASP, appendStartPattern},
	{AE, appendEnd},
	{AEI, appendEndInteresting},
	{AEP, appendEndPattern},
	{OS, overwriteStart},
	{OM, overwriteMiddle},
	{CPS, copyStart},
	{CPM, copyMiddle},
	{CPE, copyEnd},
	{CHR, changeRandomly},
	{CHG, changeGradual},
	{CHG16, changeGradual16},
	{CHG32, changeGradual32},
	{CHG64, changeGradual64},
	{CHI, changeInteresting},
	{CHI16, changeInteresting16},
	{CHI32, changeInteresting32},
	{CHP, changePattern},
	{IR, insertRandomly},
	{IG, insertGradual},
	{II, insertInteresting},
	{IP, insertPattern},
	{SW, swap},
	{PKL, preserveKindLight},
	{PKH, preserveKindHard},
	{SKL, switchKindLight},
	{SKH, switchKindHard},
}

const (
	SmlPls       MutationCodeForNotSB = iota // small plus
	SmlMns                                   // small minus
	SmlDv                                    // small division
	SmlMl                                    // small multiply
	MddlPls                                  // middle plus
	MddlMins                                 // middle minus
	MddlDv                                   // middle division
	MddlMl                                   // middcle multiply
	BgPls                                    // big plus
	BgMns                                    // big minus
	BgDv                                     // big division
	BgMl                                     // big multiply
	FctrlDv                                  // factorial division
	FctrlMl                                  // fcatorial multiply
	IntrstngChng                             // change to interesting value
	RndmChng                                 // chahge to radom value
	// NotSbBoundaryMarker is an `iota` value used to track the number of registered mutation codes.
	//
	// **DO NOT USE THIS VALUE AS A MUTATION CODE!**
	// This is strictly for tracking the total number of mutation codes.
	NotSbBoundaryMarker // tracking the number of notSbmutation codes
)

// notSbMutatorList stores a list of mutation strategy mappings for non-string/byte entities
var notSbMutatorList = []NotSbMutatorEntry{
	{SmlPls, smallPlus},
	{SmlMns, smallMinus},
	{SmlDv, smallDiv},
	{SmlMl, smallMul},
	{MddlPls, middlePlus},
	{MddlMins, middleMinus},
	{MddlDv, middleDiv},
	{MddlMl, middleMul},
	{BgPls, bigPlus},
	{BgMns, bigMinus},
	{BgDv, bigDiv},
	{BgMl, bigMul},
	{FctrlDv, factorialDiv},
	{FctrlMl, factorialMul},
	{IntrstngChng, interestingChange},
	{RndmChng, randomChange},
}

const (
	// NumMutationCodes defines the total number of mutation codes.
	// This constant is derived from `MutationBoundaryMarker` and is automatically updated
	// when new mutation codes are added.
	//
	// DO NOT modify this manually!
	NumMutationCodes = int(MutationBoundaryMarker)

	// NumNotSbMutCodes defines the total number of non-string/byte mutation codes.
	//
	// This constant is derived from `NotSbBoundaryMarker` and is automatically updated
	// when new non-string mutation codes are added.
	// DO NOT modify this manually!
	NumNotSbMutCodes = int(NotSbBoundaryMarker)
)

var (
	// entMutators maps mutation codes to their respective transformation functions.
	// NOTE: This map is initialized in the `init()` function
	entMutators map[MutationCode]MutateFunc
	// notSbEntMutators maps non-string/byte mutation codes to their respective transformation functions.
	// NOTE: This map is initialized in the `init()` function
	notSbEntMutators map[MutationCodeForNotSB]MutateFuncForNotSB
)

// init initializes mutation function mappings at runtime.
//
// It populates `entMutators` and `notSbEntMutators` based on the predefined
// mutation strategy lists (`mutatorList` and `notSbMutatorList`)
func init() {
	entMutators = make(map[MutationCode]MutateFunc, NumMutationCodes)
	for _, entry := range mutatorList {
		entMutators[entry.Code] = entry.Func
	}
	notSbEntMutators = make(map[MutationCodeForNotSB]MutateFuncForNotSB, NumNotSbMutCodes)
	for _, entry := range notSbMutatorList {
		notSbEntMutators[entry.Code] = entry.Func
	}
}

// TODO: Refine the mutation logic from this point onwards.
// The current implementation is a simple random generation and is still quite primitive.
// especially regarding the refinement of complex Ent types.
// Leverage randobyteFrom to handle value set issues when performing append and insert operations.
// Also, ensure that the function returns true in most cases to enable efficient fuzzing.
// It should only return false in unavoidable situations.
// The mutation refinement should also be performed based on the pattern value.
// The p parameter is intended for future use but is currently unused.

// processMutateStrategy applies a mutation strategy based on the type of the given Ent.
// depending on whether the entity is a `PrimitiveEnt`, `KeyValEnt`, `GroupEnt`, or `RootEnt`.
// Parameters:
//   - e: The entity (`Ent`) to be processed.
//   - primitiveFunc: A function that applies a mutation to `PrimitiveEnt` instances.
//   - keyValFunc: A function that applies a mutation to `KeyValEnt` instances.
//   - groupFunc: A function that applies a mutation to `GroupEnt` instances.
//   - rootFunc: A function that applies a mutation to `RootEnt` instances.
//
// Returns:
//   - `true` if a mutation function was successfully applied to the given entity.
//   - `false` if a mutation was failed or not implemented.
func processMutateStrategy(
	e Ent,
	primitiveFunc func(PrimitiveEnt) bool,
	keyValFunc func(*KeyValEnt) bool,
	groupFunc func(*GroupEnt) bool,
	rootFunc func(*RootEnt) bool,
) bool {
	switch ent := e.(type) {
	case PrimitiveEnt:
		return primitiveFunc(ent)
	case *KeyValEnt:
		return keyValFunc(ent)
	case *GroupEnt:
		return groupFunc(ent)
	case *RootEnt:
		return rootFunc(ent)
	default:
		return false
	}
}

// handleEmptyRoot ensures that a `RootEnt` is not empty.
// If empty, it inserts a random alphabetic character.
func handleEmptyRoot(root *RootEnt) bool {
	if len(root.data) == 0 {
		randomChar := byte(OncePCG().Intn(26) + 65) // 'A' ~ 'Z'
		if OncePCG().Intn(2) == 0 {
			randomChar = byte(OncePCG().Intn(26) + 97) // 'a' ~ 'z'
		}
		root.data = []Ent{&LetterEnt{data: []byte{randomChar}}}
	}
	return true
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
// deleteStart removes a portion of the beginning of the entity's data.
func deleteStart(e Ent, pattern string) bool {
	// pFunc applies a mutation to `PrimitiveEnt` instances by removing
	// a portion of their data from the start.
	pFunc := func(pe PrimitiveEnt) bool {
		// TODO: Fine-tuning is, of course, needed in the future.
		// Must add mutation guidelines for complex types.

		data := pe.PrimitiveData()
		if data == nil || len(data) == 0 {
			return false
		}
		maxToDelete := int(float64(len(data)) * 0.6)
		if maxToDelete <= 0 {
			return false
		}
		deleteCount := OncePCG().Intn(maxToDelete + 1)
		if deleteCount == 0 {
			return false
		}
		newdata := data[deleteCount:]
		pe.SetPrimitiveData(newdata)
		return true
	}

	// TODO: Add mutation guidelines for complex types.
	// Composite types are not handled yet.
	// kvFunc serves as a placeholder for `KeyValEnt` mutation logic.
	// Currently, it does nothing and always returns false.
	kvFunc := func(kv *KeyValEnt) bool {
		return false
	}
	// gFunc serves as a placeholder for `GroupEnt` mutation logic.
	// Currently, it does nothing and always returns false.
	gFunc := func(g *GroupEnt) bool {
		return false
	}
	// rFunc applies a mutation to `RootEnt` instances.
	rFunc := handleEmptyRoot
	return processMutateStrategy(e, pFunc, kvFunc, gFunc, rFunc)
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func deleteMiddle(e Ent, pattern string) bool {
	pFunc := func(pe PrimitiveEnt) bool {
		data := pe.PrimitiveData()
		if data == nil || len(data) < 3 {
			return true
		}
		maxToDelete := int(float64(len(data)) * 0.5)
		if maxToDelete <= 0 {
			return false
		}
		deleteCount := OncePCG().Intn(maxToDelete) + 1
		start := OncePCG().Intn(len(data)-deleteCount-1) + 1
		newdata := append(data[:start], data[start+deleteCount:]...)
		pe.SetPrimitiveData(newdata)
		return true
	}

	kvFunc := func(kv *KeyValEnt) bool {
		return false
	}

	gFunc := func(g *GroupEnt) bool {
		return false
	}

	rFunc := handleEmptyRoot
	return processMutateStrategy(e, pFunc, kvFunc, gFunc, rFunc)
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func deleteEnd(e Ent, pattern string) bool {
	pFunc := func(pe PrimitiveEnt) bool {
		data := pe.PrimitiveData()
		if data == nil || len(data) == 0 {
			return false
		}
		maxToDelete := int(float64(len(data)) * 0.5)
		if maxToDelete <= 0 {
			return false
		}
		deleteCount := OncePCG().Intn(maxToDelete) + 1
		newdata := data[:len(data)-deleteCount]
		pe.SetPrimitiveData(newdata)
		return true
	}
	kvFunc := func(kv *KeyValEnt) bool {
		return false
	}

	gFunc := func(g *GroupEnt) bool {
		return false
	}

	rFunc := handleEmptyRoot
	return processMutateStrategy(e, pFunc, kvFunc, gFunc, rFunc)
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func appendStart(e Ent, pattern string) bool {
	pFunc := func(pe PrimitiveEnt) bool {
		data := pe.PrimitiveData()

		randomByte := byte(OncePCG().Intn(256))
		newdata := append([]byte{randomByte}, data...)
		pe.SetPrimitiveData(newdata)
		return true
	}
	kvFunc := func(kv *KeyValEnt) bool {
		return false
	}

	gFunc := func(g *GroupEnt) bool {
		return false
	}

	rFunc := handleEmptyRoot
	return processMutateStrategy(e, pFunc, kvFunc, gFunc, rFunc)
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func appendStartInteresting(e Ent, pattern string) bool {
	return appendStart(e, pattern)
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func appendStartPattern(e Ent, pattern string) bool {
	return appendStart(e, pattern)
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func appendEnd(e Ent, pattern string) bool {
	pFunc := func(pe PrimitiveEnt) bool {
		data := pe.PrimitiveData()

		randomByte := byte(OncePCG().Intn(256))
		newdata := append(data, randomByte)
		pe.SetPrimitiveData(newdata)
		return true
	}
	kvFunc := func(kv *KeyValEnt) bool {
		return false
	}

	gFunc := func(g *GroupEnt) bool {
		return false
	}

	rFunc := handleEmptyRoot
	return processMutateStrategy(e, pFunc, kvFunc, gFunc, rFunc)
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func appendEndInteresting(e Ent, pattern string) bool {
	return appendEnd(e, pattern)
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func appendEndPattern(e Ent, pattern string) bool {
	return appendEnd(e, pattern)
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func overwriteStart(e Ent, pattern string) bool {
	pFunc := func(pe PrimitiveEnt) bool {
		data := pe.PrimitiveData()
		if data == nil || len(data) == 0 {
			return false
		}
		maxToOverwrite := int(float64(len(data)) * 0.3)
		if maxToOverwrite <= 0 {
			maxToOverwrite = 1
		}
		overwriteCount := OncePCG().Intn(maxToOverwrite) + 1
		for i := 0; i < overwriteCount && i < len(data); i++ {
			data[i] = byte(OncePCG().Intn(256))
		}
		return true
	}
	kvFunc := func(kv *KeyValEnt) bool {
		return false
	}

	gFunc := func(g *GroupEnt) bool {
		return false
	}

	rFunc := handleEmptyRoot
	return processMutateStrategy(e, pFunc, kvFunc, gFunc, rFunc)
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func overwriteMiddle(e Ent, pattern string) bool {
	pFunc := func(pe PrimitiveEnt) bool {
		data := pe.PrimitiveData()
		if data == nil || len(data) < 3 {
			return false
		}
		start := OncePCG().Intn(len(data)/2) + 1
		maxOverwrite := len(data) - start
		if maxOverwrite <= 0 {
			return false
		}
		overwriteCount := OncePCG().Intn(maxOverwrite)
		if overwriteCount == 0 {
			overwriteCount = 1
		}
		for i := start; i < start+overwriteCount && i < len(data); i++ {
			data[i] = byte(OncePCG().Intn(256))
		}
		return true
	}
	kvFunc := func(kv *KeyValEnt) bool {
		return false
	}

	gFunc := func(g *GroupEnt) bool {
		return false
	}

	rFunc := handleEmptyRoot
	return processMutateStrategy(e, pFunc, kvFunc, gFunc, rFunc)
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func copyStart(e Ent, pattern string) bool {
	pFunc := func(pe PrimitiveEnt) bool {
		data := pe.PrimitiveData()
		if data == nil || len(data) < 1 {
			return false
		}
		if len(data) < 1 {
			return false
		}
		copyLength := OncePCG().Intn(len(data)) + 1
		segment := data[:copyLength]
		newdata := append(segment, data...)
		pe.SetPrimitiveData(newdata)
		return true
	}
	kvFunc := func(kv *KeyValEnt) bool {
		return false
	}

	gFunc := func(g *GroupEnt) bool {
		return false
	}

	rFunc := handleEmptyRoot
	return processMutateStrategy(e, pFunc, kvFunc, gFunc, rFunc)
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func copyMiddle(e Ent, pattern string) bool {
	pFunc := func(pe PrimitiveEnt) bool {
		data := pe.PrimitiveData()
		if data == nil || len(data) < 3 {
			return false
		}
		start := OncePCG().Intn(len(data)/2) + 1
		maxCopy := len(data) - start
		if maxCopy <= 0 {
			return false
		}
		copyLength := OncePCG().Intn(maxCopy) + 1
		segment := data[start : start+copyLength]
		newdata := append(data[:start], append(segment, data[start:]...)...)
		pe.SetPrimitiveData(newdata)
		return true
	}
	kvFunc := func(kv *KeyValEnt) bool {
		return false
	}

	gFunc := func(g *GroupEnt) bool {
		return false
	}

	rFunc := handleEmptyRoot
	return processMutateStrategy(e, pFunc, kvFunc, gFunc, rFunc)
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func copyEnd(e Ent, pattern string) bool {
	pFunc := func(pe PrimitiveEnt) bool {
		data := pe.PrimitiveData()
		if data == nil || len(data) < 1 {
			return false
		}
		if len(data) == 0 {
			return false
		}
		copyLength := OncePCG().Intn(len(data)) + 1
		segment := data[len(data)-copyLength:]
		newdata := append(data, segment...)
		pe.SetPrimitiveData(newdata)
		return true
	}
	kvFunc := func(kv *KeyValEnt) bool {
		return false
	}

	gFunc := func(g *GroupEnt) bool {
		return false
	}

	rFunc := handleEmptyRoot
	return processMutateStrategy(e, pFunc, kvFunc, gFunc, rFunc)
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func changeRandomly(e Ent, pattern string) bool {
	pFunc := func(pe PrimitiveEnt) bool {
		data := pe.PrimitiveData()
		if data == nil || len(data) == 0 {
			return false
		}

		idx := OncePCG().Intn(len(data))
		data[idx] = byte(OncePCG().Intn(256))
		return true
	}
	kvFunc := func(kv *KeyValEnt) bool {
		return false
	}

	gFunc := func(g *GroupEnt) bool {
		return false
	}

	rFunc := handleEmptyRoot
	return processMutateStrategy(e, pFunc, kvFunc, gFunc, rFunc)
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func changeGradual(e Ent, pattern string) bool {
	pFunc := func(pe PrimitiveEnt) bool {
		data := pe.PrimitiveData()
		if data == nil || len(data) == 0 {
			return false
		}
		idx := OncePCG().Intn(len(data))
		if OncePCG().Intn(2) == 0 && data[idx] > 0 {
			data[idx]--
		} else {
			data[idx]++
		}
		return true
	}
	kvFunc := func(kv *KeyValEnt) bool {
		return false
	}

	gFunc := func(g *GroupEnt) bool {
		return false
	}

	rFunc := handleEmptyRoot
	return processMutateStrategy(e, pFunc, kvFunc, gFunc, rFunc)
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func changeGradual16(e Ent, pattern string) bool {
	pFunc := func(pe PrimitiveEnt) bool {
		data := pe.PrimitiveData()
		if data == nil || len(data) < 2 {
			return false
		}
		idx := OncePCG().Intn(len(data) - 1)
		for i := 0; i < 2; i++ {
			if OncePCG().Intn(2) == 0 && data[idx+i] > 0 {
				data[idx+i]--
			} else {
				data[idx+i]++
			}
		}
		return true
	}
	kvFunc := func(kv *KeyValEnt) bool {
		return false
	}

	gFunc := func(g *GroupEnt) bool {
		return false
	}

	rFunc := handleEmptyRoot
	return processMutateStrategy(e, pFunc, kvFunc, gFunc, rFunc)
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func changeGradual32(e Ent, pattern string) bool {
	pFunc := func(pe PrimitiveEnt) bool {
		data := pe.PrimitiveData()
		if data == nil || len(data) < 4 {
			return false
		}
		idx := OncePCG().Intn(len(data) - 3)
		for i := 0; i < 4; i++ {
			if OncePCG().Intn(2) == 0 && data[idx+i] > 0 {
				data[idx+i]--
			} else {
				data[idx+i]++
			}
		}
		return true
	}
	kvFunc := func(kv *KeyValEnt) bool {
		return false
	}

	gFunc := func(g *GroupEnt) bool {
		return false
	}

	rFunc := handleEmptyRoot
	return processMutateStrategy(e, pFunc, kvFunc, gFunc, rFunc)
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func changeGradual64(e Ent, pattern string) bool {
	pFunc := func(pe PrimitiveEnt) bool {
		data := pe.PrimitiveData()
		if data == nil || len(data) < 8 {
			return false
		}
		idx := OncePCG().Intn(len(data) - 7)
		for i := 0; i < 8; i++ {
			if OncePCG().Intn(2) == 0 && data[idx+i] > 0 {
				data[idx+i]--
			} else {
				data[idx+i]++
			}
		}
		return true
	}
	kvFunc := func(kv *KeyValEnt) bool {
		return false
	}

	gFunc := func(g *GroupEnt) bool {
		return false
	}

	rFunc := handleEmptyRoot
	return processMutateStrategy(e, pFunc, kvFunc, gFunc, rFunc)
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func changeInteresting(e Ent, pattern string) bool {
	interestingValues := []byte{0, 1, 0x7F, 0xFF}
	pFunc := func(pe PrimitiveEnt) bool {
		data := pe.PrimitiveData()
		if data == nil || len(data) == 0 {
			return false
		}
		idx := OncePCG().Intn(len(data))

		data[idx] = interestingValues[OncePCG().Intn(len(interestingValues))]
		return true
	}
	kvFunc := func(kv *KeyValEnt) bool {
		return false
	}

	gFunc := func(g *GroupEnt) bool {
		return false
	}

	rFunc := handleEmptyRoot
	return processMutateStrategy(e, pFunc, kvFunc, gFunc, rFunc)
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func changeInteresting16(e Ent, pattern string) bool {
	interesting16 := [][]byte{{0, 0}, {0, 1}, {0x7F, 0xFF}, {0xFF, 0xFF}}

	pFunc := func(pe PrimitiveEnt) bool {
		data := pe.PrimitiveData()
		if data == nil || len(data) < 2 {
			return false
		}
		idx := OncePCG().Intn(len(data) - 1)
		pair := interesting16[OncePCG().Intn(len(interesting16))]
		data[idx] = pair[0]
		data[idx+1] = pair[1]
		return true
	}
	kvFunc := func(kv *KeyValEnt) bool {
		return false
	}

	gFunc := func(g *GroupEnt) bool {
		return false
	}

	rFunc := handleEmptyRoot
	return processMutateStrategy(e, pFunc, kvFunc, gFunc, rFunc)
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func changeInteresting32(e Ent, pattern string) bool {
	interesting32 := [][]byte{
		{0, 0, 0, 0},
		{0, 0, 0, 1},
		{0x7F, 0xFF, 0xFF, 0xFF},
		{0xFF, 0xFF, 0xFF, 0xFF},
	}
	pFunc := func(pe PrimitiveEnt) bool {
		data := pe.PrimitiveData()
		if data == nil || len(data) < 4 {
			return false
		}
		idx := OncePCG().Intn(len(data) - 3)
		pattern := interesting32[OncePCG().Intn(len(interesting32))]
		for i := 0; i < 4; i++ {
			data[idx+i] = pattern[i]
		}
		return true
	}
	kvFunc := func(kv *KeyValEnt) bool {
		return false
	}

	gFunc := func(g *GroupEnt) bool {
		return false
	}

	rFunc := handleEmptyRoot
	return processMutateStrategy(e, pFunc, kvFunc, gFunc, rFunc)
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func changePattern(e Ent, pattern string) bool {
	p := byte(0xAA)

	pFunc := func(pe PrimitiveEnt) bool {
		data := pe.PrimitiveData()
		if data == nil || len(data) == 0 {
			return false
		}
		idx := OncePCG().Intn(len(data))
		data[idx] = p
		return true
	}
	kvFunc := func(kv *KeyValEnt) bool {
		return false
	}

	gFunc := func(g *GroupEnt) bool {
		return false
	}

	rFunc := handleEmptyRoot
	return processMutateStrategy(e, pFunc, kvFunc, gFunc, rFunc)
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func insertRandomly(e Ent, pattern string) bool {
	pFunc := func(pe PrimitiveEnt) bool {
		data := pe.PrimitiveData()
		randomByte := byte(OncePCG().Intn(256))
		pos := 0
		if len(data) > 0 {
			pos = OncePCG().Intn(len(data))
		}
		newdata := append(data[:pos], append([]byte{randomByte}, data[pos:]...)...)
		pe.SetPrimitiveData(newdata)
		return true
	}
	kvFunc := func(kv *KeyValEnt) bool {
		return false
	}

	gFunc := func(g *GroupEnt) bool {
		return false
	}

	rFunc := handleEmptyRoot
	return processMutateStrategy(e, pFunc, kvFunc, gFunc, rFunc)
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func insertGradual(e Ent, pattern string) bool {
	return insertRandomly(e, pattern)
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func insertInteresting(e Ent, pattern string) bool {
	interestingValues := []byte{0, 1, 0x7F, 0xFF}
	pFunc := func(pe PrimitiveEnt) bool {
		data := pe.PrimitiveData()
		value := interestingValues[OncePCG().Intn(len(interestingValues))]
		pos := 0
		if len(data) > 0 {
			pos = OncePCG().Intn(len(data))
		}
		newdata := append(data[:pos], append([]byte{value}, data[pos:]...)...)
		pe.SetPrimitiveData(newdata)
		return true
	}
	kvFunc := func(kv *KeyValEnt) bool {
		return false
	}
	gFunc := func(g *GroupEnt) bool {
		return false
	}
	rFunc := handleEmptyRoot
	return processMutateStrategy(e, pFunc, kvFunc, gFunc, rFunc)
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func insertPattern(e Ent, pattern string) bool {
	return insertRandomly(e, pattern)
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func swap(e Ent, pattern string) bool {
	pFunc := func(pe PrimitiveEnt) bool {
		data := pe.PrimitiveData()
		if data == nil || len(data) < 2 {
			return false
		}
		idx1 := OncePCG().Intn(len(data))
		idx2 := OncePCG().Intn(len(data))
		data[idx1], data[idx2] = data[idx2], data[idx1]
		return true
	}
	kvFunc := func(kv *KeyValEnt) bool {
		return false
	}

	gFunc := func(g *GroupEnt) bool {
		return false
	}

	rFunc := handleEmptyRoot
	return processMutateStrategy(e, pFunc, kvFunc, gFunc, rFunc)
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func preserveKindLight(e Ent, pattern string) bool {
	return false
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func preserveKindHard(e Ent, pattern string) bool {
	return false
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func switchKindLight(e Ent, pattern string) bool {
	return false
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func switchKindHard(e Ent, pattern string) bool {
	return false
}

///
// From this point onward are mutation functions for EntForNotSB.
///

// TODO: Fine-tuning is, of course, needed in the future.
func mutateBool(ens *EntForNotSB) bool {
	b, ok := ens.data.(bool)
	if !ok {
		return false
	}
	ens.data = !b
	return true
}

// TODO: Fine-tuning is, of course, needed in the future.
func smallPlus(ens *EntForNotSB) bool {
	if ens.isBool {
		return mutateBool(ens)
	}
	switch v := ens.data.(type) {
	case int:
		ens.data = v + int(1)
	case int8:
		ens.data = v + int8(1)
	case int16:
		ens.data = v + int16(1)
	case int32:
		ens.data = v + int32(1)
	case int64:
		ens.data = v + int64(1)
	case uint:
		ens.data = v + uint(1)
	case uint8:
		ens.data = v + uint8(1)
	case uint16:
		ens.data = v + uint16(1)
	case uint32:
		ens.data = v + uint32(1)
	case uint64:
		ens.data = v + uint64(1)
	case float32:
		ens.data = v + float32(1.0)
	case float64:
		ens.data = v + float64(1.0)
	default:
		return false
	}
	return true
}

// TODO: Fine-tuning is, of course, needed in the future.
func smallMinus(ens *EntForNotSB) bool {
	if ens.isBool {
		return mutateBool(ens)
	}
	switch v := ens.data.(type) {
	case int:
		ens.data = v - int(1)
	case int8:
		ens.data = v - int8(11)
	case int16:
		ens.data = v - int16(111)
	case int32:
		ens.data = v - int32(1111)
	case int64:
		ens.data = v - int64(11111)
	case uint:
		if v == 0 {
			ens.data = uint(0)
		} else {
			ens.data = v - uint(1)
		}
	case uint8:
		if v == 0 {
			ens.data = uint8(0)
		} else {
			ens.data = v - uint8(1)
		}
	case uint16:
		if v == 0 {
			ens.data = uint16(0)
		} else {
			ens.data = v - uint16(1)
		}
	case uint32:
		if v == 0 {
			ens.data = uint32(0)
		} else {
			ens.data = v - uint32(1)
		}
	case uint64:
		if v == 0 {
			ens.data = uint64(0)
		} else {
			ens.data = v - uint64(1)
		}
	case float32:
		ens.data = v - float32(1.0)
	case float64:
		ens.data = v - float64(1.0)
	default:
		return false
	}
	return true
}

// TODO: Fine-tuning is, of course, needed in the future.
func smallDiv(ens *EntForNotSB) bool {
	if ens.isBool {
		return mutateBool(ens)
	}
	switch v := ens.data.(type) {
	case int:
		if v == 0 {
			return false
		}
		ens.data = v / int(2)
	case int8:
		if v == 0 {
			return false
		}
		ens.data = v / int8(2)
	case int16:
		if v == 0 {
			return false
		}
		ens.data = v / int16(2)
	case int32:
		if v == 0 {
			return false
		}
		ens.data = v / int32(2)
	case int64:
		if v == 0 {
			return false
		}
		ens.data = v / int64(2)
	case uint:
		if v == 0 {
			return false
		}
		ens.data = v / uint(2)
	case uint8:
		if v == 0 {
			return false
		}
		ens.data = v / uint8(2)
	case uint16:
		if v == 0 {
			return false
		}
		ens.data = v / uint16(2)
	case uint32:
		if v == 0 {
			return false
		}
		ens.data = v / uint32(2)
	case uint64:
		if v == 0 {
			return false
		}
		ens.data = v / uint64(2)
	case float32:
		if v == 0.0 {
			return false
		}
		ens.data = v / float32(2.0)
	case float64:
		if v == 0.0 {
			return false
		}
		ens.data = v / float64(2.0)
	default:
		return false
	}
	return true
}

// TODO: Fine-tuning is, of course, needed in the future.
func smallMul(ens *EntForNotSB) bool {
	if ens.isBool {
		return mutateBool(ens)
	}
	switch v := ens.data.(type) {
	case int:
		ens.data = v * int(2)
	case int8:
		ens.data = v * int8(2)
	case int16:
		ens.data = v * int16(2)
	case int32:
		ens.data = v * int32(2)
	case int64:
		ens.data = v * int64(2)
	case uint:
		ens.data = v * uint(2)
	case uint8:
		ens.data = v * uint8(2)
	case uint16:
		ens.data = v * uint16(2)
	case uint32:
		ens.data = v * uint32(2)
	case uint64:
		ens.data = v * uint64(2)
	case float32:
		ens.data = v * float32(2.0)
	case float64:
		ens.data = v * float64(2.0)
	default:
		return false
	}
	return true
}

// TODO: Fine-tuning is, of course, needed in the future.
func middlePlus(ens *EntForNotSB) bool {
	if ens.isBool {
		return mutateBool(ens)
	}
	switch v := ens.data.(type) {
	case int:
		ens.data = v + int(10)
	case int8:
		ens.data = v + int8(10)
	case int16:
		ens.data = v + int16(10)
	case int32:
		ens.data = v + int32(10)
	case int64:
		ens.data = v + int64(10)
	case uint:
		ens.data = v + uint(10)
	case uint8:
		ens.data = v + uint8(10)
	case uint16:
		ens.data = v + uint16(10)
	case uint32:
		ens.data = v + uint32(10)
	case uint64:
		ens.data = v + uint64(10)
	case float32:
		ens.data = v + float32(10.0)
	case float64:
		ens.data = v + float64(10.0)
	default:
		return false
	}
	return true
}

// TODO: Fine-tuning is, of course, needed in the future.
func middleMinus(ens *EntForNotSB) bool {
	if ens.isBool {
		return mutateBool(ens)
	}
	switch v := ens.data.(type) {
	case int:
		ens.data = v - int(10)
	case int8:
		ens.data = v - int8(10)
	case int16:
		ens.data = v - int16(10)
	case int32:
		ens.data = v - int32(10)
	case int64:
		ens.data = v - int64(10)
	case uint:
		if v < uint(10) {
			ens.data = uint(0)
		} else {
			ens.data = v - uint(10)
		}
	case uint8:
		if v < uint8(10) {
			ens.data = uint8(0)
		} else {
			ens.data = v - uint8(10)
		}
	case uint16:
		if v < uint16(10) {
			ens.data = uint16(0)
		} else {
			ens.data = v - uint16(10)
		}
	case uint32:
		if v < uint32(10) {
			ens.data = uint32(0)
		} else {
			ens.data = v - uint32(10)
		}
	case uint64:
		if v < uint64(10) {
			ens.data = uint64(0)
		} else {
			ens.data = v - uint64(10)
		}
	case float32:
		ens.data = v - float32(10.0)
	case float64:
		ens.data = v - float64(10.0)
	default:
		return false
	}
	return true
}

// TODO: Fine-tuning is, of course, needed in the future.
func middleDiv(ens *EntForNotSB) bool {
	if ens.isBool {
		return mutateBool(ens)
	}
	switch v := ens.data.(type) {
	case int:
		if v == 0 {
			return false
		}
		ens.data = v / int(2)
	case int8:
		if v == 0 {
			return false
		}
		ens.data = v / int8(2)
	case int16:
		if v == 0 {
			return false
		}
		ens.data = v / int16(2)
	case int32:
		if v == 0 {
			return false
		}
		ens.data = v / int32(2)
	case int64:
		if v == 0 {
			return false
		}
		ens.data = v / int64(2)
	case uint:
		if v == 0 {
			return false
		}
		ens.data = v / uint(2)
	case uint8:
		if v == 0 {
			return false
		}
		ens.data = v / uint8(2)
	case uint16:
		if v == 0 {
			return false
		}
		ens.data = v / uint16(2)
	case uint32:
		if v == 0 {
			return false
		}
		ens.data = v / uint32(2)
	case uint64:
		if v == 0 {
			return false
		}
		ens.data = v / uint64(2)
	case float32:
		if v == 0.0 {
			return false
		}
		ens.data = v / float32(2.0)
	case float64:
		if v == 0.0 {
			return false
		}
		ens.data = v / float64(2.0)
	default:
		return false
	}
	return true
}

// TODO: Fine-tuning is, of course, needed in the future.
func middleMul(ens *EntForNotSB) bool {
	if ens.isBool {
		return mutateBool(ens)
	}
	switch v := ens.data.(type) {
	case int:
		ens.data = v * int(2)
	case int8:
		ens.data = v * int8(2)
	case int16:
		ens.data = v * int16(2)
	case int32:
		ens.data = v * int32(2)
	case int64:
		ens.data = v * int64(2)
	case uint:
		ens.data = v * uint(2)
	case uint8:
		ens.data = v * uint8(2)
	case uint16:
		ens.data = v * uint16(2)
	case uint32:
		ens.data = v * uint32(2)
	case uint64:
		ens.data = v * uint64(2)
	case float32:
		ens.data = v * float32(2.0)
	case float64:
		ens.data = v * float64(2.0)
	default:
		return false
	}
	return true
}

// TODO: Fine-tuning is, of course, needed in the future.
func bigPlus(ens *EntForNotSB) bool {
	if ens.isBool {
		return mutateBool(ens)
	}
	switch v := ens.data.(type) {
	case int:
		ens.data = v + int(1000)
	case int8:
		ens.data = v + int8(100)
	case int16:
		ens.data = v + int16(1000)
	case int32:
		ens.data = v + int32(1000)
	case int64:
		ens.data = v + int64(1000)
	case uint:
		ens.data = v + uint(1000)
	case uint8:
		ens.data = v + uint8(100)
	case uint16:
		ens.data = v + uint16(1000)
	case uint32:
		ens.data = v + uint32(1000)
	case uint64:
		ens.data = v + uint64(1000)
	case float32:
		ens.data = v + float32(1000.0)
	case float64:
		ens.data = v + float64(1000.0)
	default:
		return false
	}
	return true
}

// TODO: Fine-tuning is, of course, needed in the future.
func bigMinus(ens *EntForNotSB) bool {
	if ens.isBool {
		return mutateBool(ens)
	}
	switch v := ens.data.(type) {
	case int:
		ens.data = v - int(1000)
	case int8:
		ens.data = v - int8(100)
	case int16:
		ens.data = v - int16(1000)
	case int32:
		ens.data = v - int32(1000)
	case int64:
		ens.data = v - int64(1000)
	case uint:
		if v < uint(1000) {
			ens.data = uint(0)
		} else {
			ens.data = v - uint(1000)
		}
	case uint8:
		if v < uint8(100) {
			ens.data = uint8(0)
		} else {
			ens.data = v - uint8(100)
		}
	case uint16:
		if v < uint16(1000) {
			ens.data = uint16(0)
		} else {
			ens.data = v - uint16(1000)
		}
	case uint32:
		if v < uint32(1000) {
			ens.data = uint32(0)
		} else {
			ens.data = v - uint32(1000)
		}
	case uint64:
		if v < uint64(1000) {
			ens.data = uint64(0)
		} else {
			ens.data = v - uint64(1000)
		}
	case float32:
		ens.data = v - float32(1000.0)
	case float64:
		ens.data = v - float64(1000.0)
	default:
		return false
	}
	return true
}

// TODO: Fine-tuning is, of course, needed in the future.
func bigDiv(ens *EntForNotSB) bool {
	if ens.isBool {
		return mutateBool(ens)
	}
	switch v := ens.data.(type) {
	case int:
		if v == 0 {
			return false
		}
		ens.data = v / int(2)
	case int8:
		if v == 0 {
			return false
		}
		ens.data = v / int8(2)
	case int16:
		if v == 0 {
			return false
		}
		ens.data = v / int16(2)
	case int32:
		if v == 0 {
			return false
		}
		ens.data = v / int32(2)
	case int64:
		if v == 0 {
			return false
		}
		ens.data = v / int64(2)
	case uint:
		if v == 0 {
			return false
		}
		ens.data = v / uint(2)
	case uint8:
		if v == 0 {
			return false
		}
		ens.data = v / uint8(2)
	case uint16:
		if v == 0 {
			return false
		}
		ens.data = v / uint16(2)
	case uint32:
		if v == 0 {
			return false
		}
		ens.data = v / uint32(2)
	case uint64:
		if v == 0 {
			return false
		}
		ens.data = v / uint64(2)
	case float32:
		if v == 0.0 {
			return false
		}
		ens.data = v / float32(2.0)
	case float64:
		if v == 0.0 {
			return false
		}
		ens.data = v / float64(2.0)
	default:
		return false
	}
	return true
}

// TODO: Fine-tuning is, of course, needed in the future.
func bigMul(ens *EntForNotSB) bool {
	if ens.isBool {
		return mutateBool(ens)
	}
	switch v := ens.data.(type) {
	case int:
		ens.data = v * int(2)
	case int8:
		ens.data = v * int8(2)
	case int16:
		ens.data = v * int16(2)
	case int32:
		ens.data = v * int32(2)
	case int64:
		ens.data = v * int64(2)
	case uint:
		ens.data = v * uint(2)
	case uint8:
		ens.data = v * uint8(2)
	case uint16:
		ens.data = v * uint16(2)
	case uint32:
		ens.data = v * uint32(2)
	case uint64:
		ens.data = v * uint64(2)
	case float32:
		ens.data = v * float32(2.0)
	case float64:
		ens.data = v * float64(2.0)
	default:
		return false
	}
	return true
}

// TODO: Fine-tuning is, of course, needed in the future.
func factorialMul(ens *EntForNotSB) bool {
	if ens.isBool {
		return mutateBool(ens)
	}
	switch v := ens.data.(type) {
	case int:
		ens.data = v * int(3)
	case int8:
		ens.data = v * int8(3)
	case int16:
		ens.data = v * int16(3)
	case int32:
		ens.data = v * int32(3)
	case int64:
		ens.data = v * int64(3)
	case uint:
		ens.data = v * uint(3)
	case uint8:
		ens.data = v * uint8(3)
	case uint16:
		ens.data = v * uint16(3)
	case uint32:
		ens.data = v * uint32(3)
	case uint64:
		ens.data = v * uint64(3)
	case float32:
		ens.data = v * float32(3.0)
	case float64:
		ens.data = v * float64(3.0)
	default:
		return false
	}
	return true
}

// TODO: Fine-tuning is, of course, needed in the future.
func factorialDiv(ens *EntForNotSB) bool {
	if ens.isBool {
		return mutateBool(ens)
	}
	switch v := ens.data.(type) {
	case int:
		if v == 0 {
			return false
		}
		ens.data = v / int(3)
	case int8:
		if v == 0 {
			return false
		}
		ens.data = v / int8(3)
	case int16:
		if v == 0 {
			return false
		}
		ens.data = v / int16(3)
	case int32:
		if v == 0 {
			return false
		}
		ens.data = v / int32(3)
	case int64:
		if v == 0 {
			return false
		}
		ens.data = v / int64(3)
	case uint:
		if v == 0 {
			return false
		}
		ens.data = v / uint(3)
	case uint8:
		if v == 0 {
			return false
		}
		ens.data = v / uint8(3)
	case uint16:
		if v == 0 {
			return false
		}
		ens.data = v / uint16(3)
	case uint32:
		if v == 0 {
			return false
		}
		ens.data = v / uint32(3)
	case uint64:
		if v == 0 {
			return false
		}
		ens.data = v / uint64(3)
	case float32:
		if v == 0.0 {
			return false
		}
		ens.data = v / float32(3.0)
	case float64:
		if v == 0.0 {
			return false
		}
		ens.data = v / float64(3.0)
	default:
		return false
	}
	return true
}

// TODO: Fine-tuning is, of course, needed in the future.
func interestingChange(ens *EntForNotSB) bool {
	if ens.isBool {
		return mutateBool(ens)
	}
	switch ens.data.(type) {
	case int:
		ens.data = int(42)
	case int8:
		ens.data = int8(42)
	case int16:
		ens.data = int16(42)
	case int32:
		ens.data = int32(42)
	case int64:
		ens.data = int64(42)
	case uint:
		ens.data = uint(42)
	case uint8:
		ens.data = uint8(42)
	case uint16:
		ens.data = uint16(42)
	case uint32:
		ens.data = uint32(42)
	case uint64:
		ens.data = uint64(42)
	case float32:
		ens.data = float32(42)
	case float64:
		ens.data = float64(42)
	default:
		return false
	}
	return true
}

// TODO: Fine-tuning is, of course, needed in the future.
func randomChange(ens *EntForNotSB) bool {
	if ens.isBool {
		return mutateBool(ens)
	}
	r := OncePCG().Intn(101)
	switch ens.data.(type) {
	case int:
		ens.data = int(r)
	case int8:
		ens.data = int8(r)
	case int16:
		ens.data = int16(r)
	case int32:
		ens.data = int32(r)
	case int64:
		ens.data = int64(r)
	case uint:
		ens.data = uint(r)
	case uint8:
		ens.data = uint8(r)
	case uint16:
		ens.data = uint16(r)
	case uint32:
		ens.data = uint32(r)
	case uint64:
		ens.data = uint64(r)
	case float32:
		ens.data = float32(r)
	case float64:
		ens.data = float64(r)
	default:
		return false
	}
	return true
}

// Various helper functions.
// randomXXXFrom gradually mutates XXX and return XXX
func randomByteFrom(seedByte byte) byte {
	p := OncePCG().Float64() // 0.0 <= p < 1.0

	var b uint8

	currentCase := determineCase(rune(seedByte))

	isProgress := GenerateRandomBool(0.5)
	if currentCase == 2 {
		switch {
		case p < 0.45:
			currentCase = (currentCase + 3) % 4
			break
		case p < 0.55:
			break

		case p <= 1.0:
			currentCase = (currentCase + 1) % 4
		}
	} else {
		switch {
		case p < 0.1:
			currentCase = (currentCase + 2) % 4
			isProgress = false
			break
		case p >= 0.1 && p < 0.25:
			currentCase = (currentCase + 3) % 4
			isProgress = false
			break
		case p >= 0.25 && p < 0.75:
			break

		case p >= 0.75 && p < 0.90:
			isProgress = true
			currentCase = (currentCase + 1) % 4
		case p >= 0.90 && p <= 1.0:
			isProgress = true
			currentCase = (currentCase + 2) % 4
		}
	}

	if currentCase == 2 {
		if isProgress {
			if UniformRandomBool(0.8) {
				currentCase = (currentCase + 1) % 4
			}
		} else {
			if UniformRandomBool(0.8) {
				currentCase = (currentCase + 3) % 4
			}
		}
	}

	switch currentCase {
	case 0:
		b = uint8(RandRange(AbsoluteMin, SpecialMAX+1))
	case 1:
		b = uint8(RandRange(LowAsciiMIN, LowAsciiMAX+1))
	case 2:

		b = Del
	case 3:
		b = uint8(RandRange(HighAsciiMin, HighAscillMax+1))

	}

	return byte(b)
}

const (
	AbsoluteMin = 0x00
	SpecialMAX  = 0x1F

	LowAsciiMIN = 0x20
	LowAsciiMAX = 0x7E
	Del         = 0x7F

	HighAsciiMin  = 0x80
	HighAscillMax = 0xFF
)

func determineCase(seedRune rune) int {
	switch {
	case seedRune >= AbsoluteMin && seedRune <= SpecialMAX:
		return 0
	case seedRune >= LowAsciiMIN && seedRune <= LowAsciiMAX:
		return 1
	case seedRune == Del:
		return 2
	case seedRune >= HighAsciiMin && seedRune <= HighAscillMax:
		return 3
	default:
		panic("default case must not be caught")
	}
}

// randomXXXFrom gradually mutates XXX and return XXX
func randomIntFrom(i interface{}) interface{} {
	p := OncePCG().Float64()

	switch v := i.(type) {
	case int:
		var i interface{}
		if v == 0 {
			return int(RandInt64())
		}
		switch {
		case p < 0.15:
			min := int64(v) * (-2)
			max := int64(v) * 2
			if min > max {
				min, max = max, min
			}
			i = int(RandRange(min, max))
		case p < 0.3:
			min := int64(v) * (-4)
			max := int64(v) * (4)
			if min > max {
				min, max = max, min
			}
			i = int(RandRange(min, max))
		case p < 0.45:
			min := int64(v) * (-8)
			max := int64(v) * (8)
			if min > max {
				min, max = max, min
			}
			i = int(RandRange(min, max))
		case p < 0.60:
			min := int64(v) * (-16)
			max := int64(v) * (16)
			if min > max {
				min, max = max, min
			}
			i = int(RandRange(min, max))
		default:
			i = int(OncePCG().Int())
		}
		return i

	case int8:
		if v == 0 {
			return int8(RandInt64())
		}
		var i8 interface{}
		switch {
		case p < 0.3:
			min := int64(v) * (-2)
			max := int64(v) * (2)
			if min > max {
				min, max = max, min
			}
			i8 = int8(RandRange(min, max))
		case p < 0.5:
			min := int64(v) * (-4)
			max := int64(v) * (4)
			if min > max {
				min, max = max, min
			}
			i8 = int8(RandRange(min, max))
		default:
			i8 = int8(RandRange(-128, 128))
		}
		return i8

	case int16:
		if v == 0 {
			return int16(RandInt64())
		}
		var i16 interface{}
		switch {
		case p < 0.3:
			min := int64(v) * (-2)
			max := int64(v) * (2)
			if min > max {
				min, max = max, min
			}
			i16 = int16(RandRange(min, max))
		case p < 0.5:
			min := int64(v) * (-4)
			max := int64(v) * (4)
			if min > max {
				min, max = max, min
			}
			i16 = int16(RandRange(min, max))
		default:
			i16 = int16(RandRange(-32768, 32768))
		}
		return i16

	case int32:
		if v == 0 {
			return int32(RandInt64())
		}
		var i16 interface{}
		switch {
		case p < 0.2:
			min := int64(v) * (-2)
			max := int64(v) * (2)
			if min > max {
				min, max = max, min
			}
			i16 = int32(RandRange(min, max))
		case p < 0.4:
			min := int64(v) * (-4)
			max := int64(v) * (4)
			if min > max {
				min, max = max, min
			}
			i16 = int32(RandRange(min, max))
		case p < 0.6:
			min := int64(v) * (-8)
			max := int64(v) * (8)
			if min > max {
				min, max = max, min
			}
			i16 = int32(RandRange(min, max))
		default:
			i16 = OncePCG().Int32
		}
		return i16

	case int64:
		if v == 0 {
			return RandInt64()
		}
		var i64 interface{}
		switch {
		case p < 0.15:
			min := v * (-2)
			max := v * (2)
			if min > max {
				min, max = max, min
			}
			i64 = RandRange(min, max)
		case p < 0.3:
			min := v * (-4)
			max := v * (4)
			if min > max {
				min, max = max, min
			}
			i64 = RandRange(min, max)
		case p < 0.45:
			min := v * (-8)
			max := v * (8)
			if min > max {
				min, max = max, min
			}
			i64 = RandRange(min, max)
		case p < 0.60:
			min := v * (-16)
			max := v * (16)
			if min > max {
				min, max = max, min
			}
			i64 = RandRange(min, max)
		default:
			i64 = OncePCG().Int64()
		}
		return i64

	default:
		panic("it's not supported int type")
	}
}

// randomXXXFrom gradually mutates XXX and return XXX
func randomUintFrom(u interface{}) interface{} {
	p := OncePCG().Float64()

	switch v := u.(type) {
	case uint:
		if v == 0 {
			return uint(RandUint64())
		}
		var u interface{}
		switch {
		case p < 0.3:
			min := v / 256
			max := v * 2
			u = uint(UintRandRange(uint64(min), uint64(max)))
		case p < 0.6:
			min := v / 9096
			max := v * 8
			u = uint(UintRandRange(uint64(min), uint64(max)))
		default:
			u = uint(OncePCG().Uint64())
		}
		return u

	case uint8:
		if v == 0 {
			return uint8(RandUint64())
		}
		var u8 interface{}
		switch {
		case p < 0.3:
			min := int64(v) / 8
			max := int64(v) * 2
			u8 = uint8(RandRange(min, max))
		case p < 0.6:
			min := int64(v) / 32
			max := int64(v) * 8
			u8 = uint8(RandRange(min, max))
		default:
			u8 = uint8(RandRange(0, 256))
		}
		return u8

	case uint16:
		if v == 0 {
			return uint16(RandUint64())
		}
		var u16 interface{}
		switch {
		case p < 0.3:
			min := int64(v) / 256
			max := int64(v) * 2
			u16 = uint16(RandRange(min, max))
		case p < 0.6:
			min := int64(v) / 9096
			max := int64(v) * 8
			u16 = uint16(RandRange(min, max))
		default:
			u16 = uint16(RandRange(0, 65536))
		}
		return u16

	case uint32:
		if v == 0 {
			return uint32(RandUint64())
		}
		var u32 interface{}
		switch {
		case p < 0.2:
			min := int64(v) / 256
			max := int64(v) * 2
			u32 = uint32(RandRange(min, max))
		case p < 0.4:
			min := int64(v) / 9096
			max := int64(v) * 8
			u32 = uint32(RandRange(min, max))
		case p < 0.6:
			min := int64(v) / (9096 * 9096)
			max := int64(v) * 16
			u32 = uint32(RandRange(min, max))
		default:
			u32 = uint32(RandRange(0, int64(^uint32(0))))
		}
		return u32

	case uint64:
		if v == 0 {
			return RandUint64()
		}
		var u64 interface{}
		switch {
		case p < 0.2:
			min := v / 256
			max := v * 2
			u64 = UintRandRange(min, max)
		case p < 0.5:
			min := v / 9096
			max := v * 8
			u64 = UintRandRange(min, max)
		case p < 0.5:
			min := v / (9096 * 9096)
			max := v * 16
			u64 = UintRandRange(min, max)
		default:
			u64 = RandUint64()
		}
		return u64

	default:
		panic("it's not a supported uint type")
	}
}

// randomXXXFrom gradually mutates XXX and return XXX
func randomFloatFrom(f interface{}) interface{} {
	switch v := f.(type) {
	case float32:
		var f32 float32
		f32 = randFloat32From(float32(v))
		return float32(f32)
	case float64:
		var f64 float64
		f64 = randFloat64From(float64(v))
		return float64(f64)
	default:
		panic("argument is not float 32 or float4")
	}
}

func randFloat32From(f float32) float32 {
	return randomFloat32(f)
}

func randFloat64From(f float64) float64 {
	return randomFloat64(f)
}

func randomBool() bool {
	return UniformRandomBool(0.5)
}

func randomFloat32(a float32) float32 {
	bits := math.Float32bits(a)

	exponent := int32((bits >> 23) & 0xFF)
	mantissa := bits & 0x7FFFFF
	sign := bits & 0x80000000
	t := uint32(UnixNano())
	manshift := 1 + (t % 7)

	var shift int32
	if exponent <= 1 {
		shift = 1 + int32(mantissa&1)
	} else if exponent >= 0xFE {
		shift = -1 - int32(mantissa&1)
	} else {
		shift = -2 + int32(mantissa&3)
	}

	newExp := exponent + shift
	newExponent := uint32(newExp)

	newMantissa := mantissa ^ (mantissa >> manshift)

	newBits := sign | (newExponent << 23) | (newMantissa & 0x7FFFFF)

	return math.Float32frombits(newBits)
}

func randomFloat64(a float64) float64 {
	bits := math.Float64bits(a)

	exponent := int64((bits >> 52) & 0x7FF)

	mantissa := bits & 0xFFFFFFFFFFFFF

	sign := bits & 0x8000000000000000

	t := uint64(UnixNano())
	manshift := 1 + (t % 7)

	var shift int64
	if exponent <= 1 {
		shift = 1 + int64(mantissa&1)
	} else if exponent >= 0x7FE {
		shift = -1 - int64(mantissa&1)
	} else {
		shift = -2 + int64(mantissa&3)
	}

	newExp := exponent + shift
	newExponent := uint64(newExp)

	newMantissa := mantissa ^ (mantissa >> manshift)

	newBits := sign | (newExponent << 52) | (newMantissa & 0xFFFFFFFFFFFFF)

	return math.Float64frombits(newBits)
}
