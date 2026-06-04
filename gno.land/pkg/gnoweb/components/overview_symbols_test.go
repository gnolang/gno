package components

import (
	"bytes"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/doc"
	"github.com/stretchr/testify/require"
)

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

func TestBuildSymbols_PureHasNoActionLink(t *testing.T) {
	t.Parallel()
	jdoc := &doc.JSONDocumentation{
		Funcs: []*doc.JSONFunc{{Name: "Hello"}},
	}
	funcs, _ := buildSymbols(jdoc, noopRenderer{}, "/p/demo/foo")
	require.Len(t, funcs, 1)
	require.Empty(t, funcs[0].ActionURL, "pure packages (/p/) expose no actions")
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
