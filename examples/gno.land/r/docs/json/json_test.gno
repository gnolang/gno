package json

import "testing"

func TestPrettyPrintJSON(t *testing.T) {
	out, err := PrettyPrintJSON(`{"name":"Bob"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "{\n  \"name\": \"Bob\"\n}"
	if out != want {
		t.Errorf("object mismatch:\ngot:\n%s\nwant:\n%s", out, want)
	}
}

func TestPrettyPrintJSONArray(t *testing.T) {
	out, err := PrettyPrintJSON(`["apple","banana"]`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "[\n  \"apple\",\n  \"banana\"\n]"
	if out != want {
		t.Errorf("array mismatch:\ngot:\n%s\nwant:\n%s", out, want)
	}
}

func TestPrettyPrintJSONInvalid(t *testing.T) {
	if _, err := PrettyPrintJSON(`{not valid`); err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}
