// plugintext is a content type for representing organizations, categories or sections.
package plugintitle

const Name = "post-title"

type (
	Plugin struct{}

	Content struct {
		Title       string
		Description string
		Tags        []string
	}
)

func New() Plugin {
	return Plugin{}
}

func (p Plugin) Name() string {
	return Name
}

func (p Plugin) Render() string {
	// TODO: Implement render support for title
	return ""
}

func (p Plugin) Content(pst *post.Post) Content {
	return pst.Body[Name].(*Content)
}

func (p Plugin) SetContent(pst *post.Post, c Content) {
	pst.Body[Name] = c
}
