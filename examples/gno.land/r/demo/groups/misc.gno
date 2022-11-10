package groups

import (
	"std"
	"strconv"
	"strings"

	"gno.land/r/demo/users"
)

//----------------------------------------
// private utility methods
// XXX ensure these cannot be called from public.

func getGroup(gid GroupID) *Group {
	gidkey := groupIDKey(gid)
	group_, exists := gGroups.Get(gidkey)
	if !exists {
		return nil
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
	} else {
		return strings.Repeat(" ", length-len(str)) + str
	}
}

func padZero(u64 uint64, length int) string {
	str := strconv.Itoa(int(u64))
	if len(str) >= length {
		return str
	} else {
		return strings.Repeat("0", length-len(str)) + str
	}
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

func displayAddressMD(addr std.Address) string {
	user := users.GetUserByAddress(addr)
	if user == nil {
		return "[" + addr.String() + "](/r/users:" + addr.String() + ")"
	} else {
		return "[@" + user.Name() + "](/r/users:" + user.Name() + ")"
	}
}

func usernameOf(addr std.Address) string {
	user := users.GetUserByAddress(addr)
	if user == nil {
		return ""
	} else {
		return user.Name()
	}
}

func membInpParser(data string) []Member {
	var opArr []Member
	s := strings.Split(data, "|")
	for _, ele := range s {
		var op Member
		if ele == "" {
			continue
		}
		k := strings.Split(ele, ",")
		l := len(k)
		op.address = k[0]
		if l > 1 {
			w, err := strconv.Atoi(k[1])
			if err != nil {
				op.weight = w
			}
		}
		if l > 2 {
			op.metadata = k[2]
		}
		opArr = append(opArr, op)
	}
	return opArr
}
