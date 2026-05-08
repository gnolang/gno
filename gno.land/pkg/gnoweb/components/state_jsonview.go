package components

import (
	"bytes"
	"encoding/json"
	htmlpkg "html"
	"html/template"
	"sort"
	"strconv"
	"strings"
)

// RenderJSONTree turns a JSON payload into a collapsible HTML tree using
// native <details>/<summary>. Top-level objects and arrays are open by
// default through depth 1; deeper levels are collapsed so the page stays
// scannable on big realms. Returns "" if the JSON can't be parsed —
// callers can fall back to a flat <pre> render.
//
// Output shape (per object):
//
//	<details class="json-obj" open>
//	  <summary>{</summary>
//	  <div class="json-body">
//	    <div class="entry"><span class="key">"name"</span>: <span class="str">"alice"</span>,</div>
//	    ...
//	  </div>
//	  <span class="json-close">}</span>
//	</details>
//
// Tokens carry semantic classes (key/str/num/bool/null) so CSS can colour
// them — equivalent to chroma's output but with collapse semantics.
//
// Note: encoding/json parses objects as map[string]any without preserving
// insertion order, so keys are emitted in lexicographic order. For state-
// explorer dumps this is acceptable: stable, scannable, and the chain JSON
// is structured enough that alphabetical order doesn't hurt readability.
func RenderJSONTree(raw []byte) template.HTML {
	var v any
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber() // keep numbers as the source string (avoid float64 stringification)
	if err := dec.Decode(&v); err != nil {
		return ""
	}
	var b strings.Builder
	writeJSONNode(&b, v, 0)
	return template.HTML(b.String())
}

// openDepthThreshold caps how deep <details> open by default. The JSON
// view exists for power users who explicitly toggled away from the
// curated UI — hiding nested data behind collapse-by-default defeats
// that purpose. Threshold is generous (covers TypedValue → T/V → type
// internals → fields) so the realistic realm payload renders fully
// expanded; users can still click any <summary> to collapse on demand.
const openDepthThreshold = 8

func writeJSONNode(b *strings.Builder, v any, depth int) {
	switch x := v.(type) {
	case nil:
		b.WriteString(`<span class="null">null</span>`)
	case bool:
		if x {
			b.WriteString(`<span class="bool" data-value="true">true</span>`)
		} else {
			b.WriteString(`<span class="bool" data-value="false">false</span>`)
		}
	case json.Number:
		b.WriteString(`<span class="num">`)
		b.WriteString(htmlpkg.EscapeString(string(x)))
		b.WriteString(`</span>`)
	case string:
		b.WriteString(`<span class="str">`)
		b.WriteString(htmlpkg.EscapeString(strconv.Quote(x)))
		b.WriteString(`</span>`)
	case []any:
		writeJSONArray(b, x, depth)
	case map[string]any:
		writeJSONObject(b, x, depth)
	default:
		// Fallback for any unexpected concrete type — shouldn't happen
		// for valid JSON parsed via encoding/json.
		b.WriteString(`<span class="raw">`)
		b.WriteString(htmlpkg.EscapeString(toString(x)))
		b.WriteString(`</span>`)
	}
}

func writeJSONArray(b *strings.Builder, items []any, depth int) {
	if len(items) == 0 {
		b.WriteString(`<span class="empty">[]</span>`)
		return
	}
	b.WriteString(`<details class="json-arr"`)
	if depth < openDepthThreshold {
		b.WriteString(` open`)
	}
	b.WriteString(`><summary><span class="bracket">[</span>`)
	b.WriteString(`<span class="meta"> ` + strconv.Itoa(len(items)) + ` items</span>`)
	b.WriteString(`</summary>`)
	b.WriteString(`<div class="json-body">`)
	for i, item := range items {
		b.WriteString(`<div class="entry">`)
		writeJSONNode(b, item, depth+1)
		if i < len(items)-1 {
			b.WriteString(`<span class="sep">,</span>`)
		}
		b.WriteString(`</div>`)
	}
	b.WriteString(`</div>`)
	b.WriteString(`<span class="json-close bracket">]</span></details>`)
}

func writeJSONObject(b *strings.Builder, obj map[string]any, depth int) {
	if len(obj) == 0 {
		b.WriteString(`<span class="empty">{}</span>`)
		return
	}
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	b.WriteString(`<details class="json-obj"`)
	if depth < openDepthThreshold {
		b.WriteString(` open`)
	}
	b.WriteString(`><summary><span class="bracket">{</span>`)
	b.WriteString(`<span class="meta"> ` + strconv.Itoa(len(obj)) + ` keys</span>`)
	b.WriteString(`</summary>`)
	b.WriteString(`<div class="json-body">`)
	for i, k := range keys {
		b.WriteString(`<div class="entry">`)
		b.WriteString(`<span class="key">`)
		b.WriteString(htmlpkg.EscapeString(strconv.Quote(k)))
		b.WriteString(`</span><span class="sep">:</span> `)
		writeJSONNode(b, obj[k], depth+1)
		if i < len(keys)-1 {
			b.WriteString(`<span class="sep">,</span>`)
		}
		b.WriteString(`</div>`)
	}
	b.WriteString(`</div>`)
	b.WriteString(`<span class="json-close bracket">}</span></details>`)
}

// toString is a tiny stringifier used only in the fallback branch — covers
// integer types in case json.Decoder somehow yields one despite UseNumber.
func toString(v any) string {
	switch x := v.(type) {
	case int:
		return strconv.Itoa(x)
	case int64:
		return strconv.FormatInt(x, 10)
	case uint64:
		return strconv.FormatUint(x, 10)
	case float64:
		return strconv.FormatFloat(x, 'g', -1, 64)
	}
	return ""
}
