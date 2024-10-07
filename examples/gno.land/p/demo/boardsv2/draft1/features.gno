package boardsv2

func AddBoard(s PostStore, slug string /* ... */) (path string, _ error) {
	// TODO: Finish implementation

	return slug, nil
}

// NOTE: Define a pattern to add functionality to posts by type (AddComment, AddThread, AddPoll, Repost, Upvote, ...)
// NOTE: Maybe though functions that assert the right arguments
func AddComment(s PostStore, parentPath string, creator std.Address, message string) (path string, _ error) {
	// Try to get parent as a post or a comment, otherwise parent doesn't support comments
	p, found := b.Get(LevelPost, parentPath)
	if !found {
		p, found = b.Get(LevelComment, parentPath)
		if !found {
			return "", errors.New("parent post or comment not found: " + parentPath)
		}
	}

	comment := NewComment(p /* ... */)

	// TODO: Finish implementation

	path = parentPath + "/" + comment.ID
	return path, nil
}

// NOTE: Arguments could potentially be many, consider variadic + sane defaults (?)
func AddThread(s PostStore, parentPath, slug string, creator std.Address /* ... */) (path string, _ error) {
	p, found := b.Get(LevelPost, parentPath)
	if !found {
		return "", errors.New("parent post not found: " + parentPath)
	}

	post := NewPost(p, slug /* ... */)

	// TODO: Finish implementation

	path = parentPath + "/" + post.ID
	return path, nil
}

// ----- Other features -----
// type VotesStore interface {
// 	/*...*/
// }
//
// func Upvote(s VotesStore /* ... */)   {}
// func DownVote(s VotesStore /* ... */) {}
