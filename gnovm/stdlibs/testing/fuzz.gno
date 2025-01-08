package testing

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

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

type (
	Runner func(*T, ...interface{})
	F      struct {
		manager   *fuzzManager
		hasher    *fuzzHasher
		logger    *fuzzLogger
		corpus    []seed
		msgs      []string // Stores log messages for reporting.
		failed    bool     // Indicates whether the fuzzing has encountered a failure.
		ff        Runner
		isFuzzed  bool
		seedCount uint
		// for cmd test
		output               []byte
		verbose              bool // Stores log messages for reporting.
		trials               uint // Number of iterations to run the fuzzing process.
		dur                  string
		name                 string
		stringByteCandidates []int
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

	if f.manager.seedType == nil {
		f.manager.seedType = types
	} else {
		if !isSliceEqual(f.manager.seedType, types) {
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
	switch arg.(type) {
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
		f.manager.inithashnumber = append(f.manager.inithashnumber, hashNumber)
		endInfo := f.manager.coordinateSeed(corp)
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
	for i, t := range f.manager.seedType {
		if t == String || t == ByteArray {
			stringByteCandidates = append(stringByteCandidates, i)
		}
	}

	f.stringByteCandidates = stringByteCandidates
	fmt.Fprintln(os.Stderr, "Run trials...")

	// format by init seeds
	for i := 0; i < len(f.corpus); i++ {
		hn := f.manager.inithashnumber[i]
		initSeed := []seed{f.manager.popInitSeedByHN(hn)}
		isEnd := f.updateMachines(initSeed)
		if isEnd {
			return
		}
	}

	for {
		parentSeeds := f.manager.popSeeds()
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
	hashNumber := f.hasher.registerCoverage2HashNumber(coverage)
	seed.hashNumber = hashNumber
	if isPanic {
		tr := testResult{
			panicOccurred: true,
			panicMessage:  panicMsg,
			prror:         err,
		}
		f.logger.addCase(seed, tr)
		f.Fail()
	}
	if err != nil {
		tr := testResult{
			panicOccurred: false,
			panicMessage:  "",
			prror:         err,
		}
		f.logger.addCase(seed, tr)
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
			coverage = getCoverageOfrunner(t, content)
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
	coverage = getCoverageOfrunner(t, content)

	return coverage, err, isPanic, panicMsg
}

// Fail marks the function as having failed bur continue execution.
func (f *F) Fail() {
	f.failed = true
}

func (f *F) handleFail() {
	fmt.Fprintln(os.Stderr, "\n--- FAIL:")

	f.logger.reportLastFailingInput()
	lastCrashCase := f.logger.getLastCrashCase()
	f.minimazeAndLogInput(lastCrashCase.rawContent)
	f.logger.reportLastCrashCase()
	minimaizedCase := f.logger.getLastCrashCase()
	hn := minimaizedCase.hashNumber
	coverage := f.hasher.hashNumber2Coverage(hn)
	if f.verbose {
		fmt.Fprintln(os.Stderr, "\n--- Trace:")
		fmt.Fprintf(os.Stderr, "%s\n", coverageToString(coverage))
	} else {
		fmt.Fprintln(os.Stderr, "\n--- Trace:")
		fmt.Fprintf(os.Stderr, "Last covered line: function \"%s\" in line %d\n", coverage[len(coverage)-1].coName, coverage[len(coverage)-1].coLine)
	}
	return
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

func (f *F) minimazeAndLogInput(seedContent []interface{}) {
	minimazalbeIdXs := []int{}
	for i, t := range f.manager.seedType {
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
	// confirm and log error
	f.simulateFF(sampleSeed)
	fmt.Fprintln(os.Stderr, "\n--- Minimaized input:")
}

func (f *F) checkProgress(content []interface{}, minimizableIdXs []int) ([]interface{}, bool, int) {
	changed := false

	for occeredPoint, idx := range minimizableIdXs {
		sOrb := content[idx]
		var candidates [][]byte
		var isString bool

		switch v := sOrb.(type) {
		case string:
			if len(v) < 1 {
				continue
			}
			isString = true
			candidates = generateCandidates([]byte(v))
		case []byte:
			if len(v) < 1 {
				continue
			}
			isString = false
			candidates = generateCandidates(v)
		default:
			panic("internal logic error")
		}

		for _, candidate := range candidates {
			tester := make([]interface{}, len(content))
			copy(tester, content)

			if isString {
				tester[idx] = string(candidate)
			} else {
				tester[idx] = candidate
			}

			if f.checkPreserve(tester).errorIsPreserved {
				changed = true
				return tester, changed, occeredPoint
			}
		}
	}
	return content, changed, 0
}

// returns a new slice with the letter of each index removed from a given byte slice.
func generateCandidates(b []byte) [][]byte {
	candidates := make([][]byte, 0, len(b))
	for i := 0; i < len(b); i++ {
		// Create a new slice that removes characters from a particular index
		candidate := append([]byte{}, b[:i]...)
		candidate = append(candidate, b[i+1:]...)
		candidates = append(candidates, candidate)
	}
	return candidates
}

type PreserveFailing struct {
	errorIsPreserved bool
	coverage         Coverage
	err              error
	isPanic          bool
	panicMsg         string
	hashNumber       hashNumber
}

func (f *F) checkPreserve(content []interface{}) PreserveFailing {
	coverage, err, isPanic, panicMsg := monitor(f.ff, content)
	hashNumber := f.hasher.registerCoverage2HashNumber(coverage)
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
	fmt.Fprintln(os.Stderr, "\n--- PASS:")
	fmt.Fprintf(os.Stderr, "Complete %d Trials\n", f.manager.inputCount)
	fmt.Fprintf(os.Stderr, "Inspected %d coverage\n", (uint(f.hasher.hashNumberCounter.counter) + uint(1)))
}

func (f *F) migrateMachines() {
	// Abstract existing state_machine
	summarizedSeeds := f.manager.summarize()
	oldHashNumbers := []hashNumber{}
	for _, seed := range summarizedSeeds {
		oldHashNumbers = append(oldHashNumbers, seed.hashNumber)
	}
	coveragesOfSeeds := []Coverage{}
	for _, hn := range oldHashNumbers {
		coveragesOfSeeds = append(coveragesOfSeeds, f.hasher.hashNumber2Coverage(uint(hn)))
	}
	// Sort by re-registering existing seed coverage to the new hash machine
	// Return the value to the seed afterwards
	// The hash number of the seed is 0,1,2... and also registers 0,1,2 and coverage on the hash machine
	f.hasher = newFuzzHasher()
	for i, cov := range coveragesOfSeeds {
		summarizedSeeds[i].hashNumber = f.hasher.registerCoverage2HashNumber(cov)
	}
	// Create and relocate a new FuzzManager
	prevInputCount := f.manager.inputCount
	substractedInputCount := int(prevInputCount) - len(summarizedSeeds)
	prevSeedType := make([]supportedType, len(f.manager.seedType))
	copy(prevSeedType, f.manager.seedType)
	prevInitHashNumber := f.manager.inithashnumber
	f.manager = newFuzzManager(f.trials, uint(substractedInputCount))
	f.manager.verbose = f.verbose
	f.manager.seedType = prevSeedType
	f.manager.inithashnumber = prevInitHashNumber
	for _, seed := range summarizedSeeds {
		f.manager.coordinateSeed(seed)
	}
}

func (f *F) updateMachines(parentSeeds []seed) bool {
	seedsMap := make(map[hashNumber]*[]seed)
	childSeeds := evolve(parentSeeds, &f.seedCount, f.stringByteCandidates)
	for _, child := range childSeeds {
		child = f.simulateFF(child)
		if f.failed {
			f.handleFail()
			return true
		}
		hn := child.hashNumber
		if _, exists := seedsMap[hn]; !exists {
			seedsMap[hn] = &[]seed{}
		}
		seeds := seedsMap[hn]
		*seeds = append(*seeds, child)
	}

	for hn, seeds := range seedsMap {
		if len(*seeds) == 1 {
			concreteEndInfo := f.manager.coordinateSeed((*seeds)[0])
			flag := f.handleEndInfo(concreteEndInfo, parentSeeds)
			if flag {
				return true
			}
		} else {
			seedArray := *(seeds)
			absNumber := uint(len(seedArray) - 2)
			absEndInfo := f.manager.coordinateAbstraction(hn, absNumber)
			flag := f.handleEndInfo(absEndInfo, parentSeeds)
			if flag {
				return true
			}
			selectedIdx1 := RandRange(0, int64(len(seedArray)))
			selectedIdx2 := RandRange(0, int64(len(seedArray)))
			if selectedIdx1 == selectedIdx2 {
				if selectedIdx1 == 0 {
					selectedIdx2 = 1
				} else {
					selectedIdx2 = selectedIdx1 - 1
				}
			}
			concreteEndInfo1 := f.manager.coordinateSeed(seedArray[selectedIdx1])
			flag = f.handleEndInfo(concreteEndInfo1, parentSeeds)
			if flag {
				return true
			}
			concreteEndInfo2 := f.manager.coordinateSeed(seedArray[selectedIdx2])
			flag = f.handleEndInfo(concreteEndInfo2, parentSeeds)
			if flag {
				return true
			}
		}
	}
	for _, p := range parentSeeds {
		f.manager.coordinateSeed(p)
	}
	return false
}

func (f *F) handleEndInfo(endInfo endInfo, parentSeeds []seed) bool {
	if endInfo.completeTrials {
		f.reportF()
		return true
	}
	if endInfo.maxedCAPACITY {
		for _, p := range parentSeeds {
			f.manager.coordinateSeed(p)
		}
		f.migrateMachines()
		return false
	}
	return false
}

func NewF(verbose bool, trials uint) *F {
	newManager := newFuzzManager(trials, 0)
	newManager.verbose = verbose
	newHasher := newFuzzHasher()
	newLogger := newFuzzLogger()
	return &F{
		verbose:   verbose,
		manager:   newManager,
		hasher:    newHasher,
		logger:    newLogger,
		isFuzzed:  false,
		seedCount: 0,
		trials:    trials,
	}
}
