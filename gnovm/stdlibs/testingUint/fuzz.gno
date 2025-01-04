package testingUint

import "math/rand"

// 아마도 여기가 퍼징 테스트 유닛이라고 님이 만든 듯?
// 일단 유지는 시켜 둘게. 다만 지금은 사용 못함. 커버리지 때매 완전한 테스트는 불가해서.
// 대신 testing/mock에 함수 대한 테스팅 추가해 놓았음.
// structs

//! Notice. I didn't decided whether to use the mudular, so I made it both and see the latency of the test.

// * minRange <= data used in fuzzing < maxRange
type configValues struct {
	minRange    uint64 // the min boundary of the random values
	maxRange    uint64 // the max boundary of the random values
	breakdown   bool   // if there are unexpected operations during fuzzing, stop the fuzz and return that result.
	maxIter     uint64 // max iteratation number for the fuzzer to fuzz.
	saveSuccess bool   // save the success tests
	saveFail    bool   // save the Failed tests
}

// * sorts the range values by using modular.
type configValuesMod struct {
	minRange  uint64 // the min boundary of the random values
	maxRange  uint64 // the max boundary of the random values. It's the mod p value.
	breakdown bool   // if there are unexpected operations during fuzzing, stop the fuzz and return that result.
	maxIter   uint64 // max iteratation number for the fuzzer to fuzz.
}

type Context struct {
	randomDataSet []uint64 // Dataset of the random uint value. For more information, read the howToTest.md - How data is made chapter.
	failed        bool     // Indicates whether the fuzzing has encountered a failure.
	msgs          []string // Stores log messages for reporting.
	iters         int      // Number of iterations to run the fuzzing process. It keeps increases during fuzzing, and you can use it by searching this value in the Result file to find the failed log of the fuzz test.
}

type targetFunc func(params ...interface{}) // the target function which we want to fuzz.

// Container func
func FuzzUint() {
}
