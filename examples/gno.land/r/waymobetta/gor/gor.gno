package gor

import (
	"errors"
	"std"
	"strconv"
	
	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/users"
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

func RegisterAddress(
	addr std.Address,
	username string,
) error {
	if !GnoGitMapping.Has(addr.String()) {
		added := GnoGitMapping.Set(addr.String(), username)
		if !added {
			return errors.New("could not register address")
		} 
	}
	return nil
}

func RegisterUsername(
	addr std.Address,
	username string,
) {
	if !GitGnoMapping.Has(username) {
		GitGnoMapping.Set(username, addr.String())
	}
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
	var prSlice []PR
	prSliceIface, ok := GnoPRMapping.Get(addr.String())
	if ok {
		return prSliceIface.([]PR)
	}

	return []PR{}
}

func GetGnoPRCount(addr std.Address) int {
	var prSlice []PR
	prSliceIface, ok := GnoPRMapping.Get(addr.String())
	if ok {
		prSlice = prSliceIface.([]PR)
		return len(prSlice)
	}

	return 0
}

func Render() {
	println("GOR Realm")
}
