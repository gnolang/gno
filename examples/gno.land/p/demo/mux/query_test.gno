package mux

import (
	"fmt"
	"testing"
)

func TestQuery_Get(t *testing.T) {
	cases := []struct {
		name           string
		reqPath        string
		key            string
		expectedOutput string
	}{
		// get by key
		{"get_by_key_normal", "users/?name=testname&age=12", "name", "testname"},
		{"get_same_query", "users/user?name=name1&name=name2", "name", "name1"},
		{"get_by_key_unprocesspath", "users/unprocesspath?user=thinhnx&age=12", "user", "thinhnx"},
	}

	for _, tt := range cases {
		name := fmt.Sprintf("%s-%s", tt.name, tt.reqPath)
		t.Run(name, func(t *testing.T) {
			req := &Request{
				HandlerPath: "",
				Path:        tt.reqPath,
			}
			output := req.URL().Get(tt.key)
			if output != tt.expectedOutput {
				t.Errorf("Expected %q, but got %q", tt.expectedOutput, output)
			}
		})
	}
}


func TestQuery_List(t *testing.T) {
	cases := []struct {
		name           string
		reqPath        string
		key            string
		expectedOutput []string
	}{
		// get all queries
		{"get_the_queries", "users/user?name=testname&age=12&name=thinhnx", "name", []string{"testname", "thinhnx"}},
	}

	for _, tt := range cases {
		name := fmt.Sprintf("%s-%s", tt.name, tt.reqPath)
		t.Run(name, func(t *testing.T) {
			req := &Request{
				HandlerPath: "",
				Path:        tt.reqPath,
			}
			output := req.URL()[tt.key]
			if len(output) != len(tt.expectedOutput) {
				t.Errorf("Expected %q, but got %q", tt.expectedOutput, output)
			}
			listOutput := req.URL()[tt.key]
			for i, ttExpt := range tt.expectedOutput {
				if ttExpt != listOutput[i] {
					t.Errorf("Expected %q, but got %q", ttExpt, listOutput[i])
				}
			}
		})
	}
}