package std

import (
	"testing"
)

func TestAllocTrackerNotFinish(t *testing.T) {
	at := CreateATInstance()
	at.Register("Test AllocTracker Not Finish")
	isGood, errOnValidate := at.Validate()
	if isGood {
		t.Fatalf("Error on AllocTracker.Validate(), detail: %s", errOnValidate)
	}
}

func TestAllocTrackerNotRegistered(t *testing.T) {
	at := CreateATInstance()
	isGood, errOnValidate := at.Validate()
	if isGood {
		t.Fatalf("Error on AllocTracker.Validate(), detail: %s", errOnValidate)
	}
}

func TestAllocTrackerGetResult(t *testing.T) {
	at := CreateATInstance()
	result := at.GetResult()
	if result < 0 {
		t.Fatalf("Error on AllocTracker.GetResult()")
	}
}

