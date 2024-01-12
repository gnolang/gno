package static

import "embed"

// EmbeddedStatic holds static web server content.
//
//go:embed *
var EmbeddedStatic embed.FS
