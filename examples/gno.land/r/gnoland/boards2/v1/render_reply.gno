package boards2

import (
	"strconv"
	"strings"

	"gno.land/p/jeronimoalbi/pager"
	"gno.land/p/moul/md"
	"gno.land/p/nt/mux"
)

func renderReply(res *mux.ResponseWriter, req *mux.Request) {
	name := req.GetVar("board")
	board, found := getBoardByName(name)
	if !found {
		res.Write("Board does not exist: " + name)
		return
	}

	rawID := req.GetVar("thread")
	tID, err := strconv.Atoi(rawID)
	if err != nil {
		res.Write("Invalid thread ID: " + rawID)
		return
	}

	rawID = req.GetVar("reply")
	rID, err := strconv.Atoi(rawID)
	if err != nil {
		res.Write("Invalid reply ID: " + rawID)
		return
	}

	thread, found := board.GetThread(PostID(tID))
	if !found {
		res.Write("Thread does not exist with ID: " + req.GetVar("thread"))
		return
	}

	reply, found := thread.GetReply(PostID(rID))
	if !found {
		res.Write("Reply does not exist with ID: " + rawID)
		return
	}

	// Call render even for hidden replies to display children.
	// Original comment content will be hidden under the hood.
	// See: #3480
	res.Write(renderPostInner(reply))
}

func renderTopLevelReplies(post *Post, path, indent string, levels int) string {
	p, err := pager.New(path, post.replies.Size(), pager.WithPageSize(pageSizeReplies))
	if err != nil {
		panic(err)
	}

	var (
		b              strings.Builder
		commentsIndent = indent + "> "
	)

	render := func(_ string, v any) bool {
		reply := v.(*Post)
		b.WriteString(indent + "\n" + renderPost(reply, "", commentsIndent, levels-1))
		return false
	}

	b.WriteString("\n" + md.HorizontalRule() + "Sort by: ")
	r := parseRealmPath(path)
	if r.Query.Get("order") == "desc" {
		r.Query.Set("order", "asc")
		b.WriteString(md.Link("newest first", r.String()) + "\n")
		post.replies.ReverseIterateByOffset(p.Offset(), p.PageSize(), render)

	} else {
		r.Query.Set("order", "desc")
		b.WriteString(md.Link("oldest first", r.String()) + "\n")
		post.replies.IterateByOffset(p.Offset(), p.PageSize(), render)
	}

	if p.HasPages() {
		b.WriteString(md.HorizontalRule())
		b.WriteString(pager.Picker(p))
	}
	return b.String()
}

func renderSubReplies(post *Post, indent string, levels int) string {
	var (
		b              strings.Builder
		commentsIndent = indent + "> "
	)

	post.replies.Iterate("", "", func(_ string, v any) bool {
		reply := v.(*Post)
		b.WriteString(indent + "\n" + renderPost(reply, "", commentsIndent, levels-1))
		return false
	})
	return b.String()
}
