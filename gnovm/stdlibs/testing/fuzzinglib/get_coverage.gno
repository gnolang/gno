package fuzzinglib

import (
	"errors"
	"strings"
)

// 가상의 함수를 설정하였고, 가상 함수의 설정에 맞추어 입력 따라 각기 다른 커버리지 변환하도록 함
// TODO: 이후 일반화된 함수에서도 돌아가게 하려면, reflect등의 네이티브 코드를 잘 짤것!!
func SomeFunc(input string) interface{} {
	suffixes := []string{"of", "aw", "cr"}
	found := false
	if strings.Contains(input, "dd") {
		if strings.Contains(input, "ddkd") {
			println("고착화 통과")
		} else {
			println("one_depth")
		}
	} else if strings.Contains(input, "e") {
		if strings.Contains(input, "d") {
			if strings.HasPrefix(input, "st") {
				if strings.HasPrefix(input, "start") {
					if strings.HasSuffix(input, "end") {
						panic("five_depth panic!")
					} else {
						println("four_depth")
					}
				} else {
					println("three_depth")
				}
			} else {
				println("two_depth")
			}
		} else {
			println("one_depth")
		}
	} else {
		for _, suffix := range suffixes {
			if strings.HasSuffix(input, suffix) {
				found = true
				break
			}
		}
		if found {
			if strings.HasPrefix(input, "Err") {
				if strings.HasPrefix(input, "Error") {
					return errors.New("simulated error")
				}
			} else {
				println("one_depth")
			}
		}
	}
	return nil
}

func get_counter(start int) func() int {
	count := start
	return func() int {
		count += 1
		return count
	}
}

func Get_AllCoverage() Coverage {
	// 배열 생성
	var allLines Coverage
	for i := 10; i <= 54; i++ {
		allLines = append(allLines, uint(i))
	}
	return allLines
}

// SomeFunc의 커버리지 계산 함수
func Get_PseudoCoverage_of_SomeFunc(inp string) (result TestResult) {
	result.PanicOccurred = false
	result.PanicMessage = "no panic"
	result.Error = nil
	result.Coverage = []uint{}

	// 인풋에 따라 리턴하는 커버리지가 변화하도록 하기
	// 위의 SomeFunc에 대해서만은 제대로 동작하는 커버리지 함수

	suffixes := []string{"of", "aw", "cr"}
	found := false

	// 기본적으로 커버된 라인에 초기 값을 추가
	result.Coverage = append(result.Coverage, 10, 11, 12, 13)
	if strings.Contains(inp, "dd") {
		result.Coverage = append(result.Coverage, 14)
		if strings.Contains(inp, "ddkd") {
			result.Coverage = append(result.Coverage, 15)
		} else {
			result.Coverage = append(result.Coverage, 16, 17, 18)
		}
		result.Coverage = append(result.Coverage, 19)
	} else if strings.Contains(inp, "e") {
		result.Coverage = append(result.Coverage, 20)
		if strings.Contains(inp, "d") {
			result.Coverage = append(result.Coverage, 21)
			if strings.HasPrefix(inp, "st") {
				result.Coverage = append(result.Coverage, 22)
				if strings.HasPrefix(inp, "start") {
					result.Coverage = append(result.Coverage, 23)
					if strings.HasSuffix(inp, "end") {
						result.Coverage = append(result.Coverage, 24)
						result.PanicOccurred = true
						result.PanicMessage = "five_depth panic!"
						return result
					} else {
						result.Coverage = append(result.Coverage, 25, 26, 27)
					}
				} else {
					result.Coverage = append(result.Coverage, 28, 29, 30)
				}
			} else {
				result.Coverage = append(result.Coverage, 31, 32, 33)
			}
		} else {
			result.Coverage = append(result.Coverage, 34, 35, 36)
		}
	} else {
		result.Coverage = append(result.Coverage, 37)
		for _, suffix := range suffixes {
			result.Coverage = append(result.Coverage, 38, 39, 43)
			if strings.HasSuffix(inp, suffix) {
				result.Coverage = append(result.Coverage, 40, 41, 42)
				found = true
				break
			}
		}
		result.Coverage = append(result.Coverage, 44)
		if found {
			result.Coverage = append(result.Coverage, 45)
			if strings.HasPrefix(inp, "Err") {
				result.Coverage = append(result.Coverage, 46)
				if strings.HasPrefix(inp, "Error") {
					result.Error = errors.New("simulated error")
					result.Coverage = append(result.Coverage, 47, 48)
					return result
				}
			} else {
				result.Coverage = append(result.Coverage, 49, 50, 51)
			}
			result.Coverage = append(result.Coverage, 52)
		}
		result.Coverage = append(result.Coverage, 53)
	}
	result.Coverage = append(result.Coverage, 54)
	return result
}

// uint 슬라이스를 맵으로 변환
func ListToDict(nums []uint) map[uint]uint {
	countMap := make(map[uint]uint)

	for _, num := range nums {
		countMap[num]++
	}

	return countMap
}

// 맵의 키를 슬라이스로 변환
func MapKeysToList(countMap map[uint]uint) []uint {
	uniqueList := []uint{}
	for key := range countMap {
		uniqueList = append(uniqueList, key)
	}
	return uniqueList
}

// func main() {
// 	result := Get_PseudoCoverage_of_SomeFunc("Error_of")
// 	coveredDicts := ListToDict(result.Coverage)

// 	println("Coverage Map:")
// 	for k, v := range coveredDicts {
// 		println(k, ":", v)
// 	}
// 	println("Panic Message:", result.PanicMessage)
// 	println("Panic Occurred:", result.PanicOccurred)
// 	if result.Error != nil {
// 		println("Error:", result.Error.Error())
// 	} else {
// 		println("Error: nil")
// 	}
// }
