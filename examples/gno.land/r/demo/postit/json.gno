package postit

import (
	"bytes"
	"strconv"
	"time"
)

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

	// Write Timestamp field
	b.WriteString(`,"Timestamp":"`)
	b.WriteString(post.Timestamp.Format(time.RFC3339))
	b.WriteString(`"`)

	// TODO: Write Comments

	// Write the closing brace
	b.WriteString(`}`)

	return b.String()
}
