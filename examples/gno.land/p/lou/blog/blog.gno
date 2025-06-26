package blog

import (
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ownable/exts/authorizable"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/md"
)

type Blog struct {
	Authorizable     *authorizable.Authorizable
	title            string
	prefix           string
	PostId           seqid.ID
	Posts            *avl.Tree // id --> *Post
	PostsBySlug      *avl.Tree // slug --> *Post
	PostsByUpdatedAt *avl.Tree // updatedAt --> *Post
	TagsIndex        *avl.Tree // tagName --> int
	TagsSorted       *avl.Tree // "<count>::tag" --> tag
	AuthorsIndex     *avl.Tree // authorAddress --> int
	AuthorsSorted    *avl.Tree // "<count>::author" --> author

	// CustomHeader TO DO
	previewFooter string // additional text, e.g. "via @lou" or txlink calls
	DisableLikes  bool
	UserResolver  *UserResolver
}

func (b Blog) Title() string {
	return b.title
}

func (b Blog) Prefix() string {
	return b.prefix
}

func NewBlog(title, prefix string) (*Blog, error) {
	if prefix == "" {
		return nil, ErrEmptyPrefix
	}
	auth := authorizable.NewAuthorizable()
	return &Blog{
		Authorizable:     auth,
		title:            title,
		prefix:           prefix,
		PostId:           seqid.ID(0),
		Posts:            avl.NewTree(),
		PostsBySlug:      avl.NewTree(),
		PostsByUpdatedAt: avl.NewTree(),
		TagsIndex:        avl.NewTree(),
		TagsSorted:       avl.NewTree(),
		AuthorsIndex:     avl.NewTree(),
		AuthorsSorted:    avl.NewTree(),
		previewFooter:    "",
		DisableLikes:     false,
		UserResolver:     nil,
	}, nil
}

func (b *Blog) AddPost(post *Post) error {
	b.Authorizable.AssertPreviousOnAuthList()
	post.id = b.PostId.Next()
	if _, err := b.GetPostBySlug(post.Slug()); err == nil {
		return ErrPostAlreadyExists
	}
	if err := CheckAddr(post.Publisher()); err != nil {
		return err
	}
	post.DisableLikes = b.DisableLikes
	if b.UserResolver != nil {
		post.UserResolver = b.UserResolver
	}
	post.authors = post.resolveAuthors(post.Authors())
	post.publisher, _ = CheckUser(post.Publisher(), b.UserResolver)

	b.addToIndex(post)
	added := b.Posts.Set(b.PostId.String(), post)
	addedSlug := b.PostsBySlug.Set(post.Slug(), post)
	addedUpdated := b.PostsByUpdatedAt.Set(post.UpdatedAt().String(), post)
	if !added || !addedSlug || !addedUpdated {
		return ErrPostAddFailed
	}
	return nil
}

func (b *Blog) UpdatePostById(id seqid.ID, newPost *Post) error {
	b.Authorizable.AssertPreviousOnAuthList()
	post, err := b.GetPostByID(id.String())
	if err != nil {
		return err
	}
	postBySlug, _ := b.PostsBySlug.Get(post.Slug())
	postByUpdated, _ := b.PostsByUpdatedAt.Get(post.UpdatedAt().String())
	post.UpdatePost(newPost)
	postById.(*Post).UpdatePost(newPost)
	postByUpdated.(*Post).UpdatePost(newPost)
	return nil
}

func (b *Blog) UpdatePostBySlug(slug string, newPost *Post) error {
	b.Authorizable.AssertPreviousOnAuthList()
	post, err := b.GetPostBySlug(slug)
	if err != nil {
		return err
	}
	postById, _ := b.Posts.Get(post.ID())
	postByUpdated, _ := b.PostsByUpdatedAt.Get(post.UpdatedAt().String())
	post.UpdatePost(newPost)
	postById.(*Post).UpdatePost(newPost)
	postByUpdated.(*Post).UpdatePost(newPost)
	return nil
}

func (b *Blog) DeletePostById(id string) error {
	b.Authorizable.AssertPreviousOnAuthList()
	post, err := b.GetPostByID(id)
	if err != nil {
		return err
	}
	return b.DeletePost(post)
}

func (b *Blog) DeletePostBySlug(slug string) error {
	b.Authorizable.AssertPreviousOnAuthList()
	post, err := b.GetPostBySlug(slug)
	if err != nil {
		return err
	}
	return b.DeletePost(post)
}

func (b *Blog) DeletePost(post *Post) error {
	_, removed := b.Posts.Remove(post.ID())
	_, removedSlug := b.PostsBySlug.Remove(post.Slug())
	_, removedUpdated := b.PostsByUpdatedAt.Remove(post.UpdatedAt().String())
	if !removed || !removedSlug || !removedUpdated {
		return ErrDeleteFailed
	}
	return nil
}

func (b Blog) GetPostByID(id string) (*Post, error) {
	post, found := b.Posts.Get(id)
	if !found {
		return nil, ErrPostNotFound
	}
	return post.(*Post), nil
}

func (b Blog) GetPostBySlug(slug string) (*Post, error) {
	post, found := b.PostsBySlug.Get(slug)
	if !found {
		return nil, ErrPostNotFound
	}
	return post.(*Post), nil
}

func (b *Blog) SetDisablePostLikes(disable bool) {
	b.DisableLikes = disable
	b.Posts.Iterate("", "", func(_ string, value any) bool {
		post := value.(*Post)
		post.DisableLikes = disable
		return false
	})
	b.PostsBySlug.Iterate("", "", func(_ string, value any) bool {
		post := value.(*Post)
		post.DisableLikes = disable
		return false
	})
	b.PostsByUpdatedAt.Iterate("", "", func(_ string, value any) bool {
		post := value.(*Post)
		post.DisableLikes = disable
		return false
	})
}

func (b *Blog) SetUserResolver(resolver UserResolver) {
	b.UserResolver = &resolver
}

func (b *Blog) addToIndex(post *Post) {
	for _, tag := range post.Tags() {
		oldCount := 0
		if val, found := b.TagsIndex.Get(tag); found {
			oldCount = val.(int)
			b.TagsSorted.Remove(ufmt.Sprintf("%05d::%s", oldCount, tag))
		}
		newCount := oldCount + 1
		b.TagsIndex.Set(tag, newCount)
		b.TagsSorted.Set(ufmt.Sprintf("%05d::%s", newCount, tag), newCount)
	}
	for _, author := range post.Authors() {
		oldCount := 0
		if val, found := b.AuthorsIndex.Get(author); found {
			oldCount = val.(int)
			b.AuthorsSorted.Remove(ufmt.Sprintf("%05d::%s", oldCount, author))
		}
		newCount := oldCount + 1
		b.AuthorsIndex.Set(author, newCount)
		b.AuthorsSorted.Set(ufmt.Sprintf("%05d::%s", newCount, author), newCount)
	}
}

func (b Blog) Mention(role, recipient string) string {
	if role == "author" {
		return md.Bold(md.Link("@"+recipient, b.Prefix()+":authors/"+recipient))
	}
	if role == "commenter" {
		return md.Bold(md.Link("@"+recipient, b.Prefix()+":commenters/"+recipient))
	}
	if role == "tag" {
		return md.Bold(md.Link("#"+recipient, b.Prefix()+":tags/"+recipient))
	}
	return ""
}

func (b Blog) filterPostsStartEnd(tree *avl.Tree, start, end *time.Time) *avl.Tree {
	filtered := avl.NewTree()
	tree.Iterate("", "", func(k string, v interface{}) bool {
		post := v.(*Post)
		if (start == nil || post.CreatedAt().After(*start)) &&
			(end == nil || post.CreatedAt().Before(*end)) {
			filtered.Set(k, post)
		}
		return false
	})
	return filtered
}

func (b Blog) filterPostsByField(field, value, sort string) (*avl.Tree, bool) {
	recentPosts := avl.NewTree()
	alphaPosts := avl.NewTree()
	updatedPosts := avl.NewTree()

	switch field {
	case "tag", "author":
		b.Posts.ReverseIterate("", "", func(k string, v interface{}) bool {
			post := v.(*Post)
			var match bool
			if field == "tag" {
				match = hasField(post.Tags(), value)
			} else {
				match = hasField(post.Authors(), value)
			}
			if match {
				recentPosts.Set(k, post)
				alphaPosts.Set(post.Slug(), post)
				updatedPosts.Set(post.UpdatedAt().String(), post)
			}
			return false
		})

	case "commenter":
		commenterId := seqid.ID(0)
		b.Posts.ReverseIterate("", "", func(k string, v interface{}) bool {
			post := v.(*Post)
			comments := post.GetCommentsByAuthor(value)
			for _, comment := range comments {
				keyPrefix := commenterId.Next().String() + "::"
				recentPosts.Set(keyPrefix+post.ID(), comment)
				alphaPosts.Set(keyPrefix+post.Slug(), comment)
				updatedPosts.Set(keyPrefix+post.UpdatedAt().String(), comment)
			}
			return false
		})
	}

	switch sort {
	case "alpha":
		return alphaPosts, alphaPosts.Size() > 0
	case "update":
		return updatedPosts, updatedPosts.Size() > 0
	default:
		return recentPosts, recentPosts.Size() > 0
	}
}

func (b Blog) findPostBySlug(value string) (string, bool) {
	var foundKey string
	var found bool
	b.Posts.Iterate("", "", func(k string, v interface{}) bool {
		post := v.(*Post)
		if post.Slug() == value {
			foundKey = k
			found = true
		}
		return found
	})
	return foundKey, found
}
