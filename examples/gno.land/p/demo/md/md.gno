package md

import (
	"strings"
)

// escape a URL for markdown.
func EscMDURL(s string) string {
	return escape(s)
}

func escape(s string) string {
	schars := map[string]string{
		"%":  "%25",
		" ":  "%20",
		"!":  "%21",
		"\"": "%22",
		"#":  "%23",
		"$":  "%24",
		"&":  "%26",
		"'":  "%27",
		"(":  "%28",
		")":  "%29",
		"*":  "%2A",
		"+":  "%2B",
		",":  "%2C",
		":":  "%3A",
		";":  "%3B",
		"<":  "%3C",
		"=":  "%3D",
		">":  "%3E",
		"?":  "%3F",
		"@":  "%40",
		"[":  "%5B",
		"\\": "%5C",
		"]":  "%5D",
		"^":  "%5E",
		"_":  "%5F",
		"`":  "%60",
		"{":  "%7B",
		"|":  "%7C",
		"}":  "%7D",
		"~":  "%7E",
		"\n": "%0A",
	}
	for k, v := range schars {
		s = strings.Replace(s, k, v, -1)
	}
	return s
}

func Link(text, url string) string {
	return "[" + text + "](" + escape(url) + ")"
}
