package txindexer

// ErrInvalidConfigMissingDBPath is returned when the db path is missing in
// the configuration.
const ErrInvalidConfigMissingDBPath Error = "db path config is required for tx-indexer"

// Error defines helper to create custom errors for the txindexer package.
type Error string

// Error returns the error message as a string.
func (e Error) Error() string {
	return string(e)
}
