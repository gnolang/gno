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
		fsm    *StateMachine
		fhm    *HashMachine
		corpus []Seed
		msgs   []string // Stores log messages for reporting.
		// TODO: CLI flag to set this.
		failed    bool // Indicates whether the fuzzing has encountered a failure.
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
	// int32과 rune은 동일 타입이므로 중복 제거
	// case int32:
	// 	return Int32, nil
	case int64:
		return Int64, nil
	case uint:
		return Uint, nil
	// uint8과 byte는 동일 타입이므로 중복 제거
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

// 일단 다변수함수를 지원하게는 해 놓았지만,
// 현재로썬 걍 type string만 반환하는 상태.
// 하나의 시드를 추가하는 코드.
func (f *F) Add(args ...interface{}) {
	if f.isFuzzed {
		panic("Add after Fuzz")
	}
	var values []interface{}
	var types []SupportedType
	if len(args) == 0 {
		panic("no-argument is denied")
	}
	// 타입 검사 및 값 적층
	for i := range args {
		t, e := typeof(args[i])
		if e != nil {
			panic("not supported type")
		}
		values = append(values, args[i])
		types = append(types, t)
	}

	// 타입 기록 및 타입 일관성 검사
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

	// 초기 시드(코퍼스)들 바탕으로 머신 포멧팅
	f.ff = run
	for i, corp := range f.corpus {
		println("Adding corpus", i+1, "...")
		corp = f.simulateFF(corp)
		if f.failed {
			f.handleFail()
			return
		}
		// 초기 시드들의 해시넘버를 케싱
		hashNumber := corp.HashNumber
		f.fsm.init_hashnumber = append(f.fsm.init_hashnumber, hashNumber)
		endInfo := f.fsm.CoordinateMachine(corp)
		if endInfo.Complete_Trials {
			f.reportF()
			return
		}
		if endInfo.MAXed_CAPACITY {
			// 일종의 새로운 f.Add()
			f.migrateMachines()
			continue
		}

	}
	println("Run trials")
	string_byte_candidates := []int{}
	for i, t := range f.fsm.SeedType {
		if t == String || t == Byte_Array {
			string_byte_candidates = append(string_byte_candidates, i)
		}
	}
	// 상태 기계에 저징
	f.fsm.string_byte_candidates = string_byte_candidates

	// 초기 시드 바탕으로 1회 업데이트
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

func (f *F) updateMachines(parentSeeds []Seed) bool {
	// 단, parentSeeds는 하나여도 가능
	// 애초에 Evlove를 그에 맞게 설계함
	childSeeds := Evolve(parentSeeds, &f.seedCount, f.fsm.string_byte_candidates)
	for _, child := range childSeeds {
		child = f.simulateFF(child)
		if f.failed {
			f.handleFail()
			return true
		}
		endInfo := f.fsm.CoordinateMachine(child)
		if endInfo.Complete_Trials {
			f.reportF()
			return true
		}
		if endInfo.MAXed_CAPACITY {
			// 마저 정리
			for _, p := range parentSeeds {
				f.fsm.CoordinateMachine(p)
			}
			// 일종의 새로운 f.Add()
			f.migrateMachines()
			return false
		}
	}
	// 마저 정리
	// 여기선 인풋 수에 변동이 없다.
	for _, p := range parentSeeds {
		f.fsm.CoordinateMachine(p)
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
	// 로그가 업데이트 됨
	minimaizedCase := log[len(log)-1]
	println(CrashCaseToString(minimaizedCase))
	hn := minimaizedCase.HashNumber
	coverage := f.fhm.HashNumber2Coverage(hn)

	println("----------------------------- Trace --------------------------------")
	println(CoverageToString(coverage))
	println("--------------------------------------------------------------------")
	return
}

// 로직 상 머신의 배열이 다 들어차는 순간이 옴 (인풋이 5만개 이상 시)
// 이 경우 머신을 축약후, 새 머신으로 옮김으로써 이어 시작함.
func (f *F) migrateMachines() {
	println("Machine capacity is full. Start migration...")
	// 기존의 스테이트_머신을 축약
	summarized_seeds := f.fsm.Summarize()
	old_hashNumbers := []HashNumber{}
	for _, seed := range summarized_seeds {
		old_hashNumbers = append(old_hashNumbers, seed.HashNumber)
	}
	coverages_of_seeds := []Coverage{}
	for _, hn := range old_hashNumbers {
		coverages_of_seeds = append(coverages_of_seeds, f.fhm.HashNumber2Coverage(uint(hn)))
	}
	// 새 해시머신에 기존 시드의 커버리지를 재등록함으로써 정렬
	// 이후 그 값을 시드에 리턴
	// 시드의 헤시넘버는 0,1,2...이며, 해시머신에도 0,1,2 및 커버리지 등록
	f.fhm = New_HashMachine()
	for i, cov := range coverages_of_seeds {
		summarized_seeds[i].HashNumber = f.fhm.RegisterCoverage2HashNumber(cov)
	}
	// 새로운 스테이트_머신을 생성 후 이전시키기기
	prev_inputCount := f.fsm.Input_count
	// re-코디네이팅 시의 인풋 중복 제거.
	substracted_inputCount := int(prev_inputCount) - len(summarized_seeds)
	// 이주할 새 fsm생성
	f.fsm = New_StateMachine(f.trials, uint(substracted_inputCount))
	f.fsm.verbose = f.verbose
	// 시뮬레이팅 할 필요 없이 곧바로 코디네이팅 (hashNumber가 이미 정해졌으므로로)
	for _, seed := range summarized_seeds {
		f.fsm.CoordinateMachine(seed)
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
	// 샘플을 훼손하지 않아야 함.
	content := make([]interface{}, len(sample))
	for i, v := range sample {
		content[i] = v // 안전하게 캐스팅
	}
	// progressive, preserveness이용해서 최소화시키기
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
	// 다시 한 번 확인. 커버리지 받아오기
	// 다시 로그에 기록
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

	// seed.hn 연산
	// 이로써 seed의 상태변화는 로깅 전에 마무리됨
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
		// sOrb를 하나 줄이기고 check
		switch v := sOrb.(type) {
		case string:
			if len(v) < 1 {
				continue
			}
			for i := 0; i < len(v); i++ {
				b := []byte(v)
				candidate := append(b[:i], b[i+1:]...)
				// 테스트용 데이터 생성
				tester := make([]interface{}, len(content))
				for i, v := range content {
					tester[i] = v // 안전하게 캐스팅
				}
				tester[idx] = string(candidate)
				if f.checkPreserve(tester).errorIsPreserved {
					// 제거 확정
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
				// 테스트용 데이터 생성
				tester := make([]interface{}, len(content))
				for i, v := range content {
					tester[i] = v // 안전하게 캐스팅
				}
				tester[idx] = []byte(candidate)
				if f.checkPreserve(tester).errorIsPreserved {
					// 제거 확정
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

// TODO: 여기서 커버리지 관련은 반드시 수정하기!!!
// TODO: 여기선 test결과 따라 커버리지 하드코딩해 두었음. 이거 추후 수정!!
func monitor(run Runner, content []interface{}) (coverage Coverage, err error, isPanic bool, panicMsg string) {
	isPanic = false
	panicMsg = ""
	err = nil
	coverage = Coverage{}
	// panic을 감지하기 위한 defer
	defer func() {
		if r := recover(); r != nil {
			t := NewT("fuzzing")
			coverage = Get_Coverage_of_runner(t, content)
			isPanic = true
			// r이 error 타입이면 .Error() 문자열을 추출
			if err, ok := r.(error); ok {
				panicMsg = err.Error()
				return
			}
			// r이 string 타입이면 그대로 문자열로 사용
			if s, ok := r.(string); ok {

				panicMsg = s
				return
			}
			// 그 외에는 직접 문자열 리턴이 불가능하므로 예시로 "unknown panic" 처리
			panicMsg = "unknown panic"
		}
	}()
	t := NewT("fuzzing")

	// 콘텐트의 불변성 보장
	copied := make([]interface{}, len(content))
	for i, v := range content {
		copied[i] = v // 안전하게 캐스팅
	}
	// 배열을 스프레딩
	run(t, copied...)
	// 테스팅 결과 핸들링링
	info := t.GetResult()
	if info.Failed {
		err = errors.New(string(info.Output))
	}
	// TODO: 이 함수를, 진짜 커버리지 얻는 함수로 수정하기
	// TODO: 구현 따라서, 이것도 카피해서 넣어야 함.
	coverage = Get_Coverage_of_runner(t, content)

	// println(string(t.output))
	return coverage, err, isPanic, panicMsg
}

// Slice 비교 함수
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
