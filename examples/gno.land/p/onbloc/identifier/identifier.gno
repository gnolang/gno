// Package identifier issues short, unique, realm-scoped identifiers.
//
// A Generator is bound to its realm's package path and holds a per-block counter.
// NextID hashes the package path, block height, and counter into a short fixed-length code;
// block height plus counter make each id unique within the realm.
// Keep one persistent Generator per realm.
package identifier

import (
	"chain/runtime"
	"crypto/sha256"
	"strconv"

	"gno.land/p/nt/cford32/v0"
	"gno.land/p/nt/seqid/v0"
)

// idBytes: leading SHA-256 bytes kept per code.
// 8 bytes (64 bits) yield a 13-char code with negligible within-realm collision risk.
const idBytes = 8

// Generator issues realm-scoped identifiers. Construct with NewGenerator.
type Generator struct {
	packagePath  string   // owning realm, captured at construction
	latestHeight int64    // block height of the last NextID; sequence resets when it changes
	sequence     seqid.ID // per-block sequence, making each id unique within a block
}

// NewGenerator returns a Generator bound to the caller's realm.
// rlm must be the caller's own captured cur (asserted via IsCurrent);
// the leading unused int keeps rlm out of first position so this stays a non-crossing helper.
func NewGenerator(_ int, rlm realm) *Generator {
	if !rlm.IsCurrent() {
		panic(ErrSpoofedRealm)
	}

	pkgPath := rlm.PkgPath()
	if pkgPath == "" {
		panic(ErrNotRealm)
	}

	return &Generator{
		packagePath: pkgPath,
		sequence:    seqid.ID(0),
	}
}

// PackagePath returns the realm the generator is bound to.
func (g *Generator) PackagePath() string {
	return g.packagePath
}

// NextID returns the Crockford base32 of SHA-256(packagePath ":" height ":" sequence)[:idBytes] for the current block height.
// The sequence resets each block and bumps per call, so (height, sequence) is unique within the realm.
func (g *Generator) NextID() string {
	currentHeight := runtime.ChainHeight()

	// If the current height is greater than the latest height, reset the sequence.
	if currentHeight > g.latestHeight {
		g.latestHeight = currentHeight
		g.sequence = seqid.ID(0)
	}

	g.sequence.Next()

	seed := g.packagePath + ":" +
		strconv.FormatInt(currentHeight, 10) + ":" +
		g.sequence.String()
	sum := sha256.Sum256([]byte(seed))

	return cford32.EncodeToStringLower(sum[:idBytes])
}
