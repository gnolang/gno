package boards2

import (
	"std"
	"strconv"
	"strings"

	"gno.land/r/demo/users"
)

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
			res += "\n"
		}
		res += indent + line
	}
	return res
}

// NOTE: length must be greater than 3.
func summaryOf(text string, length int) string {
	lines := strings.SplitN(text, "\n", 2)
	line := lines[0]
	if len(line) > length {
		line = line[:(length-3)] + "..."
	} else if len(lines) > 1 {
		// len(line) <= 80
		line = line + "..."
	}
	return line
}

// newLink returns a Markdown link.
func newLink(label, uri string) string {
	return "[" + label + "](" + uri + ")"
}

// newUserLink returns a Markdown link for an account to the users realm.
func newUserLink(addr std.Address) string {
	user := users.GetUserByAddress(addr)
	if user == nil {
		return newLink(addr.String(), "/r/demo/users:"+addr.String())
	}
	return newLink("@"+user.Name, "/r/demo/users:"+user.Name)
}
