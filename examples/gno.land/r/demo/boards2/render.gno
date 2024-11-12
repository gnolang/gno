package boards

import (
	"strconv"
	"strings"
)

//----------------------------------------
// Render functions

// TODO: Is this function really needed or Render("BOARD_NAME") is enough?
// TODO: Alternatively support ID or BOARD_NAME in Render()
func RenderBoard(bid BoardID) string {
	board := getBoard(bid)
	if board == nil {
		return "missing board"
	}
	return board.RenderBoard()
}

func Render(path string) string {
	// TODO: Split into four private render functions for each case
	if path == "" {
		s := "These are all the boards of this realm:\n\n"
		gBoards.Iterate("", "", func(_ string, value interface{}) bool {
			board := value.(*Board)
			s += " * [" + board.url + "](" + board.url + ")\n"
			return false
		})
		return s
	}

	parts := strings.Split(path, "/")
	if len(parts) == 1 {
		// /r/demo/boards:BOARD_NAME
		name := parts[0]
		v, found := gBoardsByName.Get(name)
		if !found {
			return "board does not exist: " + name
		}
		return v.(*Board).RenderBoard()
	} else if len(parts) == 2 {
		// /r/demo/boards:BOARD_NAME/THREAD_ID
		name := parts[0]
		v, found := gBoardsByName.Get(name)
		if !found {
			return "board does not exist: " + name
		}

		pid, err := strconv.Atoi(parts[1])
		if err != nil {
			return "invalid thread id: " + parts[1]
		}

		board := v.(*Board)
		thread := board.GetThread(PostID(pid))
		if thread == nil {
			return "thread does not exist with id: " + parts[1]
		}
		return thread.RenderPost("", 5)
	} else if len(parts) == 3 {
		// /r/demo/boards:BOARD_NAME/THREAD_ID/REPLY_ID
		name := parts[0]
		v, found := gBoardsByName.Get(name)
		if !found {
			return "board does not exist: " + name
		}

		pid, err := strconv.Atoi(parts[1])
		if err != nil {
			return "invalid thread id: " + parts[1]
		}

		board := v.(*Board)
		thread := board.GetThread(PostID(pid))
		if thread == nil {
			return "thread does not exist with id: " + parts[1]
		}

		rid, err := strconv.Atoi(parts[2])
		if err != nil {
			return "invalid reply id: " + parts[2]
		}

		reply := thread.GetReply(PostID(rid))
		if reply == nil {
			return "reply does not exist with id: " + parts[2]
		}
		return reply.RenderInner()
	} else {
		return "unrecognized path " + path
	}
}
