package params

import (
	"testing"
)

func TestNewStringPropRequest(t *testing.T) {
	request := NewStringPropRequest("foo", "bar")
	if request.Title != "Set new sys/params key" {
		t.Errorf("wrong request title")
	}
}
