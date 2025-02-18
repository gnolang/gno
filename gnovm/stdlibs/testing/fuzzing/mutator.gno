package fuzzing

import (
	"math"
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
	boundarySamplePopularity := [8]uint64{2, 8, 32, 128, 512, 2048, 8096, 16192}
	mappedGeneratingNumber := [8]int{1440, 1080, 720, 600, 480, 360, 240, 120}
	genIdx := sort.Search(len(boundarySamplePopularity), func(i int) bool {
		return mutationSample.SamplePopularity <= boundarySamplePopularity[i]
	})
	var generatingNumber int
	if genIdx < len(boundarySamplePopularity) {
		generatingNumber = mappedGeneratingNumber[genIdx]
	} else {
		generatingNumber = 120
	}
	boundaryClusterPopularity := [4]uint64{32, 256, 4096, 259072}
	mappedMutationStrength := [4]int{5, 4, 3, 2}
	mutIdx := sort.Search(len(boundaryClusterPopularity), func(i int) bool {
		return m.data[mutationSample.Seed.HashNumber].ClusterPopularity <= boundaryClusterPopularity[i]
	})
	var mutationStrength int
	if mutIdx < len(boundaryClusterPopularity) {
		mutationStrength = mappedMutationStrength[mutIdx]
	} else {
		mutationStrength = 1
	}

	mutationSample.UpdatePowerSetCache()
	cum := mutationSample.GetPowerSetCache()

	chunk := 120 //* Must be a divisor of the generating number.

	chunkIters := generatingNumber / chunk
	var generatedSeeds []Seed
	for i := 0; i < chunkIters; i++ {
		var selectedPowerSetIdx int
		if len(cum) == 0 || cum[len(cum)-1] == 0 {
			selectedPowerSetIdx = 0
		} else {
			total := cum[len(cum)-1]
			r := GetPCG().Uint64n(total)
			selectedPowerSetIdx = sort.Search(len(cum), func(i int) bool {
				return cum[i] > r
			})
		}
		selectedSubSetArgs := m.argsPowerListExceptNull[selectedPowerSetIdx]

		// Completed the decision of the argument set to mutate.
		// From this loop onwards, actual seed generation begins.

		for j := 0; j < chunk; j++ {
			castedContent := efficientCasting(mutationSample.Seed.Content)
			mutationsInfo := make(map[int]MutationInfo)

			for _, idx := range selectedSubSetArgs {
				// Mutate the selected index element.
				var mutationInfo MutationInfo
				// Modify the value and receive the mutated data.
				castedContent[idx], mutationInfo = mutationSample.MutateArg(castedContent[idx], mutationStrength, idx)
				mutationsInfo[idx] = mutationInfo

			}
			// Synchronize seed counting.
			*seedCount += 1
			// Generate seed.
			newSeed := Seed{
				Content:        castedContent,
				Pid:            mutationSample.Seed.Id,
				Id:             *seedCount,
				Gen:            mutationSample.Seed.Gen + 1,
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
		if sample.Seed.Pid == pid {
			return id, true
		}
	}
	return 0, false
}

// FindParentID locates a sample that acts as a parent
// to the given child ID (pid).
func (mss *MutSampleSet) FindParentID(pid uint) (uint, bool) {
	for id, sample := range mss.Samples {
		if sample.Seed.Id == pid {
			return id, true
		}
	}
	return 0, false
}

// Add inserts a new sample into the MutSampleSet.
func (mss *MutSampleSet) Add(sample *MutSample) {
	mss.Samples[sample.Seed.Id] = sample
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
	Seed              Seed             // Original seed data
	ArgPowerSetScores []uint64         // Argument scores, considering aspects like [arguments, entity, strategy]
	SamplePopularity  uint64           // Number of times the sample has been leveraged
	RootEntsDict      map[int]*RootEnt // Mapping of argument index to RootEnt
	NotSbEntsDict     map[int]*EntForNotSB

	// Caching fields for cumulative sum calculations
	CachedPowerSetCumSum []uint64
	NeedsUpdateCache     bool
}

// MutateArg applies a mutation to the given input argument based on the specified strength
// and argument index. The mutation method is determined by the data type of the input.
func (ms *MutSample) MutateArg(input interface{}, strength int, argIdx int) (interface{}, MutationInfo) {
	switch input.(type) {
	case string:
		result, entID, usedStrategyMap := MutateRoot2String(ms.RootEntsDict[argIdx], strength)
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
		result, entID, usedStrategyMap := MutateRoot2String(ms.RootEntsDict[argIdx], strength)
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
		result, usedStrategyMap := MutateNotSbEnt2interface(ms.NotSbEntsDict[argIdx], strength)
		boolResult, ok := result.(bool)
		if !ok {
			panic("logical Error. If it occred, Needs to fix")
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
		result, usedStrategyMap := MutateNotSbEnt2interface(ms.NotSbEntsDict[argIdx], strength)
		float32Result, ok := result.(float32)
		if !ok {
			panic("logical Error. If it occred, Needs to fix")
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
		result, usedStrategyMap := MutateNotSbEnt2interface(ms.NotSbEntsDict[argIdx], strength)
		float64Result, ok := result.(float64)
		if !ok {
			panic("logical Error. If it occred, Needs to fix")
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
		result, usedStrategyMap := MutateNotSbEnt2interface(ms.NotSbEntsDict[argIdx], strength)
		intResult, ok := result.(int)
		if !ok {
			panic("logical Error. If it occred, Needs to fix")
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
		result, usedStrategyMap := MutateNotSbEnt2interface(ms.NotSbEntsDict[argIdx], strength)
		int8Result, ok := result.(int8)
		if !ok {
			panic("logical Error. If it occurred, Needs to fix")
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
		result, usedStrategyMap := MutateNotSbEnt2interface(ms.NotSbEntsDict[argIdx], strength)
		int16Result, ok := result.(int16)
		if !ok {
			panic("logical Error. If it occurred, Needs to fix")
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
		result, usedStrategyMap := MutateNotSbEnt2interface(ms.NotSbEntsDict[argIdx], strength)
		int32Result, ok := result.(int32)
		if !ok {
			panic("logical Error. If it occurred, Needs to fix")
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
		result, usedStrategyMap := MutateNotSbEnt2interface(ms.NotSbEntsDict[argIdx], strength)
		int64Result, ok := result.(int64)
		if !ok {
			panic("logical Error. If it occurred, Needs to fix")
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
		result, usedStrategyMap := MutateNotSbEnt2interface(ms.NotSbEntsDict[argIdx], strength)
		uintResult, ok := result.(uint)
		if !ok {
			panic("logical Error. If it occurred, Needs to fix")
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
		result, usedStrategyMap := MutateNotSbEnt2interface(ms.NotSbEntsDict[argIdx], strength)
		uint8Result, ok := result.(uint8)
		if !ok {
			panic("logical Error. If it occurred, Needs to fix")
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
		result, usedStrategyMap := MutateNotSbEnt2interface(ms.NotSbEntsDict[argIdx], strength)
		uint16Result, ok := result.(uint16)
		if !ok {
			panic("logical Error. If it occurred, Needs to fix")
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
		result, usedStrategyMap := MutateNotSbEnt2interface(ms.NotSbEntsDict[argIdx], strength)
		uint32Result, ok := result.(uint32)
		if !ok {
			panic("logical Error. If it occurred, Needs to fix")
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
		result, usedStrategyMap := MutateNotSbEnt2interface(ms.NotSbEntsDict[argIdx], strength)
		uint64Result, ok := result.(uint64)
		if !ok {
			panic("logical Error. If it occurred, Needs to fix")
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
		panic("logical Error. Type not implemented")
	}
}

// MutateRoot2String mutates a RootEnt structure and returns the resulting string,
// the ID of the mutated entity, and a map of the strategies used during mutation.
func MutateRoot2String(r *RootEnt, strength int) (string, int, map[MutationCode]int) {
	// Update score cache for selecting mutation target
	r.updateScoreCache()
	cachedScoreIds, cachedScoreCum := r.getScoreCache()

	// Select an entity ID for mutation based on weighted probability
	selectedID := selectID(cachedScoreIds, cachedScoreCum)
	// Extract surrounding string content while removing the selected entity
	startStr, endStr := SliceWithOutID(selectedID, r)
	targetPtr := r.Id2Ent[selectedID]
	// Perform mutation on the selected entity
	mutatedString, usedStrategy := MutateEnt2String(targetPtr, strength, r.Pattern, r.Strategy2count)
	// Reconstruct the final mutated string
	finalString := startStr + mutatedString + endStr

	return finalString, selectedID, usedStrategy
}

// selectID selects an ID from the cumulative sum slice (cum) using a random value.
func selectID(keys []int, cum []uint64) int {
	total := cum[len(cum)-1]
	r := GetPCG().Uint64n(total)
	idx := sort.Search(len(cum), func(i int) bool { return cum[i] > r })
	return keys[idx]
}

// MutateEnt2String mutates an entity (Ent) into a string using cached strategy weights.
// It selects mutation strategies probabilistically and applies them until the mutation
// strength limit is reached. If a mutation fails, the strategy is removed.
func MutateEnt2String(e Ent, strength int, pattern string, strategy2count map[MutationCode]uint64) (string, map[MutationCode]int) {
	clone := cloneEnt(e)

	strategyCounts := make(map[MutationCode]int)

	cnt := 0
	e.updateStrategyCache()
	cachedKeys, cachedCums := e.getStrategyCache()
	// Local copies for mutation
	localKeys := make([]MutationCode, len(cachedKeys))
	copy(localKeys, cachedKeys)
	localCums := make([]uint64, len(cachedCums))
	copy(localCums, cachedCums)

	for {
		// If all available strategies are exhausted, exit the loop.

		if len(localKeys) == 0 {
			return SerializeToString(clone), nil
			break
		}
		strategy, idx := selectStrategyWithIndex(localKeys, localCums)
		selectedMutator := entMutators[strategy]

		isSuccess := selectedMutator(clone, pattern)
		if isSuccess {
			strategyCounts[strategy]++
			cnt++
		} else {
			// If a failure occurs, compute the contribution value (delta) of the strategy
			// and remove it from localKeys and localCum.
			delta := strategy2count[strategy]
			localKeys = removeStrategyAt(localKeys, idx)
			localCums = updateCumulativeAfterRemoval(localCums, idx, delta)
			continue
		}
		if cnt >= strength {
			break
		}
	}
	result := SerializeToString(clone)
	return result, strategyCounts
}

// selectStrategyWithIndex selects a mutation strategy and its index from the cumulative sum slice.
func selectStrategyWithIndex(keys []MutationCode, cum []uint64) (MutationCode, int) {
	total := cum[len(cum)-1]
	rng := GetPCG()
	r := rng.Uint64n(total)
	idx := sort.Search(len(cum), func(i int) bool { return cum[i] > r })
	return keys[idx], idx
}

// removeStrategyAt removes an element from a slice at a given index.
func removeStrategyAt(keys []MutationCode, idx int) []MutationCode {
	return append(keys[:idx], keys[idx+1:]...)
}

// updateCumulativeAfterRemoval updates the cumulative sum slice after removing a strategy.
func updateCumulativeAfterRemoval(localCum []uint64, idx int, delta uint64) []uint64 {
	newCum := append([]uint64{}, localCum[:idx]...)
	for i := idx + 1; i < len(localCum); i++ {
		newCum = append(newCum, localCum[i]-delta)
	}
	return newCum
}

// MutateNotSbEnt2interface mutates a NotSb entity using cached strategy probabilities.
// It applies strategies until the strength limit is reached, removing failed strategies.
func MutateNotSbEnt2interface(ens *EntForNotSB, strength int) (interface{}, map[MutationCodeForNotSB]int) {
	clone := cloneEntForNotSb(ens)
	strategyCounts := make(map[MutationCodeForNotSB]int)
	cnt := 0
	ens.updateStrategyCache()
	cachedKeys, cachedCums := ens.getStrategyCache()

	localKeys := make([]MutationCodeForNotSB, len(cachedKeys))
	copy(localKeys, cachedKeys)
	localCums := make([]uint64, len(cachedCums))
	copy(localCums, cachedCums)

	for {
		if len(localKeys) == 0 {
			break
		}
		strategy, idx := selectStrategyWithIndexForNotSB(localKeys, localCums)
		selectedMutator := notSbEntMutators[strategy]
		if selectedMutator(clone) {
			strategyCounts[strategy]++
			cnt++
		} else {
			delta := ens.Strategy2count[strategy]
			localKeys = removeStrategyAtForNotSB(localKeys, idx)
			localCums = updateCumulativeAfterRemoval(localCums, idx, delta)
			continue
		}
		if cnt >= strength {
			break
		}
	}
	result := clone.Data
	return result, strategyCounts
}

// selectStrategyWithIndexForNotSB selects a NotSb mutation strategy and its index using a random value.
func selectStrategyWithIndexForNotSB(keys []MutationCodeForNotSB, cum []uint64) (MutationCodeForNotSB, int) {
	total := cum[len(cum)-1]
	rng := GetPCG()
	r := rng.Uint64n(total)
	idx := sort.Search(len(cum), func(i int) bool { return cum[i] > r })
	return keys[idx], idx
}

// removeStrategyAtForNotSB removes an element from a NotSb strategy slice at a given index.
func removeStrategyAtForNotSB(keys []MutationCodeForNotSB, idx int) []MutationCodeForNotSB {
	return append(keys[:idx], keys[idx+1:]...)
}

// ReflectSeed updates the mutation sample using the given seed's mutation information.
func (ms *MutSample) ReflectSeed(seed Seed) {
	ms.PlusArgPowerSetScore(uint64(seed.ArgPowerSetIdx), 2)
	ms.reflectMutationsInfo(seed.MutationsInfo)
	ms.PlusSamplePopularity(1)
}

// PlusArgPowerSetScore increases the score of a given argument index in the power set.
func (ms *MutSample) PlusArgPowerSetScore(i uint64, delta uint64) {
	if int(i) >= len(ms.ArgPowerSetScores) {
		newSize := int(i) + 1
		newScores := make([]uint64, newSize)
		copy(newScores, ms.ArgPowerSetScores)
		ms.ArgPowerSetScores = newScores
	}
	ms.ArgPowerSetScores[i] += delta
	ms.NeedsUpdateCache = true
}

// reflectMutationsInfo updates mutation data for all mutated arguments in the sample.
func (ms *MutSample) reflectMutationsInfo(mutInfoDict map[int]MutationInfo) {
	for argIdx, mi := range mutInfoDict {
		if !mi.IsMutated {
			continue
		}
		if mi.IsStringOrByteArray {
			if root, exists := ms.RootEntsDict[argIdx]; exists {
				root.ReflectMutationInfo(mi)
			}
		} else {
			if ens, exists := ms.NotSbEntsDict[argIdx]; exists {
				ens.ReflectMutationInfo(mi)
			}
		}
	}
}

// PlusSamplePopularity increases the sample's popularity count.
func (ms *MutSample) PlusSamplePopularity(delta int) {
	ms.SamplePopularity += uint64(delta)
}

// GetArgPowerSetScore returns the ArgPowerSetScore value at index i.
func (ms *MutSample) GetArgPowerSetScore(i uint64) uint64 {
	if int(i) >= len(ms.ArgPowerSetScores) {
		// If the index is out of bounds, it returns 0.
		return 0
	}
	return ms.ArgPowerSetScores[i]
}

// UpdatePowerSetCache updates the cached cumulative sum array based on ArgPowerSetScores.
func (ms *MutSample) UpdatePowerSetCache() {
	if !ms.NeedsUpdateCache && ms.CachedPowerSetCumSum != nil {
		return
	}
	n := len(ms.ArgPowerSetScores)
	ms.CachedPowerSetCumSum = make([]uint64, n)
	var total uint64 = 0
	for i, score := range ms.ArgPowerSetScores {
		total += score
		ms.CachedPowerSetCumSum[i] = total
	}
	ms.NeedsUpdateCache = false
}

// GetPowerSetCache retrieves the cached cumulative sum array. If it is outdated, it updates it first.
func (ms *MutSample) GetPowerSetCache() []uint64 {
	if ms.NeedsUpdateCache {
		ms.UpdatePowerSetCache()
	}
	return ms.CachedPowerSetCumSum
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
		Seed:                 seed,
		ArgPowerSetScores:    scores,
		SamplePopularity:     1,
		RootEntsDict:         formattedRootEntsDict,
		NotSbEntsDict:        formattedNotSbEntsDict,
		CachedPowerSetCumSum: nil, // Cache starts empty
		NeedsUpdateCache:     true,
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

// MutateFunc defines the type of functions that mutate Ent(string and []byte) values.
type MutateFunc func(Ent, string) bool

// MutationCode defines various mutation operations for modifying entities.
type MutationCode int

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

)

// MutateFuncForNotSB defines the type of functions that mutate EntForNotSB (not string or []byte) values.
type MutateFuncForNotSB func(*EntForNotSB) bool

// MutationCodeForNotSB defines mutation strategies for non-string/byte entities.
type MutationCodeForNotSB int

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
	PttrnChng                                // change to pattern value
	RndmChng                                 // chahge to radom value
)

// entMutators maps mutation codes to their respective transformation functions.
var entMutators = map[MutationCode]MutateFunc{
	DS:    deleteStart,
	DM:    deleteMiddle,
	DE:    deleteEnd,
	AS:    appendStart,
	ASI:   appendStartInteresting,
	ASP:   appendStartPattern,
	AE:    appendEnd,
	AEI:   appendEndInteresting,
	AEP:   appendEndPattern,
	OS:    overwriteStart,
	OM:    overwriteMiddle,
	CPS:   copyStart,
	CPM:   copyMiddle,
	CPE:   copyEnd,
	CHR:   changeRandomly,
	CHG:   changeGradual,
	CHG16: changeGradual16,
	CHG32: changeGradual32,
	CHG64: changeGradual64,
	CHI:   changeInteresting,
	CHI16: changeInteresting16,
	CHI32: changeInteresting32,
	CHP:   changePattern,
	IR:    insertRandomly,
	IG:    insertGradual,
	II:    insertInteresting,
	IP:    insertPattern,
	SW:    swap,
	PKL:   preserveKindLight,
	PKH:   preserveKindHard,
	SKL:   switchKindLight,
	SKH:   switchKindHard,
}

// notSbEntMutators maps non-string/byte mutation codes to their respective transformation functions.
var notSbEntMutators = map[MutationCodeForNotSB]MutateFuncForNotSB{
	SmlPls:       smallPlus,         // small plus
	SmlMns:       smallMinus,        // small minus
	SmlDv:        smallDiv,          // small division
	SmlMl:        smallMul,          // small multiply
	MddlPls:      middlePlus,        // middle plus
	MddlMins:     middleMinus,       // middle minus
	MddlDv:       middleDiv,         // middle division
	MddlMl:       middleMul,         // middcle multiply
	BgPls:        bigPlus,           // big plus
	BgMns:        bigMinus,          // big minus
	BgDv:         bigDiv,            // big division
	BgMl:         bigMul,            // big multiply
	FctrlDv:      factorialDiv,      // factorial division
	FctrlMl:      factorialMul,      // fcatorial multiply
	IntrstngChng: interestingChange, // change to interesting value
	RndmChng:     randomChange,
}

// TODO: Refine the mutation logic from this point onwards.
// The current implementation is a simple random generation and is still quite primitive.
// especially regarding the refinement of complex Ent types.
// Leverage randobyteFrom to handle value set issues when performing append and insert operations.
// Also, ensure that the function returns true in most cases to enable efficient fuzzing.
// It should only return false in unavoidable situations.
// The mutation refinement should also be performed based on the pattern value.
// The p parameter is intended for future use but is currently unused.

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
// deleteStart removes a portion of the beginning of the entity's Data.
func deleteStart(e Ent, p string) bool {
	switch e.(type) {
	case *LetterEnt, *NumEnt, *WsEnt, *MarkEnt:
		var data []byte
		switch t := e.(type) {
		case *LetterEnt:
			data = t.Data
		case *NumEnt:
			data = t.Data
		case *WsEnt:
			data = t.Data
		case *MarkEnt:
			data = t.Data
		}
		if data == nil || len(data) == 0 {
			return false
		}
		maxToDelete := int(float64(len(data)) * 0.6)
		if maxToDelete <= 0 {
			return false
		}
		deleteCount := GetPCG().Intn(maxToDelete + 1)
		if deleteCount == 0 {
			return false
		}
		newData := data[deleteCount:]
		switch t := e.(type) {
		case *LetterEnt:
			t.Data = newData
		case *NumEnt:
			t.Data = newData
		case *WsEnt:
			t.Data = newData
		case *MarkEnt:
			t.Data = newData
		}
		return true
	case *RootEnt:
		root := e.(*RootEnt)
		if len(root.Data) == 0 {
			// Block handling edge case for empty strings. (This is essential.)
			randomChar := byte(GetPCG().Intn(26) + 65)
			if GetPCG().Intn(2) == 0 {
				randomChar = byte(GetPCG().Intn(26) + 97)
			}
			root.Data = []Ent{
				&LetterEnt{Data: []byte{randomChar}},
			}
		}
		return true
	default:
		// TODO: Add mutation guidelines for complex types.
		// Composite types are not handled yet.
		return false
	}
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func deleteMiddle(e Ent, p string) bool {
	switch e.(type) {
	case *LetterEnt, *NumEnt, *WsEnt, *MarkEnt:
		var data []byte
		switch t := e.(type) {
		case *LetterEnt:
			data = t.Data
		case *NumEnt:
			data = t.Data
		case *WsEnt:
			data = t.Data
		case *MarkEnt:
			data = t.Data
		}
		if data == nil || len(data) < 3 {
			return true
		}
		maxToDelete := int(float64(len(data)) * 0.5)
		if maxToDelete <= 0 {
			return false
		}
		deleteCount := GetPCG().Intn(maxToDelete) + 1
		start := GetPCG().Intn(len(data)-deleteCount-1) + 1
		newData := append(data[:start], data[start+deleteCount:]...)
		switch t := e.(type) {
		case *LetterEnt:
			t.Data = newData
		case *NumEnt:
			t.Data = newData
		case *WsEnt:
			t.Data = newData
		case *MarkEnt:
			t.Data = newData
		}
		return true
	case *RootEnt:
		root := e.(*RootEnt)
		if len(root.Data) == 0 {

			randomChar := byte(GetPCG().Intn(26) + 65)
			if GetPCG().Intn(2) == 0 {
				randomChar = byte(GetPCG().Intn(26) + 97)
			}
			root.Data = []Ent{
				&LetterEnt{Data: []byte{randomChar}},
			}
		}
		return true
	default:
		return false
	}
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func deleteEnd(e Ent, p string) bool {
	switch e.(type) {
	case *LetterEnt, *NumEnt, *WsEnt, *MarkEnt:
		var data []byte
		switch t := e.(type) {
		case *LetterEnt:
			data = t.Data
		case *NumEnt:
			data = t.Data
		case *WsEnt:
			data = t.Data
		case *MarkEnt:
			data = t.Data
		}
		if data == nil || len(data) == 0 {
			return false
		}
		maxToDelete := int(float64(len(data)) * 0.5)
		if maxToDelete <= 0 {
			return false
		}
		deleteCount := GetPCG().Intn(maxToDelete) + 1
		newData := data[:len(data)-deleteCount]
		switch t := e.(type) {
		case *LetterEnt:
			t.Data = newData
		case *NumEnt:
			t.Data = newData
		case *WsEnt:
			t.Data = newData
		case *MarkEnt:
			t.Data = newData
		}
		return true
	case *RootEnt:
		root := e.(*RootEnt)
		if len(root.Data) == 0 {
			randomChar := byte(GetPCG().Intn(26) + 65)
			if GetPCG().Intn(2) == 0 {
				randomChar = byte(GetPCG().Intn(26) + 97)
			}
			root.Data = []Ent{
				&LetterEnt{Data: []byte{randomChar}},
			}
		}
		return true
	default:
		return false
	}
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func appendStart(e Ent, p string) bool {
	switch e.(type) {
	case *LetterEnt, *NumEnt, *WsEnt, *MarkEnt:
		var data []byte
		switch t := e.(type) {
		case *LetterEnt:
			data = t.Data
		case *NumEnt:
			data = t.Data
		case *WsEnt:
			data = t.Data
		case *MarkEnt:
			data = t.Data
		}
		randomByte := byte(GetPCG().Intn(256))
		newData := append([]byte{randomByte}, data...)
		switch t := e.(type) {
		case *LetterEnt:
			t.Data = newData
		case *NumEnt:
			t.Data = newData
		case *WsEnt:
			t.Data = newData
		case *MarkEnt:
			t.Data = newData
		}
		return true
	case *RootEnt:
		root := e.(*RootEnt)
		if len(root.Data) == 0 {

			randomChar := byte(GetPCG().Intn(26) + 65)
			if GetPCG().Intn(2) == 0 {
				randomChar = byte(GetPCG().Intn(26) + 97)
			}
			root.Data = []Ent{
				&LetterEnt{Data: []byte{randomChar}},
			}
		}
		return true

	default:
		return false
	}
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func appendStartInteresting(e Ent, p string) bool {
	return appendStart(e, p)
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func appendStartPattern(e Ent, p string) bool {
	return appendStart(e, p)
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func appendEnd(e Ent, p string) bool {
	switch e.(type) {
	case *LetterEnt, *NumEnt, *WsEnt, *MarkEnt:
		var data []byte
		switch t := e.(type) {
		case *LetterEnt:
			data = t.Data
		case *NumEnt:
			data = t.Data
		case *WsEnt:
			data = t.Data
		case *MarkEnt:
			data = t.Data
		}
		randomByte := byte(GetPCG().Intn(256))
		newData := append(data, randomByte)
		switch t := e.(type) {
		case *LetterEnt:
			t.Data = newData
		case *NumEnt:
			t.Data = newData
		case *WsEnt:
			t.Data = newData
		case *MarkEnt:
			t.Data = newData
		}
		return true
	case *RootEnt:
		root := e.(*RootEnt)
		if len(root.Data) == 0 {

			randomChar := byte(GetPCG().Intn(26) + 65)
			if GetPCG().Intn(2) == 0 {
				randomChar = byte(GetPCG().Intn(26) + 97)
			}
			root.Data = []Ent{
				&LetterEnt{Data: []byte{randomChar}},
			}
		}
		return true
	default:
		return false
	}
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func appendEndInteresting(e Ent, p string) bool {
	return appendEnd(e, p)
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func appendEndPattern(e Ent, p string) bool {
	return appendEnd(e, p)
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func overwriteStart(e Ent, p string) bool {
	switch e.(type) {
	case *LetterEnt, *NumEnt, *WsEnt, *MarkEnt:
		var data []byte
		switch t := e.(type) {
		case *LetterEnt:
			data = t.Data
		case *NumEnt:
			data = t.Data
		case *WsEnt:
			data = t.Data
		case *MarkEnt:
			data = t.Data
		}
		if data == nil || len(data) == 0 {
			return false
		}
		maxToOverwrite := int(float64(len(data)) * 0.3)
		if maxToOverwrite <= 0 {
			maxToOverwrite = 1
		}
		overwriteCount := GetPCG().Intn(maxToOverwrite) + 1
		for i := 0; i < overwriteCount && i < len(data); i++ {
			data[i] = byte(GetPCG().Intn(256))
		}
		switch t := e.(type) {
		case *LetterEnt:
			t.Data = data
		case *NumEnt:
			t.Data = data
		case *WsEnt:
			t.Data = data
		case *MarkEnt:
			t.Data = data
		}
		return true
	case *RootEnt:
		root := e.(*RootEnt)
		if len(root.Data) == 0 {

			randomChar := byte(GetPCG().Intn(26) + 65)
			if GetPCG().Intn(2) == 0 {
				randomChar = byte(GetPCG().Intn(26) + 97)
			}
			root.Data = []Ent{
				&LetterEnt{Data: []byte{randomChar}},
			}
		}
		return true
	default:
		return false
	}
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func overwriteMiddle(e Ent, p string) bool {
	switch e.(type) {
	case *LetterEnt, *NumEnt, *WsEnt, *MarkEnt:
		var data []byte
		switch t := e.(type) {
		case *LetterEnt:
			data = t.Data
		case *NumEnt:
			data = t.Data
		case *WsEnt:
			data = t.Data
		case *MarkEnt:
			data = t.Data
		}
		if data == nil || len(data) < 3 {
			return false
		}

		start := GetPCG().Intn(len(data)/2) + 1

		maxOverwrite := len(data) - start
		if maxOverwrite <= 0 {
			return false
		}
		overwriteCount := GetPCG().Intn(maxOverwrite)
		if overwriteCount == 0 {
			overwriteCount = 1
		}
		for i := start; i < start+overwriteCount && i < len(data); i++ {
			data[i] = byte(GetPCG().Intn(256))
		}
		switch t := e.(type) {
		case *LetterEnt:
			t.Data = data
		case *NumEnt:
			t.Data = data
		case *WsEnt:
			t.Data = data
		case *MarkEnt:
			t.Data = data
		}
		return true
	case *RootEnt:
		root := e.(*RootEnt)
		if len(root.Data) == 0 {

			randomChar := byte(GetPCG().Intn(26) + 65)
			if GetPCG().Intn(2) == 0 {
				randomChar = byte(GetPCG().Intn(26) + 97)
			}
			root.Data = []Ent{
				&LetterEnt{Data: []byte{randomChar}},
			}
		}
		return true
	default:
		return false
	}
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func copyStart(e Ent, p string) bool {
	switch e.(type) {
	case *LetterEnt, *NumEnt, *WsEnt, *MarkEnt:
		var data []byte
		switch t := e.(type) {
		case *LetterEnt:
			data = t.Data
		case *NumEnt:
			data = t.Data
		case *WsEnt:
			data = t.Data
		case *MarkEnt:
			data = t.Data
		}
		if data == nil || len(data) < 1 {
			return false
		}
		if len(data) < 1 {
			return false
		}
		copyLength := GetPCG().Intn(len(data)) + 1
		segment := data[:copyLength]
		newData := append(segment, data...)
		switch t := e.(type) {
		case *LetterEnt:
			t.Data = newData
		case *NumEnt:
			t.Data = newData
		case *WsEnt:
			t.Data = newData
		case *MarkEnt:
			t.Data = newData
		}
		return true
	case *RootEnt:
		root := e.(*RootEnt)
		if len(root.Data) == 0 {

			randomChar := byte(GetPCG().Intn(26) + 65)
			if GetPCG().Intn(2) == 0 {
				randomChar = byte(GetPCG().Intn(26) + 97)
			}
			root.Data = []Ent{
				&LetterEnt{Data: []byte{randomChar}},
			}
		}
		return true
	default:
		return false
	}
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func copyMiddle(e Ent, p string) bool {
	switch e.(type) {
	case *LetterEnt, *NumEnt, *WsEnt, *MarkEnt:
		var data []byte
		switch t := e.(type) {
		case *LetterEnt:
			data = t.Data
		case *NumEnt:
			data = t.Data
		case *WsEnt:
			data = t.Data
		case *MarkEnt:
			data = t.Data
		}
		if data == nil || len(data) < 3 {
			return false
		}

		start := GetPCG().Intn(len(data)/2) + 1
		maxCopy := len(data) - start
		if maxCopy <= 0 {
			return false
		}
		copyLength := GetPCG().Intn(maxCopy) + 1
		segment := data[start : start+copyLength]
		newData := append(data[:start], append(segment, data[start:]...)...)
		switch t := e.(type) {
		case *LetterEnt:
			t.Data = newData
		case *NumEnt:
			t.Data = newData
		case *WsEnt:
			t.Data = newData
		case *MarkEnt:
			t.Data = newData
		}
		return true
	case *RootEnt:
		root := e.(*RootEnt)
		if len(root.Data) == 0 {

			randomChar := byte(GetPCG().Intn(26) + 65)
			if GetPCG().Intn(2) == 0 {
				randomChar = byte(GetPCG().Intn(26) + 97)
			}
			root.Data = []Ent{
				&LetterEnt{Data: []byte{randomChar}},
			}
		}
		return true
	default:
		return false
	}
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func copyEnd(e Ent, p string) bool {
	switch e.(type) {
	case *LetterEnt, *NumEnt, *WsEnt, *MarkEnt:
		var data []byte
		switch t := e.(type) {
		case *LetterEnt:
			data = t.Data
		case *NumEnt:
			data = t.Data
		case *WsEnt:
			data = t.Data
		case *MarkEnt:
			data = t.Data
		}
		if data == nil || len(data) < 1 {
			return false
		}
		if len(data) == 0 {
			return false
		}
		copyLength := GetPCG().Intn(len(data)) + 1
		segment := data[len(data)-copyLength:]
		newData := append(data, segment...)
		switch t := e.(type) {
		case *LetterEnt:
			t.Data = newData
		case *NumEnt:
			t.Data = newData
		case *WsEnt:
			t.Data = newData
		case *MarkEnt:
			t.Data = newData
		}
		return true
	case *RootEnt:
		root := e.(*RootEnt)
		if len(root.Data) == 0 {

			randomChar := byte(GetPCG().Intn(26) + 65)
			if GetPCG().Intn(2) == 0 {
				randomChar = byte(GetPCG().Intn(26) + 97)
			}
			root.Data = []Ent{
				&LetterEnt{Data: []byte{randomChar}},
			}
		}
		return true
	default:
		return false
	}
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func changeRandomly(e Ent, p string) bool {
	switch e.(type) {
	case *LetterEnt, *NumEnt, *WsEnt, *MarkEnt:
		var data []byte
		switch t := e.(type) {
		case *LetterEnt:
			data = t.Data
		case *NumEnt:
			data = t.Data
		case *WsEnt:
			data = t.Data
		case *MarkEnt:
			data = t.Data
		}
		if data == nil || len(data) == 0 {
			return false
		}

		idx := GetPCG().Intn(len(data))
		data[idx] = byte(GetPCG().Intn(256))
		return true
	case *RootEnt:
		root := e.(*RootEnt)
		if len(root.Data) == 0 {

			randomChar := byte(GetPCG().Intn(26) + 65)
			if GetPCG().Intn(2) == 0 {
				randomChar = byte(GetPCG().Intn(26) + 97)
			}
			root.Data = []Ent{
				&LetterEnt{Data: []byte{randomChar}},
			}
		}
		return true
	default:
		return false
	}
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func changeGradual(e Ent, p string) bool {
	switch e.(type) {
	case *LetterEnt, *NumEnt, *WsEnt, *MarkEnt:
		var data []byte
		switch t := e.(type) {
		case *LetterEnt:
			data = t.Data
		case *NumEnt:
			data = t.Data
		case *WsEnt:
			data = t.Data
		case *MarkEnt:
			data = t.Data
		}
		if data == nil || len(data) == 0 {
			return false
		}
		idx := GetPCG().Intn(len(data))
		if GetPCG().Intn(2) == 0 && data[idx] > 0 {
			data[idx]--
		} else {
			data[idx]++
		}
		return true
	case *RootEnt:
		root := e.(*RootEnt)
		if len(root.Data) == 0 {
			root.Data = []Ent{
				&LetterEnt{Data: []byte{byte(GetPCG().Intn(256))}},
			}
		}
		return true
	default:
		return false
	}
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func changeGradual16(e Ent, p string) bool {
	switch e.(type) {
	case *LetterEnt, *NumEnt, *WsEnt, *MarkEnt:
		var data []byte
		switch t := e.(type) {
		case *LetterEnt:
			data = t.Data
		case *NumEnt:
			data = t.Data
		case *WsEnt:
			data = t.Data
		case *MarkEnt:
			data = t.Data
		}
		if data == nil || len(data) < 2 {
			return false
		}
		idx := GetPCG().Intn(len(data) - 1)
		for i := 0; i < 2; i++ {
			if GetPCG().Intn(2) == 0 && data[idx+i] > 0 {
				data[idx+i]--
			} else {
				data[idx+i]++
			}
		}
		return true
	case *RootEnt:

		root := e.(*RootEnt)
		if len(root.Data) == 0 {
			root.Data = []Ent{
				&LetterEnt{Data: []byte{byte(GetPCG().Intn(256))}},
			}
		}
		return true
	default:
		return false
	}
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func changeGradual32(e Ent, p string) bool {
	switch e.(type) {
	case *LetterEnt, *NumEnt, *WsEnt, *MarkEnt:
		var data []byte
		switch t := e.(type) {
		case *LetterEnt:
			data = t.Data
		case *NumEnt:
			data = t.Data
		case *WsEnt:
			data = t.Data
		case *MarkEnt:
			data = t.Data
		}
		if data == nil || len(data) < 4 {
			return false
		}
		idx := GetPCG().Intn(len(data) - 3)
		for i := 0; i < 4; i++ {
			if GetPCG().Intn(2) == 0 && data[idx+i] > 0 {
				data[idx+i]--
			} else {
				data[idx+i]++
			}
		}
		return true
	default:
		return false
	}
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func changeGradual64(e Ent, p string) bool {
	switch e.(type) {
	case *LetterEnt, *NumEnt, *WsEnt, *MarkEnt:
		var data []byte
		switch t := e.(type) {
		case *LetterEnt:
			data = t.Data
		case *NumEnt:
			data = t.Data
		case *WsEnt:
			data = t.Data
		case *MarkEnt:
			data = t.Data
		}
		if data == nil || len(data) < 8 {
			return false
		}
		idx := GetPCG().Intn(len(data) - 7)
		for i := 0; i < 8; i++ {
			if GetPCG().Intn(2) == 0 && data[idx+i] > 0 {
				data[idx+i]--
			} else {
				data[idx+i]++
			}
		}
		return true
	default:
		return false
	}
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func changeInteresting(e Ent, p string) bool {
	interestingValues := []byte{0, 1, 0x7F, 0xFF}
	switch e.(type) {
	case *LetterEnt, *NumEnt, *WsEnt, *MarkEnt:
		var data []byte
		switch t := e.(type) {
		case *LetterEnt:
			data = t.Data
		case *NumEnt:
			data = t.Data
		case *WsEnt:
			data = t.Data
		case *MarkEnt:
			data = t.Data
		}
		if data == nil || len(data) == 0 {
			return false
		}
		idx := GetPCG().Intn(len(data))
		data[idx] = interestingValues[GetPCG().Intn(len(interestingValues))]
		return true
	default:
		return false
	}
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func changeInteresting16(e Ent, p string) bool {
	interesting16 := [][]byte{{0, 0}, {0, 1}, {0x7F, 0xFF}, {0xFF, 0xFF}}
	switch e.(type) {
	case *LetterEnt, *NumEnt, *WsEnt, *MarkEnt:
		var data []byte
		switch t := e.(type) {
		case *LetterEnt:
			data = t.Data
		case *NumEnt:
			data = t.Data
		case *WsEnt:
			data = t.Data
		case *MarkEnt:
			data = t.Data
		}
		if data == nil || len(data) < 2 {
			return false
		}
		idx := GetPCG().Intn(len(data) - 1)
		pair := interesting16[GetPCG().Intn(len(interesting16))]
		data[idx] = pair[0]
		data[idx+1] = pair[1]
		return true
	default:
		return false
	}
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func changeInteresting32(e Ent, p string) bool {
	interesting32 := [][]byte{
		{0, 0, 0, 0},
		{0, 0, 0, 1},
		{0x7F, 0xFF, 0xFF, 0xFF},
		{0xFF, 0xFF, 0xFF, 0xFF},
	}
	switch e.(type) {
	case *LetterEnt, *NumEnt, *WsEnt, *MarkEnt:
		var data []byte
		switch t := e.(type) {
		case *LetterEnt:
			data = t.Data
		case *NumEnt:
			data = t.Data
		case *WsEnt:
			data = t.Data
		case *MarkEnt:
			data = t.Data
		}
		if data == nil || len(data) < 4 {
			return false
		}
		idx := GetPCG().Intn(len(data) - 3)
		pattern := interesting32[GetPCG().Intn(len(interesting32))]
		for i := 0; i < 4; i++ {
			data[idx+i] = pattern[i]
		}
		return true
	default:
		return false
	}
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func changePattern(e Ent, p string) bool {
	pattern := byte(0xAA)
	switch e.(type) {
	case *LetterEnt, *NumEnt, *WsEnt, *MarkEnt:
		var data []byte
		switch t := e.(type) {
		case *LetterEnt:
			data = t.Data
		case *NumEnt:
			data = t.Data
		case *WsEnt:
			data = t.Data
		case *MarkEnt:
			data = t.Data
		}
		if data == nil || len(data) == 0 {
			return false
		}
		idx := GetPCG().Intn(len(data))
		data[idx] = pattern
		return true
	default:
		return false
	}
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func insertRandomly(e Ent, p string) bool {
	switch e.(type) {
	case *LetterEnt, *NumEnt, *WsEnt, *MarkEnt:
		var data []byte
		switch t := e.(type) {
		case *LetterEnt:
			data = t.Data
		case *NumEnt:
			data = t.Data
		case *WsEnt:
			data = t.Data
		case *MarkEnt:
			data = t.Data
		}
		randomByte := byte(GetPCG().Intn(256))
		pos := 0
		if len(data) > 0 {
			pos = GetPCG().Intn(len(data))
		}
		newData := append(data[:pos], append([]byte{randomByte}, data[pos:]...)...)
		switch t := e.(type) {
		case *LetterEnt:
			t.Data = newData
		case *NumEnt:
			t.Data = newData
		case *WsEnt:
			t.Data = newData
		case *MarkEnt:
			t.Data = newData
		}
		return true
	default:
		return false
	}
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func insertGradual(e Ent, p string) bool {
	return insertRandomly(e, p)
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func insertInteresting(e Ent, p string) bool {
	interestingValues := []byte{0, 1, 0x7F, 0xFF}
	switch e.(type) {
	case *LetterEnt, *NumEnt, *WsEnt, *MarkEnt:
		var data []byte
		switch t := e.(type) {
		case *LetterEnt:
			data = t.Data
		case *NumEnt:
			data = t.Data
		case *WsEnt:
			data = t.Data
		case *MarkEnt:
			data = t.Data
		}
		value := interestingValues[GetPCG().Intn(len(interestingValues))]
		pos := 0
		if len(data) > 0 {
			pos = GetPCG().Intn(len(data))
		}
		newData := append(data[:pos], append([]byte{value}, data[pos:]...)...)
		switch t := e.(type) {
		case *LetterEnt:
			t.Data = newData
		case *NumEnt:
			t.Data = newData
		case *WsEnt:
			t.Data = newData
		case *MarkEnt:
			t.Data = newData
		}
		return true
	default:
		return false
	}
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func insertPattern(e Ent, p string) bool {
	return insertRandomly(e, p)
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func swap(e Ent, p string) bool {
	switch e.(type) {
	case *LetterEnt, *NumEnt, *WsEnt, *MarkEnt:
		var data []byte
		switch t := e.(type) {
		case *LetterEnt:
			data = t.Data
		case *NumEnt:
			data = t.Data
		case *WsEnt:
			data = t.Data
		case *MarkEnt:
			data = t.Data
		}
		if data == nil || len(data) < 2 {
			return false
		}
		idx1 := GetPCG().Intn(len(data))
		idx2 := GetPCG().Intn(len(data))
		data[idx1], data[idx2] = data[idx2], data[idx1]
		return true
	default:
		return false
	}
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func preserveKindLight(e Ent, p string) bool {
	return false
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func preserveKindHard(e Ent, p string) bool {
	return false
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func switchKindLight(e Ent, p string) bool {
	return false
}

// TODO: Fine-tuning is, of course, needed in the future.
// Must add mutation guidelines for complex types.
func switchKindHard(e Ent, p string) bool {
	return false
}

///
// From this point onward are mutation functions for EntForNotSB.
///

// TODO: Fine-tuning is, of course, needed in the future.
func mutateBool(ens *EntForNotSB) bool {
	b, ok := ens.Data.(bool)
	if !ok {
		return false
	}
	ens.Data = !b
	return true
}

// TODO: Fine-tuning is, of course, needed in the future.
func smallPlus(ens *EntForNotSB) bool {
	if ens.IsBool {
		return mutateBool(ens)
	}
	switch v := ens.Data.(type) {
	case int:
		ens.Data = v + int(1)
	case int8:
		ens.Data = v + int8(1)
	case int16:
		ens.Data = v + int16(1)
	case int32:
		ens.Data = v + int32(1)
	case int64:
		ens.Data = v + int64(1)
	case uint:
		ens.Data = v + uint(1)
	case uint8:
		ens.Data = v + uint8(1)
	case uint16:
		ens.Data = v + uint16(1)
	case uint32:
		ens.Data = v + uint32(1)
	case uint64:
		ens.Data = v + uint64(1)
	case float32:
		ens.Data = v + float32(1.0)
	case float64:
		ens.Data = v + float64(1.0)
	default:
		return false
	}
	return true
}

// TODO: Fine-tuning is, of course, needed in the future.
func smallMinus(ens *EntForNotSB) bool {
	if ens.IsBool {
		return mutateBool(ens)
	}
	switch v := ens.Data.(type) {
	case int:
		ens.Data = v - int(1)
	case int8:
		ens.Data = v - int8(1)
	case int16:
		ens.Data = v - int16(1)
	case int32:
		ens.Data = v - int32(1)
	case int64:
		ens.Data = v - int64(1)
	case uint:
		if v == 0 {
			ens.Data = uint(0)
		} else {
			ens.Data = v - uint(1)
		}
	case uint8:
		if v == 0 {
			ens.Data = uint8(0)
		} else {
			ens.Data = v - uint8(1)
		}
	case uint16:
		if v == 0 {
			ens.Data = uint16(0)
		} else {
			ens.Data = v - uint16(1)
		}
	case uint32:
		if v == 0 {
			ens.Data = uint32(0)
		} else {
			ens.Data = v - uint32(1)
		}
	case uint64:
		if v == 0 {
			ens.Data = uint64(0)
		} else {
			ens.Data = v - uint64(1)
		}
	case float32:
		ens.Data = v - float32(1.0)
	case float64:
		ens.Data = v - float64(1.0)
	default:
		return false
	}
	return true
}

// TODO: Fine-tuning is, of course, needed in the future.
func smallDiv(ens *EntForNotSB) bool {
	if ens.IsBool {
		return mutateBool(ens)
	}
	switch v := ens.Data.(type) {
	case int:
		if v == 0 {
			return false
		}
		ens.Data = v / int(2)
	case int8:
		if v == 0 {
			return false
		}
		ens.Data = v / int8(2)
	case int16:
		if v == 0 {
			return false
		}
		ens.Data = v / int16(2)
	case int32:
		if v == 0 {
			return false
		}
		ens.Data = v / int32(2)
	case int64:
		if v == 0 {
			return false
		}
		ens.Data = v / int64(2)
	case uint:
		if v == 0 {
			return false
		}
		ens.Data = v / uint(2)
	case uint8:
		if v == 0 {
			return false
		}
		ens.Data = v / uint8(2)
	case uint16:
		if v == 0 {
			return false
		}
		ens.Data = v / uint16(2)
	case uint32:
		if v == 0 {
			return false
		}
		ens.Data = v / uint32(2)
	case uint64:
		if v == 0 {
			return false
		}
		ens.Data = v / uint64(2)
	case float32:
		if v == 0.0 {
			return false
		}
		ens.Data = v / float32(2.0)
	case float64:
		if v == 0.0 {
			return false
		}
		ens.Data = v / float64(2.0)
	default:
		return false
	}
	return true
}

// TODO: Fine-tuning is, of course, needed in the future.
func smallMul(ens *EntForNotSB) bool {
	if ens.IsBool {
		return mutateBool(ens)
	}
	switch v := ens.Data.(type) {
	case int:
		ens.Data = v * int(2)
	case int8:
		ens.Data = v * int8(2)
	case int16:
		ens.Data = v * int16(2)
	case int32:
		ens.Data = v * int32(2)
	case int64:
		ens.Data = v * int64(2)
	case uint:
		ens.Data = v * uint(2)
	case uint8:
		ens.Data = v * uint8(2)
	case uint16:
		ens.Data = v * uint16(2)
	case uint32:
		ens.Data = v * uint32(2)
	case uint64:
		ens.Data = v * uint64(2)
	case float32:
		ens.Data = v * float32(2.0)
	case float64:
		ens.Data = v * float64(2.0)
	default:
		return false
	}
	return true
}

// TODO: Fine-tuning is, of course, needed in the future.
func middlePlus(ens *EntForNotSB) bool {
	if ens.IsBool {
		return mutateBool(ens)
	}
	switch v := ens.Data.(type) {
	case int:
		ens.Data = v + int(10)
	case int8:
		ens.Data = v + int8(10)
	case int16:
		ens.Data = v + int16(10)
	case int32:
		ens.Data = v + int32(10)
	case int64:
		ens.Data = v + int64(10)
	case uint:
		ens.Data = v + uint(10)
	case uint8:
		ens.Data = v + uint8(10)
	case uint16:
		ens.Data = v + uint16(10)
	case uint32:
		ens.Data = v + uint32(10)
	case uint64:
		ens.Data = v + uint64(10)
	case float32:
		ens.Data = v + float32(10.0)
	case float64:
		ens.Data = v + float64(10.0)
	default:
		return false
	}
	return true
}

// TODO: Fine-tuning is, of course, needed in the future.
func middleMinus(ens *EntForNotSB) bool {
	if ens.IsBool {
		return mutateBool(ens)
	}
	switch v := ens.Data.(type) {
	case int:
		ens.Data = v - int(10)
	case int8:
		ens.Data = v - int8(10)
	case int16:
		ens.Data = v - int16(10)
	case int32:
		ens.Data = v - int32(10)
	case int64:
		ens.Data = v - int64(10)
	case uint:
		if v < uint(10) {
			ens.Data = uint(0)
		} else {
			ens.Data = v - uint(10)
		}
	case uint8:
		if v < uint8(10) {
			ens.Data = uint8(0)
		} else {
			ens.Data = v - uint8(10)
		}
	case uint16:
		if v < uint16(10) {
			ens.Data = uint16(0)
		} else {
			ens.Data = v - uint16(10)
		}
	case uint32:
		if v < uint32(10) {
			ens.Data = uint32(0)
		} else {
			ens.Data = v - uint32(10)
		}
	case uint64:
		if v < uint64(10) {
			ens.Data = uint64(0)
		} else {
			ens.Data = v - uint64(10)
		}
	case float32:
		ens.Data = v - float32(10.0)
	case float64:
		ens.Data = v - float64(10.0)
	default:
		return false
	}
	return true
}

// TODO: Fine-tuning is, of course, needed in the future.
func middleDiv(ens *EntForNotSB) bool {
	if ens.IsBool {
		return mutateBool(ens)
	}
	switch v := ens.Data.(type) {
	case int:
		if v == 0 {
			return false
		}
		ens.Data = v / int(2)
	case int8:
		if v == 0 {
			return false
		}
		ens.Data = v / int8(2)
	case int16:
		if v == 0 {
			return false
		}
		ens.Data = v / int16(2)
	case int32:
		if v == 0 {
			return false
		}
		ens.Data = v / int32(2)
	case int64:
		if v == 0 {
			return false
		}
		ens.Data = v / int64(2)
	case uint:
		if v == 0 {
			return false
		}
		ens.Data = v / uint(2)
	case uint8:
		if v == 0 {
			return false
		}
		ens.Data = v / uint8(2)
	case uint16:
		if v == 0 {
			return false
		}
		ens.Data = v / uint16(2)
	case uint32:
		if v == 0 {
			return false
		}
		ens.Data = v / uint32(2)
	case uint64:
		if v == 0 {
			return false
		}
		ens.Data = v / uint64(2)
	case float32:
		if v == 0.0 {
			return false
		}
		ens.Data = v / float32(2.0)
	case float64:
		if v == 0.0 {
			return false
		}
		ens.Data = v / float64(2.0)
	default:
		return false
	}
	return true
}

// TODO: Fine-tuning is, of course, needed in the future.
func middleMul(ens *EntForNotSB) bool {
	if ens.IsBool {
		return mutateBool(ens)
	}
	switch v := ens.Data.(type) {
	case int:
		ens.Data = v * int(2)
	case int8:
		ens.Data = v * int8(2)
	case int16:
		ens.Data = v * int16(2)
	case int32:
		ens.Data = v * int32(2)
	case int64:
		ens.Data = v * int64(2)
	case uint:
		ens.Data = v * uint(2)
	case uint8:
		ens.Data = v * uint8(2)
	case uint16:
		ens.Data = v * uint16(2)
	case uint32:
		ens.Data = v * uint32(2)
	case uint64:
		ens.Data = v * uint64(2)
	case float32:
		ens.Data = v * float32(2.0)
	case float64:
		ens.Data = v * float64(2.0)
	default:
		return false
	}
	return true
}

// TODO: Fine-tuning is, of course, needed in the future.
func bigPlus(ens *EntForNotSB) bool {
	if ens.IsBool {
		return mutateBool(ens)
	}
	switch v := ens.Data.(type) {
	case int:
		ens.Data = v + int(1000)
	case int8:
		ens.Data = v + int8(100)
	case int16:
		ens.Data = v + int16(1000)
	case int32:
		ens.Data = v + int32(1000)
	case int64:
		ens.Data = v + int64(1000)
	case uint:
		ens.Data = v + uint(1000)
	case uint8:
		ens.Data = v + uint8(100)
	case uint16:
		ens.Data = v + uint16(1000)
	case uint32:
		ens.Data = v + uint32(1000)
	case uint64:
		ens.Data = v + uint64(1000)
	case float32:
		ens.Data = v + float32(1000.0)
	case float64:
		ens.Data = v + float64(1000.0)
	default:
		return false
	}
	return true
}

// TODO: Fine-tuning is, of course, needed in the future.
func bigMinus(ens *EntForNotSB) bool {
	if ens.IsBool {
		return mutateBool(ens)
	}
	switch v := ens.Data.(type) {
	case int:
		ens.Data = v - int(1000)
	case int8:
		ens.Data = v - int8(100)
	case int16:
		ens.Data = v - int16(1000)
	case int32:
		ens.Data = v - int32(1000)
	case int64:
		ens.Data = v - int64(1000)
	case uint:
		if v < uint(1000) {
			ens.Data = uint(0)
		} else {
			ens.Data = v - uint(1000)
		}
	case uint8:
		if v < uint8(100) {
			ens.Data = uint8(0)
		} else {
			ens.Data = v - uint8(100)
		}
	case uint16:
		if v < uint16(1000) {
			ens.Data = uint16(0)
		} else {
			ens.Data = v - uint16(1000)
		}
	case uint32:
		if v < uint32(1000) {
			ens.Data = uint32(0)
		} else {
			ens.Data = v - uint32(1000)
		}
	case uint64:
		if v < uint64(1000) {
			ens.Data = uint64(0)
		} else {
			ens.Data = v - uint64(1000)
		}
	case float32:
		ens.Data = v - float32(1000.0)
	case float64:
		ens.Data = v - float64(1000.0)
	default:
		return false
	}
	return true
}

// TODO: Fine-tuning is, of course, needed in the future.
func bigDiv(ens *EntForNotSB) bool {
	if ens.IsBool {
		return mutateBool(ens)
	}
	switch v := ens.Data.(type) {
	case int:
		if v == 0 {
			return false
		}
		ens.Data = v / int(2)
	case int8:
		if v == 0 {
			return false
		}
		ens.Data = v / int8(2)
	case int16:
		if v == 0 {
			return false
		}
		ens.Data = v / int16(2)
	case int32:
		if v == 0 {
			return false
		}
		ens.Data = v / int32(2)
	case int64:
		if v == 0 {
			return false
		}
		ens.Data = v / int64(2)
	case uint:
		if v == 0 {
			return false
		}
		ens.Data = v / uint(2)
	case uint8:
		if v == 0 {
			return false
		}
		ens.Data = v / uint8(2)
	case uint16:
		if v == 0 {
			return false
		}
		ens.Data = v / uint16(2)
	case uint32:
		if v == 0 {
			return false
		}
		ens.Data = v / uint32(2)
	case uint64:
		if v == 0 {
			return false
		}
		ens.Data = v / uint64(2)
	case float32:
		if v == 0.0 {
			return false
		}
		ens.Data = v / float32(2.0)
	case float64:
		if v == 0.0 {
			return false
		}
		ens.Data = v / float64(2.0)
	default:
		return false
	}
	return true
}

// TODO: Fine-tuning is, of course, needed in the future.
func bigMul(ens *EntForNotSB) bool {
	if ens.IsBool {
		return mutateBool(ens)
	}
	switch v := ens.Data.(type) {
	case int:
		ens.Data = v * int(2)
	case int8:
		ens.Data = v * int8(2)
	case int16:
		ens.Data = v * int16(2)
	case int32:
		ens.Data = v * int32(2)
	case int64:
		ens.Data = v * int64(2)
	case uint:
		ens.Data = v * uint(2)
	case uint8:
		ens.Data = v * uint8(2)
	case uint16:
		ens.Data = v * uint16(2)
	case uint32:
		ens.Data = v * uint32(2)
	case uint64:
		ens.Data = v * uint64(2)
	case float32:
		ens.Data = v * float32(2.0)
	case float64:
		ens.Data = v * float64(2.0)
	default:
		return false
	}
	return true
}

// TODO: Fine-tuning is, of course, needed in the future.
func factorialMul(ens *EntForNotSB) bool {
	if ens.IsBool {
		return mutateBool(ens)
	}
	switch v := ens.Data.(type) {
	case int:
		ens.Data = v * int(3)
	case int8:
		ens.Data = v * int8(3)
	case int16:
		ens.Data = v * int16(3)
	case int32:
		ens.Data = v * int32(3)
	case int64:
		ens.Data = v * int64(3)
	case uint:
		ens.Data = v * uint(3)
	case uint8:
		ens.Data = v * uint8(3)
	case uint16:
		ens.Data = v * uint16(3)
	case uint32:
		ens.Data = v * uint32(3)
	case uint64:
		ens.Data = v * uint64(3)
	case float32:
		ens.Data = v * float32(3.0)
	case float64:
		ens.Data = v * float64(3.0)
	default:
		return false
	}
	return true
}

// TODO: Fine-tuning is, of course, needed in the future.
func factorialDiv(ens *EntForNotSB) bool {
	if ens.IsBool {
		return mutateBool(ens)
	}
	switch v := ens.Data.(type) {
	case int:
		if v == 0 {
			return false
		}
		ens.Data = v / int(3)
	case int8:
		if v == 0 {
			return false
		}
		ens.Data = v / int8(3)
	case int16:
		if v == 0 {
			return false
		}
		ens.Data = v / int16(3)
	case int32:
		if v == 0 {
			return false
		}
		ens.Data = v / int32(3)
	case int64:
		if v == 0 {
			return false
		}
		ens.Data = v / int64(3)
	case uint:
		if v == 0 {
			return false
		}
		ens.Data = v / uint(3)
	case uint8:
		if v == 0 {
			return false
		}
		ens.Data = v / uint8(3)
	case uint16:
		if v == 0 {
			return false
		}
		ens.Data = v / uint16(3)
	case uint32:
		if v == 0 {
			return false
		}
		ens.Data = v / uint32(3)
	case uint64:
		if v == 0 {
			return false
		}
		ens.Data = v / uint64(3)
	case float32:
		if v == 0.0 {
			return false
		}
		ens.Data = v / float32(3.0)
	case float64:
		if v == 0.0 {
			return false
		}
		ens.Data = v / float64(3.0)
	default:
		return false
	}
	return true
}

// TODO: Fine-tuning is, of course, needed in the future.
func interestingChange(ens *EntForNotSB) bool {
	if ens.IsBool {
		return mutateBool(ens)
	}
	switch ens.Data.(type) {
	case int:
		ens.Data = int(42)
	case int8:
		ens.Data = int8(42)
	case int16:
		ens.Data = int16(42)
	case int32:
		ens.Data = int32(42)
	case int64:
		ens.Data = int64(42)
	case uint:
		ens.Data = uint(42)
	case uint8:
		ens.Data = uint8(42)
	case uint16:
		ens.Data = uint16(42)
	case uint32:
		ens.Data = uint32(42)
	case uint64:
		ens.Data = uint64(42)
	case float32:
		ens.Data = float32(42)
	case float64:
		ens.Data = float64(42)
	default:
		return false
	}
	return true
}

// TODO: Fine-tuning is, of course, needed in the future.
func randomChange(ens *EntForNotSB) bool {
	if ens.IsBool {
		return mutateBool(ens)
	}
	r := GetPCG().Intn(10001)
	switch ens.Data.(type) {
	case int:
		ens.Data = int(r)
	case int8:
		ens.Data = int8(r)
	case int16:
		ens.Data = int16(r)
	case int32:
		ens.Data = int32(r)
	case int64:
		ens.Data = int64(r)
	case uint:
		ens.Data = uint(r)
	case uint8:
		ens.Data = uint8(r)
	case uint16:
		ens.Data = uint16(r)
	case uint32:
		ens.Data = uint32(r)
	case uint64:
		ens.Data = uint64(r)
	case float32:
		ens.Data = float32(r)
	case float64:
		ens.Data = float64(r)
	default:
		return false
	}
	return true
}

// Various helper functions.
// randomXXXFrom gradually mutates XXX and return XXX
func randomByteFrom(seedByte byte) byte {
	p := GetPCG().Float64() // 0.0 <= p < 1.0

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
		panic("logical error: default case must not be caught")
	}
}

// randomXXXFrom gradually mutates XXX and return XXX
func randomIntFrom(i interface{}) interface{} {
	p := GetPCG().Float64()

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
			i = int(GetPCG().Int())
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
			i16 = GetPCG().Int32
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
			i64 = GetPCG().Int64()
		}
		return i64

	default:
		panic("it's not supported int type")
	}
}

// randomXXXFrom gradually mutates XXX and return XXX
func randomUintFrom(u interface{}) interface{} {
	p := GetPCG().Float64()

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
			u = uint(GetPCG().Uint64())
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

	exponent := (bits >> 23) & 0xFF
	mantissa := bits & 0x7FFFFF
	sign := bits & 0x80000000
	t := uint32(UnixNano())
	manshift := 1 + (t % 7)

	var shift int8
	if exponent <= 1 {
		shift = int8(1 + int(mantissa%2))
	} else if exponent >= 0xFE {
		shift = int8(-1 - int(mantissa%2))
	} else {
		shift = int8(-2 + int(mantissa%5))
	}

	newExp := int32(exponent) + int32(shift)
	newExponent := uint32(newExp)

	newMantissa := mantissa ^ (mantissa >> manshift)

	newBits := sign | (newExponent << 23) | (newMantissa & 0x7FFFFF)

	return math.Float32frombits(newBits)
}

func randomFloat64(a float64) float64 {
	bits := math.Float64bits(a)

	exponent := (bits >> 52) & 0x7FF

	mantissa := bits & 0xFFFFFFFFFFFFF

	sign := bits & 0x8000000000000000

	t := uint64(UnixNano())
	manshift := 1 + (t % 7)

	var shift int16
	if exponent <= 1 {
		shift = int16(1 + int64(mantissa%2))
	} else if exponent >= 0x7FE {
		shift = int16(-1 - int64(mantissa%2))
	} else {
		shift = int16(-2 + int64(mantissa%5))
	}

	newExp := int64(exponent) + int64(shift)
	newExponent := uint64(newExp)

	newMantissa := mantissa ^ (mantissa >> manshift)

	newBits := sign | (newExponent << 52) | (newMantissa & 0xFFFFFFFFFFFFF)

	return math.Float64frombits(newBits)
}
