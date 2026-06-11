//go:build !cgo

package config

// Pure-Go builds cannot use the cgo-only LMDB/MDBX backends; the default
// falls back to the best pure-Go backend.
const defaultDBBackend = "pebbledb"
