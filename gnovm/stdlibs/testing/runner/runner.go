package runner

import (
	"reflect"
)

// TestResult 구조체 정의
type TestResult struct {
	PanicOccurred bool          // panic 여부
	PanicMessage  interface{}   // panic 메시지
	Error         error         // 반환된 에러
	Result        []interface{} // 반환된 값 (슬라이스 형태)
}

// CollectResults_then_CheckErr 함수 정의
func CollectResults_then_CheckErr(results []reflect.Value) (bool, []interface{}, error) {
	outputs := make([]interface{}, len(results))
	for i, result := range results {
		outputs[i] = result.Interface()
	}

	var err error
	if len(outputs) > 0 {
		// 마지막 반환값이 error 타입인지 확인
		if lastResult, ok := outputs[len(outputs)-1].(error); ok {
			err = lastResult
			outputs = outputs[:len(outputs)-1] // error는 Result에서 제외
		}
	}
	return err != nil, outputs, err
}

// Panic 및 Error 탐지 함수
func Detect_Crash(fn interface{}, args ...interface{}) (result TestResult) {
	// Panic 처리
	defer func() {
		if r := recover(); r != nil {
			result.PanicOccurred = true
			result.PanicMessage = r
		}
	}()
	result.PanicOccurred = false
	result.PanicMessage = nil

	// 리플렉션을 통해 fn 호출
	fnValue := reflect.ValueOf(fn)
	if fnValue.Kind() != reflect.Func {
		panic("fn is not a function")
	}

	// 인자 준비
	in := make([]reflect.Value, len(args))
	for i, arg := range args {
		in[i] = reflect.ValueOf(arg)
	}

	// 함수 실행 및 결과 처리
	isErr, results, errContent := CollectResults_then_CheckErr(fnValue.Call(in))
	result.Result = results
	if isErr {
		result.Error = errContent
	} else {
		result.Error = nil
	}

	return result
}

// 모니터링 함수 실행
func X_monitorRuning(fn interface{}, args ...interface{}) TestResult {
	return Detect_Crash(fn, args...)
}

// func add(a, b int) int {
// 	return a + b
// }

// // 테스트용 함수 (panic 발생)
// func causePanic() {
// 	panic("Something went wrong!")
// }

// // 테스트용 함수 (error 반환)
// func returnError(a int) (string, error) {
// 	if a < 0 {
// 		return "", fmt.Errorf("negative value: %d", a)
// 	}
// 	return fmt.Sprintf("Value: %d", a), nil
// }

// func main() {
// 	// 정상 동작 테스트
// 	result1 := X_monitorRuning(add, 1, 2)
// 	fmt.Println("Result1:", result1.Result)               // [3]
// 	fmt.Println("Error1:", result1.Error)                 // <nil>
// 	fmt.Println("PanicOccurred1:", result1.PanicOccurred) // false

// 	// Panic 발생 테스트
// 	result2 := X_monitorRuning(causePanic)
// 	fmt.Println("PanicOccurred2:", result2.PanicOccurred) // true
// 	fmt.Println("PanicMessage2:", result2.PanicMessage)   // Something went wrong!

// 	// Error 반환 테스트
// 	result3 := X_monitorRuning(returnError, -5)
// 	fmt.Println("Result3:", result3.Result)               // []
// 	fmt.Println("Error3:", result3.Error)                 // negative value: -5
// 	fmt.Println("PanicOccurred3:", result3.PanicOccurred) // false
// }
