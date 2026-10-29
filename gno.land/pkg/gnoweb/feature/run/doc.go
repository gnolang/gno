// Package run implements the gnoweb run feature.
//
// It serves the maketx-run scratchpad at any package or realm URL
// carrying the ?run query (e.g. /r/demo/boards?run). The page is
// client-side: it renders an editable code template and lets the user
// copy the resulting `gnokey maketx run` command.
package run
