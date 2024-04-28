package boards

import (
	"strconv"
	"strings"

	"gno.land/p/demo/boards"
)

//----------------------------------------
// Render functions

func RenderBoardById(bid boards.BoardID) string {
	board := getBoard(bid)
	if board == nil {
		return "missing board"
	}
	return RenderBoard(board)
}

func Render(path string) string {
	if path == "" {
		str := "These are all the boards of this realm:\n\n"
		gBoards.Iterate("", "", func(key string, value interface{}) bool {
			board := value.(*boards.Board)
			str += " * [" + board.Url() + "](" + board.Url() + ")\n"
			return false
		})
		return str
	}
	parts := strings.Split(path, "/")
	if len(parts) == 1 {
		// /r/demo/boards:BOARD_NAME
		name := parts[0]
		boardI, exists := gBoardsByName.Get(name)
		if !exists {
			return "board does not exist: " + name
		}
		return RenderBoard(boardI.(*boards.Board))
	} else if len(parts) == 2 {
		// /r/demo/boards:BOARD_NAME/THREAD_ID
		name := parts[0]
		boardI, exists := gBoardsByName.Get(name)
		if !exists {
			return "board does not exist: " + name
		}
		pid, err := strconv.Atoi(parts[1])
		if err != nil {
			return "invalid thread id: " + parts[1]
		}
		board := boardI.(*boards.Board)
		thread := board.GetThread(boards.PostID(pid))
		if thread == nil {
			return "thread does not exist with id: " + parts[1]
		}
		return RenderPost(thread, "", 5)
	} else if len(parts) == 3 {
		// /r/demo/boards:BOARD_NAME/THREAD_ID/REPLY_ID
		name := parts[0]
		boardI, exists := gBoardsByName.Get(name)
		if !exists {
			return "board does not exist: " + name
		}
		pid, err := strconv.Atoi(parts[1])
		if err != nil {
			return "invalid thread id: " + parts[1]
		}
		board := boardI.(*boards.Board)
		thread := board.GetThread(boards.PostID(pid))
		if thread == nil {
			return "thread does not exist with id: " + parts[1]
		}
		rid, err := strconv.Atoi(parts[2])
		if err != nil {
			return "invalid reply id: " + parts[2]
		}
		reply := thread.GetReply(boards.PostID(rid))
		if reply == nil {
			return "reply does not exist with id: " + parts[2]
		}
		return RenderInner(reply)
	} else {
		return "unrecognized path " + path
	}
}
