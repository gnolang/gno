package fuzzing

import (
	"unicode/utf8"
)

// TODO: This file implements just pseudo-coverage functions.
// TODO: Not real functions!
// TODO: Must implement a real get-coverage function and replace this file.

// Reverse reverses the input string and returns the reversed string.
func Reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

// GetCoverageOfReverse records coverage information for the Reverse function.
func GetCoverageOfReverse(c *Coverage, s string) string {
	r := []rune(s)
	*c = append(*c, CoveredLine{CoName: "Reverse", CoLine: 37})
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		*c = append(*c, CoveredLine{CoName: "Reverse", CoLine: 39})
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

// byteToHexChar converts a byte to its corresponding hexadecimal character.
// It returns "0-9" for values less than 10 and "a-f" for values 10-15.
func byteToHexChar(b byte) string {
	if b < 10 {
		return string('0' + b) // 0-9
	}
	return string('a' + (b - 10)) // a-f
}

// TODO: This is a pseudo-coverage function. Must replace with a real coverage function in the future.

// GetCoverageOfrunner is a pseudo-coverage function. It must be replaced with a real coverage function in the future.
func GetCoverageOfrunner(content []interface{}) Coverage {
	// TODO: Ensure to get coverage.
	// TODO: The format is (function name, line)
	if len(content) == 1 {
		coverage := Coverage{}
		s, ok := content[0].(string)
		if !ok {
			panic("not a string")
		}
		coverage = append(coverage, CoveredLine{CoName: "symbolicPath", CoLine: 1})
		if len(s) > 0 && s[0] == 'b' {
			coverage = append(coverage, CoveredLine{CoName: "symbolicPath", CoLine: 2})
			if len(s) > 1 && s[1] == 'a' {
				coverage = append(coverage, CoveredLine{CoName: "symbolicPath", CoLine: 3})

				if len(s) > 2 && s[2] == 'd' {
					coverage = append(coverage, CoveredLine{CoName: "symbolicPath", CoLine: 4})

					if len(s) > 3 && s[3] == '!' {
						coverage = append(coverage, CoveredLine{CoName: "symbolicPath", CoLine: 5})
					}
				}
			}
		}
		return coverage

	}

	coverage := Coverage{}
	coverage = append(coverage, CoveredLine{CoName: "closure", CoLine: 13})
	v, ok := content[0].(string)
	if !ok {
		panic("did not implement other type")
	}
	orig := string(v)
	u, ok2 := content[1].(int)
	if !ok2 {
		panic("did not implement other type")
	}
	s1 := GetCoverageOfReverse(&coverage, orig)
	s2 := GetCoverageOfReverse(&coverage, s1)

	if orig != s2 && u > 300 && u < 1000 {

		coverage = append(coverage, CoveredLine{CoName: "closure", CoLine: 21})
		return coverage
	}

	if utf8.ValidString(orig) && !utf8.ValidString(s1) && u > 300 && u < 1000 {
		coverage = append(coverage, CoveredLine{CoName: "closure", CoLine: 24})
		return coverage
	}
	return coverage
}
