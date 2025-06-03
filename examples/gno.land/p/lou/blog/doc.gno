/*
Package blog is a simple blogging system that allows users to create, update, delete, and render blog posts.

It supports features like pagination, filtering by tags and authors, and rendering posts in Markdown format. The package is designed to be extensible and can be integrated into larger applications.

Example:

	import "gno.land/p/lou/blog"

	func Render(path string) string {
		// Create a new blog post
		post, err := blog.NewPost("my-first-post", "My First Post", "This is the body of my first post.", "", []string{"author1"}, []string{"tag1", "tag2"})
		if err != nil {
			panic(err)
		}

		return post.Render()
	}
*/
package blog // import "gno.land/p/lou/blog"