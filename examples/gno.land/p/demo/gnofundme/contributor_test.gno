package gnofundme

import (
	"testing"
	"time"

	"gno.land/p/demo/uassert"
	"gno.land/r/demo/users"
)

func TestNewContributor(t *testing.T) {
	user := users.GetUserByName("demo")
	amount := int64(500)
	display := true
	date := time.Unix(time.Now().Unix(), 0)
	contributor := NewContributor(user, date, amount, display)

	uassert.Equal(t, contributor.GetUser().Address, user.Address)

	uassert.False(t, contributor.Date.IsZero())

	uassert.Equal(t, contributor.Amount, amount)

	uassert.Equal(t, contributor.GetDisplay(), display)

	json := `{"Address":"g13ek2zz9qurzynzvssyc4sthwppnruhnp0gdz8n","date":"13 Feb 09 23:31 UTC","amount":500,"display":"true"}`

	uassert.Equal(t, contributor.String(), json)
}
