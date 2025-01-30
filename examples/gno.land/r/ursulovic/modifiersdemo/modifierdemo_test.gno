package modifiersdemo

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
)

func TestRegisterModifier(t *testing.T) {
	var user = testutils.TestAddress("user")
	std.TestSetOrigCaller(user)

	// Test non-owner calling
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Expected panic for non-owner registering modifier")
		}
	}()
	RegisterModifier("TestName", 0.5)

	// Test registering 2 modifiers with the same name
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Expected panic for trying to register 2 modifiers with the same name!")
		}
	}()
	RegisterModifier("TestName", 0.5)
	RegisterModifier("TestName", 0.5)

	// Test registering modifier with an empty name
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Expected panic for trying to register modifier with an empty name!")
		}
	}()
	RegisterModifier("", 0.5)

	// Test registering modifier with ApprovalPercentage higher than 1 or lower than 0
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Expected panic for trying to register modifier with approval percentage 0!")
		}
	}()
	RegisterModifier("ValidName", 0)
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Expected panic for trying to register modifier with approval percentage above 1!")
		}
	}()
	RegisterModifier("ValidName", 1.1)

	// Test register proper modifier
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Unexpected panic for registering proper modifier")
		}
	}()
	RegisterModifier("ValidName", 0.8)

}

func TestApproveModifier(t *testing.T) {
	var user = testutils.TestAddress("user")
	std.TestSetOrigCaller(user)

	// Test non-owner calling
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Expected panic for non-owner approving modifier")
		}
	}()
	ApproveModifier("ValidName")

	// Test approving modifier with invalid name (empty string)
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Expected panic for owner trying to approve an empty string modifier name")
		}
	}()
	ApproveModifier("")

}
