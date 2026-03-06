package groups

import (
	"strconv"
	"strings"

	"gno.land/r/sys/users"
)

//----------------------------------------
// private utility methods
// XXX ensure these cannot be called from public.

func getGroup(gid GroupID) *Group {
	gidkey := groupIDKey(gid)
	group_, exists := gGroups.Get(gidkey)
	if !exists {
		panic("group id (" + gid.String() + ") does not exists")
	}
	group := group_.(*Group)
	return group
}

func incGetGroupID() GroupID {
	gGroupsCtr++
	return GroupID(gGroupsCtr)
}

func padLeft(str string, length int) string {
	if len(str) >= length {
		return str
	}
	return strings.Repeat(" ", length-len(str)) + str
}

func padZero(u64 uint64, length int) string {
	str := strconv.Itoa(int(u64))
	if len(str) >= length {
		return str
	}
	return strings.Repeat("0", length-len(str)) + str
}

func groupIDKey(gid GroupID) string {
	return padZero(uint64(gid), 10)
}

func memberIDKey(mid MemberID) string {
	return padZero(uint64(mid), 10)
}

func indentBody(indent string, body string) string {
	lines := strings.Split(body, "\n")
	res := ""
	for i, line := range lines {
		if i > 0 {
			res += "\n"
		}
		res += indent + line
	}
	return res
}

// NOTE: length must be greater than 3.
func summaryOf(str string, length int) string {
	lines := strings.SplitN(str, "\n", 2)
	line := lines[0]
	if len(line) > length {
		line = line[:(length-3)] + "..."
	} else if len(lines) > 1 {
		// len(line) <= 80
		line = line + "..."
	}
	return line
}

func displayAddressMD(addr address) string {
	user := users.ResolveAddress(addr)
	if user == nil {
		return "[" + addr.String() + "](/u/" + addr.String() + ")"
	}
	return "[@" + user.Name() + "](/u/" + user.Name() + ")"
}

func usernameOf(addr address) string {
	user := users.ResolveAddress(addr)
	if user == nil {
		panic("user not found")
	}
	return user.Name()
}

func isValidPermission(perm Permission) bool {
	return perm == EditPermission || perm == DeletePermission
}
