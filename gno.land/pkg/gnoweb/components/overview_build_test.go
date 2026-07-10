package components

import (
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/gnolang/gno/gnovm/pkg/doc"
	"github.com/stretchr/testify/require"
)

func TestComputeStats(t *testing.T) {
	t.Parallel()
	files := []string{"main.gno", "util.gno", "main_test.gno", "README.md", "gnomod.toml"}
	jdoc := &doc.JSONDocumentation{
		Funcs: []*doc.JSONFunc{
			{Name: "Hello"}, {Name: "internal"}, {Name: "WithCross", Crossing: true},
		},
		// "hidden" is unexported: it must not be counted (the page only lists
		// exported types), matching the render path's filter.
		Types: []*doc.JSONType{{Name: "Config"}, {Name: "State"}, {Name: "hidden"}},
		// Const/var counts follow the exported inner names, so the unexported
		// "secret" var group is excluded — sidebar totals match the sections shown.
		Values: []*doc.JSONValueDecl{
			{Const: true, Values: []*doc.JSONValue{{Name: "MaxN"}}},
			{Const: true, Values: []*doc.JSONValue{{Name: "MinN"}}},
			{Const: false, Values: []*doc.JSONValue{{Name: "Cache"}}},
			{Const: false, Values: []*doc.JSONValue{{Name: "secret"}}},
		},
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

func TestBugsNotInDoc(t *testing.T) {
	t.Parallel()
	// A BUG note whose text is already in the package doc (go/doc keeps it there
	// when the note lives in the package comment) is dropped to avoid a double
	// render; a floating note absent from the doc is kept.
	got := bugsNotInDoc(
		[]string{"thing A is broken", "thing B is broken"},
		"Package foo does things.\n\nBUG(alice): thing A is broken\n",
	)
	require.Equal(t, []string{"thing B is broken"}, got)

	require.Nil(t, bugsNotInDoc(nil, "anything"))
}

func TestBugsNotInDoc_KeepsNoteThatIsPrefixOfAnother(t *testing.T) {
	t.Parallel()
	// Only the longer note lives in the package doc. The shorter one is a
	// distinct floating note and must survive, even though its text is a
	// substring of the inline one.
	got := bugsNotInDoc(
		[]string{"Foo panics on nil", "Foo panics on nil input; use Bar"},
		"Package foo does things.\n\nBUG(alice): Foo panics on nil input; use Bar\n",
	)
	require.Equal(t, []string{"Foo panics on nil"}, got)
}

func TestBugsNotInDoc_TabIndentedNoteIsInline(t *testing.T) {
	t.Parallel()
	// An indented note is still the same note, whichever whitespace indents it.
	for _, indent := range []string{" ", "\t"} {
		got := bugsNotInDoc(
			[]string{"thing A is broken"},
			"Package foo does things.\n\n"+indent+"thing A is broken\n",
		)
		require.Nil(t, got, "note indented with %q must count as already rendered", indent)
	}
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
			want:   PackageQuality{},
		},
		{
			name:   "full package",
			files:  []string{"main.gno", "main_test.gno", "README.md", "LICENSE"},
			pkgDoc: "Package foo does things.",
			want:   PackageQuality{HasReadme: true, HasTests: true, HasLicense: true, HasPkgDoc: true},
		},
		{
			name:   "filetest counts as tests",
			files:  []string{"foo_filetest.gno"},
			pkgDoc: "",
			want:   PackageQuality{HasTests: true},
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
	imports := []ImportLink{{Path: "strings"}}
	files := []FileLink{{Name: "foo.gno", Link: "/r/demo/foo$source&file=foo.gno"}}
	subpacks := []SubpackageLink{{Name: "sub", Path: "/r/demo/foo/sub"}}
	quality := PackageQuality{HasPkgDoc: true, HasReadme: true}

	toc := buildOverviewTOC(quality, true, funcs, types, values, imports, files, subpacks)
	got := make([]string, 0, len(toc))
	for _, item := range toc {
		got = append(got, item.Title)
	}
	require.Equal(t, []string{"Overview", "README", "Constants", "Variables", "Functions", "Types", "Imports", "Files", "Directories"}, got)

	// Files hang under their section and link straight into the source view.
	filesTOC := toc[7]
	require.Len(t, filesTOC.Items, 1)
	require.Equal(t, "foo.gno", filesTOC.Items[0].Title)
	require.Equal(t, "/r/demo/foo$source&file=foo.gno", filesTOC.Items[0].Anchor())
	require.Equal(t, "#files", filesTOC.Anchor(), "the section header still anchors on the page")

	// A README that never rendered must not get a table-of-contents entry.
	unrendered := buildOverviewTOC(quality, false, funcs, types, values, imports, files, subpacks)
	titles := make([]string, 0, len(unrendered))
	for _, item := range unrendered {
		titles = append(titles, item.Title)
	}
	require.NotContains(t, titles, "README")

	// Each symbol leaf carries its kind glyph; section/group lines stay bare.
	funcsTOC := toc[4]
	require.Empty(t, funcsTOC.Icon, "group header carries no glyph")
	require.Equal(t, "kind-func", funcsTOC.Items[0].Icon)
	typesTOC := toc[5]
	require.Equal(t, "kind-type", typesTOC.Items[0].Icon, "type with unset Kind falls back to generic box")
	require.Equal(t, "kind-func", typesTOC.Items[0].Items[0].Icon, "method")
}

func TestTypeKindIcon(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"struct":    "kind-struct",
		"interface": "kind-interface",
		"slice":     "kind-slice",
		"array":     "kind-slice",
		"map":       "kind-map",
		"pointer":   "kind-pointer",
		"func":      "kind-func",
		"ident":     "kind-type",
		"channel":   "kind-type",
		"":          "kind-type",
	}
	for kind, want := range cases {
		require.Equal(t, want, typeKindIcon(kind), "kind=%q", kind)
	}
}

func TestDeriveInfo_RealmVsPure(t *testing.T) {
	t.Parallel()
	u, err := weburlParseForTest("/r/demo/foo")
	require.NoError(t, err)
	info := deriveInfo(u, nil)
	require.Equal(t, "realm", info.PackageType)
	require.Equal(t, "demo", info.Namespace)

	u, err = weburlParseForTest("/p/demo/foo")
	require.NoError(t, err)
	info = deriveInfo(u, nil)
	require.Equal(t, "pure", info.PackageType)
}

func TestDeriveInfo_GnoVersionFromGnomod(t *testing.T) {
	t.Parallel()
	u, err := weburlParseForTest("/r/demo/foo")
	require.NoError(t, err)
	gnomod := []byte(`module = "gno.land/r/demo/foo"` + "\n" + `gno = "0.1"`)
	info := deriveInfo(u, gnomod)
	require.Equal(t, "0.1", info.GnoVersion)
}

func TestDeriveInfo_DraftPrivateFromGnomod(t *testing.T) {
	t.Parallel()
	u, err := weburlParseForTest("/r/demo/foo")
	require.NoError(t, err)
	mod := []byte("module = \"gno.land/r/demo/foo\"\ngno = \"0.9\"\ndraft = true\nprivate = true")
	info := deriveInfo(u, mod)
	require.True(t, info.Draft)
	require.True(t, info.Private)
	require.Equal(t, "0.9", info.GnoVersion)
}

func TestDeriveInfo_CreatorHeightFromGnomod(t *testing.T) {
	t.Parallel()
	u, err := weburlParseForTest("/r/demo/foo")
	require.NoError(t, err)
	mod := []byte("module = \"gno.land/r/demo/foo\"\ngno = \"0.9\"\n\n[addpkg]\ncreator = \"g1abc\"\nheight = 42")
	info := deriveInfo(u, mod)
	require.Equal(t, "g1abc", info.Creator)
	require.Equal(t, 42, info.Height)
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

func TestClassifyFile(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		want FileClass
	}{
		{"source", "foo.gno", FileClass{IsGno: true}},
		{"unit test", "foo_test.gno", FileClass{IsGno: true, IsTest: true}},
		{"filetest", "foo_filetest.gno", FileClass{IsGno: true, IsTest: true}},
		{"readme", "README.md", FileClass{IsReadme: true}},
		{"license", "LICENSE", FileClass{IsLicense: true}},
		{"license md", "LICENSE.md", FileClass{IsLicense: true}},
		{"plain", "gnomod.toml", FileClass{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.want, ClassifyFile(tc.in))
		})
	}
}

func TestBuildOverview_BoundsSymbols(t *testing.T) {
	t.Parallel()
	u, err := weburl.Parse("/r/demo/foo")
	require.NoError(t, err)

	funcs := make([]*doc.JSONFunc, maxOverviewSymbols+10)
	for i := range funcs {
		funcs[i] = &doc.JSONFunc{Name: "Hello"} // exported; top-level
	}

	data := BuildOverview(OverviewInput{
		URL:         u,
		DocRenderer: noopRenderer{},
		Domain:      "gno.land",
		Doc:         &doc.JSONDocumentation{Funcs: funcs},
	})

	require.Len(t, data.Funcs, maxOverviewSymbols)
	require.True(t, data.SymbolsTruncated)
}
