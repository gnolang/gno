package components

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRenderJSONTree_BasicShape locks the per-token semantic classes and
// the <details>/<summary> structure — those names are the contract the
// CSS in 06-blocks.css depends on. Adding a new token type without
// updating CSS would silently render it unstyled; this test catches that
// by checking each token's class.
func TestRenderJSONTree_BasicShape(t *testing.T) {
	t.Parallel()

	html := string(RenderJSONTree([]byte(`{
		"name": "alice",
		"age": 30,
		"active": true,
		"deleted": null,
		"tags": ["admin", "user"]
	}`)))

	// Top-level object is a collapsible.
	assert.Contains(t, html, `<details class="json-obj"`)
	assert.Contains(t, html, `<summary>`)
	// Token classes that downstream CSS targets.
	assert.Contains(t, html, `class="key"`, "object keys carry .key class")
	assert.Contains(t, html, `class="str"`, "string values carry .str class")
	assert.Contains(t, html, `class="num"`, "numbers carry .num class")
	assert.Contains(t, html, `class="bool"`, "booleans carry .bool class")
	assert.Contains(t, html, `class="null"`, "nulls carry .null class")
	// Arrays render as their own collapsible.
	assert.Contains(t, html, `<details class="json-arr"`)
	// Encoded values (string with quotes already in output).
	assert.Contains(t, html, `&#34;alice&#34;`)
	assert.Contains(t, html, `>30<`)
	assert.Contains(t, html, `>true<`)
	assert.Contains(t, html, `>null<`)
}

// TestRenderJSONTree_OpenByDefault verifies the depth-2 open contract: the
// top two levels expand on first paint, deeper ones are collapsed. Saves a
// click on the most-visible structure while keeping huge realms scannable.
func TestRenderJSONTree_OpenByDefault(t *testing.T) {
	t.Parallel()

	// Three levels deep — outer + middle should be open, inner collapsed.
	html := string(RenderJSONTree([]byte(`{
		"L0": {"L1": {"L2": {"leaf": 1}}}
	}`)))

	openCount := strings.Count(html, "<details class=\"json-obj\" open>")
	closedCount := strings.Count(html, "<details class=\"json-obj\">")
	// All four levels (0..3) fall below the generous open threshold so
	// the JSON view shows the realm's payload fully expanded by default.
	assert.Equal(t, 4, openCount, "every level under the threshold opens")
	assert.Equal(t, 0, closedCount, "no level should be collapsed at this depth")
}

// TestRenderJSONTree_BadJSON: invalid input returns "" so the handler can
// fall back to chroma or plain <pre> — never a partial / corrupt render.
func TestRenderJSONTree_BadJSON(t *testing.T) {
	t.Parallel()

	cases := []string{
		`{not json}`,
		`{`,
		``,
		`<script>`,
	}
	for _, raw := range cases {
		out := RenderJSONTree([]byte(raw))
		assert.Empty(t, out, "bad input %q must yield empty so caller can fall back", raw)
	}
}

// TestRenderJSONTree_HTMLEscaping is the security gate: any string content
// from the chain that contains HTML-active characters must be escaped, not
// rendered as live markup. A regression here = realm-supplied XSS.
func TestRenderJSONTree_HTMLEscaping(t *testing.T) {
	t.Parallel()

	html := string(RenderJSONTree([]byte(`{
		"<script>alert(1)</script>": "<img src=x onerror=alert(1)>"
	}`)))

	assert.NotContains(t, html, "<script>alert(1)</script>",
		"key with script tag must be escaped")
	assert.NotContains(t, html, "<img src=x onerror=alert(1)>",
		"value with img injection must be escaped")
	assert.Contains(t, html, "&lt;script&gt;",
		"keys are escaped via html.EscapeString")
	assert.Contains(t, html, "&lt;img",
		"values are escaped via html.EscapeString")
}

// TestRenderJSONTree_EmptyContainers handles edges: empty objects/arrays
// render as inline `[]`/`{}` placeholders, not as open <details> with no
// content (which would look broken).
func TestRenderJSONTree_EmptyContainers(t *testing.T) {
	t.Parallel()

	html := string(RenderJSONTree([]byte(`{"empty_arr": [], "empty_obj": {}}`)))
	assert.Contains(t, html, `<span class="empty">[]</span>`)
	assert.Contains(t, html, `<span class="empty">{}</span>`)
}
