package main

import (
	"errors"
	"strings"
)

// 가상의 함수를 설정하였고, 가상 함수의 설정에 맞추어 입력 따라 각기 다른 커버리지 변환하도록 함
// TODO: 이후 일반화된 함수에서도 돌아가게 하려면, reflect등의 네이티브 코드를 잘 짤것!!
func SomeFunc(inp interface{}) interface{} {
	suffixes := []string{"of", "aw", "cr"}
	found := false

	// inp를 문자열로 변환
	input, ok := inp.(string)
	if !ok {
		return errors.New("invalid input type, expected string")
	}
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

// 에러, 패닉에 따른 크래싱은 검사 x. 단지 실행 라인만 검사.
func Get_PseudoCoverage_of_SomeFunc(inp string) []int {
	// 인풋에 따라 리턴하는 커버리지가 변화하도록 하기
	// 위의 SomeFunc에 대해서만은 제대로 동작하는 커버리지 함수
	covered_lines := []int{}

	suffixes := []string{"of", "aw", "cr"}
	found := false
	covered_lines = append(covered_lines, 10, 11)
	covered_lines = append(covered_lines, 12)

	if strings.Contains(inp, "dd") {
		covered_lines = append(covered_lines, 13)
		if strings.Contains(inp, "ddkd") {
			covered_lines = append(covered_lines, 14)
		} else {
			covered_lines = append(covered_lines, 15, 16, 17)
		}
		covered_lines = append(covered_lines, 18)
	} else if strings.Contains(inp, "e") {
		covered_lines = append(covered_lines, 19)
		if strings.Contains(inp, "d") {
			covered_lines = append(covered_lines, 20)
			if strings.HasPrefix(inp, "st") {
				covered_lines = append(covered_lines, 21)
				if strings.HasPrefix(inp, "start") {
					covered_lines = append(covered_lines, 22)
					if strings.HasSuffix(inp, "end") {
						covered_lines = append(covered_lines, 23)
						return covered_lines
					} else {
						covered_lines = append(covered_lines, 24, 25, 26)
					}

				} else {
					covered_lines = append(covered_lines, 27, 28, 29)
				}
			} else {
				covered_lines = append(covered_lines, 30, 31, 32)
			}
		} else {
			covered_lines = append(covered_lines, 33, 34, 35)
		}
	} else {
		covered_lines = append(covered_lines, 36)
		for _, suffix := range suffixes {
			covered_lines = append(covered_lines, 37, 38, 42)
			if strings.HasSuffix(inp, suffix) {
				covered_lines = append(covered_lines, 39, 40, 41)
				found = true
				break
			}
		}
		covered_lines = append(covered_lines, 43)
		if found {
			covered_lines = append(covered_lines, 44)
			if strings.HasPrefix(inp, "Err") {
				covered_lines = append(covered_lines, 45)
				if strings.HasPrefix(inp, "Error") {
					covered_lines = append(covered_lines, 46, 47)
				}
			} else {
				covered_lines = append(covered_lines, 48, 49, 50)
			}
			covered_lines = append(covered_lines, 51)
		}
		covered_lines = append(covered_lines, 52)
	}
	covered_lines = append(covered_lines, 53)
	return covered_lines
}

func ListToDict(nums []int) map[int]int {
	countMap := make(map[int]int)

	for _, num := range nums {
		countMap[num]++
	}

	return countMap
}

func MapKeysToList(countMap map[int]int) []int {
	uniqueList := []int{}
	for key := range countMap {
		uniqueList = append(uniqueList, key)
	}
	return uniqueList
}

// func main() {
// 	covered_lines := Get_PseudoCoverage_of_SomeFunc("e_body_d")
// 	covered_dicts := ListToDict(covered_lines)
// 	println("dicts:", covered_dicts)
// 	cover_set := MapKeysToList(covered_dicts)
// 	println("sets:", cover_set)
// 	result := Detect_Crash(SomeFunc, "start_ed_end") // painc input
// 	println("PanicOccurred:", result.PanicOccurred)
// 	println("PanicMessage:", result.PanicMessage)
// 	println("ErrorMessage:", result.Error)
// 	println("Result:", result.Result)
// 	println("-----------------------------------")
// 	result2 := Detect_Crash(SomeFunc, "Errorof") // error input
// 	println("PanicOccurred:", result2.PanicOccurred)
// 	println("PanicMessage:", result2.PanicMessage)
// 	println("ErrorMessage:", result2.Error)
// 	println("Result:", result2.Result)
// }
