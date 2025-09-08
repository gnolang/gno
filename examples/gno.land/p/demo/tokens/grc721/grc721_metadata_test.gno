package grc721

import (
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/p/nt/uassert"
)

func TestSetMetadata(t *testing.T) {
	// Create a new dummy NFT with metadata
	dummy := NewNFTWithMetadata(dummyNFTName, dummyNFTSymbol)
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

	// Set the original caller to addr1
	testing.SetOriginCaller(addr1) // addr1

	// Mint a new token for addr1
	dummy.mint(addr1, TokenID("1"))

	// Set metadata for token 1
	derr := dummy.SetTokenMetadata(TokenID("1"), Metadata{
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

	// Test case: Invalid token ID
	err := dummy.SetTokenMetadata(TokenID("3"), Metadata{
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

	// Check if the error returned matches the expected error
	uassert.ErrorIs(t, err, ErrInvalidTokenId)

	// Set the original caller to addr2
	testing.SetOriginCaller(addr2) // addr2

	// Try to set metadata for token 1 from addr2 (should fail)
	cerr := dummy.SetTokenMetadata(TokenID("1"), Metadata{
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

	// Check if the error returned matches the expected error
	uassert.ErrorIs(t, cerr, ErrCallerIsNotOwner)

	// Set the original caller back to addr1
	testing.SetOriginCaller(addr1) // addr1

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
