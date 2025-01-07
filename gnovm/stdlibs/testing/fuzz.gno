package testing

import (
	"errors"
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
		endInfo := f.fsm.coordinateSeed(corp)
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
		initSeed := []seed{f.fsm.popInitSeedByHN(hn)}
		isEnd := f.updateMachines(initSeed)
		if isEnd {
			return
		}
	}

	for {
		parentSeeds := f.fsm.popSeeds()
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
	hashNumber := f.fhm.registerCoverage2HashNumber(coverage)
	seed.hashNumber = hashNumber
	if isPanic {
		tr := testResult{
			panicOccurred: true,
			panicMessage:  panicMsg,
			prror:         err,
		}
		f.fsm.crashLogger.addCase(seed, tr)
		f.Fail()
	}
	if err != nil {
		tr := testResult{
			panicOccurred: false,
			panicMessage:  "",
			prror:         err,
		}
		f.fsm.crashLogger.addCase(seed, tr)
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
	coverage := f.fhm.hashNumber2Coverage(hn)
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
	hashNumber := f.fhm.registerCoverage2HashNumber(coverage)
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
	summarizedSeeds := f.fsm.summarize()
	oldHashNumbers := []HashNumber{}
	for _, seed := range summarizedSeeds {
		oldHashNumbers = append(oldHashNumbers, seed.hashNumber)
	}
	coveragesOfSeeds := []Coverage{}
	for _, hn := range oldHashNumbers {
		coveragesOfSeeds = append(coveragesOfSeeds, f.fhm.hashNumber2Coverage(uint(hn)))
	}
	// Sort by re-registering existing seed coverage to the new hash machine
	// Return the value to the seed afterwards
	// The hash number of the seed is 0,1,2... and also registers 0,1,2 and coverage on the hash machine
	f.fhm = newHashMachine()
	for i, cov := range coveragesOfSeeds {
		summarizedSeeds[i].hashNumber = f.fhm.registerCoverage2HashNumber(cov)
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
		f.fsm.coordinateSeed(seed)
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
			concreteEndInfo := f.fsm.coordinateSeed(seeds[0])
			flag := f.handleEndInfo(concreteEndInfo, parentSeeds)
			if flag {
				return true
			}
		} else {
			absNumber := absSeedInfo.abstractNumber - 2
			absEndInfo := f.fsm.coordinateAbstraction(hn, absNumber)
			flag := f.handleEndInfo(absEndInfo, parentSeeds)
			if flag {
				return true
			}
			concreteEndInfo1 := f.fsm.coordinateSeed(seeds[0])
			flag = f.handleEndInfo(concreteEndInfo1, parentSeeds)
			if flag {
				return true
			}
			concreteEndInfo2 := f.fsm.coordinateSeed(seeds[1])
			flag = f.handleEndInfo(concreteEndInfo2, parentSeeds)
			if flag {
				return true
			}
		}

	}
	for _, p := range parentSeeds {
		f.fsm.coordinateSeed(p)
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
			f.fsm.coordinateSeed(p)
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
