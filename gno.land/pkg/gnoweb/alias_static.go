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

	for line := range strings.SplitSeq(head, "\n") {
		key, value, found := strings.Cut(line, ":")
		if !found {
			continue
		}
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		switch strings.TrimSpace(key) {
		case "title":
			target.Title = value
		case "description":
			target.Description = markdown.TruncateDescription(value)
		}
	}
	target.Value = strings.TrimLeft(body, "\n")

	return target
}
