package run

import (
	"html/template"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gnovm/pkg/doc"
)

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
		want  string // the script's statements, below its first comment
	}{
		"render": {
			[]*doc.JSONFunc{{Name: "Increment"}, render},
			"// println(counter.Render(\"\"))\n\t_ = counter.Render",
		},
		"crossing render": {
			[]*doc.JSONFunc{crossingRender},
			"// println(counter.Render(cross(cur), \"\"))\n\t_ = counter.Render",
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
			assert.Contains(t, script, "e.g.:\n\t"+tc.want+"\n}\n")
			// Render runs only once uncommented: its gas is the package's.
			assert.NotContains(t, script, "\tprintln(", script)
			// Only a crossing call needs main's realm.
			assert.Equal(t, strings.Contains(tc.want, "cross"), strings.Contains(script, "func main(cur realm) {"), script)
		})
	}

	t.Run("versioned path", func(t *testing.T) {
		t.Parallel()
		script := starterScript("gno.land/r/gnoland/boards2/v0", []*doc.JSONFunc{render})
		assert.Contains(t, script, "\t_ = boards2.Render\n")
	})

	t.Run("nothing to call", func(t *testing.T) {
		t.Parallel()
		script := starterScript("gno.land/p/demo/consts", []*doc.JSONFunc{method})
		assert.Contains(t, script, "import _ \"gno.land/p/demo/consts\"\n")
	})
}

// TestPageServedWhole checks that the page holds the script and the command
// as served, without JS: the controller only keeps them in step.
func TestPageServedWhole(t *testing.T) {
	t.Parallel()

	data := RunData{
		PkgPath: "gno.land/r/demo/counter",
		Remote:  "https://rpc.example:443",
		ChainId: "test-1",
		Gas:     components.RunGas,
		Script:  starterScript("gno.land/r/demo/counter", nil),
	}
	page := renderPage(t, data)
	script := template.HTMLEscapeString(data.Script)

	// The script, shown in the editor's place and written by the heredoc.
	assert.Contains(t, page, `<pre class="chroma-chroma">`+script+`</pre>`)
	assert.Contains(t, page, "cat &gt; script.gno &lt;&lt;'GNO_EOF'\n"+script+"GNO_EOF\n\n")

	// Full Security is shown and copied first, as on $help; Fast is there to
	// switch to, with $help's warning.
	assert.Contains(t, page, `<span data-run-target="mode" data-run-mode-value="secure" data-copy-target="run-cmd" class="u-inline">`)
	assert.Contains(t, page, `<span data-run-target="mode" data-run-mode-value="fast" class="u-hidden"># WARNING`)
	for _, line := range []string{
		`gnokey query -remote "https://rpc.example:443" auth/accounts/<span data-run-target="address-arg">ADDRESS</span>`,
		`gnokey maketx run -gas-fee <span data-run-target="gas-fee-arg">100000ugnot</span> -gas-wanted <span data-run-target="gas-wanted-arg">50_000_000</span><span data-run-target="send-arg"></span> -broadcast=false <span data-run-target="key-arg">KEYNAME</span> script.gno &gt; run.tx`,
		`gnokey sign -tx-path run.tx -chainid "test-1" -account-number ACCOUNTNUMBER -account-sequence SEQUENCENUMBER <span data-run-target="key-arg">KEYNAME</span>`,
		`gnokey broadcast -remote "https://rpc.example:443" run.tx`,
		`-chainid "test-1" -remote "https://rpc.example:443" <span data-run-target="key-arg">KEYNAME</span> script.gno</span>`,
	} {
		assert.Contains(t, page, line)
	}
	assert.NotContains(t, page, "&lt;key-name&gt;", "a shell reads <key-name> as a redirect")

	// It says what running it does, and takes $help's mode and key.
	assert.Contains(t, page, `signs a transaction with <code class="u-font-mono" data-run-target="key-arg">KEYNAME</code>`)
	assert.Contains(t, page, `to the <code class="u-font-mono">test-1</code> chain`)
	assert.Contains(t, page, `even when the script runs out of gas: raise <code class="u-font-mono">-gas-wanted</code>`)
	assert.Contains(t, page, `data-controller="action-header"`)
	assert.Contains(t, page, `data-action-header-target="mode"`)
	assert.Contains(t, page, `data-action-header-target="address"`)
	assert.Contains(t, page, `class="b-btn b-switch" data-run-target="js-only" hidden>`)

	// Reset, Download and the script switch do nothing without JS: the
	// controller shows them.
	assert.Contains(t, page, `<div class="b-btns" data-run-target="js-only" hidden>`)
}

func TestPageHighlightedScript(t *testing.T) {
	t.Parallel()

	page := renderPage(t, RunData{Script: "package main\n", ScriptHTML: `<pre class="chroma-chroma"><span class="chroma-kn">package</span> main</pre>`})
	assert.Contains(t, page, `<div class="b-source-code"><pre class="chroma-chroma"><span class="chroma-kn">package</span> main</pre></div>`)
}
