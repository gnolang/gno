package gnoland

import (
	"errors"
	"fmt"
	"strings"
)

// allowPathsEntry is one parsed AllowPaths entry. Either:
//   - a wildcard "*" (Wildcard is true; matches any msg that is not in the
//     always-denied list), or
//   - a (route, type) pair, optionally constrained to a realm path (only
//     legal for "vm/exec").
type allowPathsEntry struct {
	Wildcard bool   // true for the bare "*" entry
	Route    string // module name (e.g. "vm", "bank"); empty for wildcard
	Type     string // msg type (e.g. "exec", "send"); empty for wildcard
	Path     string // optional realm path; non-empty only for vm/exec entries
}

// allowPathsWildcard matches any msg type (subject to the always-denied list
// applied at ante-time). See ADR-001.
const allowPathsWildcard = "*"

// validSessionRouteTypes is the whitelist of <route>/<type> pairs permitted
// in a session's AllowPaths. Bare-route entries ("bank", "vm") and unknown
// types are rejected; relaxing this is a future-fork concern.
var validSessionRouteTypes = map[string]struct{}{
	"vm/exec":        {},
	"vm/run":         {},
	"bank/send":      {},
	"bank/multisend": {},
}

// pathBearingRouteType is the only entry that may carry a ":<path>" suffix.
const pathBearingRouteType = "vm/exec"

// parseAllowPathsEntry parses one entry into structured form. Splits on the
// first colon so realm paths containing ':' survive intact. The bare "*"
// token is a wildcard that matches any msg type (subject to the always-denied
// list); "*:<path>" is rejected.
func parseAllowPathsEntry(s string) (allowPathsEntry, error) {
	if s == "" {
		return allowPathsEntry{}, errors.New("empty allow-paths entry")
	}
	if s == allowPathsWildcard {
		return allowPathsEntry{Wildcard: true}, nil
	}
	routeType, path, hasPath := strings.Cut(s, ":")
	if routeType == allowPathsWildcard {
		return allowPathsEntry{}, errors.New("wildcard '*' must not have a path suffix")
	}
	if _, ok := validSessionRouteTypes[routeType]; !ok {
		return allowPathsEntry{}, fmt.Errorf(
			"unknown route_type %q (want one of: *, vm/exec, vm/run, bank/send, bank/multisend)",
			routeType,
		)
	}

	slash := strings.IndexByte(routeType, '/')
	e := allowPathsEntry{
		Route: routeType[:slash],
		Type:  routeType[slash+1:],
	}

	if !hasPath {
		return e, nil
	}
	if routeType != pathBearingRouteType {
		return allowPathsEntry{}, fmt.Errorf(
			"only vm/exec accepts a path suffix; %q does not", routeType,
		)
	}
	if path == "" {
		return allowPathsEntry{}, errors.New("vm/exec entry requires a non-empty path after ':'")
	}
	if strings.HasSuffix(path, "/") {
		return allowPathsEntry{}, fmt.Errorf("path %q has a trailing slash", path)
	}
	e.Path = path
	return e, nil
}

// parseAllowPaths parses the full AllowPaths slice into structured entries.
// Returns an error if the slice is empty: AllowPaths is required at create
// time (use ["*"] for unrestricted). Per-entry errors are wrapped with the
// offending index for clearer feedback.
func parseAllowPaths(paths []string) ([]allowPathsEntry, error) {
	if len(paths) == 0 {
		return nil, errors.New("AllowPaths is required (use [\"*\"] for unrestricted)")
	}
	out := make([]allowPathsEntry, 0, len(paths))
	for i, s := range paths {
		e, err := parseAllowPathsEntry(s)
		if err != nil {
			return nil, fmt.Errorf("allow_paths[%d]: %w", i, err)
		}
		out = append(out, e)
	}
	return out, nil
}
