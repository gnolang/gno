// Package debug provides utilities for logging and displaying debug information
// within Gno realms. It supports conditional rendering of logs and metadata,
// toggleable via query parameters.
//
// Key Features:
// - Log collection and display using Markdown formatting.
// - Metadata display for realm path, address, and height.
// - Collapsible debug section for cleaner presentation.
// - Query-based debug toggle using `?debug=1`.
package debug

import (
	"chain/runtime"
	"time"

	"gno.land/p/moul/md"
	"gno.land/p/moul/mdtable"
	"gno.land/p/moul/realmpath"
	"gno.land/p/nt/ufmt"
)

// Debug encapsulates debug information, including logs and metadata.
type Debug struct {
	Logs         []string
	HideMetadata bool
}

// Log appends a new line of debug information to the Logs slice.
func (d *Debug) Log(line string) {
	d.Logs = append(d.Logs, line)
}

// Render generates the debug content as a collapsible Markdown section.
// It conditionally renders logs and metadata if enabled via the `?debug=1` query parameter.
func (d Debug) Render(path string) string {
	if realmpath.Parse(path).Query.Get("debug") != "1" {
		return ""
	}

	var content string

	if d.Logs != nil {
		content += md.H3("Logs")
		content += md.BulletList(d.Logs)
	}

	if !d.HideMetadata {
		content += md.H3("Metadata")
		table := mdtable.Table{
			Headers: []string{"Key", "Value"},
		}
		table.Append([]string{"`std.CurrentRealm().PkgPath()`", string(runtime.CurrentRealm().PkgPath())})
		table.Append([]string{"`std.CurrentRealm().Address()`", string(runtime.CurrentRealm().Address())})
		table.Append([]string{"`std.PreviousRealm().PkgPath()`", string(runtime.PreviousRealm().PkgPath())})
		table.Append([]string{"`std.PreviousRealm().Address()`", string(runtime.PreviousRealm().Address())})
		table.Append([]string{"`std.ChainHeight()`", ufmt.Sprintf("%d", runtime.ChainHeight())})
		table.Append([]string{"`time.Now().Format(time.RFC3339)`", time.Now().Format(time.RFC3339)})
		content += table.String()
	}

	if content == "" {
		return ""
	}

	return md.CollapsibleSection("debug", content)
}

// Render displays metadata about the current realm but does not display logs.
// This function uses a default Debug struct with metadata enabled and no logs.
func Render(path string) string {
	return Debug{}.Render(path)
}

// IsEnabled checks if the `?debug=1` query parameter is set in the given path.
// Returns true if debugging is enabled, otherwise false.
func IsEnabled(path string) bool {
	req := realmpath.Parse(path)
	return req.Query.Get("debug") == "1"
}

// ToggleURL modifies the given path's query string to toggle the `?debug=1` parameter.
// If debugging is currently enabled, it removes the parameter.
// If debugging is disabled, it adds the parameter.
func ToggleURL(path string) string {
	req := realmpath.Parse(path)
	if IsEnabled(path) {
		req.Query.Del("debug")
	} else {
		req.Query.Add("debug", "1")
	}
	return req.String()
}
