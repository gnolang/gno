package boards2

import (
	"std"
	"strconv"
	"strings"

	"gno.land/r/sys/users"
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

// newUserLink returns a Markdown link for an account to the users
// realm or the address as string when a user is not registered.
func newUserLink(addr std.Address) string {
	user := users.ResolveAddress(addr)
	if user != nil {
		return user.RenderLink("")
	}
	return addr.String()
}
