package main

type SecretProvider func(user, realm_XXX string) string

type BasicAuth struct {
	Realm   string
	Secrets SecretProvider
}

func (a *BasicAuth) CheckAuth() string { return a.Secrets("me", a.Realm) }

func secretBasic(user, realm_XXX string) string { return user + "-" + realm_XXX }

func main() {
	b := &BasicAuth{"test", secretBasic}
	s := b.CheckAuth()
	println(s)
}

// Output:
// me-test
