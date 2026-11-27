package diff

import "strings"

// Minimize deletes lines from src while diverges reports that the reduced
// program still disagrees. budget is the number of candidate programs the
// predicate may be asked about. The package clause is kept so the result
// stays recognisable as a Go file.
//
// The predicate has to be deterministic. A candidate that fails to compile
// on both sides is not a divergence and is discarded, which stops the
// reducer from eating the statement that caused the disagreement.
func Minimize(src string, budget int, diverges func(string) bool) string {
	if budget < 1 {
		return src
	}
	lines := strings.Split(src, "\n")
	for budget > 0 {
		progressed := false
		for i := 0; i < len(lines); i++ {
			if budget <= 0 {
				break
			}
			if strings.HasPrefix(strings.TrimSpace(lines[i]), "package ") {
				continue
			}
			trial := make([]string, 0, len(lines)-1)
			trial = append(trial, lines[:i]...)
			trial = append(trial, lines[i+1:]...)
			budget--
			if !diverges(strings.Join(trial, "\n")) {
				continue
			}
			lines = trial
			progressed = true
			i--
		}
		if !progressed {
			break
		}
	}
	out := strings.Join(lines, "\n")
	if !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	return out
}
