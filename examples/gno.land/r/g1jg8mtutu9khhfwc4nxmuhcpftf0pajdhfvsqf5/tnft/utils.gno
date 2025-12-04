package tnft

import (
	"math/rand"
	"time"

	"gno.land/p/demo/tokens/grc721"
	"gno.land/p/nt/ufmt"
)

// tid converts uint64 to grc721.TokenID.
func tid(id uint64) grc721.TokenID {
	return grc721.TokenID(ufmt.Sprintf("%d", id))
}

// generateRandInstance generates a new random instance.
func generateRandInstance() *rand.Rand {
	seed1 := time.Now().Unix() + nft.TokenCount()
	seed2 := time.Now().UnixNano() + nft.TokenCount()
	pcg := rand.NewPCG(uint64(seed1), uint64(seed2))
	return rand.New(pcg)
}
