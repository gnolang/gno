package awesomegno

import (
	"errors"
	"std"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
)

type Category struct {
	ID          seqid.ID
	Name        string
	Description string
	CreatedAt   time.Time
	dapps       *avl.Tree
}

type Dapp struct {
	ID          seqid.ID
	Title       string
	Description string
	Author      std.Address
	URL         string
	CreatedAt   time.Time
	categories  *avl.Tree
}

var (
	ErrTitleEmpty            = errors.New("title cannot be empty")
	ErrDescriptionEmpty      = errors.New("description cannot be empty")
	ErrURLCannotBeEmpty      = errors.New("URL cannot be empty")
	ErrNotAuthorized         = errors.New("not authorized: only admins can perform this action")
	ErrInvalidAddress        = errors.New("invalid address")
	ErrAdminAlreadyExists    = errors.New("admin already exists")
	ErrAdminNotFound         = errors.New("admin not found")
	ErrCannotRemoveLastAdmin = errors.New("cannot remove the last admin")
	ErrCategoryExists        = errors.New("category already exists")
	ErrDappNotFound          = errors.New("dapp not found")
	ErrCategoryNotFound      = errors.New("category not found")
	ErrProposalNotFound      = errors.New("proposal not found")
	ErrProposalNotPending    = errors.New("proposal is not pending")
	ErrNoCategoriesProvided  = errors.New("no categories provided")
	ErrCategoryNameEmpty     = errors.New("category name cannot be empty")
)
