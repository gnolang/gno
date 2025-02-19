package users

import (
	"regexp"
	"std"
	"strconv"
	"strings"
	"testing"

	"gno.land/p/demo/avl"

	"gno.land/p/demo/uassert"
	"gno.land/p/demo/users"
)

// 테스트용 타입 정의
type Profile struct {
	Header  string
	Profile string
	Image   string
	Blog    *avl.Tree
	Post    *avl.Tree
}

type User struct {
	Name    string
	Address std.Address
	Invites int
	Inviter std.Address
	Profile Profile
	Number  int
}

func TestPreRegisteredTest1(t *testing.T) {
	names := ListUsersByPrefix("test1", 1)
	uassert.Equal(t, len(names), 1)
	uassert.Equal(t, names[0], "test1")
}

// 테스트를 위한 mock 함수들
func mockRegister(name string, profile string) {
	// 기존 validation 로직 중 체인 관련 검증을 제외
	if _, ok := name2User.Get(name); ok {
		panic("name already registered: " + name)
	}

	// Profile 생성
	profile1 := new(Profile)
	profile1.Header = name
	profile1.Profile = profile
	profile1.Image = ""
	profile1.Blog = avl.NewTree()
	profile1.Post = avl.NewTree()

	// counter 증가
	counter++

	// User 생성
	user := &User{
		Name:    name,
		Profile: profile1,
		Number:  counter,
	}

	// 저장
	name2User.Set(name, user)
}

func TestRegister(t *testing.T) {
	testCases := []struct {
		name        string
		username    string
		profile     string
		shouldPanic bool
	}{
		{
			name:        "Valid Registration",
			username:    "test_user123",
			profile:     "Test Profile",
			shouldPanic: false,
		},
		{
			name:        "Duplicate Username",
			username:    "test_user123",
			profile:     "Another Profile",
			shouldPanic: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if tc.shouldPanic && r == nil {
					t.Errorf("Expected panic but got none")
				}
				if !tc.shouldPanic && r != nil {
					t.Errorf("Unexpected panic: %v", r)
				}
			}()

			// mock Register 호출
			mockRegister(tc.username, tc.profile)

			if !tc.shouldPanic {
				userI, ok := name2User.Get(tc.username)
				if !ok {
					t.Error("User was not created")
					return
				}

				user := userI.(*User)
				if user.Name != tc.username {
					t.Errorf("Expected username %s, got %s", tc.username, user.Name)
				}
			}
		})
	}
}

// 테스트 후 cleanup
func cleanup() {
	name2User = avl.NewTree()
	addr2User = avl.NewTree()
	invites = avl.NewTree()
	counter = 0
}
