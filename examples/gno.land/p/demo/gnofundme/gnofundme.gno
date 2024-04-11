package gnofundme

import (
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/avl"
)

type Campaign struct {
	title        string
	description  string
	goal         uint64
	current      uint64
	deadline     time.Time
	owner        *User
	beneficiary  *User
	contributors *avl.Tree
}

type User struct {
	name      string
	email     string
	address   std.Address
	campaigns []*Campaign
}

type Contributor struct {
	user    *User
	date    time.Time
	amount  uint64
	display bool
}

func NewUser(name string, email string, address std.Address) *User {
	return &User{
		name:    name,
		email:   email,
		address: address,
	}
}

func NewCampaign(title string, description string, goal uint64, deadline time.Time, owner *User, beneficiary *User) *Campaign {
	return &Campaign{
		title:        title,
		description:  description,
		goal:         goal,
		current:      0,
		deadline:     deadline,
		owner:        owner,
		beneficiary:  beneficiary,
		contributors: avl.NewTree(),
	}
}

func (c Campaign) AddContributor(user *User, amount uint64, display bool) {
	c.current += amount
	key := c.contributors.Size()
	c.contributors.Set(strconv.Itoa(key), &Contributor{
		user:    user,
		date:    time.Now(),
		amount:  amount,
		display: display,
	})
}

func (u User) GetName() string {
	return u.name
}

func (u User) GetEmail() string {
	return u.email
}

func (u User) GetAddress() std.Address {
	return u.address
}