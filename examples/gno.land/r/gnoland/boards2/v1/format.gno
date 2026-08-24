package boards2

import (
	"strconv"
	"strings"

	"gno.land/p/gnoland/boards"
	"gno.land/p/moul/md"
	"gno.land/p/nt/markdown/foreign/v0"

	"gno.land/r/sys/users"
)

const dateFormat = "2006-01-02 3:04pm MST"

func padLeft(s string, length int) string {
	if len(s) >= length {
		return s
	}
	return strings.Repeat(" ", length-len(s)) + s
}

func padZero(u64 uint64, length int) string {
	s := strconv.Itoa(int(u64))
	if len(s) >= length {
		return s
	}
	return strings.Repeat("0", length-len(s)) + s
}

func indentBody(indent string, body string) string {
	var (
		res   string
		lines = strings.Split(body, "\n")
	)
	for i, line := range lines {
		if i > 0 {
			// Add two spaces to keep newlines within Markdown
			res += "  \n"
		}
		res += indent + line
	}
	return res
}

// indentForeignBody is the single "render a user body" operation: it
// sandboxes the body in a <gno-foreign> block, indents it like
// indentBody, AND charges one unit against the per-render budget — so
// wrapping and budget-accounting can't drift apart (the *int signals
// the mutation). The body renders inside the sandbox: its block
// structure (headings, blockquotes, lists, columns, alerts) is
// contained and cannot hijack realm chrome, so boards no longer needs
// the write-time markdown blacklist. The opener survives the "> "
// comment indentation at any depth (goldmark strips the "> " prefix
// before the block parser runs).
func indentForeignBody(indent, body string, budget *int) string {
	*budget-- // one <gno-foreign> block
	return indentBody(indent, foreign.Foreign(body))
}

func summaryOf(text string, length int) string {
	lines := strings.SplitN(text, "\n", 2)
	line := lines[0]
	if len(line) > length {
		line = line[:(length-3)] + "..."
	} else if len(lines) > 1 {
		line = line + "..."
	}
	return line
}

func userLink(addr address) string {
	if u := users.ResolveAddress(addr); u != nil {
		return md.UserLink(u.Name())
	}
	return md.UserLink(addr.String())
}

func getRoleBadge(post *boards.Post) string {
	if post == nil || post.Board == nil || post.Board.Permissions == nil {
		return ""
	}

	perms := post.Board.Permissions
	creator := post.Creator

	// Check roles in order of priority
	if perms.HasRole(creator, RoleOwner) {
		return " `owner`"
	}
	if perms.HasRole(creator, RoleAdmin) {
		return " `admin`"
	}
	if perms.HasRole(creator, RoleModerator) {
		return " `mod`"
	}
	return ""
}
