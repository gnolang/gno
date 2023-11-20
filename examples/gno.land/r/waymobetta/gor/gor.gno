package gor

import (
	"std"
	
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
		adminAddr.String(),  // The address of our admin user (stored in a variable called `adminAddr`)
		"waymobetta",        // The corresponding GitHub username for this admin user
	)
	GitGnoMapping.Set(
		"waymobetta",         // The GitHub username of our admin user
		adminAddr.String(),  // The address associated with this username (stored in the `adminAddr` variable)
	)
}

// RegisterAddress registers an Ethereum address in our mapping (std.Address -> GitHub username).
func RegisterAddress(
	addr std.Address, 
	username string,
	) {
	if GnoGitMapping.Has(addr.String()) { // Check if the address is already registered
		panic("address already registered") // If yes, revert the transaction and throw a panic with an error message
	}
	GnoGitMapping.Set(addr.String(), username) // Otherwise, insert the new mapping into our tree
}

// RegisterUsername registers a GitHub username in our mapping (Github username -> std.Address)
func RegisterUsername(addr std.Address, username string) {
	if GitGnoMapping.Has(username) { // Check if the username is already registered
		panic("username already registered") // If yes, revert the transaction and throw a panic with an error message
	}
	GitGnoMapping.Set(username, addr.String()) // Otherwise, insert the new mapping into our tree
}

// NewPR creates a new PR object and stores it in our mapping from standard addresses to slices of previous PRs.
func NewPR(
	id, // ID of the new PR (int)
	evaluation, // Evaluation score for this PR (int)
	maxEvaluation, // Maximum evaluation score allowed for this category/subcategory (int)
	commits, // Number of commits in this PR (int)
	additions, // Number of lines added by this PR (int)
	deletions, // Number of lines deleted by this PR (int)
	totalEffectiveLines, // Total number of effective lines changed by this PR (int)
	avgCharsPerLine int, // Average number of characters per line in this PR (int)
	username string, // GitHub username associated with the author of this PR (string)
	category, // Category for this PR (string)
	subCategory string, // Sub-category for this PR (optional, can be empty) (string)
	) {
	var addrStr string // Temporary variable to store the standard address corresponding to the given username
	// Get the standard address associated with the given GitHub username from our mapping
	addrIface, ok := GitGnoMapping.Get(username)
	if ok {
		addrStr = addrIface.(string) // Convert interface value (interface{}) to string and store it in `addrStr`
	}

	// Get the previous slice of PR IDs for this address from our mapping
	prStateSliceIface, ok := GnoPRMapping.Get(addrStr)
	var prSlice []PR // Initialize a new empty slice called `prSlice` to store the updated list of PRs
	if ok {
		// If we found an existing slice for this address in our mapping...
		prSlice = prStateSliceIface.([]PR) // Convert interface value (interface{}) to []PR and assign it to `prSlice`
	}

	// Add the new PR object to the end of the list stored in `prSlice`
	prSlice = append(
		prSlice,
		PR{
			ID: id, // Set ID field for our new PR object
			Evaluation: evaluation, // Set evaluation score field for our new PR object
			MaxEvaluation: maxEvaluation, // Set maximum evaluation score allowed field for our new PR object (based on category/subcategory)
			Category: category, // Set category field for our new PR object
			SubCategory: subCategory, // Set sub-category field for our new PR object (optional)
			Commits: commits, // Set number of commits field for our new PR object
			Additions: additions, // Set number of lines added field for our new PR object
			Deletions: deletions, // Set number of lines deleted field for our new PR object
			TotalEffectiveLines: totalEffectiveLines, // Set total number of effective lines changed field for our new PR object
			AvgCharsPerLine: avgCharsPerLine, // Set average number of characters per line field for our new PR object
		},
	)

	// Store the updated list of PRs in our mapping using `addrStr` as the key
	GnoPRMapping.Set(
		addrStr,  // Key (address string)
		prSlice, // Value (slice of PR objects)
	)
}

// getters

// Returns the username associated with a given Gno address
func GetGnoGitMapping(addr std.Address) (string, bool) {
	return GnoGitMapping.Get(addr.String())
}

// GetGitGnoMapping returns the standard address associated with a given GitHub username, or an empty string if not found
func GetGitGnoMapping(username string) (std.Address, bool) {
	return GitGnoMapping.Get(username)
}

// GetGnoPRMapping returns a list of PR objects associated with a given GitHub user's standard address, or an empty slice if not found
func GetGnoPRMapping(addr std.Address) []PR {
	existing, ok := GnoPRMapping.Get(addr.String()) // Retrieve interface value from our mapping for a provided user's address
	if ok {
		return existing.([]PR)
	}
	return []PR{}
}

// GetGnoPRCount returns the number of PRs associated with a given user's address; returns 0 if none found
func GetGnoPRCount(addr std.Address) int {
	existing, ok := GnoPRMapping.Get(addr.String()) // Retrieve interface value from our mapping for this user's standard address
	if ok {
		return len(existing.([]PR))
	}
	return 0
}

// GetGnoGitMappingSize returns the total number of standard addresses currently stored in our GitHub-to-standard address mapping, or 0 if empty
func GetGnoGitMappingSize() int {
	return GnoGitMapping.Size()                         // Return the size (number of key-value pairs) of our `GnoGitMapping` map as an integer
}

// GetGitGnoMappingSize returns the total number of standard addresses currently stored in our GitHub user-to-standard address mapping, or 0 if empty
func GetGitGnoMappingSize() int {
	return GitGnoMapping.Size()                          // Return the size (number of key-value pairs) of our `GitGnoMapping` map as an integer
}
