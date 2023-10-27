package gor

import (
	"errors"
	"std"
	"strconv"
	
	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

var (
	// std.Address => github username (string)
	GnoGitMapping = avl.NewTree()
	// github username (string) => std.Address
	GitGnoMapping = avl.NewTree()
	// std.Address => []string
	GnoPRMapping = avl.NewTree()
)

// constructor
func init() {
	// set admin
	GnoGitMapping.Set(
		adminAddr.String(), 
		"waymobetta",
	)
	GitGnoMapping.Set(
		"waymobetta", 
		adminAddr.String(),
	)
	GnoPRMapping.Set(
		adminAddr.String(), 
		[]interface{},
	)
}

func RegisterAddress(
	addr std.Address,
	username string,
) string {
	if !GnoGitMapping.Has(addr.String()) {
		added := GnoGitMapping.Set(addr.String(), username)
		if !added {
			return ufmt.Sprintf("%v", errors.New("could not register address"))
		} 
	}
	return ""
}

func RegisterUsername(
	addr std.Address,
	username string,
) string {
	if !GitGnoMapping.Has(username) {
		added := GitGnoMapping.Set(username, addr.String())
		if !added {
			return ufmt.Sprintf("%v", errors.New("could not register username"))
		}
	}
	return ""
}

// func NewPR(
// 	id,
// 	evaluation,
// 	maxEvaluation,
// 	commits,
// 	additions,
// 	deletions,
// 	totalEffectiveLines,
// 	avgCharsPerLine int,
// 	username,
// 	category,
// 	subCategory string,
// 	) {

func NewPR(
		addr std.Address,
		id string,
		) string {
		// get previous slice of PR IDs
		prInterface, exists := GnoPRMapping.Get(addr.String())
		if !exists {
			return ufmt.Sprintf("%v", errors.New("Gno address not found"))
		}

		var prSlice []interface{}

		// cast interface into string
		prSlice = append(
			prSlice, 
			prInterface,
		)

		// append new id to slice of PR IDs
		// update tree with new slice
		updated := GnoPRMapping.Set(
			addr.String(), 
			prSlice,
		)
		if !updated {
			return ufmt.Sprintf("%v", errors.New("could not update tree"))
		}
		
		return ""

		// pr := PR{
		// 	ID: id,
		// 	Evaluation: evaluation,
		// 	MaxEvaluation: maxEvaluation,
		// 	Category: category,
		// 	SubCategory: subCategory,
		// 	Commits: commits,
		// 	Additions: additions,
		// 	Deletions: deletions,
		// 	TotalEffectiveLines: totalEffectiveLines,
		// 	AvgCharsPerLine: avgCharsPerLine,
		// }

		// prSlice = append(prSlice, pr)

		// // update mapping
		// GitPRMapping.Set(username, pr)

		// // get address from reverse mapping lookup
		// addr := GitGnoMapping.Get())

		// // update mapping
		// GnoPRMapping[addr] = append(prSlice, pr)
}

// getters

func GetGnoGitMapping(addr std.Address) (string, bool) {
	return GnoGitMapping.Get(addr.String())
}

func GetGitGnoMapping(username string) (std.Address, bool) {
	return GitGnoMapping.Get(username)
}

func GetGnoPRMapping(addr std.Address) ([]PR, bool) {
	return GnoPRMapping.Get(addr.String())
}

// func GnoPRCount(addr std.Address) int {
// 	PRSlice, exists := GnoPRMapping(addr.String())
// 	if !exists {
// 		return 0
// 	}
// 	return len(PRSlice)
// }

// func GitPRCount(username string) int {
// 	PRSlice, exists := GitPRMapping(addr.String())
// 	if !exists {
// 		return 0
// 	}
// 	return len(PRSlice)
// }

func Render() {
	println("GOR Realm")
}
