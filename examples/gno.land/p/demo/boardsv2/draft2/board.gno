package boards

type Board struct {
	post.Post
	c Context
}

func (b Board) Content() BoardContent {
	return b.c.Plugin(pluginbasiccontent.Name).Content(b.Post)
}

func (b Board) Render() string {

}
