// This package provides functions for managing administrative tasks, such as setting the administrator's address and pausing system operations.
package gor

import (
	"std"
)

// The following variables are used for storing administrative information:
var (
	adminAddr     std.Address    // Holds the address of the current administrator.
	inPause       bool           // Indicates whether system operations are currently paused or not.
)

// Initialization function that sets up the admin's initial address when the package is imported:
func init() {
	adminAddr = std.GetOrigCaller()
}

// Function for setting the administrator's address. Only authorized users (i.e., admins) can call this function:
func AdminSetAdminAddr(addr std.Address) {
	assertIsAdmin() // Calls assertIsAdmin to ensure that only an admin is making this request.
	adminAddr = addr // Updates the administrator's address with the new value passed as a parameter.
}

// Function for setting whether system operations are paused or not:
func AdminSetInPause(state bool) {
	assertIsAdmin() // Calls assertIsAdmin to ensure that only an admin is making this request.
	inPause = state // Updates the inPause variable with the new value passed as a parameter.
}

// Function for checking whether a given address is authorized (i.e., belongs to the administrator):
func isAdmin(addr std.Address) bool {
	return addr == adminAddr // Returns true if the addresses match, false otherwise.
}

// Helper function that panics with an error message when called by unauthorized users:
func assertIsAdmin() {
	caller := std.GetOrigCaller() // Retrieves the address of the caller (i.e., the user making this request).
	if !isAdmin(caller) { // Checks whether the caller's address matches that of the administrator.
		panic("access restricted") // If not, panics with an error message indicating access restriction.
	}
}

// Helper function for checking whether system operations are currently paused:
func assertNotInPause() {
	if inPause { // Checks whether the inPause variable is true (i.e., indicates that system operations are paused).
		panic("access restricted (pause)") // If so, panics with an error message indicating access restriction due to pause mode.
	}
}

// Function for retrieving the current administrator's address as a string:
func GetAdmin() string {
	return adminAddr.String() // Returns the administrator's address in string format using std.Address's String method.
}
