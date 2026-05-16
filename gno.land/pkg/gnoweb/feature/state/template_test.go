package state

import (
	"bytes"
	"html/template"
	"strings"
	"testing"
)

// TestTemplateFunc_HeadingForKind pins the parent-Kind → child-heading
// mapping used by the decl-children grid.
func TestTemplateFunc_HeadingForKind(t *testing.T) {
	t.Parallel()
	fn := funcMap["headingForKind"].(func(string) string)
	cases := map[string]string{
		KindMap:    "Key",
		KindSlice:  "Index",
		KindArray:  "Index",
		KindStruct: "Field",
		"unknown":  "Field",
	}
	for in, want := range cases {
		if got := fn(in); got != want {
			t.Errorf("headingForKind(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestTemplateFunc_KindIconID pins Kind+Type → sprite ID. Symbols are
// defined in ui/icons.html; any rename here must update the SVG file.
func TestTemplateFunc_KindIconID(t *testing.T) {
	t.Parallel()
	fn := funcMap["kindIconID"].(func(string, string) string)
	cases := []struct {
		kind, typ string
		want      string
	}{
		{KindPrimitive, "string", "kind-string"},
		{KindPrimitive, "bool", "kind-bool"},
		{KindPrimitive, "int64", "kind-number"},
		{KindStruct, "", "kind-struct"},
		{KindMap, "", "kind-map"},
		{KindSlice, "", "kind-slice"},
		{KindArray, "", "kind-slice"},
		{KindPointer, "", "kind-pointer"},
		{KindFunc, "", "kind-func"},
		{KindClosure, "", "kind-closure"},
		{KindRef, "", "kind-ref"},
		{KindNil, "", "kind-nil"},
		{KindPackage, "", "kind-package"},
		{KindType, "", "kind-type"},
		{KindInterface, "", "kind-interface"},
		{"unknown", "", "kind-unknown"},
	}
	for _, c := range cases {
		if got := fn(c.kind, c.typ); got != c.want {
			t.Errorf("kindIconID(%q,%q) = %q, want %q", c.kind, c.typ, got, c.want)
		}
	}
}

// TestTemplateFunc_KindGroup pins the kind-filter taxonomy used by the
// CSS attribute selectors that hide unrelated cards.
func TestTemplateFunc_KindGroup(t *testing.T) {
	t.Parallel()
	fn := funcMap["kindGroup"].(func(string) string)
	cases := map[string]string{
		KindStruct:    "state",
		KindMap:       "state",
		KindSlice:     "state",
		KindArray:     "state",
		KindPointer:   "state",
		KindRef:       "state",
		KindFunc:      "code",
		KindClosure:   "code",
		KindType:      "types",
		KindInterface: "types",
		KindPrimitive: "other",
		"unknown":     "other",
	}
	for in, want := range cases {
		if got := fn(in); got != want {
			t.Errorf("kindGroup(%q) = %q, want %q", in, got, want)
		}
	}
}

// mustParse panics on parse failure, so init alone guards the embed
// glob; this test pins the var names.
func TestTemplatesParse(t *testing.T) {
	cases := []struct {
		name string
		tmpl *template.Template
	}{
		{"PageTemplate", PageTemplate},
		{"FragNodeTemplate", FragNodeTemplate},
		{"FragSourceTemplate", FragSourceTemplate},
		{"FragErrorTemplate", FragErrorTemplate},
	}
	for _, c := range cases {
		if c.tmpl == nil {
			t.Errorf("%s is nil", c.name)
		}
	}
}

func TestPageTemplateRendersBasic(t *testing.T) {
	data := StateData{
		PkgPath:    "/r/test",
		CountLabel: "1 declaration",
		Nodes: []StateNode{
			{
				Name:       "MyMap",
				Type:       "map[string]int",
				Kind:       KindMap,
				ObjectID:   "abc123:1",
				Expandable: true,
				Anchor:     "my-map",
			},
		},
		KindCounts: KindCounts{All: 1, State: 1},
	}
	var buf bytes.Buffer
	if err := PageTemplate.ExecuteTemplate(&buf, "renderPage", data); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := buf.String()

	must := []string{
		`<meta name="htmx-config"`,
		`"allowEval":false`,
		`"allowScriptTags":false`,
		`"selfRequestsOnly":true`,
		`<script type="application/json" id="state-doc-index">`,
		// Gnoweb $webargs URL (alphabetical sort puts `frag` first, `state` last).
		// `&` in the attribute interpolation gets HTML-escaped to `&amp;`.
		`hx-get="/r/test$frag=node`,
		`class="b-state-permalink"`,
	}
	for _, m := range must {
		if !strings.Contains(out, m) {
			t.Errorf("output missing %q\n--- output (head) ---\n%s", m, head(out, 1500))
		}
	}
	// htmx is bundled into controller-state.js via esbuild now — no
	// separate <script src="...htmx..."> tag should appear in the page.
	if strings.Contains(out, "htmx-2.0.4.min.js") {
		t.Errorf("unexpected vendored htmx script tag in page output (htmx should be bundled into controller-state.js)")
	}
}

func TestFragNodeTemplateRenders(t *testing.T) {
	data := FragNodeData{
		Node: StateNode{
			Name: "root",
			Kind: KindStruct,
			Children: []StateNode{
				{Name: "field1", Type: "string", Kind: KindPrimitive, Value: "hello"},
			},
		},
		HeightParam: "42",
	}
	var buf bytes.Buffer
	if err := FragNodeTemplate.ExecuteTemplate(&buf, "fragNode", data); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := buf.String()
	// `<head>` / `<head ` — not bare `<head`, which substring-matches the
	// legit `<header class="fields-head">` in the pretty fields table.
	forbid := []string{"<html", "<head>", "<head ", "<script", "<!doctype"}
	for _, f := range forbid {
		if strings.Contains(strings.ToLower(out), f) {
			t.Errorf("fragment unexpectedly contains chrome %q", f)
		}
	}
	if !strings.Contains(out, "field1") {
		t.Errorf("fragment missing child name; got:\n%s", out)
	}
}

func TestFragErrorTemplateRenders(t *testing.T) {
	data := FragErrorData{Message: "rate limit exceeded", RetryHint: "retry in 30s"}
	var buf bytes.Buffer
	if err := FragErrorTemplate.ExecuteTemplate(&buf, "fragError", data); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := buf.String()
	for _, m := range []string{`role="alert"`, "rate limit exceeded", "retry in 30s"} {
		if !strings.Contains(out, m) {
			t.Errorf("output missing %q; got:\n%s", m, out)
		}
	}
}

func TestFragSourceRenders(t *testing.T) {
	data := FragSourceData{
		SourceHTML:  template.HTML(`<pre class="chroma">func Hi() {}</pre>`),
		PkgPath:     "/r/test",
		File:        "main.gno",
		Line:        12,
		HeightParam: "99",
	}
	var buf bytes.Buffer
	if err := FragSourceTemplate.ExecuteTemplate(&buf, "fragSource", data); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := buf.String()
	for _, m := range []string{
		"main.gno:12",
		`<pre class="chroma">func Hi() {}</pre>`,
		// permalink uses the routable $webargs grammar with the pkg path.
		`/r/test$`,
		`file=main.gno`,
		`source`,
		`height=99`,
		`#L12`,
		"See in code",
	} {
		if !strings.Contains(out, m) {
			t.Errorf("output missing %q; got:\n%s", m, out)
		}
	}
	if strings.Contains(out, `href="?source`) {
		t.Errorf("permalink must not use the dead relative ?source form; got:\n%s", out)
	}
}

// Stale-while-revalidate invariant: every hx-get URL inherits HeightParam.
func TestHeightStampedIntoHxGet(t *testing.T) {
	data := StateData{
		PkgPath:     "/r/test",
		HeightParam: "12345",
		Nodes: []StateNode{
			{Name: "A", Kind: KindStruct, ObjectID: "abc:1", Expandable: true, Anchor: "a"},
			{Name: "B", Kind: KindMap, ObjectID: "def:2", Expandable: true, Anchor: "b"},
			{Name: "Fn", Kind: KindFunc, Source: &SourceLocation{File: "f.gno", StartLine: 5, EndLine: 9}},
		},
	}
	var buf bytes.Buffer
	if err := PageTemplate.ExecuteTemplate(&buf, "renderPage", data); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := buf.String()

	hxGetCount := strings.Count(out, `hx-get="`)
	if hxGetCount == 0 {
		t.Fatalf("expected at least one hx-get attribute, got none\n%s", head(out, 1500))
	}
	// Helper-built template.URL values are HTML-escaped on interpolation,
	// so the height stamp appears as `&amp;height=12345` in the body.
	heightCount := strings.Count(out, `&amp;height=12345`)
	if heightCount < hxGetCount {
		t.Errorf("hx-get count %d > stamped &amp;height=12345 count %d — at least one fragment URL is missing the height stamp",
			hxGetCount, heightCount)
	}
}

// Every htmx-driven row must have a no-JS fallback. Tree view +
// source-details use the sibling b-state-permalink; pretty view falls
// back to the enclosing card's Open/Inspect CTA. Heuristic:
// permalink + card-CTA count ≥ hx-get count.
func TestPermanentLinkPresentForEveryHxGet(t *testing.T) {
	data := StateData{
		PkgPath: "/r/test",
		Nodes: []StateNode{
			{Name: "A", Kind: KindStruct, ObjectID: "abc:1", Expandable: true, Anchor: "a"},
			{Name: "B", Kind: KindMap, ObjectID: "def:2", Expandable: true, Anchor: "b"},
			{Name: "Fn", Kind: KindFunc, Source: &SourceLocation{File: "f.gno", StartLine: 5, EndLine: 9}},
		},
	}
	// Production always enriches links before render — the pretty-view
	// card CTA is the no-JS fallback for state/node-details.
	EnrichLinks(data.Nodes, data.PkgPath, "", "")
	var buf bytes.Buffer
	if err := PageTemplate.ExecuteTemplate(&buf, "renderPage", data); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := buf.String()
	hxGetCount := strings.Count(out, `hx-get="`)
	fallbacks := strings.Count(out, `class="b-state-permalink"`) +
		strings.Count(out, `b-btn--ghost cta`)
	if hxGetCount == 0 {
		t.Fatalf("expected hx-get nodes in output\n%s", head(out, 1500))
	}
	if fallbacks < hxGetCount {
		t.Errorf("no-JS fallback count %d < hx-get count %d — some htmx rows are dead-ends without JS", fallbacks, hxGetCount)
	}
}

// TestNodeDetailsSummaryCarriesHint — an unresolved lazy node's <summary>
// must read as an expansion affordance (name + muted "details" hint), not
// a bare repeat of the card header.
func TestNodeDetailsSummaryCarriesHint(t *testing.T) {
	data := StateData{
		PkgPath: "/r/test",
		Nodes: []StateNode{
			{Name: "Matrix", Type: "[9]int", Kind: KindArray, ObjectID: "abc:42", Expandable: true, Anchor: "matrix"},
		},
	}
	var buf bytes.Buffer
	if err := PageTemplate.ExecuteTemplate(&buf, "renderPage", data); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `class="b-state-node-name">Matrix<`) {
		t.Errorf("node-details summary missing name span\n%s", head(out, 1500))
	}
	if !strings.Contains(out, `class="b-state-node-hint">details<`) {
		t.Errorf("node-details summary missing 'details' hint affordance\n%s", head(out, 1500))
	}
}

// TestPageTemplateHasHtmxConfigMeta pins the forced htmx-config meta
// flags. htmx itself is bundled into controller-state.js by esbuild
// (no separate <script src> tag), and reads this meta at boot whether
// loaded externally or bundled.
func TestPageTemplateHasHtmxConfigMeta(t *testing.T) {
	data := StateData{
		PkgPath: "/r/test",
		Nodes:   []StateNode{{Name: "X", Kind: KindStruct}},
	}
	var buf bytes.Buffer
	if err := PageTemplate.ExecuteTemplate(&buf, "renderPage", data); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := buf.String()
	must := []string{
		`<meta name="htmx-config"`,
		`"allowEval":false`,
		`"allowScriptTags":false`,
		`"selfRequestsOnly":true`,
		`"includeIndicatorStyles":false`,
		`"defaultSwapStyle":"innerHTML"`,
		`"historyCacheSize":0`,
	}
	for _, m := range must {
		if !strings.Contains(out, m) {
			t.Errorf("output missing %q\n--- output (head) ---\n%s", m, head(out, 1500))
		}
	}
}

// head returns the first n bytes of s for error context.
func head(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…(truncated)"
}

// TestPageTemplateHTMLEscapesValues — Names/Types/Values carry untrusted
// realm-source bytes; html/template MUST escape them in attribute and
// text contexts. (Ported from the deleted legacy view_test.go.)
func TestPageTemplateHTMLEscapesValues(t *testing.T) {
	data := StateData{
		PkgPath:    "/r/test",
		CountLabel: "1 decl",
		Nodes: []StateNode{{
			Name: `Evil<script>`, Type: `map[string]<bad>`, Kind: KindMap,
			Value: `"; alert(1); //`, Anchor: "evil",
		}},
		KindCounts: KindCounts{All: 1, State: 1},
	}
	var buf bytes.Buffer
	if err := PageTemplate.ExecuteTemplate(&buf, "renderPage", data); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := buf.String()
	for _, bad := range []string{
		"<script>",
		"<bad>",
		`"; alert(1)`,
	} {
		if strings.Contains(out, bad) {
			t.Errorf("output contains raw %q — escaping failed", bad)
		}
	}
}

// TestPageTemplateSidebarTOC — TOC entries must surface per-card anchors
// (#state-<name>-pretty / -tree). Ported from view_test.go sidebar tests.
func TestPageTemplateSidebarTOC(t *testing.T) {
	data := StateData{
		PkgPath: "/r/test",
		Nodes: []StateNode{
			{Name: "Counter", Kind: KindStruct, Anchor: "state-counter"},
		},
		Sidebar: &StateSidebar{
			Heading: "On this page",
			TOC: []StateTOCEntry{
				{
					Label: "Counter", Anchor: "state-counter", Kind: KindStruct, Type: "Counter",
					PrettyHref: "#state-counter-pretty", TreeHref: "#state-counter-tree", OnPage: true,
				},
			},
		},
		KindCounts: KindCounts{All: 1, State: 1},
	}
	var buf bytes.Buffer
	if err := PageTemplate.ExecuteTemplate(&buf, "renderPage", data); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := buf.String()
	for _, m := range []string{
		`href="#state-counter-pretty"`,
		`href="#state-counter-tree"`,
		`data-name="Counter"`,
	} {
		if !strings.Contains(out, m) {
			t.Errorf("output missing %q\n--- head ---\n%s", m, head(out, 1500))
		}
	}
}

// TestPageTemplateStatsStripForObject — non-primitive nodes with audit
// metadata (OID, Hash, Size, RefCount, Owner) surface a b-state-stats
// strip in the decl card. Ported from view_test.go.
func TestPageTemplateStatsStripForObject(t *testing.T) {
	data := StateData{
		PkgPath:    "/r/test",
		CountLabel: "1 decl",
		Nodes: []StateNode{{
			Name: "C", Kind: KindStruct, ObjectID: "abc:1",
			OwnerID: "own:9", RefCount: "3", LastObjectSize: "128",
			Anchor: "c",
		}},
		KindCounts: KindCounts{All: 1, State: 1},
	}
	var buf bytes.Buffer
	if err := PageTemplate.ExecuteTemplate(&buf, "renderPage", data); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := buf.String()
	for _, m := range []string{
		`class="b-state-stats"`, `>OID<`, `>Size<`, `>Refs<`, `>Owner<`,
	} {
		if !strings.Contains(out, m) {
			t.Errorf("stats strip missing %q\n--- head ---\n%s", m, head(out, 1500))
		}
	}
}

// TestPageTemplateNoStatsForPrimitive — primitive leafs (no OID, no
// audit fields) must NOT render the stats strip. Ported from view_test.go.
func TestPageTemplateNoStatsForPrimitive(t *testing.T) {
	data := StateData{
		PkgPath:    "/r/test",
		CountLabel: "1 decl",
		Nodes: []StateNode{{
			Name: "n", Kind: KindPrimitive, Type: "int", Value: "42", Anchor: "n",
		}},
		KindCounts: KindCounts{All: 1},
	}
	var buf bytes.Buffer
	if err := PageTemplate.ExecuteTemplate(&buf, "renderPage", data); err != nil {
		t.Fatalf("render: %v", err)
	}
	if strings.Contains(buf.String(), `class="b-state-stats"`) {
		t.Errorf("primitive leaf must not emit b-state-stats")
	}
}

// TestPageTemplateCrumbsBackLink — when Crumbs[0] has an Href, the
// header emits a back-link to the parent. Ported from view_test.go.
func TestPageTemplateCrumbsBackLink(t *testing.T) {
	data := StateData{
		PkgPath:    "/r/test",
		CountLabel: "Object abcd…",
		Crumbs: []StateCrumb{
			{Label: "demo", Href: template.URL("/r/demo")},
		},
		Nodes:      []StateNode{},
		KindCounts: KindCounts{},
	}
	var buf bytes.Buffer
	if err := PageTemplate.ExecuteTemplate(&buf, "renderPage", data); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `class="back-link"`) {
		t.Errorf("back-link missing; head=\n%s", head(out, 1200))
	}
	if !strings.Contains(out, `aria-label="Back to demo"`) {
		t.Errorf("back-link aria-label missing")
	}
	// The active label comes from CountLabel, not the trailing crumb.
	if !strings.Contains(out, "Object abcd…") {
		t.Errorf("CountLabel must surface in the page header")
	}
}
