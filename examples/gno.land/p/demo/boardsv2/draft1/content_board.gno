package boardsv2

const ContentTypeBoard = "boards:board"

var _ Content = (*BoardContent)(nil)

type BoardContent struct {
	Name string
}

func NewBoard() *Post {
	return &Post{
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
