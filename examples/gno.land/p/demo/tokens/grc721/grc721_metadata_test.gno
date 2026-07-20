package grc721

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)

func newDummyMetadataNFT(_ int, rlm realm) *metadataNFT {
	var nft *metadataNFT
	func(cur realm) {
		nft = NewNFTWithMetadata(0, cur, dummyNFTName, dummyNFTSymbol)
	}(cross(rlm))
	return nft
}

func TestSetMetadata(cur realm, t *testing.T) {
	// Create a new dummy NFT with metadata
	dummy := newDummyMetadataNFT(0, cur)
	if dummy == nil {
		t.Errorf("should not be nil")
	}

	// Define addresses for testing purposes
	addr1 := testutils.TestAddress("alice")
	addr2 := testutils.TestAddress("bob")

	// Define metadata attributes
	name := "test"
	description := "test"
	image := "test"
	imageData := "test"
	externalURL := "test"
	attributes := []Trait{}
	backgroundColor := "test"
	animationURL := "test"
	youtubeURL := "test"

	// Mint a new token for addr1
	dummy.mint(addr1, TokenID("1"))

	// Set metadata for token 1 as its owner (addr1)
	derr := dummy.SetTokenMetadata(addr1, TokenID("1"), Metadata{
		Name:            name,
		Description:     description,
		Image:           image,
		ImageData:       imageData,
		ExternalURL:     externalURL,
		Attributes:      attributes,
		BackgroundColor: backgroundColor,
		AnimationURL:    animationURL,
		YoutubeURL:      youtubeURL,
	})

	// Check if there was an error setting metadata
	uassert.NoError(t, derr, "Should not result in error")

	// Test case: invalid token ID. Setting metadata for a non-existent token fails.
	err := dummy.SetTokenMetadata(addr1, TokenID("3"), Metadata{
		Name:            name,
		Description:     description,
		Image:           image,
		ImageData:       imageData,
		ExternalURL:     externalURL,
		Attributes:      attributes,
		BackgroundColor: backgroundColor,
		AnimationURL:    animationURL,
		YoutubeURL:      youtubeURL,
	})

	uassert.ErrorIs(t, err, ErrInvalidTokenId)

	// Test case: caller is not the owner. Setting metadata from a non-owner (addr2) fails.
	cerr := dummy.SetTokenMetadata(addr2, TokenID("1"), Metadata{
		Name:            name,
		Description:     description,
		Image:           image,
		ImageData:       imageData,
		ExternalURL:     externalURL,
		Attributes:      attributes,
		BackgroundColor: backgroundColor,
		AnimationURL:    animationURL,
		YoutubeURL:      youtubeURL,
	})

	uassert.ErrorIs(t, cerr, ErrCallerIsNotOwner)

	// Retrieve metadata for token 1
	dummyMetadata, err := dummy.TokenMetadata(TokenID("1"))
	uassert.NoError(t, err, "Metadata error")

	// Check if metadata attributes match expected values
	uassert.Equal(t, image, dummyMetadata.Image)
	uassert.Equal(t, imageData, dummyMetadata.ImageData)
	uassert.Equal(t, externalURL, dummyMetadata.ExternalURL)
	uassert.Equal(t, description, dummyMetadata.Description)
	uassert.Equal(t, name, dummyMetadata.Name)
	uassert.Equal(t, len(attributes), len(dummyMetadata.Attributes))
	uassert.Equal(t, backgroundColor, dummyMetadata.BackgroundColor)
	uassert.Equal(t, animationURL, dummyMetadata.AnimationURL)
	uassert.Equal(t, youtubeURL, dummyMetadata.YoutubeURL)
}
