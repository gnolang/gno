package awesomegno

import (
	"std"
	"time"
)

type Category struct {
	ID          uint64
	Name        string
	Description string
	CreatedAt   time.Time
	DappIDs     []uint64
}

type Dapp struct {
	ID          uint64
	Title       string
	Description string
	Author      std.Address
	URL         string
	Approved    bool
	CreatedAt   time.Time
	Categories  []string
}

const (
	StatusPending  = "pending"
	StatusApproved = "approved"
	StatusRejected = "rejected"
)

type Proposal struct {
	ID        uint64
	Dapp      Dapp
	Status    string
	Proposer  std.Address
	Reason    string
	CreatedAt time.Time
}
