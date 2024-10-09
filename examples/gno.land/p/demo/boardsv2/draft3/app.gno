package boards

import (
	"plugin"

	"golang.org/x/mod/sumdb/storage"
)

const (
	LevelBoard = iota
	LevelThread
	LevelComment
)

type App struct {
	cntx Context
}

func New(st storage.PostStorage, o ...Option) App {
	 // TODO: avl.Tree feels wrong here but maybe get rid of the map anyway
	p := map[string]plugin.Plugin{
		plugintitle.Name: plugintitle.New(st), // content for boards
		plugintext.Name: plugintext.New(st),// content for text based threads
		pluginpoll.Name: pluginpoll.New(st),// content for poll based threads
		plugincomment.Name: plugincomment.New(st),// content for comments to the threads
		pluginreputation.Name: pluginreputation.New(
			pluginreputation.UseTokenBasePolicy(),
			pluginreputation.AllowedPostLevels(post.LevelPost, post.LevelComment),
		),
	}

	return App{
		cntx: Context{
			storage: st,
			plugins: p,
		},
	}
}

func (a App) Board(path string) (Board, error) {
	a.c.Get(level, path func(){})
}

func (a App) Boards(c post.Cursor) ([]Board, error) {

}

func (a App) CreateBoard(c BoardContent) (Board, error) {}

func (b App) Thread(path string) (Thread, error) {
	return ThreadWithComments(path, nil)
}

// Fork forks either a board or a thread by their path.
func (a App) Fork(path, newPath string) error {}

// Lock locks either a board or a thread by their path.
// Once a board is locked new threads to the board and comments to the existing
// threads won't be allowed.
// Once a thread is locked new comments to the thread won't be allowed.
func (a App) Lock(path string) error {}


// ThreadWithComments returns a thread with its comments with the comment depth
// configured with commentDepth for direct and child comments.
// For ex.
// To get a thread with only 10 direct (parent level) comments use:
// - []int{10}
// To get a thread with 10 direct comments and 3 of their child comments use:
// - []int{10, 3}
// You can define configure this for more levels until you reach to value defined
// by MaxCommentDepth.
// By default the configuration is as follows:
// - []int{20, 3}
func (b App) ThreadWithComments(path string, commentDepth []int) (Thread, error) {}
func (b App) Threads(c post.Cursor) ([]Thread, error) {}
func (b App) CreateTextThread(c ThreadTextContent) (Thread, error) {}
func (b App) CreatePollThread(c ThreadPollContent) (Thread, error) {}

// parentPath could be a path to thread (root), or path to any of the
// nested comments.
func (b App) Comments(parentPath string, c Cursor) ([]Comment, error) {}
func (b App) CreateComment(path string, c plugincomment.Content) (Comment, error) {
	post, err := a.c.Plugin(plugincomment.Name).NewPost(c, LevelComment)
	if err != nil {
		return Comment{}, err
	}
	return Comment{Post: post, c: a.c}
}

func (a App) Render(path string) string {}
