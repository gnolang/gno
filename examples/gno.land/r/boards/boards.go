package boards

import (
	"fmt"
	"regexp"
	"std"
	"strconv"
	"strings"

	"gno.land/p/avl"
)

//----------------------------------------
// Realm (package) state

var gBoards *avl.Tree       // id -> *Board
var gBoardsCtr int          // increments Board.id
var gBoardsByName *avl.Tree // name -> *Board

//----------------------------------------
// Constants

var reName = regexp.MustCompile(`^[a-z]+[_a-z0-9]*$`)

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
	if !std.IsOriginCall() {
		// TODO: consider making this a function
		// tag/decorator.
		panic("CreateBoard is public facing")
	}
	bid := incGetBoardID()
	caller := std.GetCaller()
	board := newBoard(bid, name, caller)
	bidkey := strconv.Itoa(int(bid))
	gBoards, _ = gBoards.Set(bidkey, board)
	gBoardsByName, _ = gBoardsByName.Set(name, board)
	return board.id
}

func CreatePost(bid BoardID, title string, body string) {
	if !std.IsOriginCall() {
		// TODO: consider making this a function
		// tag/decorator.
		panic("CreateBoard is public facing")
	}
	caller := std.GetCaller()
	board := getBoard(bid)
	board.AddPost(caller, title, body)
}

func CreateReply(bid BoardID, postid PostID, body string) {
	if !std.IsOriginCall() {
		// TODO: consider making this a function
		// tag/decorator.
		panic("CreateBoard is public facing")
	}
	caller := std.GetCaller()
	board := getBoard(bid)
	post := board.GetPost(postid)
	post.AddReply(caller, body)
}

// If dstBoard is private, does not ping back.
// If board specified by bid is private, panics.
func CreateRepost(bid BoardID, postid PostID, title string, body string, dstBoardID BoardID) {
	if !std.IsOriginCall() {
		// TODO: consider making this a function
		// tag/decorator.
		panic("CreateBoard is public facing")
	}
	caller := std.GetCaller()
	board := getBoard(bid)
	if board.IsPrivate() {
		panic("cannot repost from a private board")
	}
	dst := getBoard(dstBoardID)
	post := board.GetPost(postid)
	post.AddRepostTo(caller, title, body, dst)
}

//----------------------------------------
// Query methods

func RenderBoard(bid BoardID) string {
	board := getBoard(bid)
	if board == nil {
		return "missing board"
	}
	return board.Render()
}

func Render(path string) string {
	if path == "" {
		str := ""
		gBoards.Iterate("", "", func(n *avl.Tree) bool {
			board := n.Value().(*Board).name
			str += " * [" + board + "](gno.land/r/boards/" + board + ")\n"
			return false
		})
		return str
	}
	parts := strings.Split(path, "/")
	if len(parts) == 1 {
		name := parts[0]
		_, boardI, exists := gBoardsByName.Get(name)
		if !exists {
			return "board does not exist: " + name
		}
		return boardI.(*Board).Render()
	} else {
		return "unrecognized path " + path
	}
}

//----------------------------------------
// Board

type BoardID uint64

type Board struct {
	id       BoardID // only set for public boards.
	name     string
	creator  std.Address
	posts    *avl.Tree // Post.id -> *Post
	postsCtr uint64    // increments Post.id
}

func newBoard(id BoardID, name string, creator std.Address) *Board {
	if !reName.MatchString(name) {
		panic(fmt.Sprintf("invalid name %q", name))
	}
	exists := gBoardsByName.Has(name)
	if exists {
		panic("board already exists")
	}
	return &Board{
		id:      id,
		name:    name,
		creator: creator,
	}
}

// A private board is not tracked by gBoards*,
// but must be persisted by the caller's realm.
// Private boards have 0 id and does not ping
// back the remote board on reposts.
func NewPrivateBoard(name string, creator std.Address) *Board {
	return newBoard(0, name, creator)
}

func (board *Board) IsPrivate() bool {
	return board.id == 0
}

func (board *Board) GetPost(pid PostID) *Post {
	pidkey := strconv.Itoa(int(pid))
	_, postI, exists := board.posts.Get(pidkey)
	if !exists {
		return nil
	}
	return postI.(*Post)
}

func (board *Board) AddPost(creator std.Address, title string, body string) *Post {
	pid := board.incGetPostID()
	pidkey := strconv.Itoa(int(pid))
	post := &Post{
		board:   board,
		id:      pid,
		creator: creator,
		title:   title,
		body:    body,
	}
	board.posts, _ = board.posts.Set(pidkey, post)
	return post
}

// Renders the board for display suitable as plaintext in
// console.  This is suitable for demonstration or tests,
// but not for prod.
func (board *Board) Render() string {
	str := ""
	if board.id == 0 {
		str += "### (private) " + board.name + " ###\n\n"
	} else {
		str += "### gno.land/r/boards/" + board.name + " ###\n\n"
	}
	if board.posts.Size() > 0 {
		board.posts.Iterate("", "", func(n *avl.Tree) bool {
			str += "----------------------------------------\n"
			str += n.Value().(*Post).RenderSummary()
			return false
		})
	}
	return str
}

func (board *Board) incGetPostID() PostID {
	board.postsCtr++
	return PostID(board.postsCtr)
}

//----------------------------------------
// Post

// NOTE: a PostID is relative to the board.
type PostID uint64

type Post struct {
	board       *Board
	id          PostID
	creator     std.Address
	title       string // optional
	body        string
	replies     *avl.Tree // Post.id -> *Post
	reposts     *avl.Tree // Board.id -> Post.id
	replyTo     PostID    // original Post.id (if reply or repost)
	repostBoard BoardID   // original Board.id (if repost)
}

func (post *Post) AddReply(creator std.Address, body string) *Post {
	board := post.board
	pid := board.incGetPostID()
	pidkey := strconv.Itoa(int(pid))
	reply := &Post{
		board:   board,
		id:      pid,
		creator: creator,
		body:    body,
		replyTo: post.id,
	}
	post.replies, _ = post.replies.Set(pidkey, pid)
	return reply
}

func (post *Post) AddRepostTo(creator std.Address, title, body string, dst *Board) *Post {
	pid := dst.incGetPostID()
	pidkey := strconv.Itoa(int(pid))
	repost := &Post{
		board:       dst,
		id:          pid,
		creator:     creator,
		title:       title,
		body:        body,
		replyTo:     post.id,
		repostBoard: post.board.id,
	}
	dst.posts, _ = dst.posts.Set(pidkey, repost)
	if !dst.IsPrivate() {
		bidkey := strconv.Itoa(int(dst.id))
		post.reposts, _ = post.reposts.Set(bidkey, pid)
	}
	return repost
}

func (post *Post) Summary() string {
	lines := strings.SplitN(post.body, "\n", 2)
	line := lines[0]
	if len(line) > 80 {
		line = line[:77] + "..."
	}
	return line
}

func (post *Post) RenderSummary() string {
	str := ""
	if post.title != "" {
		str += "# " + post.title + "\n"
		str += "\n"
	}
	str += post.Summary() + "\n\n"
	str += padLeft("("+strconv.Itoa(post.replies.Size())+" replies)", 40) + "\n"
	return str
}

func (post *Post) Render(indent string) string {
	str := "----------------------------------------\n"
	if post.title != "" {
		str += indent + "# " + post.title + "\n"
		str += indent + "\n"
	}
	str += indent + post.body // TODO: indent body lines.
	str += indent + "\n"
	if post.replies.Size() > 0 {
		post.replies.Iterate("", "", func(n *avl.Tree) bool {
			str += "\n"
			str += n.Value().(*Post).Render(indent + "| ")
			return false
		})
	}
	return str
}

//----------------------------------------
// private utility methods
// XXX ensure these cannot be called from public.

func getBoard(bid BoardID) *Board {
	bidkey := strconv.Itoa(int(bid))
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
