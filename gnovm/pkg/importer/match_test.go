package importer

import (
	"fmt"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
)

func TestMatchPatterns(t *testing.T) {
	tt := []struct {
		name     string
		patterns []string
		// if both nil, expected to error
		success []string
		fail    []string
	}{
		{
			"basic",
			[]string{"*.gno", "!*_test.gno"},
			[]string{"hello.gno", "path/to/welcome.gno", "hello/.gno", ".gno"},
			[]string{"hello.go", "path/to/welcome.go", "x.gnox", ".gnox"},
		},
		{
			"globsReceiveLast",
			[]string{"path/*.gno"},
			[]string{},
			[]string{"path/hello.gno", "path/hello.go", "path/hello.gno/xx.gno", "path.gno/hello.gno"},
		},
		{
			"globInvalid",
			[]string{"[unterminated"},
			nil, nil,
		},
		{
			"negate",
			[]string{"!*.gno", "!*.go"},
			[]string{"negate.sol", "noext", "///", "", ".goa", "a.goa"},
			[]string{"hello.go", "path/to/welcome.go", "x.gno", ".gno"},
		},
		{
			"basicRegex",
			[]string{"/hello/"},
			[]string{"hello.gno", "x/to/hello/dir", "hello.go"},
			[]string{"Hello", "he/llo", "olleh", ".hel.lo"},
		},
		{
			"basicRegexNegate",
			[]string{"!/hello/"},
			[]string{"Hello", "he/llo", "olleh", ".hel.lo"},
			[]string{"hello.gno", "x/to/hello/dir", "hello.go"},
		},
		{
			"regexInvalid",
			[]string{"/[unmatched/"},
			nil, nil,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.success == nil && tc.fail == nil {
				_, err := MatchPatterns("xx", tc.patterns...)
				if err == nil {
					t.Errorf("%v expected to error but didn't", tc.patterns)
				}
				return
			}

			for _, s := range tc.success {
				ok, err := MatchPatterns(s, tc.patterns...)
				if err != nil {
					t.Fatal(err)
				}
				if !ok {
					t.Errorf("%q: does not match %v", s, tc.patterns)
				}
			}
			for _, s := range tc.fail {
				ok, err := MatchPatterns(s, tc.patterns...)
				if err != nil {
					t.Fatal(err)
				}
				if ok {
					t.Errorf("%q: matches %v", s, tc.patterns)
				}
			}
		})
	}
}

func TestIsGnoFile(t *testing.T) {
	for _, s := range [...]string{
		"hello.gno",
		"file/to/hello.gno",
		"hello_test.gno",
		"hello_filetest.gno",
	} {
		if !IsGnoFile(s) {
			t.Errorf("%q marked (incorrectly) as not a gno file", s)
		}
	}

	for _, s := range [...]string{
		"hidden/.hidden.gno",
		"onlysuffix/.gno",
		"notgno/hello.go",
	} {
		if IsGnoFile(s) {
			t.Errorf("%q marked (incorrectly) as a gno file", s)
		}
	}
}

func TestFilter(t *testing.T) {
	testSlice := []string{
		"README",
		"LICENSE",
		"gno.mod",
		"file.gno",
		"hello.gno",
		"x_filetest.gno",
		"x_test.gno",
	}
	assert.Equal(t, Filter(testSlice, "!*.gno"), []string{"README", "LICENSE", "gno.mod"})
	assert.Equal(t, Filter(testSlice, "gno.*"), []string{"gno.mod"})
	assert.Equal(t, Filter(testSlice, "*.gno"), []string{"file.gno", "hello.gno", "x_filetest.gno", "x_test.gno"})
	assert.Equal(t, Filter(testSlice, "*_test.gno"), []string{"x_test.gno"})
	assert.Equal(t, Filter(testSlice, "*_filetest.gno"), []string{"x_filetest.gno"})
	assert.Equal(t, Filter(testSlice, "/_(file)?test.gno$/"), []string{"x_filetest.gno", "x_test.gno"})
}

const matchStructure1 = `README
LICENSE
doc.gno
gno.mod
p1/hello.gno
p1/hello_test.gno
p1/hello_filetest.gno
p1/README
p2/goodbye.gno
p2/goodbye_test.gno
p2/goodbye.go
pno/invalid.go
.hidden/
.hidden/test.gno`

const matchStructure2 = `d1/d2/hello/d3/pkg.gno
d1/d2/world/d3/pkg.gno
d1/d2/d3/pkg.gno`

const matchStructure3 = `foo/bar/x.gno
foo/bar/baz/x.gno
foo/baz/bar/x.gno
foo/x.gno
bar/x.gno`

func generateFS(structure string) fs.FS {
	parts := strings.Split(structure, "\n")
	mfs := make(fstest.MapFS, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		mfs[part] = &fstest.MapFile{Data: []byte("Xxx")}
	}
	return mfs
}

func Test_generateStructure(t *testing.T) {
	t.Skip("Skipping (activate for debugging)")
	buf := new(strings.Builder)
	fs.WalkDir(generateFS(matchStructure1), ".", func(path string, d fs.DirEntry, err error) error {
		fi, err := d.Info()
		if err != nil {
			return err
		}
		fmt.Fprintf(buf, "%s: %v\n", path, fi.Mode())
		return nil
	})
	t.Log(buf.String())
}

func TestMatch(t *testing.T) {
	tt := []struct {
		name        string
		structure   string
		paths       []string
		opts        []MatchOption
		exp         []string
		errContains string
	}{
		{
			"basic", matchStructure1,

			[]string{"./p1"},
			nil,

			[]string{"p1"},
			"",
		},
		{
			"basicFiles", matchStructure1,

			[]string{"./p1"},
			[]MatchOption{MatchFiles("!*_filetest.gno")},

			[]string{"p1/hello.gno", "p1/hello_test.gno"},
			"",
		},
		{
			"explicit", matchStructure1,

			[]string{"./p1", "./p1/hello_test.gno"},
			[]MatchOption{MatchFiles("!*_test.gno", "!*_filetest.gno")},

			[]string{"p1/hello.gno", "p1/hello_test.gno"},
			"",
		},
		{
			"root", matchStructure1,

			[]string{"/", "/p1"},
			nil,

			[]string{"/", "/p1"},
			"",
		},
		{
			"ellipsis", matchStructure1,

			[]string{"..."},
			nil,

			[]string{".", "p1", "p2"},
			"",
		},
		{
			"onlyTestPackages", matchStructure1,

			[]string{"..."},
			[]MatchOption{MatchPackages("*_test.gno")},

			[]string{"p1", "p2"},
			"",
		},
		{
			"hidden", matchStructure1,

			[]string{"...", "./.hidden"},
			nil,

			[]string{".", "p1", "p2", ".hidden"},
			"",
		},
		{
			"subEllipsis", matchStructure2,

			[]string{"/d1/d2/.../d3"},
			nil,

			[]string{"/d1/d2/d3", "/d1/d2/hello/d3", "/d1/d2/world/d3"},
			"",
		},
		{
			"subEllipsisPackage", matchStructure2,

			[]string{"/d1/d2/.../d3"},
			[]MatchOption{MatchFiles()},

			[]string{"/d1/d2/d3/pkg.gno", "/d1/d2/hello/d3/pkg.gno", "/d1/d2/world/d3/pkg.gno"},
			"",
		},
		{
			"matchNone", matchStructure1,

			[]string{"pno/..."},
			nil,

			nil,
			"",
		},
		{
			"goMatchTest", matchStructure3,

			[]string{"foo/bar/..."},
			nil,

			[]string{"foo/bar", "foo/bar/baz"},
			"",
		},
		{
			"goMatchTest2", matchStructure3,

			[]string{"foo/.../baz"},
			nil,

			[]string{"foo/bar/baz"},
			"",
		},

		{
			"errInvalidPath", matchStructure1,

			[]string{"/", "../../../../"},
			nil,

			nil,
			`invalid path: "../../`,
		},
		{
			"errNotFound", matchStructure1,

			[]string{"not_exist"},
			nil,

			nil,
			`file does not exist`,
		},
		{
			"errNotPackage", matchStructure1,

			[]string{"p1", "pno", "p2"},
			nil,

			nil,
			`dir pno: no valid gno files`,
		},
		{
			"errBadPattern", matchStructure1,

			[]string{"p1"},
			[]MatchOption{MatchFiles("[unmatched")},

			nil,
			`syntax error in pattern`,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			f := generateFS(tc.structure)
			res, err := Match(tc.paths, append(tc.opts, func(c *matchOptions) {
				c.fs = f
			})...)
			if tc.errContains == "" {
				assert.Nil(t, err, "%v", err)
				assert.Equal(t, res, tc.exp)
			} else {
				_ = assert.NotNil(t, err) &&
					assert.Contains(t, err.Error(), tc.errContains)
				assert.Equal(t, res, tc.exp)
			}
		})
	}
}

func Test_revertPath(t *testing.T) {
	tt := []struct {
		name                 string
		cwd, original, clean string
		result               string
	}{
		{"basic", "home/morgan", "file.gno", "home/morgan/file.gno", "file.gno"},
		{"sub", "home/morgan", "t1/file.gno", "home/morgan/t1/file.gno", "t1/file.gno"},
		{"dot", "home/moul", ".", "home/moul", "."},
		{"dotdot", "home/morgan", "../file.gno", "home/file.gno", "../file.gno"},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, revertPath(tc.cwd, tc.original, tc.clean), tc.result, "result should match")
		})
	}
}
