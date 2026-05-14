package state

import (
	"context"
	"log/slog"
	"sync"
)

// Preview bounds per ADR-004 §Resource bounds.
const (
	// PreviewMaxFetches caps object RPCs across all rounds. Type RPCs are
	// extra but TID-deduped (same as legacy EnrichInlinePreviews), so a
	// resolve costs ≤ PreviewMaxFetches objects + ≤ uniqueTIDs types.
	PreviewMaxFetches = 15
	// PreviewMaxRounds peels Gno's nested indirection. 3 rounds because a
	// closure capture is a 3-hop chain: top-level closure-ref → captured
	// heap-item ref → value. The plain heap→ref→struct (*T) pattern only
	// needs 2 — the 3rd round just finds no candidates and returns. Total
	// fetches still capped at PreviewMaxFetches; more rounds peel deeper,
	// they don't add budget.
	PreviewMaxRounds = 3
	// previewRound2Reserve holds back part of PreviewMaxFetches so the
	// deeper rounds (the heap→ref→struct / closure-capture peel) aren't
	// starved when round 1 already has ≥PreviewMaxFetches top-level
	// candidates. Total stays ≤PreviewMaxFetches — re-allocates, not adds.
	previewRound2Reserve = 5
	// PreviewMaxConcurrent is the in-flight fetcher pool size.
	PreviewMaxConcurrent = 8
	// previewChildDepth bounds the untyped-fallback decode of one fetched
	// object's payload. 2, not 1: decodeFuncInline consumes a depth level
	// before walking captures, so a closure's captured values need the
	// extra slot — at depth 1 they decode straight into the depth ceiling
	// and surface as opaque "<heapItemType>" / "(too deep)" nodes.
	// Purely a per-payload walk bound; the RPC count stays capped by
	// PreviewMaxFetches.
	previewChildDepth = 2
)

// ObjectFetcher fetches one object payload by OID. Optionally also
// fetches a type payload by TID — without it, preview struct fields
// degrade to positional indices instead of named members.
type ObjectFetcher interface {
	FetchObject(ctx context.Context, oid string) ([]byte, error)
}

// TypeFetcher pairs with ObjectFetcher for the typed-preview path.
// Separate interface so callers that don't need type lookups (rare)
// can still satisfy ObjectFetcher alone.
type TypeFetcher interface {
	FetchType(ctx context.Context, tid string) ([]byte, error)
}

// ResolvePreviews fetches up to PreviewMaxFetches refs in BFS order
// across PreviewMaxRounds rounds, inlining the resolved content into
// the input tree in place. Bounded by ctx cancellation.
//
// typeFetcher is optional: when non-nil, type payloads are fetched
// alongside objects so struct previews carry named fields instead of
// positional indices. A nil typeFetcher silently downgrades to the
// positional path.
//
// Failure isolation: each fetch is independent; a failed fetch leaves
// its ref bare (Children empty, ObjectID preserved so the user can
// click through) and does NOT abort the rest of the resolve.
//
// Returns the count of RPCs actually spent (0 ≤ n ≤ PreviewMaxFetches).
// The error return is reserved for systemic failures (currently never
// produced — individual fetch errors are absorbed). nil fetcher is a
// no-op that returns (0, nil).
func ResolvePreviews(ctx context.Context, logger *slog.Logger, fetcher ObjectFetcher, typeFetcher TypeFetcher, nodes []StateNode) (int, error) {
	if fetcher == nil || len(nodes) == 0 {
		return 0, nil
	}
	if logger == nil {
		logger = slog.Default()
	}

	// Cross-round dedupe: an OID resolved in round 1 must not be re-
	// fetched in round 2 even if a cycle re-exposes it as a bare ref.
	fetched := make(map[string]struct{})
	spent := 0

	for round := 0; round < PreviewMaxRounds; round++ {
		if err := ctx.Err(); err != nil {
			return spent, nil
		}
		remaining := PreviewMaxFetches - spent
		if remaining <= 0 {
			return spent, nil
		}

		var candidates []*StateNode
		collectPreviewCandidates(nodes, &candidates)
		filtered := candidates[:0]
		for _, n := range candidates {
			if _, seen := fetched[n.ObjectID]; seen {
				continue
			}
			filtered = append(filtered, n)
		}
		candidates = filtered
		if len(candidates) == 0 {
			return spent, nil
		}

		// Round 1 holds back previewRound2Reserve so the round-2 peel
		// (heap→ref→struct, closure captures) gets guaranteed budget even
		// when the package has ≥PreviewMaxFetches top-level refs. Later
		// rounds may use whatever is left.
		roundCap := remaining
		if round == 0 && PreviewMaxRounds > 1 {
			if r := PreviewMaxFetches - previewRound2Reserve; r < roundCap {
				roundCap = r
			}
		}
		if roundCap < 0 {
			roundCap = 0
		}
		if len(candidates) > roundCap {
			candidates = candidates[:roundCap]
		}

		// Budget consumes attempts, not just successes — a flood of
		// failing OIDs in round 1 must not let round 2 retry past the cap.
		spent += fetchPreviewRound(ctx, logger, candidates, fetcher, typeFetcher, fetched)
	}
	return spent, nil
}

// fetchPreviewRound issues fetches for the given candidates under a
// shared concurrency cap, decodes each successful payload, and inlines
// the result into the candidate nodes. Returns the number of unique
// OIDs attempted (always = len(unique candidate OIDs); each attempt
// counts toward the cap whether it succeeded or failed).
//
// Typed-preview path: when typeFetcher is non-nil and a candidate
// carries a TypeID, the type payload is fetched in parallel so the
// resolved struct's field names replace positional indices. Type
// fetches share the same concurrency pool but DO NOT count against
// PreviewMaxFetches — they're a small follow-up RPC on already-bounded
// objects, not a separate amplification surface.
func fetchPreviewRound(ctx context.Context, logger *slog.Logger, candidates []*StateNode, fetcher ObjectFetcher, typeFetcher TypeFetcher, fetched map[string]struct{}) int {
	byOID := make(map[string][]*StateNode)
	tidByOID := make(map[string]string)
	for _, n := range candidates {
		byOID[n.ObjectID] = append(byOID[n.ObjectID], n)
		if tidByOID[n.ObjectID] == "" && n.TypeID != "" {
			tidByOID[n.ObjectID] = n.TypeID
		}
	}

	type result struct {
		raw []byte
		err error
	}
	results := make(map[string]result, len(byOID))
	typeResults := make(map[string][]byte) // tid → raw (success only)
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, PreviewMaxConcurrent)

	for oid := range byOID {
		wg.Add(1)
		go func(oid string) {
			defer wg.Done()
			defer recoverFetcher(logger, "preview", "oid", oid)
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return
			}
			defer func() { <-sem }()
			raw, err := fetcher.FetchObject(ctx, oid)
			mu.Lock()
			results[oid] = result{raw: raw, err: err}
			mu.Unlock()
		}(oid)
	}

	// Dedupe TIDs so two refs sharing a struct type only cost one qtype RPC.
	if typeFetcher != nil {
		uniqTIDs := make(map[string]struct{})
		for _, tid := range tidByOID {
			if tid != "" {
				uniqTIDs[tid] = struct{}{}
			}
		}
		for tid := range uniqTIDs {
			wg.Add(1)
			go func(tid string) {
				defer wg.Done()
				defer recoverFetcher(logger, "preview-type", "tid", tid)
				select {
				case sem <- struct{}{}:
				case <-ctx.Done():
					return
				}
				defer func() { <-sem }()
				raw, err := typeFetcher.FetchType(ctx, tid)
				if err != nil || len(raw) == 0 {
					if err != nil {
						logger.Debug("preview type fetch failed", "tid", tid, "err", err)
					}
					return
				}
				mu.Lock()
				typeResults[tid] = raw
				mu.Unlock()
			}(tid)
		}
	}
	wg.Wait()

	// Untyped fallback uses fragment-mode cfg (previewChildDepth bound).
	cfg := RenderConfig{
		MaxChildrenPerNode: maxChildrenPerNode,
		MaxDecodeDepth:     previewChildDepth,
	}

	for oid, refs := range byOID {
		// Every attempted OID counts toward the budget — record it as
		// fetched so round 2 won't retry it.
		fetched[oid] = struct{}{}
		res, ok := results[oid]
		if !ok || res.err != nil || len(res.raw) == 0 {
			if ok && res.err != nil {
				logger.Debug("preview fetch failed", "oid", oid, "err", res.err)
			}
			continue
		}

		// One unmarshal feeds both shape probes below. The struct path
		// still re-decodes via DecodeObject* — its callers need the raw
		// bytes + ObjectID — but that's the pre-existing cost, not new.
		val, derr := decodeObjectValue(res.raw)
		if derr != nil {
			logger.Debug("preview decode failed", "oid", oid, "err", derr)
			continue
		}

		// Heap-item box (closure-captured / escaping variable): unwrap it
		// and promote the inner value onto the ref in place — so a capture
		// renders as `name = value`, not `name : <heapItemType>` → expand.
		// Keep the ref's identity (Name + OID/TID) so a not-fully-resolved
		// inner value still clicks through.
		if inner, isHeap := decodeHeapItemInner(val, cfg); isHeap {
			for _, n := range refs {
				name, oid, tid := n.Name, n.ObjectID, n.TypeID
				*n = inner
				n.Name = name
				if n.ObjectID == "" {
					n.ObjectID = oid
				}
				if n.TypeID == "" {
					n.TypeID = tid
				}
			}
			continue
		}

		// Func/closure object: set Kind only — it drives the closure
		// badge. Body + captures stay lazy via serveFragNode; decoding
		// children here would nest a redundant "(function)" wrapper.
		if kind, isFunc := decodeFuncKind(val); isFunc {
			for _, n := range refs {
				n.Kind = kind
			}
			continue
		}

		var (
			children []StateNode
			info     StateObjectInfoView
		)
		if tid := tidByOID[oid]; tid != "" {
			if rawType := typeResults[tid]; len(rawType) > 0 {
				decoded, err := DecodeObjectFull(res.raw, rawType, cfg)
				if err == nil {
					children = decoded.Nodes
					info = decoded.Info
				} else {
					logger.Debug("preview typed decode failed; falling back to positional", "oid", oid, "tid", tid, "err", err)
				}
			}
		}
		if children == nil {
			root, err := DecodeObject(ctx, res.raw, cfg)
			if err != nil {
				logger.Debug("preview decode failed", "oid", oid, "err", err)
				continue
			}
			children = root.Children
		}
		for _, n := range refs {
			n.Children = children
			n.Preview = buildChildrenPreview(children)
			// Back-fill audit metadata from the typed-decode result so the
			// pretty-view stats block (OID/Size/Refs/Owner/Modified/Hash)
			// populates on inline-resolved refs. Only ModTime/RefCount/
			// LastObjectSize need copying — Hash/OwnerID/ObjectID came
			// straight from the ref node's parent walk.
			if info.Hash != "" && n.Hash == "" {
				n.Hash = info.Hash
			}
			if info.OwnerID != "" && n.OwnerID == "" {
				n.OwnerID = info.OwnerID
			}
			if info.ModTime != "" && n.ModTime == "" {
				n.ModTime = info.ModTime
			}
			if info.RefCount != "" && n.RefCount == "" {
				n.RefCount = info.RefCount
			}
			if info.LastObjectSize != "" && n.LastObjectSize == "" {
				n.LastObjectSize = info.LastObjectSize
			}
		}
	}
	return len(byOID)
}
