package gnoweb

import "testing"

func TestNegotiatesMarkdown(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		accept string
		want   bool
	}{
		{"empty", "", false},
		{"plain markdown", "text/markdown", true},
		{"x-markdown alias", "text/x-markdown", true},
		{"case insensitive", "TEXT/Markdown", true},
		{"markdown with charset param", "text/markdown; charset=utf-8", true},
		{"markdown q half", "text/markdown;q=0.5", true},
		{"markdown q one", "text/markdown;q=1.0", true},
		{"markdown q zero", "text/markdown;q=0", false},
		{"markdown q zero point zero", "text/markdown;q=0.0", false},
		{"markdown negative q", "text/markdown;q=-1", false},
		{"markdown malformed q", "text/markdown;q=abc", true},
		{"html then markdown", "text/html, text/markdown", true},
		{"browser accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8", false},
		{"claude-code webfetch accept", "text/markdown, text/html, */*", true},
		{"wildcard only", "*/*", false},
		{"text wildcard", "text/*", false},
		{"json", "application/json", false},
		{"surrounding spaces", "  text/markdown  ", true},
		{"markdown present with non-zero q among others", "text/html;q=0.9, text/markdown;q=0.8", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := negotiatesMarkdown(tc.accept); got != tc.want {
				t.Errorf("negotiatesMarkdown(%q) = %v, want %v", tc.accept, got, tc.want)
			}
		})
	}
}
