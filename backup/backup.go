package backup

import (
	"github.com/gnolang/tx-archive/backup/client"
	"github.com/gnolang/tx-archive/log"
)

// ExecuteBackup executes the node backup process
func ExecuteBackup(
	_ client.Client,
	_ log.Logger,
	_ Config,
) error {
	// Verify the output file can be generated
	// TODO add functionality
	// Determine the right bound
	// TODO add functionality
	// Gather the chain data from the node
	// TODO add functionality
	// Write the chain data to a file
	// TODO add functionality
	return nil
}
