// plugintext is a content type for representing a tweet, blog post or a thread like Reddit.
package plugintext

const Name = "post-text"

type (
	Plugin struct{}
	Text   struct {
		Title string
		Body  string
		Tags  []string
	}
)

func New() Plugin {
	return Plugin{}
}

func (p Plugin) Content(pst *post.Post) Content {
	return pst.Body[Name].(*Text)
}

func (p Plugin) SetContent(pst *post.Post, t Text) {
	pst.Body[Name] = c
}
