package diff

import (
	"context"
	"go/parser"
	"go/token"
	"strings"
)

// Minimize deletes lines from src while diverges reports that the reduced
// program still disagrees. budget is the number of candidate programs the
// predicate may be asked about. The package clause is kept so the result
// stays recognisable as a Go file.
//
// The predicate has to be deterministic and preserve the divergence shape so
// the reducer does not collapse a runtime bug into an unrelated compile error.
func Minimize(src string, budget int, diverges func(string) bool) string {
	if budget < 1 {
		return src
	}
	lines := strings.Split(strings.TrimSuffix(src, "\n"), "\n")
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

func minimizeDivergence(src string, origGo, origGno Result, opt Options, budget int) string {
	origParses := parses(src)
	return Minimize(src, budget, func(candidate string) bool {
		if Unsupported(candidate) != "" {
			return false
		}
		if origParses && !parses(candidate) {
			return false
		}
		left := RunGo(context.Background(), candidate, opt)
		right := RunGno(candidate, opt)
		lint := RunLint(candidate, opt)
		if left.Kind != origGo.Kind || right.Kind != origGno.Kind {
			return false
		}
		cl, _ := Classify(Program{Source: candidate}, left, right, lint)
		return cl == ClassReal
	})
}

func parses(src string) bool {
	_, err := parser.ParseFile(token.NewFileSet(), "main.go", src, parser.SkipObjectResolution)
	return err == nil
}
