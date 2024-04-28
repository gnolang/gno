package boards

import (
	"std"
	"strings"

	"gno.land/p/demo/boards"
	"gno.land/r/demo/users"
)

//----------------------------------------
// private utility methods
// XXX ensure these cannot be called from public.

func getBoard(bid boards.BoardID) *boards.Board {
	bidkey := boards.BoardIDKey(bid)
	board_, exists := gBoards.Get(bidkey)
	if !exists {
		return nil
	}
	board := board_.(*boards.Board)
	return board
}

func incGetBoardID() boards.BoardID {
	gBoardsCtr++
	return boards.BoardID(gBoardsCtr)
}

func displayAddressMD(addr std.Address) string {
	user := users.GetUserByAddress(addr)
	if user == nil {
		return "[" + addr.String() + "](/r/demo/users:" + addr.String() + ")"
	} else {
		return "[@" + user.Name + "](/r/demo/users:" + user.Name + ")"
	}
}

func usernameOf(addr std.Address) string {
	user := users.GetUserByAddress(addr)
	if user == nil {
		return ""
	}
	return user.Name
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
