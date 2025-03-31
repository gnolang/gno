package gh

import "errors"

// Account represents a GitHub user account or organization.
type Account struct {
	id   string
	name string
	kind string
}

func (a Account) ID() string   { return a.id }
func (a Account) Name() string { return a.name }
func (a Account) Kind() string { return a.kind }
func (a Account) URL() string  { return "https://github.com/" + a.id }
func (a Account) IsUser() bool { return a.kind == UserAccount }
func (a Account) IsOrg() bool  { return a.kind == OrgAccount }

// TODO: func (a Account) RepoByID() Repo ...

func (a Account) Validate() error {
	if a.id == "" {
		return errors.New("empty id")
	}
	if a.kind == "" {
		return errors.New("empty kind")
	}
	if a.name == "" {
		return errors.New("empty name")
	}
	// TODO: validate
	return nil
}

func (a Account) String() string {
	// XXX: better idea?
	return a.URL()
}

const (
	UserAccount string = "user"
	OrgAccount  string = "org"
)

func AccountByID(id string) *Account {
	res, ok := accounts.Get(id)
	if !ok {
		return nil
	}

	return res.(*Account)
}
