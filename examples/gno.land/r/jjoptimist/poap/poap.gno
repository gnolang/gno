package poap

import (
    "std"
    "time"
    
    "gno.land/p/demo/avl"
    "gno.land/p/demo/grc/grc721"
)

type POAP struct {
    id          uint64    // Unique identifier
    name        string    // Event name
    description string    // Event description
    date        time.Time // Event date
    maxMints    int      // Maximum number of POAPs
    mintCount   int      // Current mint count
}

var (
    poaps      *avl.Tree                               // Store POAPs
    poapCount  uint64                                  // Counter for POAP IDs
    nftTokens  = grc721.NewBasicNFT("POAP", "POAP")  // NFT implementation
)

func init() {
    poaps = avl.NewTree()
} 