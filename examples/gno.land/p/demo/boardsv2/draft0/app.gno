package boardsv2

import (
	"gno.land/demo/p/boardsv2/post"
	contentplugin "gno.land/demo/p/boardsv2/post/plugins/content"
)

// type Rating struct{}
//
// var ratingIndex = avl.Tree{}
// app.AddBoardHook(func (changeType int, change ChangeSet) {
// 	if changeType == 0 {
//    ratingIndex.Set("...", )
//  }
// })

type App struct {
	st     Storage
	boards []Board
}

func New(s Storage, o ...Option) App {
	a := App{
		st: Storage,
	}
	// Define the rule for a spesific view.
	boardsView := view.New(view.Filter{
		Level: 0, // this will give me the list of the boards.
	})

	return a
}

func (a *App) AddBoard(name, title, description string) (*Board, error) {
	p := post.New(contentplugin.TitleBasedContent{
		Title:       title,
		Description: description,
	})

	// I want to create a query for listing threads under this new board.
	threadView := view.New(view.Filter{
		Level: 1,
		SlugPrefix: name,
	})
	userActivityView := view.New(view.Filter{
		LevelGte: 2,
		By: func(content Content) []View {
			c.Author // by account address
		}
	})

	if err := post.Add(a.st, name, p); err != nil {
		nil, err
	}
	return a.GetBoard(name), nil
}

func (a *App) GetBoard(name string) (board *Board, found bool) {

}

func (a *App) ListBoards() ([]*Board, error) {

}
