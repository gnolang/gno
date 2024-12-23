package fuzzer

import (
	"errors"
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
}
