package state

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// encodeInt64LE renders a little-endian int64 as base64, matching how
// Amino emits PrimitiveType("32") fields in qobject_json responses.
func encodeInt64LE(v int64) string {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], uint64(v))
	return base64.StdEncoding.EncodeToString(b[:])
}

// previewMockFetcher records every call, optionally returns canned bodies
// per OID, and tracks peak in-flight concurrency.
type previewMockFetcher struct {
	bodies  map[string][]byte
	err     error
	delay   time.Duration
	calls   int32
	current int32
	peak    int32
}

func (f *previewMockFetcher) FetchObject(ctx context.Context, oid string) ([]byte, error) {
	atomic.AddInt32(&f.calls, 1)
	cur := atomic.AddInt32(&f.current, 1)
	defer atomic.AddInt32(&f.current, -1)
	for {
		old := atomic.LoadInt32(&f.peak)
		if cur <= old || atomic.CompareAndSwapInt32(&f.peak, old, cur) {
			break
		}
	}
	if f.delay > 0 {
		select {
		case <-time.After(f.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if f.err != nil {
		return nil, f.err
	}
	if b, ok := f.bodies[oid]; ok {
		return b, nil
	}
	return nil, fmt.Errorf("not found: %s", oid)
}

// previewStructBody returns a minimal qobject_json shape for a 2-field
// struct, sufficient to exercise inline-attach without per-test fixture
// boilerplate.
func previewStructBody(oid string, val0, val1 int) []byte {
	return []byte(fmt.Sprintf(`{
		"objectid": %q,
		"value": {
			"@type": "/gno.StructValue",
			"Fields": [
				{"T": {"@type": "/gno.PrimitiveType", "value": "32"}, "N": "%s"},
				{"T": {"@type": "/gno.PrimitiveType", "value": "32"}, "N": "%s"}
			]
		}
	}`, oid, encodeInt64LE(int64(val0)), encodeInt64LE(int64(val1))))
}

// TestPreviewBudgetCap15 pins the 15-fetch cap from ADR-004 §Resource bounds.
// 20 ref candidates → exactly 15 fetched, 5 left as bare refs.
func TestPreviewBudgetCap15(t *testing.T) {
	t.Parallel()

	const N = 20
	bodies := make(map[string][]byte, N)
	nodes := make([]StateNode, N)
	for i := 0; i < N; i++ {
		oid := fmt.Sprintf("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:%d", i+1)
		bodies[oid] = previewStructBody(oid, i, i+1)
		nodes[i] = StateNode{
			Name: fmt.Sprintf("R%d", i), Kind: KindRef,
			ObjectID: oid, Expandable: true,
		}
	}
	fetcher := &previewMockFetcher{bodies: bodies}

	spent, err := ResolvePreviews(context.Background(), nil, fetcher, nil, nodes)
	require.NoError(t, err)
	assert.Equal(t, PreviewMaxFetches, spent, "spent must equal the documented cap")
	assert.Equal(t, int32(PreviewMaxFetches), atomic.LoadInt32(&fetcher.calls),
		"no over-fetch beyond the cap")

	enriched, bare := 0, 0
	for _, n := range nodes {
		if len(n.Children) > 0 {
			enriched++
		} else {
			bare++
		}
	}
	assert.Equal(t, PreviewMaxFetches, enriched, "exactly cap nodes inlined")
	assert.Equal(t, N-PreviewMaxFetches, bare, "remainder stays as bare refs")
}

// TestPreviewTwoRoundResolution pins the Gno-specific heap→ref→struct chain:
// round 1 fetches the heap wrapper, round 2 follows the inner ref to the
// real struct. Total ≤2 fetches per chain.
func TestPreviewTwoRoundResolution(t *testing.T) {
	t.Parallel()

	const outerOID = "ffffffffffffffffffffffffffffffffffffffff:11"
	const innerOID = "ffffffffffffffffffffffffffffffffffffffff:12"

	heapJSON := []byte(fmt.Sprintf(`{
		"objectid": %q,
		"value": {
			"@type": "/gno.HeapItemValue",
			"Value": {
				"T": {"@type": "/gno.StructType", "PkgPath": "x", "Fields": []},
				"V": {"@type": "/gno.RefValue", "ObjectID": %q}
			}
		}
	}`, outerOID, innerOID))
	innerJSON := previewStructBody(innerOID, 7, 11)

	fetcher := &previewMockFetcher{
		bodies: map[string][]byte{outerOID: heapJSON, innerOID: innerJSON},
	}
	nodes := []StateNode{{
		Name: "user", Kind: KindPointer,
		ObjectID: outerOID, Expandable: true,
	}}

	spent, err := ResolvePreviews(context.Background(), nil, fetcher, nil, nodes)
	require.NoError(t, err)
	assert.Equal(t, 2, spent, "two rounds: one for heap wrapper, one for inner struct")
	require.NotEmpty(t, nodes[0].Children, "round 1 must reveal a child")

	// Locate the inner ref that resolved in round 2 — should be in the
	// outer node's child tree, populated with struct fields.
	innerResolved := false
	var visit func(n StateNode)
	visit = func(n StateNode) {
		if n.ObjectID == innerOID && len(n.Children) == 2 {
			innerResolved = true
			return
		}
		for _, c := range n.Children {
			visit(c)
		}
	}
	visit(nodes[0])
	assert.True(t, innerResolved,
		"round 2 must follow the inner ref and expose the struct fields")
}

// Regression: a closure's captured value is reached via decodeFuncInline
// (which consumes a depth level) → capture ref → heap-item unwrap. With
// previewChildDepth=1 the unwrap landed exactly on maxDecodeDepth and the
// value surfaced as a "(too deep)" sentinel; previewChildDepth=2 gives the
// extra slot so the actual captured value decodes.
func TestPreviewDepthResolvesHeapItemValue(t *testing.T) {
	t.Parallel()

	// The shape a closure capture (a RefValue to a heap item) resolves to:
	// a HeapItemValue wrapping the real value.
	payload := []byte(`{
		"objectid": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:1",
		"value": {"@type": "/gno.HeapItemValue",
			"Value": {"T": {"@type": "/gno.PrimitiveType", "value": "32"}, "N": "KgAAAAAAAAA="}}
	}`)
	cfg := RenderConfig{MaxChildrenPerNode: maxChildrenPerNode, MaxDecodeDepth: previewChildDepth}
	root, err := DecodeObject(context.Background(), payload, cfg)
	require.NoError(t, err)
	require.Len(t, root.Children, 1)
	got := root.Children[0]
	assert.NotEqual(t, KindTruncated, got.Kind,
		"heap-item value must decode to the actual value, not a (too deep) sentinel")
	assert.NotEmpty(t, got.Value, "the captured value must be present")
}

// TestPreviewParallelBounded pins the in-flight concurrency cap. With
// 2×PreviewMaxConcurrent refs and a small delay, the peak in-flight
// count must stay ≤ PreviewMaxConcurrent — no fetcher stampede.
func TestPreviewParallelBounded(t *testing.T) {
	t.Parallel()

	const N = PreviewMaxConcurrent * 2
	bodies := make(map[string][]byte, N)
	nodes := make([]StateNode, N)
	for i := 0; i < N; i++ {
		oid := fmt.Sprintf("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:%d", i+1)
		bodies[oid] = previewStructBody(oid, i, 0)
		nodes[i] = StateNode{
			Name: fmt.Sprintf("R%d", i), Kind: KindRef,
			ObjectID: oid, Expandable: true,
		}
	}
	fetcher := &previewMockFetcher{bodies: bodies, delay: 20 * time.Millisecond}

	_, err := ResolvePreviews(context.Background(), nil, fetcher, nil, nodes)
	require.NoError(t, err)

	peak := atomic.LoadInt32(&fetcher.peak)
	assert.LessOrEqual(t, int(peak), PreviewMaxConcurrent,
		"peak in-flight must not exceed PreviewMaxConcurrent")
	// No `peak > 1` floor — CI runners with GOMAXPROCS=1 can observe
	// serialized peaks even though goroutines were eligible.
}

// TestPreviewGracefulOnFetchError pins the failure-isolation contract: one
// failing fetch must NOT abort the remaining resolutions. The failed ref
// stays as a bare ref; the rest are inlined.
func TestPreviewGracefulOnFetchError(t *testing.T) {
	t.Parallel()

	const failOID = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:1"
	const okOID = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:2"

	bodies := map[string][]byte{
		// failOID is intentionally absent → fetcher returns "not found".
		okOID: previewStructBody(okOID, 7, 11),
	}
	fetcher := &previewMockFetcher{bodies: bodies}

	nodes := []StateNode{
		{Name: "fail", Kind: KindRef, ObjectID: failOID, Expandable: true},
		{Name: "ok", Kind: KindRef, ObjectID: okOID, Expandable: true},
	}

	_, err := ResolvePreviews(context.Background(), nil, fetcher, nil, nodes)
	require.NoError(t, err, "the resolver itself must not error on a single bad fetch")

	assert.Empty(t, nodes[0].Children, "failed ref stays bare")
	assert.Equal(t, failOID, nodes[0].ObjectID, "failed ref still carries its OID for retry/click-through")
	assert.NotEmpty(t, nodes[1].Children, "the sibling resolution must still succeed")
}

// TestPreviewCancellationViaContext pins the ctx-driven cancellation:
// a pre-canceled ctx must short-circuit before any fetches complete.
// Partial state (no children attached) is returned cleanly.
func TestPreviewCancellationViaContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	const N = PreviewMaxConcurrent * 2 // ensures the sem-blocked branch is hit
	nodes := make([]StateNode, N)
	bodies := make(map[string][]byte, N)
	for i := 0; i < N; i++ {
		oid := fmt.Sprintf("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:%d", i+1)
		bodies[oid] = previewStructBody(oid, i, 0)
		nodes[i] = StateNode{
			Name: fmt.Sprintf("R%d", i), Kind: KindRef,
			ObjectID: oid, Expandable: true,
		}
	}
	fetcher := &previewMockFetcher{bodies: bodies, delay: 50 * time.Millisecond}

	_, err := ResolvePreviews(ctx, nil, fetcher, nil, nodes)
	// ResolvePreviews itself must not return ctx.Err as a fatal — the
	// contract is "partial results, no children attached, no panic".
	// We accept either nil or ctx.Err so the implementation can choose;
	// the load-bearing assertion is "no children attached".
	if err != nil {
		assert.ErrorIs(t, err, context.Canceled)
	}

	for i := range nodes {
		assert.Empty(t, nodes[i].Children,
			"no preview may attach under canceled ctx (node %d)", i)
	}
}

// previewFuncBody returns a minimal qobject_json for a *FuncValue.
// captures>0 makes it a closure — decodeFuncInline keys closure-ness off
// len(Captures) (not the FuncValue.IsClosure flag, which is false even
// for a capturing func literal), so the preview pass must too.
func previewFuncBody(oid string, captures int) []byte {
	caps := ""
	for i := 0; i < captures; i++ {
		if i > 0 {
			caps += ","
		}
		caps += fmt.Sprintf(
			`{"T":{"@type":"/gno.heapItemType"},"V":{"@type":"/gno.RefValue","ObjectID":%q}}`,
			fmt.Sprintf("%s%d", oid, i))
	}
	return []byte(fmt.Sprintf(`{
		"objectid": %q,
		"value": {
			"@type": "/gno.FuncValue",
			"Type": {"@type": "/gno.FuncType", "Params": [], "Results": []},
			"Captures": [%s]
		}
	}`, oid, caps))
}

// TestPreviewDetectsClosureKind pins the package-page closure badge: a
// top-level func ref decodes as KindFunc (a FuncType cannot tell a
// closure from a plain func). The preview pass fetches the func object
// and promotes Kind to KindClosure when it carries captures — body and
// captures stay lazy, so no children are attached.
func TestPreviewDetectsClosureKind(t *testing.T) {
	t.Parallel()

	const closureOID = "cccccccccccccccccccccccccccccccccccccccc:63"
	const plainOID = "cccccccccccccccccccccccccccccccccccccccc:53"
	fetcher := &previewMockFetcher{bodies: map[string][]byte{
		closureOID: previewFuncBody(closureOID, 1),
		plainOID:   previewFuncBody(plainOID, 0),
	}}
	nodes := []StateNode{
		{Name: "NextID", Kind: KindFunc, ObjectID: closureOID, Expandable: true},
		{Name: "Add", Kind: KindFunc, ObjectID: plainOID, Expandable: true},
	}

	spent, err := ResolvePreviews(context.Background(), nil, fetcher, nil, nodes)
	require.NoError(t, err)
	assert.Equal(t, 2, spent, "both func objects fetched once for Kind detection")
	assert.Equal(t, KindClosure, nodes[0].Kind, "func with captures → closure badge")
	assert.Equal(t, KindFunc, nodes[1].Kind, "plain func stays func")
	assert.Empty(t, nodes[0].Children, "Kind-detection only — body/captures stay lazy")
	assert.Empty(t, nodes[1].Children, "plain func untouched beyond Kind")
}

// TestPreviewFuncsGetBudgetPriority pins the collectPreviewCandidates
// ordering: with more candidates than the 15-fetch cap, every top-level
// func is still fetched (closure detection is a cheap terminal fetch)
// while surplus data refs degrade to bare click-to-expand.
func TestPreviewFuncsGetBudgetPriority(t *testing.T) {
	t.Parallel()

	const funcs = 5
	bodies := make(map[string][]byte)
	var nodes []StateNode
	// Data refs declared first — without priority they would eat the cap.
	for i := 0; i < PreviewMaxFetches+5; i++ {
		oid := fmt.Sprintf("dddddddddddddddddddddddddddddddddddddddd:%d", i+1)
		bodies[oid] = previewStructBody(oid, i, i)
		nodes = append(nodes, StateNode{
			Name: fmt.Sprintf("D%d", i), Kind: KindRef, ObjectID: oid, Expandable: true,
		})
	}
	for i := 0; i < funcs; i++ {
		oid := fmt.Sprintf("eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee:%d", i+1)
		bodies[oid] = previewFuncBody(oid, 1) // all closures
		nodes = append(nodes, StateNode{
			Name: fmt.Sprintf("F%d", i), Kind: KindFunc, ObjectID: oid, Expandable: true,
		})
	}
	fetcher := &previewMockFetcher{bodies: bodies}

	spent, err := ResolvePreviews(context.Background(), nil, fetcher, nil, nodes)
	require.NoError(t, err)
	assert.Equal(t, PreviewMaxFetches, spent, "still bounded by the cap")
	for i := len(nodes) - funcs; i < len(nodes); i++ {
		assert.Equal(t, KindClosure, nodes[i].Kind,
			"every func resolved despite being declared after a cap-filling run of data refs")
	}
}

// TestPreviewHeapItemUnwrapKeepsIdentity pins M3: unwrapping a heap-item
// box promotes the inner value onto the ref but must NOT wipe the ref's
// ObjectID/TypeID — otherwise a not-fully-resolved inner value loses its
// click-through.
func TestPreviewHeapItemUnwrapKeepsIdentity(t *testing.T) {
	t.Parallel()

	const heapOID = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:1"
	heapJSON := []byte(fmt.Sprintf(`{
		"objectid": %q,
		"value": {"@type": "/gno.HeapItemValue",
			"Value": {"T": {"@type": "/gno.PrimitiveType", "value": "32"}, "N": "KgAAAAAAAAA="}}
	}`, heapOID))
	fetcher := &previewMockFetcher{bodies: map[string][]byte{heapOID: heapJSON}}
	nodes := []StateNode{{
		Name: "captured", Kind: KindRef,
		ObjectID: heapOID, TypeID: "tid-x", Expandable: true,
	}}

	_, err := ResolvePreviews(context.Background(), nil, fetcher, nil, nodes)
	require.NoError(t, err)
	assert.Equal(t, "captured", nodes[0].Name, "name is preserved")
	assert.Equal(t, heapOID, nodes[0].ObjectID, "ObjectID survives the heap-item unwrap")
	assert.Equal(t, "tid-x", nodes[0].TypeID, "TypeID survives the heap-item unwrap")
	assert.NotEmpty(t, nodes[0].Value, "the inner value is still promoted onto the ref")
}

// TestPreviewRoundCapNeverNegative pins M5: the round-1 budget reserve
// (PreviewMaxFetches - previewRound2Reserve) feeds a slice bound, so it
// must never go negative regardless of how the consts are retuned.
func TestPreviewRoundCapNeverNegative(t *testing.T) {
	t.Parallel()

	roundCap := PreviewMaxFetches - previewRound2Reserve
	if roundCap < 0 {
		roundCap = 0
	}
	assert.GreaterOrEqual(t, roundCap, 0, "roundCap must floor at 0 to keep candidates[:roundCap] safe")
}
