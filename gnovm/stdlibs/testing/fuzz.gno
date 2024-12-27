package testing

import (
	"errors"
	"strconv"
	"strings"
	"testing"
	"testing/fuzzinglib"
	"unicode/utf8"
)

type (
	Runner func(*testing.T, ...interface{})
	F      struct {
		fsm    *fuzzinglib.StateMachine
		fhm    *fuzzinglib.HashMachine
		corpus []fuzzinglib.Seed
		msgs   []string // Stores log messages for reporting.
		iters  int      // Number of iterations to run the fuzzing process.
		// TODO: CLI flag to set this.
		failed    bool // Indicates whether the fuzzing has encountered a failure.
		ff        Runner
		isFuzzed  bool
		seedCount uint
	}
)

func New_F() *F {
	new_fsm := fuzzinglib.New_StateMachine()
	new_fhm := fuzzinglib.New_HashMachine()
	return &F{
		fsm:       new_fsm,
		fhm:       new_fhm,
		isFuzzed:  false,
		seedCount: 0,
	}
}

func typeof(arg interface{}) (fuzzinglib.SupportedType, error) {
	switch v := arg.(type) {
	case []byte:
		return fuzzinglib.Byte_Array, nil
	case string:
		return fuzzinglib.String, nil
	case bool:
		return fuzzinglib.Bool, nil
	case byte:
		return fuzzinglib.Byte, nil
	case rune:
		return fuzzinglib.Rune, nil
	case float32:
		return fuzzinglib.Float32, nil
	case float64:
		return fuzzinglib.Float64, nil
	case int:
		return fuzzinglib.Int, nil
	case int8:
		return fuzzinglib.Int8, nil
	case int16:
		return fuzzinglib.Int16, nil
	// int32과 rune은 동일 타입이므로 중복 제거
	// case int32:
	// 	return fuzzinglib.Int32, nil
	case int64:
		return fuzzinglib.Int64, nil
	case uint:
		return fuzzinglib.Uint, nil
	// uint8과 byte는 동일 타입이므로 중복 제거
	// case uint8:
	// 	return fuzzinglib.Uint8, nil
	case uint16:
		return fuzzinglib.Uint16, nil
	case uint32:
		return fuzzinglib.Uint32, nil
	case uint64:
		return fuzzinglib.Uint64, nil
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
	var types []fuzzinglib.SupportedType

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
	f.corpus = append(f.corpus, fuzzinglib.Seed{
		Pid: 0,
		Id:  f.seedCount,
		Gen: 1, IsCoordinated: false, Content: values,
	})
}

func (f *F) Fuzz(run Runner) {
	if !f.isFuzzed {
		f.isFuzzed = true
	} else {
		panic("fuzz called more than once")
	}

	// 초기 시드(코퍼스)들 바탕으로 머신 포멧팅
	f.ff = run
	for i, corp := range f.corpus {
		println("run corpus", i, "...")
		coverage := f.runFF(corp.Content)
		// 러닝 처리
		hashNumber := f.fhm.RegisterCoverage2HashNumber(coverage)
		corp.HashNumber = hashNumber
		is_end := f.fsm.CoordinateMachine(corp)
		if is_end {
			f.report()
			return
		}
		// println("코디네이팅 결과 h0->p: ", f.fsm.HashNumber2Priority[0])
		// println("코디네이팅 결과 h1->p: ", f.fsm.HashNumber2Priority[1])
	}
	for {

		parentSeeds := f.fsm.PopSeeds()
		childSeeds := fuzzinglib.Evolve(parentSeeds, &f.seedCount)

		for _, child := range childSeeds {
			coverage := f.runFF(child.Content)
			// 러닝 처리
			hashNumber := f.fhm.RegisterCoverage2HashNumber(coverage)
			child.HashNumber = hashNumber
			is_end := f.fsm.CoordinateMachine(child)
			if is_end {
				f.report()

				return
			}
		}
		for _, p := range parentSeeds {
			is_end := f.fsm.CoordinateMachine(p)
			if is_end {
				f.report()
				return
			}
		}
	}

	// println("formatted machine")
	// println("h0-seeds")
	// f.fsm.HashNumber2Seeds[0].Display()
	// println("h1-seeds", f.fsm.HashNumber2Seeds[1])
	// f.fsm.HashNumber2Seeds[1].Display()
	// println("h0->p", f.fsm.HashNumber2Priority[0])
	// println("h1->p", f.fsm.HashNumber2Priority[1])
	// h0_p := f.fsm.HashNumber2Priority[0]
	// h1_p := f.fsm.HashNumber2Priority[1]
	// println("0번째 우선순위 해시넘버들", f.fsm.Priority2HashNumberULL[0])
	// println("1번째 우선순위 해시넘버들", f.fsm.Priority2HashNumberULL[1])
	// println("2번째 우선순위 해시넘버들", f.fsm.Priority2HashNumberULL[2])
	// println("h0의 우선순위==", h0_p, "실제", h0_p, "번째 우선순위 해시넘버들==")
	// f.fsm.Priority2HashNumberULL[h0_p].Display()
	// println("h1의 우선순위==", h1_p, "실제", h1_p, "번째 우선순위 해시넘버들==")
	// f.fsm.Priority2HashNumberULL[h1_p].Display()
	// println(f.msgs)
}

func (f *F) report() {
	println("done all trials")
	println("개수", len(f.msgs))
}

func (f *F) runFF(content []interface{}) fuzzinglib.Coverage {
	coverage, err, isPanic, panicMsg := monitor(f.ff, content)
	if isPanic {
		var sb strings.Builder
		inputStr := fuzzinglib.ContentToString(content)
		sb.WriteString("Input: ")
		sb.WriteString(inputStr)
		sb.WriteString(" -> ")
		msg := sb.String() + "Panic: " + panicMsg
		f.Fatal(msg)
	}
	if err != nil {
		var sb strings.Builder
		inputStr := fuzzinglib.ContentToString(content)
		sb.WriteString("Input: ")
		sb.WriteString(inputStr)
		sb.WriteString(" -> ")
		msg := sb.String() + "Error: " + strings.TrimSpace(err.Error())
		f.Fatal(msg)
	}
	return coverage
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

func monitor(run Runner, content []interface{}) (coverage fuzzinglib.Coverage, err error, isPanic bool, panicMsg string) {
	isPanic = false
	panicMsg = ""
	err = nil
	coverage = fuzzinglib.Coverage{}
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
	t := testing.NewT("fuzzing")
	// 배열을 스프레딩
	run(t, content...)
	// 테스팅 결과 핸들링링
	info := t.GetResult()
	if info.Failed {
		err = errors.New(string(info.Output))
	}

	// TODO: 이 함수를, 진짜 커버리지 얻는 함수로 수정하기
	coverage = fuzzinglib.Get_Coverage_of_runner(t, content)

	// println(string(t.output))
	return coverage, err, isPanic, panicMsg
}

// Slice 비교 함수
func isSliceEqual(a, b []fuzzinglib.SupportedType) bool {
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

func main() {
	f := New_F()

	f.Add("apple")
	f.Add("banana")

	// f.fsm.Format_StateMachine("some_func")

	f.Fuzz(func(t *testing.T, orig ...interface{}) {
		v, ok := orig[0].(string)
		if !ok {
			panic("dont match")
		}
		rev := fuzzinglib.Reverse1(v)
		doubleRev := fuzzinglib.Reverse1(rev)
		if v != doubleRev {
			t.Errorf("Before: %q, after: %q", orig, doubleRev)
		}
		if utf8.ValidString(v) && !utf8.ValidString(rev) {
			t.Errorf("Reverse produced invalid UTF-8 string %q", rev)
		}
	})

	// new_seed := fuzzinglib.Seed{
	// 	Gen: 0,
	// }
	// slicetest([]fuzzinglib.Seed{new_seed})
	// println(new_seed)
}

// func slicetest(seeds []fuzzinglib.Seed) {
// 	seeds[0].Gen = 100
// }
