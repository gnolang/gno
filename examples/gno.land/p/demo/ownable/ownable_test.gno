package ownable

import (
	"std"
	"testing"
)

func TestNew(t *testing.T) {
	std.TestSetOrigCaller("g1l9aypkr8xfvs82zeux486ddzec88ty69lue9de")

	expected := Ownable{
		owner: "g1l9aypkr8xfvs82zeux486ddzec88ty69lue9de",
	}

	result := New()
	if expected.owner != result.owner {
		t.Fatalf("Expected %s, got: %s\n", expected.owner.String(), result.owner.String())
	}
}

func TestGetOwner(t *testing.T) {
	std.TestSetOrigCaller("g1l9aypkr8xfvs82zeux486ddzec88ty69lue9de")

	result := New()
	resultOwner := result.GetOwner()

	expected := "g1l9aypkr8xfvs82zeux486ddzec88ty69lue9de"
	if resultOwner.String() != expected {
		t.Fatalf("Expected %s, got: %s\n", expected, result)
	}
}

func TestTransferOwnership(t *testing.T) {
	std.TestSetOrigCaller("g1l9aypkr8xfvs82zeux486ddzec88ty69lue9de")

	o := New()
	secondOwner := "g127jydsh6cms3lrtdenydxsckh23a8d6emqcvfa"

	err := o.TransferOwnership(std.Address(secondOwner))
	if err != nil {
		t.Fatalf("TransferOwnership failed, %v", err)
	}

	result := o.GetOwner()
	if secondOwner != result {
		t.Fatalf("Expected: %s, got: %s\n", secondOwner, result)
	}
}

func TestCallerIsOwner(t *testing.T) {
	std.TestSetOrigCaller("g1l9aypkr8xfvs82zeux486ddzec88ty69lue9de")

	o := New()
	unauthorizedCaller := "g127jydsh6cms3lrtdenydxsckh23a8d6emqcvfa"

	std.TestSetOrigCaller(std.Address(unauthorizedCaller))

	if o.CallerIsOwner() {
		t.Fatalf("Expected: %s to not be owner\n", unauthorizedCaller)
	}
}

func TestDropOwnership(t *testing.T) {
	std.TestSetOrigCaller("g1l9aypkr8xfvs82zeux486ddzec88ty69lue9de")

	o := New()

	err := o.DropOwnership()
	if err != nil {
		t.Fatalf("DropOwnership failed, %v", err)
	}

	owner := o.GetOwner()
	if owner != "" {
		t.Fatalf("Expected owner to be empty, not %s\n", owner)
	}
}

func TestErrors(t *testing.T) {

}
