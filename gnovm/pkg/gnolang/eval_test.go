package gnolang_test

import (
	"os"
	"path"
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
		wantOut, wantErr, ok := testData(dir, f)
		if !ok {
			continue
		}
		t.Run(f.Name(), func(t *testing.T) {
			out, err := evalTest("", "", path.Join(dir, f.Name()))

			if wantErr != "" && !strings.Contains(err, wantErr) ||
				wantErr == "" && err != "" {
				t.Fatalf("unexpected error\nWant: %s\n Got: %s", wantErr, err)
			}
			if wantOut != "" && out != wantOut {
				t.Fatalf("unexpected output\nWant: %s\n Got: %s", wantOut, out)
			}
		})
	}
}

// testData returns the expected output and error string, and true if entry is valid.
func testData(dir string, f os.DirEntry) (testOut, testErr string, ok bool) {
	if f.IsDir() {
		return "", "", false
	}
	name := path.Join(dir, f.Name())
	if !strings.HasSuffix(name, ".gno") || strings.HasSuffix(name, "_long.gno") {
		return "", "", false
	}
	buf, err := os.ReadFile(name)
	if err != nil {
		return "", "", false
	}
	str := string(buf)
	if strings.Contains(str, "// PKGPATH:") {
		return "", "", false
	}
	return commentFrom(str, "\n// Output:"), commentFrom(str, "\n// Error:"), true
}

// commentFrom returns the content from a trailing comment block in s starting with delim.
func commentFrom(s, delim string) string {
	index := strings.Index(s, delim)
	if index < 0 {
		return ""
	}
	return strings.TrimSpace(strings.ReplaceAll(s[index+len(delim):], "\n// ", "\n"))
}
