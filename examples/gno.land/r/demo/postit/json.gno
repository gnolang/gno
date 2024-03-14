package postit

import (
	"bytes"
	"strconv"
	"time"
)

// User

func usersToJsonString(users []User) string {
	var b bytes.Buffer

	// Write the starting bracket for the array
	b.WriteString(`[`)

	// Iterate over each user and append its JSON representation
	for i, user := range users {
		b.WriteString(userToJSONString(user))

		// Add a comma if it's not the last element
		if i < len(users)-1 {
			b.WriteString(`,`)
		}
	}

	// Write the closing bracket for the array
	b.WriteString(`]`)

	return b.String()
}

func userToJSONString(user User) string {
	var b bytes.Buffer

	// Write the starting brace
	b.WriteString(`{`)

	// Write Username field
	b.WriteString(`"Username":"`)
	b.WriteString(user.Username)
	b.WriteString(`"`)

	// Write Address field
	b.WriteString(`,"Address":"`)
	b.WriteString(user.Address.String())
	b.WriteString(`"`)

	// Write Name field
	b.WriteString(`,"Name":"`)
	b.WriteString(user.Name)
	b.WriteString(`"`)

	// Write Bio field
	b.WriteString(`,"Bio":"`)
	b.WriteString(user.Bio)
	b.WriteString(`"`)

	// Write Avatar field
	b.WriteString(`,"Avatar":"`)
	b.WriteString(user.Avatar)
	b.WriteString(`"`)

	// Write TotalPost field
	b.WriteString(`,"PostCount":"`)
	b.WriteString(strconv.Itoa(len(user.Posts)))
	b.WriteString(`"`)

	// Write CreatedAt field
	b.WriteString(`,"CreatedAt":"`)
	b.WriteString(user.CreatedAt.Format(time.RFC3339))
	b.WriteString(`"`)

	// Write the closing brace
	b.WriteString(`}`)

	return b.String()
}

// Post
//
// type GetPostResponse struct {
// 	Id        string
// 	Body      string
// 	Timestamp string
// 	user      User
// }

func postsToJSONString(posts []Post) string {
	var b bytes.Buffer

	// Write the starting bracket for the array
	b.WriteString(`[`)

	// Iterate over each post and append its JSON representation
	for i, post := range posts {
		b.WriteString(postToJSONString(post))

		// Add a comma if it's not the last element
		if i < len(posts)-1 {
			b.WriteString(`,`)
		}
	}

	// Write the closing bracket for the array
	b.WriteString(`]`)

	return b.String()
}

func postToJSONString(post Post) string {
	user := *getUserRefByUsername(post.Username)

	var b bytes.Buffer

	// Write the starting brace
	b.WriteString(`{`)

	// Write Id field
	b.WriteString(`"Id":`)
	b.WriteString(strconv.FormatUint(post.Id, 10))

	// Write Username field
	b.WriteString(`,"Username":"`)
	b.WriteString(post.Username)
	b.WriteString(`"`)

	// Write Body field
	b.WriteString(`,"Body":"`)
	b.WriteString(post.Body)
	b.WriteString(`"`)

	// Write Attachment field
	b.WriteString(`,"Attachment":"`)
	b.WriteString(post.Attachment)
	b.WriteString(`"`)

	// Write LikedBy field
	b.WriteString(`,"LikedBy":[`)
	for i, un := range post.LikedBy {
		if i > 0 {
			b.WriteString(`,`)
		}
		b.WriteString(`"`)
		b.WriteString(un)
		b.WriteString(`"`)
	}
	b.WriteString(`]`)

	// Write Timestamp field
	b.WriteString(`,"CreatedAt":"`)
	b.WriteString(post.CreatedAt.Format(time.RFC3339))
	b.WriteString(`"`)

	// Write User field
	b.WriteString(`,"User":{`)
	b.WriteString(`"Username":"`)
	b.WriteString(user.Username)
	b.WriteString(`","Address":"`)
	b.WriteString(user.Address.String())
	b.WriteString(`","Name":"`)
	b.WriteString(user.Name)
	b.WriteString(`","Bio":"`)
	b.WriteString(user.Bio)
	b.WriteString(`","Avatar":"`)
	b.WriteString(user.Avatar)
	b.WriteString(`","CreatedAt":"`)
	b.WriteString(user.CreatedAt.Format(time.RFC3339))
	b.WriteString(`"}`)

	// TODO: Write Comments

	// Write the closing brace
	b.WriteString(`}`)

	return b.String()
}

// Trending

// Skip empty fields
func trendingsToJSONString(trendings []Trending) string {
	var b bytes.Buffer

	// Write the starting bracket for the array
	b.WriteString(`[`)

	// Iterate over each Trending and append its JSON representation
	for i, trending := range trendings {
		if trending.Tag == "" {
			continue
		}

		// Append comma if not the first element
		if i > 0 {
			b.WriteString(",")
		}

		b.WriteString(trendingToJSONString(trending))
	}

	// Append closing bracket for the array
	b.WriteString("]")

	return b.String()
}

func trendingToJSONString(trending Trending) string {
	var b bytes.Buffer

	// Write the starting brace
	b.WriteString(`{`)

	// Write Tag field
	b.WriteString(`"tag":"`)
	b.WriteString(trending.Tag)
	b.WriteString(`"`)

	// Write Count field
	b.WriteString(`,"count":`)
	b.WriteString(strconv.FormatUint(trending.Count, 10))

	// Write the closing brace
	b.WriteString(`}`)

	return b.String()
}
