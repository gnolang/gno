package state

import (
	"context"
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// nestedStructValueFixture builds a qobject_json payload whose Value is a
// StructValue nested d levels deep — each level a single field "x" holding
// the next struct, the innermost an int = 1. The top-level Value is a bare
// StructValue; nested levels are TypedValue-wrapped (struct fields).
func nestedStructValueFixture(d int) string {
	var b strings.Builder
	b.WriteString(`{"objectid":"ffffffffffffffffffffffffffffffffffffffff:1","value":`)
	b.WriteString(`{"@type":"/gno.StructValue","Fields":[`)
	for i := 0; i < d-1; i++ {
		b.WriteString(`{"T":{"@type":"/gno.StructType","Fields":[]},"V":{"@type":"/gno.StructValue","Fields":[`)
	}
	b.WriteString(`{"T":{"@type":"/gno.PrimitiveType","value":"32"},"N":"AQAAAAAAAAA="}`)
	for i := 0; i < d-1; i++ {
		b.WriteString(`]}}`)
	}
	b.WriteString(`]}}`)
	return b.String()
}

// depthOfFirstChild walks the .Children[0] spine and returns how many
// levels deep the first KindTruncated sentinel sits (-1 if none found
// within limit levels).
func depthOfFirstChild(n StateNode, limit int) int {
	cur := n
	for d := 0; d <= limit; d++ {
		if cur.Kind == KindTruncated {
			return d
		}
		if len(cur.Children) == 0 {
			return -1
		}
		cur = cur.Children[0]
	}
	return -1
}

// TestDecodeObjectDepthBudget pins that DecodeObject honors
// cfg.MaxDecodeDepth: a struct nested far past the budget collapses to a
// "(too deep)" sentinel within MaxDecodeDepth levels instead of recursing
// the full walker.go maxDecodeDepth.
func TestDecodeObjectDepthBudget(t *testing.T) {
	t.Parallel()

	raw := []byte(nestedStructValueFixture(30))

	shallow, err := DecodeObject(context.Background(), raw, RenderConfig{
		MaxChildrenPerNode: maxChildrenPerNode,
		MaxDecodeDepth:     3,
	})
	require.NoError(t, err)
	require.Len(t, shallow.Children, 1)
	got := depthOfFirstChild(shallow.Children[0], 30)
	require.NotEqual(t, -1, got, "shallow cfg must produce a truncated sentinel")
	assert.LessOrEqual(t, got, 3, "sentinel must appear within MaxDecodeDepth levels")

	// Full-page cfg decodes the whole 30-deep tree — no sentinel.
	full, err := DecodeObject(context.Background(), raw, DefaultPageRenderConfig())
	require.NoError(t, err)
	require.Len(t, full.Children, 1)
	assert.Equal(t, -1, depthOfFirstChild(full.Children[0], 30),
		"full-depth cfg decodes the whole tree without truncation")
}

// TestDecodePackageDepthBudget pins the same depth-budget mirror logic for
// the qpkg_json entry point.
func TestDecodePackageDepthBudget(t *testing.T) {
	t.Parallel()

	raw := []byte(buildDeepStructFixture(30))

	nodes, total, err := DecodePackage(context.Background(), raw, RenderConfig{
		MaxChildrenPerNode: maxChildrenPerNode,
		MaxDecodeDepth:     3,
	}, 0, math.MaxInt32)
	require.NoError(t, err)
	require.Len(t, nodes, 1)
	assert.Equal(t, 1, total, "single top-level decl in fixture")
	got := depthOfFirstChild(nodes[0], 30)
	require.NotEqual(t, -1, got, "shallow cfg must truncate the deep package value")
	assert.LessOrEqual(t, got, 3)
}

// The typed fragment path (DecodeObjectFull with a &tid=) must honor the
// shallow per-fragment depth budget exactly like the untyped path. Without
// this, a deep struct under a valid tid would recurse to maxDecodeDepth
// (256), violating the per-fragment ≤3 bound.
func TestDecodeObjectFullTypedHonorsDepth(t *testing.T) {
	t.Parallel()

	// Object: a HeapItemValue wrapping a deeply nested StructValue. The
	// outermost struct is the HeapItem's inner Value; nested levels are
	// TypedValue-wrapped struct fields.
	const depth = 30
	var v strings.Builder
	v.WriteString(`{"@type":"/gno.StructValue","Fields":[`)
	for i := 0; i < depth-1; i++ {
		v.WriteString(`{"T":{"@type":"/gno.StructType","Fields":[]},"V":{"@type":"/gno.StructValue","Fields":[`)
	}
	v.WriteString(`{"T":{"@type":"/gno.PrimitiveType","value":"32"},"N":"AQAAAAAAAAA="}`)
	for i := 0; i < depth-1; i++ {
		v.WriteString(`]}}`)
	}
	v.WriteString(`]}`)
	objectJSON := `{"objectid":"ffffffffffffffffffffffffffffffffffffffff:1","value":` +
		`{"@type":"/gno.HeapItemValue","Value":{"T":{"@type":"/gno.RefType","ID":"gno.land/r/x.T"},"V":` +
		v.String() + `}}}`

	// Type: a StructType so the typed path's struct branch fires.
	const typeJSON = `{
		"typeid": "gno.land/r/x.T",
		"type": {"@type": "/gno.StructType", "PkgPath": "gno.land/r/x", "Fields": [
			{"Name": "x", "Type": {"@type": "/gno.StructType", "Fields": []}, "Embedded": false, "Tag": ""}
		]}
	}`

	// Shallow per-fragment cfg: the typed decode must truncate within the cap.
	decoded, err := DecodeObjectFull([]byte(objectJSON), []byte(typeJSON), DefaultFragmentRenderConfig())
	require.NoError(t, err)
	require.NotEmpty(t, decoded.Nodes)
	got := depthOfFirstChild(decoded.Nodes[0], depth)
	require.NotEqual(t, -1, got,
		"typed fragment decode must truncate within the depth budget")
	assert.LessOrEqual(t, got, DefaultFragmentRenderConfig().MaxDecodeDepth,
		"typed decode must honor cfg.MaxDecodeDepth, not the full maxDecodeDepth")

	// Full-page cfg decodes the whole tree — confirms the budget, not a
	// hard-coded shallow cap, is what truncates.
	full, err := DecodeObjectFull([]byte(objectJSON), []byte(typeJSON), DefaultPageRenderConfig())
	require.NoError(t, err)
	require.NotEmpty(t, full.Nodes)
	assert.Equal(t, -1, depthOfFirstChild(full.Nodes[0], depth),
		"full-page cfg decodes the typed tree without truncation")
}

// TestClampRenderConfigDefaults locks the zero-value behavior: a bare
// RenderConfig{} still yields a working bounded decode.
func TestClampRenderConfigDefaults(t *testing.T) {
	t.Parallel()

	got := clampRenderConfig(RenderConfig{})
	assert.Equal(t, maxChildrenPerNode, got.MaxChildrenPerNode)
	assert.Equal(t, 3, got.MaxDecodeDepth, "zero depth defaults to the fragment budget")

	over := clampRenderConfig(RenderConfig{MaxDecodeDepth: maxDecodeDepth + 100})
	assert.Equal(t, maxDecodeDepth, over.MaxDecodeDepth, "depth is capped at maxDecodeDepth")

	// startDepthFor is always >= 0 after the clamp.
	assert.GreaterOrEqual(t, startDepthFor(RenderConfig{}), 0)
	assert.Equal(t, maxDecodeDepth-3, startDepthFor(DefaultFragmentRenderConfig()))
	assert.Equal(t, 0, startDepthFor(DefaultPageRenderConfig()))
}
