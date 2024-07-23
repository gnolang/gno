package gnolang_test

import (
	"os"
	"path"
	"sort"
	"strings"
	"testing"
)

func TestEvalFiles(t *testing.T) {
	dir := "../../tests/files"
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		wantOut, wantErr, wantStacktrace, ok := testData(dir, f)
		if !ok {
			continue
		}
		t.Run(f.Name(), func(t *testing.T) {
			out, err, stacktrace := evalTest("", "", path.Join(dir, f.Name()))

			if wantErr != "" && !strings.Contains(err, wantErr) ||
				wantErr == "" && err != "" {
				t.Fatalf("unexpected error\nWant: %s\n Got: %s", wantErr, err)
			}

			if wantStacktrace != "" && !strings.Contains(stacktrace, wantStacktrace) ||
				wantStacktrace == "" && stacktrace != "" {
				t.Fatalf("unexpected stacktrace\nWant: %s\n Got: %s", wantStacktrace, stacktrace)
			}
			if wantOut != "" && out != wantOut {
				t.Fatalf("unexpected output\nWant: %s\n Got: %s", wantOut, out)
			}
		})
	}
}

// testData returns the expected output and error string, and true if entry is valid.
func testData(dir string, f os.DirEntry) (testOut, testErr, testStacktrace string, ok bool) {
	if f.IsDir() {
		return
	}
	name := path.Join(dir, f.Name())
	if !strings.HasSuffix(name, ".gno") || strings.HasSuffix(name, "_long.gno") {
		return
	}
	buf, err := os.ReadFile(name)
	if err != nil {
		return
	}
	str := string(buf)
	if strings.Contains(str, "// PKGPATH:") {
		return
	}

	res := commentFrom(str, []string{"\n// Output:", "\n// Error:", "\n// Stacktrace:"})

	return res[0], res[1], res[2], true
}

type directive struct {
	delim string
	res   string
	index int
}

// commentFrom returns the comments from s that are between the delimiters.
func commentFrom(s string, delims []string) []string {
	directives := make([]directive, len(delims))
	directivesFound := make([]*directive, 0, len(delims))

	for i, delim := range delims {
		index := strings.Index(s, delim)
		directives[i] = directive{delim: delim, index: index}
		if index >= 0 {
			directivesFound = append(directivesFound, &directives[i])
		}
	}
	sort.Slice(directivesFound, func(i, j int) bool {
		return directivesFound[i].index < directivesFound[j].index
	})

	for i := range directivesFound {
		next := len(s)
		if i != len(directivesFound)-1 {
			next = directivesFound[i+1].index
		}

		directivesFound[i].res = strings.TrimSpace(strings.ReplaceAll(s[directivesFound[i].index+len(directivesFound[i].delim):next], "\n// ", "\n"))
	}

	res := make([]string, len(directives))
	for i, d := range directives {
		res[i] = d.res
	}

	return res
}
