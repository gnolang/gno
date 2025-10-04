package tendermint

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
)

type ValidatorSet struct {
	Validators       []*Validator `json:"validators,omitempty"`
	Proposer         *Validator   `json:"proposer,omitempty"`
	TotalVotingPower int64        `json:"total_voting_power,omitempty"`
}

func NewValset(vals ...*Validator) *ValidatorSet {
	var total int64
	for _, v := range vals {
		total += v.VotingPower
	}
	return &ValidatorSet{
		Validators:       vals,
		Proposer:         vals[0],
		TotalVotingPower: total,
	}
}

func (vals ValidatorSet) Hash() []byte {
	var bz []byte
	for _, v := range vals.Validators {
		bz = append(bz, v.Bytes()...)
	}
	// TODO: ibc-go returns a merkle hash, ensure we can just use a simple hash
	// -> see p/demo/merkle
	// return merkle.HashFromByteSlices(bzs)
	h := sha256.Sum256(bz)
	return h[:]
}

// GetByAddress returns an index of the validator with address and validator
// itself (copy) if found. Otherwise, -1 and nil are returned.
func (vals *ValidatorSet) GetByAddress(address []byte) (index int32, val *Validator) {
	for idx, val := range vals.Validators {
		if bytes.Equal(val.Address, address) {
			return int32(idx), val
		}
	}
	return -1, nil
}

type Validator struct {
	Address          []byte `json:"address,omitempty"`
	PubKey           []byte `json:"pub_key"`
	VotingPower      int64  `json:"voting_power,omitempty"`
	ProposerPriority int64  `json:"proposer_priority,omitempty"`
}

func NewValidator(addr, pubkey string, vp int64) *Validator {
	addrBz, err := base64.StdEncoding.DecodeString(addr)
	if err != nil {
		panic(err)
	}
	pubkeyBz, err := base64.StdEncoding.DecodeString(pubkey)
	if err != nil {
		panic(err)
	}
	return &Validator{
		Address:     addrBz,
		PubKey:      pubkeyBz,
		VotingPower: vp,
	}
}

// Bytes returns a bytes representation of v, composed as follow:
// 32bytes: Pubkey
// varint: VotingPower
// TODO implement manual proto just like Vote ?
func (v Validator) Bytes() (bz []byte) {
	bz = append(bz, v.PubKey...)
	bz = binary.AppendVarint(bz, v.VotingPower)
	return
}
