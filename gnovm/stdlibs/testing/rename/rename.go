package rename

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"math"
	"math/rand/v2"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type (
	Runner func(*T, ...interface{})
	F      struct {
		fsm       *stateMachine
		fhm       *hashMachine
		corpus    []seed
		msgs      []string // Stores log messages for reporting.
		failed    bool     // Indicates whether the fuzzing has encountered a failure.
		ff        Runner
		isFuzzed  bool
		seedCount uint
		// for cmd test
		output  []byte
		verbose bool // Stores log messages for reporting.
		trials  uint // Number of iterations to run the fuzzing process.
		dur     string
		name    string
	}
)

func (f *F) Add(args ...interface{}) {
	if f.isFuzzed {
		panic("Add after Fuzz")
	}
	var values []interface{}
	var types []supportedType
	if len(args) == 0 {
		panic("zero-argument is denied")
	}
	for i := range args {
		t, e := typeof(args[i])
		if e != nil {
			panic("not supported type")
		}
		values = append(values, args[i])
		types = append(types, t)
	}

	if f.fsm.seedType == nil {
		f.fsm.seedType = types
	} else {
		if !isSliceEqual(f.fsm.seedType, types) {
			panic("added arguments not equal together")
		}
	}
	f.seedCount++
	f.corpus = append(f.corpus, seed{
		pid: f.seedCount,
		id:  f.seedCount,
		gen: 1, isCoordinated: false, content: values,
	})
}

func typeof(arg interface{}) (supportedType, error) {
	switch v := arg.(type) {
	case []byte:
		return ByteArray, nil
	case string:
		return String, nil
	case bool:
		return Bool, nil
	case byte:
		return Byte, nil
	case rune:
		return Rune, nil
	case float32:
		return Float32, nil
	case float64:
		return Float64, nil
	case int:
		return Int, nil
	case int8:
		return Int8, nil
	case int16:
		return Int16, nil
	// deduplication because int32 and rune are of the same type
	// case int32:
	//      return Int32, nil
	case int64:
		return Int64, nil
	case uint:
		return Uint, nil
	// deduplication
	// case uint8:
	//      return Uint8, nil
	case uint16:
		return Uint16, nil
	case uint32:
		return Uint32, nil
	case uint64:
		return Uint64, nil
	default:
		println("unsupported type:", v)
		return "", errors.New("unsupported type:")
	}
}

func isSliceEqual(a, b []supportedType) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (f *F) Fuzz(run Runner) {
	if !f.isFuzzed {
		f.isFuzzed = true
	} else {
		panic("fuzz called more than once")
	}

	// format machine
	f.ff = run
	for _, corp := range f.corpus {
		corp = f.simulateFF(corp)
		if f.failed {
			f.handleFail()
			return
		}

		hashNumber := corp.hashNumber
		f.fsm.inithashnumber = append(f.fsm.inithashnumber, hashNumber)
		endInfo := f.fsm.CoordinateSeed(corp)
		if endInfo.completeTrials {
			f.reportF()
			return
		}
		if endInfo.maxedCAPACITY {

			f.migrateMachines()
			continue
		}

	}

	stringByteCandidates := []int{}
	for i, t := range f.fsm.seedType {
		if t == String || t == ByteArray {
			stringByteCandidates = append(stringByteCandidates, i)
		}
	}

	f.fsm.stringByteCandidates = stringByteCandidates
	println("Run trials...")

	// format by init seeds
	for i := 0; i < len(f.corpus); i++ {
		hn := f.fsm.inithashnumber[i]
		initSeed := []seed{f.fsm.PopInitSeedByHN(hn)}
		isEnd := f.updateMachines(initSeed)
		if isEnd {
			return
		}
	}

	for {
		parentSeeds := f.fsm.PopSeeds()
		isEnd := f.updateMachines(parentSeeds)
		if isEnd {
			return
		}
	}
}

func (f *F) simulateFF(seed seed) seed {
	coverage, err, isPanic, panicMsg := monitor(f.ff, seed.content)

	// seed.hn computation
	// This completes the status change of the seed before logging
	hashNumber := f.fhm.RegisterCoverage2HashNumber(coverage)
	seed.hashNumber = hashNumber
	if isPanic {
		tr := testResult{
			panicOccurred: true,
			panicMessage:  panicMsg,
			prror:         err,
		}
		f.fsm.crashLogger.AddCase(seed, tr)
		f.Fail()
	}
	if err != nil {
		tr := testResult{
			panicOccurred: false,
			panicMessage:  "",
			prror:         err,
		}
		f.fsm.crashLogger.AddCase(seed, tr)
		f.Fail()
	}
	return seed
}

// TODO: Make sure to revise coverage here!!!
// TODO: I've hard-coded coverage according to the test results here. This will be corrected later!!
func monitor(run Runner, content []interface{}) (coverage Coverage, err error, isPanic bool, panicMsg string) {
	isPanic = false
	panicMsg = ""
	err = nil
	coverage = Coverage{}
	defer func() {
		if r := recover(); r != nil {
			t := NewT("fuzzing")
			coverage = getCoverageofrunner(t, content)
			isPanic = true
			if err, ok := r.(error); ok {
				panicMsg = err.Error()
				return
			}
			if s, ok := r.(string); ok {

				panicMsg = s
				return
			}

			panicMsg = "unknown panic"
		}
	}()
	t := NewT("fuzzing")
	// Ensuring the immutability of content
	copied := make([]interface{}, len(content))
	for i, v := range content {
		copied[i] = v
	}
	run(t, copied...)
	info := t.GetResult()
	if info.Failed {
		err = errors.New(string(info.Output))
	}
	// TODO: Modifying this function to get real coverage
	// TODO: It's just pshedo-covrage of some function
	coverage = getCoverageofrunner(t, content)

	return coverage, err, isPanic, panicMsg
}

// Fail marks the function as having failed bur continue execution.
func (f *F) Fail() {
	f.failed = true
}

func (f *F) handleFail() {
	f.Fail()
	println("\n--- FAIL:")
	log := f.fsm.crashLogger.data
	crashCase := log[len(log)-1]
	println("Found failing input", transEscapedString(crashCase.input), "at", f.fsm.inputCount, "trials")
	f.minimazeAndLogInput(crashCase.rawContent)
	log = f.fsm.crashLogger.data
	minimaizedCase := log[len(log)-1]
	println(crashCaseToString(minimaizedCase))
	hn := minimaizedCase.hashNumber
	coverage := f.fhm.HashNumber2Coverage(hn)
	if f.verbose {
		println("\n--- Trace:")
		println(coverageToString(coverage))
	} else {
		println("\n--- Trace:")
		println("Last covered line: function \""+coverage[len(coverage)-1].coName+"\" in line", coverage[len(coverage)-1].coLine)
	}
	return
}

func (f *F) minimazeAndLogInput(seedContent []interface{}) {
	minimazalbeIdXs := []int{}
	for i, t := range f.fsm.seedType {
		if t == ByteArray {
			minimazalbeIdXs = append(minimazalbeIdXs, i)
		} else if t == String {
			minimazalbeIdXs = append(minimazalbeIdXs, i)
		}
	}
	if len(minimazalbeIdXs) < 1 {
		return
	}
	sample := seedContent
	content := make([]interface{}, len(sample))
	for i, v := range sample {
		content[i] = v
	}
	// minimaze by progressive, preserveness traits of error
	for {
		progressed, isProgressed, occeredPoint := f.checkProgress(content, minimazalbeIdXs)
		if !isProgressed {
			break
		}
		content = progressed
		minimazalbeIdXs = minimazalbeIdXs[occeredPoint:]
	}
	sampleSeed := seed{
		content: content,
	}
	// re confirm error
	f.simulateFF(sampleSeed)
	println("\n--- Minimaized input:")
}

func (f *F) checkProgress(content []interface{}, minimazalbeIdXs []int) ([]interface{}, bool, int) {
	changed := false
	for _, idx := range minimazalbeIdXs {
		sOrb := content[idx]

		switch v := sOrb.(type) {
		case string:
			if len(v) < 1 {
				continue
			}
			for i := 0; i < len(v); i++ {
				b := []byte(v)
				candidate := append(b[:i], b[i+1:]...)

				tester := make([]interface{}, len(content))
				for i, v := range content {
					tester[i] = v
				}
				tester[idx] = string(candidate)
				if f.checkPreserve(tester).errorIsPreserved {

					changed = true
					return tester, changed, idx
				}
			}
		case []byte:
			if len(v) < 1 {
				continue
			}
			for i := 0; i < len(v); i++ {
				b := []byte(v)
				candidate := append(b[:i], b[i+1:]...)
				tester := make([]interface{}, len(content))
				for i, v := range content {
					tester[i] = v
				}
				tester[idx] = []byte(candidate)
				if f.checkPreserve(tester).errorIsPreserved {
					changed = true
					return tester, changed, idx
				}
			}
		default:
			panic("internal logic error")
		}
	}
	return content, changed, 0
}

type PreserveFailing struct {
	errorIsPreserved bool
	coverage         Coverage
	err              error
	isPanic          bool
	panicMsg         string
	hashNumber       HashNumber
}

func (f *F) checkPreserve(content []interface{}) PreserveFailing {
	coverage, err, isPanic, panicMsg := monitor(f.ff, content)
	hashNumber := f.fhm.RegisterCoverage2HashNumber(coverage)
	if isPanic {
		return PreserveFailing{
			errorIsPreserved: true,
			coverage:         coverage,
			err:              err,
			isPanic:          true,
			panicMsg:         panicMsg,
			hashNumber:       hashNumber,
		}
	}
	if err != nil {
		return PreserveFailing{
			errorIsPreserved: true,
			coverage:         coverage,
			err:              err,
			isPanic:          false,
			panicMsg:         "",
			hashNumber:       hashNumber,
		}
	}
	return PreserveFailing{
		errorIsPreserved: false,
	}
}

func (f *F) reportF() {
	println("\n--- PASS")
	println("Complete", f.fsm.inputCount, "Trials")
	println("Found", (uint(f.fhm.hashNumberCounter.counter) + uint(1)), "coverage")
}

func (f *F) migrateMachines() {
	println("Machine capacity is full. Start migration...")
	// Abstract existing state_machine
	summarizedSeeds := f.fsm.Summarize()
	oldHashNumbers := []HashNumber{}
	for _, seed := range summarizedSeeds {
		oldHashNumbers = append(oldHashNumbers, seed.hashNumber)
	}
	coveragesOfSeeds := []Coverage{}
	for _, hn := range oldHashNumbers {
		coveragesOfSeeds = append(coveragesOfSeeds, f.fhm.HashNumber2Coverage(uint(hn)))
	}
	// Sort by re-registering existing seed coverage to the new hash machine
	// Return the value to the seed afterwards
	// The hash number of the seed is 0,1,2... and also registers 0,1,2 and coverage on the hash machine
	f.fhm = newHashMachine()
	for i, cov := range coveragesOfSeeds {
		summarizedSeeds[i].hashNumber = f.fhm.RegisterCoverage2HashNumber(cov)
	}
	// Create and relocate a new StateMachine
	prevInputCount := f.fsm.inputCount
	substractedInputCount := int(prevInputCount) - len(summarizedSeeds)
	prevSeedType := make([]supportedType, len(f.fsm.seedType))
	copy(prevSeedType, f.fsm.seedType)
	prevInitHashNumber := f.fsm.inithashnumber
	prevStringByteCandidates := f.fsm.stringByteCandidates
	f.fsm = newStateMachine(f.trials, uint(substractedInputCount))
	f.fsm.verbose = f.verbose
	f.fsm.seedType = prevSeedType
	f.fsm.inithashnumber = prevInitHashNumber
	f.fsm.stringByteCandidates = prevStringByteCandidates
	for _, seed := range summarizedSeeds {
		f.fsm.CoordinateSeed(seed)
	}
	println("Migration completed. Resume fuzzing...")
}

func (f *F) updateMachines(parentSeeds []seed) bool {
	abstractSeedMap := make(map[HashNumber]abstractSeedInfo)
	childSeeds := evolve(parentSeeds, &f.seedCount, f.fsm.stringByteCandidates)
	for _, child := range childSeeds {
		child = f.simulateFF(child)
		if f.failed {
			f.handleFail()
			return true
		}
		hn := child.hashNumber
		equalHashNumberSeeds := abstractSeedMap[hn].seeds
		abstractNumber := abstractSeedMap[hn].abstractNumber
		if len(equalHashNumberSeeds) < 2 {
			equalHashNumberSeeds = append(equalHashNumberSeeds, child)
			abstractNumber++
		} else {
			abstractNumber++
		}
		absSeedInfo := abstractSeedInfo{
			seeds:          equalHashNumberSeeds,
			abstractNumber: abstractNumber,
		}
		abstractSeedMap[hn] = absSeedInfo
	}

	for hn, absSeedInfo := range abstractSeedMap {
		seeds := absSeedInfo.seeds
		if len(seeds) == 1 {
			concreteEndInfo := f.fsm.CoordinateSeed(seeds[0])
			flag := f.handleEndInfo(concreteEndInfo, parentSeeds)
			if flag {
				return true
			}
		} else {
			absNumber := absSeedInfo.abstractNumber - 2
			absEndInfo := f.fsm.CoordinateAbstraction(hn, absNumber)
			flag := f.handleEndInfo(absEndInfo, parentSeeds)
			if flag {
				return true
			}
			concreteEndInfo1 := f.fsm.CoordinateSeed(seeds[0])
			flag = f.handleEndInfo(concreteEndInfo1, parentSeeds)
			if flag {
				return true
			}
			concreteEndInfo2 := f.fsm.CoordinateSeed(seeds[1])
			flag = f.handleEndInfo(concreteEndInfo2, parentSeeds)
			if flag {
				return true
			}
		}

	}
	for _, p := range parentSeeds {
		f.fsm.CoordinateSeed(p)
	}
	return false
}

type abstractSeedInfo struct {
	seeds          []seed
	abstractNumber uint
}

func (f *F) handleEndInfo(endInfo endInfo, parentSeeds []seed) bool {
	if endInfo.completeTrials {
		f.reportF()
		return true
	}
	if endInfo.maxedCAPACITY {
		for _, p := range parentSeeds {
			f.fsm.CoordinateSeed(p)
		}
		f.migrateMachines()
		return false
	}
	return false
}

func NewF(verbose bool, trials uint) *F {
	newFsm := newStateMachine(trials, 0)
	newFsm.verbose = verbose
	newFhm := newHashMachine()
	return &F{
		verbose:   verbose,
		fsm:       newFsm,
		fhm:       newFhm,
		isFuzzed:  false,
		seedCount: 0,
		trials:    trials,
	}
}

func mutate(seed seed, mutationStrength int) seed {
	if len(seed.content) == 0 {
		panic("mutate logic error: content's len==0")
	}
	index := 0
	if len(seed.content) > 1 {
		index = int(RandRange(0, int64(len(seed.content))))
	}

	selected := seed.content[index]
	switch v := selected.(type) {
	case int, int8, int16, int32, int64:
		for i := 0; i < mutationStrength; i++ {
			seed.content[index] = randomIntFrom(v)
		}
	case uint, uint8, uint16, uint32, uint64:
		for i := 0; i < mutationStrength; i++ {
			seed.content[index] = randomUintFrom(v)
		}
	case float32, float64:
		for i := 0; i < mutationStrength; i++ {
			seed.content[index] = randomFloatFrom(v)
		}
	case bool:
		seed.content[index] = randomBool()
	// Cancellation due to value set issue
	// As in the comment code below, the spread in implementation does not cover all string values.
	// This is the description of the logic that deals with strings as bytes.
	// case string:
	//      runes := []rune(v)
	//      if len(runes) > 0 {
	//              runeIndex := RandRange(0, int64(len(runes)))
	//              runes[runeIndex] = randomRunefrom(runes[runeIndex])
	//      }
	//      var new_str string = string(runes)
	//      seed.Content[index] = new_str
	case string:
		bytes := []byte(v)
		if len(bytes) > 0 {
			for i := 0; i < mutationStrength; i++ {
				byteIndex := RandRange(0, int64(len(bytes)))
				bytes[byteIndex] = randomByteFrom(bytes[byteIndex])
			}
		}
		var newStr string = string(bytes)
		seed.content[index] = newStr

	case []byte:
		bytes := []byte(v)
		if len(bytes) > 0 {
			for i := 0; i < mutationStrength; i++ {
				byteIndex := RandRange(0, int64(len(bytes)))
				bytes[byteIndex] = randomByteFrom(bytes[byteIndex])
			}
		}
		var newByt []byte = []byte(bytes)
		seed.content[index] = newByt
	default:
		panic("not supported type")
	}

	return seed
}

func insertDelete(seed seed, p float64, mutationStrength int, stringByteCandidates []int) seed {
	if len(stringByteCandidates) == 0 {
		return seed
	}

	index := 0
	if len(stringByteCandidates) > 0 {

		selectedFieldidx := RandRange(0, int64(len(stringByteCandidates)))
		index = stringByteCandidates[selectedFieldidx]
	}

	selected := seed.content[index]

	switch v := selected.(type) {
	case []byte:
		bb := []byte(v)
		for i := 0; i < mutationStrength; i++ {
			l := len(bb)
			// Insert
			if GenerateRandomBool(p) {
				if l < 1 {
					var b byte = ' '
					bb = []byte{randomByteFrom(b)}
				} else {

					sample := bb[RandRange(0, int64(l))]
					bt := randomByteFrom(sample)

					pos := RandRange(0, int64(l))

					bb = append(bb, 0)

					copy(bb[pos+1:], bb[pos:])

					bb[pos] = bt
				}
			} else {
				// Del
				if l == 0 {
					return seed
				}
				pos := RandRange(0, int64(l))
				bb = append(bb[:pos], bb[pos+1:]...)
			}

		}
		var newByte []byte = bb
		seed.content[index] = newByte
	case string:
		bb := []byte(v)
		for i := 0; i < mutationStrength; i++ {
			l := len(bb)
			// Insert
			if GenerateRandomBool(p) {
				if l < 1 {
					var b byte = ' '
					bb = []byte{randomByteFrom(b)}
				} else {

					sample := bb[RandRange(0, int64(l))]
					bt := randomByteFrom(sample)

					pos := RandRange(0, int64(l))

					bb = append(bb, 0)

					copy(bb[pos+1:], bb[pos:])

					bb[pos] = bt
				}
			} else {
				if l != 0 {
					pos := RandRange(0, int64(l))
					bb = append(bb[:pos], bb[pos+1:]...)
				}
			}

		}
		var newString string = string(bb)
		seed.content[index] = newString
	default:
		println("maybe some nil string error:", v)
		panic("internal logic error")
	}
	return seed
}

// I deleted the existing fit, fitness.
// As I tried to increase the speed by integrating into AFL, I decided that it was faster to manage it with just queue, stack, and unique linked list.
// (Something gets uncomfortable if follow the afl logic I've seen and maintain that fitness management.)
// Fitness and selection logic are replaced.

// Modified existing crossover logic.
// I adjusted the number according to gen to solve the sticking problem.(a phenomenon in which they become similar as inputs increase)
// Changed to multi-intersection logic.
func twoPointCrossover(parent1, parent2 seed, seedCount *uint) (seed, seed) {
	content1 := make([]interface{}, len(parent1.content))
	for i, v := range parent1.content {
		content1[i] = v
	}
	content2 := make([]interface{}, len(parent2.content))
	for i, v := range parent2.content {
		content2[i] = v
	}

	for i := 0; i < len(parent1.content); i++ {
		switch v1 := content1[i].(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			content1[i], content2[i] = factorizationCrossover(v1, content2[i])
		case bool:
			content1[i] = v1
			content2[i] = content2[i]

		case []byte:
			byt1 := v1
			byt2, ok := parent2.content[i].([]byte)
			if !ok {
				panic("type not equal")
			}
			p1Bytes := []byte(byt1)
			p2Bytes := []byte(byt2)
			p1Len := len(p1Bytes)
			p2Len := len(p2Bytes)
			minLen := p1Len
			if p2Len < p1Len {
				minLen = p2Len
			}
			if minLen == 0 {
				maxLen := p1Len
				m := 1
				if p2Len > p1Len {
					m = 2
					maxLen = p2Len
				}
				if maxLen < 1 {
					s := ' '
					bb := byte(s)
					content1[i] = []byte([]byte{randomByteFrom(bb)})
					content2[i] = []byte([]byte{randomByteFrom(bb)})
					continue
				} else {
					if m == 1 {
						content2[i] = content1[i]
					} else {
						content1[i] = content2[i]
					}
					continue
				}
			}

			point1 := RandRange(0, int64(minLen))
			point2 := RandRange(0, int64(minLen))

			if point1 > point2 {
				point1, point2 = point2, point1
			}

			crossedByt1 := append([]byte{}, p1Bytes[:point1]...)
			crossedByt1 = append(crossedByt1, p2Bytes[point1:point2]...)
			crossedByt1 = append(crossedByt1, p1Bytes[point2:]...)

			crossedByt2 := append([]byte{}, p2Bytes[:point1]...)
			crossedByt2 = append(crossedByt2, p1Bytes[point1:point2]...)
			crossedByt2 = append(crossedByt2, p2Bytes[point2:]...)

			resultByt1 := []byte(crossedByt1)
			resultByt2 := []byte(crossedByt2)
			content1[i] = resultByt1
			content2[i] = resultByt2
		case string:
			byt1 := v1
			byt2, ok := parent2.content[i].(string)
			if !ok {
				panic("type not equal")
			}
			p1Bytes := []byte(byt1)
			p2Bytes := []byte(byt2)
			p1Len := len(p1Bytes)
			p2Len := len(p2Bytes)
			minLen := p1Len
			if p2Len < p1Len {
				minLen = p2Len
			}
			if minLen < 1 {
				maxLen := p1Len
				m := 1
				if p2Len > p1Len {
					m = 2
					maxLen = p2Len
				}
				if maxLen < 1 {
					s := ' '
					bb := byte(s)
					content1[i] = string(randomByteFrom(bb))
					content2[i] = string(randomByteFrom(bb))
					continue
				} else {
					if m == 1 {
						content2[i] = content1[i]
					} else {
						content1[i] = content2[i]
					}
					continue
				}
			}

			point1 := RandRange(0, int64(minLen))
			point2 := RandRange(0, int64(minLen))

			if point1 > point2 {
				point1, point2 = point2, point1
			}

			crossedByt1 := append([]byte{}, p1Bytes[:point1]...)
			crossedByt1 = append(crossedByt1, p2Bytes[point1:point2]...)
			crossedByt1 = append(crossedByt1, p1Bytes[point2:]...)

			crossedByt2 := append([]byte{}, p2Bytes[:point1]...)
			crossedByt2 = append(crossedByt2, p1Bytes[point1:point2]...)
			crossedByt2 = append(crossedByt2, p2Bytes[point2:]...)

			resultStr1 := string(crossedByt1)
			resultStr2 := string(crossedByt2)
			content1[i] = resultStr1
			content2[i] = resultStr2

		default:
			panic("not supported type")

		}
	}

	*seedCount++
	updatedIdv1 := seed{
		gen: parent1.gen + 1, isCoordinated: false, content: content1, result: nil,
		pid: parent1.id, id: *seedCount,
	}
	*seedCount++
	updatedIdv2 := seed{
		gen: parent2.gen + 1, isCoordinated: false, content: content2, result: nil,
		pid: parent1.id, id: *seedCount,
	}

	return updatedIdv1, updatedIdv2
}

// Returns the parents and returns the child.
// fluidized the number and degree of mutation along Gen
// Seeds received as parameters are invariant referenced within the function.
func evolve(seeds []seed, seedCount *uint, stringByteCandidates []int) []seed {
	p1 := seeds[0]
	MutationStrength := [10]int{5, 4, 3, 3, 3, 2, 2, 2, 2, 2}
	var mutationStrength int
	if int(p1.gen) > len(MutationStrength) {
		mutationStrength = 1
	} else {
		mutationStrength = MutationStrength[p1.gen-1]
	}
	MakingCount := [10]int{1920, 1440, 1200, 1024, 960, 720, 600, 480, 360, 240}
	var makingCount int
	if int(p1.gen) > len(MakingCount) {
		makingCount = 240
	} else {
		makingCount = MakingCount[p1.gen-1]
	}

	loopCount := makingCount / 2

	newGeneration := []seed{}

	if len(seeds) == 1 {

		for i := 0; i < int(makingCount); i++ {
			newContent := make([]interface{}, len(seeds[0].content))
			for i, v := range seeds[0].content {
				newContent[i] = v
			}
			*seedCount++
			newInd := seed{
				gen:           seeds[0].gen + 1,
				isCoordinated: false,
				content:       newContent,
				result:        nil,
				pid:           seeds[0].id,
				id:            *seedCount,
			}

			// 60% range
			randSeed := i % 5
			if randSeed <= 2 {
				newInd = mutate(newInd, mutationStrength)
			}
			// 40% range with 20% intersection
			if randSeed == 2 || randSeed == 3 {
				newInd = insertDelete(newInd, 0.5, mutationStrength, stringByteCandidates)
			}

			newInd.gen = seeds[0].gen + 1
			newGeneration = append(newGeneration, newInd)

		}
		return newGeneration
	}

	if len(seeds) > 3 {
		panic("not covered len")
	}

	p2 := seeds[1]

	for i := 0; i < int(loopCount); i++ {
		c1, c2 := twoPointCrossover(p1, p2, seedCount)
		newGeneration = append(newGeneration, c1)
		newGeneration = append(newGeneration, c2)

	}

	for i := range newGeneration {
		randSeed := i % 10

		// 40% range
		if randSeed <= 3 {
			newGeneration[i] = mutate(newGeneration[i], mutationStrength)
		}
		// 30%range and 10% intersection
		if randSeed >= 3 && randSeed <= 5 {
			newGeneration[i] = insertDelete(newGeneration[i], 0.5, mutationStrength, stringByteCandidates)
		}

	}

	return newGeneration
}

func randomByteFrom(seedByte byte) byte {
	p := GetSingleRand().Float64() // 0.0 <= p < 1.0

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
		case p >= 0.1 && p < 0.3:
			currentCase = (currentCase + 3) % 4
			isProgress = false
			break
		case p >= 0.3 && p < 0.7:
			break

		case p >= 0.7 && p < 0.9:
			isProgress = true
			currentCase = (currentCase + 1) % 4
		case p >= 0. && p <= 1.0:
			isProgress = true
			currentCase = (currentCase + 2) % 4
		}
	}

	if currentCase == 2 {
		p2 := GetSingleRand().Float64()
		if isProgress {
			if p2 < 0.8 {
				currentCase = (currentCase + 1) % 4
			}
		} else {
			if p2 < 0.8 {
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

	PrintUnicodeMin = 0x100
	PrintUnicodeMax = 0xD7FF

	BoundaryUnicodeMIN = 0xD800
	BoundaryUnicodeMax = 0x10FFFF

	AbsoluteMax = 0x7FFFFFFF
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
	case seedRune >= PrintUnicodeMin && seedRune <= PrintUnicodeMax:
		return 4
	case seedRune >= BoundaryUnicodeMIN && seedRune <= BoundaryUnicodeMax:
		return 5
	default:
		return 6
	}
}

func randomIntFrom(i interface{}) interface{} {
	p := GetSingleRand().Float64()

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
			i = GetSingleRand().Int()
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
			i16 = GetSingleRand().Int32()
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
			i64 = GetSingleRand().Int64()
		}
		return i64

	default:
		panic("it's not supported int type")
	}
}

func randomUintFrom(u interface{}) interface{} {
	p := GetSingleRand().Float64()

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
			u = uint(GetSingleRand().Uint64())
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

func factorizationCrossover(a interface{}, b interface{}) (interface{}, interface{}) {
	switch v1 := a.(type) {
	case int:
		v2, ok := b.(int)
		if !ok {
			panic("type not equal")
		}
		min := v1
		max := v2
		if v1 > v2 {
			min = v2
			max = v1
		}
		if min < 0 && max < 0 {
			min = max
			min = min * (-1)
		}
		if min < 0 {
			min = -1 * min
		}
		if min < 4 {
			min = 4
		}
		var newV1 int
		var newV2 int
		divisor := int(RandRange(1, int64(min)/2))
		if randomBool() {
			newV1 = v1 / divisor
			newV2 = v2 * divisor
		} else {
			newV1 = v1 * divisor
			newV2 = v2 / divisor
		}
		return newV1, newV2
	case int8:
		v2, ok := b.(int8)
		if !ok {
			panic("type not equal")
		}
		min := v1
		max := v2
		if v1 > v2 {
			min = v2
			max = v1
		}
		if min < 0 && max < 0 {
			min = max
			min = min * (-1)
		}
		if min < 0 {
			min = -1 * min
		}
		if min < 4 {
			min = 4
		}
		var newV1 int8
		var newV2 int8
		divisor := int8(RandRange(1, int64(min)/2))
		if randomBool() {
			newV1 = v1 / divisor
			newV2 = v2 * divisor
		} else {
			newV1 = v1 * divisor
			newV2 = v2 / divisor
		}
		return newV1, newV2
	case int16:
		v2, ok := b.(int16)
		if !ok {
			panic("type not equal")
		}
		min := v1
		max := v2
		if v1 > v2 {
			min = v2
			max = v1
		}
		if min < 0 && max < 0 {
			min = max
			min = min * (-1)
		}
		if min < 0 {
			min = -1 * min
		}
		if min < 4 {
			min = 4
		}
		var newV1 int16
		var newV2 int16
		divisor := int16(RandRange(1, int64(min)/2))
		if randomBool() {
			newV1 = v1 / divisor
			newV2 = v2 * divisor
		} else {
			newV1 = v1 * divisor
			newV2 = v2 / divisor
		}

		return newV1, newV2
	case int32:
		v2, ok := b.(int32)
		if !ok {
			panic("type not equal")
		}
		min := v1
		max := v2
		if v1 > v2 {
			min = v2
			max = v1
		}
		if min < 0 && max < 0 {
			min = max
			min = min * (-1)
		}
		if min < 0 {
			min = -1 * min
		}
		if min < 4 {
			min = 4
		}
		var newV1 int32
		var newV2 int32
		divisor := int32(RandRange(1, int64(min)/2))
		if randomBool() {
			newV1 = v1 / divisor
			newV2 = v2 * divisor
		} else {
			newV1 = v1 * divisor
			newV2 = v2 / divisor
		}
		return newV1, newV2

	case int64:
		v2, ok := b.(int64)
		if !ok {
			panic("type not equal")
		}
		min := v1
		max := v2
		if v1 > v2 {
			min = v2
			max = v1
		}
		if min < 0 && max < 0 {
			min = max
			min = min * (-1)
		}
		if min < 0 {
			min = -1 * min
		}
		if min < 4 {
			min = 4
		}
		var newV1 int64
		var newV2 int64
		divisor := RandRange(1, int64(min)/2)
		if randomBool() {
			newV1 = v1 / divisor
			newV2 = v2 * divisor
		} else {
			newV1 = v1 * divisor
			newV2 = v2 / divisor
		}
		return newV1, newV2

	case uint:
		v2, ok := b.(uint)
		if !ok {
			panic("type not equal")
		}
		min := v1
		if v1 > v2 {
			min = v2
		}
		if min < 4 {
			min = 4
		}
		var newV1 uint
		var newV2 uint
		divisor := uint(UintRandRange(1, uint64(min)/2))
		if randomBool() {
			newV1 = v1 / divisor
			newV2 = v2 * divisor
		} else {
			newV1 = v1 * divisor
			newV2 = v2 / divisor
		}
		return newV1, newV2
	case uint8:
		v2, ok := b.(uint8)
		if !ok {
			panic("type not equal")
		}
		min := v1
		if v1 > v2 {
			min = v2
		}
		if min < 4 {
			min = 4
		}
		var newV1 uint8
		var newV2 uint8
		divisor := uint8(UintRandRange(1, uint64(min)/2))
		if randomBool() {
			newV1 = v1 / divisor
			newV2 = v2 * divisor
		} else {
			newV1 = v1 * divisor
			newV2 = v2 / divisor
		}
		return newV1, newV2

	case uint16:
		v2, ok := b.(uint16)
		if !ok {
			panic("type not equal")
		}
		min := v1
		if v1 > v2 {
			min = v2
		}
		if min < 4 {
			min = 4
		}
		var newV1 uint16
		var newV2 uint16
		divisor := uint16(UintRandRange(1, uint64(min)/2))
		if randomBool() {
			newV1 = v1 / divisor
			newV2 = v2 * divisor
		} else {
			newV1 = v1 * divisor
			newV2 = v2 / divisor
		}
		return newV1, newV2

	case uint32:
		v2, ok := b.(uint32)
		if !ok {
			panic("type not equal")
		}
		min := v1
		if v1 > v2 {
			min = v2
		}
		if min < 4 {
			min = 4
		}
		var newV1 uint32
		var newV2 uint32
		divisor := uint32(UintRandRange(1, uint64(min)/2))
		if randomBool() {
			newV1 = v1 / divisor
			newV2 = v2 * divisor
		} else {
			newV1 = v1 * divisor
			newV2 = v2 / divisor
		}
		return newV1, newV2

	case uint64:
		v2, ok := b.(uint64)
		if !ok {
			panic("type not equal")
		}
		min := v1
		if v1 > v2 {
			min = v2
		}
		if min < 4 {
			min = 4
		}
		var newV1 uint64
		var newV2 uint64
		divisor := uint64(UintRandRange(1, uint64(min)/2))
		if randomBool() {
			newV1 = v1 / divisor
			newV2 = v2 * divisor
		} else {
			newV1 = v1 * divisor
			newV2 = v2 / divisor
		}
		return newV1, newV2

	case float32:
		v2, ok := b.(float32)
		if !ok {
			panic("type not equal")
		}
		newV1 := float32(0.7*float64(v1) + 0.3*float64(v2))
		newV2 := float32(0.3*float64(v1) + 0.7*float64(v2))
		return newV1, newV2
	case float64:
		v2, ok := b.(float64)
		if !ok {
			panic("type not equal")
		}
		newV1 := float64(0.3*float64(v1) + 0.7*float64(v2))
		newV2 := float64(0.3*float64(v1) + 0.7*float64(v2))
		return newV1, newV2
	default:
		panic("type can't be  factorization crossovered.")
	}
}

func randomFloat32(a float32) float32 {
	bits := math.Float32bits(a)

	exponent := (bits >> 23) & 0xFF
	mantissa := bits & 0x7FFFFF
	sign := bits & 0x80000000
	t := uint32(unixNano())
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

	t := uint64(time.Now().UnixNano())
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

func mock(t *T, orig ...interface{}) {
	v, ok := orig[0].(string)
	if !ok {
		panic("dont match")
	}
	rev := Reverse1(v)
	doubleRev := Reverse1(rev)
	if v != doubleRev && v == "some cond" {
		t.Errorf("Before: %q, after: %q", orig, doubleRev)
	}
	if utf8.ValidString(v) && !utf8.ValidString(rev) && v == "some cond" {
		t.Errorf("Reverse produced invalid UTF-8 string %q", rev)
	}
}

// l
// l
// l
// l
// l
// l
// l
// l
// l
func Reverse1(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

func getCoverageOfReverse1(c *Coverage, s string) string {
	r := []rune(s)
	*c = append(*c, coveredLine{coName: "Reverse1", coLine: 37})
	*c = append(*c, coveredLine{coName: "Reverse1", coLine: 38})
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		*c = append(*c, coveredLine{coName: "Reverse1", coLine: 39})
		r[i], r[j] = r[j], r[i]
		*c = append(*c, coveredLine{coName: "Reverse1", coLine: 40})
		*c = append(*c, coveredLine{coName: "Reverse1", coLine: 41})
	}
	*c = append(*c, coveredLine{coName: "Reverse1", coLine: 39})
	*c = append(*c, coveredLine{coName: "Reverse1", coLine: 42})
	return string(r)
}

func byteToHexChar(b byte) string {
	if b < 10 {
		return string('0' + b) // 0-9
	}
	return string('a' + (b - 10)) // a-f
}

func getCoverageofrunner(t *T, content []interface{}) Coverage {
	// TODO: Make sure get coverage.
	// TODO: The format is (function name, line)

	coverage := Coverage{}
	coverage = append(coverage, coveredLine{coName: "closure", coLine: 13})
	coverage = append(coverage, coveredLine{coName: "closure", coLine: 14})
	coverage = append(coverage, coveredLine{coName: "closure", coLine: 15})
	v, ok := content[0].(string)
	if !ok {
		coverage = append(coverage, coveredLine{coName: "closure", coLine: 16})
		coverage = append(coverage, coveredLine{coName: "closure", coLine: 17})

		panic("did not implement other type")
	}
	orig := string(v)
	u, ok2 := content[1].(int)
	if !ok2 {
		panic("did not implement other type")
	}
	coverage = append(coverage, coveredLine{coName: "closure", coLine: 18})
	s1 := getCoverageOfReverse1(&coverage, orig)
	coverage = append(coverage, coveredLine{coName: "closure", coLine: 19})
	s2 := getCoverageOfReverse1(&coverage, s1)

	coverage = append(coverage, coveredLine{coName: "closure", coLine: 20})
	if orig != s2 && u > 300 && u < 1000 {

		// println("orig=", orig, "doublereverse", s2)
		coverage = append(coverage, coveredLine{coName: "closure", coLine: 21})
		return coverage
	}
	coverage = append(coverage, coveredLine{coName: "closure", coLine: 22})

	coverage = append(coverage, coveredLine{coName: "closure", coLine: 23})
	if utf8.ValidString(orig) && !utf8.ValidString(s1) && u > 300 && u < 1000 {
		coverage = append(coverage, coveredLine{coName: "closure", coLine: 24})
		return coverage
	}
	coverage = append(coverage, coveredLine{coName: "closure", coLine: 25})
	coverage = append(coverage, coveredLine{coName: "closure", coLine: 26})
	return coverage
}

type coveredLine struct {
	coName string
	coLine int
}

type Coverage []coveredLine

type (
	internalHash uint64
	HashNumber   uint64
)

type hashMachine struct {
	Internal2Cov            map[internalHash]Coverage
	HashNumber2Internal     *[MaxCapacity]internalHash
	internalHash2hashNumber map[internalHash]HashNumber
	hashNumberCounter       *hashNumberCounter
}

func (hm *hashMachine) RegisterCoverage2HashNumber(coverage Coverage) HashNumber {
	internalHash := getInternalHash(coverage)
	hm.Internal2Cov[internalHash] = coverage
	hashNumber := hm.CountHashNumber(internalHash)
	hm.HashNumber2Internal[hashNumber] = internalHash
	return HashNumber(hashNumber)
}

func (hm *hashMachine) HashNumber2Coverage(hn uint) Coverage {
	internal := hm.HashNumber2Internal[hn]
	cov := hm.Internal2Cov[internal]
	return cov
}

type hashNumberCounter struct {
	counter HashNumber
}

func newHashNumberCounter(counter HashNumber) *hashNumberCounter {
	return &hashNumberCounter{
		counter: counter,
	}
}

func newHashMachine() *hashMachine {
	return &hashMachine{
		HashNumber2Internal:     &[MaxCapacity]internalHash{},
		hashNumberCounter:       newHashNumberCounter(0),
		Internal2Cov:            make(map[internalHash]Coverage),
		internalHash2hashNumber: make(map[internalHash]HashNumber),
	}
}

func (hm *hashMachine) CountHashNumber(ih internalHash) HashNumber {
	if value, exists := hm.internalHash2hashNumber[ih]; exists {
		// If the key exists, return the value
		return value
	}
	hm.internalHash2hashNumber[ih] = hm.hashNumberCounter.counter
	current := hm.hashNumberCounter.counter

	hm.hashNumberCounter.counter++
	return current
}

func coverageToBytes(coverage Coverage) []byte {
	var builder strings.Builder
	for _, line := range coverage {
		builder.WriteString(line.coName)
		builder.WriteString("|")
		builder.WriteString(intToString(line.coLine))
		builder.WriteString("|")
	}
	return []byte(builder.String())
}

func intToString(n int) string {
	if n == 0 {
		return "0"
	}

	isNegative := false
	if n < 0 {
		isNegative = true
		n = -n
	}

	var digits []byte
	for n > 0 {
		digit := n % 10
		digits = append([]byte{'0' + byte(digit)}, digits...)
		n /= 10
	}

	if isNegative {
		digits = append([]byte{'-'}, digits...)
	}

	return string(digits)
}

// calculates hash numbers without removing redundancy of line iterations in coverage
func getInternalHash(input Coverage) internalHash {
	valBytes := coverageToBytes(input)
	valArray := sha256.Sum256(valBytes)
	return internalHash(bytesToUint64(valArray))
}

// bytesToUint64 converts the first 8 bytes of a SHA256 hash to uint64
func bytesToUint64(b [32]byte) uint64 {
	return uint64(b[0])<<56 | uint64(b[1])<<48 | uint64(b[2])<<40 | uint64(b[3])<<32 |
		uint64(b[4])<<24 | uint64(b[5])<<16 | uint64(b[6])<<8 | uint64(b[7])
}

func removeDuplicates(coverage Coverage) Coverage {
	uniqueMap := make(map[coveredLine]bool)
	result := Coverage{}

	for _, line := range coverage {
		if !uniqueMap[line] {
			uniqueMap[line] = true
			result = append(result, line)
		}
	}

	return result
}

func findDifferences(cov1, cov2 Coverage) Coverage {
	diff := Coverage{}
	for _, line1 := range cov1 {
		found := false
		for _, line2 := range cov2 {
			if line1.coName == line2.coName && line1.coLine == line2.coLine {
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, line1)
		}
	}
	return diff
}

type supportedType string

const (
	ByteArray supportedType = "[]byte"
	String    supportedType = "string"
	Bool      supportedType = "bool"
	Byte      supportedType = "byte"
	Rune      supportedType = "rune"
	Float32   supportedType = "float32"
	Float64   supportedType = "float64"
	Int       supportedType = "int"
	Int8      supportedType = "int8"
	Int16     supportedType = "int16"
	Int32     supportedType = "int32"
	Int64     supportedType = "int64"
	Uint      supportedType = "uint"
	Uint8     supportedType = "uint8"
	Uint16    supportedType = "uint16"
	Uint32    supportedType = "uint32"
	Uint64    supportedType = "uint64"
)

type seed struct {
	pid           uint
	id            uint
	gen           uint
	isCoordinated bool
	hashNumber    HashNumber
	content       []interface{}
	result        interface{}
}

const (
	// Effective when path kinds(coverage) is in range 300,000~1,000,000.
	// and effiective when iters is less than 100_000_000
	// More than that is possible, but less effective.
	MaxCapacity     uint64 = 100_000
	PartialCapacity        = MaxCapacity / 5
)

type stateMachine struct {
	trials     uint
	inputCount uint

	allCoverage     Coverage // []coveredline
	coveredCoverage Coverage // []coveredline

	// for f.Add in f.Fuzz
	seedType             []supportedType
	inithashnumber       []HashNumber
	stringByteCandidates []int

	// for f.Fuzz
	priorityCache          uint
	hashNumber2Seeds       *[MaxCapacity]*seedQueue
	hashNumber2Priority    *[MaxCapacity]uint
	priority2HashNumberULL *[MaxCapacity]*uniquesLinkedList

	crashLogger           *crashLogger
	verbose               bool
	inspectingHashNumbers uint
}

func newStateMachine(trials uint, inputCount uint) *stateMachine {
	settedAllULL := func() *[MaxCapacity]*uniquesLinkedList {
		arr := &[MaxCapacity]*uniquesLinkedList{}
		for i := range arr {
			arr[i] = newUniquesLinkedList()
		}
		return arr
	}()
	return &stateMachine{
		inputCount:      inputCount,
		trials:          trials,
		allCoverage:     Coverage{{coName: "nil", coLine: 0}},
		coveredCoverage: Coverage{{coName: "nil", coLine: 0}},

		priorityCache: 1,

		hashNumber2Priority:    &[MaxCapacity]uint{},
		hashNumber2Seeds:       &[MaxCapacity]*seedQueue{},
		priority2HashNumberULL: settedAllULL,

		inspectingHashNumbers: 0,
		crashLogger:           newCrashLogger(),
	}
}

type endInfo struct {
	completeTrials bool
	maxedCAPACITY  bool
}

func (sm *stateMachine) CoordinateAbstraction(hn HashNumber, abstractNumber uint) endInfo {
	sm.inputCount = sm.inputCount + abstractNumber
	// Get current priorty
	// 1. Prior increases discontinuously
	// 2. Hn must increase continuously
	sm.hashNumber2Priority[hn] = sm.hashNumber2Priority[hn] + abstractNumber
	nextPriority := sm.hashNumber2Priority[hn]
	nextHashNumber := hn
	if sm.verbose {
		if (sm.inputCount-abstractNumber)%2000 > sm.inputCount%2000 {
			println(sm.inputCount, "times runned: inspecting", sm.inspectingHashNumbers, "coverages")
		}
	} else {
		if (sm.inputCount-abstractNumber)%(sm.trials/5) > sm.inputCount%(sm.trials/5) {
			println(sm.inputCount, "times runned: inspecting", sm.inspectingHashNumbers, "coverages")
		}
	}
	if sm.inputCount >= sm.trials {
		return endInfo{
			completeTrials: true,
			maxedCAPACITY:  false,
		}
	} else if uint64(nextPriority) >= (MaxCapacity-1) || uint64(nextHashNumber) >= (MaxCapacity-1) {
		// Case 1: Priority reaches MAX (and protect overflow)
		// Case 2 HashNumber reaches Max (and protect overflow)
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

func (sm *stateMachine) CoordinateSeed(seed seed) endInfo {
	if seed.isCoordinated {
		sm.hashNumber2Seeds[seed.hashNumber].Enqueue(seed)
		return endInfo{
			completeTrials: false,
			maxedCAPACITY:  false,
		}
	}
	hn := seed.hashNumber
	seed.isCoordinated = true
	sm.inputCount++

	if sm.hashNumber2Seeds[hn] == nil {
		sm.inspectingHashNumbers++
		sm.hashNumber2Seeds[hn] = newSeedQueue()
	}
	sm.hashNumber2Seeds[hn].Enqueue(seed)
	oldPriority := sm.hashNumber2Priority[hn]
	sm.hashNumber2Priority[hn]++
	updatedPriority := sm.hashNumber2Priority[hn]

	if updatedPriority == 1 {
		sm.priorityCache = 1
		sm.priority2HashNumberULL[updatedPriority].Append(uint(hn))
	} else {
		sm.priority2HashNumberULL[oldPriority].Delete(uint(hn))
		sm.priority2HashNumberULL[updatedPriority].Append(uint(hn))
	}
	if sm.verbose {
		if sm.inputCount%2000 == 0 {
			println(sm.inputCount, "times runned: inspecting", sm.inspectingHashNumbers, "coverages")
		}
	} else {
		if sm.inputCount%(sm.trials/5) == 0 {
			println(sm.inputCount, "times runned: inspecting", sm.inspectingHashNumbers, "coverages")
		}
	}

	if sm.inputCount >= sm.trials {
		return endInfo{
			completeTrials: true,
			maxedCAPACITY:  false,
		}
	} else if uint64(updatedPriority) >= (MaxCapacity-1) || uint64(hn) >= (MaxCapacity-1) {
		// Case 1: Priority reaches MAX
		// Case 2 HashNumber reaches Max
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

func (sm *stateMachine) PopInitSeedByHN(hn HashNumber) seed {
	popedSeed, isOnce := sm.hashNumber2Seeds[hn].Dequeue()
	if !isOnce {
		panic("logical internal error: it must has more than one seed")
	}
	return popedSeed
}

func (sm *stateMachine) PopSeeds() []seed {
	var highestProrityHashNumber uint

	for {

		hn, isExist := sm.priority2HashNumberULL[sm.priorityCache].Peek()
		if !isExist {
			sm.priorityCache++

			continue
		}

		highestProrityHashNumber = hn
		break
	}

	popedSeed1, isOnce := sm.hashNumber2Seeds[highestProrityHashNumber].Dequeue()
	if !isOnce {
		panic("logical internal error: it must has more than one seed")
	}
	peekedSeed2, err := sm.hashNumber2Seeds[highestProrityHashNumber].Peek()
	if err != nil {
		return []seed{popedSeed1}
	}
	pid1 := popedSeed1.pid
	pid2 := peekedSeed2.pid
	if pid1 == pid2 {
		popedSeed2, _ := sm.hashNumber2Seeds[highestProrityHashNumber].Dequeue()
		return []seed{popedSeed1, popedSeed2}
	} else {
		return []seed{popedSeed1}
	}
}

type priorytyAndHashNumber struct {
	priority   uint
	hashNumber int
}
type priorytyHashNumberList []priorytyAndHashNumber

func (phl priorytyHashNumberList) Len() int           { return len(phl) }
func (phl priorytyHashNumberList) Less(i, j int) bool { return phl[i].priority < phl[j].priority }
func (phl priorytyHashNumberList) Swap(i, j int)      { phl[i], phl[j] = phl[j], phl[i] }

type seedAndPrioryty struct {
	seed     seed
	priority uint
}

func (sm *stateMachine) Summarize() []seed {
	collectedPriorytyAndHN := priorytyHashNumberList{}
	for i, e := range sm.hashNumber2Priority {

		if e == 0 {
			break
		}
		collectedPriorytyAndHN = append(collectedPriorytyAndHN, priorytyAndHashNumber{
			priority:   e,
			hashNumber: i,
		})
		// Exit when slice length exceeds Partial_Capacity
		if len(collectedPriorytyAndHN) >= int(PartialCapacity) {
			break
		}
	}

	sort.Sort(collectedPriorytyAndHN)

	sampledSeedAndPrioryty := []seedAndPrioryty{}

	for _, pair := range collectedPriorytyAndHN {
		hn := pair.hashNumber
		seed := sm.hashNumber2Seeds[hn].data[0]
		priority := pair.priority
		sampledSeedAndPrioryty = append(sampledSeedAndPrioryty, seedAndPrioryty{seed, priority})
	}

	total := len(sampledSeedAndPrioryty)

	summarizedSeeds := []seed{}

	for i, seedAndPriority := range sampledSeedAndPrioryty {

		rank := i + 1

		seed := seedAndPriority.seed
		seed.pid = 0
		seed.id = uint(i)
		seed.gen = allocateGen(seedAndPriority.priority, rank, total)
		seed.isCoordinated = false

		summarizedSeeds = append(summarizedSeeds, seed)
	}
	return summarizedSeeds
}

func allocateGen(priority uint, rank int, total int) uint {
	var bigger uint

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
	c1CutLine := uint(float64(MaxCapacity) * 0.008)
	c2CutLine := uint(float64(MaxCapacity) * 0.014)
	c3CutLine := uint(float64(MaxCapacity) * 0.02)
	c4CutLine := uint(float64(MaxCapacity) * 0.03)
	c5CutLine := uint(float64(MaxCapacity) * 0.08)
	c6CutLine := uint(float64(MaxCapacity) * 0.1)
	c7CutLine := uint(float64(MaxCapacity) * 0.15)
	c8CutLine := uint(float64(MaxCapacity) * 0.30)
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

type testResult struct {
	panicOccurred bool
	panicMessage  string
	prror         error
}

type crashCase struct {
	hashNumber   uint
	input        string
	rawContent   []interface{}
	inputCount   uint
	isPanic      bool
	panicMessage string
	errorMsg     string
	timestamp    time.Time
}

// crash logger
type crashLogger struct {
	data []crashCase
}

func newCrashLogger() *crashLogger {
	return &crashLogger{
		data: make([]crashCase, 0),
	}
}

func (c *crashLogger) AddCase(seed seed, r testResult) {
	var crashcase crashCase
	if r.prror == nil {
		crashcase = crashCase{
			hashNumber:   uint(seed.hashNumber),
			input:        contentToString(seed.content),
			rawContent:   seed.content,
			inputCount:   seed.id,
			isPanic:      r.panicOccurred,
			panicMessage: r.panicMessage,
			errorMsg:     "",
			timestamp:    time.Now(),
		}
	} else {
		crashcase = crashCase{
			hashNumber:   uint(seed.hashNumber),
			input:        contentToString(seed.content),
			rawContent:   seed.content,
			inputCount:   seed.id,
			isPanic:      r.panicOccurred,
			panicMessage: r.panicMessage,
			errorMsg:     strings.TrimSpace(r.prror.Error()),
			timestamp:    time.Now(),
		}
	}
	c.data = append(c.data, crashcase)
}

func (c *crashLogger) GetCase(index int) (crashCase, error) {
	if index < 0 || index >= len(c.data) {
		return crashCase{}, errors.New("index out of bounds")
	}
	return c.data[index], nil
}

func (c *crashLogger) RemoveCase(index int) error {
	if index < 0 || index >= len(c.data) {
		return errors.New("index out of bounds")
	}
	c.data = append(c.data[:index], c.data[index+1:]...)
	return nil
}

func (c *crashLogger) ListCases() []crashCase {
	return c.data
}

func (c *crashLogger) Size() int {
	return len(c.data)
}

func (c *crashLogger) ClearLog() {
	c.data = make([]crashCase, 0)
}

// data structures
type seedQueue struct {
	data []seed
}

func newSeedQueue() *seedQueue {
	return &seedQueue{
		data: make([]seed, 0),
	}
}

func (q *seedQueue) Enqueue(seed seed) {
	q.data = append(q.data, seed)
}

func (q *seedQueue) Dequeue() (seed, bool) {
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

func (q *seedQueue) Peek() (seed, error) {
	if len(q.data) == 0 {
		return seed{}, errors.New("queue is empty")
	}
	return q.data[0], nil
}

func (q *seedQueue) Select() seed

func (q *seedQueue) IsEmpty() bool {
	return len(q.data) == 0
}

func (q *seedQueue) Size() int {
	return len(q.data)
}

func (q *seedQueue) Display() {
	if len(q.data) == 0 {
		println("Queue is empty")
		return
	}
	println("Queue seeds:")
	for i, seed := range q.data {
		println("[", i, "]: {content:", seed.content, "}")
	}
}

// ! Data must be unique!!!
// ! It's a list that reduces time complexity by using it!!
// node: node in single connection list
type node struct {
	data uint
	next *node
}

// uniquesLinkedList: Data Structure for Singly Linked List + O(1) Insertion/Deletion
type uniquesLinkedList struct {
	head      *node
	tail      *node
	nodeMap   map[uint]*node  // data -> node
	parentMap map[*node]*node // node -> parent node
}

func newUniquesLinkedList() *uniquesLinkedList {
	return &uniquesLinkedList{
		head:      nil,
		tail:      nil,
		nodeMap:   make(map[uint]*node),
		parentMap: make(map[*node]*node),
	}
}

// -------------------------------------------------------
// 1) Add a new node to O(1) at the end of the Appendix list
// -------------------------------------------------------
func (ll *uniquesLinkedList) Append(data uint) {
	newNode := &node{data: data, next: nil}

	if ll.head == nil {
		ll.head = newNode
		ll.tail = newNode
		ll.nodeMap[data] = newNode
		ll.parentMap[newNode] = nil
		return
	}

	ll.tail.next = newNode
	ll.parentMap[newNode] = ll.tail
	ll.tail = newNode
	ll.nodeMap[data] = newNode
}

// -------------------------------------------------------
// 2) Prepend: Add a new node to O(1) in the head of the list
// -------------------------------------------------------
func (ll *uniquesLinkedList) Prepend(data uint) {
	newNode := &node{data: data}

	if ll.head == nil {
		ll.head = newNode
		ll.tail = newNode
		ll.nodeMap[data] = newNode
		ll.parentMap[newNode] = nil
		return
	}

	newNode.next = ll.head

	ll.parentMap[ll.head] = newNode

	ll.head = newNode

	ll.nodeMap[data] = newNode
	ll.parentMap[newNode] = nil
}

// -------------------------------------------------------
// 3) Delete: Delete node with data == value to O(1)
// -------------------------------------------------------
func (ll *uniquesLinkedList) Delete(value uint) {
	targetNode, ok := ll.nodeMap[value]
	if !ok {
		return
	}

	if targetNode == ll.head {
		ll.head = ll.head.next
		if ll.head == nil {
			ll.tail = nil
		} else {
			ll.parentMap[ll.head] = nil
		}
		delete(ll.nodeMap, value)
		delete(ll.parentMap, targetNode)
		return
	}

	parent := ll.parentMap[targetNode]
	if parent == nil {
		return
	}
	parent.next = targetNode.next

	if targetNode == ll.tail {
		ll.tail = parent
	} else {
		ll.parentMap[targetNode.next] = parent
	}

	delete(ll.nodeMap, value)
	delete(ll.parentMap, targetNode)
}

// -------------------------------------------------------
// 4) DeleteNode: Delete directly to O(1) with node pointer
// -------------------------------------------------------
func (ll *uniquesLinkedList) DeleteNode(node *node) {
	if node == nil {
		return
	}

	if node == ll.head {
		ll.head = ll.head.next
		if ll.head == nil {
			ll.tail = nil
		} else {
			ll.parentMap[ll.head] = nil
		}
		delete(ll.nodeMap, node.data)
		delete(ll.parentMap, node)
		return
	}

	parent := ll.parentMap[node]
	if parent == nil {
		return
	}
	parent.next = node.next

	if node == ll.tail {
		ll.tail = parent
	} else {
		ll.parentMap[node.next] = parent
	}

	delete(ll.nodeMap, node.data)
	delete(ll.parentMap, node)
}

func (ll *uniquesLinkedList) SearchNode(value uint) *node {
	return ll.nodeMap[value]
}

func (ll *uniquesLinkedList) Display() {
	current := ll.head
	for current != nil {
		println("%d -> ", current.data)
		current = current.next
	}
	println("nil")
}

func (ll *uniquesLinkedList) IsEmpty() bool {
	return ll.head == nil
}

func (ll *uniquesLinkedList) Peek() (uint, bool) {
	if ll.head == nil {
		return 0, false
	}
	return ll.head.data, true
}

// token distinguishes whether it is "normal Unicode" or "broken bytes".
type token struct {
	Data  []byte // Actual bytes of the token
	Valid bool   // If true, characters that have successfully decoded UTF-8
}

// tokenizeString: Decode string s as UTF-8 as much as possible, keep broken bytes separate
func tokenizeString(s string) []token {
	var tokens []token
	b := []byte(s)
	i := 0
	for i < len(b) {
		r, size := utf8.DecodeRune(b[i:])
		switch {
		case r == utf8.RuneError && size == 1:
			tokens = append(tokens, token{
				Data:  []byte{b[i]},
				Valid: false,
			})
			i++
		default:
			tokens = append(tokens, token{
				Data:  b[i : i+size],
				Valid: true,
			})
			i += size
		}
	}
	return tokens
}

func rebuildString(tokens []token) string {
	var buf bytes.Buffer
	for _, t := range tokens {
		buf.Write(t.Data)
	}
	return buf.String()
}

func rebuildEscaped(tokens []token) string {
	var result []byte
	for _, t := range tokens {
		if t.Valid {
			result = append(result, t.Data...)
		} else {
			for _, b := range t.Data {
				result = append(result, []byte("\\x")...)
				hex := byteToHex(b)
				result = append(result, hex...)
			}
		}
	}
	return string(result)
}

func byteToHex(b byte) []byte {
	const hexdigits = "0123456789abcdef"
	hi := hexdigits[b>>4]
	lo := hexdigits[b&0x0F]
	return []byte{hi, lo}
}

// Recover what the escape characters break when they are printed. (e.g."�" -> "\xeb")
func transEscapedString(s string) string {
	toks := tokenizeString(s)
	escaped := rebuildEscaped(toks)
	return escaped
}

func uintToString(v uint) string {
	return strconv.Itoa(int(v))
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func sliceToString(slice []uint) string {
	if len(slice) == 0 {
		return "[]"
	}

	var sb strings.Builder
	sb.WriteString("[")
	for i, val := range slice {
		sb.WriteString(uintToString(val))
		if i < len(slice)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString("]")
	return sb.String()
}

func interfaceToString(i interface{}) string {
	switch v := i.(type) {
	case nil:
		return "nil"
	case string:
		return strconv.Quote(v)
	case int, int32, int64:
		return strconv.FormatInt(int64(v.(int)), 10)
	case uint, uint32, uint64:
		return strconv.FormatUint(uint64(v.(uint)), 10)
	case float32, float64:
		return strconv.FormatFloat(v.(float64), 'f', -1, 64)
	case bool:
		if v {
			return "true"
		}
		return "false"
	default:
		return "unknown"
	}
}

func interfacesliceToString(slice []interface{}) string {
	var sb strings.Builder
	sb.WriteString("[")
	for i, elem := range slice {
		sb.WriteString(interfaceToString(elem))
		if i < len(slice)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString("]")
	return sb.String()
}

func coverageToString(coverage Coverage) string {
	var sb strings.Builder
	sb.WriteString("[")
	count := 0
	for i, line := range coverage {
		sb.WriteString("{co_name: ")
		sb.WriteString(line.coName)
		sb.WriteString(", co_line: ")
		sb.WriteString(strconv.Itoa(line.coLine))
		sb.WriteString("}")
		count++
		if i < len(coverage)-1 {
			if count%3 == 0 {
				sb.WriteString(", \n")
			} else {
				sb.WriteString(", ")
			}
		}
	}
	sb.WriteString("]")
	return sb.String()
}

func contentToString(content []interface{}) string {
	var result strings.Builder

	result.WriteString("[")
	for i, elem := range content {
		switch v := elem.(type) {
		case string:
			result.WriteString("\"" + v + "\"")
		case int:

			result.WriteString(strconv.Itoa(v))
		case int8:

			result.WriteString(strconv.FormatInt(int64(v), 10))
		case int16:

			result.WriteString(strconv.FormatInt(int64(v), 10))
		case int32:

			result.WriteString(strconv.FormatInt(int64(v), 10))
		case int64:

			result.WriteString(strconv.FormatInt(v, 10))
		case uint:

			result.WriteString(strconv.FormatUint(uint64(v), 10))
		case uint8:

			result.WriteString(strconv.FormatUint(uint64(v), 10))
		case uint16:

			result.WriteString(strconv.FormatUint(uint64(v), 10))
		case uint32:

			result.WriteString(strconv.FormatUint(uint64(v), 10))
		case uint64:

			result.WriteString(strconv.FormatUint(v, 10))
		case float32:

			result.WriteString(strconv.FormatFloat(float64(v), 'f', -1, 32))
		case float64:

			result.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
		case []byte:

			result.WriteString("\"" + string(v) + "\"")
		case bool:

			if v {
				result.WriteString("true")
			} else {
				result.WriteString("false")
			}
		default:

			result.WriteString("unknown")
		}

		if i < len(content)-1 {
			result.WriteString(", ")
		}
	}
	result.WriteString("]")

	return result.String()
}

func crashCaseToString(c crashCase) string {
	var sb strings.Builder

	sb.WriteString("Input: ")
	sb.WriteString(transEscapedString(c.input))
	if c.isPanic {
		sb.WriteString("\nPanic/Error: Panic ")
		sb.WriteString("\nPanicMessage: \"")
		sb.WriteString(c.panicMessage)
	} else {
		sb.WriteString("\nPanic/Error: Error ")
		sb.WriteString("\nErrorMessage: ")
		sb.WriteString("\"" + c.errorMsg + "\"")
	}

	return sb.String()
}

// Internal state for the random number generator.
var x uint64 = 42

func ExUnixNano() int64 {
	return unixNano()
}

var (
	singleRand *rand.Rand
	isCalled   bool
)

type CustomSource struct {
	seed uint64
}

func (cs *CustomSource) Uint64() uint64 {
	cs.seed ^= cs.seed >> 12
	cs.seed ^= cs.seed << 25
	cs.seed ^= cs.seed >> 27
	return cs.seed * 2685821657736338717
}

func NewCustomSource(seed int64) *CustomSource {
	return &CustomSource{
		seed: uint64(seed),
	}
}

// used a single tone pattern.
// GetGlobalRand: Assume to be a single high-routine environment
func GetSingleRand() *rand.Rand {
	if !isCalled {
		isCalled = true
		seed := unixNano()
		source := NewCustomSource(seed)
		singleRand = rand.New(source)

	}
	return singleRand
}

// UniformRand generates a uniformly distributed random number.
// It uses the linear congrential generator method to produce the random number.
// and the result is in the range from 0 to m-1. here, m is 32768.
// To produce random number in [0, m-1], repeat this method as many times as needed.
// [1] https://en.wikipedia.org/wiki/Linear_congruential_generator
func UniformRand() uint64 {
	var a uint64 = 950213
	var c uint64 = 12345
	var m uint64 = 32768
	x = x*a + c
	return (x >> 16) % m
}

// _srand function sets the seed for the random number generator.
// This function provides an initial starting point for the sequence of random numbers.

func _srand() {
	r := GetSingleRand()
	x = uint64(r.Uint64())
}

// nrand function generates a number approximating a normal distribution[1].
// It uses the Central Limit Theorem[2] by summing multiple uniformly distributed random numbers
// to approximate a normal distribution.
//
// y = Sum(k=1, K) (x_k - K/2) / sqrt(K/12)
//
// Here, K is some integer ans x_k are uniformly distributed numbers,
// even for K as small as 10, the approximation is quite good.
// [1] https://en.wikipedia.org/wiki/Normal_distribution
// [2] https://en.wikipedia.org/wiki/Central_limit_theorem

func nrand() float64 {
	r := GetSingleRand()
	danger := r.NormFloat64()
	scaled := danger / 3
	if scaled < -1 {
		return -1
	} else if scaled > 1 {
		return 1
	}
	return scaled
}

// randRange generates a random integer between min and max (inclusive->exclusive: changend).
// ? Random question. Why did you process this intrusive?
// First of all, I changed to Exclusive
// This function leverages the UniformRand function to generate a random number in a specified range.
// Note: max should be greater than min.
func RandRange(start, end int64) int64 {
	if start >= end {
		panic("start >= end ")
	}
	r := GetSingleRand()
	randNum := r.Int64()
	hashedNum := (randNum) % (int64(end - start))
	result := int64(start) + hashedNum

	return result
}

func UintRandRange(start, end uint64) uint64 {
	if start >= end {
		panic("start >= end ")
	}
	r := GetSingleRand()
	randNum := r.Uint64()
	hashedNum := (randNum) % (uint64(end - start))
	result := uint64(start) + hashedNum

	return result
}

func RandInt64() int64 {
	r := GetSingleRand()
	randNum := r.Int64()
	return randNum
}

func RandUint64() uint64 {
	r := GetSingleRand()
	randNum := r.Uint64()
	return randNum
}

func RandUint32() uint32 {
	r := GetSingleRand()
	randNum := r.Uint32()
	return randNum
}

func GenerateRandomBool(bias float64) bool {
	// Modify to use fuzz's random function for generating boolean with bias
	if bias < 0 || bias > 1 {
		panic("bias should be in the range [0, 1]")
	}
	// Convert fuzz's normalized range random float [-1, 1] to [0, 1]
	res := (nrand() + 1) / 2
	return res > bias
}

func UniformRandomBool(probability float64) bool {
	if probability < 0.0 || probability > 1.0 {
		panic("Probability must be between 0.0 and 1.0")
	}
	r := GetSingleRand()
	return r.Float64() < probability
}
