package gnoweb

import (
	"strconv"
	"strings"
)

// negotiatesMarkdown reports whether the client explicitly accepts markdown.
//
// It returns true only when "text/markdown" (or the alias "text/x-markdown")
// is named explicitly in the Accept header with a non-zero q-value. It never
// matches the "*/*" or "text/*" wildcards, so browsers — which always accept
// "*/*" — continue to receive HTML.
func negotiatesMarkdown(accept string) bool {
	for _, part := range strings.Split(accept, ",") {
		mediaType, params, _ := strings.Cut(strings.TrimSpace(part), ";")
		mediaType = strings.ToLower(strings.TrimSpace(mediaType))
		if mediaType != "text/markdown" && mediaType != "text/x-markdown" {
			continue
		}
		if markdownQualityIsZero(params) {
			continue
		}
		return true
	}
	return false
}

// markdownQualityIsZero reports whether the media-range parameters contain an
// explicit q-value of zero (an explicit refusal). An absent or unparseable
// q-value is treated as acceptable.
func markdownQualityIsZero(params string) bool {
	for _, p := range strings.Split(params, ";") {
		key, value, ok := strings.Cut(p, "=")
		if !ok || strings.ToLower(strings.TrimSpace(key)) != "q" {
			continue
		}
		q, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err != nil {
			return false // unparseable q -> treat as acceptable
		}
		return q <= 0
	}
	return false
}
