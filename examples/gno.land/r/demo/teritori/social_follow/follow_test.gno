package social_follow

import (
	"std"
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/testutils"
)

func TestFollow_Follow(t *testing.T) {
	addr2User = avl.Tree{}
	main := testutils.TestAddress("main")
	std.TestSetOrigCaller(main)

	user := testutils.TestAddress("user")

	Follow(user)

	if c := FollowersCount(user); c != 1 {
		t.Fatalf("FollowersCount expected to have 1 has %d", c)
	}

	if followers := Followers(user, 0, 1); followers[0].String() != string(main) {
		t.Fatalf("Followers expected to have %s has %s", string(main), followers[0].String())
	}

	if followed := Followed(main, 0, 1); followed[0].String() != string(user) {
		t.Fatalf("Followed expected to have %s has %s", string(user), followed[0].String())
	}
}

func TestFollow_UnFollow(t *testing.T) {
	addr2User = avl.Tree{}

	user1 := testutils.TestAddress("user1")
	user2 := testutils.TestAddress("user2")
	std.TestSetOrigCaller(user1)

	Follow(user2)
	if IsFollower(user1, user2) != true {
		t.Fatalf("expected to be a follower")
	}

	Unfollow(user2)
	if IsFollower(user1, user2) != false {
		t.Fatalf("expected to not be a follower")
	}
}

func TestFollow_FollowedPagination(t *testing.T) {
	addr2User = avl.Tree{}

	main := testutils.TestAddress("main")
	std.TestSetOrigCaller(main)
	for i := 0; i < 10; i++ {
		user := testutils.TestAddress("user" + string(i))
		Follow(user)
	}

	for i := 0; i < 10; i++ {
		if followed := Followed(main, i, 1); len(followed) != 1 {
			t.Fatalf("at page %d expected to have 1 has %d", i, len(followed))
		}
	}

	if followed := Followed(main, 10, 1); len(followed) != 0 {
		t.Fatalf("at page 10 expected to have 0 has %d", len(followed))
	}

	if followed := Followed(main, 0, 10); len(followed) != 10 {
		t.Fatalf("at page 0 expected to have 10 has %d", len(followed))
	}

	if followed := Followed(main, 1, 8); len(followed) != 2 {
		t.Fatalf("at page 2 expected to have 2 has %d", len(followed))
	}
}

func TestFollow_FollowersPagination(t *testing.T) {
	main := testutils.TestAddress("main")

	for i := 0; i < 10; i++ {
		user := testutils.TestAddress("user" + string(i))
		std.TestSetOrigCaller(user)
		Follow(main)
	}

	for i := 0; i < 10; i++ {
		if followers := Followers(main, i, 1); len(followers) != 1 {
			t.Fatalf("at page %d expected to have 1 has %d", i, len(followers))
		}
	}

	if followers := Followers(main, 10, 1); len(followers) != 0 {
		t.Fatalf("at page 10 expected to have 0 has %d", len(followers))
	}

	if followers := Followers(main, 0, 10); len(followers) != 10 {
		t.Fatalf("at page 0 expected to have 10 has %d", len(followers))
	}

	if followers := Followers(main, 1, 8); len(followers) != 2 {
		t.Fatalf("at page 2 expected to have 2 has %d", len(followers))
	}
}
