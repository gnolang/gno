package boards

import (
	"strconv"
	"strings"
)

//----------------------------------------
// Render functions

func RenderBoard(bid BoardID) string {
	board := getBoard(bid)
	if board == nil {
		return "missing board"
	}
	return board.RenderBoard()
}

func Render(path string) string {
	if path == "" {
		str := "These are all the boards of this realm:\n\n"
		gBoards.Iterate("", "", func(key string, value any) bool {
			board := value.(*Board)
			str += " * [" + board.url + "](" + board.url + ")\n"
			return false
		})
		return str
	}
	parts := strings.Split(path, "/")
	if len(parts) == 1 {
		// /r/archive/boards:BOARD_NAME
		name := parts[0]
		boardI, exists := gBoardsByName.Get(name)
		if !exists {
			return "board does not exist: " + name
		}
		return boardI.(*Board).RenderBoard()
	} else if len(parts) == 2 {
		// /r/archive/boards:BOARD_NAME/THREAD_ID
		name := parts[0]
		boardI, exists := gBoardsByName.Get(name)
		if !exists {
			return "board does not exist: " + name
		}
		pid, err := strconv.Atoi(parts[1])
		if err != nil {
			return "invalid thread id: " + parts[1]
		}
		board := boardI.(*Board)
		thread := board.GetThread(PostID(pid))
		if thread == nil {
			return "thread does not exist with id: " + parts[1]
		}
		return thread.RenderPost("", 5)
	} else if len(parts) == 3 {
		// /r/archive/boards:BOARD_NAME/THREAD_ID/REPLY_ID
		name := parts[0]
		boardI, exists := gBoardsByName.Get(name)
		if !exists {
			return "board does not exist: " + name
		}
		pid, err := strconv.Atoi(parts[1])
		if err != nil {
			return "invalid thread id: " + parts[1]
		}
		board := boardI.(*Board)
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
