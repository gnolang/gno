package grc721

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/users"
)

func TestSetExtension(t *testing.T) {
	dummy := NewNFTWithMetadata(dummyNFTName, dummyNFTSymbol)
	if dummy == nil {
		t.Errorf("should not be nil")
	}

	addr1 := std.Address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")
	addr2 := std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

	name := "test"
	description := "test"
	image := "test"
	imageData := "test"
	externalURL := "test"
	attributes := []Trait{}
	backgroundColor := "test"
	animationURL := "test"
	youtubeURL := "test"

	std.TestSetOrigCaller(std.Address(addr1)) // addr1

	dummy.mint(addr1, TokenID("1"))

	_, derr := dummy.SetExtension(TokenID("1"), Metadata{
		Name : name,
		Description: description,
		Image: image,
		ImageData: imageData,
		ExternalURL: externalURL,
		Attributes: attributes,
		BackgroundColor: backgroundColor,
		AnimationURL: animationURL,
		YoutubeURL: youtubeURL,
	})

	if derr != nil {
		t.Errorf("Should not result in error ", derr.Error())
	}

	// Test case: Invalid token ID
	_, err := dummy.SetExtension(TokenID("3"), Metadata{
		Name : name,
		Description: description,
		Image: image,
		ImageData: imageData,
		ExternalURL: externalURL,
		Attributes: attributes,
		BackgroundColor: backgroundColor,
		AnimationURL: animationURL,
		YoutubeURL: youtubeURL,
	})
	if err != ErrInvalidTokenId {
		t.Errorf("Expected error %v, got %v", ErrInvalidTokenId, err)
	}

	std.TestSetOrigCaller(std.Address(addr2)) // addr2

	_, cerr := dummy.SetExtension(TokenID("1"), Metadata{
		Name : name,
		Description: description,
		Image: image,
		ImageData: imageData,
		ExternalURL: externalURL,
		Attributes: attributes,
		BackgroundColor: backgroundColor,
		AnimationURL: animationURL,
		YoutubeURL: youtubeURL,
	}) // addr2 trying to set Extension for token 1
	if cerr != ErrCallerIsNotOwner {
		t.Errorf("Expected error %v, got %v", ErrCallerIsNotOwner, cerr)
	}

	// Test case: Retrieving Extension
	std.TestSetOrigCaller(std.Address(addr1)) // addr1

	dummyExtension, err := dummy.Extension(TokenID("1"))
	if err != nil {
		t.Errorf("Extension error: %v", err.Error())

	} else {
		if dummyExtension.Image != image {
			t.Errorf("Expected Extension's image %v, got %v", image, dummyExtension.Image)
		}
		
		if dummyExtension.ImageData != imageData {
			t.Errorf("Expected Extension's imageData %v, got %v", imageData, dummyExtension.ImageData)
		}
		
		if dummyExtension.ExternalURL != externalURL {
			t.Errorf("Expected Extension's externalURL %v, got %v", externalURL, dummyExtension.ExternalURL)
		}
		
		if dummyExtension.Description != description {
			t.Errorf("Expected Extension's description %v, got %v", description, dummyExtension.Description)
		}
	
		if dummyExtension.Name != name {
			t.Errorf("Expected Extension's name %v, got %v", name, dummyExtension.Name)
		}
		
		if len(dummyExtension.Attributes) != len(attributes) {
			t.Errorf("Expected %d Extension's attributes %v, got %v", len(attributes), len(dummyExtension.Attributes))
		}
		 
		if dummyExtension.BackgroundColor != backgroundColor {
			t.Errorf("Expected Extension's backgroundColor %v, got %v", backgroundColor, dummyExtension.BackgroundColor)
		}
		
		if dummyExtension.AnimationURL != animationURL {
			t.Errorf("Expected Extension's animationURL %v, got %v", animationURL, dummyExtension.AnimationURL)
		}
		
		if dummyExtension.YoutubeURL != youtubeURL {
			t.Errorf("Expected Extension's youtubeURL %v, got %v", youtubeURL, dummyExtension.YoutubeURL)
		}
	}
	
	
}

