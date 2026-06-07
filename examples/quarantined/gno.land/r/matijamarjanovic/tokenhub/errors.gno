package tokenhub

import (
	"errors"
)

var (
	ErrNFTAlreadyRegistered = errors.New("NFT already registered")
	ErrNFTNotFound          = errors.New("NFT not found")
	ErrMTAlreadyRegistered  = errors.New("multi-token already registered")
	ErrMTNotFound           = errors.New("multi-token not found")
	ErrMTInfoNotFound       = errors.New("multi-token info not found")
	ErrNFTtokIDNotExists    = errors.New("NFT token ID does not exists")
	ErrNFTNotMetadata       = errors.New("NFT must implement IGRC721CollectionMetadata")
)
