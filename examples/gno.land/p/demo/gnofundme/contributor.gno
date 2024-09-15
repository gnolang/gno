package gnofundme

import (
	"strconv"
	"time"
)

type Contributor struct {
	// the user who contributed
	user *User
	// the date of the contribution
	Date time.Time
	// the amount of the contribution
	Amount int64
	// whether the contributor wants to be displayed (Anonymous in case of false)
	display bool
}

func NewContributor(user *User, date time.Time, amount int64, display bool) *Contributor {
	return &Contributor{
		user:    user,
		Date:    date,
		Amount:  amount,
		display: display,
	}
}

func (c Contributor) GetUser() *User {
	return c.user
}

func (c Contributor) GetDisplay() bool {
	return c.display
}

func (c Contributor) String() string {
	return "Contributor{" +
		"Contributor Address: " + c.user.GetAddress().String() +
		", date: " + c.Date.Format(time.RFC822) +
		", amount: " + strconv.Itoa(int(c.Amount)) +
		", display: " + FormatBool(c.display) +
		"}"
}

func FormatBool(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
