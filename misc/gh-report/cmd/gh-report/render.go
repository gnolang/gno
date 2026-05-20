package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func RenderJSON(w io.Writer, r Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}

const titleMax = 70

func truncTitle(s string) string {
	if len(s) <= titleMax {
		return s
	}
	return s[:titleMax-1] + "…"
}

func metaBracket(section Section, entry Entry) string {
	kind := "issue"
	if entry.Kind == KindPR {
		kind = "PR"
	}
	d := ageDays(entry.UpdatedAt)
	if section.Name == "Depends on other core" {
		for _, h := range OtherCore {
			if hasAny(entry.RequestedReviewer, h) {
				return fmt.Sprintf("%s/%dd/@%s", kind, d, h)
			}
		}
		for _, h := range OtherCore {
			if hasAny(entry.Assignees, h) {
				return fmt.Sprintf("%s/%dd/@%s", kind, d, h)
			}
		}
	}
	return fmt.Sprintf("%s/%dd", kind, d)
}

func RenderMarkdown(w io.Writer, r Report) error {
	var b strings.Builder
	for i, s := range r.Sections {
		if i > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "## %s (%d)\n", s.Name, s.Count())
		for _, e := range s.Entries {
			fmt.Fprintf(&b, "- [#%d](%s) [%s] %s (%s, +%dc)\n",
				e.Number, e.URL, metaBracket(s, e), truncTitle(e.Title), e.Author, e.Comments)
		}
	}
	_, err := io.WriteString(w, b.String())
	return err
}
