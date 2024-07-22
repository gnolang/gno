package mux

import (
	"fmt"
	"testing"
)

func TestRequest_GetVar(t *testing.T) {
	cases := []struct {
		handlerPath    string
		reqPath        string
		getVarKey      string
		expectedOutput string
	}{
		{"users/{id}", "users/123", "id", "123"},
		{"users/123", "users/123", "id", ""},
		{"users/{id}", "users/123", "nonexistent", ""},
		{"a/{b}/c/{d}", "a/42/c/1337", "b", "42"},
		{"a/{b}/c/{d}", "a/42/c/1337", "d", "1337"},
		{"{a}", "foo", "a", "foo"},
		{"/api/v1/{user}", "/api/v1/testuser?name=test", "user", "testuser"},
		{"/api/{v}/{user}/long/slash", "/api/v1/testuser?name=test/long/slash", "v", "v1"},
		{"/api/v1/{user}/long/slash", "/api/v1/testuser?name=test/long/slash", "user", "testuser"},
		// TODO: wildcards: a/*/c
		// TODO: multiple patterns per slashes: a/{b}-{c}/d
	}

	for _, tt := range cases {
		name := fmt.Sprintf("%s-%s", tt.handlerPath, tt.reqPath)
		t.Run(name, func(t *testing.T) {
			req := &Request{
				HandlerPath: tt.handlerPath,
				Path:        tt.reqPath,
			}

			output := req.GetVar(tt.getVarKey)
			if output != tt.expectedOutput {
				t.Errorf("Expected '%q, but got %q", tt.expectedOutput, output)
			}
		})
	}
}
