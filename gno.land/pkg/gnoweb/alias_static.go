package gnoweb

import (
	"os"
	"strings"
	"sync"
	"time"

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

	// An editor may save the file with CRLF endings or a byte order mark;
	// neither changes what the block says.
	normalized := strings.ReplaceAll(strings.TrimPrefix(content, "\uFEFF"), "\r\n", "\n")
	if !strings.HasSuffix(normalized, "\n") {
		normalized += "\n" // a file of front matter alone may end on its rule
	}
	rest, found := strings.CutPrefix(normalized, "---\n")
	if !found {
		return target
	}
	// A paragraph right above a rule needs a blank line between them, or
	// the rule makes it a heading: a block ending on a blank line is page
	// content, whatever its lines look like.
	head, body, found := strings.Cut(rest, "\n---\n")
	if !found || strings.HasSuffix(strings.TrimRight(head, " \t"), "\n") {
		return target
	}

	// A block is front matter only if every line could be YAML and at least
	// one is a key, so a page drawing a rule above and below a heading, a
	// URL or a "Note: ..." line keeps it. Fields are set only once the whole
	// block qualifies.
	var title, description string
	sawKey := false
	lines := strings.Split(head, "\n")
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		key, value, found := strings.Cut(line, ":")
		key = strings.TrimRight(key, " \t")
		// As in YAML, a colon makes a key only at the end of the line or
		// before a space or tab, so https://... or mailto:... is prose.
		isKey := found && (value == "" || value[0] == ' ' || value[0] == '\t')
		if !isKey || !isFrontMatterKey(key) {
			// Blank lines, YAML comments, list items and indented
			// continuations are front matter too; anything else is prose.
			if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") ||
				strings.HasPrefix(line, "- ") || isIndented(line) {
				continue
			}
			return target
		}
		sawKey = true
		// A value goes on over the indented and blank lines below its key.
		end := i + 1
		for end < len(lines) && (isIndented(lines[end]) || strings.TrimSpace(lines[end]) == "") {
			end++
		}
		value = scalarValue(value, lines[i+1:end])
		i = end - 1
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

func isIndented(line string) bool {
	return strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t")
}

// scalarValue reads a value from the rest of its key's line and the lines
// that continue it. A folded (>) or literal (|) block is its continuation
// lines; a plain or quoted value is its first line and its continuations.
// The head puts every field on one line, so the lines are joined with a
// space either way, and a quoted value loses its one wrapping pair of quotes.
func scalarValue(first string, more []string) string {
	first = strings.TrimSpace(first)
	parts := make([]string, 0, len(more)+1)
	block := isBlockIndicator(first)
	if !block && first != "" {
		parts = append(parts, first)
	}
	for _, line := range more {
		if line = strings.TrimSpace(line); line != "" {
			parts = append(parts, line)
		}
	}
	value := strings.Join(parts, " ")
	if block {
		return value
	}
	// A quoted value loses its wrapping pair, which a comment may follow.
	if value != "" && (value[0] == '"' || value[0] == '\'') {
		end := strings.LastIndexByte(value, value[0])
		if after := value[end+1:]; end > 0 && (after == "" || strings.HasPrefix(after, " #")) {
			if value[0] == '\'' {
				return strings.ReplaceAll(value[1:end], "''", "'")
			}
			return value[1:end]
		}
	}
	// A plain value ends at a comment.
	value, _, _ = strings.Cut(value, " #")
	return value
}

// isBlockIndicator reports whether v opens a YAML block scalar: > or |,
// then an optional chomping or indentation indicator and comment.
func isBlockIndicator(v string) bool {
	if v == "" || (v[0] != '>' && v[0] != '|') {
		return false
	}
	v, _, _ = strings.Cut(v[1:], "#")
	return strings.Trim(v, "+-0123456789 \t") == ""
}

// NewStaticFileAlias reads a static-markdown alias from path. The page is
// re-read whenever the file's size or modification time changes, so the file
// can be replaced (ideally by renaming a new one over it) without restarting
// gnoweb. If a later read fails, the last content read is kept.
func NewStaticFileAlias(path string) (AliasTarget, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return AliasTarget{}, err
	}
	f := &staticFile{path: path, target: NewStaticAlias(string(content))}
	return f.current(), nil
}

type staticFile struct {
	path string

	mu      sync.Mutex
	modTime time.Time // zero: read the file again on the next request
	size    int64
	target  AliasTarget
}

// staticFileSettle is how old a file's modification time must be before a
// read of it is trusted. A write within the same timestamp tick as the read
// (coarse filesystem clocks, an in-place write caught halfway) would
// otherwise leave stale content until the next change.
const staticFileSettle = 2 * time.Second

func (f *staticFile) current() AliasTarget {
	f.mu.Lock()
	defer f.mu.Unlock()
	if info, err := os.Stat(f.path); err == nil && (!info.ModTime().Equal(f.modTime) || info.Size() != f.size) {
		if content, err := os.ReadFile(f.path); err == nil {
			f.target, f.size, f.modTime = NewStaticAlias(string(content)), info.Size(), info.ModTime()
			if time.Since(f.modTime) < staticFileSettle {
				f.modTime = time.Time{}
			}
		}
	}
	t := f.target
	t.file = f
	return t
}
