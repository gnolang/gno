package gor

import (
	"errors"
	"std"
	"strconv"
	
	"gno.land/p/demo/avl"
)

var (
	// std.Address => github username (string)
	GnoGitMapping = avl.NewTree()
	// github username (string) => std.Address
	GitGnoMapping = avl.NewTree()
	// std.Address => []PR
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
}

// setters

func RegisterAddress(
	addr std.Address,
	username string,
) {
	if GnoGitMapping.Has(addr.String()) {
		// revert tx
		panic("address already registered")
	}
	GnoGitMapping.Set(addr.String(), username)
}

func RegisterUsername(
	addr std.Address,
	username string,
) {
	if GitGnoMapping.Has(username) {
		// revert tx
		panic("username already registered")
	}
		GitGnoMapping.Set(username, addr.String())
}

func NewPR(
	id,
	evaluation,
	maxEvaluation,
	commits,
	additions,
	deletions,
	totalEffectiveLines,
	avgCharsPerLine int,
	username,
	category,
	subCategory string,
	) {
	var addrStr string
	// get address from username
	addrIface, ok := GitGnoMapping.Get(username)
	if ok {
		addrStr = addrIface.(string)
	}

	// get previous slice of PR IDs for address
	prStateSliceIface, ok := GnoPRMapping.Get(addrStr)
	var prSlice []PR
	if ok {
		prSlice = prStateSliceIface.([]PR)
	}

	prSlice = append(
		prSlice,
		PR{
			ID: id,
			Evaluation: evaluation,
			MaxEvaluation: maxEvaluation,
			Category: category,
			SubCategory: subCategory,
			Commits: commits,
			Additions: additions,
			Deletions: deletions,
			TotalEffectiveLines: totalEffectiveLines,
			AvgCharsPerLine: avgCharsPerLine,
		},
	)

	// store updated slice of PRs
	updated := GnoPRMapping.Set(
		addrStr, 
		prSlice,
	)
}

// getters

func GetGnoGitMapping(addr std.Address) (string, bool) {
	return GnoGitMapping.Get(addr.String())
}

func GetGitGnoMapping(username string) (std.Address, bool) {
	return GitGnoMapping.Get(username)
}

func GetGnoPRMapping(addr std.Address) []PR {
	existing, ok := GnoPRMapping.Get(addr.String())
	if ok {
		return existing.([]PR)
	}

	return []PR{}
}

func GetGnoPRCount(addr std.Address) int {
	existing, ok := GnoPRMapping.Get(addr.String())
	if ok {
		return len(existing.([]PR))
	}

	return 0
}

func GetGnoGitMappingSize() int {
	return GnoGitMapping.Size()
}

func GetGitGnoMappingSize() int {
	return GitGnoMapping.Size()
}

func Render() {
	println("GOR Realm")
}
