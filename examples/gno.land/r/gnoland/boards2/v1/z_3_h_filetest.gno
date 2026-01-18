package main

import (
	"strings"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	name          = "foo123"
)

var newName string

func init() {
	newName = strings.Repeat("A", boards2.MaxBoardNameLength+1)

	testing.SetRealm(testing.NewUserRealm(owner))
	boards2.CreateBoard(cross, name, false, false)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.RenameBoard(cross, name, newName)
}

// Error:
// board name is too long, maximum allowed is 50 characters
