package custom_resource

import (
	"errors"

	"gno.land/p/nt/ufmt"
	"gno.land/p/samcrew/daokit"
)

const ActionNewPostKind = "gno.land/r/daodemo/custom_resource/blog.NewPost"

type ActionNewPost struct {
	Title   string
	Content string
}

func (m *ActionNewPost) String() string {
	return ufmt.Sprintf("## %s\n\n%s", m.Title, m.Content)
}

func (m *ActionNewPost) Type() string {
	return ActionNewPostKind
}

func NewPostAction(title, content string) daokit.Action {
	// def: daoKit.NewAction(kind: String, payload: interface{})
	return daokit.NewAction(ActionNewPostKind, &ActionNewPost{
		Title:   title,
		Content: content,
	})
}

func NewPostHandler(blog *Blog) daokit.ActionHandler {
	// def: daoKit.NewActionHandler(kind: String, payload: func(interface{}))
	return daokit.NewActionHandler(ActionNewPostKind, func(payload interface{}) {
		action, ok := payload.(*ActionNewPost)
		if !ok {
			panic(errors.New("invalid action type"))
		}
		blog.NewPost(action.Title, action.Content)
	})
}
