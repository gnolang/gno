package gnofundme

import (
	"time"

	"gno.land/p/demo/json"
	"gno.land/p/demo/ufmt"
	pusers "gno.land/p/demo/users"
)

type Contributor struct {
	// the user who contributed
	user *pusers.User
	// the date of the contribution
	Date time.Time
	// the amount of the contribution
	Amount uint64
	// whether the contributor wants to be displayed (Anonymous in case of false)
	display bool
}

func NewContributor(user *pusers.User, date time.Time, amount uint64, display bool) *Contributor {
	return &Contributor{
		user:    user,
		Date:    date,
		Amount:  amount,
		display: display,
	}
}

func (c Contributor) GetUser() *pusers.User {
	return c.user
}

func (c Contributor) GetDisplay() bool {
	return c.display
}

func (c Contributor) String() string {
	node := json.ObjectNode("Contributor", map[string]*json.Node{
		"Address": json.StringNode("Address", c.user.Address.String()),
		"date":    json.StringNode("date", c.Date.Format(time.RFC822)),
		"amount":  json.NumberNode("amount", float64(c.Amount)),
		"display": json.StringNode("display", FormatBool(c.display)),
	})
	value, err := json.Marshal(node)
	if err != nil {
		panic(ufmt.Sprintf("unexpected error: %s", err))
	}

	return string(value)
}

func FormatBool(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
