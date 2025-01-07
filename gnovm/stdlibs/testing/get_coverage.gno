package testing

import (
	"unicode/utf8"
)

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

func getCoverageOfReverse1(c *Coverage, s string) string {
	r := []rune(s)
	*c = append(*c, coveredLine{coName: "Reverse1", coLine: 37})
	*c = append(*c, coveredLine{coName: "Reverse1", coLine: 38})
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		*c = append(*c, coveredLine{coName: "Reverse1", coLine: 39})
		r[i], r[j] = r[j], r[i]
		*c = append(*c, coveredLine{coName: "Reverse1", coLine: 40})
		*c = append(*c, coveredLine{coName: "Reverse1", coLine: 41})
	}
	*c = append(*c, coveredLine{coName: "Reverse1", coLine: 39})
	*c = append(*c, coveredLine{coName: "Reverse1", coLine: 42})
	return string(r)
}

func byteToHexChar(b byte) string {
	if b < 10 {
		return string('0' + b) // 0-9
	}
	return string('a' + (b - 10)) // a-f
}

func getCoverageofrunner(t *T, content []interface{}) Coverage {
	// TODO: Make sure get coverage.
	// TODO: The format is (function name, line)

	coverage := Coverage{}
	coverage = append(coverage, coveredLine{coName: "closure", coLine: 13})
	coverage = append(coverage, coveredLine{coName: "closure", coLine: 14})
	coverage = append(coverage, coveredLine{coName: "closure", coLine: 15})
	v, ok := content[0].(string)
	if !ok {
		coverage = append(coverage, coveredLine{coName: "closure", coLine: 16})
		coverage = append(coverage, coveredLine{coName: "closure", coLine: 17})

		panic("did not implement other type")
	}
	orig := string(v)
	u, ok2 := content[1].(int)
	if !ok2 {
		panic("did not implement other type")
	}
	coverage = append(coverage, coveredLine{coName: "closure", coLine: 18})
	s1 := getCoverageOfReverse1(&coverage, orig)
	coverage = append(coverage, coveredLine{coName: "closure", coLine: 19})
	s2 := getCoverageOfReverse1(&coverage, s1)

	coverage = append(coverage, coveredLine{coName: "closure", coLine: 20})
	if orig != s2 && u > 300 && u < 1000 {

		// println("orig=", orig, "doublereverse", s2)
		coverage = append(coverage, coveredLine{coName: "closure", coLine: 21})
		return coverage
	}
	coverage = append(coverage, coveredLine{coName: "closure", coLine: 22})

	coverage = append(coverage, coveredLine{coName: "closure", coLine: 23})
	if utf8.ValidString(orig) && !utf8.ValidString(s1) && u > 300 && u < 1000 {
		coverage = append(coverage, coveredLine{coName: "closure", coLine: 24})
		return coverage
	}
	coverage = append(coverage, coveredLine{coName: "closure", coLine: 25})
	coverage = append(coverage, coveredLine{coName: "closure", coLine: 26})
	return coverage
}
