// Package run holds the gnoweb maketx-run scratchpad.
//
// The scratchpad is a section of the Actions page (?help), not a page of its
// own: it renders an editable code template and lets the user copy the
// resulting `gnokey maketx run` command. The Dry Run button posts the script
// to /_/api/dryrun, which is served by the playground feature.
//
// Unlike the other features this package ships no handler — only the
// "ui/run_script" template (parsed by components) and the frontend assets
// (controller-run.ts, run.css) that the build picks up by convention.
package run
