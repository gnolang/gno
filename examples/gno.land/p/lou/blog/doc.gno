/*
Package blog is a simple blogging system that allows users to create, update, delete, and render blog posts.

It supports features like pagination, filtering by tags and authors, and rendering posts in Markdown format. The package is designed to be extensible and can be integrated into larger applications.

Example:

	import "gno.land/p/lou/blog"

	func init() {
		myBlog, _ = blog.NewBlog(
			"Personal Blog",
			"g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs",
			blog.WithUserResolver(myResolver),
		)
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

	func Render(path string) string {
		return myBlog.Render(path)
	}

	func myResolver(input string) (string, bool) {
		data, ok := users.ResolveAny(input)
		if !ok || data == nil {
			return "", false
		}
		return data.Name(), true
	}
*/
package blog // import "gno.land/p/lou/blog"
