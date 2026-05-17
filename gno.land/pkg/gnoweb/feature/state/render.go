package state

import (
	"context"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/amino"
)

// RenderConfig bounds one Amino decode. Fragments use a shallow depth (≤3);
// the full-page path keeps walker.go's 256 for legacy parity.
type RenderConfig struct {
	// MaxChildrenPerNode caps visible children; surplus collapses to one
	// KindTruncated sentinel.
	MaxChildrenPerNode int
	// MaxDecodeDepth bounds recursion depth for this single decode.
	MaxDecodeDepth int
}

// DefaultFragmentRenderConfig is the slim per-fragment budget (depth ≤3).
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
// Refs surface as KindRef; ExportRefValue cycle markers as KindCycle.
// recoverDecodeToErr keeps amino's hard panics on hostile chain bytes
// inside the function — the caller gets an error, never a torn request.
func DecodeObject(ctx context.Context, raw []byte, cfg RenderConfig) (root StateNode, err error) {
	defer recoverDecodeToErr("decode object JSON", &err)
	if err := ctx.Err(); err != nil {
		return StateNode{}, err
	}
	cfg = clampRenderConfig(cfg)

	var resp objectResponse
	if err := amino.UnmarshalJSON(raw, &resp); err != nil {
		return StateNode{}, fmt.Errorf("decode object JSON: %w", err)
	}

	root = StateNode{
		Name:     "(object)",
		Kind:     KindStruct,
		ObjectID: resp.ObjectID,
	}
	root.Children = decodeValueChildren(cfg, resp.Value)
	root.Length = intPtr(len(root.Children))
	return root, nil
}

// DecodePackage decodes a paginated window over a qpkg_json payload's
// top-level slots, bounded by cfg. Returns (page, total, err); the caller
// builds the prev/next view-model from total via buildPagination. Splits
// into parsePackage + decodePackageSlice so the page handler can reuse
// the parsed Names+Values for the full sidebar TOC without re-decoding.
func DecodePackage(ctx context.Context, raw []byte, cfg RenderConfig, offset, limit int) ([]StateNode, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}
	resp, err := parsePackage(raw)
	if err != nil {
		return nil, 0, err
	}
	total := min(len(resp.Names), len(resp.Values))
	if limit <= 0 {
		limit = maxTopLevelDecls
	}
	start, end := clampSliceWindow(offset, limit, total)
	indices := make([]int, 0, end-start)
	for i := start; i < end; i++ {
		indices = append(indices, i)
	}
	nodes, err := decodePackageSlice(ctx, resp, cfg, indices)
	return nodes, total, err
}

// parsePackage is the amino-decode half of DecodePackage. Exposed so the
// page handler can compute full-sidebar metadata (peekTopLevelKind over
// every Value) from a single parse, instead of decoding the package twice.
// recoverDecodeToErr keeps an amino panic on hostile chain bytes inside
// the function; the caller sees a clean error and returns 500.
func parsePackage(raw []byte) (resp pkgResponse, err error) {
	defer recoverDecodeToErr("decode pkg JSON", &err)
	if err := amino.UnmarshalJSON(raw, &resp); err != nil {
		return pkgResponse{}, fmt.Errorf("decode pkg JSON: %w", err)
	}
	return resp, nil
}

// decodePackageSlice walks the selected top-level indices of an already-
// parsed pkgResponse, bounded by cfg. indices is consumed positionally so
// the caller can align anchors/kinds with the returned nodes slice.
// recoverDecodeToErr catches walker panics on hostile values so a single
// malformed top-level decl cannot tear the whole page response.
func decodePackageSlice(ctx context.Context, resp pkgResponse, cfg RenderConfig, indices []int) (nodes []StateNode, err error) {
	defer recoverDecodeToErr("decode pkg slice", &err)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	cfg = clampRenderConfig(cfg)
	startDepth := startDepthFor(cfg)
	total := min(len(resp.Names), len(resp.Values))
	nodes = make([]StateNode, 0, len(indices))
	for _, i := range indices {
		if i < 0 || i >= total {
			continue
		}
		nodes = append(nodes, decodeTypedValueAt(startDepth, resp.Names[i], resp.Values[i]))
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
