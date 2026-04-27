package main

import (
	"fmt"
	"strings"
)

// remoteOverrideArr is a flag.Value adapter that parses repeated
// `-remote-override <domain>=<rpc>` arguments into a string map. It mirrors
// the rpcpkgfetcher.New(map) signature used by Loader.
type remoteOverrideArr map[string]string

func (m *remoteOverrideArr) String() string {
	if m == nil || len(*m) == 0 {
		return ""
	}
	parts := make([]string, 0, len(*m))
	for k, v := range *m {
		parts = append(parts, k+"="+v)
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
