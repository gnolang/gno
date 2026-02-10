package doc

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveDocumentable(t *testing.T) {
	p, err := os.Getwd()
	require.NoError(t, err)
	path := func(s string) string { return filepath.Join(p, "testdata/integ", s) }
	dirs := newDirs([]string{path("")}, []string{path("mod")})
	getDir := func(p string) bfsDir { return dirs.findDir(path(p))[0] }
	pdata := func(p string, unexp bool) *pkgData {
		pd, err := newPkgData(getDir(p), unexp)
		require.NoError(t, err)
		return pd
	}

	tt := []struct {
		name        string
		args        []string
		unexp       bool
		expect      *Documentable
		errContains string
	}{
		{"package", []string{"crypto/rand"}, false, &Documentable{bfsDir: getDir("crypto/rand")}, ""},
		{"packageMod", []string{"gno.land/mod"}, false, nil, `package not found`},
		{"dir", []string{"./testdata/integ/crypto/rand"}, false, &Documentable{bfsDir: getDir("crypto/rand")}, ""},
		{"dirMod", []string{"./testdata/integ/mod"}, false, &Documentable{bfsDir: getDir("mod")}, ""},
		{"dirAbs", []string{path("crypto/rand")}, false, &Documentable{bfsDir: getDir("crypto/rand")}, ""},
		// test_notapkg exists in local dir and also path("test_notapkg").
		// ResolveDocumentable should first try local dir, and seeing as it is not a valid dir, try searching it as a package.
		{"dirLocalMisleading", []string{"test_notapkg"}, false, &Documentable{bfsDir: getDir("test_notapkg")}, ""},
		{
			"normalSymbol",
			[]string{"crypto/rand.Flag"},
			false,
			&Documentable{bfsDir: getDir("crypto/rand"), symbol: "Flag", pkgData: pdata("crypto/rand", false)}, "",
		},
		{
			"normalAccessible",
			[]string{"crypto/rand.Generate"},
			false,
			&Documentable{bfsDir: getDir("crypto/rand"), symbol: "Generate", pkgData: pdata("crypto/rand", false)}, "",
		},
		{
			"normalSymbolUnexp",
			[]string{"crypto/rand.unexp"},
			true,
			&Documentable{bfsDir: getDir("crypto/rand"), symbol: "unexp", pkgData: pdata("crypto/rand", true)}, "",
		},
		{
			"normalAccessibleFull",
			[]string{"crypto/rand.Rand.Name"},
			false,
			&Documentable{bfsDir: getDir("crypto/rand"), symbol: "Rand", accessible: "Name", pkgData: pdata("crypto/rand", false)}, "",
		},
		{
			"disambiguate",
			[]string{"rand.Flag"},
			false,
			&Documentable{bfsDir: getDir("crypto/rand"), symbol: "Flag", pkgData: pdata("crypto/rand", false)}, "",
		},
		{
			"disambiguate2",
			[]string{"rand.Crypto"},
			false,
			&Documentable{bfsDir: getDir("crypto/rand"), symbol: "Crypto", pkgData: pdata("crypto/rand", false)}, "",
		},
		{
			"disambiguate3",
			[]string{"rand.Normal"},
			false,
			&Documentable{bfsDir: getDir("rand"), symbol: "Normal", pkgData: pdata("rand", false)}, "",
		},
		{
			"disambiguate4", // just "rand" should use the directory that matches it exactly.
			[]string{"rand"},
			false,
			&Documentable{bfsDir: getDir("rand")}, "",
		},
		{
			"wdSymbol",
			[]string{"WdConst"},
			false,
			&Documentable{bfsDir: getDir("wd"), symbol: "WdConst", pkgData: pdata("wd", false)}, "",
		},

		{"errInvalidArgs", []string{"1", "2", "3"}, false, nil, "invalid arguments: [1 2 3]"},
		{"errNoCandidates", []string{"math", "Big"}, false, nil, `package not found: "math"`},
		{"errNoCandidates2", []string{"LocalSymbol"}, false, nil, `package not found`},
		{"errNoCandidates3", []string{"Symbol.Accessible"}, false, nil, `package not found`},
		{"errNonExisting", []string{"rand.NotExisting"}, false, nil, `could not resolve arguments`},
		{"errIgnoredMod", []string{"modignored"}, false, &Documentable{bfsDir: getDir("modignored")}, ""},
		{"errIgnoredMod2", []string{"./testdata/integ/modignored"}, false, &Documentable{bfsDir: getDir("modignored")}, ""},
		{"errUnexp", []string{"crypto/rand.unexp"}, false, nil, "could not resolve arguments"},
		{"errDirNotapkg", []string{"./test_notapkg"}, false, nil, `package not found: "./test_notapkg"`},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Wd prefix mean test relative to local directory -
			// mock change local dir by setting the fpAbs variable (see doc.go) to match
			// testdata/integ/wd when we call it on ".".
			if strings.HasPrefix(tc.args[0], "Wd") {
				fpAbs = func(s string) (string, error) { return filepath.Clean(filepath.Join(path("wd"), s)), nil }
				defer func() { fpAbs = filepath.Abs }()
			}
			result, err := ResolveDocumentable(
				[]string{path("")}, []string{path("mod")},
				tc.args, tc.unexp,
			)
			// we use stripFset because d.pkgData.fset contains sync/atomic values,
			// which in turn makes reflect.DeepEqual compare the two sync.Atomic values.
			assert.Equal(t, stripFset(tc.expect), stripFset(result), "Documentables should match")
			if tc.errContains == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tc.errContains)
			}
		})
	}
}

func stripFset(d *Documentable) *Documentable {
	if d != nil && d.pkgData != nil {
		d.pkgData.fset = nil
	}
	return d
}

func TestDocument(t *testing.T) {
	// the format itself can change if the design is to be changed,
	// we want to make sure that given information is available when calling
	// Document.
	abspath, err := filepath.Abs("./testdata/integ/crypto/rand")
	require.NoError(t, err)
	dir := bfsDir{
		importPath: "crypto/rand",
		dir:        abspath,
	}

	tt := []struct {
		name     string
		d        *Documentable
		opts     *WriteDocumentationOptions
		contains []string
	}{
		{"base", &Documentable{bfsDir: dir}, nil, []string{"func Crypto", "!Crypto symbol", "func NewRand", "!unexp", "type Flag", "!Name"}},
		{"func", &Documentable{bfsDir: dir, symbol: "crypto"}, nil, []string{"Crypto symbol", "func Crypto", "!func NewRand", "!type Flag"}},
		{"funcWriter", &Documentable{bfsDir: dir, symbol: "NewWriter"}, nil, []string{"func NewWriter() io.Writer", "!func Crypto", "!// crossing"}},
		{"tp", &Documentable{bfsDir: dir, symbol: "Rand"}, nil, []string{"type Rand", "comment1", "!func Crypto", "!unexp  ", "!comment4", "Has unexported"}},
		{"inter", &Documentable{bfsDir: dir, symbol: "Rander"}, nil, []string{"type Rander", "generate", "!unexp  ", "!comment1", "Has unexported"}},
		{"tpField", &Documentable{bfsDir: dir, symbol: "Rand", accessible: "Value"}, nil, []string{"type Rand", "!comment1", "comment2", "!func Crypto", "!unexp", "elided"}},
		{
			"tpUnexp",
			&Documentable{bfsDir: dir, symbol: "Rand"},
			&WriteDocumentationOptions{Unexported: true},
			[]string{"type Rand", "comment1", "!func Crypto", "unexp  ", "comment4", "!Has unexported"},
		},
		{
			"symUnexp",
			&Documentable{bfsDir: dir, symbol: "unexp"},
			&WriteDocumentationOptions{Unexported: true},
			[]string{"var unexp", "!type Rand", "!comment1", "!comment4", "!func Crypto", "!Has unexported"},
		},
		{
			"fieldUnexp",
			&Documentable{bfsDir: dir, symbol: "Rand", accessible: "unexp"},
			&WriteDocumentationOptions{Unexported: true},
			[]string{"type Rand", "!comment1", "comment4", "!func Crypto", "elided", "!Has unexported"},
		},
	}

	buf := &bytes.Buffer{}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			buf.Reset()
			err := tc.d.WriteDocumentation(buf, tc.opts)
			require.NoError(t, err)
			s := buf.String()
			for _, c := range tc.contains {
				if c[0] == '!' {
					assert.NotContains(t, s, c[1:])
				} else {
					assert.Contains(t, s, c)
				}
			}
		})
	}
}

func Test_parseArgParts(t *testing.T) {
	tt := []struct {
		name string
		args []string
		exp  *docArgs
	}{
		{"noArgs", []string{}, &docArgs{pkg: "."}},

		{"oneAmbiguous", []string{"ambiguous"}, &docArgs{pkg: "ambiguous", pkgAmbiguous: true}},
		{"onePath", []string{"pkg/path"}, &docArgs{pkg: "pkg/path"}},
		{"oneSpecial", []string{".."}, &docArgs{pkg: ".."}},
		{"oneSpecial2", []string{"../../../.."}, &docArgs{pkg: "../../../.."}},
		{"oneSpecial3", []string{"../upper/.."}, &docArgs{pkg: "../upper/.."}},
		{"oneSpecial4", []string{"."}, &docArgs{pkg: "."}},

		{"twoPkgSym", []string{"pkg.sym"}, &docArgs{pkg: "pkg", sym: "sym", pkgAmbiguous: true}},
		{"twoPkgPathSym", []string{"path/pkg.sym"}, &docArgs{pkg: "path/pkg", sym: "sym"}},
		{"twoPkgUpperSym", []string{"../pkg.sym"}, &docArgs{pkg: "../pkg", sym: "sym"}},
		{"twoPkgExportedSym", []string{"Writer.Write"}, &docArgs{pkg: ".", sym: "Writer", acc: "Write"}},
		{"twoPkgCapitalPathSym", []string{"Path/Capitalised.Sym"}, &docArgs{pkg: "Path/Capitalised", sym: "Sym"}},

		{"threePkgSymAcc", []string{"pkg.sym.acc"}, &docArgs{pkg: "pkg", sym: "sym", acc: "acc"}},
		{"threePathPkgSymAcc", []string{"./pkg.sym.acc"}, &docArgs{pkg: "./pkg", sym: "sym", acc: "acc"}},
		{"threePathPkgSymAcc2", []string{"../pkg.sym.acc"}, &docArgs{pkg: "../pkg", sym: "sym", acc: "acc"}},
		{"threePathPkgSymAcc3", []string{"path/to/pkg.sym.acc"}, &docArgs{pkg: "path/to/pkg", sym: "sym", acc: "acc"}},
		{"threePathPkgSymAcc4", []string{"path/../to/pkg.sym.acc"}, &docArgs{pkg: "path/../to/pkg", sym: "sym", acc: "acc"}},

		// the logic on the split is pretty unambiguously that the first argument
		// is the path, so we can afford to be less thorough on that regard.
		{"splitTwo", []string{"io", "Writer"}, &docArgs{pkg: "io", sym: "Writer"}},
		{"splitThree", []string{"io", "Writer.Write"}, &docArgs{pkg: "io", sym: "Writer", acc: "Write"}},

		{"errTooManyDots", []string{"io.Writer.Write.Impossible"}, nil},
		{"errTooManyDotsSplit", []string{"io", "Writer.Write.Impossible"}, nil},
		{"errTooManyArgs", []string{"io", "Writer", "Write"}, nil},
	}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			p, ok := parseArgs(tc.args)
			if ok {
				_ = assert.NotNil(t, tc.exp, "parseArgs is successful when should have failed") &&
					assert.Equal(t, *tc.exp, p)
			} else {
				assert.Nil(t, tc.exp, "parseArgs is unsuccessful")
			}
		})
	}
}
