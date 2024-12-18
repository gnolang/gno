package todos

import (
	"std"
	"gno.land/p/demo/grc/grc721"
)

type TodoListNFT struct {
	nft grc721.IGRC721MetadataOnchain
	creator std.Address
}

func NewTodoListNFT(name string, symbol string) *TodoListNFT {
	return &TodoListNFT{
		nft: grc721.NewNFTWithMetadata(name, symbol),
		creator: std.GetOrigCaller(),
	}
}

func (t *TodoListNFT) Creator() std.Address {
	return t.creator
}


// GetTokensByAssignee retrieves all tokens where the owner is the assignee.
func (t *TodoListNFT) GetTokensByAssignee(assignee std.Address) ([]grc721.TokenID, error) {
	var tokens []grc721.TokenID
	for _, tid := range t.nft.basicNFT.owners.Keys() {
		owner, err := t.OwnerOf(grc721.TokenID(tid))
		if err != nil {
			continue
		}
		// Check if the owner of the token is the assignee
		if owner == assignee {
			tokens = append(tokens, grc721.TokenID(tid))
		}
	}
	return tokens, nil
}

// GetTokenByName retrieves a token by its name.
func (t *TodoListNFT) GetTokenByName(name string) (grc721.TokenID, error) {
	for _, tid := range t.nft.basicNFT.owners.Keys() {
		metadata, err := t.TokenMetadata(grc721.TokenID(tid))
		if err != nil {
			continue
		}
		if metadata.Name == name {
			return grc721.TokenID(tid), nil
		}
	}
	return "", ErrTokenNotFound
}

// GetTokensByStatus retrieves all tokens with a specific status.
func (t *TodoListNFT) GetTokensByStatus(status string) ([]grc721.TokenID, error) {
	var tokens []grc721.TokenID
	for _, tid := range t.nft.basicNFT.owners.Keys() {
		metadata, err := t.TokenMetadata(grc721.TokenID(tid))
		if err != nil {
			continue
		}
		for _, attr := range metadata.Attributes {
			if attr.TraitType == "status" && attr.Value == status {
				tokens = append(tokens, grc721.TokenID(tid))
				break
			}
		}
	}
	return tokens, nil
}

// Mint a new task token with default metadata
func (t *TodoListNFT) MintTaskToken(to std.Address, tid grc721.TokenID, name, description, status, assignee string) error {
	// Ensure only the creator can mint new tasks
	if std.GetOrigCaller() != t.creator {
		return ErrOnlyCreatorCanMint
	}

	// Mint the token
	err := t.nft.Mint(to, tid)
	if err != nil {
		return err
	}

	// Set default metadata
	metadata := t.defaultTaskMetadata(name, description, status, assignee)
	return t.SetTokenMetadata(tid, metadata)
}

// Burn a token if its status is "done"
func (t *TodoListNFT) Burn(tid grc721.TokenID) error {
	// Retrieve the metadata for the token
	metadata, err := t.TokenMetadata(tid)
	if err != nil {
		return err
	}

	// Check if the status attribute is "done"
	for _, attr := range metadata.Attributes {
		if attr.TraitType == "status" && attr.Value == "done" {
			// Perform the burn operation
			return t.nft.Burn(tid)
		}
	}

	return ErrTokenCannotBeBurned
}
// Implementing IGRC721 interface methods

func (t *TodoListNFT) BalanceOf(owner std.Address) (uint64, error) {
	return t.nft.BalanceOf(owner)
}

func (t *TodoListNFT) OwnerOf(tid grc721.TokenID) (std.Address, error) {
	return t.nft.OwnerOf(tid)
}

func (t *TodoListNFT) SetTokenURI(tid grc721.TokenID, tURI grc721.TokenURI) (bool, error) {
	return t.nft.SetTokenURI(tid, tURI)
}

func (t *TodoListNFT) SafeTransferFrom(from, to std.Address, tid grc721.TokenID) error {
	return t.nft.SafeTransferFrom(from, to, tid)
}

func (t *TodoListNFT) TransferFrom(from, to std.Address, tid grc721.TokenID) error {
	return t.nft.TransferFrom(from, to, tid)
}

func (t *TodoListNFT) Approve(approved std.Address, tid grc721.TokenID) error {
	return t.nft.Approve(approved, tid)
}

func (t *TodoListNFT) SetApprovalForAll(operator std.Address, approved bool) error {
	return t.nft.SetApprovalForAll(operator, approved)
}

func (t *TodoListNFT) GetApproved(tid grc721.TokenID) (std.Address, error) {
	return t.nft.GetApproved(tid)
}

func (t *TodoListNFT) IsApprovedForAll(owner, operator std.Address) bool {
	return t.nft.IsApprovedForAll(owner, operator)
}

// Additional methods for metadata

func (t *TodoListNFT) SetTokenMetadata(tid grc721.TokenID, metadata grc721.Metadata) error {
	return t.nft.SetTokenMetadata(tid, metadata)
}

func (t *TodoListNFT) TokenMetadata(tid grc721.TokenID) (grc721.Metadata, error) {
	return t.nft.TokenMetadata(tid)
}

// Predefined metadata attributes for a task
func (t *TodoListNFT) defaultTaskMetadata(name, description, status, assignee string) grc721.Metadata {
	return grc721.Metadata{
		Name:        name,
		Description: description,
		Attributes: []grc721.Trait{
			{TraitType: "status", Value: status},
			{TraitType: "assignee", Value: assignee},
		},
	}
}

