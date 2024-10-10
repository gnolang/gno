package boards

import (
	"gno.land/p/demo/boards/post"
	"gno.land/p/demo/boards/post/plugin"
	pluginreputation "gno.land/p/demo/boards/post/plugin/reputation"
	plugintitle "gno.land/p/demo/boards/post/plugin/title"
)

const (
	LevelBoard = iota
	LevelThread
	LevelComment
)

// App is the boards application.
type App struct {
	posts            post.Store
	plugins          *plugin.Registry
	maxCommentsDepth int
	disableComments  bool
	reputationPolicy pluginreputation.Policy
}

// New creates a new boards application.
func New(st post.Store, options ...Option) App {
	app := App{posts: st}
	for _, apply := range options {
		apply(&app)
	}

	app.plugins := plugin.NewRegistry(
		plugintitle.New(st),   // Plugin for boards
		plugintext.New(st),    // Plugin for text based threads
		pluginpoll.New(st),    // Plugin for poll based threads
		plugincomment.New(st), // Plugin for comments to the threads
		pluginreputation.New(
			pluginreputation.UsePolicy(app.reputationPolicy),
			pluginreputation.AllowedPostLevels(post.LevelPost, post.LevelComment),
		),
	)
	return app
}

func (a App) GetPost(path string) (_ *post.Post, found bool) {
	return a.post.Get(path) // NOTE: Allowing this enables us to have generic functions like Lock
}

func (a App) GetBoard(path string) (_ Board, found bool) {
	p, found := a.posts.Get(path)
	if !found || p.Level != LevelBoard {
		return Board{}, false
	}
	return Board{p}, true
}

func (a App) GetThread(path string) (_ Thread, found bool) {
	p, found := a.posts.Get(path)
	if !found || p.Level != LevelThread {
		return Thread{}, false
	}
	return Thread{p}, true
}

func (a App) GetComment(path string) (_ Comment, found bool) {
	p, found := a.posts.Get(path)
	if !found || p.Level != LevelComment {
		return Comment{}, false
	}
	return Comment{p}, true
}

func (a App) CreateBoard(slug, title, description string, tags []string) (Board, error) {}

// Fork forks either a board or a thread by their path.
func (a App) Fork(path, newPath string) error {}

// Lock locks either a board or a thread by their path.
// Once a board is locked new threads to the board and comments to the existing
// threads won't be allowed.
// Once a thread is locked new comments to the thread won't be allowed.
func (a App) Lock(path string) error {}

// ---- TODO: Review the following list of app methods ----- //

func (a App) Boards(c post.Cursor) ([]Board, error) {
}

func (b App) Thread(path string) (Thread, error) {
	return ThreadWithComments(path, nil)
}

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
func (b App) Threads(c post.Cursor) ([]Thread, error)                            {}
func (b App) CreateTextThread(c ThreadTextContent) (Thread, error)               {}
func (b App) CreatePollThread(c ThreadPollContent) (Thread, error)               {}

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
