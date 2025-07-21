package gh

import "errors"

// Repo represents a GitHub repository.
type Repo struct {
	id        string
	owner     *Account
	name      string
	isPrivate bool
	isFork    bool
}

func (r Repo) ID() string      { return r.id }
func (r Repo) Name() string    { return r.name }
func (r Repo) Owner() *Account { return r.owner }
func (r Repo) IsPrivate() bool { return r.isPrivate }
func (r Repo) IsFork() bool    { return r.isFork }
func (r Repo) URL() string     { return r.owner.URL() + "/" + r.name }

func (r Repo) String() string {
	// XXX: better idea?
	return r.URL()
}

func (r Repo) Validate() error {
	if r.id == "" {
		return errors.New("id is empty")
	}
	if r.name == "" {
		return errors.New("name is empty")
	}
	if r.owner == nil {
		return errors.New("owner is nil")
	}
	return nil
}

func RepoByID(id string) *Repo {
	res, ok := repos.Get(id)
	if !ok {
		return nil
	}

	return res.(*Repo)
}
