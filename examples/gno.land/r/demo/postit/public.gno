package postit

import (
	"std"
	"strconv"
	"time"
)

// User

func CreateUser(username, name string) string {
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
	return user.Username
}

func UpdateBio(bio string) string {
	// TODO: add checks
	caller := std.PrevRealm().Addr()
	user := getUserRefByAddress(caller)
	user.Bio = bio
	return user.Bio
}

func SetAvatarUrl(url string) string {
	// TODO: add checks
	// TODO: later should be replaced with grc721
	caller := std.PrevRealm().Addr()
	user := getUserRefByAddress(caller)
	user.AvatarUrl = url
	return user.AvatarUrl
}

// Post

func CreatePost(body string) uint64 {
	caller := std.PrevRealm().Addr()
	user := getUserRefByAddress(caller)
	post := &Post{
		Id:        postCount,
		Username:  user.Username,
		Body:      body,
		Timestamp: time.Now(),
	}
	posts.Set(strconv.FormatUint(postCount, 10), post)
	user.Posts = append(user.Posts, postCount)
	postCount = postCount + 1
	return post.Id
}

// Test (To be removed)
