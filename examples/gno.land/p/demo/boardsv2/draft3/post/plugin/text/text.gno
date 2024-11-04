// plugintext is a content type for representing a tweet, blog post or a thread like Reddit.
package plugintext

import (
	"gno.land/p/demo/boards/post" // NOTE: Plugins should be at the same level of post package
)

const Name = "post-text"

type (
	Plugin  struct{}
	Content struct {
		Title string
		Body  string
		Tags  []string
	}
)

func New() Plugin {
	return Plugin{}
}

func (p Plugin) Name() string {
	return Name
}

func (p Plugin) Render() string {
	// TODO: Implement render support for text
	return ""
}

func (p Plugin) Content(pst *post.Post) Content {
	return pst.Body[Name].(*Content)
}

func (p Plugin) SetContent(pst *post.Post, c Content) {
	pst.Body[Name] = c
}
