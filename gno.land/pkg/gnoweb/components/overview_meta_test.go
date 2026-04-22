package components

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/gnolang/gno/gnovm/pkg/doc"
	"github.com/stretchr/testify/require"
)

// weburlParseForTest wraps weburl.Parse for test readability.
func weburlParseForTest(path string) (*weburl.GnoURL, error) {
	return weburl.Parse(path)
}

func fileContentFn(m map[string][]byte) func(string) ([]byte, bool) {
	return func(name string) ([]byte, bool) {
		v, ok := m[name]
		return v, ok
	}
}

func TestDeriveLicense(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		files   []string
		content map[string][]byte
		want    License
	}{
		{
			name:  "no license file",
			files: []string{"main.gno", "README.md"},
			want:  License{},
		},
		{
			name:    "MIT by content signature",
			files:   []string{"LICENSE"},
			content: map[string][]byte{"LICENSE": []byte("The MIT License\n\nCopyright (c) 2024 ...")},
			want:    License{Kind: "MIT", FileName: "LICENSE"},
		},
		{
			name:    "SPDX identifier takes precedence over signature",
			files:   []string{"LICENSE.md"},
			content: map[string][]byte{"LICENSE.md": []byte("SPDX-License-Identifier: Apache-2.0\n\nThe MIT License text ...")},
			want:    License{Kind: "Apache-2.0", FileName: "LICENSE.md"},
		},
		{
			name:    "unknown license type still surfaces file name",
			files:   []string{"LICENSE.txt"},
			content: map[string][]byte{"LICENSE.txt": []byte("Some custom wording with no known signature")},
			want:    License{Kind: "", FileName: "LICENSE.txt"},
		},
		{
			name:    "file exists but content not fetched",
			files:   []string{"LICENSE"},
			content: nil,
			want:    License{FileName: "LICENSE"},
		},
		{
			name:    "bounded 4KB read ignores late signature",
			files:   []string{"LICENSE"},
			content: map[string][]byte{"LICENSE": append(bytes.Repeat([]byte(" "), 5000), []byte("The MIT License")...)},
			want:    License{Kind: "", FileName: "LICENSE"},
		},
		{
			name:    "Apache detection",
			files:   []string{"LICENSE"},
			content: map[string][]byte{"LICENSE": []byte("Apache License, Version 2.0\n\n...")},
			want:    License{Kind: "Apache-2.0", FileName: "LICENSE"},
		},
		{
			name:    "BSD-3-Clause detection",
			files:   []string{"LICENSE"},
			content: map[string][]byte{"LICENSE": []byte("Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:\n\n1. ...\n2. ...\n3. Neither the name of the copyright holder ...")},
			want:    License{Kind: "BSD-3-Clause", FileName: "LICENSE"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := deriveLicense(tc.files, fileContentFn(tc.content))
			require.Equal(t, tc.want, got)
		})
	}
}

func TestParseImports_ClassifyAndLink(t *testing.T) {
	t.Parallel()
	src := []byte(`package foo
import (
	"strings"
	"gno.land/p/demo/avl"
	"gno.land/r/gnoland/users/v1"
	"github.com/external/dep"
)
func Ignored() {}`)
	got := parseImports(map[string][]byte{"main.gno": src}, "gno.land")
	require.Equal(t, []ImportLink{
		{Path: "github.com/external/dep", Kind: "external", Link: ""},
		{Path: "gno.land/p/demo/avl", Kind: "package", Link: "/p/demo/avl"},
		{Path: "gno.land/r/gnoland/users/v1", Kind: "realm", Link: "/r/gnoland/users/v1"},
		{Path: "strings", Kind: "stdlib", Link: ""},
	}, got)
}

func TestParseImports_DedupAcrossFiles(t *testing.T) {
	t.Parallel()
	src1 := []byte(`package p
import "strings"
import "gno.land/p/demo/avl"`)
	src2 := []byte(`package p
import "strings"
import "fmt"`)
	got := parseImports(map[string][]byte{"a.gno": src1, "b.gno": src2}, "gno.land")
	paths := make([]string, 0, len(got))
	for _, im := range got {
		paths = append(paths, im.Path)
	}
	require.Equal(t, []string{"fmt", "gno.land/p/demo/avl", "strings"}, paths)
}

func TestParseImports_MalformedBodyTolerated(t *testing.T) {
	t.Parallel()
	// ImportsOnly stops after imports, so .gno-only syntax later is irrelevant.
	src := []byte(`package foo
import "strings"

func WithCross(cur realm, arg string) string {
	return arg
}`)
	got := parseImports(map[string][]byte{"main.gno": src}, "gno.land")
	require.Len(t, got, 1)
	require.Equal(t, "strings", got[0].Path)
}

func TestParseImports_EmptyInput(t *testing.T) {
	t.Parallel()
	got := parseImports(nil, "gno.land")
	require.Nil(t, got)
}

func TestParseImports_UnparseableFileSilentlySkipped(t *testing.T) {
	t.Parallel()
	src := []byte(`not go at all`)
	got := parseImports(map[string][]byte{"bad.gno": src}, "gno.land")
	require.Nil(t, got)
}

func TestComputeStats(t *testing.T) {
	t.Parallel()
	files := []string{"main.gno", "util.gno", "main_test.gno", "README.md", "gnomod.toml"}
	jdoc := &doc.JSONDocumentation{
		Funcs: []*doc.JSONFunc{
			{Name: "Hello"}, {Name: "internal"}, {Name: "WithCross", Crossing: true},
		},
		Types:  []*doc.JSONType{{Name: "Config"}, {Name: "State"}},
		Values: []*doc.JSONValueDecl{{Const: true}, {Const: false}, {Const: true}},
	}
	imports := []ImportLink{{Path: "strings"}, {Path: "gno.land/p/demo/avl"}}
	got := computeStats(files, jdoc, imports)
	require.Equal(t, PackageStats{
		FileCount:     5,
		GnoFileCount:  3,
		TestCount:     1,
		FuncCount:     3,
		ExportedFunc:  2,
		TypeCount:     2,
		ConstCount:    2,
		VarCount:      1,
		ImportCount:   2,
		CrossingCount: 1,
	}, got)
}

func TestDeriveQuality(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		files  []string
		pkgDoc string
		want   PackageQuality
	}{
		{
			name:   "nothing",
			files:  []string{"main.gno"},
			pkgDoc: "",
			want:   PackageQuality{SourceVerified: true},
		},
		{
			name:   "full package",
			files:  []string{"main.gno", "main_test.gno", "README.md", "LICENSE"},
			pkgDoc: "Package foo does things.",
			want:   PackageQuality{HasReadme: true, HasTests: true, HasLicense: true, HasPkgDoc: true, SourceVerified: true},
		},
		{
			name:   "filetest counts as tests",
			files:  []string{"foo_filetest.gno"},
			pkgDoc: "",
			want:   PackageQuality{HasTests: true, SourceVerified: true},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := deriveQuality(tc.files, &doc.JSONDocumentation{PackageDoc: tc.pkgDoc})
			require.Equal(t, tc.want, got)
		})
	}
}

func TestExtractSynopsis(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in, want string
	}{
		{"", ""},
		{"single line no newline", "single line no newline"},
		{"first line\nsecond line", "first line"},
		{strings.Repeat("x", 130), strings.Repeat("x", 117) + "..."},
		{strings.Repeat("x", 120), strings.Repeat("x", 120)},
	}
	for _, tc := range tests {
		require.Equal(t, tc.want, extractSynopsis(tc.in))
	}
}

func TestBuildFileLinks(t *testing.T) {
	t.Parallel()
	got := buildFileLinks("/r/demo/foo", []string{"foo.gno", "foo_test.gno", "README.md", "LICENSE"})
	require.Equal(t, []FileLink{
		{Name: "foo.gno", Link: "/r/demo/foo$source&file=foo.gno"},
		{Name: "foo_test.gno", Link: "/r/demo/foo$source&file=foo_test.gno", IsTest: true},
		{Name: "README.md", Link: "/r/demo/foo$source&file=README.md", IsReadme: true},
		{Name: "LICENSE", Link: "/r/demo/foo$source&file=LICENSE", IsLicense: true},
	}, got)
}

func TestBuildSubpackages(t *testing.T) {
	t.Parallel()
	got := buildSubpackages("/r/demo/foo", []string{
		"/r/demo/foo",
		"/r/demo/foo/bar",
		"/r/demo/foo/bar/baz",
		"/r/demo/foo/qux/",
		"/r/demo/other",
	})
	require.Equal(t, []SubpackageLink{
		{Name: "bar", Path: "/r/demo/foo/bar"},
		{Name: "qux", Path: "/r/demo/foo/qux"},
	}, got)
}

func TestBuildOverviewTOC(t *testing.T) {
	t.Parallel()
	funcs := []FuncEntry{{Name: "Hello", AnchorID: "func-Hello"}}
	types := []TypeEntry{{Name: "Config", AnchorID: "type-Config", Methods: []FuncEntry{{Name: "Load", AnchorID: "method-Config-Load"}}}}
	values := []ValueGroup{{Kind: "const"}, {Kind: "var"}}
	quality := PackageQuality{HasPkgDoc: true, HasReadme: true}

	toc := buildOverviewTOC(quality, funcs, types, values)
	got := make([]string, 0, len(toc))
	for _, item := range toc {
		got = append(got, item.Title)
	}
	require.Equal(t, []string{"Overview", "README", "Constants", "Variables", "Functions", "Types", "Files"}, got)
}

func TestDeriveInfo_RealmVsPure(t *testing.T) {
	t.Parallel()
	u, err := weburlParseForTest("/r/demo/foo")
	require.NoError(t, err)
	info := deriveInfo(u, []string{"foo.gno"}, nil)
	require.Equal(t, "realm", info.PackageType)
	require.Equal(t, "demo", info.Namespace)

	u, err = weburlParseForTest("/p/demo/foo")
	require.NoError(t, err)
	info = deriveInfo(u, []string{"foo.gno"}, nil)
	require.Equal(t, "pure", info.PackageType)
}

func TestDeriveInfo_GnoVersionFromGnomod(t *testing.T) {
	t.Parallel()
	u, err := weburlParseForTest("/r/demo/foo")
	require.NoError(t, err)
	gnomod := []byte(`module = "gno.land/r/demo/foo"` + "\n" + `gno = "0.1"`)
	info := deriveInfo(u, []string{"foo.gno", "gnomod.toml"}, gnomod)
	require.Equal(t, "0.1", info.GnoVersion)
}

// noopRenderer renders doc strings by writing them unchanged — enough for unit tests.
type noopRenderer struct{}

func (noopRenderer) RenderDocumentation(w io.Writer, src []byte) error {
	_, err := w.Write(src)
	return err
}

func (noopRenderer) RenderSource(w io.Writer, name string, src []byte) error {
	_, err := w.Write(src)
	return err
}

func TestBuildSymbols_SplitsTopLevelAndMethods(t *testing.T) {
	t.Parallel()
	jdoc := &doc.JSONDocumentation{
		Types: []*doc.JSONType{
			{Name: "Config", Type: "type Config struct{}", Kind: "struct"},
		},
		Funcs: []*doc.JSONFunc{
			{Name: "Hello", Signature: "func Hello() string", Type: ""},
			{Name: "Load", Signature: "func (c *Config) Load() error", Type: "Config"},
		},
	}
	funcs, types := buildSymbols(jdoc, noopRenderer{}, "/r/demo/foo")
	require.Len(t, funcs, 1)
	require.Equal(t, "Hello", funcs[0].Name)
	require.Equal(t, "/r/demo/foo$help&func=Hello", funcs[0].ActionURL)
	require.Len(t, types, 1)
	require.Equal(t, "Config", types[0].Name)
	require.Len(t, types[0].Methods, 1)
	require.Equal(t, "Load", types[0].Methods[0].Name)
	require.True(t, types[0].Methods[0].IsMethod)
}

func TestBuildSymbols_FiltersUnexported(t *testing.T) {
	t.Parallel()
	jdoc := &doc.JSONDocumentation{
		Funcs: []*doc.JSONFunc{
			{Name: "public"},
			{Name: "Exported"},
		},
	}
	funcs, _ := buildSymbols(jdoc, noopRenderer{}, "/r/demo/foo")
	require.Len(t, funcs, 1)
	require.Equal(t, "Exported", funcs[0].Name)
}

func TestBuildSymbols_RenderIsNotActionLinked(t *testing.T) {
	t.Parallel()
	jdoc := &doc.JSONDocumentation{
		Funcs: []*doc.JSONFunc{{Name: "Render"}},
	}
	funcs, _ := buildSymbols(jdoc, noopRenderer{}, "/r/demo/foo")
	require.Len(t, funcs, 1)
	require.Empty(t, funcs[0].ActionURL, "Render must not get an ActionURL")
}

func TestBuildSymbols_CrossingFlag(t *testing.T) {
	t.Parallel()
	jdoc := &doc.JSONDocumentation{
		Funcs: []*doc.JSONFunc{{Name: "Mutate", Crossing: true}},
	}
	funcs, _ := buildSymbols(jdoc, noopRenderer{}, "/r/demo/foo")
	require.True(t, funcs[0].Crossing)
}

func TestBuildSymbols_OrphanMethodSkipped(t *testing.T) {
	t.Parallel()
	jdoc := &doc.JSONDocumentation{
		Funcs: []*doc.JSONFunc{{Name: "Stray", Type: "MissingType"}},
	}
	funcs, types := buildSymbols(jdoc, noopRenderer{}, "/r/demo/foo")
	require.Empty(t, funcs)
	require.Empty(t, types)
}

func TestBuildValues_OrderAndKinds(t *testing.T) {
	t.Parallel()
	jdoc := &doc.JSONDocumentation{
		Values: []*doc.JSONValueDecl{
			{Const: true, Signature: "const X = 1", Values: []*doc.JSONValue{{Name: "X"}}},
			{Const: false, Signature: "var Y int", Values: []*doc.JSONValue{{Name: "Y"}}},
			{Const: true, Signature: "const (A=1; B=2)", Values: []*doc.JSONValue{{Name: "A"}, {Name: "B"}}},
		},
	}
	got := buildValues(jdoc, noopRenderer{}, "/r/demo/foo")
	require.Len(t, got, 3)
	require.Equal(t, "const", got[0].Kind)
	require.Equal(t, "X", got[0].Names)
	require.Equal(t, "var", got[1].Kind)
	require.Equal(t, "const", got[2].Kind)
	require.Equal(t, "A, B", got[2].Names)
}

func TestBuildSourceURL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		pkg  string
		file string
		line int
		want string
	}{
		{"valid link", "/r/demo/foo", "foo.gno", 42, "/r/demo/foo$source&file=foo.gno#L42"},
		{"empty file returns empty", "/r/demo/foo", "", 42, ""},
		{"zero line returns empty", "/r/demo/foo", "foo.gno", 0, ""},
		{"negative line returns empty", "/r/demo/foo", "foo.gno", -1, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.want, buildSourceURL(tc.pkg, tc.file, tc.line))
		})
	}
}

func TestFilterNonTestSources(t *testing.T) {
	t.Parallel()
	sources := map[string][]byte{
		"foo.gno":          []byte("package foo"),
		"foo_test.gno":     []byte("package foo"),
		"bar_filetest.gno": []byte("package foo"),
		"README.md":        []byte("# readme"),
		"gnomod.toml":      []byte("module = \"x\""),
	}
	got := filterNonTestSources(sources)
	require.Len(t, got, 1, "only non-test .gno files are kept for import parsing")
	_, ok := got["foo.gno"]
	require.True(t, ok)
	require.Nil(t, filterNonTestSources(nil), "nil input → nil output")
	require.Nil(t, filterNonTestSources(map[string][]byte{}), "empty input → nil output")
}

func TestPackageTypeOf(t *testing.T) {
	t.Parallel()
	cases := []struct {
		path string
		want string
	}{
		{"/r/demo/foo", "realm"},
		{"/p/demo/foo", "pure"},
		{"/u/someone", ""}, // neither realm nor pure → empty
	}
	for _, tc := range cases {
		u, err := weburl.Parse(tc.path)
		require.NoError(t, err)
		require.Equal(t, tc.want, packageTypeOf(u))
	}
}

func TestRawHTMLComponentRender(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	c := rawHTMLComponent("<p>hello</p>")
	require.NoError(t, c.Render(&buf))
	require.Equal(t, "<p>hello</p>", buf.String())
}

func TestBuildSymbols_FiltersUnexportedTypes(t *testing.T) {
	t.Parallel()
	jdoc := &doc.JSONDocumentation{
		Types: []*doc.JSONType{
			{Name: "Public", Type: "type Public struct{}", Kind: "struct"},
			{Name: "internal", Type: "type internal struct{}", Kind: "struct"},
		},
	}
	_, types := buildSymbols(jdoc, noopRenderer{}, "/r/demo/foo")
	require.Len(t, types, 1)
	require.Equal(t, "Public", types[0].Name)
}

func TestBuildValues_FiltersUnexportedDecls(t *testing.T) {
	t.Parallel()
	jdoc := &doc.JSONDocumentation{
		Values: []*doc.JSONValueDecl{
			{Const: true, Values: []*doc.JSONValue{{Name: "Pub"}}},
			{Const: false, Values: []*doc.JSONValue{{Name: "internal"}}},
			{Const: false, Values: []*doc.JSONValue{{Name: "Mixed"}, {Name: "pvt"}}}, // mixed = kept
		},
	}
	got := buildValues(jdoc, noopRenderer{}, "/r/demo/foo")
	require.Len(t, got, 2)
	require.Equal(t, "Pub", got[0].Names)
	require.Equal(t, "Mixed, pvt", got[1].Names)
}
