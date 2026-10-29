package run

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gnovm/pkg/doc"
)

func TestRunDataPkgAlias(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		pkgPath string
		want    string
	}{
		"unversioned realm":  {"gno.land/r/demo/boards", "boards"},
		"versioned realm v0": {"gno.land/r/gnoland/boards2/v0", "boards2"},
		"versioned pure v1":  {"gno.land/p/nt/avl/v1", "avl"},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, RunData{PkgPath: tc.pkgPath}.PkgAlias())
		})
	}
}

func renderPage(t *testing.T, data RunData) string {
	t.Helper()
	var buf strings.Builder
	require.NoError(t, NewPageView(data).Render(&buf))
	return buf.String()
}

func TestPageGasDefaults(t *testing.T) {
	t.Parallel()

	page := renderPage(t, RunData{PkgPath: "gno.land/r/demo/counter", Gas: components.RunGas})
	assert.Contains(t, page, `value="`+components.RunGas.Wanted+`"`)
	assert.Contains(t, page, `value="`+components.RunGas.Fee+`"`)
	assert.NotEqual(t, components.CallGas, components.RunGas, "a script asks for less than $help's calls")
}

func TestStarterScript(t *testing.T) {
	t.Parallel()

	str := func(name string) *doc.JSONField { return &doc.JSONField{Name: name, Type: "string"} }
	render := &doc.JSONFunc{Name: "Render", Params: []*doc.JSONField{str("path")}, Results: []*doc.JSONField{str("")}}
	crossingRender := &doc.JSONFunc{
		Name: "Render", Crossing: true,
		Params:  []*doc.JSONField{{Name: "cur", Type: "realm"}, str("path")},
		Results: []*doc.JSONField{str("")},
	}
	method := &doc.JSONFunc{Name: "Render", Type: "Board", Params: []*doc.JSONField{str("path")}, Results: []*doc.JSONField{str("")}}

	cases := map[string]struct {
		funcs []*doc.JSONFunc
		want  string
	}{
		"render": {
			[]*doc.JSONFunc{{Name: "Increment"}, render},
			`println(counter.Render(""))`,
		},
		"crossing render": {
			[]*doc.JSONFunc{crossingRender},
			`println(counter.Render(cross(cur), ""))`,
		},
		"no render: first exported function": {
			[]*doc.JSONFunc{method, {Name: "helper"}, {Name: "NewGroup"}, {Name: "Other"}},
			"_ = counter.NewGroup",
		},
		"render of another shape": {
			[]*doc.JSONFunc{{Name: "Render", Results: []*doc.JSONField{str("")}}},
			"_ = counter.Render",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			script := starterScript("gno.land/r/demo/counter", tc.funcs)
			assert.Contains(t, script, "import \"gno.land/r/demo/counter\"\n")
			assert.Contains(t, script, "\n\t"+tc.want+"\n")
			// Only a crossing call needs main's realm.
			assert.Equal(t, strings.Contains(tc.want, "cross"), strings.Contains(script, "func main(cur realm) {"), script)
		})
	}

	t.Run("nothing to call", func(t *testing.T) {
		t.Parallel()
		script := starterScript("gno.land/p/demo/consts", []*doc.JSONFunc{method})
		assert.Contains(t, script, "import _ \"gno.land/p/demo/consts\"\n")
	})
}
