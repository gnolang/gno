package components

// StateViewType identifies the state-explorer body view so layout_index.go
// can branch on it (dev-mode chrome) without importing feature/state,
// which would create a cycle: feature/state already imports components.
//
// The canonical Kind* constants and OID helpers (TruncOID, ShortenOID)
// live in feature/state — components is no longer in that path.
const StateViewType ViewType = "state-view"
