package boards

const (
	LevelBoard = iota
	LevelThread
	LevelComment
)

type App struct {
	c Context
}


func New(st post.Storage) App {
	a := App{
		st: st,
		plugs: make(map[post.PluginName]plugin.Plugin),
	}

	a.plugs[pluginbody.Name] = pluginbody.New(st) // load global state from st

	return a
}

type Board struct {
	post.Post
	c Context
}

func (b Board) Content() BoardContent {
	return b.c.Plugin(pluginbasiccontent.Name).Content(b.Post)
}

func (b Board) Render() string {

}

type Thread struct {
	post.Post
	c Context
}

func (t Thread) TextContent() ThreadTextContent {

}

func (t Thread) PollContent() ThreadPollContent {}
func (t Thread) Type() ContentType {}

// Comments returns a list of comments sent to the thread.
// The comment slice will be non-nil only when Thread is initiated
// through ThreadWithComments.
func (t Thread) Comments() []Comment {}

type Comment struct {
	post.Post
	c Context
}

func (c Comment) Content() CommentContent {}

func (a App) Board(path string) ([]Board, error) {
	a.c.Get(level, path func(){})
}
func (a App) LockBoard(path string) (error) {}
func (a App) ForkBoard(path string) (error) {}

func (a App) Boards(c post.Cursor) ([]Board, error) {}
func (a App) CreateBoard(c BoardContent) (Board, error) {}

func (b App) Thread(path string) (Thread, error) {
	return ThreadWithComments(path, nil)
}
func (a App) LockThread(path string) (error) {}
func (a App) ForkThread(path string) {}

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

type Context struct {
	opts []Option
	st post.Storage
	plugs map[post.PluginName]plugin.Plugin
}

func (c Context) Plugin(n post.PluginName) post.Plugin {

}

func (c Context) Set(p *Post) (updated bool) {
	key := newKey(p.Level, p.Slug())
	return b.posts.Set(key, p)
}

func (c Context) Remove(level int, path string) (_ *Post, removed bool) {
	key := newKey(level, path)
	if v, removed := b.posts.Remove(key); removed {
		return v.(*Post), true
	}
	return nil, false
}

func (c Context) Get(level int, path string, iterator func()) (_ *Post, found bool) {
	key := newKey(level, path)
	if v, found := b.posts.Get(key); found {
		return v.(*Post), true
	}
	return "", false
}
