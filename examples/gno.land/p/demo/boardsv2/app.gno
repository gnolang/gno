package boardsv2

import (
	"gno.land/demo/p/boardsv2/post"
	commentplugin "gno.land/demo/p/boardsv2/post/plugins/content"
)

type App struct {
	s      Storage
	boards []Board
}

func New(s Storage, o ...Option) App {
	a := App{
		s: Storage,
	}
	return a
}

func (a *App) AddBoard(name, title, description string) (*Board, error) {
	p := post.New(commentplugin.TitleBasedContent{
		Title:       title,
		Description: description,
	})
	post.Add(a.s, name, p)
	return a.GetBoard(name)
}

func (a *App) GetBoard(name string) (board *Board, found bool) {

}

func (a *App) ListBoards() ([]*Board, error) {

}
