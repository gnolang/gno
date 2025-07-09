// Package cid implements Content Identifiers (CIDs) for Gno.
// CIDs let us uniquely identify content regardless of where it's stored,
// forming the foundation for content-addressed systems like IPFS.
//
// This is a starting point - we'll build it up step by step.
package cid

import (
	"errors"
)

// TODO:
// 1. Implement multiformat components:
//    - Multihash: How we hash content and identify which hash function was used
//    - Multibase: How we encode CIDs as strings (hex, base32, etc.)
//    - Multicodec: How we identify what type of content we're dealing with
//
// 2. Core CID functionality:
//    - Creation: Simple ways to make CIDs from content
//    - Parsing: Convert from bytes/strings back to CIDs
//    - Utilities: String conversion, comparison, etc.
//
// 3. Tests:
//    - Validate the basics work
//    - Make sure our multiformat implementations are solid
//    - Check compatibility with IPFS

// Version is just a number that tells us which CID format we're using.
type Version uint64

const (
	V0 Version = 0 // CIDv0: Original format, just SHA2-256 + base58
	V1 Version = 1 // CIDv1: Current format with codec info and flexible encoding
)

// Standard errors you might encounter when working with CIDs.
var (
	ErrCidTooShort      = errors.New("cid: not enough bytes for CID")
	ErrInvalidVersion   = errors.New("cid: invalid CID version")
	ErrInvalidMultihash = errors.New("cid: invalid multihash")
)

// CID is the main structure - it represents a content identifier.
// It combines a version, content type info, and a hash of the content.
// 
// For now, we're using simple types that we'll replace with proper
// implementations as we build out the package.
type CID struct {
	Version Version // 0 or 1
	Codec   uint64  // What kind of content this points to
	Hash    []byte  // The hash of the content
} 