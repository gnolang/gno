package main

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// remoteOverrideArr is a flag.Value adapter that parses repeated
// `-remote-override <domain>=<rpc>` arguments into a string map. It mirrors
// the rpcpkgfetcher.New(map) signature used by Loader.
type remoteOverrideArr map[string]string

// String returns a deterministic, comma-joined representation of the parsed
// overrides (keys sorted lexicographically). Stable output matters for the
// flag package's default-printing and for any caller logging the value.
func (m *remoteOverrideArr) String() string {
	if m == nil || len(*m) == 0 {
		return ""
	}
	keys := slices.Sorted(maps.Keys(*m))
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = k + "=" + (*m)[k]
	}
	return strings.Join(parts, ",")
}

func (m *remoteOverrideArr) Set(v string) error {
	domain, rpc, ok := strings.Cut(v, "=")
	if !ok {
		return fmt.Errorf("invalid -remote-override %q: expected domain=rpc", v)
	}
	if domain == "" {
		return fmt.Errorf("invalid -remote-override %q: empty domain", v)
	}
	if rpc == "" {
		return fmt.Errorf("invalid -remote-override %q: empty rpc", v)
	}
	if *m == nil {
		*m = remoteOverrideArr{}
	}
	(*m)[domain] = rpc
	return nil
}
