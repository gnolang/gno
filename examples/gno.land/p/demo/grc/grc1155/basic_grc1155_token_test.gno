package grc1155

import (
	"std"
	"testing"

	"gno.land/r/demo/users"
)

var dummyURI = "ipfs://xyz"

func TestNewBasicGRC1155Token(t *testing.T) {
	dummy := NewBasicGRC1155Token(dummyURI)
	if dummy == nil {
		return t.Errorf("should not be nil")
	}
}

func TestUri(t *testing.T) {
	dummy := NewBasicGRC1155Token(dummyURI)
	if dummy == nil {
		t.Errorf("should not be nil")
	}

	uri := dummy.Uri()
	if uri != dummyURI {
		t.Errorf("expected: (%s), got: (%s)", dummyURI, uri)
	}
}
