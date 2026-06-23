package gnoweb

import (
	"mime"
	"strconv"
	"strings"
)

// negotiatesMarkdown reports whether the client explicitly accepts markdown.
//
// It returns true only when "text/markdown" (or the alias "text/x-markdown")
// is named explicitly in the Accept header and not refused with an explicit
// q=0. It never matches the "*/*" or "text/*" wildcards, so browsers — which
// always accept "*/*" — continue to receive HTML.
func negotiatesMarkdown(accept string) bool {
	for _, part := range strings.Split(accept, ",") {
		mediaType, params, err := mime.ParseMediaType(strings.TrimSpace(part))
		if err != nil {
			continue
		}
		if mediaType != "text/markdown" && mediaType != "text/x-markdown" {
			continue
		}
		if q, ok := params["q"]; ok {
			if v, err := strconv.ParseFloat(q, 64); err == nil && v <= 0 {
				continue
			}
		}
		return true
	}
	return false
}
