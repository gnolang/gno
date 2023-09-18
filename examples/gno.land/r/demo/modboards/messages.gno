package boards

import (
	"encoding/binary"
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/daodao/interfaces"
	"gno.land/p/demo/ujson"
)

// Create board

type ExecutableMessageCreateBoard struct {
	dao_interfaces.ExecutableMessage

	Name string
}

func (msg *ExecutableMessageCreateBoard) Type() string {
	return "gno.land/r/demo/modboards.CreateBoard"
}

func (msg *ExecutableMessageCreateBoard) String() string {
	var ss []string
	ss = append(ss, msg.Type())
	ss = append(ss, "Name: "+msg.Name)
	return strings.Join(ss, "\n---\n")
}

func (msg *ExecutableMessageCreateBoard) ToJSON() string {
	return ujson.FormatObject([]ujson.FormatKV{
		{Key: "name", Value: msg.Name},
	})
}

type CreateBoardHandler struct {
	dao_interfaces.MessageHandler
}

func NewCreateBoardHandler() *CreateBoardHandler {
	return &CreateBoardHandler{}
}

func (h *CreateBoardHandler) Execute(imsg dao_interfaces.ExecutableMessage) {
	msg := imsg.(*ExecutableMessageCreateBoard)
	CreateBoard(msg.Name)
}

func (h *CreateBoardHandler) Type() string {
	return ExecutableMessageCreateBoard{}.Type()
}

func (h *CreateBoardHandler) MessageFromJSON(ast *ujson.JSONASTNode) dao_interfaces.ExecutableMessage {
	msg := &ExecutableMessageCreateBoard{}
	ast.ParseObject([]*ujson.ParseKV{
		{Key: "name", Value: &msg.Name},
	})
	return msg
}

// Delete post

type ExecutableMessageDeletePost struct {
	dao_interfaces.ExecutableMessage

	BoardID  BoardID
	ThreadID PostID
	PostID   PostID
	Reason   string
}

func (msg *ExecutableMessageDeletePost) Type() string {
	return "gno.land/r/demo/modboards.DeletePost"
}

func (msg *ExecutableMessageDeletePost) String() string {
	var ss []string
	ss = append(ss, msg.Type())

	board, ok := getBoard(msg.BoardID).(*Board)
	s := ""

	if ok {
		s += "Board: " + board.name + " (" + board.id.String() + ")"

		thread := board.GetThread(msg.ThreadID)
		if thread != nil {
			s += "\nThread: " + thread.title + " (" + thread.id.String() + ")"
		} else {
			s += "\nThread: " + msg.ThreadID.String() + " (not found)"
		}

		if msg.PostID != msg.ThreadID {
			post := thread.GetReply(msg.PostID)
			if post != nil {
				s += "\nPost: " + post.title + " (" + post.id.String() + ")"
			} else {
				s += "\nPost: " + msg.PostID.String() + " (not found)"
			}
		}
	} else {
		s += "Board: " + msg.BoardID.String() + " (not found)"
	}

	s += "\nReason: " + msg.Reason

	ss = append(ss, s)

	return strings.Join(ss, "\n---\n")
}

func (msg *ExecutableMessageDeletePost) ToJSON() string {
	return ujson.FormatObject([]ujson.FormatKV{
		{Key: "boardId", Value: msg.BoardID},
		{Key: "threadId", Value: msg.ThreadID},
		{Key: "postId", Value: msg.PostID},
		{Key: "reason", Value: msg.Reason},
	})
}

type DeletePostHandler struct {
	dao_interfaces.MessageHandler
}

func NewDeletePostHandler() *DeletePostHandler {
	return &DeletePostHandler{}
}

func (h *DeletePostHandler) Execute(imsg dao_interfaces.ExecutableMessage) {
	msg := imsg.(*ExecutableMessageDeletePost)
	DeletePost(msg.BoardID, msg.ThreadID, msg.PostID, msg.Reason)
}

func (h *DeletePostHandler) Type() string {
	return ExecutableMessageDeletePost{}.Type()
}

func (h *DeletePostHandler) MessageFromJSON(ast *ujson.JSONASTNode) dao_interfaces.ExecutableMessage {
	msg := &ExecutableMessageDeletePost{}
	ast.ParseObject([]*ujson.ParseKV{
		{Key: "boardId", Value: &msg.BoardID},
		{Key: "threadId", Value: &msg.ThreadID},
		{Key: "postId", Value: &msg.PostID},
		{Key: "reason", Value: &msg.Reason},
	})
	return msg
}
