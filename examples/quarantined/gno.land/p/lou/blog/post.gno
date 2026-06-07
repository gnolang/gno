package blog

import (
	"strconv"
	"strings"
	"time"

	"gno.land/p/moul/md"
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/seqid/v0"
	"gno.land/p/nt/ufmt/v0"
)

type Post struct {
	id           seqid.ID
	slug         string
	title        string
	body         string
	createdAt    time.Time
	updatedAt    time.Time
	publishedAt  time.Time
	tags         []string
	authors      []string
	publisher    string
	likes        *avl.Tree // addr --> bool
	CommentsId   seqid.ID
	Comments     *avl.Tree // id --> *Comment
	CommentsSize int

	PreviewFooter   string // additional text, e.g. "via @lou" or txlink calls
	DisableLikes    bool
	DisableComments bool
	UserResolver    UserResolver
}

func (p Post) ID() string {
	return p.id.String()
}

func (p Post) Slug() string {
	return p.slug
}

func (p Post) Title() string {
	return p.title
}

func (p Post) Body() string {
	return p.body
}

func (p Post) CreatedAt() time.Time {
	return p.createdAt
}

func (p Post) UpdatedAt() time.Time {
	return p.updatedAt
}

func (p Post) PublishedAt() time.Time {
	return p.publishedAt
}

func (p Post) Tags() []string {
	return p.tags
}

func (p Post) Authors() []string {
	return p.authors
}

func (p Post) Publisher() string {
	return p.publisher
}

func (p Post) Likes() int {
	return p.likes.Size()
}

func NewPost(slug, title, body, publicationDate, caller string, authors, tags []string) (*Post, error) {
	newSlug := handleSlug(slug, title)
	if title == "" {
		return nil, ErrEmptyTitle
	}
	if body == "" {
		return nil, ErrEmptyBody
	}

	var err error
	publishedAt := time.Now()
	if publicationDate != "" {
		publishedAt, err = time.Parse(time.RFC3339, publicationDate)
		if err != nil {
			return nil, err
		}
	}

	return &Post{
		slug:            newSlug,
		title:           title,
		body:            body,
		createdAt:       publishedAt,
		updatedAt:       publishedAt,
		publishedAt:     time.Now(),
		tags:            tags,
		authors:         authors,
		publisher:       caller,
		likes:           avl.NewTree(),
		CommentsId:      seqid.ID(0),
		Comments:        avl.NewTree(),
		CommentsSize:    0,
		PreviewFooter:   "",
		DisableLikes:    false,
		DisableComments: false,
		UserResolver:    nil,
	}, nil
}

func (p *Post) UpdatePost(newPost *Post) {
	if p.UserResolver != nil {
		p.authors = p.resolveAuthors(p.Authors())
		p.publisher, _ = CheckUser(p.Publisher(), p.UserResolver)
	} else {
		p.authors = newPost.authors
	}

	p.slug = newPost.slug
	p.title = newPost.title
	p.body = newPost.body
	p.tags = newPost.tags
	p.CommentsId = newPost.CommentsId
	p.Comments = newPost.Comments
	p.likes = newPost.likes
	p.DisableLikes = newPost.DisableLikes
	p.DisableComments = newPost.DisableComments
	p.PreviewFooter = newPost.PreviewFooter
	p.updatedAt = time.Now()
}

func (p *Post) AddComment(comment *Comment) error {
	if p == nil {
		return ErrInvalidPost
	}
	user, err := CheckUser(comment.Author(), p.UserResolver)
	if err != nil {
		return err
	}
	comment.author = user

	p.CommentsId.Next()
	p.CommentsSize++
	comment.SetID(p.CommentsId)
	p.Comments.Set(p.CommentsId.String(), comment)
	return nil
}

func (p *Post) AddReply(parentID string, reply *Comment) error {
	if p == nil {
		return ErrInvalidPost
	}
	user, err := CheckUser(reply.Author(), p.UserResolver)
	if err != nil {
		return err
	}
	reply.author = user

	parent, err := p.GetCommentByID(parentID)
	if err != nil {
		return err
	}

	p.CommentsId.Next()
	p.CommentsSize++
	reply.SetID(p.CommentsId)
	parent.Replies.Set(p.CommentsId.String(), reply)
	return nil
}

func (p *Post) EditCommentByID(id string, content string) error {
	comment, err := p.GetCommentByID(id)
	if err != nil {
		return err
	}
	if err := comment.Edit(content); err != nil {
		return err
	}
	return nil
}

func (p *Post) DeleteCommentById(id string) error {
	if _, err := p.GetCommentByID(id); err != nil {
		return err
	}
	_, removed := p.Comments.Remove(id)
	if !removed {
		return ErrCommentDeleteFailed
	}
	p.CommentsSize--
	return nil
}

func (p *Post) PinCommentById(id string) error {
	comment, err := p.GetCommentByID(id)
	if err != nil {
		return ErrCommentNotFound
	}
	comment.Pin()
	return nil
}

func (p *Post) LikePost(addr string) error {
	if p.DisableLikes {
		return ErrLikesDisabled
	}
	if _, exists := p.likes.Get(addr); exists {
		return ErrAlreadyLiked
	}
	p.likes.Set(addr, true)
	return nil
}

func (p *Post) UnlikePost(addr string) error {
	if p.DisableLikes {
		return ErrLikesDisabled
	}
	if _, exists := p.likes.Get(addr); !exists {
		return ErrNotLiked
	}
	if _, removed := p.likes.Remove(addr); !removed {
		return ErrUnlikeFailed
	}
	return nil
}

func (p *Post) SetPreviewFooter(footer string) {
	p.PreviewFooter = footer
}

func (p Post) GetCommentByID(id string) (*Comment, error) {
	comment, found := p.Comments.Get(id)
	if found {
		return comment.(*Comment), nil
	}

	var result *Comment
	p.Comments.ReverseIterate("", "", func(_ string, v interface{}) bool {
		c := v.(*Comment)
		if found := searchInReplies(c.Replies, id, &result); found {
			return true
		}
		return false
	})
	if result == nil {
		return nil, ErrCommentNotFound
	}
	return result, nil
}

func (p Post) GetCommentsByAuthor(author string) []*Comment {
	var comments []*Comment = nil
	p.Comments.ReverseIterate("", "", func(_ string, comment interface{}) bool {
		c := comment.(*Comment)
		if c.Author() == author {
			comments = append(comments, c)
		}
		comments = append(comments, p.getReplies(c.Replies, author)...)
		return false
	})
	return comments
}

func (p Post) RenderPreview(gridMode bool, timeFmt, prefix string) string {
	out := md.H2(md.Link(p.Title(), prefix+":posts/"+p.Slug())) + "\n\n"
	out += md.H5(md.Italic("/"+p.Slug())) + "\n\n"
	if p.UpdatedAt() != p.CreatedAt() {
		out += formatTime(p.UpdatedAt(), timeFmt) + " (updated)\n\n"
	} else {
		out += formatTime(p.CreatedAt(), timeFmt) + "\n\n"
	}
	if !gridMode {
		out += md.Bold("author(s):") + " " + strings.Join(p.Authors(), ", ") + "\n\n"
	}
	if len(p.Tags()) > 0 && p.Tags()[0] != "" {
		out += md.Bold("tags:") + " `" + strings.Join(p.Tags(), "`, `") + "`\n\n"
	}
	if !p.DisableComments {
		out += md.Bold(md.Link("comments ("+strconv.Itoa(p.CommentsSize), prefix+":posts/"+p.Slug()+"#comments")) + ") "
	}
	if !p.DisableLikes {
		out += ufmt.Sprintf("| ❤️ (%d)\n\n", p.Likes())
	}
	if p.PreviewFooter != "" {
		out += p.PreviewFooter + "\n\n"
	}
	out += md.HorizontalRule()
	return out
}

func (p Post) RenderPost(prefix string) string {
	resolvedAuthors := p.resolveAuthors(p.Authors())
	out := md.H1(p.Title()) + "\n\n"
	if p.Comments.Size() > 0 {
		out += md.Link(strconv.Itoa(p.CommentsSize)+" Comment(s)", "#comments") + "\n\n"
	}
	out += md.Italic("Author(s):") + " " + renderAuthorLinks(prefix, resolvedAuthors) + "\n\n"
	out += p.Body() + "\n\n"
	out += md.HorizontalRule()
	out += md.Bold("Created on:") + " " + formatTime(p.CreatedAt(), "full") + "\n\n"
	out += md.Bold("Published on:") + " " + formatTime(p.PublishedAt(), "full") + "\n\n"
	if p.UpdatedAt() != p.CreatedAt() {
		out += md.Bold("Last updated:") + " " + formatTime(p.UpdatedAt(), "full") + "\n\n"
	}
	out += md.Bold("Publisher:") + " " + md.Link(p.Publisher(), "/u/"+p.Publisher()) + "\n\n"
	out += md.Bold("Tags:") + " " + renderTagLinks(prefix, p.Tags()) + "\n\n"
	if !p.DisableLikes {
		out += ufmt.Sprintf("❤️ %d\n\n", p.Likes())
	}
	if !p.DisableComments {
		out += md.HorizontalRule()
		out += p.RenderComments(prefix)
	}
	return out
}

func (p Post) RenderComments(prefix string) string {
	out := md.H3("Comments ("+strconv.Itoa(p.CommentsSize)+")") + "\n\n"
	if p.Comments.Size() == 0 {
		out += "No comments yet.\n\n"
		return out
	}
	p.Comments.ReverseIterate("", "", func(_ string, v interface{}) bool {
		comment := v.(*Comment)
		if comment.Pinned() {
			out += comment.Render(prefix, 0, p.UserResolver) + "\n\n"
		}
		return false
	})
	p.Comments.ReverseIterate("", "", func(_ string, v interface{}) bool {
		comment := v.(*Comment)
		if !comment.Pinned() {
			out += comment.Render(prefix, 0, p.UserResolver) + "\n\n"
		}
		return false
	})
	return out
}

func (p Post) getReplies(replies *avl.Tree, author string) []*Comment {
	var comments []*Comment
	if replies == nil {
		return comments
	}
	replies.ReverseIterate("", "", func(_ string, value any) bool {
		comment := value.(*Comment)
		if comment.Author() == author {
			comments = append(comments, comment)
		}
		if comment.Replies.Size() > 0 {
			comments = append(comments, p.getReplies(comment.Replies, author)...)
		}
		return false
	})
	return comments
}

func (p Post) resolveAuthors(authors []string) []string {
	var resolvedAuthors []string
	for _, author := range authors {
		user, err := CheckUser(author, p.UserResolver)
		if err == nil {
			resolvedAuthors = append(resolvedAuthors, user)
		} else {
			resolvedAuthors = append(resolvedAuthors, author)
		}
	}
	return resolvedAuthors
}

func renderTagLinks(prefix string, tags []string) string {
	return renderLinks(prefix, "tags", tags, "#")
}

func renderAuthorLinks(prefix string, authors []string) string {
	return renderLinks(prefix, "authors", authors, "")
}
