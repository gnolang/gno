package gnolang_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func TestEvalFiles(t *testing.T) {
	dir := "../../tests/files"
	err := fs.WalkDir(os.DirFS(dir), ".", func(path string, de fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case path == "extern":
			return fs.SkipDir
		case de.IsDir():
			return nil
		}

		fullPath := filepath.Join(dir, path)
		wantOut, wantErr, wantStacktrace, ok := testData(fullPath)
		if !ok {
			return nil
		}

		t.Run(path, func(t *testing.T) {
			out, err, stacktrace := evalTest("", "", fullPath)

			if wantErr != "" && !strings.Contains(err, wantErr) ||
				wantErr == "" && err != "" {
				t.Fatalf("unexpected error\nWant: %s\n Got: %s", wantErr, err)
			}

			if wantStacktrace != "" && !strings.Contains(stacktrace, wantStacktrace) {
				t.Fatalf("unexpected stacktrace\nWant: %s\n Got: %s", wantStacktrace, stacktrace)
			}
			if wantOut != "" && out != wantOut {
				t.Fatalf("unexpected output\nWant: %s\n Got: %s", wantOut, out)
			}
		})

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// testData returns the expected output and error string, and true if entry is valid.
func testData(name string) (testOut, testErr, testStacktrace string, ok bool) {
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
	res := commentFrom(str, []string{
		"// Output:",
		"// Error:",
		"// Stacktrace:",
	})

	return res[0], res[1], res[2], true
}

type directive struct {
	delim string
	res   string
	index int
}

// (?m) makes ^ and $ match start/end of string
var reCommentPrefix = regexp.MustCompile("(?m)^//(?: |$)")

// commentFrom returns the comments from s that are between the delimiters.
func commentFrom(s string, delims []string) []string {
	directives := make([]directive, len(delims))
	directivesFound := make([]*directive, 0, len(delims))

	for i, delim := range delims {
		// must find delim isolated on one line
		delim = "\n" + delim + "\n"
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

		parsed := reCommentPrefix.ReplaceAllLiteralString(
			s[directivesFound[i].index+len(directivesFound[i].delim):next],
			"")
		directivesFound[i].res = strings.TrimSuffix(parsed, "\n")
	}

	res := make([]string, len(directives))
	for i, d := range directives {
		res[i] = d.res
	}

	return res
}
