package testing

import (
	"errors"
	"unicode/utf8"
)

// TODO: This file is currently a virtual coverage acquisition file
// TODO: To clear one day, replace it with a real coverage function.
// l
// l
// l
func mock(t *T, orig ...interface{}) {
	v, ok := orig[0].(string)
	if !ok {
		panic("dont match")
	}
	rev := Reverse1(v)
	doubleRev := Reverse1(rev)
	if v != doubleRev && v == "some cond" {
		t.Errorf("Before: %q, after: %q", orig, doubleRev)
	}
	if utf8.ValidString(v) && !utf8.ValidString(rev) && v == "some cond" {
		t.Errorf("Reverse produced invalid UTF-8 string %q", rev)
	}
}

// l
// l
// l
// l
// l
// l
// l
// l
// l
func Reverse1(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

func Get_Coverage_of_Reverse1(c *Coverage, s string) string {
	r := []rune(s)
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

func byteToHexChar(b byte) string {
	if b < 10 {
		return string('0' + b) // 0-9
	}
	return string('a' + (b - 10)) // a-f
}

func Get_Coverage_of_runner(t *T, content []interface{}) Coverage {
	// TODO: Make sure get coverage.
	// TODO: The format is (function name, line)

	coverage := Coverage{}
	coverage = append(coverage, CoveredLine{co_name: "closure", co_line: 13})
	coverage = append(coverage, CoveredLine{co_name: "closure", co_line: 14})
	coverage = append(coverage, CoveredLine{co_name: "closure", co_line: 15})
	v, ok := content[0].(string)
	if !ok {
		coverage = append(coverage, CoveredLine{co_name: "closure", co_line: 16})
		coverage = append(coverage, CoveredLine{co_name: "closure", co_line: 17})

		panic("did not implement other type")
	}
	orig := string(v)
	u, ok2 := content[1].(int)
	if !ok2 {
		panic("did not implement other type")
	}
	coverage = append(coverage, CoveredLine{co_name: "closure", co_line: 18})
	s1 := Get_Coverage_of_Reverse1(&coverage, orig)
	coverage = append(coverage, CoveredLine{co_name: "closure", co_line: 19})
	s2 := Get_Coverage_of_Reverse1(&coverage, s1)

	coverage = append(coverage, CoveredLine{co_name: "closure", co_line: 20})
	if orig != s2 && u > 300 && u < 1000 {

		// println("orig=", orig, "doublereverse", s2)
		coverage = append(coverage, CoveredLine{co_name: "closure", co_line: 21})
		return coverage
	}
	coverage = append(coverage, CoveredLine{co_name: "closure", co_line: 22})

	coverage = append(coverage, CoveredLine{co_name: "closure", co_line: 23})
	if utf8.ValidString(orig) && !utf8.ValidString(s1) && u > 300 && u < 1000 {
		coverage = append(coverage, CoveredLine{co_name: "closure", co_line: 24})
		return coverage
	}
	coverage = append(coverage, CoveredLine{co_name: "closure", co_line: 25})
	coverage = append(coverage, CoveredLine{co_name: "closure", co_line: 26})
	return coverage
}

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
