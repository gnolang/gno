package test

import (
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockStore implements the minimal Store interface for testing
type mockStore struct {
	gno.Store
	files map[string]*std.MemFile
}

func newMockStore() *mockStore {
	return &mockStore{
		files: make(map[string]*std.MemFile),
	}
}

func (ms *mockStore) GetMemFile(pkgPath, name string) *std.MemFile {
	key := pkgPath + "/" + name
	return ms.files[key]
}

func (ms *mockStore) addFile(pkgPath, name, body string) {
	key := pkgPath + "/" + name
	ms.files[key] = &std.MemFile{
		Name: name,
		Body: body,
	}
}

func TestStoreAdapter(t *testing.T) {
	store := newMockStore()

	// these files will be accessed during profiling
	store.addFile("gno.land/p/demo/users", "users.gno", `package users

import "gno.land/p/demo/avl"

type User struct {
	Name string
	Age  int
}

func GetUser(id string) *User {
	return &User{Name: "test", Age: 25}
}`)

	store.addFile("gno.land/p/demo/avl", "avl.gno", `package avl

type Tree struct {
	root *Node
}

type Node struct {
	key   string
	value any
}`)

	adapter := NewStoreAdapter(store)

	usersFile := adapter.GetMemFile("gno.land/p/demo/users", "users.gno")
	require.NotNil(t, usersFile)
	assert.Contains(t, usersFile.Body, "func GetUser")

	avlFile := adapter.GetMemFile("gno.land/p/demo/avl", "avl.gno")
	require.NotNil(t, avlFile)
	assert.Contains(t, avlFile.Body, "type Tree")

	notFound := adapter.GetMemFile("gno.land/p/demo/notfound", "notfound.gno")
	assert.Nil(t, notFound)
}
