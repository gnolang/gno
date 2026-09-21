package blog

import "errors"

var (
	ErrPostTitleMissing  = errors.New("post title is missing")
	ErrPostSlugMissing   = errors.New("post slug is missing")
	ErrPostBodyMissing   = errors.New("post body is missing")
	ErrPostSlugExists    = errors.New("post with specified slug already exists")
	ErrPostPubDateExists = errors.New("post with specified publication date exists")
	ErrPostTitleExists   = errors.New("post with specified title already exists")
	ErrNoSuchPost        = errors.New("no such post")
)
