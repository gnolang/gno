package blog

import (
	"chain/runtime/unsafe"
	"strings"

	"gno.land/p/lou/blog"
	"gno.land/p/moul/md"
	"gno.land/p/moul/txlink"
)

func CreatePost(cur realm, slug, title, body, publicationDate, authors, tags string) {
	authorsField := strings.Split(authors, " ")
	tagsField := strings.Split(tags, " ")

	post, err := blog.NewPost(
		slug,
		title,
		body,
		publicationDate,
		unsafe.OriginCaller().String(),
		authorsField,
		tagsField,
	)
	if err != nil {
		panic(err)
	}
	actionLinks := md.Link("Like", txlink.Call("LikePostBySlug", "slug", post.Slug()))
	post.SetPreviewFooter(actionLinks + "\n\n")
	if err := myBlog.AddPost(post, cur); err != nil {
		panic(err)
	}
}

func UpdatePost(cur realm, id, slug, title, body, publicationDate, authors, tags string) {
	var err error
	var post *blog.Post
	authorsField := strings.Split(authors, " ")
	tagsField := strings.Split(tags, " ")

	_, err = myBlog.GetPostById(id)
	if err != nil {
		panic(err)
	}
	post, err = blog.NewPost(
		slug,
		title,
		body,
		publicationDate,
		unsafe.OriginCaller().String(),
		authorsField,
		tagsField,
	)
	if err != nil {
		panic(err)
	}
	if err := myBlog.UpdatePostById(id, post, cur); err != nil {
		panic(err)
	}
}

func UpdatePostBySlug(cur realm, slug, title, body, publicationDate, authors, tags string) {
	var err error
	var post *blog.Post
	authorsField := strings.Split(authors, " ")
	tagsField := strings.Split(tags, " ")

	_, err = myBlog.GetPostBySlug(slug)
	if err != nil {
		panic(err)
	}
	post, err = blog.NewPost(
		slug,
		title,
		body,
		publicationDate,
		unsafe.OriginCaller().String(),
		authorsField,
		tagsField,
	)
	if err != nil {
		panic(err)
	}
	if err := myBlog.UpdatePostBySlug(slug, post, cur); err != nil {
		panic(err)
	}
}

func DeletePost(cur realm, slug string) {
	if err := myBlog.DeletePostBySlug(slug, cur); err != nil {
		panic(err)
	}
}

func ToggleLikePostBySlug(cur realm, slug string) {
	if err := myBlog.LikePostBySlug(slug, cur); err != nil {
		panic(err)
	}
}

func AddCommentToPostBySlug(cur realm, slug, comment string) {
	var err error
	post, err := myBlog.GetPostBySlug(slug)
	if err != nil {
		panic(err)
	}

	comm, err := blog.NewComment(cur.Previous().Address().String(), comment)
	if err != nil {
		panic(err)
	}
	if err := post.AddComment(comm); err != nil {
		panic(err)
	}
}
