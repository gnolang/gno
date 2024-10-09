package commentplugin

type CommentContent struct {
	Body string
}

func (c CommentContent) Render() string {

}

type TitleBasedContent struct{}

type TextContent struct {
	Title string
	Body  string
	Tags  []string
}

func (c TextContent) Render() string {

}

type PollContent struct {
	Question string
	Options  []string
	Votes    []struct {
		Address std.Adress
		Option  string
	}
}
