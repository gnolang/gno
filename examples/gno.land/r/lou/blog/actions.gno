package blog

import (
	"chain/runtime"
	"strings"

	"gno.land/p/lou/blog"
	"gno.land/p/moul/md"
	"gno.land/p/moul/txlink"
)

func CreatePost(_ realm, slug, title, body, publicationDate, authors, tags string) {
	authorsField := strings.Split(authors, " ")
	tagsField := strings.Split(tags, " ")

	post, err := blog.NewPost(
		slug,
		title,
		body,
		publicationDate,
		runtime.OriginCaller().String(),
		authorsField,
		tagsField,
	)
	if err != nil {
		panic(err)
	}
	actionLinks := md.Link("Like", txlink.Call("LikePostBySlug", "slug", post.Slug()))
	post.SetPreviewFooter(actionLinks + "\n\n")
	if err := myBlog.AddPost(post); err != nil {
		panic(err)
	}
}

func UpdatePost(_ realm, id, slug, title, body, publicationDate, authors, tags string) {
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
		runtime.OriginCaller().String(),
		authorsField,
		tagsField,
	)
	if err != nil {
		panic(err)
	}
	if err := myBlog.UpdatePostById(id, post); err != nil {
		panic(err)
	}
}

func UpdatePostBySlug(_ realm, slug, title, body, publicationDate, authors, tags string) {
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
		runtime.OriginCaller().String(),
		authorsField,
		tagsField,
	)
	if err != nil {
		panic(err)
	}
	if err := myBlog.UpdatePostBySlug(slug, post); err != nil {
		panic(err)
	}
}

func DeletePost(_ realm, slug string) {
	if err := myBlog.DeletePostBySlug(slug); err != nil {
		panic(err)
	}
}

func ToggleLikePostBySlug(_ realm, slug string) {
	if err := myBlog.LikePostBySlug(slug); err != nil {
		panic(err)
	}
}

func AddCommentToPostBySlug(_ realm, slug, comment string) {
	var err error
	post, err := myBlog.GetPostBySlug(slug)
	if err != nil {
		panic(err)
	}

	comm, err := blog.NewComment(runtime.PreviousRealm().Address().String(), comment)
	if err != nil {
		panic(err)
	}
	if err := post.AddComment(comm); err != nil {
		panic(err)
	}
}
