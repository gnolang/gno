package boards

import (
	"strings"

	"gno.land/p/demo/json"
	"gno.land/p/demo/teritori/dao_interfaces"
)

// Create board

type ExecutableMessageCreateBoard struct {
	dao_interfaces.ExecutableMessage

	Name string
}

var _ dao_interfaces.ExecutableMessage = &ExecutableMessageCreateBoard{}

func (msg ExecutableMessageCreateBoard) Type() string {
	return "gno.land/r/demo/teritori/modboards.CreateBoard"
}

func (msg *ExecutableMessageCreateBoard) String() string {
	var ss []string
	ss = append(ss, msg.Type())
	ss = append(ss, "Name: "+msg.Name)
	return strings.Join(ss, "\n---\n")
}

func (msg *ExecutableMessageCreateBoard) ToJSON() *json.Node {
	return json.ObjectNode("", map[string]*json.Node{
		"name": json.StringNode("", msg.Name),
	})
}

func (msg *ExecutableMessageCreateBoard) FromJSON(ast *json.Node) {
	obj := ast.MustObject()
	msg.Name = obj["name"].MustString()
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

func (h CreateBoardHandler) Type() string {
	return ExecutableMessageCreateBoard{}.Type()
}

func (h *CreateBoardHandler) Instantiate() dao_interfaces.ExecutableMessage {
	return &ExecutableMessageCreateBoard{}
}

// Delete post

type ExecutableMessageDeletePost struct {
	dao_interfaces.ExecutableMessage

	BoardID  BoardID
	ThreadID PostID
	PostID   PostID
	Reason   string
}

var _ dao_interfaces.ExecutableMessage = &ExecutableMessageDeletePost{}

func (msg ExecutableMessageDeletePost) Type() string {
	return "gno.land/r/demo/teritori/modboards.DeletePost"
}

func (msg *ExecutableMessageDeletePost) String() string {
	var ss []string
	ss = append(ss, msg.Type())

	board := getBoard(msg.BoardID)
	s := ""

	if board != nil {
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

func (msg *ExecutableMessageDeletePost) ToJSON() *json.Node {
	return json.ObjectNode("", map[string]*json.Node{
		"boardId":  msg.BoardID.ToJSON(),
		"threadId": msg.ThreadID.ToJSON(),
		"postId":   msg.PostID.ToJSON(),
		"reason":   json.StringNode("", msg.Reason),
	})
}

func (msg *ExecutableMessageDeletePost) FromJSON(ast *json.Node) {
	obj := ast.MustObject()
	msg.BoardID.FromJSON(obj["boardId"])
	msg.ThreadID.FromJSON(obj["threadId"])
	msg.PostID.FromJSON(obj["postId"])
	msg.Reason = obj["reason"].MustString()
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

func (h DeletePostHandler) Type() string {
	return ExecutableMessageDeletePost{}.Type()
}

func (h *DeletePostHandler) Instantiate() dao_interfaces.ExecutableMessage {
	return &ExecutableMessageDeletePost{}
}
