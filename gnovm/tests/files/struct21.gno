package main

type SecretProvider func(user, realm_XXX string) string

type BasicAuth struct {
	Realm   string
	Secrets SecretProvider
}

func (a *BasicAuth) CheckAuth() string { return a.Secrets("me", a.Realm) }

func (a *BasicAuth) secretBasic(user, realm_XXX string) string {
	return a.Realm + "-" + user + "-" + realm_XXX
}

func main() {
	b := &BasicAuth{Realm: "test"}
	b.Secrets = b.secretBasic
	s := b.CheckAuth()
	println(s)
}

// Output:
// test-me-test
