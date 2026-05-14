package state

import (
	"context"
	"fmt"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/amino"
)

// RenderConfig holds the per-call decode bounds the slim ADR-004 path
// applies to one object's Amino payload. Distinct from walker.go's
// package-wide constants because fragments use a shallow depth budget
// (≤3) while the legacy full-page path stays at the original 256.
type RenderConfig struct {
	// MaxChildrenPerNode caps the visible children of any collection
	// node — surplus collapses to one KindTruncated sentinel. Mirrors
	// walker.go's package-wide bound; exposed here so fragment-mode
	// callers can tighten it if needed.
	MaxChildrenPerNode int
	// MaxDecodeDepth bounds recursion depth for this single decode.
	// ADR-004 §Resource bounds: ≤3 for fragment-mode rendering, 256
	// for full-page legacy parity.
	MaxDecodeDepth int
}

// DefaultFragmentRenderConfig is the slim per-fragment budget from
// ADR-004 §Resource bounds.
func DefaultFragmentRenderConfig() RenderConfig {
	return RenderConfig{
		MaxChildrenPerNode: maxChildrenPerNode,
		MaxDecodeDepth:     3,
	}
}

// DefaultPageRenderConfig is the full-depth budget for the legacy
// full-page path — parity with walker.go's package-wide constants.
func DefaultPageRenderConfig() RenderConfig {
	return RenderConfig{
		MaxChildrenPerNode: maxChildrenPerNode,
		MaxDecodeDepth:     maxDecodeDepth,
	}
}

// startDepthFor pre-offsets the recursion depth so the walker's global
// maxDecodeDepth bound fires after exactly cfg.MaxDecodeDepth levels.
// After clampRenderConfig the result is always >= 0.
func startDepthFor(cfg RenderConfig) int {
	return maxDecodeDepth - clampRenderConfig(cfg).MaxDecodeDepth
}

// DecodeObject decodes one qobject_json payload into a root StateNode
// whose Children are the decoded fields/elements, bounded by cfg.
// Refs surface as KindRef nodes; ExportRefValue cycle markers render as
// KindCycle (the gnovm exporter's per-export scope is what bounds
// cycles — ADR-004 §Consequences §Negative).
func DecodeObject(ctx context.Context, raw []byte, cfg RenderConfig) (StateNode, error) {
	if err := ctx.Err(); err != nil {
		return StateNode{}, err
	}
	cfg = clampRenderConfig(cfg)

	var resp objectResponse
	if err := amino.UnmarshalJSON(raw, &resp); err != nil {
		return StateNode{}, fmt.Errorf("decode object JSON: %w", err)
	}

	root := StateNode{
		Name:     "(object)",
		Kind:     KindStruct,
		ObjectID: resp.ObjectID,
	}
	root.Children = decodeValueChildren(cfg, resp.Value)
	root.Length = intPtr(len(root.Children))
	return root, nil
}

// decodeObjectValue unmarshals a qobject_json payload to its gno.Value.
// Shared by the preview-pass shape probes so one fetched payload is
// parsed once, not once per probe.
func decodeObjectValue(raw []byte) (gno.Value, error) {
	var resp objectResponse
	if err := amino.UnmarshalJSON(raw, &resp); err != nil {
		return nil, fmt.Errorf("decode object JSON: %w", err)
	}
	return resp.Value, nil
}

// decodeHeapItemInner unwraps a heap-item box — Gno's transparent wrapper
// around closure-captured / escaping variables — so callers can promote
// the inner value over the redundant "<heapItemType>" layer.
func decodeHeapItemInner(v gno.Value, cfg RenderConfig) (StateNode, bool) {
	hiv, ok := v.(*gno.HeapItemValue)
	if !ok {
		return StateNode{}, false
	}
	return decodeTypedValueAt(startDepthFor(cfg), "value", hiv.Value), true
}

// decodeFuncKind classifies a func object: only a fetched *FuncValue
// carries .Captures — the package-level FuncType the walker sees cannot.
func decodeFuncKind(v gno.Value) (string, bool) {
	fv, ok := v.(*gno.FuncValue)
	if !ok {
		return "", false
	}
	return funcKind(fv), true
}

// DecodePackage decodes a qpkg_json payload into the top-level slots,
// bounded by cfg. Previews are NOT resolved here — refs stay as
// expandable ref nodes for ResolvePreviews to follow.
func DecodePackage(ctx context.Context, raw []byte, cfg RenderConfig) ([]StateNode, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	cfg = clampRenderConfig(cfg)

	var resp pkgResponse
	if err := amino.UnmarshalJSON(raw, &resp); err != nil {
		return nil, fmt.Errorf("decode pkg JSON: %w", err)
	}
	startDepth := startDepthFor(cfg)
	nodes := make([]StateNode, 0, len(resp.Names))
	for i, name := range resp.Names {
		if i >= len(resp.Values) {
			break
		}
		nodes = append(nodes, decodeTypedValueAt(startDepth, name, resp.Values[i]))
	}
	return nodes, nil
}

// clampRenderConfig defaults zero/negative fields to safe values so a
// caller passing RenderConfig{} still gets a working bounded decode.
func clampRenderConfig(cfg RenderConfig) RenderConfig {
	if cfg.MaxChildrenPerNode <= 0 {
		cfg.MaxChildrenPerNode = maxChildrenPerNode
	}
	if cfg.MaxDecodeDepth <= 0 {
		cfg.MaxDecodeDepth = 3
	}
	if cfg.MaxDecodeDepth > maxDecodeDepth {
		cfg.MaxDecodeDepth = maxDecodeDepth
	}
	return cfg
}
