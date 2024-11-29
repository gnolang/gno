package main

// 커링을 위한 함수 타입 정의
type CurryFunc func(...int) int

// curry 함수: n-인자 함수를 커링 가능하게 변환
// func curry(fun func(int, int, int) int) func(args ...int) interface{} {
// 	// 함수를 리턴함
// 	return func(args ...int) interface{} {
// 		// 리턴된 함수는, 인풋 길이가 초과 시 값 반환
// 		if len(args) >= 3 {
// 			return fun(args[0], args[1], args[2])
// 		}

// 		// 리턴된 함수는
// 		///인자를 통해 호출 받으면 다시
// 		return func(nextArgs ...int) interface{}
// 		{
// 			return curry(fun)(append(args, nextArgs...)...)
// 		}
// 	}
// }

// sum 함수
func sum(a, b, c int) int {
	return a + b + c
}

// func main() {
// 	// sum 함수에 커링 적용
// 	curriedSum := curry(sum)

// 	// 호출 1: 모든 인수를 한 번에 전달
// 	result1 := curriedSum(1, 2, 3).(int)
// 	println("Result 1:", result1) // 6

// 	// 호출 2: 커링으로 부분적으로 전달
// 	result2 := curriedSum(1).(func(...int) interface{})(2, 3).(int)
// 	println("Result 2:", result2) // 6

// 	// 호출 3: 모든 인수를 하나씩 전달
// 	result3 := curriedSum(1).(func(...int) interface{})(2).(func(...int) interface{})(3).(int)
// 	println("Result 3:", result3) // 6
// }
