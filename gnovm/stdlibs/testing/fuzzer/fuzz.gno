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

func pshedo_typeof(arg interface{}) (fuzzinglib.SupportedType, error) {
	pshedoType := "string"
	s_type, flag := isSupportedType(pshedoType)
	if !flag {
		return s_type, errors.New("not supported type")
	}
	return s_type, nil
}

func isSupportedType(value string) (fuzzinglib.SupportedType, bool) {
	switch fuzzinglib.SupportedType(value) {
	case fuzzinglib.Byte_Array, fuzzinglib.String, fuzzinglib.Bool, fuzzinglib.Byte, fuzzinglib.Rune, fuzzinglib.Float32, fuzzinglib.Float64,
		fuzzinglib.Int, fuzzinglib.Int8, fuzzinglib.Int16, fuzzinglib.Int32, fuzzinglib.Int64, fuzzinglib.Uint, fuzzinglib.Uint8, fuzzinglib.Uint16, fuzzinglib.Uint32, fuzzinglib.Uint64:
		return fuzzinglib.SupportedType(value), true
	default:
		return "", false
	}
}

// 일단 다변수함수를 지원하게는 해 놓았지만,
// 현재로썬 걍 type string만 반환하는 상태.
// 하나의 시드를 추가하는 코드.
func (f *F) Add(args ...interface{}) {
	var values []interface{}
	var types []fuzzinglib.SupportedType

	for i := range args {
		t, e := pshedo_typeof(args[i])
		if e != nil {
			panic("not supported type")
		}
		values = append(values, args[i])
		types = append(types, t)
	}
	f.fsm.InitSeeds = append(f.fsm.InitSeeds, fuzzinglib.Seed{Gen: 1, IsUsed: false, Content: values})
}

func main() {
	f := New_F()
	f.fsm.Format_StateMachine("some_func")
	println(f.fsm.ToString())
}
