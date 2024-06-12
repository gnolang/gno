package grc721

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/users"
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
	std.TestSetOrigCaller(addr1) // addr1

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
	if derr != nil {
		t.Errorf("Should not result in error : %s", derr.Error())
	}

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
	if err != ErrInvalidTokenId {
		t.Errorf("Expected error %s, got %s", ErrInvalidTokenId, err)
	}

	// Set the original caller to addr2
	std.TestSetOrigCaller(addr2) // addr2

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
	if cerr != ErrCallerIsNotOwner {
		t.Errorf("Expected error %s, got %s", ErrCallerIsNotOwner, cerr)
	}

	// Set the original caller back to addr1
	std.TestSetOrigCaller(addr1) // addr1

	// Retrieve metadata for token 1
	dummyMetadata, err := dummy.TokenMetadata(TokenID("1"))
	if err != nil {
		t.Errorf("Metadata error: %s", err.Error())
	}

	// Check if metadata attributes match expected values
	if dummyMetadata.Image != image {
		t.Errorf("Expected Metadata's image %s, got %s", image, dummyMetadata.Image)
	}
	if dummyMetadata.ImageData != imageData {
		t.Errorf("Expected Metadata's imageData %s, got %s", imageData, dummyMetadata.ImageData)
	}
	if dummyMetadata.ExternalURL != externalURL {
		t.Errorf("Expected Metadata's externalURL %s, got %s", externalURL, dummyMetadata.ExternalURL)
	}
	if dummyMetadata.Description != description {
		t.Errorf("Expected Metadata's description %s, got %s", description, dummyMetadata.Description)
	}
	if dummyMetadata.Name != name {
		t.Errorf("Expected Metadata's name %s, got %s", name, dummyMetadata.Name)
	}
	if len(dummyMetadata.Attributes) != len(attributes) {
		t.Errorf("Expected %d Metadata's attributes, got %d", len(attributes), len(dummyMetadata.Attributes))
	}
	if dummyMetadata.BackgroundColor != backgroundColor {
		t.Errorf("Expected Metadata's backgroundColor %s, got %s", backgroundColor, dummyMetadata.BackgroundColor)
	}
	if dummyMetadata.AnimationURL != animationURL {
		t.Errorf("Expected Metadata's animationURL %s, got %s", animationURL, dummyMetadata.AnimationURL)
	}
	if dummyMetadata.YoutubeURL != youtubeURL {
		t.Errorf("Expected Metadata's youtubeURL %s, got %s", youtubeURL, dummyMetadata.YoutubeURL)
	}
}
