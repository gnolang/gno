// Package run implements the gnoweb run feature.
//
// It serves the maketx-run scratchpad at the $run web query of an existing
// realm or package (e.g. /r/demo/boards$run). The page is served whole: a
// starter script that runs against the package, and the `gnokey maketx
// run` command in $help's two modes. The browser makes the script editable
// and keeps the command in step with the fields.
package run
