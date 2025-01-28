package modifiersdemo

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
)

func TestRegisterModifier(t *testing.T) {
	var user = testutils.TestAddress("user")
	std.TestSetOrigCaller(user)

	//Test non-owner calling
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Excpected panic for non-owner registering modifier")
		}
	}()
	RegisterModifier("TestName", 0.5)

	//Test registering 2 modifiers with same name
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Excpected panic for trying to register 2 modifiers with same name!")
		}
	}()
	RegisterModifier("TestName", 0.5)

	//Test registering modifier with an empty name
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Excpected panic for trying for trying to register modifier with an empty name!")
		}
	}()
	RegisterModifier("", 0.5)

	//Test registering modifier with ApprovalPercentage higher than 1 or less or lower than 0
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Excpected panic for trying for trying to register modifier with approval percentage 0!")
		}
	}()
	RegisterModifier("ValidName", 0)
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Excpected panic for trying for trying to register modifier with approval percentage above 1!")
		}
	}()
	RegisterModifier("ValidName", 1.1)

}
