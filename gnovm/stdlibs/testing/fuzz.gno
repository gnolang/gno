package testing

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing/fuzzing"
)

type (
	// Runner is the type of the target function that f.Fuzz will execute.
	Runner func(*T, ...interface{})

	// F is the structure used for fuzzing.
	F struct {
		manager              *fuzzing.FuzzManager // Maneger manages seed,hashNumber(coverage) and priority for fuzzing, sorting and organizing them to quickly find the next input candidate.
		hasher               *fuzzing.FuzzHasher  // Hasher assigns unique hashNumbers to coverage used in fuzzing.
		logger               *fuzzing.FuzzLogger  // Logger records and manages various messages from the fuzzer
		mutator              *fuzzing.Mutator     // Mutator modifies and generates input values used for fuzzing
		corpus               []fuzzing.Seed       // Copus are "Real seeds" added through f.Add.
		msgs                 []string             // Msgs stores failing messages for reporting.
		failed               bool                 // Failed indicates whether the fuzzing process has encountered a failure.
		ff                   Runner               // Ff is the target function for f.Fuzz.
		isFuzzed             bool                 // IsFuzzed monitors whether f.Fuzz has been called
		seedCount            uint                 // SeedCount counts the number of seeds attempted.
		stringByteCandidates []int                // StringByteCandidates records the indices of string and []byte arguments among Runner's args.
		dur                  string               // Dur records the time taken for fuzzing.

		// For command-line tests
		verbose bool // verbose determines the frequency of logging based on CLI input.
		trials  uint // trials is number of iterations to run the fuzzing process.
	}
)

// Add adds initial seeds to the f.corpus.
// It prevents calling f.Add after f.Fuzz.
func (f *F) Add(args ...interface{}) {
	// Disallow calling f.Add after f.Fuzz.
	if f.isFuzzed {
		fmt.Fprintln(os.Stderr, "Add after Fuzz")
		f.Fail()
		return
	}

	// Store the arguments received by f.Add
	var values []interface{}
	var types []fuzzing.SupportedType
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "zero-argument is denied")
		f.Fail()
		return
	}
	for i := range args {
		t, e := typeof(args[i])
		if e != nil {
			fmt.Fprintln(os.Stderr, "not supported type")
			f.Fail()
			return
		}
		values = append(values, args[i])
		types = append(types, t)
	}
	// Initialize SeedType if not already set.
	if f.manager.SeedType == nil {
		f.manager.SeedType = types
		// The mutator is initialized based on the types.
		f.mutator = fuzzing.NewMutator(types)
	} else {
		// Ensure that all added arguments have the same type.
		if !isSliceEqual(f.manager.SeedType, types) {
			fmt.Fprintln(os.Stderr, "added arguments not equal together")
			f.Fail()
			return
		}
	}
	// Increment the seed count and add to corpus.
	f.seedCount++
	f.corpus = append(f.corpus, fuzzing.Seed{
		Pid:           f.seedCount,
		Id:            f.seedCount,
		Gen:           1,
		IsCoordinated: false,
		Content:       values,
	})
}

// typeof takes an interface{} and returns its type and whether it is supported.
func typeof(arg interface{}) (fuzzing.SupportedType, error) {
	switch arg.(type) {
	case []byte:
		return fuzzing.ByteArray, nil
	case string:
		return fuzzing.String, nil
	case bool:
		return fuzzing.Bool, nil
	// duduplication. uint8 and byte are of the same type.
	// case byte:
	// 	return fuzzing.Byte, nil
	// deduplication because int32 and rune are of the same type
	// case rune:
	// 	return fuzzing.Rune, nil
	case float32:
		return fuzzing.Float32, nil
	case float64:
		return fuzzing.Float64, nil
	case int:
		return fuzzing.Int, nil
	case int8:
		return fuzzing.Int8, nil
	case int16:
		return fuzzing.Int16, nil
	case int32:
		return fuzzing.Int32, nil
	case int64:
		return fuzzing.Int64, nil
	case uint:
		return fuzzing.Uint, nil
	case uint8:
		return fuzzing.Uint8, nil
	case uint16:
		return fuzzing.Uint16, nil
	case uint32:
		return fuzzing.Uint32, nil
	case uint64:
		return fuzzing.Uint64, nil
	default:
		return "", errors.New("unsupported type:")
	}
}

// isSliceEqual checks if two slices of supportedType are equal.
func isSliceEqual(a, b []fuzzing.SupportedType) bool {
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

// Fuzz orchestrates input mutation and execution for fuzz testing.
// It utilizes f.manager to manage inputs and f.mutator to generate variations
// of seeds added via f.Add. The target function is executed using f.Fuzz.
//
// The fuzzing process follows these steps:
// 1. Retrieve input from f.manager.
// 2. Mutate the input using f.mutator, storing mutation details, including:
//   - argumentsSet: Specifies the subset of arguments modified.
//   - entity: Indicates which part of an argument was changed.
//   - mutationStrategy: Describes the transformation applied.
//
// 3. Execute the mutated inputs.
// 4. Update the hasher, manager, and mutator based on execution results:
//   - The hasher records coverage data and maps it to hash values.
//   - The manager clusters seeds with similar hash values and adjusts priority rankings.
//   - The mutator refines mutation strategies using (execution results, original input, mutation strategy).
//
// 5. If an error occurs or the iteration limit is reached, the process outputs results and terminates.
func (f *F) Fuzz(run Runner) {
	if !f.isFuzzed {
		f.isFuzzed = true
	} else {
		fmt.Fprintln(os.Stderr, "fuzz called more than once")
		return
	}
	// If f.Fatal or f.Fail was called in FuzzXXX, return immediately.
	if f.failed {
		return
	}

	// Initialize fuzzing by registering the target function.
	f.ff = run

	// Format manager and mutator using corpus from f.Add.
	for _, corp := range f.corpus {
		corp = f.simulateFF(corp)
		if f.failed {
			f.handleFail()
			return
		}
		hashNumber := corp.HashNumber
		f.manager.Inithashnumber = append(f.manager.Inithashnumber, hashNumber)
		endInfo := f.manager.CoordinateSeed(corp)
		f.mutator.Coordinate2SampleSet(corp, 1)
		// TODO: Add it as a self-referential structure.
		// Remove duplicate coverage entries and ensure each unique coverage retains a minimal self-referential structure.

		if endInfo.CompleteTrials {
			f.reportF()
			return
		}
		if endInfo.MaxedCAPACITY {
			f.migrateMachines()
			continue
		}
	}

	// Guidelines for evolving seeds.
	// Used by the mutator to determine mutation strategies.
	// Strings and []byte inputs are parsed and analyzed to update mutation strategies accordingly.

	stringByteCandidates := []int{}
	for i, t := range f.manager.SeedType {
		if t == fuzzing.String || t == fuzzing.ByteArray {
			stringByteCandidates = append(stringByteCandidates, i)
		}
	}

	f.stringByteCandidates = stringByteCandidates
	fmt.Fprintln(os.Stderr, "Run trials...")
	// Independently generate inputs from the initially formatted seeds.
	// Ensures that all initial seeds are executed and mutated at least once.
	// Inputs provided via f.Add are executed independently, regardless of priority.

	for i := 0; i < len(f.corpus); i++ {
		hn := f.manager.Inithashnumber[i]
		initSeed := f.manager.PopInitSeedByHN(hn)
		isEnd := f.updateMachines(initSeed)
		if isEnd {
			return
		}
	}
	for {

		// 1) Pop input from the manager
		parentSeed := f.manager.PopSeed()
		// 2) Generate evolved input from the retrieved input.
		// 3) Simulate and log the evolved input when executing the target function.
		// 4) Coordinate input management and repeatedly update the mutator.

		isEnd := f.updateMachines(parentSeed)
		// 5) If an error occurs or trials (iterations) are completed, output the result and terminate.
		if isEnd {
			return
		}
	}
}

// simulateFF simulates the target function with the given seed.
func (f *F) simulateFF(seed fuzzing.Seed) fuzzing.Seed {
	coverage, err, isPanic, panicMsg := monitor(f.ff, seed.Content)

	// Compute seed.hn
	// This completes the status change of the seed before logging
	hashNumber := f.hasher.RegisterCoverage2HashNumber(coverage)
	seed.HashNumber = hashNumber
	if isPanic {
		tr := fuzzing.TestResult{
			PanicOccurred: true,
			PanicMessage:  panicMsg,
			Error:         err,
		}
		f.logger.AddCase(seed, tr)
		f.manager.CoordinateSeed(seed)
		f.mutator.Coordinate2SampleSet(seed, 1)
		f.Fail()
	}
	if err != nil {
		tr := fuzzing.TestResult{
			PanicOccurred: false,
			PanicMessage:  "",
			Error:         err,
		}
		f.logger.AddCase(seed, tr)
		f.manager.CoordinateSeed(seed)
		f.mutator.Coordinate2SampleSet(seed, 1)
		f.Fail()
	}
	return seed
}

// monitor runs the target function and captures coverage, errors, and panics.
// TODO: Make sure to revise coverage here!!!
// It's hard-coded coverage according to the test results here. This will be corrected later!!
func monitor(run Runner, content []interface{}) (coverage fuzzing.Coverage, err error, isPanic bool, panicMsg string) {
	isPanic = false
	panicMsg = ""
	err = nil
	coverage = fuzzing.Coverage{}
	defer func() {
		if r := recover(); r != nil {

			// TODO: part 1
			coverage = fuzzing.CoverageOfrunner(content)
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
	// TODO: part 2
	// Modifying this function to get real coverage
	// It's just pseudo-coverage of some function
	coverage = fuzzing.CoverageOfrunner(content)

	return coverage, err, isPanic, panicMsg
}

// Fail marks the function as having failed but continues execution.
func (f *F) Fail() {
	f.failed = true
}

// FailNow marks the fuzz test as failed and stops execution immediately.
// It is primarily used within FuzzXXX functions. While it can be used inside f.Fuzz,
// it is recommended to use the test framework's methods instead.
func (f *F) FailNow() {
	templet := "fuzzing: you called FailNow"
	f.logOrPanic(templet)
	f.failed = true
}

// Fatal logs the provided arguments and then calls FailNow.
// It is equivalent to calling Log followed by FailNow.
// While it can be used inside f.Fuzz, using the test framework's methods is preferred.
func (f *F) Fatal(args ...interface{}) {
	var sb strings.Builder

	for _, arg := range args {
		sb.WriteString(fmt.Sprint(arg)) // Supports various types using fmt.Sprint
		sb.WriteString(" ")             // Adds space between messages
	}

	message := sb.String()
	f.msgs = append(f.msgs, message)

	// Use the helper function to log and panic
	f.logOrPanic("fuzzing: you called Fatal %d time(s) \"%v\"\n", len(f.msgs), f.msgs)
	f.Fail()
}

// Fatalf logs a formatted message and then calls FailNow.
// It functions similarly to Fatal but supports formatted strings.
// While it can be used inside f.Fuzz, using the test framework's methods is preferred.
func (f *F) Fatalf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	f.msgs = append(f.msgs, message)

	f.logOrPanic("fuzzing: you called Fatalf %d time(s) \"%v\"\n", len(f.msgs), f.msgs)
	f.Fail()
}

// logOrPanic is a helper function that logs the message or then panics.
func (f *F) logOrPanic(templet string, args ...interface{}) {
	// Log the message to stderr
	fmt.Fprintf(os.Stderr, templet, args...)

	// If already fuzzed, panic with the message
	if f.isFuzzed {
		panic(fmt.Sprintf(templet, args...))
	}
}

// handleFail handles a failure during fuzzing by logging and minimizing the input.
func (f *F) handleFail() {
	fmt.Fprintln(os.Stderr, "\n--- FAIL:")

	f.logger.ReportLastFailingInput()
	lastCrashCase := f.logger.LastCrashCase()
	f.minimizeAndLogInput(lastCrashCase.RawContent)
	f.logger.ReportLastCrashCase()
	minimaizedCase := f.logger.LastCrashCase()
	hn := minimaizedCase.HashNumber
	coverage := f.hasher.HashNumber2Coverage(hn)
	if f.verbose {
		fmt.Fprintln(os.Stderr, "\n--- Trace:")
		fmt.Fprintf(os.Stderr, "%s\n", coverageToString(coverage))
	} else {
		fmt.Fprintln(os.Stderr, "\n--- Trace:")
		fmt.Fprintf(os.Stderr, "Last covered line: function \"%s\" in line %d\n", coverage[len(coverage)-1].CoName, coverage[len(coverage)-1].CoLine)
	}
	if f.verbose {
		f.analyzeResult()
	}
	return
}

// analyzeResult performs analysis on the fuzzing results and logs the findings.
func (f *F) analyzeResult() {
	fmt.Fprintln(os.Stderr, "\n--- Analytics:")
	if f.verbose {

		hashNumber := 0
		for f.manager.HashNumber2Priority[hashNumber] != 0 {
			fmt.Fprintf(os.Stdout, "coverage %d: [%d frequency | %d actual storage] \n", hashNumber, f.manager.HashNumber2Priority[hashNumber], len(f.manager.HashNumber2Seeds[hashNumber].Data))
			dataLen := len(f.manager.HashNumber2Seeds[hashNumber].Data)
			sampleCount := getSampleCount(dataLen) // Determines sample count
			fmt.Fprint(os.Stdout, "- ")
			for j := 0; j < sampleCount; j++ {
				index := j * (dataLen / sampleCount) // Selects indices at even intervals
				s := f.manager.HashNumber2Seeds[hashNumber].Data[index]

				if j > 0 {
					fmt.Fprint(os.Stdout, ", ")
				}
				fmt.Fprintf(os.Stdout, "%v", fuzzing.ContentToString(s.Content))
			}
			fmt.Fprintln(os.Stdout) // Adds a newline

			hashNumber++
		}
	}
}

// getSampleCount determines the number of samples based on input length.
func getSampleCount(length int) int {
	switch {
	case length >= 4096:
		return 5
	case length >= 512:
		return 4
	case length >= 64:
		return 3
	case length >= 8:
		return 2
	default:
		return 1
	}
}

// coverageToString converts coverage data to a string representation.
func coverageToString(coverage fuzzing.Coverage) string {
	var sb strings.Builder
	sb.WriteString("[")
	count := 0
	for i, line := range coverage {
		sb.WriteString("{co_name: ")
		sb.WriteString(line.CoName)
		sb.WriteString(", co_line: ")
		sb.WriteString(strconv.Itoa(line.CoLine))
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

// minimizeAndLogInput reduces seedContent to its minimal form while preserving error conditions.
// It then executes a simulation with the minimized input and logs the result.
func (f *F) minimizeAndLogInput(seedContent []interface{}) {
	// Condition function: Checks if an input preserves an error or panic state.
	cond := func(tester []interface{}) bool {
		_, err, isPanic, _ := monitor(f.ff, tester)
		if isPanic {
			return true
		}
		if err != nil {
			return true
		}
		return false
	}

	candidate := f.minimizeByCond(seedContent, cond)

	sampleSeed := fuzzing.Seed{
		Content: candidate,
	}
	// Execute simulation with minimized input and log results.
	f.simulateFF(sampleSeed)
	fmt.Fprintln(os.Stderr, "\n--- Minimized input:")
}

// minimizeByHashNumber minimizes seedContent while ensuring the resulting input
// produces the same hash number (hn) when monitored.
func (f *F) minimizeByHashNumber(seedContent []interface{}, hn fuzzing.HashNumber) []interface{} {
	// Condition function: Ensures monitored output produces the same hash number.
	cond := func(tester []interface{}) bool {
		coverage, _, _, _ := monitor(f.ff, tester)
		if f.hasher.IsExistingCoverage(coverage) {
			newHn := f.hasher.RegisterCoverage2HashNumber(coverage)
			return newHn == hn
		} else {
			return false
		}
	}

	return f.minimizeByCond(seedContent, cond)
}

// minimizeByCond iteratively reduces seedContent while satisfying the given condition function.
func (f *F) minimizeByCond(seedContent []interface{}, cond func([]interface{}) bool) []interface{} {
	// Identify indices of minimizable elements (e.g., ByteArray, String).
	var minimizableIndices []int
	for i, t := range f.manager.SeedType {
		if t == fuzzing.ByteArray || t == fuzzing.String {
			minimizableIndices = append(minimizableIndices, i)
		}
	}
	if len(minimizableIndices) < 1 {
		return seedContent
	}

	// Create a copy of seedContent.
	candidate := fuzzing.DeepCopyInterfaceList(seedContent)

	// Iteratively minimize input.
	for {
		newCandidate, progressed, occurredPoint := f.checkProgressWithCond(candidate, minimizableIndices, cond)
		if !progressed {
			break
		}
		candidate = newCandidate
		// Skip already processed minimizable indices.
		minimizableIndices = minimizableIndices[occurredPoint:]
	}

	return candidate
}

// checkProgressWithCond attempts to reduce minimizable elements while preserving the given condition.
// It returns the best minimized candidate, whether progress was made, and the index of the modified element.
func (f *F) checkProgressWithCond(content []interface{}, minimizableIndices []int, cond func([]interface{}) bool) ([]interface{}, bool, int) {
	for occurredPoint, idx := range minimizableIndices {
		element := content[idx]
		var candidates [][]byte
		var isString bool

		switch v := element.(type) {
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
			continue // Skip non-minimizable types.

		}

		// Test each candidate against the condition function.
		for _, candidateBytes := range candidates {
			tester := fuzzing.DeepCopyInterfaceList(content)

			if isString {
				tester[idx] = string(candidateBytes)
			} else {
				tester[idx] = candidateBytes
			}

			if cond(tester) {
				// Return the minimized candidate if condition holds.
				return tester, true, occurredPoint
			}
		}
	}
	return content, false, 0
}

// generateCandidates generates potential minimization candidates by removing one character at a time.
func generateCandidates(b []byte) [][]byte {
	candidates := make([][]byte, 0, len(b))
	for i := 0; i < len(b); i++ {
		// createSliceWithoutIndex generates a new slice with the character at the given index removed.

		candidate := append([]byte{}, b[:i]...)
		candidate = append(candidate, b[i+1:]...)
		candidates = append(candidates, candidate)
	}
	return candidates
}

// reportF reports the fuzzing results after successful completion.
func (f *F) reportF() {
	fmt.Fprintln(os.Stderr, "\n--- PASS:")
	fmt.Fprintf(os.Stderr, "Complete %d Trials\n", f.manager.InputCount)
	fmt.Fprintf(os.Stderr, "Inspected %d coverage\n", (f.manager.InspectingHashNumbers))
	if f.verbose {
		f.analyzeResult()
	}
}

// migrateMachines abstracts the existing f.manager and relocates the f.manager.
func (f *F) migrateMachines() {
	// Abstract existing f.manager
	summarizedSeeds := f.manager.SummarizeSeeds()

	prevInputCount := f.manager.InputCount
	substractedInputCount := int(prevInputCount) - len(summarizedSeeds)
	prevSeedType := make([]fuzzing.SupportedType, len(f.manager.SeedType))
	copy(prevSeedType, f.manager.SeedType)
	prevInitHashNumber := f.manager.Inithashnumber
	// Reinitialize f.manager while preserving relevant data.
	f.manager = fuzzing.NewFuzzManager(f.trials, uint(substractedInputCount))
	f.manager.Verbose = f.verbose
	f.manager.SeedType = prevSeedType
	f.manager.Inithashnumber = prevInitHashNumber
	for _, seed := range summarizedSeeds {
		f.manager.CoordinateSeed(seed)
	}
}

// updateMachines performs the following steps:
// 1) Generates mutated input from it's argument(parentSeed)
// 2) Simulates and logs evolved input to the target function
// 3) Coordinates input to the fuzzManager, updates the mutator,
// and determines whether fuzzing should be terminated.
func (f *F) updateMachines(parentSeed fuzzing.Seed) bool {
	f.manager.CoordinateSeed(parentSeed)
	// 1) Generate mutated input from the popped input by mutator.
	childSeeds := f.mutator.MutateSeed2Seeds(parentSeed, &f.seedCount)

	seedsMap := make(map[fuzzing.HashNumber]*[]fuzzing.Seed)
	// 2) Simulate and log evolved input to the target function
	// If failure occurs, log it and terminate
	for _, child := range childSeeds {
		child = f.simulateFF(child)
		if f.failed {
			f.handleFail()
			return true
		}
		hn := child.HashNumber
		if _, exists := seedsMap[hn]; !exists {
			seedsMap[hn] = &[]fuzzing.Seed{}
		}
		seeds := seedsMap[hn]
		*seeds = append(*seeds, child)
	}

	// 3) Update fuzz manager with the new inputs.
	for hn, seeds := range seedsMap {
		if !f.manager.IsCoordinatedHashNumber(hn) {
			// If the hash number has not been coordinated yet,
			// it coordinates the seed and directly adds it to the mutator.

			// TODO: Select "shortSeed" using an appropriate heuristic.
			// The smaller the initial seed, the shorter the minimization time.

			seedArray := *(seeds)
			shortSeed := seedArray[0] // Sample seed for coordination
			copiedContent := make([]interface{}, len(shortSeed.Content))
			for i, v := range shortSeed.Content {
				copiedContent[i] = v
			}

			minimizedContent := f.minimizeByHashNumber(copiedContent, hn)
			minimizedSeed := fuzzing.Seed{
				Pid:           shortSeed.Pid,
				Id:            shortSeed.Id,
				Gen:           shortSeed.Gen,
				HashNumber:    shortSeed.HashNumber,
				IsCoordinated: shortSeed.IsCoordinated,
				Content:       minimizedContent,
			}
			abstractCount := len(seedArray)

			abstractEndInfo := f.manager.CoordinateAbstraction(hn, uint(abstractCount))
			flag1 := f.handleEndInfo(abstractEndInfo)
			if flag1 {
				return true
			}
			concreteEndInfo := f.manager.CoordinateSeed(minimizedSeed)
			f.mutator.Coordinate2SampleSet(minimizedSeed, int(abstractCount+1))
			flag2 := f.handleEndInfo(concreteEndInfo)
			if flag2 {
				return true
			}
		} else {
			// If the hash number has already been coordinated,
			// it is not stored directly. Instead, it is used to adjust the manager and mutator.

			seedArray := *(seeds)
			sampleSeed := seedArray[0]
			sampleSet, foundSampleSet := f.mutator.FindMutSampleSet(hn)
			if !foundSampleSet {
				panic("Logic error: Unable to find sampleSet based on hash number")
			}
			if parentId, foundParent := sampleSet.FindParentID(sampleSeed.Pid); foundParent {
				// If a mutated input returns to the parent's coverage,
				// the parent's mutation strategy is updated accordingly.
				// This creates a bias based on survivors, making future mutation strategies
				// more likely to successfully pass program validation.

				parentSample, foundParentSample := sampleSet.FindSample(parentId)
				if !foundParentSample {
					panic("logic error: Unable to find sample sample based on parent id")
				}
				for _, seed := range seedArray {
					parentSample.ReflectSeed(seed)
					sampleSet.PlusClusterPopularity(1)
				}
				abstractCount := len(seedArray)
				abstractEndInfo := f.manager.CoordinateAbstraction(hn, uint(abstractCount))
				flag := f.handleEndInfo(abstractEndInfo)
				if flag {
					return true
				}
			} else {
				// Even if the mutated input reaches an already registered hash number,
				// if the transformation results in a different hash number than the original,
				// it is considered a special case and is directly inserted into the coordinator and mutator.

				// TODO: Consider a better way to utilize "siblings."
				// 1) Use sampleSet.FindSiblingID to find a sibling.
				// 2) Infer structural similarity using an appropriate heuristic (e.g., length tracking).
				//    If ID-based weighted operations are possible, it would be beneficial.
				//    Be cautious, as direct weighted operations may cause merge conflicts when structural changes occur.
				// TODO: Select "shortSeed" using an appropriate heuristic.
				// The smaller the initial seed, the shorter the minimization time.
				seedArray := *(seeds)
				shortSeed := seedArray[0] // Sample seed for coordination
				abstractCount := len(seedArray)
				abstractEndInfo := f.manager.CoordinateAbstraction(hn, uint(abstractCount))
				flag1 := f.handleEndInfo(abstractEndInfo)
				if flag1 {
					return true
				}
				concreteEndInfo := f.manager.CoordinateSeed(shortSeed)
				f.mutator.Coordinate2SampleSet(shortSeed, int(abstractCount+1))
				flag2 := f.handleEndInfo(concreteEndInfo)
				if flag2 {
					return true
				}
			}
		}
	}
	return false
}

// handleEndInfo handles the endInfo returned by coordinateSeed or coordinateAbstraction.
func (f *F) handleEndInfo(endInfo fuzzing.EndInfo) bool {
	if endInfo.CompleteTrials {
		f.reportF()
		return true
	}
	if endInfo.MaxedCAPACITY {
		f.migrateMachines()
		return false
	}
	return false
}

// NewF creates a new instance of F with the specified verbosity and number of trials.
func NewF(verbose bool, trials uint) *F {
	newManager := fuzzing.NewFuzzManager(trials, 0)
	newManager.Verbose = verbose
	newHasher := fuzzing.NewFuzzHasher()
	newLogger := fuzzing.NewFuzzLogger()
	return &F{
		verbose: verbose,
		manager: newManager,
		hasher:  newHasher,
		logger:  newLogger,
		// The mutator is created after receiving the seed type via Add.
		mutator:   nil,
		isFuzzed:  false,
		seedCount: 0,
		trials:    trials,
	}
}
