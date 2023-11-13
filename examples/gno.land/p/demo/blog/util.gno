package blog

import "strings"

func breadcrumb(parts []string) string {
	return "# " + strings.Join(parts, " / ") + "\n\n"
}

func merge(left, right []*Post) []*Post {
	result := make([]*Post, 0, len(left)+len(right))

	for len(left) > 0 || len(right) > 0 {
		if len(left) == 0 {
			return append(result, right...)
		}
		if len(right) == 0 {
			return append(result, left...)
		}
		if left[0].CreatedAt.After(right[0].CreatedAt) {
			result = append(result, left[0])
			left = left[1:]
		} else {
			result = append(result, right[0])
			right = right[1:]
		}
	}

	return result
}

func mergeSort(posts []*Post) []*Post {
	if len(posts) <= 1 {
		return posts
	}

	middle := len(posts) / 2
	left := mergeSort(posts[:middle])
	right := mergeSort(posts[middle:])

	return merge(left, right)
}
