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

		CreatedAt: time.Now(),
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

func SetAvatar(base64 string) string {
	// TODO: add checks
	// TODO: later should be replaced with grc721
	caller := std.PrevRealm().Addr()
	user := getUserRefByAddress(caller)
	user.Avatar = base64
	return user.Avatar
}

// Post

func CreatePost(body string) string {
	caller := std.PrevRealm().Addr()
	user := getUserRefByAddress(caller)
	post := &Post{
		Id:       postCount,
		Username: user.Username,
		Body:     body,

		CreatedAt: time.Now(),
	}
	strId := strconv.FormatUint(post.Id, 10)
	posts.Set(strId, post)
	feed = append(feed, postCount)
	user.Posts = append(user.Posts, postCount)
	postCount = postCount + 1

	// Update hashtags
	htl := parsehashTags(post.Body)
	for _, ht := range htl {
		if res, ok := hashtags.Get(ht); ok {
			count := res.(uint64)
			hashtags.Set(ht, count+1)
			updateTrending(ht, count+1)
			continue
		}
		hashtags.Set(ht, uint64(1))
		updateTrending(ht, uint64(1))
	}

	return strId
}

// Test (To be removed)
