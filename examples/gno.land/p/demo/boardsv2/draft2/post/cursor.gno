package post

type Cursor struct {
	FromID string
	Count int
}

func NewCursor(fromID string, count int) Cursor {}
