package gnofundme

import "time"

type Contributor struct {
	user    *User
	date    time.Time
	amount  uint
	display bool
}

func (c Contributor) GetUser() *User {
	return c.user
}

func (c Contributor) GetDate() time.Time {
	return c.date
}

func (c Contributor) GetAmount() uint {
	return c.amount
}

func (c Contributor) GetDisplay() bool {
	return c.display
}
