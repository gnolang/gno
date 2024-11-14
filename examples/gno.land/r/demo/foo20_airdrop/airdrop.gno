package foo20_airdrop

import (
	"encoding/hex"
	"std"
	"strconv"

	"gno.land/p/demo/airdrop"
	"gno.land/p/demo/airdrop/grc20_merkle_airdrop"
	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/json"
	"gno.land/p/demo/merkle"
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/foo20"
)

var (
	token grc20.Teller

	foo20airdrop *grc20_merkle_airdrop.MerkleAirdrop
)

func RegisterMerkleRoot(root string) {
	token = foo20.Token.RealmTeller()

	if foo20airdrop != nil {
		panic("foo20 airdrop merkle root is already registered")
	}
	foo20airdrop = grc20_merkle_airdrop.NewMerkleAirdrop(root, token)
}

func Claim(data airdrop.AirdropData, proofs []merkle.Node) {
	err := foo20airdrop.Claim(data, proofs)
	if err != nil {
		panic(err.Error())
	}
}

type ClaimRequest struct {
	Address string `json:"address"`
	Amount  string `json:"amount"`
	Proof   []struct {
		Position float64 `json:"position"`
		Data     string  `json:"data"`
	} `json:"proof"`
}

func (req ClaimRequest) MustGetAmount() uint64 {
	n, err := strconv.ParseInt(req.Amount, 10, 64)
	if err != nil {
		panic("failed to parse amount: " + req.Amount)
	}
	return uint64(n)
}

func ClaimJSON(in string) {
	n, err := json.Unmarshal([]byte(in))
	if err != nil {
		panic(err.Error())
	}

	req := &ClaimRequest{
		Address: n.MustKey("address").MustString(),
		Amount:  n.MustKey("amount").MustString(),
	}

	arr := n.MustKey("proof").MustArray()
	proof := make([]merkle.Node, len(arr))

	for i, e := range arr {
		pos := e.MustKey("position").MustNumeric()
		data, err := hex.DecodeString(e.MustKey("data").MustString())
		if err != nil {
			panic(err.Error())
		}

		proof[i] = merkle.NewNode([]byte(data), uint8(pos))
	}

	data := airdrop.NewData(std.Address(req.Address), req.MustGetAmount())

	Claim(data, proof)
}

func TotalClaimed() uint64 {
	return foo20airdrop.TotalClaimed()
}

func Render(path string) string {
	if foo20airdrop == nil {
		return "Airdrop is not registered yet"
	}

	return ufmt.Sprintf("total claimed: %d", foo20airdrop.TotalClaimed())
}
