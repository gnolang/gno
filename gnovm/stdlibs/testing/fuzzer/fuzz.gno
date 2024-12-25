package fuzzer

import (
	"errors"
	"strings"
	"testing"
	"testing/fuzzinglib"
)

// type F struct {
// 	common
// 	fstate *fuzzState
// 	tstate *testState

// 	// inFuzzFn is true when the fuzz function is running. Most F methods cannot
// 	// be called when inFuzzFn is true.
// 	inFuzzFn bool

// 	// corpus is a set of seed corpus entries, added with F.Add and loaded
// 	// from testdata.
// 	corpus []corpusEntry

// 	result     fuzzResult
// 	fuzzCalled bool
// }

type F struct {
	fsm    *fuzzinglib.StateMachine
	corpus []fuzzinglib.Seed
	msgs   []string // Stores log messages for reporting.
	iters  int      // Number of iterations to run the fuzzing process. TODO: CLI flag to set this.
	failed bool     // Indicates whether the fuzzing has encountered a failure.
}

func New_F() *F {
	new_fsm := fuzzinglib.New_StateMachine()
	return &F{
		fsm: new_fsm,
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

	// 타입 검사 및 값 적층층
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

	// 초기화용 시드에 추가
	f.fsm.InitSeeds = append(f.fsm.InitSeeds, fuzzinglib.Seed{Gen: 1, IsUsed: false, Content: values})
}

type Runner func(*testing.T, ...interface{})

func (f *F) Fuzz(run Runner) {
	_, err, isPanic, panicMsg := monitor(run, "apple")
	if isPanic {
		msg := "Panic occered in target function: " + panicMsg
		f.Fatal(msg)
	}
	if err != nil {

		msg := "Error occered in target function: " + err.Error()
		f.Fatal(msg)

	}
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
	println(f.msgs[0])
	f.Fail()
}

func monitor(run Runner, inputs ...interface{}) (coverage fuzzinglib.Coverage, err error, isPanic bool, panicMsg string) {
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
	run(t, inputs)
	v, ok := inputs[0].(string)
	if !ok {
		panic("didnt implement other type")
	}
	coverage = fuzzinglib.Get_Coverage_of_runner(t, v)

	info := t.GetResult()
	if info.Failed {
		err = errors.New(string(info.Output))
	}

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
	var s []byte = []byte("hell")
	var z []byte = []byte("apple")
	f.Add("strrr", s)
	f.Add("strrr22", z)
	f.Add("sadaasd", s)
	f.fsm.Format_StateMachine("some_func")
	println(f.fsm.ToString())
	println(f.fsm.SeedType)
	println(f.fsm.InitSeeds)
	f.Fuzz(func(t *testing.T, a ...interface{}) {
		println(a[0])
		panic("크아아악")
	})
	println("hello world")
}
