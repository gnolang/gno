package blog

import (
	"gno.land/p/demo/ownable"
)

func ResetPosts(admin ownable.Ownable) {
	admin.AssertOwnedByPrevious()
}
