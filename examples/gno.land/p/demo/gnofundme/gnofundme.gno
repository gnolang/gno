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
	goal         uint
	current      uint
	begin        time.Time
	deadline     time.Time
	owner        *User
	beneficiary  *User
	contributors *avl.Tree
}

func NewCampaign(title string, description string, goal uint, begin time.Time, deadline time.Time, owner *User, beneficiary *User) *Campaign {
	return &Campaign{
		title:        title,
		description:  description,
		goal:         goal,
		current:      0,
		begin:        begin,
		deadline:     deadline,
		owner:        owner,
		beneficiary:  beneficiary,
		contributors: avl.NewTree(),
	}
}

func (c Campaign) GetTitle() string {
	return c.title
}

func (c Campaign) GetDescription() string {
	return c.description
}

func (c Campaign) GetGoal() uint {
	return c.goal
}

func (c Campaign) GetCurrent() uint {
	return c.current
}

func (c Campaign) GetBegin() time.Time {
	return c.begin
}

func (c Campaign) GetDeadline() time.Time {
	return c.deadline
}

func (c Campaign) GetOwner() *User {
	return c.owner
}

func (c Campaign) GetBeneficiary() *User {
	return c.beneficiary
}

func (c *Campaign) AddContributor(user *User, amount uint, display bool) {
	c.current += amount
	key := c.contributors.Size()
	contributor := &Contributor{
		user:    user,
		date:    time.Now(),
		amount:  amount,
		display: display,
	}
	c.contributors.Set(strconv.Itoa(key), contributor)
}

func (c Campaign) GetContributors() []*Contributor {
	contributors := make([]*Contributor, 0, c.contributors.Size())
	c.contributors.Iterate("", "", func(key string, value interface{}) bool {
		contributors = append(contributors, value.(*Contributor))
		return false
	})
	return contributors
}
