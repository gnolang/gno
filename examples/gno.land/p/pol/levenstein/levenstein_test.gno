package levenstein

import "testing"

func TestExample(t *testing.T) {
	if Distance("float", "boats") != 3 {
		t.Fatal("expected float vs boats to be 3")
	}
}