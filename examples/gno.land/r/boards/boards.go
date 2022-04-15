package boards

import (
	"regexp"
	"std"
	"strconv"
	"strings"

	"gno.land/p/avl"
	"gno.land/r/users"
)

//----------------------------------------
// Realm (package) state

var gBoards *avl.Tree       // id -> *Board
var gBoardsCtr int          // increments Board.id
var gBoardsByName *avl.Tree // name -> *Board

//----------------------------------------
// Constants

var reName = regexp.MustCompile(`^[a-z]+[_a-z0-9]{2,29}$`)

//----------------------------------------
// Public facing functions

func GetBoardIDFromName(name string) (BoardID, bool) {
	_, boardI, exists := gBoardsByName.Get(name)
	if !exists {
		return 0, false
	}
	return boardI.(*Board).id, true
}

func CreateBoard(name string) BoardID {
	std.AssertOriginCall()
	bid := incGetBoardID()
	caller := std.GetOrigCaller()
	url := "/r/boards:" + name
	board := newBoard(bid, url, name, caller)
	bidkey := boardIDKey(bid)
	gBoards, _ = gBoards.Set(bidkey, board)
	gBoardsByName, _ = gBoardsByName.Set(name, board)
	return board.id
}

func CreatePost(bid BoardID, title string, body string) PostID {
	std.AssertOriginCall()
	caller := std.GetOrigCaller()
	board := getBoard(bid)
	post := board.AddPost(caller, title, body)
	return post.id
}

func CreateReply(bid BoardID, threadid, postid PostID, body string) PostID {
	std.AssertOriginCall()
	caller := std.GetOrigCaller()
	board := getBoard(bid)
	thread := board.GetThread(threadid)
	if postid == threadid {
		reply := thread.AddReply(caller, body)
		return reply.id
	} else {
		post := thread.GetReply(postid)
		reply := post.AddReply(caller, body)
		return reply.id
	}
}

// If dstBoard is private, does not ping back.
// If board specified by bid is private, panics.
func CreateRepost(bid BoardID, postid PostID, title string, body string, dstBoardID BoardID) PostID {
	std.AssertOriginCall()
	caller := std.GetOrigCaller()
	board := getBoard(bid)
	if board.IsPrivate() {
		panic("cannot repost from a private board")
	}
	dst := getBoard(dstBoardID)
	thread := board.GetThread(postid)
	repost := thread.AddRepostTo(caller, title, body, dst)
	return repost.id
}

//----------------------------------------
// Query methods

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
		gBoards.Iterate("", "", func(n *avl.Tree) bool {
			board := n.Value().(*Board)
			str += " * [" + board.url + "](" + board.url + ")\n"
			return false
		})
		return str
	}
	parts := strings.Split(path, "/")
	if len(parts) == 1 {
		// /r/boards:BOARD_NAME
		name := parts[0]
		_, boardI, exists := gBoardsByName.Get(name)
		if !exists {
			return "board does not exist: " + name
		}
		return boardI.(*Board).RenderBoard()
	} else if len(parts) == 2 {
		// /r/boards:BOARD_NAME/THREAD_ID
		name := parts[0]
		_, boardI, exists := gBoardsByName.Get(name)
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
		// /r/boards:BOARD_NAME/THREAD_ID/REPLY_ID
		name := parts[0]
		_, boardI, exists := gBoardsByName.Get(name)
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

//----------------------------------------
// Board

type BoardID uint64

type Board struct {
	id        BoardID // only set for public boards.
	url       string
	name      string
	creator   std.Address
	posts     *avl.Tree // Post.id -> *Post
	postsCtr  uint64    // increments Post.id
	createdAt std.Time
}

func newBoard(id BoardID, url string, name string, creator std.Address) *Board {
	if !reName.MatchString(name) {
		panic("invalid name: " + name)
	}
	exists := gBoardsByName.Has(name)
	if exists {
		panic("board already exists")
	}
	return &Board{
		id:        id,
		url:       url,
		name:      name,
		creator:   creator,
		createdAt: std.GetTimestamp(),
	}
}

/* TODO support this once we figure out how to ensure URL correctness.
// A private board is not tracked by gBoards*,
// but must be persisted by the caller's realm.
// Private boards have 0 id and does not ping
// back the remote board on reposts.
func NewPrivateBoard(url string, name string, creator std.Address) *Board {
	return newBoard(0, url, name, creator)
}
*/

func (board *Board) IsPrivate() bool {
	return board.id == 0
}

func (board *Board) GetThread(pid PostID) *Post {
	pidkey := postIDKey(pid)
	_, postI, exists := board.posts.Get(pidkey)
	if !exists {
		return nil
	}
	return postI.(*Post)
}

func (board *Board) AddPost(creator std.Address, title string, body string) *Post {
	pid := board.incGetPostID()
	pidkey := postIDKey(pid)
	post := &Post{
		board:     board,
		id:        pid,
		creator:   creator,
		title:     title,
		body:      body,
		threadID:  pid,
		createdAt: std.GetTimestamp(),
	}
	board.posts, _ = board.posts.Set(pidkey, post)
	return post
}

// Renders the board for display suitable as plaintext in
// console.  This is suitable for demonstration or tests,
// but not for prod.
func (board *Board) RenderBoard() string {
	str := ""
	str += "\\[[post](" + board.GetPostFormURL() + ")]\n\n"
	if board.posts.Size() > 0 {
		board.posts.Iterate("", "", func(n *avl.Tree) bool {
			if str != "" {
				str += "----------------------------------------\n"
			}
			str += n.Value().(*Post).RenderSummary() + "\n"
			return false
		})
	}
	return str
}

func (board *Board) incGetPostID() PostID {
	board.postsCtr++
	return PostID(board.postsCtr)
}

func (board *Board) GetURLFromThreadAndReplyID(threadID, replyID PostID) string {
	if replyID == 0 {
		return board.url + "/" + strconv.Itoa(int(threadID))
	} else {
		return board.url + "/" + strconv.Itoa(int(threadID)) + "/" + strconv.Itoa(int(replyID))
	}
}

func (board *Board) GetPostFormURL() string {
	return "/r/boards?help&__func=CreatePost" +
		"&bid=" + strconv.Itoa(int(board.id)) +
		"&body.type=textarea"
}

//----------------------------------------
// Post

// NOTE: a PostID is relative to the board.
type PostID uint64

// A Post is a "thread" or a "reply" depending on context.
// A thread is a Post of a Board that holds other replies.
type Post struct {
	board       *Board
	id          PostID
	creator     std.Address
	title       string // optional
	body        string
	replies     *avl.Tree // Post.id -> *Post
	repliesAll  *avl.Tree // Post.id -> *Post (all replies, for top-level posts)
	reposts     *avl.Tree // Board.id -> Post.id
	threadID    PostID    // original Post.id
	replyTo     PostID    // parent Post.id (if reply or repost)
	repostBoard BoardID   // original Board.id (if repost)
	createdAt   std.Time
}

func (post *Post) IsThread() bool {
	return post.replyTo == 0
}

func (post *Post) GetPostID() PostID {
	return post.id
}

func (post *Post) AddReply(creator std.Address, body string) *Post {
	board := post.board
	pid := board.incGetPostID()
	pidkey := postIDKey(pid)
	reply := &Post{
		board:     board,
		id:        pid,
		creator:   creator,
		body:      body,
		threadID:  post.threadID,
		replyTo:   post.id,
		createdAt: std.GetTimestamp(),
	}
	post.replies, _ = post.replies.Set(pidkey, reply)
	if post.threadID == post.id {
		post.repliesAll, _ = post.repliesAll.Set(pidkey, reply)
	} else {
		thread := board.GetThread(post.threadID)
		thread.repliesAll, _ = thread.repliesAll.Set(pidkey, reply)
	}
	return reply
}

func (thread *Post) GetReply(pid PostID) *Post {
	pidkey := postIDKey(pid)
	_, replyI, ok := thread.repliesAll.Get(pidkey)
	if !ok {
		return nil
	} else {
		return replyI.(*Post)
	}
}

func (post *Post) AddRepostTo(creator std.Address, title, body string, dst *Board) *Post {
	pid := dst.incGetPostID()
	pidkey := postIDKey(pid)
	repost := &Post{
		board:       dst,
		id:          pid,
		creator:     creator,
		title:       title,
		body:        body,
		threadID:    pid,
		replyTo:     post.id,
		repostBoard: post.board.id,
		createdAt:   std.GetTimestamp(),
	}
	dst.posts, _ = dst.posts.Set(pidkey, repost)
	if !dst.IsPrivate() {
		bidkey := boardIDKey(dst.id)
		post.reposts, _ = post.reposts.Set(bidkey, pid)
	}
	return repost
}

func (post *Post) GetSummary() string {
	return summaryOf(post.body, 80)
}

func (post *Post) GetURL() string {
	if post.IsThread() {
		return post.board.GetURLFromThreadAndReplyID(
			post.id, 0)
	} else {
		return post.board.GetURLFromThreadAndReplyID(
			post.threadID, post.id)
	}
}

func (post *Post) GetReplyFormURL() string {
	return "/r/boards?help&__func=CreateReply" +
		"&bid=" + strconv.Itoa(int(post.board.id)) +
		"&threadid=" + strconv.Itoa(int(post.threadID)) +
		"&postid=" + strconv.Itoa(int(post.id)) +
		"&body.type=textarea"
}

func (post *Post) RenderSummary() string {
	str := ""
	if post.title != "" {
		str += "## [" + summaryOf(post.title, 80) + "](" + post.GetURL() + ")\n"
		str += "\n"
	}
	str += post.GetSummary() + "\n"
	str += "- by " + displayAddress(post.creator) + ", "
	str += "[" + std.FormatTimestamp(post.createdAt, "2006-01-02 3:04pm MST") + "](" + post.GetURL() + ") "

	str += "(" + strconv.Itoa(post.replies.Size()) + " replies)" + "\n"
	return str
}

func (post *Post) RenderPost(indent string, levels int) string {
	if post == nil {
		return "nil post"
	}
	str := ""
	if post.title != "" {
		str += indent + "# " + post.title + "\n"
		str += indent + "\n"
	}
	str += indentBody(indent, post.body) + "\n" // TODO: indent body lines.
	str += indent + "- by " + displayAddress(post.creator) + ", "
	str += "[" + std.FormatTimestamp(post.createdAt, "2006-01-02 3:04pm (MST)") + "](" + post.GetURL() + ")"
	str += " [reply](" + post.GetReplyFormURL() + ")" + "\n"
	if levels > 0 {
		if post.replies.Size() > 0 {
			post.replies.Iterate("", "", func(n *avl.Tree) bool {
				str += indent + "\n"
				str += n.Value().(*Post).RenderPost(indent+"> ", levels-1)
				return false
			})
		}
	} else {
		if post.replies.Size() > 0 {
			str += indent + "\n"
			str += indent + "_[see all " + strconv.Itoa(post.replies.Size()) + " replies](" + post.GetURL() + ")_\n"
		}
	}
	return str
}

// render reply and link to context thread
func (post *Post) RenderInner() string {
	if post.IsThread() {
		panic("unexpected thread")
	}
	threadID := post.threadID
	replyID := post.id
	parentID := post.replyTo
	str := ""
	str += "_[see thread](" + post.board.GetURLFromThreadAndReplyID(
		threadID, 0) + ")_\n\n"
	thread := post.board.GetThread(post.threadID)
	var parent *Post
	if thread.id == parentID {
		parent = thread
	} else {
		parent = thread.GetReply(parentID)
	}
	str += parent.RenderPost("", 0)
	str += "\n"
	str += post.RenderPost("> ", 5)
	return str
}

//----------------------------------------
// private utility methods
// XXX ensure these cannot be called from public.

func getBoard(bid BoardID) *Board {
	bidkey := boardIDKey(bid)
	_, board_, exists := gBoards.Get(bidkey)
	if !exists {
		return nil
	}
	board := board_.(*Board)
	return board
}

func incGetBoardID() BoardID {
	gBoardsCtr++
	return BoardID(gBoardsCtr)
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

func boardIDKey(bid BoardID) string {
	return padZero(uint64(bid), 10)
}

func postIDKey(pid PostID) string {
	return padZero(uint64(pid), 10)
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

func displayAddress(input std.Address) string {
	user := users.GetUserByAddress(input)
	if user == nil {
		return input.String()
	}
	return user.Name() + " (" + input.String() + ")"
}
