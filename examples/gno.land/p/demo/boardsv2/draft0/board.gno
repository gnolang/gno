package boardsv2

type Board struct {
}

func (b *Board) AddPost() error {

}

func (b *Board) GetThread(id string) (post *Post, found bool) {

}

func (b *Board) ListThreads(id string) (post *Post, found bool) {
	threadView.List() // there should be an iterator, pagination
}

func (b *Board) Fork() error {

}

func (b *Board) Lock() error {

}
