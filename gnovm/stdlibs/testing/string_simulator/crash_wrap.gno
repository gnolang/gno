package main

import (
	"errors"
)

// TestResult 구조체: 실행 결과 저장
type TestResult struct {
	PanicOccurred bool        // panic 여부
	PanicMessage  interface{} // panic 메시지
	Error         error       // 반환된 에러
	Result        interface{} // 반환된 값
}

// TestFunc 타입: 테스트할 함수 시그니처
type TestFunc func(...interface{}) (interface{}, error)

// RunWithMonitor: 함수 실행 중 에러와 panic 감지

// WrapFunction: 다양한 함수 타입을 래핑하여 TestFunc으로 변환
// 상당 부분 컨벤션에 의존
// 솔직히 panic만 잡아내면 상관이 없는데
// go <- 얘는 err를 명시적 처리로 하는거라서
// 내가 그걸 잡아야 함. 근데 그건 뭐 구문 분석의 영역이라..
// 어캐든 err 잡아내려면, 지금으로썬 컨벤션에 의지해야 함.
// err타입을 의존
// 아 걍 3중 반복 떄리면서 primitive타입 거렁야 겠는데???

func WrapFunction(f interface{}) TestFunc {
	return func(inputs ...interface{}) (interface{}, error) {
		if len(inputs) != 1 {
			return nil, errors.New("invalid number of arguments")
		}
		switch fn := f.(type) {
		// 가정을 두었음. err는 반드시 (r,e)타입이거나 (e)로만 나와야 함
		// int가 error리턴 시
		case func(int) error:
			val, ok := inputs[0].(int)
			if !ok {
				return nil, errors.New("invalid input type")
			}
			return nil, fn(val)
			// int가 (r,e) 컨벤션의 에러 리턴 시
		case func(int) (int, error):
			val, ok := inputs[0].(int)
			if !ok {
				return nil, errors.New("invalid input type")
			}
			return fn(val)

		// 그 외엔 값 리턴이라 가정
		// 이 경우 panic만 검사하면 됨.
		case func(int) int:
			val, ok := inputs[0].(int)
			if !ok {
				return nil, errors.New("invalid input type")
			}

			return fn(val), nil
			///대충 이짓거릴 3중 반복

		case func() (interface{}, error):
			return fn()
		default:
			return nil, errors.New("unsupported function type")
		}
	}
}

// 테스트 함수들
func Double(x int) int {
	return x * 2
}

func Concat(s string) string {
	return s + " is awesome!"
}

func ErrorFunction(x string) (interface{}, error) {
	println(x)
	return 1, errors.New("simulated error")
}

func notConventionError(x interface{}) interface{} {
	println(x)
	return errors.New("not conventionalError. not tuple")
}

func PanicFunction(x int) int {
	println(x)
	panic("something went wrong!")
	return 3
}

func CollectResults_then_CheckErr(r ...interface{}) (bool, error, []interface{}) {
	var results []interface{}
	var is_err bool = false
	var err_content error = nil

	for _, value := range r {
		if err, ok := value.(error); ok {
			println("Error detected:", err)
			err_content = err
			is_err = true
			return is_err, err_content, results
		} else {
			results = append(results, value)
		}
	}

	return is_err, err_content, results
}

// err, panic 전부 탐지하도록 하기
func Detect_Crash(testFunc interface{}, inp interface{}) (result TestResult) {
	// input-> panic은 여기서 거름
	defer func() {
		if r := recover(); r != nil {
			result.PanicOccurred = true
			result.PanicMessage = r
		}
	}()
	result.PanicOccurred = false
	result.PanicMessage = nil

	if f, ok := testFunc.(func(interface{}) interface{}); ok {
		isErr, errContent, results := CollectResults_then_CheckErr(f(inp))
		result.Result = results
		if isErr {
			result.Error = errContent
		} else {
			result.Error = nil
		}
	} else {
		result.Error = errors.New("testFunc is not of type func(interface{}) interface{}")
	}
	return result
}

// func CaseTest() {
// 	// 테스트 케이스 목록
// 	testCases := []struct {
// 		fn    interface{}
// 		input []interface{}
// 	}{
// 		{Double, []interface{}{10}},
// 		{Concat, []interface{}{"Go programming"}},
// 		{ErrorFunction, []interface{}{"Go programming"}},
// 		{notConventionError, []interface{}{2}},
// 		{PanicFunction, []interface{}{11}},
// 	}

// 	// 각 테스트 실행
// 	for i, testCase := range testCases {
// 		println("Test Case %d:\n", i+1)
// 		result := Detect_Crash(testCase.fn, testCase.input...)

// 		if result.PanicOccurred {
// 			println("  Panic Detected: %v\n", result.PanicMessage)
// 		} else if result.Error != nil {
// 			println("  Error: %v\n", result.Error)
// 		} else {
// 			switch v := result.Result.(type) {
// 			case string:
// 				println("  Result (string): %v\n", v)
// 			case int:
// 				println("  Result (int): %v\n", v)
// 			default:
// 				println("  Result (unknown type): %v\n", v)
// 			}
// 		}
// 	}
// }
