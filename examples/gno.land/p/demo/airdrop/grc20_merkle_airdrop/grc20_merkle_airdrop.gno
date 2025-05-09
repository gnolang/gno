package grc20_merkle_airdrop

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"gno.land/p/demo/airdrop"
	"gno.land/p/demo/avl"
	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/merkle"
)

var (
	ErrAlreadyClaimed = errors.New("already claimed")
	ErrInvalidProof   = errors.New("invalid merkle proof")
)

type MerkleAirdrop struct {
	root string

	token   grc20.Teller
	claimed *avl.Tree
}

func NewMerkleAirdrop(merkleroot string, token grc20.Teller) *MerkleAirdrop {
	return &MerkleAirdrop{
		root: merkleroot,

		token:   token,
		claimed: avl.NewTree(),
	}
}

func (ma *MerkleAirdrop) Root() string {
	return ma.root
}

func (ma *MerkleAirdrop) Claim(data airdrop.AirdropData, proofs []merkle.Node) error {
	shasum := sha256.Sum256(data.Bytes())
	hash := hex.EncodeToString(shasum[:])

	if ma.claimed.Has(hash) {
		return ErrAlreadyClaimed
	}

	if !merkle.Verify(ma.root, data, proofs) {
		return ErrInvalidProof
	}

	err := ma.token.Transfer(data.Address(), data.Amount())
	if err != nil {
		return err
	}

	ma.claimed.Set(hash, data.Amount())
	return nil
}

func (ma MerkleAirdrop) TotalClaimed() uint64 {
	var claimed uint64 = 0

	ma.claimed.Iterate("", "", func(k string, v interface{}) bool {
		claimed += v.(uint64)
		return false
	})

	return claimed
}
