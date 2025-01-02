package testing

import (
	"errors"
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
		trials uint     // Number of iterations to run the fuzzing process.
		// TODO: CLI flag to set this.
		failed    bool // Indicates whether the fuzzing has encountered a failure.
		ff        Runner
		isFuzzed  bool
		seedCount uint
	}
)

func New_F(trials uint) *F {
	println("creating fuzzing enviroment...")
	new_fsm := New_StateMachine(trials, 0)
	new_fhm := New_HashMachine()
	return &F{
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
	var values []interface{}
	var types []SupportedType

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
		Pid: 0,
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
		println("Adding corpus", i, "...")
		corp = f.simulateFF(corp)
		endInfo := f.fsm.CoordinateMachine(corp)
		if endInfo.Complete_Trials {
			f.report()
			return
		}
		if endInfo.MAXed_CAPACITY {
			// 일종의 새로운 f.Add()
			f.migrateStateMachine()
			continue
		}
		// println("코디네이팅 결과 h0->p: ", f.fsm.HashNumber2Priority[0])
		// println("코디네이팅 결과 h1->p: ", f.fsm.HashNumber2Priority[1])
	}
	println("running trials...")

	for {
		parentSeeds := f.fsm.PopSeeds()
		childSeeds := Evolve(parentSeeds, &f.seedCount)

		for _, child := range childSeeds {
			child = f.simulateFF(child)
			endInfo := f.fsm.CoordinateMachine(child)
			if endInfo.Complete_Trials {
				f.report()
				return
			}
			if endInfo.MAXed_CAPACITY {
				// 마저 정리
				for _, p := range parentSeeds {
					f.fsm.CoordinateMachine(p)
				}
				// 일종의 새로운 f.Add()
				f.migrateStateMachine()
				continue
			}
		}
		// 마저 정리
		// 여기선 인풋 수에 변동이 없다.
		for _, p := range parentSeeds {
			f.fsm.CoordinateMachine(p)
		}
	}
}

// 로직 상 머신의 배열이 다 들어차는 순간이 옴 (인풋이 5만개 이상 시)
// 이 경우 머신을 축약후, 새 머신으로 옮김으로써 이어 시작함.
func (f *F) migrateStateMachine() {
	println("Machine capacity is full. Start migration...")
	// 기존의 스테이트_머신을 축약
	summarized_seeds := f.fsm.Summarize()
	// 새로운 스테이트_머신을 생성 후 이전시키기기
	prev_inputCount := f.fsm.Input_count
	// re-코디네이팅 시의 인풋 중복 제거.
	substracted_inputCount := int(prev_inputCount) - len(summarized_seeds)
	// 이주할 새 fsm생성
	f.fsm = New_StateMachine(f.trials, uint(substracted_inputCount))
	// 시뮬레이팅 할 필요 없이 곧바로 코디네이팅 (hashNumber가 이미 정해졌으므로로)
	for _, seed := range summarized_seeds {
		f.fsm.CoordinateMachine(seed)
	}
	println("Migration completed. Resume fuzzing.")
}

func (f *F) report() {
	println("----------------------------- Complete Trials --------------------------------")
	println("----------------------------- Fuzz Report --------------------------------")
	kinds := f.fsm.Crash_Logger.Kind()
	summary := f.fsm.Crash_Logger.Summary()
	PrintKinds(kinds)
	PrintSummary(summary)
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
	}
	if err != nil {
		tr := TestResult{
			PanicOccurred: false,
			PanicMessage:  "",
			Error:         err,
		}

		f.fsm.Crash_Logger.AddCase(seed, tr)
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
	// 배열을 스프레딩
	run(t, content...)
	// 테스팅 결과 핸들링링
	info := t.GetResult()
	if info.Failed {
		err = errors.New(string(info.Output))
	}
	// TODO: 이 함수를, 진짜 커버리지 얻는 함수로 수정하기
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

// new_seed := Seed{
// 	Gen: 0,
// }
// slicetest([]Seed{new_seed})
// println(new_seed)

// func slicetest(seeds []Seed) {
// 	seeds[0].Gen = 100
// }
