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
type TestFunc func() (interface{}, error)

// RunWithMonitor: 함수 실행 중 에러와 panic 감지
func RunWithMonitor(testFunc TestFunc) (result TestResult) {
	// 우선 result선언 우 defer에서 반환시키기
	// panic 감지
	defer func() {
		if r := recover(); r != nil {
			println("recovered panic!!!")
			result.PanicOccurred = true
			result.PanicMessage = r
		}
	}()
	// 함수 실행
	res, err := testFunc()
	result.Result = res
	result.Error = err

	return result
}

// 테스트 함수 1: 정상 실행
func NormalFunction() (interface{}, error) {
	return "Hello, World!", nil
}

// 테스트 함수 2: 에러 반환
func ErrorFunction() (interface{}, error) {
	return nil, errors.New("simulated error")
}

// 테스트 함수 3: panic 발생
func PanicFunction() (interface{}, error) {
	println("Start functionPF")
	panic("something went wrong!")
	println("it must not be done!")
	return 1, nil
}

func main() {
	// 테스트 케이스 목록
	testCases := []TestFunc{
		NormalFunction,
		ErrorFunction,
		PanicFunction,
	}

	// 각 테스트 실행
	for i, testFunc := range testCases {
		println("Test Case\n", i+1)
		result := RunWithMonitor(testFunc)

		if result.PanicOccurred {
			println("  Panic Detected: \n", result.PanicMessage)
		} else if result.Error != nil {
			println("  Error: \n", result.Error)
		} else {
			println("  No Pand E Result:\n", result.Result)
		}
	}
}
