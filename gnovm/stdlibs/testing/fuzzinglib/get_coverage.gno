package fuzzinglib

import (
	"errors"
	"testing"
	"unicode/utf8"
)

func Get_AllCoverage() Coverage {
	return Coverage{
		{"closure", 13},
		{"closure", 14},
		{"Reverse1", 37},
		{"Reverse1", 38},
		{"Reverse1", 39},
		{"Reverse1", 40},
		{"Reverse1", 41},
		{"Reverse1", 42},
		{"closure", 15},
		{"closure", 16},
		{"closure", 17},
		{"closure", 18},
		{"closure", 19},
		{"closure", 20},
		{"closure", 21},
		{"closure", 22},
		{"closure", 23},
	}
}

// TODO: 리버스, 리버스 관련 처리도 원시 모델로 처리하기. 점진적 수정 시뮬 필요.
func Reverse3(s string) (string, error) {
	if !utf8.ValidString(s) {
		return s, errors.New("input is not valid UTF-8")
	}
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r), nil
}

func Reverse2(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

func Reverse1(s string) string {
	r := []byte(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

func Get_Coverage_of_runner(t *testing.T, orig string) Coverage {
	// 커버리지 받아오도록 하기.
	// 형식은 (함수명, line)

	coverage := Coverage{}
	coverage = append(coverage, CoveredLine{co_name: "closure", co_line: 13})
	coverage = append(coverage, CoveredLine{co_name: "closure", co_line: 14})
	s1 := Get_Coverage_of_Reverse1(&coverage, orig)
	coverage = append(coverage, CoveredLine{co_name: "closure", co_line: 15})
	s2 := Get_Coverage_of_Reverse1(&coverage, orig)
	coverage = append(coverage, CoveredLine{co_name: "closure", co_line: 16})
	if orig != s2 {
		coverage = append(coverage, CoveredLine{co_name: "closure", co_line: 17})
		return coverage
	}
	coverage = append(coverage, CoveredLine{co_name: "closure", co_line: 18})
	coverage = append(coverage, CoveredLine{co_name: "closure", co_line: 19})
	if utf8.ValidString(orig) && !utf8.ValidString(s1) {
		coverage = append(coverage, CoveredLine{co_name: "closure", co_line: 20})
		return coverage
	}
	coverage = append(coverage, CoveredLine{co_name: "closure", co_line: 21})
	coverage = append(coverage, CoveredLine{co_name: "closure", co_line: 22})
	return coverage
}

func Get_Coverage_of_Reverse1(c *Coverage, s string) string {
	r := []byte(s)
	*c = append(*c, CoveredLine{co_name: "Reverse1", co_line: 37})
	*c = append(*c, CoveredLine{co_name: "Reverse1", co_line: 38})
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		*c = append(*c, CoveredLine{co_name: "Reverse1", co_line: 39})
		r[i], r[j] = r[j], r[i]
		*c = append(*c, CoveredLine{co_name: "Reverse1", co_line: 40})
		*c = append(*c, CoveredLine{co_name: "Reverse1", co_line: 41})
	}
	*c = append(*c, CoveredLine{co_name: "Reverse1", co_line: 39})
	*c = append(*c, CoveredLine{co_name: "Reverse1", co_line: 42})
	return string(r)
}

// func main() {
// 	coverage1 := Get_Coverage_of_target_func("ssss")
// 	coverage2 := Get_Coverage_of_target_func("ǁ")
// 	fmt.Println("Coverage1:")
// 	for i, c := range coverage1 {
// 		fmt.Printf("  covered line %d: co_name = %q, co_line = %d\n", i, c.co_name, c.co_line)
// 	}

// 	// coverage2 출력
// 	fmt.Println("Coverage2:")
// 	for i, c := range coverage2 {
// 		fmt.Printf("  covered line %d: co_name = %q, co_line = %d\n", i, c.co_name, c.co_line)

// 	} // 두 배열의 차이(diff) 계산
// 	fmt.Println("\nDifferences:")

// 	// coverage1에만 있는 항목
// 	fmt.Println("In Coverage1 but not in Coverage2:")
// 	for _, c1 := range coverage1 {
// 		found := false
// 		for _, c2 := range coverage2 {
// 			if c1.co_name == c2.co_name && c1.co_line == c2.co_line {
// 				found = true
// 				break
// 			}
// 		}
// 		if !found {
// 			fmt.Printf("  co_name = %q, co_line = %d\n", c1.co_name, c1.co_line)
// 		}
// 	}

// 	// coverage2에만 있는 항목
// 	fmt.Println("\nIn Coverage2 but not in Coverage1:")
// 	for _, c2 := range coverage2 {
// 		found := false
// 		for _, c1 := range coverage1 {
// 			if c1.co_name == c2.co_name && c1.co_line == c2.co_line {
// 				found = true
// 				break
// 			}
// 		}
// 		if !found {
// 			fmt.Printf("  co_name = %q, co_line = %d\n", c2.co_name, c2.co_line)
// 		}
// 	}

// input := "The quick brown fox jumped over the lazy dog"
// rev := Reverse1(input)
// doubleRev := Reverse1(rev)
// fmt.Printf("original: %q\n", input)
// fmt.Printf("original: %q\n", rev)
// fmt.Printf("original: %q\n", doubleRev)
//fmt.Printf("reversed: %q, err: %v\n", rev, revErr)
//fmt.Printf("reversed again: %q, err: %v\n", doubleRev, doubleRevErr)
//}
