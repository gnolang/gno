package testing

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

type (
	Runner func(*T, ...interface{})
	F      struct {
		fsm       *StateMachine
		fhm       *HashMachine
		corpus    []Seed
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

func New_F(verbose bool, trials uint) *F {
	println("creating fuzzing enviroment...")
	new_fsm := New_StateMachine(trials, 0)
	new_fsm.verbose = verbose
	new_fhm := New_HashMachine()
	return &F{
		verbose:   verbose,
		fsm:       new_fsm,
		fhm:       new_fhm,
		isFuzzed:  false,
		seedCount: 0,
		trials:    trials,
	}
}

func typeof(arg interface{}) (SupportedType, error) {
	switch v := arg.(type) {
	case []byte:
		return Byte_Array, nil
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
	// 	return Int32, nil
	case int64:
		return Int64, nil
	case uint:
		return Uint, nil
	// deduplication
	// case uint8:
	// 	return Uint8, nil
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

func (f *F) Add(args ...interface{}) {
	if f.isFuzzed {
		panic("Add after Fuzz")
	}
	var values []interface{}
	var types []SupportedType
	if len(args) == 0 {
		panic("no-argument is denied")
	}
	for i := range args {
		t, e := typeof(args[i])
		if e != nil {
			panic("not supported type")
		}
		values = append(values, args[i])
		types = append(types, t)
	}

	if f.fsm.SeedType == nil {
		f.fsm.SeedType = types
	} else {
		if !isSliceEqual(f.fsm.SeedType, types) {
			panic("added arguments not equal together")
		}
	}
	f.seedCount++
	f.corpus = append(f.corpus, Seed{
		Pid: f.seedCount,
		Id:  f.seedCount,
		Gen: 1, IsCoordinated: false, Content: values,
	})
}

func (f *F) Fuzz(run Runner) {
	println("----------------------------- Fuzz Start --------------------------------")
	if !f.isFuzzed {
		f.isFuzzed = true
	} else {
		panic("fuzz called more than once")
	}

	// format machine
	f.ff = run
	for i, corp := range f.corpus {
		println("Adding corpus", i+1)
		corp = f.simulateFF(corp)
		if f.failed {
			f.handleFail()
			return
		}

		hashNumber := corp.HashNumber
		f.fsm.init_hashnumber = append(f.fsm.init_hashnumber, hashNumber)
		endInfo := f.fsm.CoordinateSeed(corp)
		if endInfo.Complete_Trials {
			f.reportF()
			return
		}
		if endInfo.MAXed_CAPACITY {

			f.migrateMachines()
			continue
		}

	}

	string_byte_candidates := []int{}
	for i, t := range f.fsm.SeedType {
		if t == String || t == Byte_Array {
			string_byte_candidates = append(string_byte_candidates, i)
		}
	}

	f.fsm.string_byte_candidates = string_byte_candidates
	println("Run trials...")

	// format by init seeds
	for i := 0; i < len(f.corpus); i++ {
		hn := f.fsm.init_hashnumber[i]
		initSeed := []Seed{f.fsm.PopInitSeedByHN(hn)}
		is_end := f.updateMachines(initSeed)
		if is_end {
			return
		}
	}

	for {
		parentSeeds := f.fsm.PopSeeds()
		is_end := f.updateMachines(parentSeeds)
		if is_end {
			return
		}
	}
}

type abstractSeedInfo struct {
	seeds          []Seed
	abstractNumber uint
}

func (f *F) handleEndInfo(endInfo EndInfo, parentSeeds []Seed) bool {
	if endInfo.Complete_Trials {
		f.reportF()
		return true
	}
	if endInfo.MAXed_CAPACITY {
		for _, p := range parentSeeds {
			f.fsm.CoordinateSeed(p)
		}
		f.migrateMachines()
		return false
	}
	return false
}

func (f *F) updateMachines(parentSeeds []Seed) bool {
	abstractSeedMap := make(map[HashNumber]abstractSeedInfo)
	childSeeds := Evolve(parentSeeds, &f.seedCount, f.fsm.string_byte_candidates)
	// println("childlen", len(childSeeds))

	for _, child := range childSeeds {
		child = f.simulateFF(child)
		if f.failed {
			f.handleFail()
			return true
		}
		hn := child.HashNumber
		equal_hn_seeds := abstractSeedMap[hn].seeds
		abstract_number := abstractSeedMap[hn].abstractNumber
		if len(equal_hn_seeds) < 2 {
			equal_hn_seeds = append(equal_hn_seeds, child)
			abstract_number++
		} else {
			abstract_number++
		}
		absSeedInfo := abstractSeedInfo{
			seeds:          equal_hn_seeds,
			abstractNumber: abstract_number,
		}
		abstractSeedMap[hn] = absSeedInfo
	}
	// println("len!", len(abstractSeedMap))

	for hn, absSeedInfo := range abstractSeedMap {
		seeds := absSeedInfo.seeds
		if len(seeds) == 1 {
			// println("hashNumber", hn, ": 1")
			concreteEndInfo := f.fsm.CoordinateSeed(seeds[0])
			flag := f.handleEndInfo(concreteEndInfo, parentSeeds)
			if flag {
				return flag
			}
		} else {
			absNumber := absSeedInfo.abstractNumber - 2
			// println("hashNumber", hn, ":", absNumber+2)
			absEndInfo := f.fsm.CoordinateAbstraction(hn, absNumber)
			flag := f.handleEndInfo(absEndInfo, parentSeeds)
			if flag {
				return flag
			}
			concreteEndInfo1 := f.fsm.CoordinateSeed(seeds[0])
			flag = f.handleEndInfo(concreteEndInfo1, parentSeeds)
			if flag {
				return flag
			}
			concreteEndInfo2 := f.fsm.CoordinateSeed(seeds[1])
			flag = f.handleEndInfo(concreteEndInfo2, parentSeeds)
			if flag {
				return flag
			}
		}

	}
	// println("iters!")
	for _, p := range parentSeeds {
		f.fsm.CoordinateSeed(p)
	}
	return false
}

func (f *F) handleFail() {
	f.Fail()
	println("----------------------------- FAIL --------------------------------")
	log := f.fsm.Crash_Logger.data
	crashCase := log[len(log)-1]
	println("found failing input", TransForHuman(crashCase.Input), "at", f.fsm.Input_count, "trials")
	println("start minimazing failing input")
	f.minimazeAndLogInput(crashCase.RawContent)
	log = f.fsm.Crash_Logger.data
	minimaizedCase := log[len(log)-1]
	println(CrashCaseToString(minimaizedCase))
	hn := minimaizedCase.HashNumber
	coverage := f.fhm.HashNumber2Coverage(hn)

	println("----------------------------- Trace --------------------------------")
	println(CoverageToString(coverage))
	println("--------------------------------------------------------------------")
	return
}

func (f *F) migrateMachines() {
	println("Machine capacity is full. Start migration...")
	// Abstract existing state_machine
	summarized_seeds := f.fsm.Summarize()
	old_hashNumbers := []HashNumber{}
	for _, seed := range summarized_seeds {
		old_hashNumbers = append(old_hashNumbers, seed.HashNumber)
	}
	coverages_of_seeds := []Coverage{}
	for _, hn := range old_hashNumbers {
		coverages_of_seeds = append(coverages_of_seeds, f.fhm.HashNumber2Coverage(uint(hn)))
	}
	// Sort by re-registering existing seed coverage to the new hash machine
	// Return the value to the seed afterwards
	// The hash number of the seed is 0,1,2... and also registers 0,1,2 and coverage on the hash machine
	f.fhm = New_HashMachine()
	for i, cov := range coverages_of_seeds {
		summarized_seeds[i].HashNumber = f.fhm.RegisterCoverage2HashNumber(cov)
	}
	// Create and relocate a new State_machine
	prev_inputCount := f.fsm.Input_count
	substracted_inputCount := int(prev_inputCount) - len(summarized_seeds)
	f.fsm = New_StateMachine(f.trials, uint(substracted_inputCount))
	f.fsm.verbose = f.verbose
	for _, seed := range summarized_seeds {
		f.fsm.CoordinateSeed(seed)
	}
	println("Migration completed. Resume fuzzing...")
}

func (f *F) minimazeAndLogInput(seedContent []interface{}) {
	minimazalbeIdXs := []int{}
	for i, t := range f.fsm.SeedType {
		if t == Byte_Array {
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
	sampleSeed := Seed{
		Content: content,
	}
	// re confirm error
	f.simulateFF(sampleSeed)
	println("minimaized input")
}

func (f *F) reportF() {
	println("----------------------------- PASS --------------------------------")
	println("Complete", f.trials, "Trials")
	println("Found", (uint(f.fhm.hashNumber_counter.counter) + uint(1)), "coverage")
}

func (f *F) simulateFF(seed Seed) Seed {
	coverage, err, isPanic, panicMsg := monitor(f.ff, seed.Content)

	// seed.hn computation
	// This completes the status change of the seed before logging
	hashNumber := f.fhm.RegisterCoverage2HashNumber(coverage)
	seed.HashNumber = hashNumber
	if isPanic {
		tr := TestResult{
			PanicOccurred: true,
			PanicMessage:  panicMsg,
			Error:         err,
		}
		f.fsm.Crash_Logger.AddCase(seed, tr)
		f.Fail()
	}
	if err != nil {
		tr := TestResult{
			PanicOccurred: false,
			PanicMessage:  "",
			Error:         err,
		}
		f.fsm.Crash_Logger.AddCase(seed, tr)
		f.Fail()
	}
	return seed
}

// Fail marks the function as having failed bur continue execution.
func (f *F) Fail() {
	f.failed = true
}

// Fatal is equivalent to Log followed by FailNow.
// It logs the message and marks the fuzzing as failed.
func (f *F) Fatal(args ...interface{}) {
	var sb strings.Builder

	for _, arg := range args {
		sb.WriteString(arg.(string))
	}

	f.msgs = append(f.msgs, sb.String())
	f.Fail()
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
			coverage = Get_Coverage_of_runner(t, content)
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
	coverage = Get_Coverage_of_runner(t, content)

	// println(string(t.output))
	return coverage, err, isPanic, panicMsg
}

func isSliceEqual(a, b []SupportedType) bool {
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
