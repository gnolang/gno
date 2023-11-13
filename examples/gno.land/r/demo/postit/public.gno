package postit

import "std"

// User

func CreateUser(username string) {
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
	}
	users.Set(username, user)
	addressToUsername.Set(caller.String(), username)
}

func UpdateBio(bio string) {
	caller := std.PrevRealm().Addr()
	user := GetUserByAddress(caller)
	user.Bio = bio
}

// Post

// Test (To be removed)
