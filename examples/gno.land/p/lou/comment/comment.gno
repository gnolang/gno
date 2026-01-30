package comment

import (
	"errors"
	"std"

	"gno.land/p/demo/seqid"
)

var (
	ratings = []string{
		"☆☆☆☆☆",
		"★☆☆☆☆",
		"★★☆☆☆",
		"★★★☆☆",
		"★★★★☆",
		"★★★★★",
	}

	ErrInvalidRating = errors.New("rating is invalid, should be from 0-5")
)

type Comment struct {
	rating  uint64
	comment *string
	author  std.Address
	id      seqid.ID
}

func (c Comment) Rating() uint64 {
	return c.rating
}

func (c Comment) Comment() *string {
	return c.comment
}

func (c Comment) Author() std.Address {
	return c.author
}

func (c Comment) ID() seqid.ID {
	return c.id
}

func NewComment(rating uint64, comment string, author std.Address, id seqid.ID) (*Comment, error) {
	if rating > 5 || rating < 0 {
		return nil, ErrInvalidRating
	}
	return &Comment{
		rating:  rating,
		comment: &comment,
		author:  author,
		id:      id,
	}, nil
}

func RenderRating(rating uint64) string {
	if rating > 5 {
		rating = 5
	}
	return ratings[rating]
}
