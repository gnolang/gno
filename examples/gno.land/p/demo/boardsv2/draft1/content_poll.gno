package boardsv2

const ContentTypePoll = "boards:poll"

var _ Content = (*PollContent)(nil)

type PollContent struct {
	Question string
	Options  []string
	Votes    []struct {
		Address std.Adress
		Option  string
	}
}

func NewPoll() *Post {
	return &Post{
		// ...
		Level:   LevelPost,
		Content: &PollContent{
			// ...
		},
	}
}

func (c PollContent) Type() string {
	return ContentTypePoll
}

func (c PollContent) Render() string {
	return ""
}
