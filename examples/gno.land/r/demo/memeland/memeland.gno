package memeland

import (
	_ "gno.land/p/demo/memeland"
)

var memeland *Memeland

func init() {
	// todo add admin?
	memeland = NewMemeland()
}

func PostMeme(data, timestamp string) string {
	if memeland == nil {
		return "memeland does not exist"
	}
	return memeland.PostMeme(data, timestamp)
}
func Upvote(id string) string {
	if memeland == nil {
		return "memeland does not exist"
	}
	return memeland.Upvote(id)
}

func GetPostsByTimestamp(startTimestamp, endTimestamp string, page, pageSize int) string {
	if memeland == nil {
		return "memeland does not exist"
	}

	return memeland.GetPostsByTimestamp(startTimestamp, endTimestamp, page, pageSize)
}
