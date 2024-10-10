package store

// TODO: Define how cursors should be used alongside stores
type Cursor struct {
	FromID string
	Count  int
}

func NewCursor(fromID string, count int) Cursor {}
