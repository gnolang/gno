package main

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// remoteArr is a flag.Value adapter that parses repeated
// `-remote <domain>=<rpc>` arguments into the Config.Remotes map consumed
// by the Loader.
type remoteArr map[string]string

// String returns a deterministic, comma-joined representation of the parsed
// entries (keys sorted lexicographically). Stable output matters for the
// flag package's default-printing and for any caller logging the value.
func (m *remoteArr) String() string {
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

func (m *remoteArr) Set(v string) error {
	domain, rpc, ok := strings.Cut(v, "=")
	if !ok {
		return fmt.Errorf("invalid -remote %q: expected domain=rpc", v)
	}
	if domain == "" {
		return fmt.Errorf("invalid -remote %q: empty domain", v)
	}
	if rpc == "" {
		return fmt.Errorf("invalid -remote %q: empty rpc", v)
	}
	if *m == nil {
		*m = remoteArr{}
	}
	(*m)[domain] = rpc
	return nil
}
