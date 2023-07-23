package airdrop

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/merkle"
	"gno.land/r/demo/foo20"
	"gno.land/r/demo/users"
)

var (
	ErrAlreadyClaimed = errors.New("already claimed")
	ErrInvalidProof   = errors.New("invalid merkle proof")
)

type AirdropData struct {
	Address std.Address
	// TODO: use std.Coin
	Amount uint64
	// Amount std.Coin
}

func (data AirdropData) Bytes() []byte {
	// TODO: use binary.Write
	// var buf bytes.Buffer
	// binary.Write(&buf, binary.BigEndian, d)
	// return buf.Bytes()
	// OR: use json.Marshal for frontend compatibilities

	s := fmt.Sprintf("%v", data)
	return []byte(s)
}

type MerkleAirdrop struct {
	root string

	token   grc20.IGRC20
	claimed *avl.Tree
}

func NewMerkleAirdrop(merkleroot string, token grc20.IGRC20) *MerkleAirdrop {
	return &MerkleAirdrop{
		root: merkleroot,

		token:   token,
		claimed: avl.NewTree(),
	}
}

func (ma *MerkleAirdrop) Root() string {
	return ma.root
}

func (ma *MerkleAirdrop) Claim(data AirdropData, proofs []merkle.Node) error {
	shasum := sha256.Sum256(data.Bytes())
	hash := hex.EncodeToString(shasum[:])

	if ma.claimed.Has(hash) {
		return ErrAlreadyClaimed
	}

	if !merkle.Verify(ma.root, data, proofs) {
		return ErrInvalidProof
	}

	// THIS WORKS
	// foo20.Transfer(users.AddressOrName(data.Address), data.Amount)
	// foo20.GRC20().Transfer(data.Address, data.Amount)

	// THIS DOES NOT WORK
	err := ma.token.Transfer(data.Address, data.Amount)
	if err != nil {
		return err
	}

	ma.claimed.Set(hash, data.Amount)
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
