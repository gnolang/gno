// The `protected.gno` file contains public realm functions that expose Boards2
// data to other realms. This allows sub realms to access boards data to be able
// to migrate content from one version to another or to implement specific
// features in separate sub realms.

package boards2

import (
	"gno.land/p/gnoland/boards"
)

// TODO: Authorize sub realms to be able to call protected functions (use DAOs)

// GetRealmPermissions returns Boards2 realm permissions.
func GetRealmPermissions() boards.Permissions {
	return gPerms
}
