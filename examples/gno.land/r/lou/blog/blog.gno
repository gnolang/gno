package blog

import (
	"std"
	"strings"

	"gno.land/p/lou/blog"
	// "gno.land/p/moul/md"
	// "gno.land/p/moul/txlink"

	"gno.land/r/sys/users"
)

var (
	myBlog *blog.Blog
)

func myResolver(input string) (string, bool) {
	data, ok := users.ResolveAny(input)
	if !ok || data == nil {
		return "", false
	}
	println(data.RenderLink(""))
	return data.Name(), true
}

func CreateSampleBlog() *blog.Blog {
	myBlog.SetUserResolver(myResolver)
	post1, err := blog.NewPost(
		"ai-future",
		"The Future of AI",
		"Artificial Intelligence is transforming every industry, from healthcare to entertainment.",
		"2024-03-01T10:00:00Z",
		"g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs",
		[]string{"g1abcde12345"},
		[]string{"ai", "technology", "future"},
	)
	if err != nil {
		panic(err)
	}
	comm, err := blog.NewComment(
		"g1abcde12345",
		"Exciting times ahead!",
	)
	if err != nil {
		panic(err)
	}
	post1.AddComment(comm)

	post2, err := blog.NewPost(
		"blockchain-real-world",
		"Blockchain in the Real World",
		"Beyond cryptocurrencies, blockchain is finding uses in supply chains, voting systems, and more.",
		"2023-12-15T08:30:00Z",
		"g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs",
		[]string{"g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs"},
		[]string{"blockchain", "tech", "supplychain"},
	)
	if err != nil {
		panic(err)
	}
	reply, err := blog.NewComment(
		"g1ffzxha57dh0qgv9ma5v393ur0zexfvp6lsjpae",
		"heyyy",
	)
	if err != nil {
		panic(err)
	}
	comm2, err := blog.NewComment(
		"g14x2crt492f84shxa9s7e5ulk7lnf2p3euz7r9n",
		"This tech still needs better UX.",
	)
	if err != nil {
		panic(err)
	}
	comm3, err := blog.NewComment(
		"g1fjnxhv3v5x02j49fyl49epmhjyzscm8e4au94t",
		"Absolutely, usability is key for adoption.",
	)
	if err != nil {
		panic(err)
	}
	comm4, err := blog.NewComment(
		"g14x2crt492f84shxa9s7e5ulk7lnf2p3euz7r9n",
		"... still needs better UX.",
	)
	if err != nil {
		panic(err)
	}
	post2.AddComment(comm2)
	post2.AddComment(comm3)
	post2.AddReply(comm2.ID(), reply)
	post2.AddReply(reply.ID(), comm4)

	post3, err := blog.NewPost(
		"getting-things-done",
		"Getting Things Done in 2025",
		"With distractions everywhere, productivity tools are more essential than ever.",
		"2025-01-10T09:15:00Z",
		"g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs",
		[]string{"g1produser1"},
		[]string{"productivity", "lifehacks", "focus", "ai"},
	)
	if err != nil {
		panic(err)
	}
	myBlog.AddPost(post1)
	myBlog.AddPost(post2)
	myBlog.AddPost(post3)
	return myBlog
}

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
	post.AddComment(comm)
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
	post.AddComment(comm)
}

func CreatePost(_ realm, slug, title, body, publicationDate, authors, tags string) {
	authorsField := strings.Split(authors, " ")
	tagsField := strings.Split(tags, " ")

	post, err := blog.NewPost(
		slug,
		title,
		body,
		publicationDate,
		std.CurrentRealm().Address().String(),
		authorsField,
		tagsField,
	)
	if err != nil {
		panic(err)
	}
	myBlog.AddPost(post)
}

func Start(_ realm) {
	var err error
	myBlog, err = blog.NewBlog("Lou's Blog", "/r/lou/blog")
	if err != nil {
		panic(err)
	}
}

func Test(_ realm) {
	CreateSampleBlog()
}

func Render(path string) string {
	out := ""
	if myBlog != nil {
		out += myBlog.Render(path)
	}
	return out
}
