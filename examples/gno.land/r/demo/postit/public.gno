package postit

import (
	"std"
	"strconv"
	"time"
)

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

func CreatePost(body string) {
	caller := std.PrevRealm().Addr()
	user := getUserRefByAddress(caller)
	post := &Post{
		Id:        postCount,
		Username:  user.Username,
		Body:      body,
		Timestamp: time.Now(),
	}
	posts.Set(strconv.FormatUint(postCount, 64), post)
	user.Posts = append(user.Posts, postCount)
	postCount = postCount + 1
}

// Test (To be removed)
