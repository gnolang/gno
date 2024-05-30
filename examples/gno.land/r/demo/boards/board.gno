package boards

import (
	"gno.land/p/demo/boards"
)

// Renders the board for display suitable as plaintext in
// console.  This is suitable for demonstration or tests,
// but not for prod.
func RenderBoard(board *boards.Board) string {
	str := ""
	str += "\\[[post](" + GetPostFormURL(board) + ")]\n\n"
	threads := board.Threads()
	if threads.Size() > 0 {
		threads.Iterate("", "", func(key string, value interface{}) bool {
			if str != "" {
				str += "----------------------------------------\n"
			}
			str += RenderSummary(value.(*boards.Post)) + "\n"
			return false
		})
	}
	return str
}

func GetPostFormURL(board *boards.Board) string {
	return "/r/demo/boards?help&__func=CreateThread" +
		"&bid=" + board.Id().String() +
		"&body.type=textarea"
}
