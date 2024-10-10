package boards

type Thread struct {
	post.Post
	c Context
}

func (t Thread) TextContent() ThreadTextContent {

}

func (t Thread) PollContent() ThreadPollContent {}
func (t Thread) Type() ContentType              {}

// Comments returns a list of comments sent to the thread.
// The comment slice will be non-nil only when Thread is initiated
// through ThreadWithComments.
func (t Thread) Comments() []Comment {}
