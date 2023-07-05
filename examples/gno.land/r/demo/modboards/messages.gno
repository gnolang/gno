package boards

import (
	"encoding/binary"
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/daodao/interfaces"
)

// Create board

type ExecutableMessageCreateBoard struct {
	dao_interfaces.ExecutableMessage

	Name string
}

func (msg *ExecutableMessageCreateBoard) Type() string {
	return "CreateBoard"
}

func (msg *ExecutableMessageCreateBoard) String() string {
	var ss []string
	ss = append(ss, msg.Type())
	ss = append(ss, "Name: "+msg.Name)
	return strings.Join(ss, "\n---\n")
}

func (msg *ExecutableMessageCreateBoard) Binary() []byte {
	b := []byte{}

	t := msg.Type()
	b = binary.BigEndian.AppendUint16(b, uint16(len(t)))
	b = append(b, []byte(t)...)

	b = binary.BigEndian.AppendUint16(b, uint16(len(msg.Name)))
	b = append(b, []byte(msg.Name)...)

	return b
}

func ExecutableMessageCreateBoardFromBinary(b []byte) *ExecutableMessageCreateBoard {
	msg := &ExecutableMessageCreateBoard{}

	if len(b) < 2 {
		panic("invalid length - less than 2")
	}
	tl := binary.BigEndian.Uint16(b[:2])
	b = b[2:]
	if len(b) < int(tl) {
		panic("invalid length - less than expected")
	}
	t := string(b[:tl])
	if t != msg.Type() {
		panic("invalid type")
	}
	b = b[tl:]

	nl := binary.BigEndian.Uint16(b[:2])
	b = b[2:]
	if len(b) < int(nl) {
		panic("invalid length - less than expected")
	}
	n := string(b[:nl])
	// b = b[nl:]
	msg.Name = n

	return msg
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

func (h *CreateBoardHandler) FromBinary(b []byte) dao_interfaces.ExecutableMessage {
	return ExecutableMessageCreateBoardFromBinary(b)
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
	return "DeletePost"
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

func (msg *ExecutableMessageDeletePost) Binary() []byte {
	b := []byte{}

	t := msg.Type()
	b = binary.BigEndian.AppendUint16(b, uint16(len(t)))
	b = append(b, []byte(t)...)

	b = binary.BigEndian.AppendUint64(b, uint64(msg.BoardID))
	b = binary.BigEndian.AppendUint64(b, uint64(msg.ThreadID))
	b = binary.BigEndian.AppendUint64(b, uint64(msg.PostID))

	b = binary.BigEndian.AppendUint16(b, uint16(len(msg.Reason)))
	b = append(b, []byte(msg.Reason)...)

	return b
}

func ExecutableMessageDeletePostFromBinary(b []byte) *ExecutableMessageDeletePost {
	msg := &ExecutableMessageDeletePost{}

	if len(b) < 2 {
		panic("invalid length - less than 2")
	}
	tl := binary.BigEndian.Uint16(b[:2])
	b = b[2:]
	if len(b) < int(tl) {
		panic("invalid length - less than expected")
	}
	t := string(b[:tl])
	if t != msg.Type() {
		panic("invalid type")
	}
	b = b[tl:]

	if len(b) < 8 {
		panic("invalid length - less than 8")
	}
	msg.BoardID = BoardID(binary.BigEndian.Uint64(b[:8]))
	b = b[8:]

	if len(b) < 8 {
		panic("invalid length - less than 8")
	}
	msg.ThreadID = PostID(binary.BigEndian.Uint64(b[:8]))
	b = b[8:]

	if len(b) < 8 {
		panic("invalid length - less than 8")
	}
	msg.PostID = PostID(binary.BigEndian.Uint64(b[:8]))
	b = b[8:]

	rl := binary.BigEndian.Uint16(b[:2])
	b = b[2:]
	if len(b) < int(rl) {
		panic("invalid length - less than expected")
	}
	r := string(b[:rl])
	msg.Reason = r
	// b = b[rl:]

	return msg
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

func (h *DeletePostHandler) FromBinary(b []byte) dao_interfaces.ExecutableMessage {
	return ExecutableMessageDeletePostFromBinary(b)
}
