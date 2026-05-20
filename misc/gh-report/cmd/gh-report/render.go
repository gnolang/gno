package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
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

func ansi(color, s string) string {
	if os.Getenv("NO_COLOR") != "" {
		return s
	}
	return "\x1b[" + color + "m" + s + "\x1b[0m"
}

func RenderANSI(w io.Writer, r Report) error {
	var b strings.Builder
	bold := func(s string) string { return ansi("1;36", s) }
	num := func(s string) string { return ansi("33", s) }
	red := func(s string) string { return ansi("31", s) }

	for i, s := range r.Sections {
		if i > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "%s\n", bold(fmt.Sprintf("%s (%d)", s.Name, s.Count())))
		for _, e := range s.Entries {
			d := ageDays(e.UpdatedAt)
			dateStr := fmt.Sprintf("%dd", d)
			if d >= StaleDays {
				dateStr = red(dateStr)
			}
			kind := "issue"
			if e.Kind == KindPR {
				kind = "PR"
			}
			fmt.Fprintf(&b, "- %s [%s/%s] %s (%s, +%dc)\n",
				num("#"+strconv.Itoa(e.Number)), kind, dateStr, truncTitle(e.Title), e.Author, e.Comments)
		}
	}
	_, err := io.WriteString(w, b.String())
	return err
}
