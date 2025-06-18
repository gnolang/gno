package blog

import (
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/md"
)

type Blog struct {
	Title            string
	Prefix           string
	PostId           seqid.ID
	Posts            *avl.Tree // id --> *Post
	PostsBySlug      *avl.Tree // slug --> *Post
	PostsByUpdatedAt *avl.Tree // updatedAt --> *Post
	TagsIndex        *avl.Tree // tagName --> int
	TagsSorted       *avl.Tree // "<count>::tag" --> tag
	AuthorsIndex     *avl.Tree // authorAddress --> int
	AuthorsSorted    *avl.Tree // "<count>::author" --> author

	// CustomHeader
	PreviewFooter string // additional text, e.g. "via @lou" or txlink calls
	DisableLikes  bool
	UserResolver  *UserResolver
}

func NewBlog(title, prefix string) (*Blog, error) {
	if prefix == "" {
		return nil, ErrEmptyPrefix
	}
	return &Blog{
		Title:            title,
		Prefix:           prefix,
		PostId:           seqid.ID(0),
		Posts:            avl.NewTree(),
		PostsBySlug:      avl.NewTree(),
		PostsByUpdatedAt: avl.NewTree(),
		TagsIndex:        avl.NewTree(),
		TagsSorted:       avl.NewTree(),
		AuthorsIndex:     avl.NewTree(),
		AuthorsSorted:    avl.NewTree(),
		PreviewFooter:    "",
		DisableLikes:     false,
		UserResolver:     nil,
	}, nil
}

func (b *Blog) AddPost(post *Post) error {
	post.ID = b.PostId.Next()
	if _, exists := b.PostsBySlug.Get(post.Slug); exists {
		return ErrPostAlreadyExists
	}
	if _, err := CheckUser(post.Publisher, b.UserResolver); err != nil {
		return err
	}
	post.SetDisableLikes(b.DisableLikes)
	if b.UserResolver != nil {
		post.SetUserResolver(*b.UserResolver)
	}

	b.addToIndex(post)
	added := b.Posts.Set(b.PostId.String(), post)
	addedSlug := b.PostsBySlug.Set(post.Slug, post)
	addedUpdated := b.PostsByUpdatedAt.Set(post.UpdatedAt.String(), post)
	if !added || !addedSlug || !addedUpdated {
		return ErrPostAddFailed
	}
	return nil
}

func (b *Blog) UpdatePostById(id seqid.ID, newPost *Post) error {
	existingPost, found := b.Posts.Get(id.String())
	if !found {
		return ErrPostNotFound
	}
	post := existingPost.(*Post)
	postBySlug, _ := b.PostsBySlug.Get(post.Slug)
	postByUpdated, _ := b.PostsByUpdatedAt.Get(post.UpdatedAt.String())
	UpdatePost(post, newPost)
	UpdatePost(postBySlug.(*Post), newPost)
	UpdatePost(postByUpdated.(*Post), newPost)
	return nil
}

func (b *Blog) UpdatePostBySlug(slug string, newPost *Post) error {
	existingPost, found := b.PostsBySlug.Get(slug)
	if !found {
		return ErrPostNotFound
	}
	post := existingPost.(*Post)
	postById, _ := b.Posts.Get(post.ID.String())
	postByUpdated, _ := b.PostsByUpdatedAt.Get(post.UpdatedAt.String())
	UpdatePost(post, newPost)
	UpdatePost(postById.(*Post), newPost)
	UpdatePost(postByUpdated.(*Post), newPost)
	return nil
}

func UpdatePost(existingPost, newPost *Post) {
	existingPost.Slug = newPost.Slug
	existingPost.Title = newPost.Title
	existingPost.Body = newPost.Body
	existingPost.Authors = newPost.Authors
	existingPost.Tags = newPost.Tags
	existingPost.CommentId = newPost.CommentId
	existingPost.Comments = newPost.Comments
	existingPost.Likes = newPost.Likes
	existingPost.DisableLikes = newPost.DisableLikes
	existingPost.PreviewFooter = newPost.PreviewFooter
	existingPost.UpdatedAt = time.Now()
}

func (b *Blog) DeletePostById(id string) error {
	post, found := b.Posts.Get(id)
	if !found {
		return ErrPostNotFound
	}
	return b.DeletePost(post.(*Post))
}

func (b *Blog) DeletePostBySlug(slug string) error {
	post, found := b.PostsBySlug.Get(slug)
	if !found {
		return ErrPostNotFound
	}
	return b.DeletePost(post.(*Post))
}

func (b *Blog) DeletePost(post *Post) error {
	_, removed := b.Posts.Remove(post.ID.String())
	_, removedSlug := b.PostsBySlug.Remove(post.Slug)
	_, removedUpdated := b.PostsByUpdatedAt.Remove(post.UpdatedAt.String())
	if !removed || !removedSlug || !removedUpdated {
		return ErrDeleteFailed
	}
	return nil
}

func (b *Blog) SetDisablePostLikes(disable bool) {
	b.DisableLikes = disable
	b.Posts.ReverseIterate("", "", func(_ string, value any) bool {
		post := value.(*Post)
		post.SetDisableLikes(disable)
		return false
	})
	b.PostsBySlug.ReverseIterate("", "", func(_ string, value any) bool {
		post := value.(*Post)
		post.SetDisableLikes(disable)
		return false
	})
	b.PostsByUpdatedAt.ReverseIterate("", "", func(_ string, value any) bool {
		post := value.(*Post)
		post.SetDisableLikes(disable)
		return false
	})
}

func (b *Blog) SetUserResolver(resolver UserResolver) {
	b.UserResolver = &resolver
}

func (b Blog) Mention(role, recipient string) string {
	if role == "author" {
		return md.Bold(md.Link("@"+recipient, b.Prefix+":authors/"+recipient))
	}
	if role == "commenter" {
		return md.Bold(md.Link("@"+recipient, b.Prefix+":commenters/"+recipient))
	}
	if role == "tag" {
		return md.Bold(md.Link("#"+recipient, b.Prefix+":tags/"+recipient))
	}
	return ""
}

func (b Blog) determinePager(options RenderOptions, field, value string) (*pager.Pager, bool) {
	if value == "" {
		if options.IsAlphabetical && field == "author" {
			return pager.NewPager(b.AuthorsIndex, 12, !options.Ascending), true
		} else if options.IsCommon && field == "author" {
			return pager.NewPager(b.AuthorsSorted, 12, options.Ascending), true
		} else if options.IsAlphabetical && field == "tag" {
			return pager.NewPager(b.TagsIndex, 12, !options.Ascending), true
		} else if options.IsCommon && field == "tag" {
			return pager.NewPager(b.TagsSorted, 12, options.Ascending), true
		} else if field == "author" {
			return pager.NewPager(b.AuthorsIndex, 12, options.Ascending), true
		} else if field == "tag" {
			return pager.NewPager(b.TagsIndex, 12, options.Ascending), true
		}
	}
	pageSize := 9
	if !options.IsGrid {
		pageSize = 3
	}

	filteredPosts, exists := b.filterPostsByField(field, value, options.Sort)
	if options.StartTime != nil || options.EndTime != nil {
		filteredPosts = b.filterPostsStartEnd(filteredPosts, options.StartTime, options.EndTime)
	}
	if options.Sort == "alpha" {
		return pager.NewPager(filteredPosts, pageSize, !options.Ascending), exists
	}
	return pager.NewPager(filteredPosts, pageSize, options.Ascending), exists
}

func (b Blog) filterPostsStartEnd(tree *avl.Tree, start, end *time.Time) *avl.Tree {
	filtered := avl.NewTree()
	tree.Iterate("", "", func(k string, v interface{}) bool {
		post := v.(*Post)
		if (start == nil || post.CreatedAt.After(*start)) &&
			(end == nil || post.CreatedAt.Before(*end)) {
			filtered.Set(k, post)
		}
		return false
	})
	return filtered
}

func (b Blog) filterPostsByField(field, value, sort string) (*avl.Tree, bool) {
	commenterId := seqid.ID(0)
	recentPosts := avl.NewTree()
	alphaPosts := avl.NewTree()
	updatedPosts := avl.NewTree()

	switch field {
	case "tag", "author":
		b.Posts.ReverseIterate("", "", func(k string, v interface{}) bool {
			post := v.(*Post)
			var match bool
			if field == "tag" {
				match = hasField(post.Tags, value)
			} else {
				match = hasField(post.Authors, value)
			}
			if match {
				recentPosts.Set(k, post)
				alphaPosts.Set(post.Slug, post)
				updatedPosts.Set(post.UpdatedAt.String(), post)
			}
			return false
		})

	case "commenter":
		b.Posts.ReverseIterate("", "", func(k string, v interface{}) bool {
			post := v.(*Post)
			comments := post.GetCommentsByAuthor(value)
			for _, comment := range comments {
				commenterId.Next()
				keyPrefix := commenterId.String() + "::"
				recentPosts.Set(keyPrefix+post.ID.String(), comment)
				alphaPosts.Set(keyPrefix+post.Slug, comment)
				updatedPosts.Set(keyPrefix+post.UpdatedAt.String(), comment)
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
		if post.Slug == value {
			foundKey = k
			found = true
		}
		return found
	})
	return foundKey, found
}

func (b *Blog) addToIndex(post *Post) {
	for _, tag := range post.Tags {
		oldCount := 0
		if val, found := b.TagsIndex.Get(tag); found {
			oldCount = val.(int)
			b.TagsSorted.Remove(ufmt.Sprintf("%05d::%s", oldCount, tag))
		}
		newCount := oldCount + 1
		b.TagsIndex.Set(tag, newCount)
		b.TagsSorted.Set(ufmt.Sprintf("%05d::%s", newCount, tag), newCount)
	}
	for _, author := range post.Authors {
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
