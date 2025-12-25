package boards2

import (
	"strconv"
	"strings"

	"gno.land/p/moul/md"

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
