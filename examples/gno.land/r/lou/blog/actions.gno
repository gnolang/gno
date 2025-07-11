package blog

import (
	"std"
	"strings"

	"gno.land/p/lou/blog"
)

func AddCommentToPostById(_ realm, id, comment string) {
	var err error
	comm, err := blog.NewComment(std.PreviousRealm().Address().String(), comment)
	if err != nil {
		panic(err)
	}
	post, err := myBlog.GetPostByID(id)
	if err != nil {
		panic(err)
	}
	if err := post.AddComment(comm); err != nil {
		panic(err)
	}
}

func AddCommentToPostBySlug(_ realm, slug, comment string) {
	var err error
	comm, err := blog.NewComment(std.PreviousRealm().Address().String(), comment)
	if err != nil {
		panic(err)
	}

	post, err := myBlog.GetPostBySlug(slug)
	if err != nil {
		panic(err)
	}
	if err := post.AddComment(comm); err != nil {
		panic(err)
	}
}

func CreatePost(_ realm, slug, title, body, publicationDate, authors, tags string) {
	authorsField := strings.Split(authors, " ")
	tagsField := strings.Split(tags, " ")

	post, err := blog.NewPost(
		slug,
		title,
		body,
		publicationDate,
		std.OriginCaller().String(),
		authorsField,
		tagsField,
	)
	if err != nil {
		panic(err)
	}
	if err := myBlog.AddPost(post); err != nil {
		panic(err)
	}
}
