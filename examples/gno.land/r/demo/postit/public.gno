package postit

import "std"

// User

func CreateUser(username, name string) {
	caller := std.PrevRealm().Addr()
	if username == "" {
		panic("username cannot be empty")
	}
	if _, ok := addressToUsername.Get(caller.String()); ok {
		panic("user with address already exists")
	}
	if _, ok := users.Get(username); ok {
		panic("username already taken")
	}
	user := &User{
		Address:  caller,
		Username: username,
		Name:     name,
	}
	users.Set(username, user)
	addressToUsername.Set(caller.String(), username)
	userCount = userCount + 1
}

func UpdateBio(bio string) {
	caller := std.PrevRealm().Addr()
	user := getUserRefByAddress(caller)
	user.Bio = bio
}

// Post

// Test (To be removed)
