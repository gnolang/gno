// This is a support file for toml_testgen_test.go
package toml

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func testgenInvalid(t *testing.T, input string) {
	t.Logf("Input TOML:\n%s", input)
	tree, err := Load(input)
	if err != nil {
		return
	}

	typedTree := testgenTranslate(*tree)

	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(typedTree); err != nil {
		return
	}

	t.Fatalf("test did not fail. resulting tree:\n%s", buf.String())
}

func testgenValid(t *testing.T, input string, jsonRef string) {
	t.Logf("Input TOML:\n%s", input)
	tree, err := Load(input)
	if err != nil {
		t.Fatalf("failed parsing toml: %s", err)
	}

	typedTree := testgenTranslate(*tree)

	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(typedTree); err != nil {
		t.Fatalf("failed translating to JSON: %s", err)
	}

	var jsonTest any
	if err := json.NewDecoder(buf).Decode(&jsonTest); err != nil {
		t.Logf("translated JSON:\n%s", buf.String())
		t.Fatalf("failed decoding translated JSON: %s", err)
	}

	var jsonExpected any
	if err := json.NewDecoder(bytes.NewBufferString(jsonRef)).Decode(&jsonExpected); err != nil {
		t.Logf("reference JSON:\n%s", jsonRef)
		t.Fatalf("failed decoding reference JSON: %s", err)
	}

	if !reflect.DeepEqual(jsonExpected, jsonTest) {
		t.Logf("Diff:\n%#+v\n%#+v", jsonExpected, jsonTest)
		t.Fatal("parsed TOML tree is different than expected structure")
	}
}

func testgenTranslate(tomlData any) any {
	switch orig := tomlData.(type) {
	case map[string]any:
		typed := make(map[string]any, len(orig))
		for k, v := range orig {
			typed[k] = testgenTranslate(v)
		}
		return typed
	case *Tree:
		return testgenTranslate(*orig)
	case Tree:
		keys := orig.Keys()
		typed := make(map[string]any, len(keys))
		for _, k := range keys {
			typed[k] = testgenTranslate(orig.GetPath([]string{k}))
		}
		return typed
	case []*Tree:
		typed := make([]map[string]any, len(orig))
		for i, v := range orig {
			typed[i] = testgenTranslate(v).(map[string]any)
		}
		return typed
	case []map[string]any:
		typed := make([]map[string]any, len(orig))
		for i, v := range orig {
			typed[i] = testgenTranslate(v).(map[string]any)
		}
		return typed
	case []any:
		typed := make([]any, len(orig))
		for i, v := range orig {
			typed[i] = testgenTranslate(v)
		}
		return testgenTag("array", typed)
	case time.Time:
		return testgenTag("datetime", orig.Format("2006-01-02T15:04:05Z"))
	case bool:
		return testgenTag("bool", fmt.Sprintf("%v", orig))
	case int64:
		return testgenTag("integer", fmt.Sprintf("%d", orig))
	case float64:
		return testgenTag("float", fmt.Sprintf("%v", orig))
	case string:
		return testgenTag("string", orig)
	}

	panic(fmt.Sprintf("Unknown type: %T", tomlData))
}

func testgenTag(typeName string, data any) map[string]any {
	return map[string]any{
		"type":  typeName,
		"value": data,
	}
}
