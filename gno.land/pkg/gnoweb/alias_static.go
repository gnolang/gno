package gnoweb

import (
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown"
)

// NewStaticAlias builds a static-markdown alias from a file's content, lifting
// an optional front matter block off the top:
//
//	---
//	title: About
//	description: What gno.land is and why it exists.
//	---
//
// It reads only the pages gnoweb is configured with, which the operator ships.
// Realm content never goes through it: a realm is permissionless, and metadata
// it could set without displaying it is the abuse #3910 warned about.
func NewStaticAlias(content string) AliasTarget {
	target := AliasTarget{Kind: StaticMarkdown, Value: content}

	rest, found := strings.CutPrefix(content, "---\n")
	if !found {
		return target
	}
	head, body, found := strings.Cut(rest, "\n---\n")
	if !found {
		return target
	}

	// A block is front matter only if every line could be YAML and at least
	// one is a key, so a page drawing a rule above and below a heading, a
	// URL or a "Note: ..." line keeps it. Fields are set only once the whole
	// block qualifies.
	var title, description string
	sawKey := false
	for line := range strings.SplitSeq(head, "\n") {
		key, value, found := strings.Cut(line, ":")
		key = strings.TrimRight(key, " \t")
		// As in YAML, a colon makes a key only at the end of the line or
		// before a space or tab, so https://... or mailto:... is prose.
		isKey := found && (value == "" || value[0] == ' ' || value[0] == '\t')
		if !isKey || !isFrontMatterKey(key) {
			// Blank lines, YAML comments, list items and indented
			// continuations are front matter too; anything else is prose.
			if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") ||
				strings.HasPrefix(line, "- ") || strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
				continue
			}
			return target
		}
		sawKey = true
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		switch key {
		case "title":
			title = markdown.TruncateTitle(value)
		case "description":
			description = markdown.TruncateDescription(value)
		}
	}
	if !sawKey {
		return target
	}
	target.Title, target.Description = title, description
	target.Value = strings.TrimLeft(body, "\n")

	return target
}

// isFrontMatterKey reports whether key looks like a front matter key: a
// lowercase word, as Jekyll and Hugo write them. Sentence-case prose such as
// "Note: read this." does not qualify.
func isFrontMatterKey(key string) bool {
	if key == "" || key[0] < 'a' || key[0] > 'z' {
		return false
	}
	for _, c := range key {
		if (c < 'a' || c > 'z') && (c < '0' || c > '9') && c != '_' && c != '-' {
			return false
		}
	}
	return true
}
