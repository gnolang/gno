//go:build cgo

package config

// Cgo builds default to LMDB: fastest reads at scale (see the DBBackend TOML
// comment), but the backend requires cgo.
const defaultDBBackend = "lmdbdb"
