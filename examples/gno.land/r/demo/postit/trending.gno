package postit

import (
	"sort"
)

type Trending struct {
	Tag   string
	Count uint64
}

func updateTrending(tag string, count uint64) {
	// check if we can return early, since we know
	// `trending` is sorted and have constant size i.e 10.
	if trending[len(trending)-1].Count > count {
		return
	}

	// if already in trending, update count and get the index
	index := -1
	for i := 0; i < len(trending); i++ {
		if trending[i].Tag == tag {
			trending[i].Count = count
			index = i
		}
	}
	if index != -1 { // means already in trending, move to correct position and return
		for i := index; i > 0; i-- {
			if trending[i-1].Count < trending[i].Count { // swap
				temp := trending[i-1]
				trending[i-1] = trending[i]
				trending[i] = temp
			}
		}
		return
	}

	// Find the index to insert
	index = -1
	for i := 0; i < len(trending); i++ {
		if trending[i].Count <= count {
			index = i
			break
		}
	}
	if index == -1 {
		panic("should not happen")
	}
	// Shift elements to make room for the new one
	for i := len(trending) - 1; i > index; i-- {
		trending[i] = trending[i-1]
	}
	// finally update
	trending[index] = Trending{
		Tag:   tag,
		Count: count,
	}
}
