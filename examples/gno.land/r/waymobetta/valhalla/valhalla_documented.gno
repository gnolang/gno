package valhalla // This package provides a membership management system using AVL trees. 

import (
	"errors"
	// The "demo" package from gno.land/p contains utility functions for formatting output. We'll use ufmt to format our member list output nicely. 	
	"gno.land/p/demo/ufmt"
	// AVL trees are a type of self-balancing binary search tree, which we're using here to efficiently store and retrieve membership information based on unique keys. 
	"gno.land/p/demo/avl"
)

var MemberList avl.Tree // This variable holds our AVL tree containing the list of members and their associated status (true for active member, false otherwise). // We're using a global variable here because we want to be able to access it from multiple functions in this package. 

// func init() {
// // When the package is imported, we create a new empty AVL tree called MemberList. 
// 	MemberList = avl.NewTree()
// }

func AddMember(memberKey string) error { // This function adds a member to our list using their unique key as an identifier. It returns an error if the member already exists or if there's another issue adding them (e.g., they don't have admin privileges). 
	assertIsAdmin() // We check that the current user has administrator privileges before allowing them to add a new member, since this is a sensitive operation. 
	if IsMember(memberKey) {
		// Check if the member already exists in our list (using their unique key as an identifier). 
		return errors.New("member already exists") // If they do, we return an error indicating that the member has already been added to the system. 
	}
	MemberList.Set(memberKey, true) // Otherwise, we add them to the AVL tree using Set(). This sets their status to true (meaning they're now a valid member). 

	return nil // We don't need to return anything else since there shouldn't be any other errors at this point. 
}

func RemoveMember(memberKey string) { // This function removes a member from our list using their unique key as an identifier. It returns an error if the member doesn't exist or if there's another issue removing them (e.g., they don't have admin privileges). 
	assertIsAdmin() // We check that the current user has administrator privileges before allowing them to remove a member, since this is a sensitive operation. 
	if !IsMember(memberKey) {
	// Check if the member exists in our list (using their unique key as an identifier). 
		return errors.New("member does not exist") // If they don't, we return an error indicating that the member isn't currently part of the system. 
	}
	MemberList.Remove(memberKey) // Otherwise, we remove them from the AVL tree using Remove(). This sets their status to false (meaning they're no longer a valid member). 

	return nil // We don't need to return anything else since there shouldn't be any other errors at this point.
}

func IsMember(memberKey string) bool { // This function checks whether a given member is currently part of our list using their unique key as an identifier. 
	return MemberList.Has(memberKey) // We simply check if the AVL tree contains the specified key (which should correspond to the member's status). 
}

func MemberCount() int { // This function returns the number of members currently in our list, based on the size of the AVL tree. 
	return MemberList.Size() // The Size() method returns the current size of the AVL tree, which is equal to the number of keys (and therefore members) it contains. 
}
