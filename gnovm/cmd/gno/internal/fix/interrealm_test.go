package fix

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_interrealm_PosBuilding(t *testing.T) {
	// This test ensures that whenever we add a `cur realm` parameter,
	// it does not display like `func ResetPosts(cur realm,)` - ie. with an
	// additional, redundant comma.

	const src = `package minisocial

import "gno.land/p/demo/ownable"

var Ownable = ownable.NewWithAddress("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5") // @leohhhn

// ResetPosts allows admin deletion of the posts
func ResetPosts() {
	crossing()
	Ownable.AssertOwnedByPrevious()
	posts = nil
}
`
	const want = `package minisocial

import "gno.land/p/demo/ownable"

var Ownable = ownable.NewWithAddress("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5") // @leohhhn

// ResetPosts allows admin deletion of the posts
func ResetPosts(cur realm) {
	Ownable.AssertOwnedByPrevious()
	posts = nil
}
`
	fset, f := mustParse(src)
	interrealm(f)
	assert.Equal(t, want, doFormat(fset, f))
}
