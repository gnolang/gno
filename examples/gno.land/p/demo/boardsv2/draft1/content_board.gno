package boardsv2

// TODO: Move content types to `boardsv2` API

const ContentTypeBoard = "boards:board"

var _ Content = (*BoardContent)(nil)

type BoardContent struct {
	Name string
}

func NewBoard() *Post {
	return &Post{ // TODO: Use a contructor to be able to use private fields (use options), NewPost
		// ...
		Level:   LevelBoard,
		Content: &BoardContent{
			// ...
		},
	}
}

func (c BoardContent) Type() string {
	return ContentTypeBoard
}

func (c BoardContent) Render() string {
	return ""
}
